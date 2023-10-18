package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type MailerConfig struct {
	Port      int             `yaml:"port"`
	Protocol  string          `yaml:"protocol"`
	Dkim      DkimConfig      `yaml:"dkim"`
	Forwarder ForwarderConfig `yaml:"forwarder"`
}

type ForwarderConfig struct {
	CommonPorts []int `yaml:"common_ports"`
}

type DkimConfig struct {
	Selector       string   `yaml:"selector"`
	Domain         string   `yaml:"domain"`
	Headers        []string `yaml:"headers"`
	PrivateKeyPath string   `yaml:"private_key_path"`
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
