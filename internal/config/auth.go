package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type User struct {
	Password string `yaml:"password"`
	Email    string `yaml:"email"`
}

type HashParams struct {
	Iterations  uint32 `yaml:"iterations"`
	Parallelism uint8  `yaml:"parallelism"`
	Memory      uint32 `yaml:"memory"`
	HashLength  uint32 `yaml:"hash_length"`
	SaltLength  uint32 `yaml:"salt_length"`
}

type AuthConfig struct {
	General        Configuration `yaml:"general"`
	DefaultUser    *User         `yaml:"default_user"`
	PasswordConfig *HashParams   `yaml:"password_config"`
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
