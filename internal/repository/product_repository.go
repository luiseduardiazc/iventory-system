package repository

import (
	"context"
	"database/sql"
	"fmt"

	"inventory-system/internal/domain"
)

// ProductRepository maneja las operaciones de persistencia para productos
type ProductRepository struct {
	db *sql.DB
}

// NewProductRepository crea una nueva instancia del repositorio
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create crea un nuevo producto
func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (id, sku, name, description, category, price, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	_, err := r.db.ExecContext(ctx, query,
		product.ID,
		product.SKU,
		product.Name,
		product.Description,
		product.Category,
		product.Price,
	)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID obtiene un producto por su ID
func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `
		SELECT id, sku, name, description, category, price, created_at, updated_at
		FROM products
		WHERE id = ?
	`

	var product domain.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.SKU,
		&product.Name,
		&product.Description,
		&product.Category,
		&product.Price,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{Resource: "Product", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// GetBySKU obtiene un producto por su SKU
func (r *ProductRepository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	query := `
		SELECT id, sku, name, description, category, price, created_at, updated_at
		FROM products
		WHERE sku = ?
	`

	var product domain.Product
	err := r.db.QueryRowContext(ctx, query, sku).Scan(
		&product.ID,
		&product.SKU,
		&product.Name,
		&product.Description,
		&product.Category,
		&product.Price,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{Resource: "Product", ID: sku}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}

	return &product, nil
}

// List obtiene una lista paginada de productos
func (r *ProductRepository) List(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	query := `
		SELECT id, sku, name, description, category, price, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var product domain.Product
		err := rows.Scan(
			&product.ID,
			&product.SKU,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.Price,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

// ListByCategory obtiene productos por categoría
func (r *ProductRepository) ListByCategory(ctx context.Context, category string, limit, offset int) ([]*domain.Product, error) {
	query := `
		SELECT id, sku, name, description, category, price, created_at, updated_at
		FROM products
		WHERE category = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, category, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list products by category: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var product domain.Product
		err := rows.Scan(
			&product.ID,
			&product.SKU,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.Price,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

// Update actualiza un producto existente
func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products
		SET sku = ?, name = ?, description = ?, category = ?, price = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		product.SKU,
		product.Name,
		product.Description,
		product.Category,
		product.Price,
		product.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.NotFoundError{Resource: "Product", ID: product.ID}
	}

	return nil
}

// Delete elimina un producto (soft delete podría implementarse)
func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.NotFoundError{Resource: "Product", ID: id}
	}

	return nil
}

// Count retorna el total de productos
func (r *ProductRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM products`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return count, nil
}
