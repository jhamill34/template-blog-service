package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AuthConfig struct {
	General Configuration `yaml:"general"`
}

func LoadAuthConfig(filename string) (AuthConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return AuthConfig{}, err
	}

	var config AuthConfig
	decoder := yaml.NewDecoder(file)

	err = decoder.Decode(&config)
	if err != nil {
		return AuthConfig{}, err
	}

	return config, nil
}
