package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SmtpSender struct {
	smtpServer   string
	smtpUser     string
	smtpPassword string
	user         string
	domain       string
	tlsConfig    *tls.Config
}

func NewSmtpSender(
	smtpServer string,
	smtpUser string,
	smtpPassword string,
	user string,
	domain string,
	tlsConfig *tls.Config,
) *SmtpSender {
	return &SmtpSender{
		smtpServer:   smtpServer,
		smtpUser:     smtpUser,
		smtpPassword: smtpPassword,
		user:         user,
		domain:       domain,
		tlsConfig:    tlsConfig,
	}
}

// SendEmail implements services.EmailSender.
func (self *SmtpSender) SendEmail(
	ctx context.Context,
	to string,
	subject string,
	body string,
) error {
	log.Println("DIAL")
	client, err := smtp.Dial(self.smtpServer)
	if err != nil {
		return err
	}

	log.Println("STARTTLS")
	if err := client.StartTLS(self.tlsConfig); err != nil {
		return err
	}

	host := strings.Split(self.smtpServer, ":")[0]
	if err := client.Auth(smtp.PlainAuth("", self.smtpUser, self.smtpPassword, host)); err != nil {
		return err
	}

	fromEmail := fmt.Sprintf("%s@%s", self.user, self.domain)

	if err := client.Mail(fromEmail); err != nil {
		return err
	}

	if err := client.Rcpt(to); err != nil {
		return err
	}

	headers := map[string]string{
		"Message-ID":   fmt.Sprintf("<%s@%s>", uuid.New().String(), self.domain),
		"Subject":      subject,
		"From":         fmt.Sprintf("<%s>", fromEmail),
		"To":           fmt.Sprintf("<%s>", to),
		"Content-Type": "text/html",
		"Date":         time.Now().Format(time.RFC1123Z),
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	for key, value := range headers {
		_, err = fmt.Fprintf(writer, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}

	fmt.Fprint(writer, "\r\n")
	_, err = fmt.Fprint(writer, body)
	if err != nil {
		return err
	}

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
