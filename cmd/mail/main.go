package main

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/app/mail"
)

func main() {
	ctx := context.Background()
	mail.Configure().Start(ctx)
}
