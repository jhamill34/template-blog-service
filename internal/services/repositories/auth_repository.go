package repositories

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"golang.org/x/crypto/argon2"
)

const ROOT_NAME = "ROOT"

type AuthRepository struct {
	userDao            *dao.UserDao
	passwordConfig     *config.HashParams
	verifyTokenService services.VerifyTokenService
	emailService       services.EmailService
	templateService    services.TemplateService
}

func NewAuthRepository(
	userDao *dao.UserDao,
	passwordConfig *config.HashParams,
	verifyTokenService services.VerifyTokenService,
	emailService services.EmailService,
	templateService services.TemplateService,
) *AuthRepository {
	return &AuthRepository{
		userDao:            userDao,
		passwordConfig:     passwordConfig,
		verifyTokenService: verifyTokenService,
		emailService:       emailService,
		templateService:    templateService,
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

	return repo.sendVerifyEmail(ctx, user)
}

type RegisterEmailData struct {
	Token string
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
	data := RegisterEmailData{ token }
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

func (repo *AuthRepository) VerifyUser(ctx context.Context, token string) error {
	userId, err := repo.verifyTokenService.Verify(ctx, token)

	if err != nil {
		return err
	}

	return repo.userDao.VerifyUser(ctx, userId)
}

func comparePasswords(password, encodedPassword string) (bool, error) {
	params, salt, storedHash, err := decodeHash(encodedPassword)

	if err != nil {
		return false, err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.HashLength,
	)

	if subtle.ConstantTimeCompare(storedHash, hash) == 1 {
		return true, nil
	}

	return false, nil
}

// "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",

func createHash(params *config.HashParams, password string) (string, error) {
	salt, err := randomBytes(params.SaltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.HashLength,
	)
	base64hash := base64.RawStdEncoding.EncodeToString(hash)
	base64salt := base64.RawStdEncoding.EncodeToString(salt)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.Memory,
		params.Iterations,
		params.Parallelism,
		base64salt,
		base64hash,
	), nil
}

func decodeHash(
	encodedPassword string,
) (params *config.HashParams, salt []byte, hash []byte, err error) {
	vals := strings.Split(encodedPassword, "$")

	if len(vals) != 6 || vals[0] != "" || vals[1] != "argon2id" {
		return nil, nil, nil, fmt.Errorf("Invalid hash format")
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}

	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("Invalid hash version")
	}

	params = &config.HashParams{}
	_, err = fmt.Sscanf(
		vals[3],
		"m=%d,t=%d,p=%d",
		&params.Memory,
		&params.Iterations,
		&params.Parallelism,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	params.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	params.HashLength = uint32(len(hash))

	return params, salt, hash, nil
}

func randomBytes(length uint32) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	return bytes, nil
}

// var _ services.AuthService = (*AuthRepository)(nil)
