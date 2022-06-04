package locker_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib" // init driver
	"github.com/kei2100/locker/postgres"
)

func ExampleLocker() {
	// setup *sql.DB
	db := setupPostgres()
	defer db.Close()

	// Acquire the lock by specified key. (Using PostgreSQL pg_advisory_lock)
	key := randHex(16)
	locker := postgres.NewLocker(db)
	lock, err := locker.Get(context.Background(), key)
	if err != nil {
		panic(fmt.Sprintf("failed to acquire: %+v", err))
	}
	// Releases the lock at the end of the function
	defer lock.Release()
	fmt.Println("lock acquired")

	// The same key will be blocked until the lock is released
	done := make(chan interface{})
	go func() {
		defer close(done)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		lock, err := locker.Get(ctx, key)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return
			}
			panic(fmt.Sprintf("unexpected err %+v", err))
		}
		lock.Release()
		fmt.Println("unexpected lock acquired")
	}()
	<-done

	// Output:
	// lock acquired
}

func setupPostgres() *sql.DB {
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
		panic(err)
	}
	return db
}

func randHex(bytes int) string {
	randBytes := make([]byte, bytes)
	if _, err := rand.Read(randBytes); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", randBytes)
}
