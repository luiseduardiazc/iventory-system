package testutil

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// SetupTestDB crea una base de datos SQLite en memoria para tests
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Aplicar schema
	schema := `
	CREATE TABLE IF NOT EXISTS products (
		id TEXT PRIMARY KEY,
		sku TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		category TEXT,
		price REAL NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS stock (
		id TEXT PRIMARY KEY,
		product_id TEXT NOT NULL,
		store_id TEXT NOT NULL,
		quantity INTEGER NOT NULL DEFAULT 0,
		reserved INTEGER NOT NULL DEFAULT 0,
		min_stock INTEGER NOT NULL DEFAULT 0,
		max_stock INTEGER NOT NULL DEFAULT 0,
		version INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(product_id, store_id),
		FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS reservations (
		id TEXT PRIMARY KEY,
		product_id TEXT NOT NULL,
		store_id TEXT NOT NULL,
		customer_id TEXT NOT NULL,
		quantity INTEGER NOT NULL,
		status TEXT NOT NULL CHECK(status IN ('PENDING', 'CONFIRMED', 'CANCELLED', 'EXPIRED')),
		reference_id TEXT,
		expires_at DATETIME NOT NULL,
		confirmed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		event_type TEXT NOT NULL,
		aggregate_type TEXT NOT NULL,
		aggregate_id TEXT NOT NULL,
		store_id TEXT NOT NULL,
		payload TEXT NOT NULL,
		synced INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Índices para optimización
	CREATE INDEX IF NOT EXISTS idx_stock_product_store ON stock(product_id, store_id);
	CREATE INDEX IF NOT EXISTS idx_stock_store ON stock(store_id);
	CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status);
	CREATE INDEX IF NOT EXISTS idx_reservations_product_store ON reservations(product_id, store_id);
	CREATE INDEX IF NOT EXISTS idx_events_synced ON events(synced);
	CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);

	-- Datos de ejemplo para tests
	INSERT INTO products (id, sku, name, description, category, price) VALUES
		('550e8400-e29b-41d4-a716-446655440000', 'PROD-001', 'Laptop HP Pavilion 15', 'Laptop para tests', 'electronics', 599.99),
		('550e8400-e29b-41d4-a716-446655440001', 'PROD-002', 'Mouse Logitech', 'Mouse para tests', 'accessories', 99.99),
		('550e8400-e29b-41d4-a716-446655440002', 'PROD-003', 'Teclado Mecanico', 'Teclado para tests', 'accessories', 89.99),
		('550e8400-e29b-41d4-a716-446655440003', 'PROD-004', 'Monitor LG 27', 'Monitor para tests', 'electronics', 349.99),
		('550e8400-e29b-41d4-a716-446655440004', 'PROD-005', 'Webcam Logitech', 'Webcam para tests', 'accessories', 79.99);

	INSERT INTO stock (id, product_id, store_id, quantity, reserved, version) VALUES
		('stock-mad-001', '550e8400-e29b-41d4-a716-446655440000', 'MAD-001', 10, 0, 1),
		('stock-mad-002', '550e8400-e29b-41d4-a716-446655440001', 'MAD-001', 50, 5, 1),
		('stock-mad-003', '550e8400-e29b-41d4-a716-446655440002', 'MAD-001', 20, 2, 1),
		('stock-mad-004', '550e8400-e29b-41d4-a716-446655440003', 'MAD-001', 5, 1, 1),
		('stock-mad-005', '550e8400-e29b-41d4-a716-446655440004', 'MAD-001', 15, 0, 1),
		('stock-bcn-001', '550e8400-e29b-41d4-a716-446655440000', 'BCN-001', 15, 2, 1),
		('stock-bcn-002', '550e8400-e29b-41d4-a716-446655440001', 'BCN-001', 30, 3, 1),
		('stock-bcn-003', '550e8400-e29b-41d4-a716-446655440002', 'BCN-001', 25, 0, 1),
		('stock-bcn-004', '550e8400-e29b-41d4-a716-446655440003', 'BCN-001', 8, 0, 1),
		('stock-bcn-005', '550e8400-e29b-41d4-a716-446655440004', 'BCN-001', 20, 1, 1),
		('stock-val-001', '550e8400-e29b-41d4-a716-446655440000', 'VAL-001', 5, 1, 1),
		('stock-val-002', '550e8400-e29b-41d4-a716-446655440001', 'VAL-001', 40, 5, 1),
		('stock-val-003', '550e8400-e29b-41d4-a716-446655440002', 'VAL-001', 0, 0, 1),
		('stock-val-004', '550e8400-e29b-41d4-a716-446655440003', 'VAL-001', 3, 0, 1),
		('stock-val-005', '550e8400-e29b-41d4-a716-446655440004', 'VAL-001', 12, 2, 1),
		('stock-sev-001', '550e8400-e29b-41d4-a716-446655440000', 'SEV-001', 20, 3, 1),
		('stock-sev-002', '550e8400-e29b-41d4-a716-446655440001', 'SEV-001', 35, 2, 1),
		('stock-sev-003', '550e8400-e29b-41d4-a716-446655440002', 'SEV-001', 18, 1, 1),
		('stock-sev-004', '550e8400-e29b-41d4-a716-446655440003', 'SEV-001', 6, 0, 1),
		('stock-sev-005', '550e8400-e29b-41d4-a716-446655440004', 'SEV-001', 10, 0, 1);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

// CleanupTestDB cierra la conexión a la base de datos de test
func CleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Errorf("Failed to close test database: %v", err)
	}
}

// TruncateTables limpia todas las tablas (útil para integration tests)
func TruncateTables(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{"events", "reservations", "stock", "products"}
	for _, table := range tables {
		_, err := db.Exec("DELETE FROM " + table)
		if err != nil {
			t.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}
}

// BeginTx inicia una transacción para tests
func BeginTx(t *testing.T, db *sql.DB) *sql.Tx {
	t.Helper()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	return tx
}

// RollbackTx hace rollback de una transacción (útil para cleanup)
func RollbackTx(t *testing.T, tx *sql.Tx) {
	t.Helper()

	if err := tx.Rollback(); err != nil {
		t.Errorf("Failed to rollback transaction: %v", err)
	}
}
