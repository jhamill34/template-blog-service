package services

import (
	"context"
	"io"

	"github.com/jhamill34/notion-provisioner/internal/models"
)

type AuthService interface {
	LoginUser(ctx context.Context, email string, password string) (*models.SessionData, *AuthServiceError)
	CreateUser(ctx context.Context, username, email, password string) *AuthServiceError
	VerifyInvite(ctx context.Context, id, token string, predicate func(*models.InviteData) bool) *AuthServiceError
	InviteUser(ctx context.Context, fromUserId, email string) *AuthServiceError
	InvalidateInvite(ctx context.Context, id string) *AuthServiceError
	ResendVerifyEmail(ctx context.Context, email string) *AuthServiceError  
	CreateRootUser(ctx context.Context, email, password string) *AuthServiceError
	ChangePassword(ctx context.Context, id, currentPassword, newPassword string) *AuthServiceError
	ChangePasswordWithToken(ctx context.Context, id, token, newPassword string) *AuthServiceError
	VerifyUser(ctx context.Context, id, token string) *AuthServiceError
	CreateForgotPasswordToken(ctx context.Context, email string) *AuthServiceError 

	GetUserByEmail(ctx context.Context, email string) (*models.User, *AuthServiceError)
	GetUserByUsername(ctx context.Context, username string) (*models.User, *AuthServiceError)
}

type UserService interface {
	ListUsers(ctx context.Context) []models.User
}

type VerifyTokenService interface {
	Verify(ctx context.Context, id string, token string) *TokenError
	Create(ctx context.Context, id string) string
}

type TokenClaimsService interface {
	VerifyWithClaims(ctx context.Context, id string, token string, data interface{}) *TokenError
	CreateWithClaims(ctx context.Context, id string, data interface{}) string
	Destroy(ctx context.Context, id string) 
} 

type SessionService interface {
	Create(ctx context.Context, data *models.SessionData) string
	Find(ctx context.Context, id string, data *models.SessionData) *SessionError
	Update(ctx context.Context, data *models.SessionData) *SessionError 
	Destroy(ctx context.Context, id string) 
}

type TemplateService interface {
	Render(w io.Writer, template string, layout string, model models.TemplateModel) error
}

type EmailService interface {
	SendEmail(ctx context.Context, to, subject, body string) 
}

type AccessControlService interface {
	Enforce(ctx context.Context, resource string, action string) *AccessControlError
}

