package unit

import (
	"context"
	"testing"
	"time"

	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
	"inventory-system/internal/service"
	"inventory-system/test/mocks"
	"inventory-system/test/testutil"
)

func TestReservationService_ExpirationWorker(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	reservationRepo := repository.NewReservationRepository(db)
	stockRepo := repository.NewStockRepository(db)
	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	publisher := mocks.NewNoOpPublisher()

	reservationService := service.NewReservationService(
		reservationRepo,
		stockRepo,
		productRepo,
		eventRepo,
		publisher,
	)

	ctx := context.Background()

	t.Run("ExpirationWorker_ExpiresOldReservations", func(t *testing.T) {
		// Create product and stock
		product := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "EXPIRE-001"
		})
		if err := productRepo.Create(ctx, product); err != nil {
			t.Fatalf("Error creating product: %v", err)
		}

		stock := testutil.CreateTestStock(product.ID, "STORE-EXPIRE", func(s *domain.Stock) {
			s.Quantity = 100
			s.Reserved = 5
		})
		if err := stockRepo.Create(ctx, stock); err != nil {
			t.Fatalf("Error creating stock: %v", err)
		}

		// Create reservation that already expired
		now := time.Now()
		expiredReservation := &domain.Reservation{
			ID:         testutil.GenerateID(),
			ProductID:  product.ID,
			StoreID:    "STORE-EXPIRE",
			CustomerID: "customer-1",
			Quantity:   5,
			Status:     domain.ReservationStatusPending,
			ExpiresAt:  time.Now().Add(-1 * time.Minute), // Already expired
			CreatedAt:  time.Now().Add(-10 * time.Minute),
			UpdatedAt:  &now,
		}
		reservationRepo.Create(ctx, expiredReservation)

		// Run expiration worker once
		processed, err := reservationService.ProcessExpiredReservations(ctx)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if processed != 1 {
			t.Errorf("Expected 1 reservation processed, got %d", processed)
		}

		// Verify reservation was expired
		updated, err := reservationRepo.GetByID(ctx, expiredReservation.ID)
		if err != nil {
			t.Fatalf("Failed to get reservation: %v", err)
		}

		if updated.Status != domain.ReservationStatusExpired {
			t.Errorf("Expected status EXPIRED, got %s", updated.Status)
		}

		// Verify stock was released
		finalStock, err := stockRepo.GetByProductAndStore(ctx, product.ID, "STORE-EXPIRE")
		if err != nil {
			t.Fatalf("Failed to get stock: %v", err)
		}

		if finalStock.Reserved != 0 {
			t.Errorf("Expected reserved 0 (stock released), got %d", finalStock.Reserved)
		}
	})

	t.Run("ExpirationWorker_IgnoresActiveReservations", func(t *testing.T) {
		// Create product and stock
		product := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "EXPIRE-002"
		})
		if err := productRepo.Create(ctx, product); err != nil {
			t.Fatalf("Error creating product: %v", err)
		}

		stock := testutil.CreateTestStock(product.ID, "STORE-EXPIRE-2", func(s *domain.Stock) {
			s.Quantity = 100
			s.Reserved = 10
		})
		if err := stockRepo.Create(ctx, stock); err != nil {
			t.Fatalf("Error creating stock: %v", err)
		}

		// Create active reservation (not expired)
		now := time.Now()
		activeReservation := &domain.Reservation{
			ID:         testutil.GenerateID(),
			ProductID:  product.ID,
			StoreID:    "STORE-EXPIRE-2",
			CustomerID: "customer-2",
			Quantity:   10,
			Status:     domain.ReservationStatusPending,
			ExpiresAt:  time.Now().Add(10 * time.Minute), // Still valid
			CreatedAt:  time.Now(),
			UpdatedAt:  &now,
		}
		reservationRepo.Create(ctx, activeReservation)

		// Run expiration worker
		processed, err := reservationService.ProcessExpiredReservations(ctx)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if processed != 0 {
			t.Errorf("Expected 0 reservations processed, got %d", processed)
		}

		// Verify reservation is still PENDING
		updated, err := reservationRepo.GetByID(ctx, activeReservation.ID)
		if err != nil {
			t.Fatalf("Failed to get reservation: %v", err)
		}

		if updated.Status != domain.ReservationStatusPending {
			t.Errorf("Expected status PENDING, got %s", updated.Status)
		}
	})

	t.Run("ExpirationWorker_IgnoresConfirmedReservations", func(t *testing.T) {
		// Create product and stock
		product := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "EXPIRE-003"
		})
		if err := productRepo.Create(ctx, product); err != nil {
			t.Fatalf("Error creating product: %v", err)
		}

		stock := testutil.CreateTestStock(product.ID, "STORE-EXPIRE-3", func(s *domain.Stock) {
			s.Quantity = 100
			s.Reserved = 0
		})
		if err := stockRepo.Create(ctx, stock); err != nil {
			t.Fatalf("Error creating stock: %v", err)
		}

		// Create confirmed reservation (already processed)
		now := time.Now()
		confirmedReservation := &domain.Reservation{
			ID:         testutil.GenerateID(),
			ProductID:  product.ID,
			StoreID:    "STORE-EXPIRE-3",
			CustomerID: "customer-3",
			Quantity:   15,
			Status:     domain.ReservationStatusConfirmed,
			ExpiresAt:  time.Now().Add(-5 * time.Minute), // Expired time, but already confirmed
			CreatedAt:  time.Now().Add(-20 * time.Minute),
			UpdatedAt:  &now,
		}
		reservationRepo.Create(ctx, confirmedReservation)

		// Run expiration worker
		processed, err := reservationService.ProcessExpiredReservations(ctx)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if processed != 0 {
			t.Errorf("Expected 0 reservations processed (confirmed ignored), got %d", processed)
		}

		// Verify reservation is still CONFIRMED (not changed to EXPIRED)
		updated, err := reservationRepo.GetByID(ctx, confirmedReservation.ID)
		if err != nil {
			t.Fatalf("Failed to get reservation: %v", err)
		}

		if updated.Status != domain.ReservationStatusConfirmed {
			t.Errorf("Expected status CONFIRMED, got %s", updated.Status)
		}
	})
}

