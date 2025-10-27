package database

import (
	"database/sql"
	"fmt"
	"log"

	"inventory-system/internal/config"

	_ "github.com/lib/pq"  // PostgreSQL driver
	_ "modernc.org/sqlite" // SQLite driver (pure Go)
)

// NewDatabaseClient crea una conexión a la base de datos según la configuración
func NewDatabaseClient(cfg *config.Config) (*sql.DB, error) {
	var db *sql.DB
	var err error

	switch cfg.DatabaseDriver {
	case "postgres":
		connStr := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.PostgresHost,
			cfg.PostgresPort,
			cfg.PostgresUser,
			cfg.PostgresPassword,
			cfg.PostgresDB,
		)
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to open postgres connection: %w", err)
		}
		log.Printf("✅ Connected to PostgreSQL database: %s", cfg.PostgresDB)

	case "sqlite":
		db, err = sql.Open("sqlite", cfg.SQLitePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite connection: %w", err)
		}

		// Para :memory:, usar una sola conexión (evita crear múltiples DBs en memoria)
		if cfg.SQLitePath == ":memory:" {
			db.SetMaxOpenConns(1)
		}

		log.Printf("✅ Connected to SQLite database: %s", cfg.SQLitePath)

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.DatabaseDriver)
	}

	// Verificar la conexión
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configurar connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * 60) // 5 minutes

	return db, nil
}
