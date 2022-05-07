package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"hash"
	"hash/fnv"
	"io"

	"github.com/kei2100/locker"
	"github.com/kei2100/sync-until-succeed-once"
	"github.com/twmb/murmur3"
)

type Locker struct {
	Logger  locker.Logger
	db      *sql.DB
	Hash32A func() hash.Hash32
	Hash32B func() hash.Hash32
}

// NewLocker creates a new Locker
func NewLocker(db *sql.DB) *Locker {
	return &Locker{
		Logger:  locker.DefaultLogger,
		db:      db,
		Hash32A: fnv.New32a,
		Hash32B: murmur3.New32,
	}
}

type lock struct {
	logger locker.Logger
	conn   *sql.Conn
	key1   int32
	key2   int32
	once   sync.UntilSucceedOnce
}

func (r *Locker) Get(ctx context.Context, key string) (locker.Lock, error) {
	h1 := r.Hash32A()
	if _, err := io.WriteString(h1, key); err != nil {
		return nil, fmt.Errorf("postgres: write string to Hash32A hash function: %w", err)
	}
	h2 := r.Hash32B()
	if _, err := io.WriteString(h2, key); err != nil {
		return nil, fmt.Errorf("postgres: write string to Hash32B hash function: %w", err)
	}
	return r.Get32(ctx, int32(h1.Sum32()), int32(h2.Sum32()))
}

func (r *Locker) Get32(ctx context.Context, key1, key2 int32) (locker.Lock, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres: get connection: %w", err)
	}
	onerror := func() {
		if err := conn.Close(); err != nil {
			r.Logger.Printf("postgres: an error occurred while closing the connection: %+v\n", err)
		}
	}
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1, $2)", key1, key2); err != nil {
		defer onerror()
		return nil, fmt.Errorf("postgres: SELECT pg_advisory_lock: %w", err)
	}
	return &lock{
		conn:   conn,
		key1:   key1,
		key2:   key2,
		logger: locker.DefaultLogger,
	}, nil
}

func (k *lock) Release(ctx context.Context) error {
	if err := k.once.Do(func() error {
		row := k.conn.QueryRowContext(ctx, "SELECT pg_advisory_unlock($1, $2)", k.key1, k.key2)
		var released bool
		if err := row.Scan(&released); err != nil {
			return fmt.Errorf("postgres: release lock: %w", err)
		}
		if !released {
			k.logger.Println("postgres: lock already released")
		}
		if err := k.conn.Close(); err != nil {
			k.logger.Printf("postgres: an error occurred while closing the connection: %+v\n", err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
