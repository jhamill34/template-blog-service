package auth

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strings"

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

func Configure() *Auth {
	cfg, err := config.LoadAuthConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		panic(err)
	}

	privateKey := loadPrivateKey(cfg.AccessToken.PrivateKeyPath.String())
	publicKey := loadPublicKey(cfg.AccessToken.PublicKeyPath.String())
	signer := rca_signer.NewRcaSigner(rca_signer.NewStaticPublicKeyProvider(publicKey), privateKey)

	db := database.NewMySQLDbProvider(cfg.Database.GetConnectionString())
	kv := database.NewRedisProvider("AUTH:", cfg.Cache.Addr.String(), cfg.Cache.Password.String())

	templateRepository := repositories.
		NewTemplateRepository(cfg.Template.Common...).
		AddTemplates(cfg.Template.Paths...)

	sessionStore := session.NewRedisSessionStore(
		kv,
		cfg.Session.TTL,
		[]byte(cfg.Session.SigningKey.String()),
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
	inviteToOrgTokenService := repositories.NewHashedVerifyTokenRepository(
		kv,
		cfg.InviteTTL,
		repositories.VerificationTypeInviteToOrg,
		cfg.PasswordConfig,
	)

	smtpAddr := fmt.Sprintf("%s:%d", cfg.Email.SmtpDomain, cfg.Email.SmtpPort)

	smtpCredentials := strings.Split(cfg.Email.SmtpCredentials.String(), ":")
	if len(smtpCredentials) != 2 {
		panic("Invalid email credentials in configuration")
	}

	emailService := email.NewSmtpSender(
		smtpAddr,
		smtpCredentials[0],
		smtpCredentials[1],
		cfg.Email.User.String(),
		cfg.Email.Domain.String(),
		&tls.Config{
			ServerName: cfg.Email.SmtpDomain.String(),
			InsecureSkipVerify: true,
		},
	)

	userDao := dao.NewUserDao(db)
	authRepo := repositories.NewAuthRepository(
		cfg.Server.BaseUrl.String(),
		userDao,
		cfg.PasswordConfig,
		verifyTokenRepository,
		emailService,
		templateRepository,
		forgotPasswordTokenRepository,
		inviteService,
	)

	appDao := dao.NewApplicationDao(db)
	orgDao := dao.NewOrganizationDao(db)

	publisher := database.NewRedisPublisherProvider(
		cfg.PubSub.Addr.String(),
		cfg.PubSub.Password.String(),
	)
	permissionModel := config.LoadRbacModel(os.Getenv("RBAC_MODEL_FILE"))
	policyProvider := rbac.NewDatabasePolicyProvider(userDao, orgDao)
	accessControlService := rbac.NewCasbinAccessControl(
		permissionModel,
		kv,
		publisher,
		policyProvider,
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

	orgRepo := repositories.NewOrganizationRepository(
		cfg.Server.BaseUrl.String(),
		orgDao,
		userDao,
		accessControlService,
		inviteToOrgTokenService,
		emailService,
		templateRepository,
	)

	return &Auth{
		server: transport.NewServer(
			cfg.Server,
			routes.NewAuthRoutes(
				cfg.Server.BaseUrl.String(),
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
				accessControlService,
				policyProvider,
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
						cfg.DefaultUser.Email.String(),
						cfg.DefaultUser.Password.String(),
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

				_, err := appService.GetAppByClientId(newContext, cfg.DefaultApp.ClientId.String())
				if err == services.AppNotFound {
					_, err = appService.CreateApp(
						newContext,
						cfg.DefaultApp.ClientId.String(),
						cfg.DefaultApp.ClientSecret.String(),
						cfg.DefaultApp.RedirectUri.String(),
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
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	if publicKey, ok := publicKey.(*rsa.PublicKey); ok {
		return publicKey
	}

	panic("Could not load public key")
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
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	if privateKey, ok := privateKey.(*rsa.PrivateKey); ok {
		return privateKey
	}

	panic("Not an RSA private key")
}
