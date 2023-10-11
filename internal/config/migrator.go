package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type MigratorConfig struct {
	Database   DatabaseConfig `yaml:"database"`
	Migrations []string       `yaml:"migrations"`
}

func LoadMigrationConfig(filename string) (MigratorConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return MigratorConfig{}, err
	}
	defer file.Close()

	var config MigratorConfig
	decoder := yaml.NewDecoder(file)

	err = decoder.Decode(&config)
	if err != nil {
		return MigratorConfig{}, err
	}

	return config, nil
}
