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

	kv := database.NewRedisProvider("APP:", cfg.Cache.Addr, cfg.Cache.Password)

	permissionModel := config.LoadRbacModel("configs/rbac_model.conf")
	accessControlService := rbac.NewCasbinAccessControl(
		permissionModel,
		kv,
		nil,
		rbac.NewRemotePolicyProvider(cfg.AuthServer+"/policy", http.DefaultClient),
	)

	postDao := dao.NewPostDao(db)
	postService := repositories.NewPostRepository(postDao, accessControlService)

	subscriber := database.NewRedisSubscriberProvider(cfg.PubSub.Addr, cfg.PubSub.Password)

	go listenForPolicyInvalidation(context.Background(), subscriber, kv)

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
			// kv.Close()
			// subscriber.Close()
		},
		setup: func(ctx context.Context) {
			err := database.Migrate(db, "APP", cfg.Database.Migrations)
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
	subscriber database.SubscriberProvider,
	policyStore database.KeyValueStoreProvider,
) {
	channel, err := subscriber.Get().Subscribe(ctx, "policy_invalidate")
	if err != nil {
		panic(err)
	}

	for {
		msg := <-channel
		policyStore.Get().Del(ctx, msg)
	}
}
