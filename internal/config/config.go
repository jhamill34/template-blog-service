package config

import "time"

type NotificationsConfig struct {
	Timeout time.Duration `yaml:"timeout"`
}

type ServerConfig struct {
	Port         int           `yaml:"port"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	BaseUrl      string        `yaml:"base_url"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type RedisConfig struct {
	Addr     string `yaml:"address"`
	Password string `yaml:"password"`
}

type TemplateConfig struct {
	Common []string `yaml:"common"`
	Paths  []string `yaml:"paths"`
}

type SessionConfig struct {
	TTL        time.Duration `yaml:"ttl"`
	CookieTTL  time.Duration `yaml:"cookie_ttl"`
	SigningKey []byte        `yaml:"signing_key"`
}
