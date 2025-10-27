package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"inventory-system/internal/config"

	_ "modernc.org/sqlite" // SQLite driver (pure Go)
)

// NewDatabaseClient crea una conexión a la base de datos SQLite
func NewDatabaseClient(cfg *config.Config) (*sql.DB, error) {
	var db *sql.DB
	var err error

	// Solo soportamos SQLite
	if cfg.DatabaseDriver != "sqlite" {
		return nil, fmt.Errorf("unsupported database driver: %s (only sqlite is supported)", cfg.DatabaseDriver)
	}

	db, err = sql.Open("sqlite", cfg.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite connection: %w", err)
	}

	// Para :memory:, usar una sola conexión (evita crear múltiples DBs en memoria)
	if cfg.SQLitePath == ":memory:" {
		db.SetMaxOpenConns(1)
	}

	log.Printf("✅ Connected to SQLite database: %s", cfg.SQLitePath)

	// Verificar la conexión
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configurar connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
