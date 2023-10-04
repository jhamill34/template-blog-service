package repositories

import (
	"bytes"
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
)

const ROOT_NAME = "ROOT"

type AuthRepository struct {
	userDao               *dao.UserDao
	passwordConfig        *config.HashParams
	verifyTokenService    services.VerifyTokenService
	emailService          services.EmailService
	templateService       services.TemplateService
	passwordForgotService services.VerifyTokenService
	inviteTokenService    services.TokenClaimsService
}

func NewAuthRepository(
	userDao *dao.UserDao,
	passwordConfig *config.HashParams,
	verifyTokenService services.VerifyTokenService,
	emailService services.EmailService,
	templateService services.TemplateService,
	passwordForgotService services.VerifyTokenService,
	inviteTokenService services.TokenClaimsService,
) *AuthRepository {
	return &AuthRepository{
		userDao:               userDao,
		passwordConfig:        passwordConfig,
		verifyTokenService:    verifyTokenService,
		emailService:          emailService,
		templateService:       templateService,
		passwordForgotService: passwordForgotService,
		inviteTokenService:    inviteTokenService,
	}
}

// LoginUser implements services.AuthService.
func (repo *AuthRepository) LoginUser(
	ctx context.Context,
	email, password string,
) (*models.SessionData, models.Notifier) {
	password = strings.TrimSpace(password)
	user, err := repo.userDao.FindByEmail(ctx, email)

	if err == database.NotFound {
		return nil, services.InvalidPassword
	}

	if err != nil {
		panic(err)
	}

	ok, err := comparePasswords(password, user.HashedPassword)
	if err != nil {
		panic(err)
	}

	if !ok {
		return nil, services.InvalidPassword
	}

	if !user.Verified {
		return nil, services.UnverifiedUser
	}

	return &models.SessionData{
		SessionId: uuid.New().String(),
		UserId:    user.Id,
		Name:      user.Name,
		Email:     user.Email,
		CsrfToken: uuid.New().String(),
	}, nil
}

func (repo *AuthRepository) GetUserByEmail(
	ctx context.Context,
	email string,
) (*models.User, models.Notifier) {
	user, err := repo.userDao.FindByEmail(ctx, email)

	if err == database.NotFound {
		return nil, services.AccountNotFound
	}

	if err != nil {
		panic(err)
	}

	return &models.User{
		UserId: user.Id,
		Name:   user.Name,
		Email:  user.Email,
	}, nil
}

func (repo *AuthRepository) GetUserByUsername(
	ctx context.Context,
	username string,
) (*models.User, models.Notifier) {
	user, err := repo.userDao.FindByUsername(ctx, username)

	if err == database.NotFound {
		return nil, services.AccountNotFound
	}

	if err != nil {
		panic(err)
	}

	return &models.User{
		UserId: user.Id,
		Name:   user.Name,
		Email:  user.Email,
	}, nil
}

func (repo *AuthRepository) CreateUser(
	ctx context.Context,
	username, email, password string,
) models.Notifier {
	if username == ROOT_NAME {
		return services.EmailAlreadyInUse
	}

	encodedHash, err := createHash(repo.passwordConfig, password)
	if err != nil {
		panic(err)
	}

	id := uuid.New().String()
	user, err := repo.userDao.CreateUser(ctx, id, username, email, encodedHash, false)
	if err == database.Duplicate {
		return services.EmailAlreadyInUse
	}

	if err != nil {
		panic(err)
	}

	repo.sendVerifyEmail(ctx, user)

	return nil
}

func (repo *AuthRepository) ResendVerifyEmail(
	ctx context.Context,
	email string,
) models.Notifier {
	user, err := repo.userDao.FindByEmail(ctx, email)
	if err == database.NotFound {
		return services.AccountNotFound
	}

	if err != nil {
		panic(err)
	}

	if user.Verified {
		return services.AccountAlreadyVerified
	}

	repo.sendVerifyEmail(ctx, user)

	return nil
}

type EmailWithTokenData struct {
	Token string
	Id    string
}

func (repo *AuthRepository) sendVerifyEmail(
	ctx context.Context,
	user *database.UserEntity,
) {
	token := repo.verifyTokenService.Create(ctx, user.Id)

	buffer := bytes.Buffer{}
	data := EmailWithTokenData{token, user.Id}
	repo.templateService.Render(
		&buffer,
		"register_email.html",
		"layout",
		models.NewTemplateData(data),
	)

	repo.emailService.SendEmail(ctx, user.Email, "Verify your email", buffer.String())
}

