package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"inventory-system/internal/domain"
)

// ReservationRepository maneja las operaciones de persistencia para reservas
type ReservationRepository struct {
	db *sql.DB
}

// NewReservationRepository crea una nueva instancia del repositorio
func NewReservationRepository(db *sql.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

// Create crea una nueva reserva
func (r *ReservationRepository) Create(ctx context.Context, reservation *domain.Reservation) error {
	query := `
		INSERT INTO reservations (id, product_id, store_id, customer_id, quantity, status, expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	updatedAt := reservation.CreatedAt // Por defecto, igual a created_at
	if reservation.UpdatedAt != nil {
		updatedAt = *reservation.UpdatedAt
	}

	_, err := r.db.ExecContext(ctx, query,
		reservation.ID,
		reservation.ProductID,
		reservation.StoreID,
		reservation.CustomerID,
		reservation.Quantity,
		reservation.Status,
		reservation.ExpiresAt,
		reservation.CreatedAt,
		updatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create reservation: %w", err)
	}

	return nil
}

// GetByID obtiene una reserva por su ID
func (r *ReservationRepository) GetByID(ctx context.Context, id string) (*domain.Reservation, error) {
	query := `
		SELECT id, product_id, store_id, quantity, status, expires_at, created_at, updated_at
		FROM reservations
		WHERE id = ?
	`

	var reservation domain.Reservation
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&reservation.ID,
		&reservation.ProductID,
		&reservation.StoreID,
		&reservation.Quantity,
		&reservation.Status,
		&reservation.ExpiresAt,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{
			Resource: "Reservation",
			ID:       id,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	return &reservation, nil
}

// UpdateStatus actualiza el estado de una reserva
func (r *ReservationRepository) UpdateStatus(ctx context.Context, id string, status domain.ReservationStatus) error {
	query := `
		UPDATE reservations
		SET status = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.NotFoundError{
			Resource: "Reservation",
			ID:       id,
		}
	}

	return nil
}

// GetPendingExpired obtiene todas las reservas pendientes que ya expiraron
func (r *ReservationRepository) GetPendingExpired(ctx context.Context) ([]*domain.Reservation, error) {
	query := `
		SELECT id, product_id, store_id, quantity, status, expires_at, created_at, updated_at
		FROM reservations
		WHERE status = ?
		  AND expires_at < ?
		ORDER BY expires_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, domain.ReservationStatusPending, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get expired reservations: %w", err)
	}
	defer rows.Close()

	var reservations []*domain.Reservation
	for rows.Next() {
		var reservation domain.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.ProductID,
			&reservation.StoreID,
			&reservation.Quantity,
			&reservation.Status,
			&reservation.ExpiresAt,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, &reservation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservations: %w", err)
	}

	return reservations, nil
}

// GetByProductAndStore obtiene todas las reservas de un producto en una tienda
func (r *ReservationRepository) GetByProductAndStore(ctx context.Context, productID, storeID string, status *domain.ReservationStatus) ([]*domain.Reservation, error) {
	var query string
	var args []interface{}

	if status != nil {
		query = `
			SELECT id, product_id, store_id, quantity, status, expires_at, created_at, updated_at
			FROM reservations
			WHERE product_id = ? AND store_id = ? AND status = ?
			ORDER BY created_at DESC
		`
		args = []interface{}{productID, storeID, *status}
	} else {
		query = `
			SELECT id, product_id, store_id, quantity, status, expires_at, created_at, updated_at
			FROM reservations
			WHERE product_id = ? AND store_id = ?
			ORDER BY created_at DESC
		`
		args = []interface{}{productID, storeID}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservations: %w", err)
	}
	defer rows.Close()

	var reservations []*domain.Reservation
	for rows.Next() {
		var reservation domain.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.ProductID,
			&reservation.StoreID,
			&reservation.Quantity,
			&reservation.Status,
			&reservation.ExpiresAt,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, &reservation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservations: %w", err)
	}

	return reservations, nil
}

// GetPendingByStore obtiene todas las reservas pendientes de una tienda
func (r *ReservationRepository) GetPendingByStore(ctx context.Context, storeID string) ([]*domain.Reservation, error) {
	query := `
		SELECT id, product_id, store_id, quantity, status, expires_at, created_at, updated_at
		FROM reservations
		WHERE store_id = ? AND status = ?
		ORDER BY expires_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, storeID, domain.ReservationStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending reservations: %w", err)
	}
	defer rows.Close()

	var reservations []*domain.Reservation
	for rows.Next() {
		var reservation domain.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.ProductID,
			&reservation.StoreID,
			&reservation.Quantity,
			&reservation.Status,
			&reservation.ExpiresAt,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, &reservation)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservations: %w", err)
	}

	return reservations, nil
}

// Delete elimina una reserva (usado para limpieza de reservas antiguas)
func (r *ReservationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM reservations WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete reservation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &domain.NotFoundError{
			Resource: "Reservation",
			ID:       id,
		}
	}

	return nil
}

// DeleteOldCompleted elimina reservas completadas/canceladas antiguas
func (r *ReservationRepository) DeleteOldCompleted(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `
		DELETE FROM reservations
		WHERE status IN (?, ?)
		  AND updated_at < ?
	`

	result, err := r.db.ExecContext(ctx, query,
		domain.ReservationStatusConfirmed,
		domain.ReservationStatusCancelled,
		olderThan,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to delete old reservations: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

// CountByStatus cuenta las reservas por estado
func (r *ReservationRepository) CountByStatus(ctx context.Context, status domain.ReservationStatus) (int, error) {
	query := `SELECT COUNT(*) FROM reservations WHERE status = ?`

	var count int
	err := r.db.QueryRowContext(ctx, query, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count reservations: %w", err)
	}

	return count, nil
}
