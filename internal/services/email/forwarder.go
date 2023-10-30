package email

import (
	"bufio"
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/smtp"
	"net/textproto"
	"sort"
	"strings"

	"github.com/jhamill34/notion-provisioner/internal/models"
)

type EmailForwarder struct {
	ports          []int
	signingOptions SigningOptions
}

func NewEmailForwarder(ports []int, options SigningOptions) *EmailForwarder {
	return &EmailForwarder{ports, options}
}

// HandleMessage implements services.EmailReciever.
func (f *EmailForwarder) HandleMessage(ctx context.Context, message *models.Envelope) {
	for _, recipient := range message.Recipients {
		parts := strings.Split(recipient, "@")
		domain := parts[1]

		client, err := makeClient(f.ports, domain)
		if err != nil {
			log.Println(err)
			continue
		}

		if err := forward(client, message, f.signingOptions); err != nil {
			log.Println(err)
			continue
		}
	}
}

func forward(client *smtp.Client, message *models.Envelope, options SigningOptions) error {
	if err := client.Hello(options.Domain); err != nil {
		return err
	}

	if err := client.Mail(message.Sender); err != nil {
		return err
	}

	for _, recipient := range message.Recipients {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	Sign(bytes.NewReader(message.Data), writer, options)

	err = writer.Close()
	if err != nil {
		return err
	}

	err = client.Quit()
	if err != nil {
		return err
	}

	return nil
}

func makeClient(ports []int, domain string) (client *smtp.Client, err error) {
	mx, err := net.LookupMX(domain)
	if err != nil {
		return nil, err
	}

	for _, record := range mx {
		for j := range ports {
			server := strings.TrimSuffix(record.Host, ".")
			addr := fmt.Sprintf("%s:%d", server, ports[j])

			client, err = smtp.Dial(addr)
			if err != nil {
				if j == len(ports)-1 {
					return nil, err
				}

				continue
			}

			return client, nil
		}
	}

	return nil, fmt.Errorf("Unable to connect to any servers for %s on any common port", domain)
}

type SigningOptions struct {
	Selector   string
	Domain     string
	Headers    []string
	PrivateKey crypto.Signer
}

type Signer struct {
	signatureFields map[string]string
	w               io.WriteCloser
	done            <-chan error
}

func NewSigner(options SigningOptions) *Signer {
	r, w := io.Pipe()
	done := make(chan error, 1)

	s := &Signer{nil, w, done}

	go s.signRoutine(r, done, options)

	return s
}

func (s *Signer) Signature() (string, error) {
	if s.signatureFields == nil {
		return "", nil
	}

	return formatDkimHeader(s.signatureFields)
}

func (s *Signer) Write(p []byte) (n int, err error) {
	return s.w.Write(p)
}

func (s *Signer) Close() error {
	if err := s.w.Close(); err != nil {
		return err
	}

	return <-s.done
}

func (s *Signer) signRoutine(data io.Reader, done chan<- error, options SigningOptions) {
	defer close(done)
	reader := bufio.NewReader(data)

	var headers []string
	headerReader := textproto.NewReader(reader)
	for {
		line, err := headerReader.ReadLine()
		if err != nil {
			done <- err
			return
		}

		if line == "" {
			break
		} else if line[0] == ' ' || line[0] == '\t' {
			headers[len(headers)-1] += line + "\r\n"
		} else {
			headers = append(headers, line)
		}
	}

	hasher := sha256.New()
	can := NewBodyCanonicalizer(hasher)

	if _, err := io.Copy(can, reader); err != nil {
		done <- err
		return
	}
	bodyHash := hasher.Sum(nil)

	fields := map[string]string{
		"v":  "1",
		"a":  "rsa-sha256",
		"c":  "relaxed/relaxed",
		"d":  options.Domain,
		"s":  options.Selector,
		"h":  strings.Join(options.Headers, ":"),
		"bh": base64.StdEncoding.EncodeToString(bodyHash),
	}

	hasher.Reset()
	for _, headerKey := range options.Headers {
		for _, line := range headers {
			if strings.HasPrefix(line, headerKey+":") {
				canonHeader := canonicalizeHeader(line)
				io.WriteString(hasher, canonHeader)
			}
		}
	}
	dkimHeader, err := formatDkimHeader(fields)
	if err != nil {
		done <- err
		return
	}

	canonDkimHeader := strings.Trim(canonicalizeHeader(dkimHeader), "\r\n")
	io.WriteString(hasher, canonDkimHeader)
	allHashed := hasher.Sum(nil)
	sig, err := options.PrivateKey.Sign(rand.Reader, allHashed, crypto.SHA256)
	if err != nil {
		done <- err
		return
	}

	fields["b"] = base64.StdEncoding.EncodeToString(sig)

	s.signatureFields = fields
}

func Sign(input io.Reader, output io.Writer, options SigningOptions) error {
	var buffer bytes.Buffer
	signer := NewSigner(options)

	writer := io.MultiWriter(&buffer, signer)

	_, err := io.Copy(writer, input)
	if err != nil {
		return err
	}

	err = signer.Close()
	if err != nil {
		return err
	}

	sig, err := signer.Signature()
	if err != nil {
		return err
	}

	_, err = io.WriteString(output, sig)
	if err != nil {
		return err
	}

	_, err = io.Copy(output, &buffer)
	if err != nil {
		return err
	}

	return nil
}

func foldSignature(signature string) (result string, err error) {
	buffer := bytes.NewBufferString(signature)

	line := make([]byte, 75)
	for {
		n, err := buffer.Read(line)
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", err
		}

		if n == 0 {
			continue
		}

		if result != "" {
			result += "\r\n "
		}

		result += string(line[:n])
	}

	return result, nil
}

