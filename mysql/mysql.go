package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kei2100/locker"
	"github.com/kei2100/sync-until-succeed-once"
)

type Locker struct {
	db *sql.DB
}

// NewLocker creates a new Locker
func NewLocker(db *sql.DB) *Locker {
	return &Locker{
		db: db,
	}
}

type lock struct {
	conn *sql.Conn
	key  string
	once sync.UntilSucceedOnce
}

func (r *Locker) Get(ctx context.Context, key string) (locker.Lock, error) {
	conn, err := r.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("mysql: get connection: %w", err)
	}
	onerror := func() {
		if err := conn.Close(); err != nil {
			// TODO log
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
		conn: conn,
		key:  key,
	}, nil
}

func (k *lock) Release(ctx context.Context) error {
	if err := k.once.Do(func() error {
		row := k.conn.QueryRowContext(ctx, "SELECT RELEASE_LOCK(?)", k.key)
		var result sql.NullInt32
		if err := row.Scan(&result); err != nil {
			return fmt.Errorf("mysql: release lock: %w", err)
		}
		if !result.Valid {
			// TODO log
		}
		if err := k.conn.Close(); err != nil {
			// TODO log
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
