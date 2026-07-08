package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"Extreme-Solutions/internal/repository/postgres"
)

func main() {
	// 1. Grab DB configuration from environment variables (fallback to local dev values)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/extreme_isp?sslmode=disable"
	}

	log.Println("Connecting to the database to run schema migration loops...")
	db, err := postgres.NewConnectionPool(dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize database pool: %v", err)
	}
	defer db.Close()

	// 2. Locate your migrations directory path
	migrationFile := filepath.Join("internal", "migrations", "000001_init_schema.up.sql")
	log.Printf("Reading migration script: %s", migrationFile)

	scriptBytes, err := os.ReadFile(migrationFile)
	if err != nil {
		log.Fatalf("Critical error tracking down down schema sql file: %v", err)
	}

	// 3. Fire migration statements inside a strict transactional context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to spin up database transaction: %v", err)
	}

	// Rollback cleanly if execution stops prematurely
	defer tx.Rollback()

	log.Println("Executing migration schema script updates...")
	_, err = tx.ExecContext(ctx, string(scriptBytes))
	if err != nil {
		log.Fatalf("Migration Loop CRASHED. Rolling back database updates. Error: %v", err)
	}

	// Commit updates firmly
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit schema migration tracking details: %v", err)
	}

	fmt.Println("🚀 Migration complete! Database schemas, network indexing, and billing profiles are up to date.")
}
