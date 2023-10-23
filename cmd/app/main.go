package main

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/app/app"
)

func main() {
	ctx := context.Background()
	api.Configure().Start(ctx)
}
