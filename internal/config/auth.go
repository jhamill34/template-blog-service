package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type User struct {
	Password string `yaml:"password"`
	Email    string `yaml:"email"`
}

type App struct {
	ClientId     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectUri  string `yaml:"redirect_uri"`
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
}

type HashParams struct {
	Iterations  uint32 `yaml:"iterations"`
	Parallelism uint8  `yaml:"parallelism"`
	Memory      uint32 `yaml:"memory"`
	HashLength  uint32 `yaml:"hash_length"`
	SaltLength  uint32 `yaml:"salt_length"`
	Secret      string `yaml:"secret"`
}

type EmailParams struct {
	Domain     string `yaml:"domain"`
	User       string `yaml:"user"`
	SmtpDomain string `yaml:"smtp_domain"`
	SmtpPort   int    `yaml:"smtp_port"`
}

type AuthConfig struct {
	Server            ServerConfig             `yaml:"server"`
	Database          DatabaseConfig           `yaml:"database"`
	Template          TemplateConfig           `yaml:"template"`
	PubSub            RedisConfig              `yaml:"pubsub"`
	Cache             RedisConfig              `yaml:"cache"`
	Notifications     NotificationsConfig      `yaml:"notifications"`
	DefaultUser       *User                    `yaml:"default_user"`
	DefaultApp        *App                     `yaml:"default_app"`
	PasswordConfig    *HashParams              `yaml:"password_config"`
	Session           SessionConfig            `yaml:"session"`
	VerifyTTL         time.Duration            `yaml:"verify_ttl"`
	PasswordForgotTTL time.Duration            `yaml:"password_forgot_ttl"`
	InviteTTL         time.Duration            `yaml:"invite_ttl"`
	AuthCodeTTL       time.Duration            `yaml:"auth_code_ttl"`
	AccessToken       AccessTokenConfiguration `yaml:"access_token"`
	Email             EmailParams              `yaml:"email"`
}

type AccessTokenConfiguration struct {
	PublicKeyPath  string        `yaml:"public_key_path"`
	PrivateKeyPath string        `yaml:"private_key_path"`
	TTL            time.Duration `yaml:"ttl"`
}

func LoadAuthConfig(filename string) (AuthConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return AuthConfig{}, err
	}
	defer file.Close()

	var config AuthConfig
	decoder := yaml.NewDecoder(file)

	err = decoder.Decode(&config)
	if err != nil {
		return AuthConfig{}, err
	}

	return config, nil
}