func (repo *AuthRepository) CreateRootUser(
	ctx context.Context,
	email, password string,
) models.Notifier {
	encodedHash, err := createHash(repo.passwordConfig, password)
	if err != nil {
		panic(err)
	}

	_, err = repo.userDao.CreateUser(ctx, ROOT_NAME, ROOT_NAME, email, encodedHash, true)

	if err == database.Duplicate {
		return services.EmailAlreadyInUse
	}

	if err != nil {
		panic(err)
	}

	return nil
}

func (repo *AuthRepository) ChangePassword(
	ctx context.Context,
	id, currentPassword, newPassword string,
) models.Notifier {
	user, err := repo.userDao.FindById(ctx, id)

	if err == database.NotFound {
		return services.AccountNotFound
	}

	if err != nil {
		panic(err)
	}

	ok, err := comparePasswords(currentPassword, user.HashedPassword)
	if err != nil {
		panic(err)
	}

	if ok {
		encodedHash, err := createHash(repo.passwordConfig, newPassword)
		if err != nil {
			panic(err)
		}

		err = repo.userDao.ChangePassword(ctx, id, encodedHash)
		if err != nil {
			panic(err)
		}
	} else {
		return services.InvalidPassword
	}

	return nil
}

func (repo *AuthRepository) ChangePasswordWithToken(
	ctx context.Context,
	id, token, newPassword string,
) models.Notifier {
	err := repo.passwordForgotService.Verify(ctx, id, token)
	if err == services.InvalidToken {
		return services.InvalidPasswordToken
	}

	if err != nil {
		panic(err)
	}

	encodedHash, hashErr := createHash(repo.passwordConfig, newPassword)
	if hashErr != nil {
		panic(hashErr)
	}

	daoErr := repo.userDao.ChangePassword(ctx, id, encodedHash)
	if daoErr != nil {
		panic(daoErr)
	}

	return nil
}

func (repo *AuthRepository) VerifyUser(
	ctx context.Context,
	id, token string,
) models.Notifier {
	err := repo.verifyTokenService.Verify(ctx, id, token)

	if err != nil {
		return services.InvalidRegistrationToken
	}

	daoErr := repo.userDao.VerifyUser(ctx, id)
	if daoErr != nil {
		panic(daoErr)
	}
	return nil
}

func (repo *AuthRepository) CreateForgotPasswordToken(
	ctx context.Context,
	email string,
) models.Notifier {
	user, err := repo.userDao.FindByEmail(ctx, email)
	if err == database.NotFound {
		return services.AccountNotFound
	}

	if err != nil {
		panic(err)
	}

	token := repo.passwordForgotService.Create(ctx, user.Id)

	buffer := bytes.Buffer{}
	data := EmailWithTokenData{token, user.Id}
	repo.templateService.Render(
		&buffer,
		"forgot_password_email.html",
		"layout",
		models.NewTemplateData(data),
	)

	repo.emailService.SendEmail(
		ctx,
		user.Email,
		"Reset your password email",
		buffer.String(),
	)

	return nil
}

func (repo *AuthRepository) InviteUser(
	ctx context.Context,
	fromUserId,
	email string,
) models.Notifier {
	if _, err := repo.userDao.FindByEmail(ctx, email); err != database.NotFound {
		return nil
	}

	newId := uuid.New().String()
	token := repo.inviteTokenService.CreateWithClaims(
		ctx,
		newId,
		&models.InviteData{InvitedBy: fromUserId, Email: email},
	)

	buffer := bytes.Buffer{}
	data := EmailWithTokenData{token, newId}
	repo.templateService.Render(
		&buffer,
		"invite_email.html",
		"layout",
		models.NewTemplateData(data),
	)

	repo.emailService.SendEmail(ctx, email, "You have been invited", buffer.String())

	return nil
}

func (repo *AuthRepository) VerifyInvite(
	ctx context.Context,
	id, token string,
	predicate func(*models.InviteData) bool,
) models.Notifier {
	var inviteData models.InviteData
	err := repo.inviteTokenService.VerifyWithClaims(ctx, id, token, &inviteData)
	if err != nil {
		return services.InvalidInviteToken
	}

	if !predicate(&inviteData) {
		return services.InvalidInviteToken
	}

	return nil
}

func (repo *AuthRepository) InvalidateInvite(
	ctx context.Context,
	id string,
) models.Notifier {
	repo.inviteTokenService.Destroy(ctx, id)
	return nil
}

// var _ services.AuthService = (*AuthRepository)(nil)
