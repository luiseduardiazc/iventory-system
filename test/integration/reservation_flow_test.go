package integration

import (
	"context"
	"testing"
	"time"

	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
	"inventory-system/internal/service"
	"inventory-system/test/testutil"
)

// TestCompleteReservationFlow prueba el flujo completo de reserva:
// 1. Crear producto
// 2. Inicializar stock
// 3. Crear reserva
// 4. Confirmar reserva
// 5. Verificar stock actualizado
func TestCompleteReservationFlow(t *testing.T) {
	// Setup database
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	ctx := context.Background()

	// Create repositories
	productRepo := repository.NewProductRepository(db)
	stockRepo := repository.NewStockRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	eventRepo := repository.NewEventRepository(db)

	// Create services
	productService := service.NewProductService(productRepo, eventRepo)
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo)
	reservationService := service.NewReservationService(reservationRepo, stockRepo, productRepo, eventRepo, db)

	// Step 1: Create a product
	product := &domain.Product{
		ID:          testutil.GenerateTestID(),
		SKU:         testutil.GenerateTestSKU(),
		Name:        "Integration Test Laptop",
		Description: "Test product for integration testing",
		Category:    "electronics",
		Price:       1299.99,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	createdProduct, err := productService.CreateProduct(ctx, product)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	t.Logf("✅ Step 1: Product created (ID: %s, SKU: %s)", createdProduct.ID, createdProduct.SKU)

	// Step 2: Initialize stock in Madrid
	storeID := "MAD-001"
	initialStock, err := stockService.InitializeStock(ctx, createdProduct.ID, storeID, 100)
	if err != nil {
		t.Fatalf("Failed to initialize stock: %v", err)
	}

	if initialStock.Quantity != 100 {
		t.Errorf("Expected quantity 100, got %d", initialStock.Quantity)
	}
	if initialStock.Reserved != 0 {
		t.Errorf("Expected reserved 0, got %d", initialStock.Reserved)
	}

	t.Logf("✅ Step 2: Stock initialized (Quantity: %d, Available: %d)",
		initialStock.Quantity, initialStock.Available())

	// Step 3: Create reservation (reserve 10 units)
	reservation := &domain.Reservation{
		ID:         testutil.GenerateTestID(),
		ProductID:  createdProduct.ID,
		StoreID:    storeID,
		CustomerID: "CUST-INT-TEST-001",
		Quantity:   10,
		Status:     domain.ReservationStatusPending,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
		CreatedAt:  time.Now(),
	}

	createdReservation, err := reservationService.CreateReservation(ctx, reservation)
	if err != nil {
		t.Fatalf("Failed to create reservation: %v", err)
	}

	t.Logf("✅ Step 3: Reservation created (ID: %s, Quantity: %d, Status: %s)",
		createdReservation.ID, createdReservation.Quantity, createdReservation.Status)

	// Verify stock after reservation
	stockAfterReservation, err := stockService.GetStockByProductAndStore(ctx, createdProduct.ID, storeID)
	if err != nil {
		t.Fatalf("Failed to get stock after reservation: %v", err)
	}

	if stockAfterReservation.Quantity != 100 {
		t.Errorf("Total quantity should remain 100, got %d", stockAfterReservation.Quantity)
	}
	if stockAfterReservation.Reserved != 10 {
		t.Errorf("Reserved should be 10, got %d", stockAfterReservation.Reserved)
	}
	if stockAfterReservation.Available() != 90 {
		t.Errorf("Available should be 90 (100-10), got %d", stockAfterReservation.Available())
	}

	t.Logf("✅ Step 3.5: Stock after reservation (Total: %d, Reserved: %d, Available: %d)",
		stockAfterReservation.Quantity, stockAfterReservation.Reserved, stockAfterReservation.Available())

	// Step 4: Confirm reservation (complete the sale)
	confirmedReservation, err := reservationService.ConfirmReservation(ctx, createdReservation.ID)
	if err != nil {
		t.Fatalf("Failed to confirm reservation: %v", err)
	}

	if confirmedReservation.Status != domain.ReservationStatusConfirmed {
		t.Errorf("Expected status CONFIRMED, got %s", confirmedReservation.Status)
	}

	t.Logf("✅ Step 4: Reservation confirmed (Status: %s)", confirmedReservation.Status)

	// Step 5: Verify final stock
	finalStock, err := stockService.GetStockByProductAndStore(ctx, createdProduct.ID, storeID)
	if err != nil {
		t.Fatalf("Failed to get final stock: %v", err)
	}

	if finalStock.Quantity != 90 {
		t.Errorf("Final quantity should be 90 (100-10), got %d", finalStock.Quantity)
	}
	if finalStock.Reserved != 0 {
		t.Errorf("Final reserved should be 0, got %d", finalStock.Reserved)
	}
	if finalStock.Available() != 90 {
		t.Errorf("Final available should be 90, got %d", finalStock.Available())
	}

	t.Logf("✅ Step 5: Final stock (Total: %d, Reserved: %d, Available: %d)",
		finalStock.Quantity, finalStock.Reserved, finalStock.Available())

	// Verify events were created
	events, err := eventRepo.GetByAggregate(ctx, createdReservation.ID)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(events) == 0 {
		t.Error("Expected at least one event to be created")
	}

	t.Logf("✅ Step 6: Events created: %d events", len(events))
	t.Logf("🎉 COMPLETE RESERVATION FLOW SUCCESS!")
}