func formatDkimHeader(params map[string]string) (string, error) {
	header := "DKIM-Signature:"

	signature := ""
	keys := make([]string, 0, len(params))
	for key := range params {
		if key == "b" {
			signature = params[key]
		} else {
			keys = append(keys, key)
		}
	}

	sort.Strings(keys)

	line := ""
	for _, key := range keys {
		value := params[key]

		// The three is to account for the space, equals, semicolon,
		// and \r\n at the end and space at the beginning
		if len(line)+len(key)+len(value)+6 > 78 {
			header += line + "\r\n"
			line = ""
		}

		line = line + " " + key + "=" + value + ";"
	}

	if line != "" {
		header += line
	}

	sig, err := foldSignature("b="+signature)
	if err != nil {
		return "", err
	}
	header += "\r\n " + sig

	return header + "\r\n", nil
}

func canonicalizeHeader(header string) string {
	kv := strings.SplitN(header, ":", 2)

	key := strings.ToLower(kv[0])
	value := kv[1]

	newValue := make([]byte, 0, len(value))

	i := 0
	for i < len(value) {
		current := byte(value[i])

		i++
		if match(current, ' ', '\t', '\r', '\n') {
			for i < len(value) && match(byte(value[i]), ' ', '\t', '\r', '\n') {
				i++
			}

			// Only add a space if we've seen non space characters first
			// or if we're not at the end of the input
			if len(newValue) > 0 && i < len(value) {
				newValue = append(newValue, ' ')
			}
		} else {
			newValue = append(newValue, current)
		}
	}

	return key + ":" + string(newValue) + "\r\n"
}

type BodyCanonicalizer struct {
	w io.Writer
}

func NewBodyCanonicalizer(w io.Writer) *BodyCanonicalizer {
	return &BodyCanonicalizer{w}
}

func (c *BodyCanonicalizer) Write(p []byte) (n int, err error) {
	newValue := make([]byte, 0, len(p))
	newLines := make([]byte, 0, 2)
	ws := false
	cr := false

	for _, current := range p {
		prevCr := cr
		cr = false
		if current == '\t' || current == ' ' {
			ws = true
		} else if current == '\r' {
			cr = true
			ws = false
			newLines = append(newLines, current)
		} else if current == '\n' {
			if !prevCr {
				newLines = append(newLines, '\r')
			}

			newLines = append(newLines, current)
			ws = false
		} else {
			if len(newLines) > 0 {
				newValue = append(newValue, newLines...)
				newLines = newLines[:0]
			}
			if ws {
				newValue = append(newValue, ' ')
				ws = false
			}

			newValue = append(newValue, current)
		}
	}

	n, err = c.w.Write(newValue)
	if err != nil {
		return n, err
	}

	n, err = c.w.Write([]byte("\r\n"))
	if err != nil {
		return n, err
	}

	return len(p), nil
}

func match(current byte, expected ...byte) bool {
	for _, b := range expected {
		if current == b {
			return true
		}
	}

	return false
}
