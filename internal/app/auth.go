package app

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"

	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/services"
	"github.com/jhamill34/notion-provisioner/internal/services/email"
	"github.com/jhamill34/notion-provisioner/internal/services/rbac"
	"github.com/jhamill34/notion-provisioner/internal/services/rca_signer"
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

	privateKey := loadPrivateKey(cfg.AccessToken.PrivateKeyPath)
	publicKey := loadPublicKey(cfg.AccessToken.PublicKeyPath)
	signer := rca_signer.NewRcaSigner(rca_signer.NewStaticPublicKeyProvider(publicKey), privateKey)

	db := database.NewMySQLDbProvider(cfg.General.Database.Path)

	templateRepository := repositories.
		NewTemplateRepository(cfg.General.Template.Common...).
		AddTemplates(cfg.General.Template.Paths...)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.General.Cache.Addr,
		Password: cfg.General.Cache.Password,
	})

	sessionStore := session.NewRedisSessionStore(
		redisClient,
		cfg.Session.TTL,
		cfg.Session.SigningKey,
	)
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
	authCodeService := repositories.NewHashedVerifyTokenRepository(
		redisClient,
		cfg.AuthCodeTTL,
		repositories.VerificationTypeAuthCode,
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

	appDao := dao.NewApplicationDao(db)
	
	subscriber := redis.NewClient(&redis.Options{
		Addr:     cfg.General.PubSub.Addr,
		Password: cfg.General.PubSub.Password,
	})

	permissionModel := config.LoadRbacModel("configs/rbac_model.conf")
	accessControlService := rbac.NewCasbinAccessControl(
		permissionModel,
		redisClient,
		subscriber,
		rbac.NewDatabasePolicyProvider(userDao),
	)

	userService := repositories.NewUserRepository(userDao, accessControlService)
	appService := repositories.NewApplicationRepository(
		appDao,
		accessControlService,
		cfg.PasswordConfig,
		authCodeService,
		signer,
		cfg.AccessToken,
	)


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
				appService,
				sessionStore,
				templateRepository,
				cfg.General.Notifications,
			),
			routes.NewUserRoutes(
				cfg.General.Notifications,
				sessionStore,
				templateRepository,
				userService,
			),
			routes.NewKeyRoutes(
				&privateKey.PublicKey,
			),
			routes.NewPolicyRoutes(
				sessionStore,
				signer,
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

func loadPublicKey(path string) *rsa.PublicKey {
	publicFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer publicFile.Close()

	publicKeyBytes, err := io.ReadAll(publicFile)
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(publicKeyBytes)
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	return publicKey
}

func loadPrivateKey(path string) *rsa.PrivateKey {
	privateFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer privateFile.Close()

	privateKeyBytes, err := io.ReadAll(privateFile)
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(privateKeyBytes)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	return privateKey
}
