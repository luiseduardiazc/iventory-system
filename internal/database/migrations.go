package database

import (
	"database/sql"
	"fmt"
	"os"

	"inventory-system/internal/config"
)

// InitializeSchema aplica las migraciones a la base de datos
func InitializeSchema(db *sql.DB, cfg *config.Config) error {
	// Leer archivo de migración
	schemaSQL, err := os.ReadFile("migrations/001_initial_schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Ejecutar schema
	if _, err := db.Exec(string(schemaSQL)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// HealthCheck verifica que la base de datos esté funcionando
func HealthCheck(db *sql.DB) error {
	return db.Ping()
}
