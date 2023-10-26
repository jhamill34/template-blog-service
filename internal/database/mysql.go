package database

import (
	"log"
	"time"

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

// TODO: add retry config to global config
func (p *MySQLDbProvider) Get() *sqlx.DB {
	if p.db == nil {
		var db *sqlx.DB
		var err error

		for i := 0; i < 10; i++ {
			log.Println("Trying to connect to db")
			db, err = sqlx.Open("mysql", p.path)
			if err == nil {
				break
			}
				
			log.Println(err.Error())
			time.Sleep(5 * time.Second)
		}

		if err != nil {
			log.Fatal(err.Error())
		}

		for i := 0; i < 10; i++ {
			log.Println("Trying to ping db")
			err = db.Ping()
			if err == nil {
				break
			}
			log.Println(err.Error())
			time.Sleep(5 * time.Second)

		}

		if err != nil {
			log.Fatal(err.Error())
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
