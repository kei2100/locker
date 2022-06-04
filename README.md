locker
=======

## Description

locker provides an interface and implementations of named lock for exclusivity control between processes and threads.

```go
func ExampleLocker() {
	// setup *sql.DB
	db := setupPostgres()
	defer db.Close()

	// Acquire the lock by specified key. (Using PostgreSQL pg_advisory_lock)
	key := randHex(16)
	locker := postgres.NewLocker(db)
	lock, err := locker.Get(context.Background(), key)
	if err != nil {
		panic(fmt.Sprintf("failed to acquire: %+v", err))
	}
	// Releases the lock at the end of the function
	defer lock.Release()
	fmt.Println("lock acquired")

	// The same key will be blocked until the lock is released
	done := make(chan interface{})
	go func() {
		defer close(done)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		lock, err := locker.Get(ctx, key)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return
			}
			panic(fmt.Sprintf("unexpected err: %+v", err))
		}
		lock.Release()
		fmt.Println("unexpected lock acquired")
	}()
	<-done

	// Output:
	// lock acquired
}
```

## Lock implementations

Currently provides following implementations.

* postgres.Locker
  * postgres.Locker provides a simple lock mechanism using PostgreSQL [pg_advisory_lock](https://www.postgresql.org/docs/14/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS)
  * Recommend that using the [pgx](https://github.com/jackc/pgx) PostgreSQL Driver
* mysql.Locker
  * mysql.Locker provides a simple lock mechanism using MySQL [GET_LOCK](https://dev.mysql.com/doc/refman/8.0/en/locking-functions.html#function_get-lock)

## Installation

`go get` the implementation you need.

```bash
$ go get github.com/kei2100/locker/postgres
$ go get github.com/kei2100/locker/mysql
```
