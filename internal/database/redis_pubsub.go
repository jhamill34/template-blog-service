package database

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisPublisherProvider struct {
	publisher *RedisPublisher
	addr      string
	password  string
}

// Get implements PublisherProvider.
func (self *RedisPublisherProvider) Get() Publisher {
	if self.publisher == nil {
		client := redis.NewClient(&redis.Options{
			Addr:     self.addr,
			Password: self.password,
		})

		var err error
		for i := 0; i < 10; i++ {
			log.Println("Trying to ping redis")
			err = client.Ping(context.Background()).Err()
			if err == nil {
				break
			}

			time.Sleep(5 * time.Second)
		}

		if err != nil {
			log.Fatal(err.Error())
		}

		self.publisher = &RedisPublisher{
			redisClient: client,
		}
	}

	return self.publisher
}

// Close implements PublisherProvider.
func (self *RedisPublisherProvider) Close() error {
	if self.publisher != nil {
		return self.publisher.redisClient.Close()
	}

	return nil
}

func NewRedisPublisherProvider(addr, password string) *RedisPublisherProvider {
	return &RedisPublisherProvider{
		addr:     addr,
		password: password,
	}
}

// ======================

type RedisPublisher struct {
	redisClient *redis.Client
}

// Publish implements Publisher.
func (self *RedisPublisher) Publish(
	ctx context.Context,
	channel string,
	message interface{},
) error {
	return self.redisClient.Publish(ctx, channel, message).Err()
}

// ======================

type RedisSubscriberProvider struct {
	subscriber *RedisSubscriber
	addr       string
	password   string
}

func (self *RedisSubscriberProvider) Get() Subscriber {
	if self.subscriber == nil {
		client := redis.NewClient(&redis.Options{
			Addr:     self.addr,
			Password: self.password,
		})

		var err error
		for i := 0; i < 10; i++ {
			log.Println("Trying to ping redis")
			err = client.Ping(context.Background()).Err()
			if err == nil {
				break
			}

			time.Sleep(5 * time.Second)
		}

		if err != nil {
			log.Fatal(err.Error())
		}

		self.subscriber = &RedisSubscriber{
			redisClient: client,
		}
	}

	return self.subscriber
}

func (self *RedisSubscriberProvider) Close() error {
	if self.subscriber != nil {
		return self.subscriber.redisClient.Close()
	}

	return nil
}

func NewRedisSubscriberProvider(addr, password string) *RedisSubscriberProvider {
	return &RedisSubscriberProvider{
		addr:     addr,
		password: password,
	}
}

// ======================

type RedisSubscriber struct {
	redisClient *redis.Client
}

// Subscribe implements Subscriber.
func (self *RedisSubscriber) Subscribe(ctx context.Context, channel string) (<-chan string, error) {
	subscriber := self.redisClient.Subscribe(ctx, channel)
	ch := make(chan string)

	go func() {
		for {
			msg, err := subscriber.ReceiveMessage(ctx)
			if err != nil {
				panic(err)
			}

			ch <- msg.Payload
		}
	}()

	return ch, nil
}
