package session

import (
	"github.com/jhamill34/notion-provisioner/internal"
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
func (self *RedisSessionStore) Create(data interface{}) string {
	id := internal.GenerateId(256)
	return id
}

// Destroy implements services.SessionService.
func (self *RedisSessionStore) Destroy(id string) {
	panic("unimplemented")
}

// Find implements services.SessionService.
func (self *RedisSessionStore) Find(id string) (interface{}, error) {
	panic("unimplemented")
}

// var _ services.SessionService = (*RedisSessionStore)(nil)
