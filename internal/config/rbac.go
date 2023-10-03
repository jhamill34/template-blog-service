package config

import (
	"io"
	"os"
)

func LoadRbacModel(location string) string {
	file, err := os.Open(location)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	return string(data)
}
