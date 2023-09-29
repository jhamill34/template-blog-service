package dao

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jhamill34/notion-provisioner/internal/database"
)

type UserDao struct {
	databaseProvider database.DatabaseProvider
}

func NewUserDao(databaseProvider database.DatabaseProvider) *UserDao {
	return &UserDao{
		databaseProvider: databaseProvider,
	}
}

func (dao *UserDao) FindById(ctx context.Context, id string) (*database.UserEntity, error) {
	db := dao.databaseProvider.Get()

	var user database.UserEntity
	err := db.GetContext(ctx, &user, `
		SELECT 
			id, name, email, hashed_password, verified, created_at, updated_at 
		FROM 
			user 
		WHERE 
			id = ?
	`, id)

	if err == sql.ErrNoRows {
		return nil, database.NotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (dao *UserDao) FindByEmail(ctx context.Context, email string) (*database.UserEntity, error) {
	db := dao.databaseProvider.Get()

	var user database.UserEntity
	err := db.GetContext(ctx, &user, `
		SELECT 
			id, name, email, hashed_password, verified, created_at, updated_at 
		FROM 
			user 
		WHERE 
			email = ?
	`, email)

	if err == sql.ErrNoRows {
		return nil, database.NotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (dao *UserDao) FindByUsername(
	ctx context.Context,
	username string,
) (*database.UserEntity, error) {
	db := dao.databaseProvider.Get()

	var user database.UserEntity
	err := db.GetContext(ctx, &user, `
		SELECT 
			id, name, email, hashed_password, verified, created_at, updated_at 
		FROM 
			user 
		WHERE 
			name = ?
	`, username)

	if err == sql.ErrNoRows {
		return nil, database.NotFound
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (dao *UserDao) CreateUser(
	ctx context.Context,
	name, email, hashedPassword string,
	verified bool,
) error {
	db := dao.databaseProvider.Get()

	if _, err := dao.FindByEmail(ctx, email); err == database.NotFound {
		id := uuid.New().String()
		_, err := db.ExecContext(ctx, `
		INSERT INTO user 
			(id, name, email, hashed_password, verified)
		VALUES 
			(?, ?, ?, ?, ?)
	`, id, name, email, hashedPassword, verified)

		if err != nil {
			return err
		}

		return nil
	} else {
		return database.Duplicate
	}
}

func (dao *UserDao) ChangePassword(ctx context.Context, id, hashedPassword string) error {
	db := dao.databaseProvider.Get()

	_, err := db.ExecContext(ctx, `
		UPDATE user 
		SET hashed_password = ?
		WHERE id = ?
	`, hashedPassword, id)

	if err != nil {
		return err
	}

	return nil
}
