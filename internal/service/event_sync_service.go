package service

import (
	"context"
	"fmt"
	"time"

	"inventory-system/internal/repository"
)

// EventSyncService maneja la sincronización de eventos con message brokers
type EventSyncService struct {
	eventRepo *repository.EventRepository
}

// NewEventSyncService crea una nueva instancia del servicio
func NewEventSyncService(eventRepo *repository.EventRepository) *EventSyncService {
	return &EventSyncService{
		eventRepo: eventRepo,
	}
}

// SyncPendingEvents sincroniza eventos pendientes con el sistema central
func (s *EventSyncService) SyncPendingEvents(ctx context.Context, batchSize int) (int, error) {
	if batchSize <= 0 {
		batchSize = 100
	}

	// Obtener eventos pendientes
	events, err := s.eventRepo.GetPendingEvents(ctx, batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending events: %w", err)
	}

	if len(events) == 0 {
		return 0, nil
	}

	syncedCount := 0
	eventIDs := make([]string, 0, len(events))

	for _, event := range events {
		// NOTA: En producción, EventPublisher maneja la publicación a brokers.
		// Este servicio solo gestiona el estado de sincronización en BD.
		// Por ahora simulamos el éxito para el prototipo.
		// Implementación de producción:
		// err := s.publisher.Publish(ctx, event)
		// if err != nil {
		//     log.Printf("Error publishing event %s: %v", event.ID, err)
		//     continue
		// }

		eventIDs = append(eventIDs, event.ID)
		syncedCount++
	}

	// Marcar como sincronizados
	if len(eventIDs) > 0 {
		err = s.eventRepo.MarkMultipleAsSynced(ctx, eventIDs)
		if err != nil {
			return syncedCount, fmt.Errorf("failed to mark events as synced: %w", err)
		}
	}

	return syncedCount, nil
}

// GetPendingEventsCount retorna la cantidad de eventos pendientes
func (s *EventSyncService) GetPendingEventsCount(ctx context.Context) (int, error) {
	return s.eventRepo.CountPending(ctx)
}

// CleanupOldEvents limpia eventos sincronizados antiguos
func (s *EventSyncService) CleanupOldEvents(ctx context.Context, daysOld int) (int64, error) {
	if daysOld <= 0 {
		daysOld = 30 // Default: 30 días
	}

	olderThan := time.Now().AddDate(0, 0, -daysOld)
	return s.eventRepo.DeleteOldSynced(ctx, olderThan)
}

// GetEventsByStore obtiene eventos de una tienda
func (s *EventSyncService) GetEventsByStore(ctx context.Context, storeID string, limit, offset int) (interface{}, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	return s.eventRepo.GetByStore(ctx, storeID, limit, offset)
}

// GetEventsByProduct obtiene el historial de eventos de un producto
func (s *EventSyncService) GetEventsByProduct(ctx context.Context, productID string) (interface{}, error) {
	return s.eventRepo.GetByAggregateID(ctx, productID)
}
