package main

import (
	"github.com/jhamill34/notion-provisioner/internal/app/migrator"
)

func main() {
	migrator.Configure().Run()
}
