package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
	"inventory-system/internal/service"
	"inventory-system/test/mocks"
	"inventory-system/test/testutil"
)

// Test para verificar que EventSyncService re-intenta publicar eventos fallidos
func TestEventSyncService_RetryMechanism(t *testing.T) {
	t.Run("Retry_PublishFailedEvents_Success", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		defer db.Close()

		eventRepo := repository.NewEventRepository(db)
		ctx := context.Background()

		// Crear un MockPublisher que falla la primera vez y tiene éxito la segunda
		failingPublisher := &FailingPublisher{
			failCount: 1, // Falla 1 vez, luego tiene éxito
		}

		syncService := service.NewEventSyncService(eventRepo, failingPublisher)

		// Crear un evento no sincronizado
		event := &domain.Event{
			ID:            testutil.GenerateID(),
			EventType:     "stock.updated",
			AggregateID:   "product-retry-1",
			AggregateType: "stock",
			StoreID:       "STORE-RETRY-001",
			Payload:       `{"quantity": 50}`,
			Synced:        false,
			CreatedAt:     time.Now(),
		}
		eventRepo.Save(ctx, event)

		// Primer intento: DEBE FALLAR
		syncedCount, err := syncService.SyncPendingEvents(ctx, 10)
		if err != nil {
			t.Errorf("SyncPendingEvents should not return error: %v", err)
		}
		if syncedCount != 0 {
			t.Errorf("Expected 0 synced events on first attempt (should fail), got %d", syncedCount)
		}

		// Verificar que el evento NO se marcó como sincronizado
		events, _ := eventRepo.GetPendingEvents(ctx, 10)
		if len(events) != 1 {
			t.Errorf("Expected 1 pending event after failed sync, got %d", len(events))
		}

		// Segundo intento: DEBE TENER ÉXITO
		syncedCount, err = syncService.SyncPendingEvents(ctx, 10)
		if err != nil {
			t.Errorf("SyncPendingEvents should not return error: %v", err)
		}
		if syncedCount != 1 {
			t.Errorf("Expected 1 synced event on second attempt (should succeed), got %d", syncedCount)
		}

		// Verificar que ya NO hay eventos pendientes
		events, _ = eventRepo.GetPendingEvents(ctx, 10)
		if len(events) != 0 {
			t.Errorf("Expected 0 pending events after successful sync, got %d", len(events))
		}
	})

	t.Run("Retry_PartialFailure_OnlySuccessfulMarked", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		defer db.Close()

		eventRepo := repository.NewEventRepository(db)
		ctx := context.Background()

		// Publisher que falla selectivamente
		selectivePublisher := &SelectiveFailPublisher{
			failIDs: map[string]bool{"fail-1": true, "fail-2": true},
		}

		syncService := service.NewEventSyncService(eventRepo, selectivePublisher)

		// Crear 4 eventos: 2 que fallarán, 2 que tendrán éxito
		events := []*domain.Event{
			{
				ID:            "fail-1",
				EventType:     "stock.updated",
				AggregateID:   "product-1",
				AggregateType: "stock",
				StoreID:       "STORE-001",
				Payload:       `{"quantity": 10}`,
				Synced:        false,
				CreatedAt:     time.Now(),
			},
			{
				ID:            "success-1",
				EventType:     "stock.updated",
				AggregateID:   "product-2",
				AggregateType: "stock",
				StoreID:       "STORE-001",
				Payload:       `{"quantity": 20}`,
				Synced:        false,
				CreatedAt:     time.Now(),
			},
			{
				ID:            "fail-2",
				EventType:     "reservation.created",
				AggregateID:   "reservation-1",
				AggregateType: "reservation",
				StoreID:       "STORE-001",
				Payload:       `{"quantity": 5}`,
				Synced:        false,
				CreatedAt:     time.Now(),
			},
			{
				ID:            "success-2",
				EventType:     "reservation.created",
				AggregateID:   "reservation-2",
				AggregateType: "reservation",
				StoreID:       "STORE-001",
				Payload:       `{"quantity": 3}`,
				Synced:        false,
				CreatedAt:     time.Now(),
			},
		}

		for _, evt := range events {
			eventRepo.Save(ctx, evt)
		}

		// Intentar sincronizar
		syncedCount, err := syncService.SyncPendingEvents(ctx, 10)
		if err != nil {
			t.Errorf("SyncPendingEvents should not return error: %v", err)
		}

		if syncedCount != 2 {
			t.Errorf("Expected 2 synced events (only successful ones), got %d", syncedCount)
		}

		// Verificar que SOLO quedan pendientes los que fallaron
		pendingEvents, _ := eventRepo.GetPendingEvents(ctx, 10)
		if len(pendingEvents) != 2 {
			t.Errorf("Expected 2 pending events (failed ones), got %d", len(pendingEvents))
		}

		// Verificar que los pendientes son los correctos
		pendingIDs := make(map[string]bool)
		for _, evt := range pendingEvents {
			pendingIDs[evt.ID] = true
		}

		if !pendingIDs["fail-1"] || !pendingIDs["fail-2"] {
			t.Errorf("Expected fail-1 and fail-2 to still be pending")
		}
	})

	t.Run("Retry_NoOpPublisher_AlwaysSucceeds", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		defer db.Close()

		eventRepo := repository.NewEventRepository(db)
		ctx := context.Background()

		// NoOpPublisher nunca falla
		noOpPublisher := mocks.NewNoOpPublisher()
		syncService := service.NewEventSyncService(eventRepo, noOpPublisher)

		// Crear evento
		event := &domain.Event{
			ID:            testutil.GenerateID(),
			EventType:     "stock.created",
			AggregateID:   "product-noop",
			AggregateType: "stock",
			StoreID:       "STORE-NOOP-001",
			Payload:       `{"quantity": 100}`,
			Synced:        false,
			CreatedAt:     time.Now(),
		}
		eventRepo.Save(ctx, event)

		// Debe sincronizar exitosamente en el primer intento
		syncedCount, err := syncService.SyncPendingEvents(ctx, 10)
		if err != nil {
			t.Errorf("SyncPendingEvents should not return error: %v", err)
		}
		if syncedCount != 1 {
			t.Errorf("Expected 1 synced event with NoOpPublisher, got %d", syncedCount)
		}

		// No deben quedar eventos pendientes
		events, _ := eventRepo.GetPendingEvents(ctx, 10)
		if len(events) != 0 {
			t.Errorf("Expected 0 pending events with NoOpPublisher, got %d", len(events))
		}
	})
}

