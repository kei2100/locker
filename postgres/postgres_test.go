package postgres

import (
	"os"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib" // init driver
	"github.com/kei2100/locker/internal/test"
)

func TestLocker(t *testing.T) {
	db := test.SetupPostgres(t)
	locker := NewLocker(db)
	test.TestSpec(t, locker)
}

func TestMultiProcess(t *testing.T) {
	db := test.SetupPostgres(t)
	locker := NewLocker(db)
	environ := []string{
		"HOST_POSTGRES_PORT", os.Getenv("HOST_POSTGRES_PORT"),
	}
	test.TestMultiProcess(t, "postgres", environ, locker)
}

// go test -fuzz=FuzzLocker_sum64Key$ -fuzztime 30s github.com/kei2100/locker/postgres
func FuzzLocker_sum64Key(f *testing.F) {
	locker := NewLocker(nil)
	corpus := [][2]string{
		{"costarring", "liquid"},
		{"declinate", "macallums"},
		{"altarage", "zinke"},
		{"altarages", "zinkes"},
	}
	for _, ss := range corpus {
		f.Add(ss[0], ss[1])
	}
	f.Fuzz(func(t *testing.T, keyA, keyB string) {
		if keyA == keyB {
			return
		}
		a, err := locker.sum64Key(keyA)
		if err != nil {
			t.Errorf("sum64Key(%x[%s]) returns an error %+v", keyA, keyA, err)
			return
		}
		b, err := locker.sum64Key(keyB)
		if err != nil {
			t.Errorf("sum64Key(%x[%s]) returns an error %+v", keyB, keyB, err)
			return
		}
		if a == b {
			t.Errorf("collision:\n%x[%s]\n%x[%s]", keyA, keyA, keyB, keyB)
			return
		}
	})
}

// go test -fuzz=FuzzLocker_sum64Key_table$ -fuzztime 30s github.com/kei2100/locker/postgres
func FuzzLocker_sum64Key_table(f *testing.F) {
	locker := NewLocker(nil)
	table := make(map[int64]string, 0)
	f.Fuzz(func(t *testing.T, key string) {
		k, err := locker.sum64Key(key)
		if err != nil {
			t.Errorf("sum64Key(%x[%s]) returns an error %+v", key, key, err)
			return
		}
		existing, ok := table[k]
		if !ok {
			table[k] = key
			return
		}
		if existing != key {
			t.Errorf("collision:\n%x[%s]\n%x[%s]", existing, existing, key, key)
		}
	})
}
