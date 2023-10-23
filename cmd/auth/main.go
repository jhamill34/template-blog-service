package main

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/app/auth"
)

func main() {
	ctx := context.Background()
	auth.Configure().Start(ctx)
}
