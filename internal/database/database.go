package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/svergin/go-home-o-matic/internal/config"
)

func Provide(cfg *config.Config) *sql.DB {
	dbconfig := cfg.DB
	db, err := sql.Open("sqlite3", dbconfig.File)
	if err != nil {
		panic(fmt.Sprintf("failed to open sqlite3 connect: %s", err))
	}

	return db
}
