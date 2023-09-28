package app

import (
	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/services/repositories"
	"github.com/jhamill34/notion-provisioner/internal/services/session"
	"github.com/jhamill34/notion-provisioner/internal/transport"
	"github.com/jhamill34/notion-provisioner/internal/transport/routes"
	_ "github.com/mattn/go-sqlite3"
)

type Auth struct {
	server transport.Server
	db     database.DatabaseProvider
	cfg    config.AuthConfig
}

func ConfigureAuth() *Auth {
	cfg, err := config.LoadAuthConfig("configs/auth.yaml")
	if err != nil {
		panic(err)
	}

	db := database.NewSqliteDbProvider(cfg.General.Database.Path)

	templateRepository := repositories.
		NewTemplateRepository(cfg.General.Template.Common...).
		AddTemplates(cfg.General.Template.Paths...)

	return &Auth{
		server: transport.NewServer(
			cfg.General.Server.Port,
			routes.NewAuthRoutes(
				repositories.NewAuthRepository(
					dao.NewUserDao(db),
				),
				session.NewInMemorySessionStore(),
				templateRepository,
			),
		),
		db:  db,
		cfg: cfg,
	}
}

func (a *Auth) Start() {
	defer a.db.Close()

	err := database.Migrate(a.db, a.cfg.General.Database.Migrations)
	if err != nil {
		panic(err)
	}

	a.server.Start()
}
