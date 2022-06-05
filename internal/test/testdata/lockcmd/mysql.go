//go:build mysql

package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql" // init driver
	"github.com/kei2100/locker"
	"github.com/kei2100/locker/mysql"
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
		return mysql.NewLocker(db), cleanup, nil
	}
}

func newDB() (*sql.DB, error) {
	user := "develop"
	password := "develop"
	host := "localhost"
	port := os.Getenv("HOST_MYSQL_PORT")
	if port == "" {
		port = "3306"
	}
	database := "develop"
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&collation=utf8mb4_bin&loc=UTC&parseTime=true",
		user,
		password,
		host,
		port,
		database,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("main: open db: %w", err)
	}
	return db, nil
}
