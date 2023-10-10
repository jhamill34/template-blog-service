package app

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/services/session"
	"github.com/jhamill34/notion-provisioner/internal/transport"
	"github.com/jhamill34/notion-provisioner/internal/transport/routes"
	"github.com/redis/go-redis/v9"
)

type Gateway struct {
	server transport.Server
}

func ConfigureGateway() *Gateway {
	cfg, err := config.LoadGatewayConfig("configs/gateway.yaml")
	if err != nil {
		panic(err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Cache.Addr,
		Password: cfg.Cache.Password,
	})

	sessionStore := session.NewRedisSessionStore(
		redisClient,
		cfg.SessionConfig.TTL,
		cfg.SessionConfig.SigningKey,
	)

	return &Gateway{
		transport.NewServer(
			cfg.Server,
			routes.NewGatewayRoutes(sessionStore),
		),
	}
}

func (self *Gateway) Start(ctx context.Context) {
	self.server.Start(ctx)
}
