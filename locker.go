package locker

import (
	"context"
)

// Locker is an interface for acquiring named locks.
type Locker interface {
	// Get acquires a lock with the name specified by key.
	// If another session has already acquired a lock with the same name, it blocks until the lock is released or ctx.Done.
	Get(ctx context.Context, key string) (Lock, error)
}

// Lock interface represents an acquired named lock object.
type Lock interface {
	// Release releases the acquired named lock object.
	Release(ctx context.Context)
}
