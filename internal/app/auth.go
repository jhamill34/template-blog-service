package app

import (
	"context"

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
	server  transport.Server
	setup   func()
	cleanup func()
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

	sessionStore := session.NewInMemorySessionStore()

	userDao := dao.NewUserDao(db)
	authRepo := repositories.NewAuthRepository(userDao, cfg.PasswordConfig)

	return &Auth{
		server: transport.NewServer(
			cfg.General.Server.Port,
			routes.NewAuthRoutes(
				authRepo,
				sessionStore,
				templateRepository,
			),
			routes.NewOauthRoutes(
				sessionStore,
				templateRepository,
			),
		),
		cleanup: func() {
			db.Close()
		},
		setup: func() {
			err := database.Migrate(db, cfg.General.Database.Migrations)
			if err != nil {
				panic(err)
			}

			if cfg.DefaultUser != nil {
				user, err := authRepo.GetUserByUsername("ROOT")
				if err != nil {
					panic(err)
				}

				if user == nil {
					err = authRepo.CreateRootUser(
						cfg.DefaultUser.Email,
						cfg.DefaultUser.Password,
					)
					if err != nil {
						panic(err)
					}
				}
			}

		},
	}
}

func (a *Auth) Start(ctx context.Context) {
	a.setup()
	defer a.cleanup()

	a.server.Start(ctx)
}
