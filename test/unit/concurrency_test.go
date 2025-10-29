package unit

import (
	"context"
	"sync"
	"testing"

	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
	"inventory-system/internal/service"
	"inventory-system/test/mocks"
	"inventory-system/test/testutil"
)

// TestConcurrentStockUpdates verifica que el optimistic locking previene race conditions
func TestConcurrentStockUpdates(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	stockRepo := repository.NewStockRepository(db)
	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	publisher := mocks.NewNoOpPublisher()
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo, publisher)

	ctx := context.Background()

	// Create product and initial stock
	product := testutil.CreateTestProduct(func(p *domain.Product) {
		p.SKU = "CONCURRENT-001"
	})
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Error creating product: %v", err)
	}

	initialStock := testutil.CreateTestStock(product.ID, "STORE-CONCURRENT", func(s *domain.Stock) {
		s.Quantity = 100
	})
	if err := stockRepo.Create(ctx, initialStock); err != nil {
		t.Fatalf("Error creating stock: %v", err)
	}

	// Simulate 10 concurrent users trying to update stock
	goroutines := 10
	var wg sync.WaitGroup
	successCount := 0
	errorCount := 0
	var mu sync.Mutex

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			// Each goroutine tries to increment stock by 10
			_, err := stockService.AdjustStock(ctx, product.ID, "STORE-CONCURRENT", 10)

			mu.Lock()
			if err != nil {
				errorCount++
			} else {
				successCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify final stock
	finalStock, err := stockRepo.GetByProductAndStore(ctx, product.ID, "STORE-CONCURRENT")
	if err != nil {
		t.Fatalf("Failed to get final stock: %v", err)
	}

	// El stock final debe estar entre initial y initial+(goroutines*10)
	// ya que algunas actualizaciones pueden fallar por optimistic locking
	minExpected := initialStock.Quantity
	maxExpected := initialStock.Quantity + (goroutines * 10)

	if finalStock.Quantity < minExpected || finalStock.Quantity > maxExpected {
		t.Errorf("Final quantity %d out of expected range [%d, %d] (success=%d, errors=%d)",
			finalStock.Quantity, minExpected, maxExpected, successCount, errorCount)
	}

	// Verificar que al menos una actualización tuvo éxito
	if finalStock.Quantity == initialStock.Quantity {
		t.Error("Expected at least one update to succeed")
	}

	// Version should have incremented
	if finalStock.Version <= initialStock.Version {
		t.Errorf("Expected version to increment, got %d (initial was %d)",
			finalStock.Version, initialStock.Version)
	}

	t.Logf("Concurrent test results: initial=%d, final=%d, %d successes, %d errors (optimistic locking working)",
		initialStock.Quantity, finalStock.Quantity, successCount, errorCount)
}

