package config

import "time"

type Configuration struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Template TemplateConfig `yaml:"template"`
}

type ServerConfig struct {
	Port         int `yaml:"port"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type DatabaseConfig struct {
	Path       string   `yaml:"path"`
	Migrations []string `yaml:"migrations"`
}

type TemplateConfig struct {
	Common []string `yaml:"common"`
	Paths  []string `yaml:"paths"`
}