// ========== Mock Publishers para Testing ==========

// FailingPublisher simula un publisher que falla N veces antes de tener éxito
type FailingPublisher struct {
	failCount    int // Cuántas veces debe fallar
	attemptCount int // Contador de intentos
}

func (p *FailingPublisher) Publish(ctx context.Context, event *domain.Event) error {
	p.attemptCount++
	if p.attemptCount <= p.failCount {
		return errors.New("simulated publish failure (Redis down)")
	}
	return nil // Éxito después de N fallos
}

func (p *FailingPublisher) PublishBatch(ctx context.Context, events []*domain.Event) error {
	return errors.New("not implemented")
}

func (p *FailingPublisher) Close() error {
	return nil
}

// SelectiveFailPublisher falla solo para ciertos event IDs
type SelectiveFailPublisher struct {
	failIDs map[string]bool
}

func (p *SelectiveFailPublisher) Publish(ctx context.Context, event *domain.Event) error {
	if p.failIDs[event.ID] {
		return errors.New("simulated failure for event " + event.ID)
	}
	return nil
}

func (p *SelectiveFailPublisher) PublishBatch(ctx context.Context, events []*domain.Event) error {
	return errors.New("not implemented")
}

func (p *SelectiveFailPublisher) Close() error {
	return nil
}