func TestEventSyncService_SyncWorker(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	eventRepo := repository.NewEventRepository(db)
	syncService := service.NewEventSyncService(eventRepo)

	ctx := context.Background()

	t.Run("SyncWorker_RetryUnsyncedEvents", func(t *testing.T) {
		// Create unsynced events
		event1 := &domain.Event{
			ID:            testutil.GenerateID(),
			EventType:     "stock.created",
			AggregateID:   "product-1",
			AggregateType: "stock",
			StoreID:       "STORE-SYNC-001",
			Payload:       `{"quantity": 100}`,
			Synced:        false, // Not synced yet
			CreatedAt:     time.Now(),
		}
		eventRepo.Save(ctx, event1)

		event2 := &domain.Event{
			ID:            testutil.GenerateID(),
			EventType:     "reservation.created",
			AggregateID:   "reservation-1",
			AggregateType: "reservation",
			StoreID:       "STORE-SYNC-001",
			Payload:       `{"quantity": 5}`,
			Synced:        false,
			CreatedAt:     time.Now(),
		}
		eventRepo.Save(ctx, event2)

		// Run sync worker with batch size
		synced, err := syncService.SyncPendingEvents(ctx, 10)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if synced != 2 {
			t.Errorf("Expected 2 events synced, got %d", synced)
		}

		// Verify events are marked as synced
		updatedEvent1, _ := eventRepo.GetByID(ctx, event1.ID)
		if !updatedEvent1.Synced {
			t.Error("Expected event1 to be marked as synced")
		}

		updatedEvent2, _ := eventRepo.GetByID(ctx, event2.ID)
		if !updatedEvent2.Synced {
			t.Error("Expected event2 to be marked as synced")
		}
	})

	t.Run("SyncWorker_IgnoresAlreadySyncedEvents", func(t *testing.T) {
		// Create already synced event
		syncedEvent := &domain.Event{
			ID:            testutil.GenerateID(),
			EventType:     "stock.updated",
			AggregateID:   "product-2",
			AggregateType: "stock",
			StoreID:       "STORE-SYNC-002",
			Payload:       `{"quantity": 200}`,
			Synced:        true, // Already synced
			CreatedAt:     time.Now(),
		}
		eventRepo.Save(ctx, syncedEvent)

		// Run sync worker with batch size
		synced, err := syncService.SyncPendingEvents(ctx, 10)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Should not re-sync already synced events
		if synced > 0 {
			t.Errorf("Expected 0 events synced (already synced), got %d", synced)
		}
	})
}
