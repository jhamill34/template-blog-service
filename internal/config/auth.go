package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type User struct {
	Password StringFromEnv `yaml:"password"`
	Email    StringFromEnv `yaml:"email"`
}

type App struct {
	ClientId     StringFromFile `yaml:"client_id"`
	ClientSecret StringFromFile `yaml:"client_secret"`
	Name         string         `yaml:"name"`
	Description  string         `yaml:"description"`
	RedirectUri  StringFromEnv  `yaml:"redirect_uri"`
}

type HashParams struct {
	Iterations  uint32        `yaml:"iterations"`
	Parallelism uint8         `yaml:"parallelism"`
	Memory      uint32        `yaml:"memory"`
	HashLength  uint32        `yaml:"hash_length"`
	SaltLength  uint32        `yaml:"salt_length"`
	Secret      StringFromEnv `yaml:"secret"`
}

type EmailParams struct {
	Domain          StringFromEnv  `yaml:"domain"`
	User            StringFromEnv  `yaml:"user"`
	SmtpDomain      StringFromEnv  `yaml:"smtp_domain"`
	SmtpCredentials StringFromFile `yaml:"smtp_credentials"`
	SmtpPort        int            `yaml:"smtp_port"`
}

type AccessTokenConfiguration struct {
	PublicKeyPath  StringFromEnv `yaml:"public_key_path"`
	PrivateKeyPath StringFromEnv `yaml:"private_key_path"`
	TTL            time.Duration `yaml:"ttl"`
}

type AuthConfig struct {
	Server            ServerConfig             `yaml:"server"`
	Cache             RedisConfig              `yaml:"cache"`
	PubSub            RedisConfig              `yaml:"pubsub"`
	Database          DatabaseConfig           `yaml:"database"`
	Notifications     NotificationsConfig      `yaml:"notifications"`
	Template          TemplateConfig           `yaml:"template"`
	DefaultUser       *User                    `yaml:"default_user"`
	DefaultApp        *App                     `yaml:"default_app"`
	PasswordConfig    *HashParams              `yaml:"password_config"`
	VerifyTTL         time.Duration            `yaml:"verify_ttl"`
	PasswordForgotTTL time.Duration            `yaml:"password_forgot_ttl"`
	InviteTTL         time.Duration            `yaml:"invite_ttl"`
	AuthCodeTTL       time.Duration            `yaml:"auth_code_ttl"`
	AccessToken       AccessTokenConfiguration `yaml:"access_token"`
	Session           SessionConfig            `yaml:"session"`
	Email             EmailParams              `yaml:"email"`
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
