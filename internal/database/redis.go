package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisProvider struct {
	store    *RedisStore
	prefix   string
	addr     string
	password string
}

// Get implements KeyValueStoreProvider.
func (self *RedisProvider) Get() KeyValueStore {
	if self.store == nil {
		self.store = NewRedisStore(self.prefix, redis.NewClient(&redis.Options{
			Addr:     self.addr,
			Password: self.password,
		}))
	}

	return self.store
}

// Close implements KeyValueStoreProvider.
func (self *RedisProvider) Close() error {
	if self.store != nil {
		return self.store.redisClient.Close()
	}
	return nil
}

func NewRedisProvider(prefix, addr, password string) *RedisProvider {
	return &RedisProvider{
		prefix:   prefix,
		addr:     addr,
		password: password,
	}
}

// ======================

type RedisStore struct {
	prefix      string
	redisClient *redis.Client
}

// Del implements KeyValueStore.
func (self *RedisStore) Del(ctx context.Context, key string) error {
	return self.redisClient.Del(ctx, self.prefix+key).Err()
}

// Expire implements KeyValueStore.
func (self *RedisStore) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return self.redisClient.Expire(ctx, self.prefix+key, expiration).Err()
}

// Get implements KeyValueStore.
func (self *RedisStore) Get(ctx context.Context, key string) (string, error) {
	return self.redisClient.Get(ctx, self.prefix+key).Result()
}

// Set implements KeyValueStore.
func (self *RedisStore) Set(
	ctx context.Context,
	key string,
	value interface{},
	expiration time.Duration,
) error {
	return self.redisClient.Set(ctx, self.prefix+key, value, expiration).Err()
}

func NewRedisStore(prefix string, redisClient *redis.Client) *RedisStore {
	return &RedisStore{
		prefix:      prefix,
		redisClient: redisClient,
	}
}
