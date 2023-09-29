package database

import (
	"github.com/jmoiron/sqlx"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteDbProvider struct {
	path string
	db   *sqlx.DB
}

func NewSqliteDbProvider(path string) *SqliteDbProvider {
	return &SqliteDbProvider{
		path: path,
		db:   nil,
	}
}

func (p *SqliteDbProvider) Get() *sqlx.DB {
	if p.db == nil {
		db, err := sqlx.Open("sqlite3", p.path)
		if err != nil {
			panic(err)
		}

		p.db = db
	}

	return p.db
}

func (p *SqliteDbProvider) Close() error {
	if p.db != nil {
		return p.db.Close()
	}

	return nil
}
