package main

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/app/gateway"
)

func main() {
	ctx := context.Background()
	gateway.Configure().Start(ctx)
}
