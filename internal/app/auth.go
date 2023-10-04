package app

import (
	"context"

	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/services/email"
	"github.com/jhamill34/notion-provisioner/internal/services/rbac"
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

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.General.Redis.Addr,
		Password: cfg.General.Redis.Password,
	})

	sessionStore := session.NewRedisSessionStore(redisClient, cfg.Session.TTL, cfg.Session.SigningKey)
	verifyTokenRepository := repositories.NewHashedVerifyTokenRepository(
		redisClient,
		cfg.VerifyTTL,
		repositories.VerificationTypeRegistration,
		cfg.PasswordConfig,
	)
	forgotPasswordTokenRepository := repositories.NewHashedVerifyTokenRepository(
		redisClient,
		cfg.PasswordForgotTTL,
		repositories.VerificationTypeForgotPassword,
		cfg.PasswordConfig,
	)
	inviteService := repositories.NewHashedVerifyTokenRepository(
		redisClient,
		cfg.InviteTTL,
		repositories.VerificationTypeInvite,
		cfg.PasswordConfig,
	)

	emailService := &email.MockEmailService{}

	userDao := dao.NewUserDao(db)
	authRepo := repositories.NewAuthRepository(
		userDao,
		cfg.PasswordConfig,
		verifyTokenRepository,
		emailService,
		templateRepository,
		forgotPasswordTokenRepository,
		inviteService,
	)

	permissionModel := config.LoadRbacModel("configs/rbac_model.conf")
	accessControlService := rbac.NewCasbinAccessControl(permissionModel, userDao)

	userService := repositories.NewUserRepository(userDao, accessControlService)

	return &Auth{
		server: transport.NewServer(
			cfg.General.Server,
			routes.NewAuthRoutes(
				cfg.General.Notifications,
				cfg.Session,
				authRepo,
				sessionStore,
				templateRepository,
				accessControlService,
			),
			routes.NewOauthRoutes(
				sessionStore,
				templateRepository,
			),
			routes.NewUserRoutes(
				sessionStore,
				templateRepository,
				userService,
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
				_, err := authRepo.GetUserByUsername(ctx, "ROOT")
				if err == services.AccountNotFound {
					err = authRepo.CreateRootUser(
						ctx,
						cfg.DefaultUser.Email,
						cfg.DefaultUser.Password,
					)
					if err != nil {
						panic(err)
					}
				}
				
				if err != nil {
					panic(err)
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
