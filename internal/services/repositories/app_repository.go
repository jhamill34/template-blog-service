package repositories

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

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
	tokenClaimService    services.TokenClaimsService
	signer               services.Signer
}

func NewApplicationRepository(
	appDao *dao.ApplicationDao,
	accessControlService services.AccessControlService,
	passwordConfig *config.HashParams,
	tokenClaimService services.TokenClaimsService,
	signer services.Signer,
) *ApplicationRepository {
	return &ApplicationRepository{
		appDao:               appDao,
		accessControlService: accessControlService,
		passwordConfig:       passwordConfig,
		tokenClaimService:    tokenClaimService,
		signer:               signer,
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

// NewSecret implements services.ApplicationService.
func (self *ApplicationRepository) NewSecret(
	ctx context.Context,
	id string,
) (string, models.Notifier) {
	if err := self.accessControlService.Enforce(ctx, "/oauth/application/"+id+"/secret", "update"); err != nil {
		return "", err
	}

	clientSecret := uuid.New().String()
	hashedClientSecret, err := createHash(self.passwordConfig, clientSecret)
	if err != nil {
		panic(err)
	}

	if err := self.appDao.UpdateSecret(ctx, id, hashedClientSecret); err != nil {
		panic(err)
	}

	return clientSecret, nil
}

type AuthCodeClaims struct {
	UserId string `json:"UserId"`
	AppId  string `json:"AppId"`
}

func (self *ApplicationRepository) NewAuthCode(
	ctx context.Context,
	userId, appId string,
) string {
	codeBytes, err := randomBytes(32)
	if err != nil {
		panic(err)
	}
	code := base64.RawURLEncoding.EncodeToString(codeBytes)

	token := self.tokenClaimService.CreateWithClaims(ctx, code, AuthCodeClaims{
		UserId: userId,
		AppId:  appId,
	})

	return code + "/" + token
}

func (self *ApplicationRepository) GetAuthCode(
	ctx context.Context,
	code string,
) (string, string, models.Notifier) {
	parts := strings.Split(code, "/")
	if len(parts) != 2 {
		return "", "", services.InvalidAuthCode
	}

	var authCodeClaims AuthCodeClaims
	err := self.tokenClaimService.VerifyWithClaims(ctx, parts[0], parts[1], &authCodeClaims)
	if err != nil {
		return "", "", err
	}

	return authCodeClaims.UserId, authCodeClaims.AppId, nil
}

func (self *ApplicationRepository) ValidateAppSecret(
	ctx context.Context,
	appId, clientSecret string,
) (*models.App, models.Notifier) {
	app, err := self.appDao.FindById(ctx, appId)
	if err == database.NotFound {
		return nil, services.AppNotFound
	}

	ok, err := comparePasswords(clientSecret, app.HashedClientSecret)
	if err != nil {
		panic(err)
	}

	if !ok {
		return nil, services.AccessDenied
	}

	return &models.App{
		AppId:       app.Id,
		ClientId:    app.ClientId,
		RedirectUri: app.RedirectUri,
		Name:        app.Name,
		Description: app.Description,
	}, nil
}

type AccessTokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type AccessTokenClaims struct {
	Sub string `json:"sub"`
	Aud string `json:"aud"`
	Iss string `json:"iss"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
}

func (self *ApplicationRepository) NewAccessToken(
	ctx context.Context,
	userId, clientId string,
) (*models.AccessTokenResponse, models.Notifier) {
	header := AccessTokenHeader{
		Alg: "RS256",
		Typ: "JWT",
	}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		panic(err)
	}
	headerString := base64.RawURLEncoding.EncodeToString(headerBytes)

	claims := AccessTokenClaims{
		Sub: userId,
		Aud: clientId,
		Iss: "auth_server",
		Exp: 3600,
		Iat: time.Now().Unix(),
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		panic(err)
	}
	claimsString := base64.RawURLEncoding.EncodeToString(claimsBytes)
	
	paylaod := headerString + "." + claimsString
	signature, err := self.signer.Sign([]byte(paylaod))


	return &models.AccessTokenResponse{
		AccessToken:  paylaod + "." + signature,
		RefreshToken: "",
		Expires:      3600,
	}, nil
}

func (self *ApplicationRepository) VerifyAccessToken(ctx context.Context, accessToken string) bool {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return false
	}

	signature := parts[2]
	paylaod := parts[0] + "." + parts[1]

	return self.signer.Verify([]byte(paylaod), signature) == nil
}

// var _ services.ApplicationService = (*ApplicationRepository)(nil)
