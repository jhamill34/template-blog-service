package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type MailerConfig struct {
	Port           int             `yaml:"port"`
	Protocol       string          `yaml:"protocol"`
	ReadTimeout    time.Duration   `yaml:"read_timeout"`
	WriteTimeout   time.Duration   `yaml:"write_timeout"`
	DataTimeout    time.Duration   `yaml:"data_timeout"`
	MaxMessageSize int             `yaml:"max_message_size"`
	MaxRecipients  int             `yaml:"max_recipients"`
	Dkim           DkimConfig      `yaml:"dkim"`
	Forwarder      ForwarderConfig `yaml:"forwarder"`
}

type ForwarderConfig struct {
	CommonPorts []int `yaml:"common_ports"`
}

type DkimConfig struct {
	Selector       string        `yaml:"selector"`
	Domain         StringFromEnv `yaml:"domain"`
	Headers        []string      `yaml:"headers"`
	PrivateKeyPath StringFromEnv `yaml:"private_key_path"`
}

func LoadMailConfig(filename string) (MailerConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return MailerConfig{}, err
	}
	defer file.Close()

	var config MailerConfig
	decoder := yaml.NewDecoder(file)

	err = decoder.Decode(&config)
	if err != nil {
		return MailerConfig{}, err
	}

	return config, nil
}
