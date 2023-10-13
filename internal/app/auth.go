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

	db := database.NewMySQLDbProvider(cfg.Database.Path)
	kv := database.NewRedisProvider("AUTH:", cfg.Cache.Addr, cfg.Cache.Password)

	templateRepository := repositories.
		NewTemplateRepository(cfg.Template.Common...).
		AddTemplates(cfg.Template.Paths...)

	sessionStore := session.NewRedisSessionStore(
		kv,
		cfg.Session.TTL,
		cfg.Session.SigningKey,
	)
	verifyTokenRepository := repositories.NewHashedVerifyTokenRepository(
		kv,
		cfg.VerifyTTL,
		repositories.VerificationTypeRegistration,
		cfg.PasswordConfig,
	)
	forgotPasswordTokenRepository := repositories.NewHashedVerifyTokenRepository(
		kv,
		cfg.PasswordForgotTTL,
		repositories.VerificationTypeForgotPassword,
		cfg.PasswordConfig,
	)
	inviteService := repositories.NewHashedVerifyTokenRepository(
		kv,
		cfg.InviteTTL,
		repositories.VerificationTypeInvite,
		cfg.PasswordConfig,
	)
	authCodeService := repositories.NewHashedVerifyTokenRepository(
		kv,
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

	publisher := database.NewRedisPublisherProvider(cfg.PubSub.Addr, cfg.PubSub.Password)
	permissionModel := config.LoadRbacModel("configs/rbac_model.conf")
	accessControlService := rbac.NewCasbinAccessControl(
		permissionModel,
		kv,
		publisher,
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

	orgDao := dao.NewOrganizationDao(db)
	orgRepo := repositories.NewOrganizationRepository(orgDao, accessControlService)

	return &Auth{
		server: transport.NewServer(
			cfg.Server,
			routes.NewAuthRoutes(
				cfg.Notifications,
				cfg.Session,
				authRepo,
				sessionStore,
				templateRepository,
				accessControlService,
				emailService,
			),
			routes.NewOauthRoutes(
				appService,
				sessionStore,
				templateRepository,
				cfg.Notifications,
			),
			routes.NewUserRoutes(
				cfg.Notifications,
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
			routes.NewOrganizationRoutes(
				cfg.Notifications,
				sessionStore,
				templateRepository,
				orgRepo,
			),
		),
		cleanup: func(_ context.Context) {
			db.Close()
			// kv.Close()
			// publisher.Close()
		},
		setup: func(ctx context.Context) {
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

			if cfg.DefaultApp != nil {
				// Pretend to log in as root to do this action
				newContext := context.WithValue(ctx, "user_id", "ROOT")

				_, err := appService.GetAppByClientId(newContext, cfg.DefaultApp.ClientId)
				if err == services.AppNotFound {
					_, err = appService.CreateApp(
						newContext,
						cfg.DefaultApp.ClientId,
						cfg.DefaultApp.ClientSecret,
						cfg.DefaultApp.RedirectUri,
						cfg.DefaultApp.Name,
						cfg.DefaultApp.Description,
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
