package test

import (
	"context"
	"errors"
	"fmt"
	"github.com/kei2100/locker"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestMultiProcess(t *testing.T, buildFlag string, environ []string, locker locker.Locker) {
	if err := buildLockcmd(buildFlag); err != nil {
		t.Error(err)
	}
	commandCtx, commandCancel := context.WithCancel(context.Background())
	defer commandCancel()
	key := randHex(16)
	go execLockcmd(commandCtx, environ, key)

	time.Sleep(lockTimeout)

	lockerCtx, lockerCancel := context.WithTimeout(context.Background(), lockTimeout)
	defer lockerCancel()
	_, err := locker.Get(lockerCtx, key)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("want DeadlineExceeded, got %v", err)
	}

	commandCancel()
	time.Sleep(lockTimeout)

	lockerCtx, lockerCancel = context.WithTimeout(context.Background(), lockTimeout)
	defer lockerCancel()
	lock, err := locker.Get(lockerCtx, key)
	if err != nil {
		t.Errorf("want not err, got %v", err)
	}
	lock.Release()
}

func buildLockcmd(buildFlag string) error {
	gocmd, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("test: look path go command: %+v", err)
	}
	cmd := exec.Cmd{
		Path: gocmd,
		Args: []string{"go", "build", "-tags", buildFlag},
		Dir:  lockcmdDir(),
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("test: failed to go build: %+v", err)
	}
	return nil
}

func execLockcmd(ctx context.Context, environ []string, key string) error {
	lockcmd := exec.Cmd{
		Path: path.Join(lockcmdDir(), "lockcmd"),
		Args: []string{"lockcmd", key},
		Env:  append(os.Environ(), environ...),
	}
	if err := lockcmd.Start(); err != nil {
		return fmt.Errorf("test: start lockcmd: %w", err)
	}
	<-ctx.Done()
	// stop lockcmd immediately
	if err := lockcmd.Process.Kill(); err != nil {
		return fmt.Errorf("test: kill lockcmd: %w", err)
	}
	if err := lockcmd.Wait(); err != nil {
		return fmt.Errorf("test: wait lockcmd: %w", err)
	}
	return nil
}

func lockcmdDir() string {
	_, f, _, _ := runtime.Caller(0)
	curdir := filepath.Dir(f)
	return path.Join(curdir, "testdata", "lockcmd")
}
