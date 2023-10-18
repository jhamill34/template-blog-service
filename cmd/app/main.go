package main

import (
	"context"
	"os"

	"github.com/jhamill34/notion-provisioner/internal/app"
)

func main() {
	ctx := context.Background()

	service := os.Getenv("SERVICE")
	switch service {
	case "gateway":
		app.ConfigureGateway().Start(ctx)
	case "auth":
		app.ConfigureAuth().Start(ctx)
	case "app":
		app.ConfigureApp().Start(ctx)
	case "migration":
		app.ConfigureMigrator().Run()
	case "mail":
		app.ConfigureMail().Start(ctx)
	}
}
