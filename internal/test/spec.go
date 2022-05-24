package test

import (
	"context"
	"errors"
	"github.com/kei2100/locker"
	"testing"
	"time"
)

// TestSpec tests whether the specific Locker implementation meets the specification
func TestSpec(t *testing.T, impl locker.Locker) {
	t.Helper()
	const n = 2
	const timeout = time.Millisecond * 100
	key := "key/foo"
	results := make([]<-chan getOrTimeoutResult, n)
	for i := 0; i < n; i++ {
		ch := getOrErr(t, impl, key, timeout)
		results[i] = ch
	}
	// expect: 1 aquired / 9 timeouts
	var timeouts int
	var aquired locker.Lock
	for _, ch := range results {
		result := <-ch
		if result.timeout {
			timeouts++
		} else {
			aquired = result.lock
		}
	}
	if g, w := timeouts, n-1; g != w {
		t.Errorf("timeouts\ngot :%v\nwant:%v", timeouts, n-1)
	}
	// expect 10 timeouts
	timeouts = 0
	for i := 0; i < n; i++ {
		ch := getOrErr(t, impl, key, timeout)
		results[i] = ch
	}
	for _, ch := range results {
		result := <-ch
		if result.timeout {
			timeouts++
		} else {
			aquired = result.lock
		}
	}
	if g, w := timeouts, n; g != w {
		t.Errorf("timeouts\ngot :%v\nwant:%v", timeouts, n-1)
	}
	aquired.Release()
}

type getOrTimeoutResult struct {
	lock    locker.Lock
	timeout bool
}

func getOrErr(t *testing.T, impl locker.Locker, key string, timeout time.Duration) <-chan getOrTimeoutResult {
	t.Helper()
	result := make(chan getOrTimeoutResult, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		t.Cleanup(cancel)
		lock, err := impl.Get(ctx, key)
		if err != nil {
			if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
				t.Errorf("unexpected err occurred: %+v", err)
			}
		}
		result <- getOrTimeoutResult{
			lock:    lock,
			timeout: err != nil,
		}
	}()
	return result
}
