package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq" // init driver
)

func TestLocker(t *testing.T) {
	ctx := context.Background()
	db := setupPostgres(t)
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

func setupPostgres(t testing.TB) *sql.DB {
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
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	return db
}
