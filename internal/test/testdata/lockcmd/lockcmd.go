package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/kei2100/locker"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: lockcmd [key]")
		os.Exit(1)
	}
	key := os.Args[1]
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	err := run(ctx, key)
	if !errors.Is(err, ctx.Err()) {
		fmt.Fprintf(os.Stderr, "unexpected err: %+v\n", err)
		os.Exit(1)
	}
}

type cleanup func()

var newLocker func() (locker.Locker, cleanup, error) = nil

func run(ctx context.Context, key string) error {
	locker, cleanup, err := newLocker()
	if err != nil {
		return err
	}
	defer cleanup()
	lock, err := locker.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("main: get lock: %w", err)
	}
	defer lock.Release()
	<-ctx.Done()
	return ctx.Err()
}
