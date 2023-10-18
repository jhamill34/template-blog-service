package email

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"strings"
	"time"
)

type EmailServer struct {
	router              *CommandRouter
	tlsConfig           *tls.Config
	port int
	protocol string
}

func NewEmailServer(
	router *CommandRouter,
	tlsConfig *tls.Config,
	protocol string,
	port int,
) *EmailServer {
	return &EmailServer{router, tlsConfig, port, protocol}
}

func (s *EmailServer) Start(ctx context.Context) {
	port := fmt.Sprintf(":%d", s.port)

	listener, err := net.Listen(s.protocol, port)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection: ", err)
			continue
		}

		connectionContext := context.WithValue(ctx, "remote", conn.RemoteAddr().String())
		go s.HandleConnection(connectionContext, conn)
	}
}

func (s *EmailServer) HandleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	session := NewSession(ctx, conn, false)
	session.WriteResponse(Response{
		Code:    220,
		Message: "email_service ESMTP ready",
	})

	localCtx := ctx
	for session.HasNextCommand() {
		cmd, err := session.NextCommand(localCtx)
		log.Printf("Command: %s %s\n", cmd.Verb(), cmd.Args())
		if err != nil {
			log.Println("Error reading command: ", err)
			return
		}

		if cmd.Verb() == "STARTTLS" {
			if s.UpgradeTLS(ctx, conn, session) {
				cmd.Reset()
			}
		} else {
			s.router.Handle(&cmd, session)
		}

		localCtx = cmd.ctx
	}
}

func (s *EmailServer) UpgradeTLS(ctx context.Context, conn net.Conn, session *Session) bool {
	if session.tls {
		session.WriteResponse(Response{
			Code:    502,
			Message: "TLS already established",
		})
		return false
	}

	if s.tlsConfig == nil {
		session.WriteResponse(Response{
			Code:    502,
			Message: "TLS not supported",
		})
		return false
	}

	tlsConn := tls.Server(conn, s.tlsConfig)
	session.WriteResponse(Response{
		Code:    220,
		Message: "Go ahead",
	})

	if err := tlsConn.Handshake(); err != nil {
		session.WriteResponse(Response{
			Code:    550,
			Message: "Handshake error",
		})
		return false
	}

	*session = *NewSession(ctx, tlsConn, true)

	return true
}

type ResponseWriter interface {
	WriteResponse(response Response)
	Close()
}

type CommandHandler = func(command *Request, w ResponseWriter)

type CommandRouter struct {
	commands map[string]CommandHandler
}

func NewCommandRouter() *CommandRouter {
	return &CommandRouter{
		commands: make(map[string]CommandHandler),
	}
}

func (r *CommandRouter) Register(verb string, handler CommandHandler) {
	r.commands[verb] = handler
}

func (router *CommandRouter) Handle(r *Request, w ResponseWriter) {
	handler, ok := router.commands[r.Verb()]
	if !ok {
		w.WriteResponse(Response{
			Code:    502,
			Message: "Unsupported command.",
		})
		return
	}

	handler(r, w)
}

type Session struct {
	ctx     context.Context
	conn    net.Conn
	reader  *bufio.Reader
	writer  *bufio.Writer
	scanner *bufio.Scanner

	tls bool
}

func NewSession(ctx context.Context, conn net.Conn, tls bool) *Session {
	reader := bufio.NewReader(conn)
	scanner := bufio.NewScanner(reader)
	writer := bufio.NewWriter(conn)

	return &Session{
		ctx:     ctx,
		conn:    conn,
		reader:  reader,
		writer:  writer,
		scanner: scanner,
		tls:     tls,
	}
}

type Request struct {
	ctx        context.Context
	rstContext context.Context
	verb       string
	args       []string
	reader     *bufio.Reader
}

func RequestFromString(
	ctx context.Context,
	rstCtx context.Context,
	cmd string,
	reader *bufio.Reader,
) (Request, error) {
	fields := strings.Fields(cmd)

	if len(fields) > 0 {
		verb := strings.ToUpper(fields[0])

		args := []string{}

		if len(fields) > 1 {
			args = fields[1:]
		}

		return Request{
			ctx:        ctx,
			rstContext: rstCtx,
			verb:       verb,
			args:       args,
			reader:     reader,
		}, nil
	}

	return Request{}, errors.New("Invalid command.")
}

func (c *Request) Context() context.Context {
	return c.ctx
}

func (c *Request) WithContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *Request) Reset() {
	c.ctx = c.rstContext
}

func (c *Request) Verb() string {
	return c.verb
}

func (c *Request) Args() []string {
	return c.args
}

var ErrMaxBodyLength = errors.New("Maximum body length exceeded.")

func (c *Request) Body(maxLength int) ([]byte, error) {
	buffer := bytes.Buffer{}
	reader := textproto.NewReader(c.reader).DotReader()

	_, err := io.CopyN(&buffer, reader, int64(maxLength))

	if err == io.EOF {
		return buffer.Bytes(), nil
	}

	if err != nil {
		return nil, err
	}

	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return nil, err
	}

	return nil, ErrMaxBodyLength
}

func (s *Session) HasNextCommand() bool {
	return s.scanner.Scan()
}

func (s *Session) NextCommand(ctx context.Context) (Request, error) {
	if s.scanner.Err() != nil {
		return Request{}, s.scanner.Err()
	}

	line := s.scanner.Text()
	return RequestFromString(ctx, s.ctx, line, s.reader)
}

type Response struct {
	Code    int
	Message string
}

func (r Response) String() string {
	return fmt.Sprintf("%d %s\r\n", r.Code, r.Message)
}

func (s *Session) WriteResponse(response Response) {
	s.writer.WriteString(response.String())
	log.Printf("Response: %s", response.String())
	s.writer.Flush()
}

func (s *Session) Close() {
	s.writer.Flush()
	time.Sleep(200 * time.Millisecond)
	s.conn.Close()
}
