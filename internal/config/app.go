package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Server        ServerConfig        `yaml:"server"`
	Database      DatabaseConfig      `yaml:"database"`
	PubSub        RedisConfig		  `yaml:"pubsub"`
	Cache		  RedisConfig         `yaml:"cache"`
}

func LoadAppConfig(filename string) (AppConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return AppConfig{}, err
	}
	defer file.Close()

	var config AppConfig
	decoder := yaml.NewDecoder(file)

	err = decoder.Decode(&config)
	if err != nil {
		return AppConfig{}, err
	}

	return config, nil
}
