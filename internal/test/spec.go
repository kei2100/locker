package test

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kei2100/locker"
)

const numThreads = 10
const lockTimeout = 500 * time.Millisecond

// TestSpec tests whether the specific Locker implementation meets the specification
func TestSpec(t *testing.T, impl locker.Locker) {
	t.Run("lock not acquired", func(t *testing.T) {
		t.Parallel()
		key := randHex(16)
		results := make([]*lockOrErr, numThreads)

		wg := sync.WaitGroup{}
		wg.Add(numThreads)
		for i := 0; i < numThreads; i++ {
			go func(i int) {
				defer wg.Done()
				results[i] = getLockOrError(t, impl, key, lockTimeout)
			}(i)
		}
		wg.Wait()

		assertNumAcquiredOrTimeout(t, results, 1, numThreads-1)
	})

	t.Run("lock already acquired", func(t *testing.T) {
		t.Parallel()
		key := randHex(16)
		results := make([]*lockOrErr, numThreads)

		acquired := getLockOrError(t, impl, key, lockTimeout)
		if acquired.lock == nil {
			t.Errorf("failed to acquire: %+v", acquired.err)
			return
		}
		defer acquired.lock.Release()

		wg := sync.WaitGroup{}
		wg.Add(numThreads)
		for i := 0; i < numThreads; i++ {
			go func(i int) {
				defer wg.Done()
				results[i] = getLockOrError(t, impl, key, lockTimeout)
			}(i)
		}
		wg.Wait()

		assertNumAcquiredOrTimeout(t, results, 0, numThreads)
	})

	t.Run("lock already released", func(t *testing.T) {
		t.Parallel()
		key := randHex(16)
		results := make([]*lockOrErr, numThreads)

		acquired := getLockOrError(t, impl, key, lockTimeout)
		if acquired.lock == nil {
			t.Errorf("failed to acquire: %+v", acquired.err)
			return
		}
		acquired.lock.Release()

		wg := sync.WaitGroup{}
		wg.Add(numThreads)
		for i := 0; i < numThreads; i++ {
			go func(i int) {
				defer wg.Done()
				results[i] = getLockOrError(t, impl, key, lockTimeout)
			}(i)
		}
		wg.Wait()

		assertNumAcquiredOrTimeout(t, results, 1, numThreads-1)
	})

	t.Run("acquire by different keys", func(t *testing.T) {
		t.Parallel()
		results := make([]*lockOrErr, numThreads)

		wg := sync.WaitGroup{}
		wg.Add(numThreads)
		for i := 0; i < numThreads; i++ {
			go func(i int) {
				defer wg.Done()
				key := randHex(16)
				results[i] = getLockOrError(t, impl, key, lockTimeout)
			}(i)
		}
		wg.Wait()

		assertNumAcquiredOrTimeout(t, results, numThreads, 0)
	})
}

type lockOrErr struct {
	lock locker.Lock
	err  error
}

func getLockOrError(t *testing.T, impl locker.Locker, key string, timeout time.Duration) *lockOrErr {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	lock, err := impl.Get(ctx, key)
	return &lockOrErr{
		lock: lock,
		err:  err,
	}
}

func assertNumAcquiredOrTimeout(t *testing.T, results []*lockOrErr, wantNumAcquired, wantNumTimeout int) {
	var numAcquired int
	var numTimeout int
	for _, r := range results {
		if r.lock != nil {
			numAcquired++
			r.lock.Release()
			continue
		}
		if !errors.Is(r.err, context.DeadlineExceeded) {
			t.Errorf("unexpected err (%T)%+v", r.err, r.err)
			continue
		}
		numTimeout++
	}
	if g, w := numAcquired, wantNumAcquired; g != w {
		t.Errorf("numAcquired\ngot :%v\nwant:%v", numAcquired, wantNumAcquired)
	}
	if g, w := numTimeout, wantNumTimeout; g != w {
		t.Errorf("numTimeout\ngot :%v\nwant:%v", numTimeout, wantNumTimeout)
	}
}

func randHex(bytes int) string {
	randBytes := make([]byte, bytes)
	if _, err := rand.Read(randBytes); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", randBytes)
}