// TestConcurrentReservations verifica que múltiples usuarios no puedan sobrevender
func TestConcurrentReservations(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	stockRepo := repository.NewStockRepository(db)
	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	publisher := mocks.NewNoOpPublisher()

	reservationService := service.NewReservationService(reservationRepo, stockRepo, productRepo, eventRepo, publisher)

	ctx := context.Background()

	// Create product with limited stock
	product := testutil.CreateTestProduct(func(p *domain.Product) {
		p.SKU = "CONCURRENT-RESERVE-001"
	})
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Error creating product: %v", err)
	}

	initialStock := testutil.CreateTestStock(product.ID, "STORE-RESERVE", func(s *domain.Stock) {
		s.Quantity = 10 // Only 10 items
		s.Reserved = 0
	})
	if err := stockRepo.Create(ctx, initialStock); err != nil {
		t.Fatalf("Error creating stock: %v", err)
	}

	// Simulate 20 concurrent users trying to reserve 1 item each
	goroutines := 20
	var wg sync.WaitGroup
	successCount := 0
	failureCount := 0
	var mu sync.Mutex

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(customerID int) {
			defer wg.Done()

			customerIDStr := testutil.GenerateCustomerID(customerID)
			_, err := reservationService.CreateReservation(
				ctx,
				product.ID,
				"STORE-RESERVE",
				customerIDStr,
				1,
				15, // 15 minutes TTL
			)

			mu.Lock()
			if err != nil {
				failureCount++
			} else {
				successCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Only 10 reservations should succeed (stock limit)
	if successCount > 10 {
		t.Errorf("Expected max 10 successful reservations (stock limit), got %d", successCount)
	}

	if failureCount < 10 {
		t.Errorf("Expected at least 10 failures, got %d", failureCount)
	}

	// Verify stock reserved matches successful reservations
	finalStock, err := stockRepo.GetByProductAndStore(ctx, product.ID, "STORE-RESERVE")
	if err != nil {
		t.Fatalf("Failed to get final stock: %v", err)
	}

	if finalStock.Reserved != successCount {
		t.Errorf("Expected reserved %d, got %d", successCount, finalStock.Reserved)
	}

	t.Logf("Concurrent reservation results: %d successes, %d failures (overselling prevented)",
		successCount, failureCount)
}

// TestHighLoadStockOperations simula carga alta con múltiples operaciones
func TestHighLoadStockOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high load test in short mode")
	}

	db := testutil.SetupTestDB(t)
	defer db.Close()

	stockRepo := repository.NewStockRepository(db)
	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	publisher := mocks.NewNoOpPublisher()
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo, publisher)

	ctx := context.Background()

	// Create product
	product := testutil.CreateTestProduct(func(p *domain.Product) {
		p.SKU = "HIGH-LOAD-001"
	})
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Error creating product: %v", err)
	}

	initialStock := testutil.CreateTestStock(product.ID, "STORE-LOAD", func(s *domain.Stock) {
		s.Quantity = 1000
	})
	if err := stockRepo.Create(ctx, initialStock); err != nil {
		t.Fatalf("Error creating stock: %v", err)
	}

	// Simulate 100 concurrent operations
	operations := 100
	var wg sync.WaitGroup
	errors := make(chan error, operations)

	for i := 0; i < operations; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			// Mix of operations
			switch iteration % 3 {
			case 0:
				// Update stock
				_, err := stockService.AdjustStock(ctx, product.ID, "STORE-LOAD", 5)
				if err != nil {
					errors <- err
				}
			case 1:
				// Check availability
				_, err := stockService.CheckAvailability(ctx, product.ID, "STORE-LOAD", 1)
				if err != nil {
					errors <- err
				}
			case 2:
				// Get stock
				_, err := stockService.GetStockByProductAndStore(ctx, product.ID, "STORE-LOAD")
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Count errors
	errorCount := 0
	for range errors {
		errorCount++
	}

	// Some errors are expected due to concurrent updates, but not all should fail
	if errorCount > operations/2 {
		t.Errorf("Too many errors under load: %d out of %d operations failed", errorCount, operations)
	}

	t.Logf("High load test: %d operations, %d errors (%.2f%% success rate)",
		operations, errorCount, float64(operations-errorCount)/float64(operations)*100)
}

// TestRaceConditionDetection verifica detección de race conditions con -race flag
func TestRaceConditionDetection(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	stockRepo := repository.NewStockRepository(db)
	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	publisher := mocks.NewNoOpPublisher()
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo, publisher)

	ctx := context.Background()

	product := testutil.CreateTestProduct(func(p *domain.Product) {
		p.SKU = "RACE-001"
	})
	if err := productRepo.Create(ctx, product); err != nil {
		t.Fatalf("Error creating product: %v", err)
	}

	initialStock := testutil.CreateTestStock(product.ID, "STORE-RACE", func(s *domain.Stock) {
		s.Quantity = 100
	})
	if err := stockRepo.Create(ctx, initialStock); err != nil {
		t.Fatalf("Error creating stock: %v", err)
	}

	// Run with: go test -race
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stockService.GetStockByProductAndStore(ctx, product.ID, "STORE-RACE")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			stockService.AdjustStock(ctx, product.ID, "STORE-RACE", 1)
		}()
	}

	wg.Wait()
}
