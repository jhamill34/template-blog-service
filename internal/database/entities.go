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

type UserPermissionEntity struct {
	Id         int    `db:"id"`
	UserId     string `db:"user_id"`
	Resource   string `db:"resource"`
	Action     string `db:"action"`
	Permission string `db:"permission"`
}

type RoleEntity struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type RolePermissionEntity struct {
	Id         int    `db:"id"`
	RoleId     int    `db:"role_id"`
	Resource   string `db:"resource"`
	Action     string `db:"action"`
	Permission string `db:"permission"`
}

type RoleUserEntity struct {
	Id     int    `db:"id"`
	RoleId int    `db:"role_id"`
	UserId string `db:"user_id"`
}
