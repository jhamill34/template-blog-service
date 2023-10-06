package main

import (
	"context"
	"os"

	"github.com/jhamill34/notion-provisioner/internal/app"
)

func main() {
	ctx := context.Background()
	
	service := os.Args[1]
	switch service {
	case "auth":
		app.ConfigureAuth().Start(ctx)
	case "app":
		app.ConfigureApp().Start(ctx)
	}
}
