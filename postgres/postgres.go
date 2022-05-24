package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"hash"
	"hash/fnv"
	"io"
	"sync"

	"github.com/kei2100/locker"
	"github.com/twmb/murmur3"
)

// Locker is an implementation of the locker.Locker using PostgreSQL pg_advisory_lock
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
	keyA   int32
	keyB   int32
	once   sync.Once
}

// Get calls PostgreSQL pg_advisory_lock(keyA int, keyB int) to acquire the lock.
// Since these pg_advisory_lock keys are 32-bit integers, this method uses two different hash functions to convert the argument string keys to integers
// and uses them as the keys for pg_advisory_lock.
// Default hash functions are FNV-1a hash and Murmur hash.
func (r *Locker) Get(ctx context.Context, key string) (locker.Lock, error) {
	keyA, keyB, err := r.sum32Keys(key)
	if err != nil {
		return nil, err
	}
	return r.GetByRawKey(ctx, keyA, keyB)
}

// GetByRawKey calls PostgreSQL pg_advisory_lock(keyA int, keyB int) to acquire the lock.
func (r *Locker) GetByRawKey(ctx context.Context, keyA, keyB int32) (locker.Lock, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("postgres: get connection: %w", err)
	}
	onerror := func() {
		if err := conn.Close(); err != nil {
			r.Logger.Printf("postgres: an error occurred while closing the connection: %+v", err)
		}
	}
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1, $2)", keyA, keyB); err != nil {
		defer onerror()
		return nil, fmt.Errorf("postgres: SELECT pg_advisory_lock: %w", err)
	}
	return &lock{
		conn:   conn,
		keyA:   keyA,
		keyB:   keyB,
		logger: r.Logger,
	}, nil
}

func (r *Locker) sum32Keys(key string) (keyA, keyB int32, err error) {
	hA := r.Hash32A()
	if _, err := io.WriteString(hA, key); err != nil {
		return 0, 0, fmt.Errorf("postgres: write string to Hash32A hash function: %w", err)
	}
	hB := r.Hash32B()
	if _, err := io.WriteString(hB, key); err != nil {
		return 0, 0, fmt.Errorf("postgres: write string to Hash32B hash function: %w", err)
	}
	return int32(hA.Sum32()), int32(hB.Sum32()), nil
}

func (k *lock) Release() {
	k.once.Do(func() {
		defer func() {
			if err := k.conn.Close(); err != nil {
				k.logger.Printf("postgres: an error occurred while closing the connection: %+v", err)
			}
		}()
		row := k.conn.QueryRowContext(context.Background(), "SELECT pg_advisory_unlock($1, $2)", k.keyA, k.keyB)
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
