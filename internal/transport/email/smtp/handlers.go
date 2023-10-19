package smtp

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/transport/email"
)

func SmtpProtocolRouter(messageHandler services.EmailReciever) *email.CommandRouter {
	router := email.NewCommandRouter()
	handler := SmtpHandler{messageHandler}

	router.Register("HELO", handler.Hello)
	router.Register("MAIL", handler.Mail)
	router.Register("RCPT", handler.Recipient)
	router.Register("DATA", handler.Data)
	router.Register("RSET", handler.Reset)
	router.Register("NOOP", handler.Noop)
	router.Register("QUIT", handler.Quit)

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

type SmtpHandler struct {
	messageHandler services.EmailReciever
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

func (self *SmtpHandler) Mail(r *email.Request, w email.ResponseWriter) {
	if len(r.Args()) != 1 {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Invalid syntax.",
		})
		return
	}

	parts := strings.Split(r.Args()[0], ":")
	if len(parts) != 2 || parts[0] != "FROM" {
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
	if len(r.Args()) != 1 {
		w.WriteResponse(email.Response{
			Code:    502,
			Message: "Invalid syntax.",
		})
		return
	}

	parts := strings.Split(r.Args()[0], ":")
	if len(parts) != 2 || parts[0] != "TO" {
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
			Code: 452, 
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
