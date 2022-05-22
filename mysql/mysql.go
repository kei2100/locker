package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/kei2100/locker"
)

type Locker struct {
	Logger locker.Logger
	db     *sql.DB
}

// NewLocker creates a new Locker
func NewLocker(db *sql.DB) *Locker {
	return &Locker{
		Logger: locker.DefaultLogger,
		db:     db,
	}
}

type lock struct {
	logger locker.Logger
	conn   *sql.Conn
	key    string
	once   sync.Once
}

func (r *Locker) Get(ctx context.Context, key string) (locker.Lock, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("mysql: get connection: %w", err)
	}
	onerror := func() {
		if err := conn.Close(); err != nil {
			r.Logger.Printf("mysql: an error occurred while closing the connection: %+v", err)
		}
	}
	var result sql.NullInt32
	row := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, -1)", key)
	if err := row.Scan(&result); err != nil {
		defer onerror()
		return nil, fmt.Errorf("mysql: rows scan: %w", err)
	}
	if result.Int32 != 1 {
		defer onerror()
		return nil, fmt.Errorf("mysql: get_lock(?, -1) returns unexpected result (%v %v)", result.Int32, result.Valid)
	}
	return &lock{
		conn:   conn,
		key:    key,
		logger: locker.DefaultLogger,
	}, nil
}

func (k *lock) Release(ctx context.Context) {
	k.once.Do(func() {
		defer func() {
			if err := k.conn.Close(); err != nil {
				k.logger.Printf("mysql: an error occurred while closing the connection: %+v", err)
			}
		}()
		row := k.conn.QueryRowContext(ctx, "SELECT RELEASE_LOCK(?)", k.key)
		var result sql.NullInt32
		if err := row.Scan(&result); err != nil {
			k.logger.Printf("mysql: failed to release lock: %+v", err)
			return
		}
		if !(result.Valid && result.Int32 == 1) {
			panic("mysql: lock already released")
		}
	})
}
