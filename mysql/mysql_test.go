package mysql

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // init driver
	"os"
	"testing"
	"time"
)

func TestLocker(t *testing.T) {
	ctx := context.Background()
	db := setupMySQL(t)
	locker := NewLocker(db)
	key := "key"
	// get lock
	lock, err := locker.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	// expect duplicate key
	timeout, notTimeout := make(chan struct{}), make(chan struct{})
	go func() {
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		t.Cleanup(cancel)
		lock, err := locker.Get(ctx, key)
		if err != nil {
			close(timeout)
			return
		}
		defer lock.Release(ctx)
		close(notTimeout)
	}()
	select {
	case <-notTimeout:
		t.Fatal("not timeout")
	case <-timeout:
		// ok
	}
	// release
	lock.Release(ctx)
	timeout, notTimeout = make(chan struct{}), make(chan struct{})
	go func() {
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		t.Cleanup(cancel)
		lock, err := locker.Get(ctx, key)
		if err != nil {
			close(timeout)
			return
		}
		defer lock.Release(ctx)
		close(notTimeout)
	}()
	select {
	case <-notTimeout:
		// ok
	case <-timeout:
		t.Fatal("timeout")
	}
}

func setupMySQL(t testing.TB) *sql.DB {
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
