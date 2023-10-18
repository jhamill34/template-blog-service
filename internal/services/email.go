package services

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/models"
)

type EmailSender interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

type EmailReciever interface {
	HandleMessage(ctx context.Context, message *models.Envelope)
}
