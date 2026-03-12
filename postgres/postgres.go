package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"hash"
	"io"
	"sync"

	"github.com/cespare/xxhash/v2"
	"github.com/kei2100/locker"
)

// Locker is an implementation of the locker.Locker using PostgreSQL pg_advisory_lock
type Locker struct {
	Logger locker.Logger
	db     *sql.DB
	Hash64 func() hash.Hash64
}

// NewLocker creates a new Locker
func NewLocker(db *sql.DB) *Locker {
	return &Locker{
		Logger: locker.DefaultLogger,
		db:     db,
		Hash64: func() hash.Hash64 { return xxhash.New() },
	}
}

type lock struct {
	logger locker.Logger
	conn   *sql.Conn
	key    int64
	once   sync.Once
}

// Get calls PostgreSQL pg_advisory_lock(bigint) to acquire the lock.
// It uses xxhash64 to convert the argument string key to a 64-bit integer.
func (r *Locker) Get(ctx context.Context, key string) (locker.Lock, error) {
	k, err := r.sum64Key(key)
	if err != nil {
		return nil, err
	}
	return r.GetByRawKey(ctx, k)
}

// GetByRawKey calls PostgreSQL pg_advisory_lock(bigint) to acquire the lock.
func (r *Locker) GetByRawKey(ctx context.Context, key int64) (locker.Lock, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres: get connection: %w", err)
	}
	onerror := func() {
		if err := conn.Close(); err != nil {
			r.Logger.Printf("postgres: an error occurred while closing the connection: %+v", err)
		}
	}
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", key); err != nil {
		defer onerror()
		return nil, fmt.Errorf("postgres: SELECT pg_advisory_lock: %w", err)
	}
	return &lock{
		conn:   conn,
		key:    key,
		logger: r.Logger,
	}, nil
}

func (r *Locker) sum64Key(key string) (int64, error) {
	h := r.Hash64()
	if _, err := io.WriteString(h, key); err != nil {
		return 0, fmt.Errorf("postgres: write string to Hash64 hash function: %w", err)
	}
	return int64(h.Sum64()), nil
}

func (k *lock) Release() {
	k.once.Do(func() {
		defer func() {
			if err := k.conn.Close(); err != nil {
				k.logger.Printf("postgres: an error occurred while closing the connection: %+v", err)
			}
		}()
		row := k.conn.QueryRowContext(context.Background(), "SELECT pg_advisory_unlock($1)", k.key)
		var released bool
		if err := row.Scan(&released); err != nil {
			k.logger.Printf("postgres: failed to release lock: %+v", err)
			return
		}
		if !released {
			panic("postgres: lock already released")
		}
	})
}
