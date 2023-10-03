package app

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/transport"
)

type App struct {
	server transport.Server
}

func ConfigureApp() *App {
	return &App{
		transport.NewServer(
			config.ServerConfig{},
		),
	}
}

func (a *App) Start(ctx context.Context) {
	a.server.Start(ctx)
}
