package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	
	"inventory-system/internal/domain"
)

// EventRepository maneja las operaciones de persistencia para eventos
type EventRepository struct {
	db *sql.DB
}

// NewEventRepository crea una nueva instancia del repositorio
func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

// Save guarda un nuevo evento
func (r *EventRepository) Save(ctx context.Context, event *domain.Event) error {
	query := `
		INSERT INTO events (id, event_type, aggregate_id, aggregate_type, store_id, payload, created_at, synced)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		event.ID,
		event.EventType,
		event.AggregateID,
		event.AggregateType,
		event.StoreID,
		event.Payload,
		event.CreatedAt,
		event.Synced,
	)
	
	if err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}
	
	return nil
}

// GetByID obtiene un evento por su ID
func (r *EventRepository) GetByID(ctx context.Context, id string) (*domain.Event, error) {
	query := `
		SELECT id, event_type, aggregate_id, aggregate_type, store_id, payload, created_at, synced, synced_at
		FROM events
		WHERE id = ?
	`
	
	var event domain.Event
	var syncedAt sql.NullTime
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.EventType,
		&event.AggregateID,
		&event.AggregateType,
		&event.StoreID,
		&event.Payload,
		&event.CreatedAt,
		&event.Synced,
		&syncedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{
			Resource: "Event",
			ID:       id,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	
	if syncedAt.Valid {
		event.SyncedAt = &syncedAt.Time
	}
	
	return &event, nil
}

// GetPendingEvents obtiene todos los eventos que no han sido sincronizados
func (r *EventRepository) GetPendingEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	query := `
		SELECT id, event_type, aggregate_id, aggregate_type, store_id, payload, created_at, synced, synced_at
		FROM events
		WHERE synced = false
		ORDER BY created_at ASC
		LIMIT ?
	`
	
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending events: %w", err)
	}
	defer rows.Close()
	
	var events []*domain.Event
	for rows.Next() {
		var event domain.Event
		var syncedAt sql.NullTime
		
		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.AggregateID,
			&event.AggregateType,
			&event.StoreID,
			&event.Payload,
			&event.CreatedAt,
			&event.Synced,
			&syncedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		
		if syncedAt.Valid {
			event.SyncedAt = &syncedAt.Time
		}
		
		events = append(events, &event)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}
	
	return events, nil
}

// GetByAggregateID obtiene todos los eventos de un agregado específico (Product o Stock)
func (r *EventRepository) GetByAggregateID(ctx context.Context, aggregateID string) ([]*domain.Event, error) {
	query := `
		SELECT id, event_type, aggregate_id, aggregate_type, store_id, payload, created_at, synced, synced_at
		FROM events
		WHERE aggregate_id = ?
		ORDER BY created_at ASC
	`
	
	rows, err := r.db.QueryContext(ctx, query, aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by aggregate: %w", err)
	}
	defer rows.Close()
	
	var events []*domain.Event
	for rows.Next() {
		var event domain.Event
		var syncedAt sql.NullTime
		
		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.AggregateID,
			&event.AggregateType,
			&event.StoreID,
			&event.Payload,
			&event.CreatedAt,
			&event.Synced,
			&syncedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		
		if syncedAt.Valid {
			event.SyncedAt = &syncedAt.Time
		}
		
		events = append(events, &event)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}
	
	return events, nil
}

// GetByStore obtiene todos los eventos de una tienda
func (r *EventRepository) GetByStore(ctx context.Context, storeID string, limit, offset int) ([]*domain.Event, error) {
	query := `
		SELECT id, event_type, aggregate_id, aggregate_type, store_id, payload, created_at, synced, synced_at
		FROM events
		WHERE store_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.db.QueryContext(ctx, query, storeID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by store: %w", err)
	}
	defer rows.Close()
	
	var events []*domain.Event
	for rows.Next() {
		var event domain.Event
		var syncedAt sql.NullTime
		
		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.AggregateID,
			&event.AggregateType,
			&event.StoreID,
			&event.Payload,
			&event.CreatedAt,
			&event.Synced,
			&syncedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		
		if syncedAt.Valid {
			event.SyncedAt = &syncedAt.Time
		}
		
		events = append(events, &event)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}
	
	return events, nil
}

// MarkAsSynced marca un evento como sincronizado
func (r *EventRepository) MarkAsSynced(ctx context.Context, eventID string) error {
	query := `
		UPDATE events
		SET synced = true,
		    synced_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	
	result, err := r.db.ExecContext(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to mark event as synced: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return &domain.NotFoundError{
			Resource: "Event",
			ID:       eventID,
		}
	}
	
	return nil
}

// MarkMultipleAsSynced marca múltiples eventos como sincronizados
func (r *EventRepository) MarkMultipleAsSynced(ctx context.Context, eventIDs []string) error {
	if len(eventIDs) == 0 {
		return nil
	}
	
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	query := `
		UPDATE events
		SET synced = true,
		    synced_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	
	for _, eventID := range eventIDs {
		_, err = stmt.ExecContext(ctx, eventID)
		if err != nil {
			return fmt.Errorf("failed to mark event %s as synced: %w", eventID, err)
		}
	}
	
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// DeleteOldSynced elimina eventos sincronizados antiguos (limpieza periódica)
func (r *EventRepository) DeleteOldSynced(ctx context.Context, olderThan time.Time) (int64, error) {
	query := `
		DELETE FROM events
		WHERE synced = true
		  AND synced_at < ?
	`
	
	result, err := r.db.ExecContext(ctx, query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old events: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	return rowsAffected, nil
}

// CountPending cuenta eventos pendientes de sincronización
func (r *EventRepository) CountPending(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM events WHERE synced = false`
	
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count pending events: %w", err)
	}
	
	return count, nil
}

// GetEventsByType obtiene eventos de un tipo específico
func (r *EventRepository) GetEventsByType(ctx context.Context, eventType string, limit, offset int) ([]*domain.Event, error) {
	query := `
		SELECT id, event_type, aggregate_id, aggregate_type, store_id, payload, created_at, synced, synced_at
		FROM events
		WHERE event_type = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.db.QueryContext(ctx, query, eventType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by type: %w", err)
	}
	defer rows.Close()
	
	var events []*domain.Event
	for rows.Next() {
		var event domain.Event
		var syncedAt sql.NullTime
		
		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.AggregateID,
			&event.AggregateType,
			&event.StoreID,
			&event.Payload,
			&event.CreatedAt,
			&event.Synced,
			&syncedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		
		if syncedAt.Valid {
			event.SyncedAt = &syncedAt.Time
		}
		
		events = append(events, &event)
	}
	
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}
	
	return events, nil
}
