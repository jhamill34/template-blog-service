package services

import (
	"context"
	"io"

	"github.com/jhamill34/notion-provisioner/internal/models"
)

type AuthService interface {
	LoginUser(ctx context.Context, email string, password string) (*models.SessionData, models.Notifier)
	CreateUser(ctx context.Context, username, email, password string) models.Notifier
	VerifyInvite(ctx context.Context, id, token string, predicate func(*models.InviteData) bool) models.Notifier
	InviteUser(ctx context.Context, fromUserId, email string) models.Notifier
	InvalidateInvite(ctx context.Context, id string) models.Notifier
	ResendVerifyEmail(ctx context.Context, email string) models.Notifier
	CreateRootUser(ctx context.Context, email, password string) models.Notifier
	ChangePassword(ctx context.Context, id, currentPassword, newPassword string) models.Notifier
	ChangePasswordWithToken(ctx context.Context, id, token, newPassword string) models.Notifier
	VerifyUser(ctx context.Context, id, token string) models.Notifier
	CreateForgotPasswordToken(ctx context.Context, email string) models.Notifier

	GetUserByEmail(ctx context.Context, email string) (*models.User, models.Notifier)
	GetUserByUsername(ctx context.Context, username string) (*models.User, models.Notifier)
}

type UserService interface {
	ListUsers(ctx context.Context) ([]models.User, models.Notifier)
	GetUser(ctx context.Context, id string) (*models.User, models.Notifier)
	ListPolicies(ctx context.Context, id string) ([]models.Policy, models.Notifier)
	CreatePolicy(ctx context.Context, id, resource, action, effect string) models.Notifier
	DeletePolicy(ctx context.Context, id, policyId string) models.Notifier
}

type VerifyTokenService interface {
	Verify(ctx context.Context, id string, token string) models.Notifier
	Create(ctx context.Context, id string) string
}

type TokenClaimsService interface {
	VerifyWithClaims(ctx context.Context, id string, token string, data interface{}) models.Notifier
	CreateWithClaims(ctx context.Context, id string, data interface{}) string
	Destroy(ctx context.Context, id string) 
} 

type SessionService interface {
	Create(ctx context.Context, data *models.SessionData) string
	Find(ctx context.Context, id string, data *models.SessionData) models.Notifier
	Update(ctx context.Context, data *models.SessionData) models.Notifier
	Destroy(ctx context.Context, id string) 
}

type TemplateService interface {
	Render(w io.Writer, template string, layout string, model models.TemplateModel) error
}

type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) 
}

type AccessControlService interface {
	Enforce(ctx context.Context, resource string, action string) models.Notifier
	Invalidate(ctx context.Context, id string)
}

