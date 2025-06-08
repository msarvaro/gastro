package postgres

import (
	"database/sql"
	"restaurant-management/configs"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewDB(config *configs.Config) (*DB, error) {
	db, err := sql.Open("postgres", config.GetDBConnString())
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}
