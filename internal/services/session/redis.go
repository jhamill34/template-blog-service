package session

import "github.com/redis/go-redis/v9"

type RedisSessionStore struct {
	redisClient *redis.Client
}

func NewRedisSessionStore(redisClient *redis.Client) *RedisSessionStore {
	return &RedisSessionStore{

	}
}

// Create implements services.SessionService.
func (*RedisSessionStore) Create(data interface{}) string {
	panic("unimplemented")
}

// Destroy implements services.SessionService.
func (*RedisSessionStore) Destroy(id string) {
	panic("unimplemented")
}

// Find implements services.SessionService.
func (*RedisSessionStore) Find(id string) (interface{}, error) {
	panic("unimplemented")
}

// var _ services.SessionService = (*RedisSessionStore)(nil)
