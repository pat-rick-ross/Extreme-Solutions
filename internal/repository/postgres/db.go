package postgres

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq" // Pure Go Postgres driver initialization
)

// NewConnectionPool creates an optimized, resilient connection pool for Postgres
func NewConnectionPool(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Optimize connection configurations for high-concurrency ISP accounting
	db.SetMaxOpenConns(25)                 // Max concurrent open connections
	db.SetMaxIdleConns(25)                 // Keeps idle connections alive to avoid handshakes
	db.SetConnMaxLifetime(5 * time.Minute) // Cycles connections gracefully

	// Check if connection is alive
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
