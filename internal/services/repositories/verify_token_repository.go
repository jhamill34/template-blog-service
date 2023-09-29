package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var PREFIX = "verify_token:"

type VerifyRegistrationTokenRepository struct {
	redisClient *redis.Client
	ttl         time.Duration
}

func NewVerifyRegistrationTokenRepository(
	redisClient *redis.Client,
	ttl time.Duration,
) *VerifyRegistrationTokenRepository {
	return &VerifyRegistrationTokenRepository{
		redisClient: redisClient,
		ttl:         ttl,
	}
}

// Create implements services.VerifyTokenService.
func (self *VerifyRegistrationTokenRepository) Create(ctx context.Context, id string) (string, error) {
	token := uuid.New().String()
	err := self.redisClient.Set(ctx, PREFIX+token, id, self.ttl).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}

// Verify implements services.VerifyTokenService.
func (self *VerifyRegistrationTokenRepository) Verify(ctx context.Context, token string) (string, error) {
	userId, err := self.redisClient.Get(ctx, PREFIX+token).Result()
	if err != nil {
		return "", err
	}

	err = self.redisClient.Del(ctx, PREFIX+token).Err()
	if err != nil {
		return "", err
	}

	return userId, nil
}

// var _ services.VerifyTokenService = (*VerifyTokenRepository)(nil)
