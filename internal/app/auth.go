package app

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/services/repositories"
	"github.com/jhamill34/notion-provisioner/internal/services/session"
	"github.com/jhamill34/notion-provisioner/internal/transport"
	"github.com/jhamill34/notion-provisioner/internal/transport/routes"
	"github.com/redis/go-redis/v9"
)

type Auth struct {
	server  transport.Server
	setup   func(ctx context.Context)
	cleanup func(ctx context.Context)
}

func ConfigureAuth() *Auth {
	cfg, err := config.LoadAuthConfig("configs/auth.yaml")
	if err != nil {
		panic(err)
	}

	db := database.NewMySQLDbProvider(cfg.General.Database.Path)

	templateRepository := repositories.
		NewTemplateRepository(cfg.General.Template.Common...).
		AddTemplates(cfg.General.Template.Paths...)

	var sessionStore services.SessionService
	if cfg.General.Session != nil {
		redisClient := redis.NewClient(&redis.Options{
			Addr: cfg.General.Session.Addr,
			Password: cfg.General.Session.Password,
		})
		sessionStore = session.NewRedisSessionStore(redisClient, cfg.General.Session.TTL)
	} else {
		sessionStore = session.NewInMemorySessionStore()
	}

	userDao := dao.NewUserDao(db)
	authRepo := repositories.NewAuthRepository(userDao, cfg.PasswordConfig)

	return &Auth{
		server: transport.NewServer(
			cfg.General.Server,
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
		cleanup: func(_ context.Context) {
			db.Close()
		},
		setup: func(ctx context.Context) {
			err := database.Migrate(db, cfg.General.Database.Migrations)
			if err != nil {
				panic(err)
			}

			if cfg.DefaultUser != nil {
				user, err := authRepo.GetUserByUsername(ctx, "ROOT")
				if err != nil {
					panic(err)
				}

				if user == nil {
					err = authRepo.CreateRootUser(
						ctx,
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
	a.setup(ctx)
	defer a.cleanup(ctx)

	a.server.Start(ctx)
}
