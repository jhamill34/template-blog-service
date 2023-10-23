package gateway

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

func Configure() *Gateway {
	cfg, err := config.LoadGatewayConfig("configs/gateway.yaml")
	if err != nil {
		panic(err)
	}

	templateRepository := repositories.
		NewTemplateRepository(cfg.Template.Common...).
		AddTemplates(cfg.Template.Paths...)

	kv := database.NewRedisProvider("GATEWAY:", cfg.Cache.Addr.String(), cfg.Cache.Password.String())

	sessionStore := session.NewRedisSessionStore(
		kv,
		cfg.SessionConfig.TTL,
		[]byte(cfg.SessionConfig.SigningKey.String()),
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
				cfg.AppServer.String(),
				cfg.Notifications,
				cfg.Server.BaseUrl.String(),
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
