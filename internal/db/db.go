package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func InitializeDB(dsn string) (*sql.DB, error) {
	return initializeWithRetry(dsn, 5, 2*time.Second)
}

func InitializeTestDB(dsn string) (*sql.DB, error) {
	// Можно использовать другую стратегию для тестов
	return initializeWithRetry(dsn, 3, 1*time.Second)
}

func initializeWithRetry(dsn string, maxRetries int, delay time.Duration) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				// Дополнительные настройки для тестовой БД
				if maxRetries == 3 { // Это тестовая инициализация
					db.SetMaxOpenConns(5)
					db.SetMaxIdleConns(2)
				}
				return db, nil
			}
		}
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("failed to connect to DB after %d retries: %w", maxRetries, err)
}
