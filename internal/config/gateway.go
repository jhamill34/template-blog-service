package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type GatewayConfig struct {
	Server        ServerConfig        `yaml:"server"`
	Template      TemplateConfig      `yaml:"template"`
	Cache         RedisConfig         `yaml:"cache"`
	SessionConfig SessionConfig       `yaml:"session"`
	Notifications NotificationsConfig `yaml:"notifications"`
}

func LoadGatewayConfig(filename string) (GatewayConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return GatewayConfig{}, err
	}
	defer file.Close()

	var config GatewayConfig
	decoder := yaml.NewDecoder(file)

	err = decoder.Decode(&config)
	if err != nil {
		return GatewayConfig{}, err
	}

	return config, nil
}
