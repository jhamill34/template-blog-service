package migrator

import (
	"os"

	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/database"
)

type Migrator struct {
	databaseProvider database.DatabaseProvider
	migrations       []string
}

func Configure() *Migrator {
	cfg, err := config.LoadMigrationConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		panic(err)
	}

	db := database.NewMySQLDbProvider(cfg.Database.GetConnectionString())

	return &Migrator{db, cfg.Migrations}
}

func (m *Migrator) Run() {
	err := database.Migrate(m.databaseProvider, "ROOT", m.migrations)
	if err != nil {
		panic(err)
	}
}