// TestReservationExpiration prueba que las reservas expiran automáticamente
func TestReservationExpiration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	ctx := context.Background()

	productRepo := repository.NewProductRepository(db)
	stockRepo := repository.NewStockRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	eventRepo := repository.NewEventRepository(db)

	productService := service.NewProductService(productRepo, eventRepo)
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo)
	reservationService := service.NewReservationService(reservationRepo, stockRepo, productRepo, eventRepo, db)

	// Create product and stock
	product := testutil.CreateTestProduct()
	_, err := productService.CreateProduct(ctx, product)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	storeID := "BCN-001"
	_, err = stockService.InitializeStock(ctx, product.ID, storeID, 50)
	if err != nil {
		t.Fatalf("Failed to initialize stock: %v", err)
	}

	// Create expired reservation (expires in the past)
	expiredReservation := &domain.Reservation{
		ID:         testutil.GenerateTestID(),
		ProductID:  product.ID,
		StoreID:    storeID,
		CustomerID: "CUST-EXPIRED-001",
		Quantity:   5,
		Status:     domain.ReservationStatusPending,
		ExpiresAt:  time.Now().Add(-10 * time.Minute), // Expired 10 minutes ago
		CreatedAt:  time.Now().Add(-20 * time.Minute),
	}

	// We need to create it directly in the repo to bypass validation
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	if err := reservationRepo.Create(ctx, tx, expiredReservation); err != nil {
		t.Fatalf("Failed to create expired reservation: %v", err)
	}

	// Reserve stock
	if err := stockRepo.ReserveStock(ctx, tx, product.ID, storeID, 5); err != nil {
		t.Fatalf("Failed to reserve stock: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	t.Logf("✅ Created expired reservation (expires_at in the past)")

	// Process expired reservations (this is what the worker does)
	count, err := reservationService.ProcessExpiredReservations(ctx)
	if err != nil {
		t.Fatalf("Failed to process expired reservations: %v", err)
	}

	if count == 0 {
		t.Error("Expected at least 1 expired reservation to be processed")
	}

	t.Logf("✅ Processed %d expired reservations", count)

	// Verify reservation status changed to EXPIRED
	updated, err := reservationRepo.GetByID(ctx, expiredReservation.ID)
	if err != nil {
		t.Fatalf("Failed to get updated reservation: %v", err)
	}

	if updated.Status != domain.ReservationStatusExpired {
		t.Errorf("Expected status EXPIRED, got %s", updated.Status)
	}

	t.Logf("✅ Reservation status updated to %s", updated.Status)

	// Verify stock was released
	stock, err := stockRepo.GetByProductAndStore(ctx, product.ID, storeID)
	if err != nil {
		t.Fatalf("Failed to get stock: %v", err)
	}

	if stock.Reserved != 0 {
		t.Errorf("Reserved should be 0 after expiration, got %d", stock.Reserved)
	}
	if stock.Available() != 50 {
		t.Errorf("Available should be 50 after releasing reservation, got %d", stock.Available())
	}

	t.Logf("✅ Stock released (Available: %d)", stock.Available())
	t.Logf("🎉 RESERVATION EXPIRATION TEST SUCCESS!")
}

// TestStockTransfer prueba la transferencia de stock entre tiendas
func TestStockTransfer(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	ctx := context.Background()

	productRepo := repository.NewProductRepository(db)
	stockRepo := repository.NewStockRepository(db)
	eventRepo := repository.NewEventRepository(db)

	productService := service.NewProductService(productRepo, eventRepo)
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo)

	// Create product
	product := testutil.CreateTestProduct()
	_, err := productService.CreateProduct(ctx, product)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Initialize stock in two stores
	fromStore := "MAD-001"
	toStore := "BCN-001"

	stockMAD, err := stockService.InitializeStock(ctx, product.ID, fromStore, 100)
	if err != nil {
		t.Fatalf("Failed to initialize stock in Madrid: %v", err)
	}

	stockBCN, err := stockService.InitializeStock(ctx, product.ID, toStore, 50)
	if err != nil {
		t.Fatalf("Failed to initialize stock in Barcelona: %v", err)
	}

	t.Logf("✅ Initial stock - Madrid: %d, Barcelona: %d",
		stockMAD.Available(), stockBCN.Available())

	// Transfer 20 units from Madrid to Barcelona
	err = stockService.TransferStock(ctx, product.ID, fromStore, toStore, 20)
	if err != nil {
		t.Fatalf("Failed to transfer stock: %v", err)
	}

	t.Logf("✅ Transferred 20 units from Madrid to Barcelona")

	// Verify Madrid stock decreased
	madridAfter, err := stockService.GetStockByProductAndStore(ctx, product.ID, fromStore)
	if err != nil {
		t.Fatalf("Failed to get Madrid stock: %v", err)
	}

	if madridAfter.Quantity != 80 {
		t.Errorf("Madrid stock should be 80 (100-20), got %d", madridAfter.Quantity)
	}

	// Verify Barcelona stock increased
	barcelonaAfter, err := stockService.GetStockByProductAndStore(ctx, product.ID, toStore)
	if err != nil {
		t.Fatalf("Failed to get Barcelona stock: %v", err)
	}

	if barcelonaAfter.Quantity != 70 {
		t.Errorf("Barcelona stock should be 70 (50+20), got %d", barcelonaAfter.Quantity)
	}

	t.Logf("✅ Final stock - Madrid: %d, Barcelona: %d",
		madridAfter.Quantity, barcelonaAfter.Quantity)
	t.Logf("🎉 STOCK TRANSFER TEST SUCCESS!")
}

