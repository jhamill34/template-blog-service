package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/redis/go-redis/v9"
)

type VerificationType string

const (
	VerificationTypeRegistration   VerificationType = "registration:"
	VerificationTypeForgotPassword VerificationType = "forgot_password:"
)

type HashedVerifyTokenRepository struct {
	redisClient    *redis.Client
	ttl            time.Duration
	prefix         VerificationType
	passwordParams *config.HashParams
}

func NewHashedVerifyTokenRepository(
	redisClient *redis.Client,
	ttl time.Duration,
	prefix VerificationType,
	passwordParams *config.HashParams,
) *HashedVerifyTokenRepository {
	return &HashedVerifyTokenRepository{
		redisClient:    redisClient,
		ttl:            ttl,
		prefix:         prefix,
		passwordParams: passwordParams,
	}
}

// Create implements services.VerifyTokenService.
func (self *HashedVerifyTokenRepository) Create(
	ctx context.Context,
	id string,
) (string, error) {
	publicToken := uuid.New().String()
	token, err := createHash(self.passwordParams, publicToken)

	err = self.redisClient.Set(ctx, string(self.prefix)+id, token, self.ttl).Err()
	if err != nil {
		return "", err
	}

	return publicToken, nil
}

// Verify implements services.VerifyTokenService.
func (self *HashedVerifyTokenRepository) Verify(
	ctx context.Context,
	id, token string,
) error {
	hashedToken, err := self.redisClient.Get(ctx, string(self.prefix)+id).Result()
	if err != nil {
		return err
	}

	ok, err := comparePasswords(token, hashedToken)
	if err != nil {
		return err
	}

	if ok {
		err = self.redisClient.Del(ctx, string(self.prefix)+id).Err()
		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("Invalid Token")
}

// var _ services.VerifyTokenService = (*VerifyTokenRepository)(nil)
