package dao

import (
	"context"
	"database/sql"

	"github.com/jhamill34/notion-provisioner/internal/database"
)

type ApplicationDao struct {
	databaseProvider database.DatabaseProvider
}

func NewApplicationDao(
	databaseProvider database.DatabaseProvider,
) *ApplicationDao {
	return &ApplicationDao{
		databaseProvider: databaseProvider,
	}
}

func (self *ApplicationDao) Create(
	ctx context.Context,
	id, clientId, hashedClientSecret,
	redirect_uri, name, description string,
) (*database.ApplicationEntity, error) {
	db := self.databaseProvider.Get()

	if _, err := self.FindByClientId(ctx, clientId); err == database.NotFound {
		_, err := db.ExecContext(ctx, `
		INSERT INTO application (id, client_id, hashed_client_secret, redirect_uri, name, description)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, clientId, hashedClientSecret, redirect_uri, name, description)

		if err != nil {
			return nil, err
		}

		return &database.ApplicationEntity{
			Id:                 id,
			ClientId:           clientId,
			HashedClientSecret: hashedClientSecret,
			RedirectUri:        redirect_uri,
			Name:               name,
			Description:        description,
		}, nil

	} else {
		return nil, database.Duplicate
	}
}

func (self *ApplicationDao) FindById(
	ctx context.Context,
	id string,
) (*database.ApplicationEntity, error) {
	db := self.databaseProvider.Get()

	var app database.ApplicationEntity
	err := db.GetContext(ctx, &app, `
		SELECT 
			id, client_id, hashed_client_secret, redirect_uri, name, description, created_at, updated_at
		FROM application
		WHERE id = ?
	`, id)

	if err == sql.ErrNoRows {
		return nil, database.NotFound
	}

	if err != nil {
		return nil, err
	}

	return &app, nil
}

func (self *ApplicationDao) FindByClientId(
	ctx context.Context,
	clientId string,
) (*database.ApplicationEntity, error) {
	db := self.databaseProvider.Get()

	var app database.ApplicationEntity
	err := db.GetContext(ctx, &app, `
		SELECT 
			id, client_id, hashed_client_secret, redirect_uri, name, description, created_at, updated_at
		FROM application
		WHERE client_id = ?
	`, clientId)

	if err == sql.ErrNoRows {
		return nil, database.NotFound
	}

	if err != nil {
		return nil, err
	}

	return &app, nil
}

func (self *ApplicationDao) UpdateSecret(ctx context.Context, appId, hashedSecret string) error {
	db := self.databaseProvider.Get()

	_, err := db.ExecContext(ctx, `
		UPDATE application
		SET hashed_client_secret = ?
		WHERE id = ?
	`, hashedSecret, appId)

	if err != nil {
		return err
	}

	return nil
}

func (self *ApplicationDao) Delete(ctx context.Context, appId string) error {
	db := self.databaseProvider.Get()

	_, err := db.ExecContext(ctx, `
		DELETE FROM application
		WHERE id = ?
	`, appId)

	if err != nil {
		return err
	}

	return nil
}

func (self *ApplicationDao) List(ctx context.Context) ([]database.ApplicationEntity, error) {
	db := self.databaseProvider.Get()

	var apps []database.ApplicationEntity
	err := db.SelectContext(ctx, &apps, `
		SELECT 
			id, client_id, hashed_client_secret, redirect_uri, name, description, created_at, updated_at
		FROM application
	`)

	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (self *ApplicationDao) CreateRefreshToken(
	ctx context.Context,
	userId, appId, token string,
) error {
	db := self.databaseProvider.Get()

	_, err := db.ExecContext(ctx, `
		INSERT INTO refresh_token (user_id, app_id, token)
		VALUES (?, ?, ?)
	`, userId, appId, token)

	if err != nil {
		return err
	}

	return nil
}

func (self *ApplicationDao) FindRefreshToken(
	ctx context.Context,
	token string,
) (*database.RefreshTokenEntity, error) {
	db := self.databaseProvider.Get()

	var refreshToken database.RefreshTokenEntity
	err := db.GetContext(ctx, &refreshToken, `
		SELECT 
			id, user_id, app_id, token
		FROM refresh_token
		WHERE token = ?
	`, token)

	if err == sql.ErrNoRows {
		return nil, database.NotFound
	}

	if err != nil {
		return nil, err
	}

	return &refreshToken, nil
}

func (self *ApplicationDao) DeleteRefreshToken(
	ctx context.Context,
	token string,
) error {
	db := self.databaseProvider.Get()

	_, err := db.ExecContext(ctx, `
		DELETE FROM refresh_token
		WHERE token = ?
	`, token)

	if err != nil {
		return err
	}

	return nil
}