// TestConcurrentReservations prueba reservas concurrentes
func TestConcurrentReservations(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	ctx := context.Background()

	productRepo := repository.NewProductRepository(db)
	stockRepo := repository.NewStockRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	eventRepo := repository.NewEventRepository(db)

	productService := service.NewProductService(productRepo, eventRepo)
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo)
	reservationService := service.NewReservationService(reservationRepo, stockRepo, productRepo, eventRepo, db)

	// Create product with limited stock
	product := testutil.CreateTestProduct()
	_, err := productService.CreateProduct(ctx, product)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	storeID := "MAD-001"
	_, err = stockService.InitializeStock(ctx, product.ID, storeID, 10) // Only 10 units
	if err != nil {
		t.Fatalf("Failed to initialize stock: %v", err)
	}

	t.Logf("✅ Product created with 10 units in stock")

	// Try to create 3 concurrent reservations of 5 units each
	// Only 2 should succeed (10 total / 5 per reservation = 2 max)
	done := make(chan error, 3)

	for i := 0; i < 3; i++ {
		go func(index int) {
			reservation := &domain.Reservation{
				ID:         testutil.GenerateTestID(),
				ProductID:  product.ID,
				StoreID:    storeID,
				CustomerID: testutil.GenerateTestID(),
				Quantity:   5,
				Status:     domain.ReservationStatusPending,
				ExpiresAt:  time.Now().Add(15 * time.Minute),
				CreatedAt:  time.Now(),
			}

			_, err := reservationService.CreateReservation(ctx, reservation)
			done <- err
		}(i)
	}

	// Collect results
	successCount := 0
	failureCount := 0

	for i := 0; i < 3; i++ {
		err := <-done
		if err == nil {
			successCount++
			t.Logf("✅ Reservation %d succeeded", i+1)
		} else {
			failureCount++
			t.Logf("❌ Reservation %d failed (expected): %v", i+1, err)
		}
	}

	if successCount != 2 {
		t.Errorf("Expected exactly 2 successful reservations, got %d", successCount)
	}

	if failureCount != 1 {
		t.Errorf("Expected exactly 1 failed reservation, got %d", failureCount)
	}

	// Verify final stock
	finalStock, err := stockService.GetStockByProductAndStore(ctx, product.ID, storeID)
	if err != nil {
		t.Fatalf("Failed to get final stock: %v", err)
	}

	if finalStock.Reserved != 10 {
		t.Errorf("Expected 10 reserved (2 * 5), got %d", finalStock.Reserved)
	}
	if finalStock.Available() != 0 {
		t.Errorf("Expected 0 available, got %d", finalStock.Available())
	}

	t.Logf("✅ Final stock - Total: %d, Reserved: %d, Available: %d",
		finalStock.Quantity, finalStock.Reserved, finalStock.Available())
	t.Logf("🎉 CONCURRENT RESERVATIONS TEST SUCCESS!")
}
