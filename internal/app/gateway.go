package app

import (
	"context"
	"net/http"

	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/services/repositories"
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

	templateRepository := repositories.
		NewTemplateRepository(cfg.Template.Common...).
		AddTemplates(cfg.Template.Paths...)

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
			routes.NewGatewayRoutes(
				sessionStore,
				templateRepository,
				http.DefaultClient,
				cfg.SessionConfig,
				cfg.Oauth,
				cfg.AuthServer,
				cfg.ExternalAuthServer,
				cfg.AppServer,
				cfg.Notifications,
			),
		),
	}
}

func (self *Gateway) Start(ctx context.Context) {
	self.server.Start(ctx)
}
