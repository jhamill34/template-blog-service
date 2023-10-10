package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type GatewayConfig struct {
	Server             ServerConfig        `yaml:"server"`
	Template           TemplateConfig      `yaml:"template"`
	Cache              RedisConfig         `yaml:"cache"`
	SessionConfig      SessionConfig       `yaml:"session"`
	Notifications      NotificationsConfig `yaml:"notifications"`
	Oauth              OauthConfig         `yaml:"oauth"`
	AuthServer         string              `yaml:"auth_server"`
	ExternalAuthServer string              `yaml:"external_auth_server"`
	AppServer          string              `yaml:"app_server"`
	BaseUrl            string              `yaml:"base_url"`
}

type OauthConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	AuthorizeUri string `yaml:"authorize_uri"`
	TokenUri     string `yaml:"token_uri"`
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
