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
	Id       int    `db:"id"`
	UserId   string `db:"user_id"`
	Resource string `db:"resource"`
	Action   string `db:"action"`
	Effect   string `db:"effect"`
}

type ApplicationEntity struct {
	Id                 string    `db:"id"`
	ClientId           string    `db:"client_id"`
	HashedClientSecret string    `db:"hashed_client_secret"`
	RedirectUri        string    `db:"redirect_uri"`
	Name               string    `db:"name"`
	Description        string    `db:"description"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

type RefreshTokenEntity struct {
	Id     int    `db:"id"`
	UserId string `db:"user_id"`
	AppId  string `db:"app_id"`
	Token  string `db:"token"`
}

type OrganizationEntity struct {
	Id          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

type OrganizationPermissionEntity struct {
	Id       int    `db:"id"`
	OrgId    string `db:"org_id"`
	Resource string `db:"resource"`
	Action   string `db:"action"`
	Effect   string `db:"effect"`
}

type OrganizationUserEntity struct {
	Id     int    `db:"id"`
	OrgId  string `db:"org_id"`
	UserId string `db:"user_id"`
}
