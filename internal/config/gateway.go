package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type GatewayConfig struct {
	Server        ServerConfig        `yaml:"server"`
	Cache         RedisConfig         `yaml:"cache"`
	SessionConfig SessionConfig       `yaml:"session"`
	Notifications NotificationsConfig `yaml:"notifications"`
	Oauth         OauthConfig         `yaml:"oauth"`
	AppServer     StringFromEnv       `yaml:"app_server"`
	Template      TemplateConfig      `yaml:"template"`
}

type OauthConfig struct {
	ClientID             StringFromEnv `yaml:"client_id"`
	ClientSecret         StringFromEnv `yaml:"client_secret"`
	RedirectAuthorizeUri StringFromEnv `yaml:"redirect_authorize_uri"`
	TokenUri             StringFromEnv `yaml:"token_uri"`
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
