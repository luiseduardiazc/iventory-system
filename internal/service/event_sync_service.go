package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
)

// EventPublisher interface para publicar eventos (inyección de dependencia)
type EventPublisher interface {
	Publish(ctx context.Context, event *domain.Event) error
	PublishBatch(ctx context.Context, events []*domain.Event) error
	Close() error
}

// EventSyncService maneja la sincronización de eventos con message brokers.
// Actúa como mecanismo de RETRY para eventos que fallaron en la publicación inicial.
type EventSyncService struct {
	eventRepo *repository.EventRepository
	publisher EventPublisher // Re-intenta publicar eventos pendientes
}

// NewEventSyncService crea una nueva instancia del servicio
func NewEventSyncService(eventRepo *repository.EventRepository, publisher EventPublisher) *EventSyncService {
	return &EventSyncService{
		eventRepo: eventRepo,
		publisher: publisher,
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
	failedCount := 0
	eventIDs := make([]string, 0, len(events))

	// RE-INTENTAR publicación de eventos pendientes
	for _, event := range events {
		// Intenta publicar en el broker (Redis/Kafka)
		err := s.publisher.Publish(ctx, event)
		if err != nil {
			log.Printf("⚠️  Failed to sync event %s: %v (will retry later)", event.ID, err)
			failedCount++
			continue // No marcar como sincronizado si falla
		}

		// Solo marcar como sincronizado si la publicación fue exitosa
		eventIDs = append(eventIDs, event.ID)
		syncedCount++
	}

	// Marcar como sincronizados solo los exitosos
	if len(eventIDs) > 0 {
		err = s.eventRepo.MarkMultipleAsSynced(ctx, eventIDs)
		if err != nil {
			return syncedCount, fmt.Errorf("failed to mark events as synced: %w", err)
		}
		log.Printf("✅ Successfully synced %d events (failed: %d)", syncedCount, failedCount)
	} else if failedCount > 0 {
		log.Printf("⚠️  All %d events failed to sync (will retry in next cycle)", failedCount)
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
