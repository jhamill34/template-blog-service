package repositories

import (
	"bytes"
	"context"
	"fmt"
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
) (*models.User, error) {
	password = strings.TrimSpace(password)
	user, err := repo.userDao.FindByEmail(ctx, email)

	if err != nil {
		return nil, err
	}

	ok, err := comparePasswords(password, user.HashedPassword)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("Invalid User Credentials")
	}

	if !user.Verified {
		return nil, fmt.Errorf("User is not verified")
	}

	return &models.User{
		UserId: user.Id,
		Name:   user.Name,
		Email:  user.Email,
	}, nil
}

func (repo *AuthRepository) GetUserByEmail(
	ctx context.Context,
	email string,
) (*models.User, error) {
	user, err := repo.userDao.FindByEmail(ctx, email)

	if err == database.NotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
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
) (*models.User, error) {
	user, err := repo.userDao.FindByUsername(ctx, username)

	if err == database.NotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
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
) error {
	if username == ROOT_NAME {
		return fmt.Errorf("Cannot create user with reserved name: %s", ROOT_NAME)
	}

	encodedHash, err := createHash(repo.passwordConfig, password)
	if err != nil {
		return err
	}

	user, err := repo.userDao.CreateUser(ctx, username, email, encodedHash, false)
	if err != nil {
		return err
	}

	return repo.sendVerifyEmail(ctx, user)
}

func (repo *AuthRepository) ResendVerifyEmail(
	ctx context.Context,
	email string,
) error {
	user, err := repo.userDao.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	if user.Verified {
		return fmt.Errorf("User is already verified")
	}

	return repo.sendVerifyEmail(ctx, user)
}

type EmailWithTokenData struct {
	Token string
	Id    string
}

func (repo *AuthRepository) sendVerifyEmail(
	ctx context.Context,
	user *database.UserEntity,
) error {
	token, err := repo.verifyTokenService.Create(ctx, user.Id)
	if err != nil {
		return err
	}

	buffer := bytes.Buffer{}
	data := EmailWithTokenData{token, user.Id}
	repo.templateService.Render(&buffer, "register_email.html", "layout", data)

	return repo.emailService.SendEmail(ctx, user.Email, "Verify your email", buffer.String())
}

func (repo *AuthRepository) CreateRootUser(ctx context.Context, email, password string) error {
	encodedHash, err := createHash(repo.passwordConfig, password)
	if err != nil {
		return err
	}

	_, err = repo.userDao.CreateUser(ctx, ROOT_NAME, email, encodedHash, true)
	return err
}

func (repo *AuthRepository) ChangePassword(
	ctx context.Context,
	id, currentPassword, newPassword string,
) error {
	user, err := repo.userDao.FindById(ctx, id)
	if err != nil {
		return err
	}

	ok, err := comparePasswords(currentPassword, user.HashedPassword)
	if err != nil {
		return err
	}

	if ok {
		encodedHash, err := createHash(repo.passwordConfig, newPassword)
		if err != nil {
			return err
		}

		return repo.userDao.ChangePassword(ctx, id, encodedHash)
	}

	return fmt.Errorf("Invalid User Credentials")
}

func (repo *AuthRepository) ChangePasswordWithToken(
	ctx context.Context,
	id, token, newPassword string,
) error {
	err := repo.passwordForgotService.Verify(ctx, id, token)
	if err != nil {
		return err
	}

	encodedHash, err := createHash(repo.passwordConfig, newPassword)
	if err != nil {
		return err
	}

	return repo.userDao.ChangePassword(ctx, id, encodedHash)
}

func (repo *AuthRepository) VerifyUser(ctx context.Context, id, token string) error {
	err := repo.verifyTokenService.Verify(ctx, id, token)

	if err != nil {
		return err
	}

	return repo.userDao.VerifyUser(ctx, id)
}

func (repo *AuthRepository) CreateForgotPasswordToken(ctx context.Context, email string) error {
	user, err := repo.userDao.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	token, err := repo.passwordForgotService.Create(ctx, user.Id)

	buffer := bytes.Buffer{}
	data := EmailWithTokenData{token, user.Id}
	repo.templateService.Render(&buffer, "forgot_password_email.html", "layout", data)

	return repo.emailService.SendEmail(
		ctx,
		user.Email,
		"Reset your password email",
		buffer.String(),
	)
}

func (repo *AuthRepository) InviteUser(ctx context.Context, email string) error {
	if _, err := repo.userDao.FindByEmail(ctx, email); err != database.NotFound{
		return nil
	}

	user := ctx.Value("user").(*models.User)

	newId := uuid.New().String()
	token, err := repo.inviteTokenService.CreateWithClaims(
		ctx,
		newId,
		&models.InviteData{InvitedBy: user.UserId, Email: email},
	)
	if err != nil {
		return err
	}

	buffer := bytes.Buffer{}
	data := EmailWithTokenData{token, newId}
	repo.templateService.Render(&buffer, "invite_email.html", "layout", data)

	return repo.emailService.SendEmail(ctx, email, "You have been invited", buffer.String())
}

func (repo *AuthRepository) VerifyInvite(ctx context.Context, id, token string, predicate func(*models.InviteData) bool) (bool, error) {
	var inviteData models.InviteData
	err := repo.inviteTokenService.VerifyWithClaims(ctx, id, token, &inviteData)
	if err != nil {
		return false, err
	}

	if !predicate(&inviteData) {
		return false, nil
	} else {
		err = repo.inviteTokenService.Destroy(ctx, id)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

// var _ services.AuthService = (*AuthRepository)(nil)
