package database

import "github.com/jmoiron/sqlx"

type DatabaseProvider interface {
	Get() *sqlx.DB
	Close() error
}
