package config

import "time"

type Configuration struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Template TemplateConfig `yaml:"template"`
	Redis    RedisConfig   `yaml:"redis"`
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

type RedisConfig struct {
	Addr     string `yaml:"address"`
	Password string `yaml:"password"`
}

type TemplateConfig struct {
	Common []string `yaml:"common"`
	Paths  []string `yaml:"paths"`
}
