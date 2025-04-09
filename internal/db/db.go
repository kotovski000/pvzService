package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func InitializeDB(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				return db, nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("failed to connect to DB after retries: %w", err)
}
