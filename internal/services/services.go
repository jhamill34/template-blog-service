package services

import (
	"context"
	"io"

	"github.com/jhamill34/notion-provisioner/internal/models"
)

type AuthService interface {
	LoginUser(ctx context.Context, email string, password string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	CreateUser(ctx context.Context, username, email, password string) error
	CreateRootUser(ctx context.Context, email, password string) error
}

type SessionService interface {
	Create(ctx context.Context, data interface{}) string
	Destroy(ctx context.Context, id string)
	Find(ctx context.Context, id string) (interface{}, error)
}

type TemplateService interface {
	Render(w io.Writer, template string, layout string, data interface{}) error
}
