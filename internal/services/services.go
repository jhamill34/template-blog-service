package services

import (
	"context"
	"io"

	"github.com/jhamill34/notion-provisioner/internal/models"
)

type TemplateService interface {
	Render(w io.Writer, template string, layout string, model models.TemplateModel) error
}

type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string)
}

type Signer interface {
	Sign(data []byte) (string, error)
	Verify(data []byte, signature string) error
}

type AccessControlService interface {
	Enforce(ctx context.Context, resource string, action string) models.Notifier
	Invalidate(ctx context.Context, id string)
}
