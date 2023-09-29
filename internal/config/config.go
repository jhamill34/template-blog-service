package config

import "time"

type Configuration struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Template TemplateConfig `yaml:"template"`
	Session  *SessionConfig `yaml:"session"`
}

type ServerConfig struct {
	Port         int           `yaml:"port"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type DatabaseConfig struct {
	Path       string   `yaml:"path"`
	Migrations []string `yaml:"migrations"`
}

type SessionConfig struct {
	Addr      string        `yaml:"address"`
	Password  string        `yaml:"password"`
	TTL       time.Duration `yaml:"ttl"`
	CookieTTL time.Duration `yaml:"cookie_ttl"`
}

type TemplateConfig struct {
	Common []string `yaml:"common"`
	Paths  []string `yaml:"paths"`
}
