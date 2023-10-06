package services

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/models"
)

type AuthService interface {
	LoginUser(
		ctx context.Context,
		email string,
		password string,
	) (*models.SessionData, models.Notifier)
	CreateUser(ctx context.Context, username, email, password string) models.Notifier
	VerifyInvite(
		ctx context.Context,
		id, token string,
		predicate func(*models.InviteData) bool,
	) models.Notifier
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

type ApplicationService interface {
	CreateApp(
		ctx context.Context,
		redirectUri, name, description string,
	) (*models.App, string, models.Notifier)
	GetApp(ctx context.Context, id string) (*models.App, models.Notifier)
	GetAppByClientId(ctx context.Context, clientId string) (*models.App, models.Notifier)
	DeleteApp(ctx context.Context, id string) models.Notifier
	ListApps(ctx context.Context) ([]models.App, models.Notifier)
	NewSecret(ctx context.Context, id string) (string, models.Notifier)
	NewAuthCode(ctx context.Context, userId, clientId string) string
	GetAuthCode(ctx context.Context, code string) (string, string, models.Notifier)
	ValidateAppSecret(ctx context.Context, id, secret string) (*models.App, models.Notifier)
	NewAccessToken(ctx context.Context, userId, clientId, refreshToken string) (*models.AccessTokenResponse, models.Notifier)
	VerifyAccessToken(ctx context.Context, accessToken string) bool
	FindRefreshToken(ctx context.Context, refreshToken string) (string, string, models.Notifier)
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
