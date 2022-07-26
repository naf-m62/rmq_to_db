package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type (
	DBManager interface {
		Repo() Repo
	}

	dbManager struct {
		repo      Repo
		tableName string
	}
)

func NewDBManager(db *sql.DB, tableName string) DBManager {
	return &dbManager{repo: &repo{db: db, tableName: tableName}}
}

func (d *dbManager) Repo() Repo {
	return d.repo
}

func ConnDB(config *Config) (db *sql.DB, err error) {
	var dsn = fmt.Sprintf("host = %s user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Username, config.Password, config.DBName)

	if db, err = sql.Open(config.DriverName, dsn); err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		log.Println(err)
		return nil, err
	}

	return db, nil
}
