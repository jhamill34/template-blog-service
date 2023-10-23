package main

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/app/api"
)

func main() {
	ctx := context.Background()
	api.Configure().Start(ctx)
}
