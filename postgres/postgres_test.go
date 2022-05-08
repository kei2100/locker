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

func FuzzLocker_sum32Keys(f *testing.F) {
	locker := NewLocker(setupPostgres(f))
	corpus := [][2]string{
		{"costarring", "liquid"},
		{"declinate", "macallums"},
		{"altarage", "zinke"},
		{"altarages", "zinkes"},
	}
	for _, ss := range corpus {
		f.Add(ss[0], ss[1])
	}
	f.Fuzz(func(t *testing.T, keyA, keyB string) {
		if keyA == keyB {
			return
		}
		a1, a2, err := locker.sum32Keys(keyA)
		if err != nil {
			t.Errorf("sum32Keys(%x[%s]) returns an error %+v", keyA, keyA, err)
			return
		}
		b1, b2, err := locker.sum32Keys(keyB)
		if err != nil {
			t.Errorf("sum32Keys(%x[%s]) returns an error %+v", keyB, keyB, err)
			return
		}
		if a1 == b1 && a2 == b2 {
			t.Errorf("collision:\n%x[%s]\n%x[%s]", keyA, keyA, keyB, keyB)
			return
		}
	})
}

func FuzzLocker_sum32Kyes_table(f *testing.F) {
	locker := NewLocker(setupPostgres(f))
	table := make(map[int32]map[int32]string, 0)
	f.Fuzz(func(t *testing.T, key string) {
		key1, key2, err := locker.sum32Keys(key)
		if err != nil {
			t.Errorf("sum32Keys(%x[%s]) returns an error %+v", key, key, err)
			return
		}
		e1, ok := table[key1]
		if !ok {
			table[key1] = map[int32]string{
				key2: key,
			}
			return
		}
		e2, ok := e1[key2]
		if !ok {
			e1[key2] = key
			return
		}
		if e2 != key {
			t.Errorf("collision:\n%x[%s]\n%x[%s]", e2, e2, key, key)
		}
	})
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
