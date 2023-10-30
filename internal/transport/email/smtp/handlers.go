package smtp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/transport/email"
)

func SmtpProtocolRouter(
	hostname string,
	messageHandler services.EmailReciever,
	authHandler services.EmailAuthHandler,
) *email.CommandRouter {
	router := email.NewCommandRouter()
	handler := SmtpHandler{hostname, messageHandler, authHandler}

	router.Register("HELO", handler.Hello)
	router.Register("EHLO", handler.Ehello)
	router.Register("MAIL", handler.Mail)
	router.Register("RCPT", handler.Recipient)
	router.Register("DATA", handler.Data)
	router.Register("RSET", handler.Reset)
	router.Register("NOOP", handler.Noop)
	router.Register("QUIT", handler.Quit)

	router.Register("AUTH", handler.Auth)

	return router
}

func HostnameFromContext(ctx context.Context) string {
	hostname, ok := ctx.Value("hostname").(string)
	if !ok {
		return ""
	}
	return hostname
}

func SenderFromContext(ctx context.Context) string {
	sender, ok := ctx.Value("sender").(string)
	if !ok {
		return ""
	}
	return sender
}

func RecipientFromContext(ctx context.Context) []string {
	recipients, ok := ctx.Value("recipients").([]string)
	if !ok {
		return []string{}
	}
	return recipients
}

func AuthFromContext(ctx context.Context) bool {
	auth, ok := ctx.Value("auth").(bool)
	if !ok {
		return false
	}
	return auth
}

type SmtpHandler struct {
	hostname       string
	messageHandler services.EmailReciever
	authHandler    services.EmailAuthHandler
}

func (self *SmtpHandler) Hello(r *email.Request, w email.ResponseWriter) {
	if len(r.Args()) != 1 {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Invalid syntax.",
		})
		return
	}

	hostname := HostnameFromContext(r.Context())
	if hostname != "" {
		r.Reset()
	}

	r.WithContext(context.WithValue(r.Context(), "hostname", r.Args()[0]))
	w.WriteResponse(email.Response{
		Code:    250,
		Message: "Go ahead",
	})
}

type Extensions struct {
	Domain string
	Names  []string
}

func (self Extensions) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("250-%s\r\n", self.Domain))

	for _, name := range self.Names[:len(self.Names)-1] {
		sb.WriteString(fmt.Sprintf("250-%s\r\n", name))
	}

	sb.WriteString(fmt.Sprintf("250 %s\r\n", self.Names[len(self.Names)-1]))

	return sb.String()
}

func (self *SmtpHandler) Ehello(r *email.Request, w email.ResponseWriter) {
	if len(r.Args()) != 1 {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Invalid syntax.",
		})
		return
	}

	hostname := HostnameFromContext(r.Context())
	if hostname != "" {
		r.Reset()
	}

	r.WithContext(context.WithValue(r.Context(), "hostname", r.Args()[0]))

	w.WriteResponse(Extensions{
		Domain: self.hostname,
		Names: []string{
			"STARTTLS",
			"AUTH LOGIN PLAIN",
		},
	})
}

func (self *SmtpHandler) Mail(r *email.Request, w email.ResponseWriter) {
	if self.authHandler != nil && !AuthFromContext(r.Context()) {
		w.WriteResponse(email.Response{
			Code:    530,
			Message: "Authentication required",
		})
		return
	}

	if len(r.Args()) != 1 {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Invalid syntax.",
		})
		return
	}

	parts := strings.Split(r.Args()[0], ":")
	if len(parts) != 2 || strings.ToUpper(parts[0]) != "FROM" {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Invalid syntax.",
		})
		return
	}

	hostname := HostnameFromContext(r.Context())
	if hostname == "" {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Please introduce yourself first.",
		})
		return
	}

	sender := SenderFromContext(r.Context())
	if sender != "" {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Duplicate MAIL",
		})
		return
	}

	addr := ""
	if parts[1] != "<>" {
		parsedAddr, err := mail.ParseAddress(parts[1])
		if err != nil {
			w.WriteResponse(email.Response{
				Code:    502,
				Message: "Malformed email address",
			})
			return
		}
		addr = parsedAddr.Address
	}

	r.WithContext(context.WithValue(r.Context(), "sender", addr))
	w.WriteResponse(email.Response{
		Code:    250,
		Message: "Go ahead",
	})
}

func (self *SmtpHandler) Recipient(r *email.Request, w email.ResponseWriter) {
	if self.authHandler != nil && !AuthFromContext(r.Context()) {
		w.WriteResponse(email.Response{
			Code:    530,
			Message: "Authentication required",
		})
		return
	}

	if len(r.Args()) != 1 {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Invalid syntax.",
		})
		return
	}

	parts := strings.Split(r.Args()[0], ":")
	if len(parts) != 2 || strings.ToUpper(parts[0]) != "TO" {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Invalid syntax.",
		})
		return
	}

	sender := SenderFromContext(r.Context())
	if sender == "" {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Missing MAIL FROM command.",
		})
		return
	}

	recipients := RecipientFromContext(r.Context())
	if len(recipients) >= r.Config().MaxRecipients {
		w.WriteResponse(email.Response{
			Code:    452,
			Message: "Too many recipients",
		})
		return
	}

	parsedAddr, err := mail.ParseAddress(parts[1])
	if err != nil {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Malformed email address",
		})
		return
	}
	addr := parsedAddr.Address
	recipients = append(recipients, addr)

	r.WithContext(context.WithValue(r.Context(), "recipients", recipients))
	w.WriteResponse(email.Response{
		Code:    250,
		Message: "Go ahead",
	})
}

