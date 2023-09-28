package main

import (
	"os"

	"github.com/jhamill34/notion-provisioner/internal/app"
)

func main() {
	service := os.Args[1]
	switch service {
	case "auth":
		app.ConfigureAuth().Start()
	}
}
