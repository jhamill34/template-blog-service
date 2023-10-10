package app

import (
	"context"
	"net/http"

	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/services/rbac"
	"github.com/jhamill34/notion-provisioner/internal/services/rca_signer"
	"github.com/jhamill34/notion-provisioner/internal/services/repositories"
	"github.com/jhamill34/notion-provisioner/internal/transport"
	"github.com/jhamill34/notion-provisioner/internal/transport/routes"
	"github.com/redis/go-redis/v9"
)

type App struct {
	server  transport.Server
	setup   func(ctx context.Context)
	cleanup func(ctx context.Context)
}

func ConfigureApp() *App {
	cfg, err := config.LoadAppConfig("configs/app.yaml")
	if err != nil {
		panic(err)
	}

	publicKeyProvider := rca_signer.NewRemotePublicKeyProvider(
		http.DefaultClient,
		cfg.AuthServer+"/key/signer",
	)
	signer := rca_signer.NewRcaSigner(publicKeyProvider, nil)

	db := database.NewMySQLDbProvider(cfg.Database.Path)
	policyStore := redis.NewClient(&redis.Options{
		Addr:     cfg.Cache.Addr,
		Password: cfg.Cache.Password,
	})

	permissionModel := config.LoadRbacModel("configs/rbac_model.conf")
	accessControlService := rbac.NewCasbinAccessControl(
		permissionModel,
		policyStore,
		nil,
		rbac.NewRemotePolicyProvider("http://auth-service:3334/policy", http.DefaultClient),
	)

	subscriber := redis.NewClient(&redis.Options{
		Addr:     cfg.PubSub.Addr,
		Password: cfg.PubSub.Password,
	})

	postDao := dao.NewPostDao(db)
	postService := repositories.NewPostRepository(postDao, accessControlService)

	go listenForPolicyInvalidation(context.Background(), subscriber, policyStore)

	return &App{
		server: transport.NewServer(
			cfg.Server,
			routes.NewBlogRoutes(
				postService,
				signer,
			),
		),
		cleanup: func(_ context.Context) {
			db.Close()
		},
		setup: func(ctx context.Context) {
			err := database.Migrate(db, cfg.Database.Migrations)
			if err != nil {
				panic(err)
			}
		},
	}
}

func (a *App) Start(ctx context.Context) {
	a.setup(ctx)
	defer a.cleanup(ctx)

	a.server.Start(ctx)
}

func listenForPolicyInvalidation(
	ctx context.Context,
	redisSubscriber *redis.Client,
	policyStore *redis.Client,
) {
	subscriber := redisSubscriber.Subscribe(ctx, "policy_invalidate")
	for {
		msg, err := subscriber.ReceiveMessage(ctx)
		if err != nil {
			panic(err)
		}

		policyStore.Del(ctx, msg.Payload)
	}
}
