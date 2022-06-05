package mysql

import (
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql" // init driver
	"github.com/kei2100/locker/internal/test"
)

func TestLocker(t *testing.T) {
	db := test.SetupMySQL(t)
	locker := NewLocker(db)
	test.TestSpec(t, locker)
}

func TestMultiProcess(t *testing.T) {
	db := test.SetupMySQL(t)
	locker := NewLocker(db)
	environ := []string{
		"HOST_MYSQL_PORT", os.Getenv("HOST_MYSQL_PORT"),
	}
	test.TestMultiProcess(t, "mysql", environ, locker)
}
