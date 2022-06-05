//go:build postgres

package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib" // init driver
	"github.com/kei2100/locker"
	"github.com/kei2100/locker/postgres"
)

func init() {
	newLocker = func() (locker.Locker, cleanup, error) {
		db, err := newDB()
		if err != nil {
			return nil, nil, err
		}
		cleanup := func() {
			db.Close()
		}
		return postgres.NewLocker(db), cleanup, nil
	}
}

func newDB() (*sql.DB, error) {
	user := "develop"
	password := "develop"
	host := "localhost"
	port := os.Getenv("HOST_POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	database := "develop"
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user,
		password,
		host,
		port,
		database,
	)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("main: open db: %w", err)
	}
	return db, nil
}
