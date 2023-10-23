package api

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
	cleanup func(ctx context.Context)
}

func Configure() *App {
	cfg, err := config.LoadAppConfig("configs/app.yaml")
	if err != nil {
		panic(err)
	}

	publicKeyProvider := rca_signer.NewRemotePublicKeyProvider(
		http.DefaultClient,
		cfg.AuthServer.BaseUrl.String()+cfg.AuthServer.KeyPath,
	)
	signer := rca_signer.NewRcaSigner(publicKeyProvider, nil)

	db := database.NewMySQLDbProvider(cfg.Database.GetConnectionString())

	kv := database.NewRedisProvider("APP:", cfg.Cache.Addr.String(), cfg.Cache.Password.String())

	permissionModel := config.LoadRbacModel("configs/rbac_model.conf")
	accessControlService := rbac.NewCasbinAccessControl(
		permissionModel,
		kv,
		nil,
		rbac.NewRemotePolicyProvider(
			cfg.AuthServer.BaseUrl.String()+cfg.AuthServer.PolicyPath,
			http.DefaultClient,
		),
	)

	postDao := dao.NewPostDao(db)
	postService := repositories.NewPostRepository(postDao, accessControlService)

	subscriber := database.NewRedisSubscriberProvider(
		cfg.PubSub.Addr.String(),
		cfg.PubSub.Password.String(),
	)

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
	}
}

func (a *App) Start(ctx context.Context) {
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
