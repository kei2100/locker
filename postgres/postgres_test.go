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

// go test -fuzz=FuzzLocker_sum32Keys -fuzztime 30s github.com/kei2100/locker/postgres
func FuzzLocker_sum32Keys(f *testing.F) {
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
		a1, a2, err := locker.sum32Keys(keyA)
		if err != nil {
			t.Errorf("sum32Keys(%x[%s]) returns an error %+v", keyA, keyA, err)
			return
		}
		b1, b2, err := locker.sum32Keys(keyB)
		if err != nil {
			t.Errorf("sum32Keys(%x[%s]) returns an error %+v", keyB, keyB, err)
			return
		}
		if a1 == b1 && a2 == b2 {
			t.Errorf("collision:\n%x[%s]\n%x[%s]", keyA, keyA, keyB, keyB)
			return
		}
	})
}

// go test -fuzz=FuzzLocker_sum32Keys_table -fuzztime 30s github.com/kei2100/locker/postgres
func FuzzLocker_sum32Keys_table(f *testing.F) {
	locker := NewLocker(nil)
	table := make(map[int32]map[int32]string, 0)
	f.Fuzz(func(t *testing.T, key string) {
		key1, key2, err := locker.sum32Keys(key)
		if err != nil {
			t.Errorf("sum32Keys(%x[%s]) returns an error %+v", key, key, err)
			return
		}
		e1, ok := table[key1]
		if !ok {
			table[key1] = map[int32]string{
				key2: key,
			}
			return
		}
		e2, ok := e1[key2]
		if !ok {
			e1[key2] = key
			return
		}
		if e2 != key {
			t.Errorf("collision:\n%x[%s]\n%x[%s]", e2, e2, key, key)
		}
	})
}
