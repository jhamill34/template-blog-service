package app

import (
	"context"
	"net/http"

	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/services/repositories"
	"github.com/jhamill34/notion-provisioner/internal/services/session"
	"github.com/jhamill34/notion-provisioner/internal/transport"
	"github.com/jhamill34/notion-provisioner/internal/transport/routes"
)

type Gateway struct {
	server transport.Server
	cleanup func(ctx context.Context)
}

func ConfigureGateway() *Gateway {
	cfg, err := config.LoadGatewayConfig("configs/gateway.yaml")
	if err != nil {
		panic(err)
	}

	templateRepository := repositories.
		NewTemplateRepository(cfg.Template.Common...).
		AddTemplates(cfg.Template.Paths...)

	kv := database.NewRedisProvider("GATEWAY:", cfg.Cache.Addr, cfg.Cache.Password)

	sessionStore := session.NewRedisSessionStore(
		kv,
		cfg.SessionConfig.TTL,
		cfg.SessionConfig.SigningKey,
	)

	return &Gateway{
		server: transport.NewServer(
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
				cfg.Server.BaseUrl,
			),
		),
		cleanup: func(ctx context.Context) {
			// kv.Close()
		},
	}
}

func (self *Gateway) Start(ctx context.Context) {
	defer self.cleanup(ctx)

	self.server.Start(ctx)
}
