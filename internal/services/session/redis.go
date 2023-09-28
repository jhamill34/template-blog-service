package session

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisSessionStore struct {
	redisClient *redis.Client
}

func NewRedisSessionStore(redisClient *redis.Client) *RedisSessionStore {
	return &RedisSessionStore{
		redisClient: redisClient,
	}
}

// Create implements services.SessionService.
func (*RedisSessionStore) Create(ctx context.Context, data interface{}) string {
	panic("unimplemented")
}

// Destroy implements services.SessionService.
func (*RedisSessionStore) Destroy(ctx context.Context, id string) {
	panic("unimplemented")
}

// Find implements services.SessionService.
func (*RedisSessionStore) Find(ctx context.Context, id string) (interface{}, error) {
	panic("unimplemented")
}

// var _ services.SessionService = (*RedisSessionStore)(nil)
