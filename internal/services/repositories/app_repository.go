package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/config"
	"github.com/jhamill34/notion-provisioner/internal/database"
	"github.com/jhamill34/notion-provisioner/internal/database/dao"
	"github.com/jhamill34/notion-provisioner/internal/models"
	"github.com/jhamill34/notion-provisioner/internal/services"
)

type ApplicationRepository struct {
	appDao               *dao.ApplicationDao
	accessControlService services.AccessControlService
	passwordConfig       *config.HashParams
}

func NewApplicationRepository(
	appDao *dao.ApplicationDao,
	accessControlService services.AccessControlService,
	passwordConfig *config.HashParams,
) *ApplicationRepository {
	return &ApplicationRepository{
		appDao:               appDao,
		accessControlService: accessControlService,
		passwordConfig:       passwordConfig,
	}
}

// CreateApp implements services.ApplicationService.
func (self *ApplicationRepository) CreateApp(
	ctx context.Context,
	redirectUri, name, description string,
) (*models.App, string, models.Notifier) {
	if err := self.accessControlService.Enforce(ctx, "/oauth/application", "create"); err != nil {
		return nil, "", err
	}

	appId := uuid.New().String()
	clientId := uuid.New().String()
	clientSecret := uuid.New().String()

	hashedClientSecret, err := createHash(self.passwordConfig, clientSecret)
	if err != nil {
		panic(err)
	}

	app, err := self.appDao.Create(
		ctx,
		appId, clientId, hashedClientSecret,
		redirectUri, name, description,
	)

	if err != nil {
		panic(err)
	}

	return &models.App{
		AppId:       app.Id,
		ClientId:    app.ClientId,
		RedirectUri: app.RedirectUri,
		Name:        app.Name,
		Description: app.Description,
	}, clientSecret, nil
}

// DeleteApp implements services.ApplicationService.
func (self *ApplicationRepository) DeleteApp(ctx context.Context, id string) models.Notifier {
	if err := self.accessControlService.Enforce(ctx, "/oauth/application/"+id, "delete"); err != nil {
		return err
	}

	if err := self.appDao.Delete(ctx, id); err != nil {
		panic(err)
	}

	return nil
}

// GetApp implements services.ApplicationService.
func (self *ApplicationRepository) GetApp(
	ctx context.Context,
	id string,
) (*models.App, models.Notifier) {
	if err := self.accessControlService.Enforce(ctx, "/oauth/application/"+id, "read"); err != nil {
		return nil, err
	}

	app, err := self.appDao.FindById(ctx, id)
	if err == database.NotFound {
		return nil, services.AppNotFound
	}
	if err != nil {
		panic(err)
	}

	return &models.App{
		AppId:       app.Id,
		ClientId:    app.ClientId,
		RedirectUri: app.RedirectUri,
		Name:        app.Name,
		Description: app.Description,
	}, nil
}

// GetAppByClientId implements services.ApplicationService.
func (self *ApplicationRepository) GetAppByClientId(
	ctx context.Context,
	clientId string,
) (*models.App, models.Notifier) {
	app, err := self.appDao.FindByClientId(ctx, clientId)
	if err == database.NotFound {
		return nil, services.AppNotFound
	}
	if err != nil {
		panic(err)
	}

	if err := self.accessControlService.Enforce(ctx, "/oauth/application/"+app.Id, "read"); err != nil {
		return nil, err
	}

	return &models.App{
		AppId:       app.Id,
		ClientId:    app.ClientId,
		RedirectUri: app.RedirectUri,
		Name:        app.Name,
		Description: app.Description,
	}, nil
}

// ListApps implements services.ApplicationService.
func (self *ApplicationRepository) ListApps(ctx context.Context) ([]models.App, models.Notifier) {
	if err := self.accessControlService.Enforce(ctx, "/oauth/application", "list"); err != nil {
		return nil, err
	}

	data, err := self.appDao.List(ctx)
	if err != nil {
		panic(err)
	}

	apps := make([]models.App, len(data))	
	i := 0
	for _, app := range data {
		if err := self.accessControlService.Enforce(ctx, "/oauth/application/"+app.Id, "read"); err == nil {
			apps[i] = models.App{
				AppId:       app.Id,
				ClientId:    app.ClientId,
				RedirectUri: app.RedirectUri,
				Name:        app.Name,
				Description: app.Description,
			}

			i++
		}
	}

	return apps[:i], nil
}

// var _ services.ApplicationService = (*ApplicationRepository)(nil)
