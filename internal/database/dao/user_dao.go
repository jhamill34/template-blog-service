package dao

import (
	"log"

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

func (dao *UserDao) FindByEmail(email string) (*database.UserEntity, error) {
	db := dao.databaseProvider.Get()

	var user database.UserEntity
	err := db.Get(&user, `
		SELECT 
			id, name, email, hashed_password, created_at, updated_at 
		FROM 
			user 
		WHERE 
			email = ?
	`, email)

	if err != nil {
		log.Printf("Error scanning user: %v", err)
		return nil, err
	}

	return &user, nil
}

func (dao *UserDao) CreateUser(name, email, hashedPassword string) error {
	db := dao.databaseProvider.Get()

	id := uuid.New().String()

	_, err := db.Exec(`
		INSERT INTO user 
			(id, name, email, hashed_password)
		VALUES 
			(?, ?, ?, ?)
	`, id, name, email, hashedPassword)

	if err != nil {
		return err
	}

	return nil
}