func (self *SmtpHandler) Data(r *email.Request, w email.ResponseWriter) {
	if self.authHandler != nil && !AuthFromContext(r.Context()) {
		w.WriteResponse(email.Response{
			Code:    530,
			Message: "Authentication required",
		})
		return
	}

	recipients := RecipientFromContext(r.Context())
	if len(recipients) == 0 {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Missing RCPT TO command.",
		})
		return
	}

	w.WriteResponse(email.Response{
		Code:    354,
		Message: "Go ahead. End your data with <CR><LF>.<CR><LF>",
	})
	w.SetDeadline(time.Now().Add(r.Config().DataTimeout))

	// TODO: configure max size
	data, err := r.Body()
	if err == email.ErrMaxBodyLength {
		w.WriteResponse(email.Response{
			Code:    552,
			Message: "Message too large",
		})
		r.Reset()
		return
	}

	if err != nil {
		return
	}

	envelope := models.Envelope{
		Sender:     SenderFromContext(r.Context()),
		Recipients: RecipientFromContext(r.Context()),
		Data:       data,
	}

	go self.messageHandler.HandleMessage(r.Context(), &envelope)

	w.WriteResponse(email.Response{
		Code:    250,
		Message: "Thank you.",
	})
	r.Reset()
}

func (self *SmtpHandler) Auth(r *email.Request, w email.ResponseWriter) {
	if self.authHandler == nil {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Authentication not supported",
		})
		return
	}

	if len(r.Args()) < 1 {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Invalid syntax.",
		})
		return
	}

	hostname := HostnameFromContext(r.Context())
	if hostname == "" {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Please introduce yourself first.",
		})
		return
	}

	if !email.TLSFromContext(r.Context()) {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "TLS required",
		})
		return
	}

	username := ""
	password := ""

	method := strings.ToUpper(r.Args()[0])
	switch method {
	case "PLAIN":
		var err error
		encodedCreds := ""
		if len(r.Args()) != 2 {
			w.WriteResponse(email.Response{
				Code:    334,
				Message: "Give me your credentials",
			})

			encodedCreds, err = r.Prompt()
			if err != nil {
				return
			}
		} else {
			encodedCreds = r.Args()[1]
		}

		decodedCreds, err := base64.StdEncoding.DecodeString(encodedCreds)
		if err != nil {
			w.WriteResponse(email.Response{
				Code:    502,
				Message: "Couldn't decode your credentials",
			})
			return
		}

		creds := bytes.Split(decodedCreds, []byte{0})
		if len(creds) != 3 {
			w.WriteResponse(email.Response{
				Code:    502,
				Message: "Couldn't decode your credentials",
			})
			return
		}

		username = string(creds[1])
		password = string(creds[2])
	case "LOGIN":
		encodedUsername := ""
		var err error
		if len(r.Args()) != 2 {
			w.WriteResponse(email.Response{
				Code:    334,
				Message: "VXNlcm5hbWU6",
			})

			encodedUsername, err = r.Prompt()
			if err != nil {
				return
			}
		} else {
			encodedUsername = r.Args()[1]
		}

		decodedUsername, err := base64.StdEncoding.DecodeString(encodedUsername)
		if err != nil {
			w.WriteResponse(email.Response{
				Code:    502,
				Message: "Couldn't decode your username",
			})
			return
		}

		username = string(decodedUsername)

		w.WriteResponse(email.Response{
			Code:    334,
			Message: "UGFzc3dvcmQ6",
		})

		encodedPassword, err := r.Prompt()
		if err != nil {
			return
		}

		decodedPassword, err := base64.StdEncoding.DecodeString(encodedPassword)
		if err != nil {
			w.WriteResponse(email.Response{
				Code:    502,
				Message: "Couldn't decode your password",
			})
			return
		}

		password = string(decodedPassword)
	default:
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Unsupported authentication method",
		})
		return
	}

	if !self.authHandler.Authenticate(r.Context(), username, password) {
		w.WriteResponse(email.Response{
			Code:    535,
			Message: "Authentication failed",
		})
		return
	}

	r.WithContext(context.WithValue(r.Context(), "auth", true))

	w.WriteResponse(email.Response{
		Code:    235,
		Message: "Authentication successful",
	})
}

func (self *SmtpHandler) Reset(r *email.Request, w email.ResponseWriter) {
	r.Reset()
	w.WriteResponse(email.Response{
		Code:    250,
		Message: "Go ahead",
	})
}

func (self *SmtpHandler) Noop(r *email.Request, w email.ResponseWriter) {
	w.WriteResponse(email.Response{
		Code:    250,
		Message: "Go ahead",
	})
}

func (self *SmtpHandler) Quit(r *email.Request, w email.ResponseWriter) {
	w.WriteResponse(email.Response{
		Code:    221,
		Message: "OK, bye",
	})
	w.Close()
}
