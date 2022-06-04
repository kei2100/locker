package mysql

import (
	"testing"

	_ "github.com/go-sql-driver/mysql" // init driver
	"github.com/kei2100/locker/internal/test"
)

func TestLocker(t *testing.T) {
	db := test.SetupMySQL(t)
	locker := NewLocker(db)
	test.TestSpec(t, locker)
}
