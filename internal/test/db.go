package test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql" // init driver
	_ "github.com/jackc/pgx/v4/stdlib" // init driver
)

func SetupMySQL(t testing.TB) *sql.DB {
	t.Helper()
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
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	return db
}

func SetupPostgres(t testing.TB) *sql.DB {
	t.Helper()
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
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	return db
}
