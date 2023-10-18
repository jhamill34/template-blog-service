package services

import (
	"context"
	"io"

	"github.com/jhamill34/notion-provisioner/internal/models"
)

type TemplateService interface {
	Render(w io.Writer, template string, layout string, model models.TemplateModel) error
}

type Signer interface {
	Sign(data []byte) (string, error)
	Verify(data []byte, signature string) error
}

type AccessControlService interface {
	Enforce(ctx context.Context, resource string, action string) models.Notifier
	Invalidate(ctx context.Context, id string)
}

type SessionService interface {
	Create(ctx context.Context, data *models.SessionData) string
	Find(ctx context.Context, id string, data *models.SessionData) models.Notifier
	UpdateCsrf(ctx context.Context, id, csrfToken string) models.Notifier
	Destroy(ctx context.Context, id string)
}
