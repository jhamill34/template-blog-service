package database

import (
	"github.com/jmoiron/sqlx"
	
	_ "github.com/go-sql-driver/mysql"
)

type MySQLDbProvider struct {
	path string
	db   *sqlx.DB
}

func NewMySQLDbProvider(path string) *MySQLDbProvider {
	return &MySQLDbProvider{
		path: path,
		db:   nil,
	}
}

func (p *MySQLDbProvider) Get() *sqlx.DB {
	if p.db == nil {
		db, err := sqlx.Open("mysql", p.path)
		if err != nil {
			panic(err)
		}

		p.db = db
	}

	return p.db
}

func (p *MySQLDbProvider) Close() error {
	if p.db != nil {
		return p.db.Close()
	}

	return nil
}
