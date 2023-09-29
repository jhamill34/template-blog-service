package database

import "time"

type UserEntity struct {
	Id             string    `db:"id"`
	Name           string    `db:"name"`
	Email          string    `db:"email"`
	HashedPassword string    `db:"hashed_password"`
	Verified       bool      `db:"verified"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
