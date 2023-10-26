package config

import (
	"fmt"
	"time"
)

type NotificationsConfig struct {
	Timeout time.Duration `yaml:"timeout"`
}

type ServerConfig struct {
	Port         int           `yaml:"port"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	BaseUrl      StringFromEnv `yaml:"base_url"`
}

type DatabaseConfig struct {
	User   StringFromEnv  `yaml:"user"`
	Pass   StringFromFile `yaml:"pass"`
	Host   StringFromEnv  `yaml:"host"`
	DbName StringFromEnv  `yaml:"db_name"`
}

func (c DatabaseConfig) GetConnectionString() string {
	conn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?parseTime=true&multiStatements=true",
		c.User,
		c.Pass,
		c.Host,
		c.DbName,
	)

	return conn
}

type RedisConfig struct {
	Addr     StringFromEnv  `yaml:"address"`
	Password StringFromFile `yaml:"password"`
}

type TemplateConfig struct {
	Common []string `yaml:"common"`
	Paths  []string `yaml:"paths"`
}

type SessionConfig struct {
	TTL        time.Duration `yaml:"ttl"`
	CookieTTL  time.Duration `yaml:"cookie_ttl"`
	SigningKey StringFromEnv `yaml:"signing_key"`
}
