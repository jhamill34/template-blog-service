package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type VerificationType string

const (
	VerificationTypeRegistration   VerificationType = "registration:"
	VerificationTypeForgotPassword VerificationType = "forgot_password:"
)

type VerifyRegistrationTokenRepository struct {
	redisClient *redis.Client
	ttl         time.Duration
	prefix      VerificationType
}

func NewVerifyRegistrationTokenRepository(
	redisClient *redis.Client,
	ttl time.Duration,
	prefix VerificationType,
) *VerifyRegistrationTokenRepository {
	return &VerifyRegistrationTokenRepository{
		redisClient: redisClient,
		ttl:         ttl,
		prefix:      prefix,
	}
}

// Create implements services.VerifyTokenService.
func (self *VerifyRegistrationTokenRepository) Create(
	ctx context.Context,
	id string,
) (string, error) {
	token := uuid.New().String()
	err := self.redisClient.Set(ctx, string(self.prefix)+token, id, self.ttl).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}

// Verify implements services.VerifyTokenService.
func (self *VerifyRegistrationTokenRepository) Verify(
	ctx context.Context,
	token string,
) (string, error) {
	userId, err := self.redisClient.Get(ctx, string(self.prefix)+token).Result()
	if err != nil {
		return "", err
	}

	err = self.redisClient.Del(ctx, string(self.prefix)+token).Err()
	if err != nil {
		return "", err
	}

	return userId, nil
}

// var _ services.VerifyTokenService = (*VerifyTokenRepository)(nil)
