package database

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type DatabaseProvider interface {
	Get() *sqlx.DB
	Close() error
}

type KeyValueStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, key string) error
	Expire(ctx context.Context, key string, expiration time.Duration) error
}

type KeyValueStoreProvider interface {
	Get() KeyValueStore
	Close() error
}

type PublisherProvider interface {
	Get() Publisher 
	Close() error
}

type Publisher interface {
	Publish(ctx context.Context, channel string, message interface{}) error
}

type SubscriberProvider interface {
	Get() Subscriber 
	Close() error
}

type Subscriber interface {
	Subscribe(ctx context.Context, channel string) (<-chan string, error)
}
