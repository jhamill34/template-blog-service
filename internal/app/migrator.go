package app

import (
	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/database"
)

type Migrator struct {
	databaseProvider database.DatabaseProvider
	migrations       []string
}

func ConfigureMigrator() *Migrator {
	cfg, err := config.LoadMigrationConfig("configs/migrator.yaml")
	if err != nil {
		panic(err)
	}

	db := database.NewMySQLDbProvider(cfg.Database.Path)

	return &Migrator{db, cfg.Migrations}
}

func (m *Migrator) Run() {
	err := database.Migrate(m.databaseProvider, "ROOT", m.migrations)
	if err != nil {
		panic(err)
	}
}
