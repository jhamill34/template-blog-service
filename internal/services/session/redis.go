package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const PREFIX = "session:"

type RedisSessionStore struct {
	redisClient *redis.Client
	ttl         time.Duration
}

// Create implements services.SessionService.
func (self *RedisSessionStore) Create(ctx context.Context, data interface{}) (string, error) {
	id, err := randomString(32)

	if err != nil {
		return "", err
	}

	// TODO: Sign a JWT and store the token instead of HSet
	err = self.redisClient.HSet(ctx, PREFIX+id, data).Err()
	if err != nil {
		return "", err
	}

	err = self.redisClient.Expire(ctx, PREFIX+id, self.ttl).Err()
	if err != nil {
		return "", err
	}

	return id, nil
}

// Destroy implements services.SessionService.
func (self *RedisSessionStore) Destroy(ctx context.Context, id string) error {
	err := self.redisClient.Del(ctx, PREFIX+id).Err()
	if err != nil {
		return err
	}

	return nil
}

// Find implements services.SessionService.
func (self *RedisSessionStore) Find(ctx context.Context, id string, result interface{}) error {
	var err error
	val, err := self.redisClient.Exists(ctx, PREFIX+id).Result()
	if err != nil {
		return err
	}

	if val == 0 {
		return fmt.Errorf("Session with id %s not found", id)
	}

	err = self.redisClient.HGetAll(ctx, PREFIX+id).Scan(result)
	if err != nil {
		return err
	}

	err = self.redisClient.Expire(ctx, PREFIX+id, self.ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

func NewRedisSessionStore(redisClient *redis.Client, ttl time.Duration) *RedisSessionStore {
	return &RedisSessionStore{
		redisClient: redisClient,
		ttl:         ttl,
	}
}

// Create implements services.SessionService.
func randomString(n int) (string, error) {
	buffer := make([]byte, n)

	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	id := base64.StdEncoding.EncodeToString(buffer)

	return id, nil
}

// var _ services.SessionService = (*RedisSessionStore)(nil)
