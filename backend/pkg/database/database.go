package database

import (
    "fmt"
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq" // Import side effects of postgres connector
)

type DB struct {
    db *sqlx.DB
}

func NewDB(driver string, host string, name string, user string, password string) *DB {
    dsn := fmt.Sprintf("%s://%s:%s@%s/%s?sslmode=%s", driver, user, password, host, name, "disable")
    return &DB{db: sqlx.MustConnect(driver, dsn)}
}
