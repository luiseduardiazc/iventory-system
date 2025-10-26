package repository

import (
	"context"
	"database/sql"
	"fmt"

	"inventory-system/internal/domain"
)

// StockRepository maneja las operaciones de persistencia para stock
type StockRepository struct {
	db *sql.DB
}

// NewStockRepository crea una nueva instancia del repositorio
func NewStockRepository(db *sql.DB) *StockRepository {
	return &StockRepository{db: db}
}

// GetByProductAndStore obtiene el stock de un producto en una tienda específica
func (r *StockRepository) GetByProductAndStore(ctx context.Context, productID, storeID string) (*domain.Stock, error) {
	query := `
		SELECT id, product_id, store_id, quantity, reserved, version, updated_at
		FROM stock
		WHERE product_id = ? AND store_id = ?
	`

	var stock domain.Stock
	err := r.db.QueryRowContext(ctx, query, productID, storeID).Scan(
		&stock.ID,
		&stock.ProductID,
		&stock.StoreID,
		&stock.Quantity,
		&stock.Reserved,
		&stock.Version,
		&stock.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{
			Resource: "Stock",
			ID:       fmt.Sprintf("product=%s, store=%s", productID, storeID),
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	return &stock, nil
}

// GetAllByProduct obtiene el stock de un producto en TODAS las tiendas
func (r *StockRepository) GetAllByProduct(ctx context.Context, productID string) ([]*domain.Stock, error) {
	query := `
		SELECT id, product_id, store_id, quantity, reserved, version, updated_at
		FROM stock
		WHERE product_id = ?
		ORDER BY store_id
	`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock by product: %w", err)
	}
	defer rows.Close()

	var stocks []*domain.Stock
	for rows.Next() {
		var stock domain.Stock
		err := rows.Scan(
			&stock.ID,
			&stock.ProductID,
			&stock.StoreID,
			&stock.Quantity,
			&stock.Reserved,
			&stock.Version,
			&stock.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock: %w", err)
		}
		stocks = append(stocks, &stock)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stocks: %w", err)
	}

	return stocks, nil
}

// GetAllByStore obtiene todo el stock de una tienda
func (r *StockRepository) GetAllByStore(ctx context.Context, storeID string) ([]*domain.Stock, error) {
	query := `
		SELECT id, product_id, store_id, quantity, reserved, version, updated_at
		FROM stock
		WHERE store_id = ?
		ORDER BY product_id
	`

	rows, err := r.db.QueryContext(ctx, query, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock by store: %w", err)
	}
	defer rows.Close()

	var stocks []*domain.Stock
	for rows.Next() {
		var stock domain.Stock
		err := rows.Scan(
			&stock.ID,
			&stock.ProductID,
			&stock.StoreID,
			&stock.Quantity,
			&stock.Reserved,
			&stock.Version,
			&stock.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock: %w", err)
		}
		stocks = append(stocks, &stock)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stocks: %w", err)
	}

	return stocks, nil
}

// Create crea un nuevo registro de stock
func (r *StockRepository) Create(ctx context.Context, stock *domain.Stock) error {
	query := `
		INSERT INTO stock (id, product_id, store_id, quantity, reserved, version, updated_at)
		VALUES (?, ?, ?, ?, ?, 1, CURRENT_TIMESTAMP)
	`

	_, err := r.db.ExecContext(ctx, query,
		stock.ID,
		stock.ProductID,
		stock.StoreID,
		stock.Quantity,
		stock.Reserved,
	)

	if err != nil {
		return fmt.Errorf("failed to create stock: %w", err)
	}

	return nil
}

// UpdateQuantity actualiza la cantidad de stock con optimistic locking
func (r *StockRepository) UpdateQuantity(ctx context.Context, stock *domain.Stock) error {
	query := `
		UPDATE stock
		SET quantity = ?,
		    version = version + 1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND version = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		stock.Quantity,
		stock.ID,
		stock.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to update stock quantity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.ConflictError{
			Message: fmt.Sprintf("optimistic lock failed: stock was modified by another transaction (current version: %d)", stock.Version),
		}
	}

	return nil
}

// ReserveStock incrementa la cantidad reservada (usado por reservas)
// Usa SELECT FOR UPDATE para lock pesimista en operaciones críticas
func (r *StockRepository) ReserveStock(ctx context.Context, productID, storeID string, quantity int) error {
	// Iniciar transacción
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// SELECT FOR UPDATE - lock pesimista
	query := `
		SELECT id, product_id, store_id, quantity, reserved, version
		FROM stock
		WHERE product_id = ? AND store_id = ?
	`

	var stock domain.Stock
	err = tx.QueryRowContext(ctx, query, productID, storeID).Scan(
		&stock.ID,
		&stock.ProductID,
		&stock.StoreID,
		&stock.Quantity,
		&stock.Reserved,
		&stock.Version,
	)

	if err == sql.ErrNoRows {
		return &domain.NotFoundError{
			Resource: "Stock",
			ID:       fmt.Sprintf("product=%s, store=%s", productID, storeID),
		}
	}
	if err != nil {
		return fmt.Errorf("failed to lock stock: %w", err)
	}

	// Validar disponibilidad
	available := stock.Quantity - stock.Reserved
	if available < quantity {
		return &domain.InsufficientStockError{
			ProductID: productID,
			StoreID:   storeID,
			Available: available,
			Requested: quantity,
		}
	}

	// Actualizar reservado
	updateQuery := `
		UPDATE stock
		SET reserved = reserved + ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err = tx.ExecContext(ctx, updateQuery, quantity, stock.ID)
	if err != nil {
		return fmt.Errorf("failed to update reserved stock: %w", err)
	}

	// Commit transacción
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReleaseReservedStock libera stock reservado (cuando se cancela una reserva)
func (r *StockRepository) ReleaseReservedStock(ctx context.Context, productID, storeID string, quantity int) error {
	query := `
		UPDATE stock
		SET reserved = reserved - ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE product_id = ? AND store_id = ?
		  AND reserved >= ?
	`

	result, err := r.db.ExecContext(ctx, query, quantity, productID, storeID, quantity)
	if err != nil {
		return fmt.Errorf("failed to release reserved stock: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.ConflictError{
			Message: fmt.Sprintf("cannot release %d units: insufficient reserved stock", quantity),
		}
	}

	return nil
}

// ConfirmReservation confirma una reserva (decrementa quantity y reserved)
func (r *StockRepository) ConfirmReservation(ctx context.Context, productID, storeID string, quantity int) error {
	query := `
		UPDATE stock
		SET quantity = quantity - ?,
		    reserved = reserved - ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE product_id = ? AND store_id = ?
		  AND quantity >= ?
		  AND reserved >= ?
	`

	result, err := r.db.ExecContext(ctx, query,
		quantity, quantity, productID, storeID, quantity, quantity)

	if err != nil {
		return fmt.Errorf("failed to confirm reservation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.ConflictError{
			Message: fmt.Sprintf("cannot confirm reservation: insufficient stock or reserved quantity"),
		}
	}

	return nil
}

// GetLowStockItems retorna productos con stock bajo (cantidad < umbral)
func (r *StockRepository) GetLowStockItems(ctx context.Context, threshold int) ([]*domain.Stock, error) {
	query := `
		SELECT id, product_id, store_id, quantity, reserved, version, updated_at
		FROM stock
		WHERE (quantity - reserved) < ?
		ORDER BY (quantity - reserved) ASC
	`

	rows, err := r.db.QueryContext(ctx, query, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock items: %w", err)
	}
	defer rows.Close()

	var stocks []*domain.Stock
	for rows.Next() {
		var stock domain.Stock
		err := rows.Scan(
			&stock.ID,
			&stock.ProductID,
			&stock.StoreID,
			&stock.Quantity,
			&stock.Reserved,
			&stock.Version,
			&stock.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock: %w", err)
		}
		stocks = append(stocks, &stock)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stocks: %w", err)
	}

	return stocks, nil
}
