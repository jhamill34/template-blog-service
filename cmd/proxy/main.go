package main

import (
	"context"
	"os"

	"github.com/jhamill34/notion-provisioner/internal/app/app"
	"github.com/jhamill34/notion-provisioner/internal/app/auth"
	"github.com/jhamill34/notion-provisioner/internal/app/gateway"
	"github.com/jhamill34/notion-provisioner/internal/app/mail"
)

func main() {
	if len(os.Args) != 2 {
		panic("Please provide a service to run")
	}

	ctx := context.Background()
	switch os.Args[1] {
	case "app":
		app.Configure().Start(ctx)
	case "auth":
		auth.Configure().Start(ctx)
	case "gateway":
		gateway.Configure().Start(ctx)
	case "mail":
		mail.Configure().Start(ctx)
	}
}
