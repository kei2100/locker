package locker

import "context"

// TODO
type Locker interface {
	// TODO
	Get(ctx context.Context, key string) (Lock, error)
}

// TODO
type Lock interface {
	// TODO
	Release(ctx context.Context) error
}
