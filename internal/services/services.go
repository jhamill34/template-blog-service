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
	ChangePassword(ctx context.Context, id, currentPassword, newPassword string) error
	ChangePasswordWithToken(ctx context.Context, id, token, newPassword string) error
	VerifyUser(ctx context.Context, id, token string) error
	ResendVerifyEmail(ctx context.Context, email string) error
	CreateForgotPasswordToken(ctx context.Context, email string) error
	InviteUser(ctx context.Context, email string) error
	VerifyInvite(ctx context.Context, id, token string, predicate func(*models.InviteData) bool) (bool, error)
}

type SessionService interface {
	Create(ctx context.Context, data interface{}) (string, error)
	Destroy(ctx context.Context, id string) error
	Find(ctx context.Context, id string, data interface{}) error
}

type TemplateService interface {
	Render(w io.Writer, template string, layout string, data interface{}) error
}

type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

type VerifyTokenService interface {
	Verify(ctx context.Context, id string, token string) error
	Create(ctx context.Context, id string) (string, error)
}


type TokenClaimsService interface {
	VerifyWithClaims(ctx context.Context, id string, token string, data interface{}) error
	CreateWithClaims(ctx context.Context, id string, data interface{}) (string, error)
	Destroy(ctx context.Context, id string) error
} 

