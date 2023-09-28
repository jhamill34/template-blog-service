package repositories

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"golang.org/x/crypto/argon2"
)

type AuthRepository struct {
	userDao *dao.UserDao
}

func NewAuthRepository(userDao *dao.UserDao) *AuthRepository {
	return &AuthRepository{
		userDao: userDao,
	}
}

// LoginUser implements services.AuthService.
func (repo *AuthRepository) LoginUser(email, password string) (*models.User, error) {
	password = strings.TrimSpace(password)
	user, err := repo.userDao.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	ok, err := comparePasswords(password, user.HashedPassword)
	if err != nil {
		return nil, err
	}

	if ok {
		return &models.User{
			UserId: user.Id,
			Name:   user.Name,
			Email:  user.Email,
		}, nil
	}

	return nil, fmt.Errorf("Invalid User Credentials")
}

func comparePasswords(password, encodedPassword string) (bool, error) {
	params, salt, storedHash, err := decodeHash(encodedPassword)

	if err != nil {
		return false, err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		params.iterations,
		params.memory,
		params.parallelism,
		params.hashLength,
	)

	if subtle.ConstantTimeCompare(storedHash, hash) == 1 {
		return true, nil
	}

	return false, nil
}

type hashParams struct {
	iterations  uint32
	parallelism uint8
	memory      uint32
	hashLength  uint32
	saltLength  uint32
}

// "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
func decodeHash(encodedPassword string) (params *hashParams, salt []byte, hash []byte, err error) {
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

	params = &hashParams{}
	_, err = fmt.Sscanf(
		vals[3],
		"m=%d,t=%d,p=%d",
		&params.memory,
		&params.iterations,
		&params.parallelism,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	params.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	params.hashLength = uint32(len(hash))

	return params, salt, hash, nil
}

// var _ services.AuthService = (*AuthRepository)(nil)
