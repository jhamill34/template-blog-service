package app

import (
	"github.com/jhamill34/notion-provisioner/internal/transport"
	"github.com/jhamill34/notion-provisioner/internal/transport/routes"
)

type App struct {
	server transport.Server
}

func ConfigureApp() *App {
	return &App{
		transport.NewServer(
			3333,
			routes.NewBlogRoutes(),
		),
	}
}

func (a *App) Start() {
	a.server.Start()
}
