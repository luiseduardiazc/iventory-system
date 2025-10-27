package database_test

import (
	"testing"

	"inventory-system/internal/config"
	"inventory-system/internal/database"
)

func TestPlaceholderConversion(t *testing.T) {
	tests := []struct {
		name     string
		driver   string
		input    string
		expected string
	}{
		{
			name:     "PostgreSQL simple query",
			driver:   "postgres",
			input:    "SELECT * FROM users WHERE id = ?",
			expected: "SELECT * FROM users WHERE id = $1",
		},
		{
			name:     "PostgreSQL multiple placeholders",
			driver:   "postgres",
			input:    "INSERT INTO products (name, price, category) VALUES (?, ?, ?)",
			expected: "INSERT INTO products (name, price, category) VALUES ($1, $2, $3)",
		},
		{
			name:     "PostgreSQL complex query",
			driver:   "postgres",
			input:    "UPDATE stock SET quantity = ?, updated_at = ? WHERE id = ? AND version = ?",
			expected: "UPDATE stock SET quantity = $1, updated_at = $2 WHERE id = $3 AND version = $4",
		},
		{
			name:     "SQLite query unchanged",
			driver:   "sqlite",
			input:    "SELECT * FROM users WHERE id = ?",
			expected: "SELECT * FROM users WHERE id = ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock DB with the specified driver
			db := &database.DB{
				Driver: tt.driver,
			}

			result := db.ConvertPlaceholders(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertPlaceholders() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDatabaseConnection(t *testing.T) {
	t.Run("SQLite connection", func(t *testing.T) {
		cfg := &config.Config{
			DatabaseDriver: "sqlite",
			SQLitePath:     ":memory:",
		}

		db, err := database.NewDatabaseClient(cfg)
		if err != nil {
			t.Fatalf("Failed to connect to SQLite: %v", err)
		}
		defer db.Close()

		if db.Driver != "sqlite" {
			t.Errorf("Expected driver 'sqlite', got '%s'", db.Driver)
		}

		// Test that connection works
		err = db.Ping()
		if err != nil {
			t.Errorf("Failed to ping SQLite database: %v", err)
		}
	})

	// Note: PostgreSQL test would require a running PostgreSQL instance
	// which might not be available in CI/CD environments
}
