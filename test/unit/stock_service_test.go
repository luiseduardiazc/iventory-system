package unit

import (
	"context"
	"testing"

	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
	"inventory-system/internal/service"
	"inventory-system/test/mocks"
	"inventory-system/test/testutil"
)

func TestStockService_InitializeStock(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	stockRepo := repository.NewStockRepository(db)
	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	publisher := mocks.NewNoOpPublisher()
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo, publisher)

	ctx := context.Background()

	t.Run("InitializeStock_Success", func(t *testing.T) {
		// Create product first
		product := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "STOCK-001"
		})
		if err := productRepo.Create(ctx, product); err != nil {
			t.Fatalf("Error creating product: %v", err)
		}

		created, err := stockService.InitializeStock(ctx, product.ID, "TEST-STORE-001", 100)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if created.ID == "" {
			t.Error("Expected stock ID to be generated")
		}

		if created.Quantity != 100 {
			t.Errorf("Expected quantity 100, got %d", created.Quantity)
		}

		if created.Version != 1 {
			t.Errorf("Expected version 1, got %d", created.Version)
		}
	})

	t.Run("InitializeStock_InvalidProduct", func(t *testing.T) {
		_, err := stockService.InitializeStock(ctx, "non-existent-product-id", "TEST-STORE-001", 50)
		if err == nil {
			t.Error("Expected error for non-existent product, got nil")
		}
	})

	t.Run("InitializeStock_NegativeQuantity", func(t *testing.T) {
		product := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "STOCK-002"
		})
		if err := productRepo.Create(ctx, product); err != nil {
			t.Fatalf("Error creating product: %v", err)
		}

		_, err := stockService.InitializeStock(ctx, product.ID, "TEST-STORE-001", -10)
		if err == nil {
			t.Error("Expected validation error for negative quantity, got nil")
		}
	})
}

func TestStockService_UpdateStock(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	stockRepo := repository.NewStockRepository(db)
	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	publisher := mocks.NewNoOpPublisher()
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo, publisher)

	ctx := context.Background()

	t.Run("UpdateStock_Success", func(t *testing.T) {
		// Create product and stock
		product := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "UPDATE-STOCK-001"
		})
		if err := productRepo.Create(ctx, product); err != nil {
			t.Fatalf("Error creating product: %v", err)
		}

		initialStock := testutil.CreateTestStock(product.ID, "STORE-001", func(s *domain.Stock) {
			s.Quantity = 100
		})
		if err := stockRepo.Create(ctx, initialStock); err != nil {
			t.Fatalf("Error creating stock: %v", err)
		}

		// Update stock
		updated, err := stockService.UpdateStock(ctx, product.ID, "STORE-001", 150)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if updated.Quantity != 150 {
			t.Errorf("Expected quantity 150, got %d", updated.Quantity)
		}

		if updated.Version != initialStock.Version+1 {
			t.Errorf("Expected version to increment, got %d", updated.Version)
		}
	})

	t.Run("UpdateStock_LessThanReserved", func(t *testing.T) {
		// Create product and stock with reservations
		product := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "UPDATE-STOCK-002"
		})
		if err := productRepo.Create(ctx, product); err != nil {
			t.Fatalf("Error creating product: %v", err)
		}

		stock := testutil.CreateTestStock(product.ID, "STORE-002", func(s *domain.Stock) {
			s.Quantity = 100
			s.Reserved = 30
		})
		if err := stockRepo.Create(ctx, stock); err != nil {
			t.Fatalf("Error creating stock: %v", err)
		}

		// Update reserved (UpdateQuantity doesn't change quantity directly in this test)
		stock.Reserved = 30
		if err := stockRepo.UpdateQuantity(ctx, stock); err != nil {
			t.Fatalf("Error updating stock: %v", err)
		}

		// Try to set quantity below reserved
		_, err := stockService.UpdateStock(ctx, product.ID, "STORE-002", 20)
		if err == nil {
			t.Error("Expected error when setting quantity below reserved, got nil")
		}

		if _, ok := err.(*domain.ValidationError); !ok {
			t.Errorf("Expected ValidationError, got %T", err)
		}
	})

	t.Run("UpdateStock_OptimisticLocking", func(t *testing.T) {
		// Create product and stock
		product := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "UPDATE-STOCK-003"
		})
		if err := productRepo.Create(ctx, product); err != nil {
			t.Fatalf("Error creating product: %v", err)
		}

		stock := testutil.CreateTestStock(product.ID, "STORE-003", func(s *domain.Stock) {
			s.Quantity = 100
		})
		if err := stockRepo.Create(ctx, stock); err != nil {
			t.Fatalf("Error creating stock: %v", err)
		}

		// Simulate concurrent update by manually incrementing version
		stock.Version++
		stockRepo.UpdateQuantity(ctx, stock)

		// Try to update with old version (should fail)
		oldStock, _ := stockRepo.GetByProductAndStore(ctx, product.ID, "STORE-003")
		oldStock.Quantity = 200
		oldStock.Version-- // Use old version

		err := stockRepo.UpdateQuantity(ctx, oldStock)
		if err == nil {
			t.Error("Expected optimistic locking error, got nil")
		}
	})
}

func TestStockService_AdjustStock(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	stockRepo := repository.NewStockRepository(db)
	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	publisher := mocks.NewNoOpPublisher()
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo, publisher)

	ctx := context.Background()

	t.Run("AdjustStock_Increment", func(t *testing.T) {
		product := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "ADJUST-001"
		})
		if err := productRepo.Create(ctx, product); err != nil {
			t.Fatalf("Error creating product: %v", err)
		}

		stock := testutil.CreateTestStock(product.ID, "STORE-ADJ-001", func(s *domain.Stock) {
			s.Quantity = 100
		})
		if err := stockRepo.Create(ctx, stock); err != nil {
			t.Fatalf("Error creating stock: %v", err)
		}

		adjusted, err := stockService.AdjustStock(ctx, product.ID, "STORE-ADJ-001", 50)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if adjusted.Quantity != 150 {
			t.Errorf("Expected quantity 150, got %d", adjusted.Quantity)
		}
	})

	t.Run("AdjustStock_Decrement", func(t *testing.T) {
		product := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "ADJUST-002"
		})
		if err := productRepo.Create(ctx, product); err != nil {
			t.Fatalf("Error creating product: %v", err)
		}

		stock := testutil.CreateTestStock(product.ID, "STORE-ADJ-002", func(s *domain.Stock) {
			s.Quantity = 100
		})
		if err := stockRepo.Create(ctx, stock); err != nil {
			t.Fatalf("Error creating stock: %v", err)
		}

		adjusted, err := stockService.AdjustStock(ctx, product.ID, "STORE-ADJ-002", -30)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if adjusted.Quantity != 70 {
			t.Errorf("Expected quantity 70, got %d", adjusted.Quantity)
		}
	})

	t.Run("AdjustStock_NegativeResult", func(t *testing.T) {
		product := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "ADJUST-003"
		})
		if err := productRepo.Create(ctx, product); err != nil {
			t.Fatalf("Error creating product: %v", err)
		}

		stock := testutil.CreateTestStock(product.ID, "STORE-ADJ-003", func(s *domain.Stock) {
			s.Quantity = 50
		})
		if err := stockRepo.Create(ctx, stock); err != nil {
			t.Fatalf("Error creating stock: %v", err)
		}

		_, err := stockService.AdjustStock(ctx, product.ID, "STORE-ADJ-003", -100)
		if err == nil {
			t.Error("Expected error when adjustment would result in negative stock, got nil")
		}
	})
}

func TestStockService_GetAvailability(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	stockRepo := repository.NewStockRepository(db)
	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	publisher := mocks.NewNoOpPublisher()
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo, publisher)

	ctx := context.Background()

	t.Run("GetAvailability_Success", func(t *testing.T) {
		product := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "AVAIL-001"
		})
		if err := productRepo.Create(ctx, product); err != nil {
			t.Fatalf("Error creating product: %v", err)
		}

		stock := testutil.CreateTestStock(product.ID, "STORE-AVAIL-001", func(s *domain.Stock) {
			s.Quantity = 100
			s.Reserved = 20
		})
		if err := stockRepo.Create(ctx, stock); err != nil {
			t.Fatalf("Error creating stock: %v", err)
		}

		available, err := stockService.CheckAvailability(ctx, product.ID, "STORE-AVAIL-001", 1)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Available should be true for quantity=1 (80 available = 100 - 20)
		if !available {
			t.Error("Expected available to be true")
		}

		// Test with quantity greater than available
		available, err = stockService.CheckAvailability(ctx, product.ID, "STORE-AVAIL-001", 100)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if available {
			t.Error("Expected available to be false for quantity 100 (only 80 available)")
		}
	})
}

func TestStockService_GetAllStockByProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	stockRepo := repository.NewStockRepository(db)
	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	publisher := mocks.NewNoOpPublisher()
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo, publisher)

	ctx := context.Background()

	t.Run("GetAllStockByProduct_MultipleStores", func(t *testing.T) {
		product := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "MULTI-STORE-001"
		})
		if err := productRepo.Create(ctx, product); err != nil {
			t.Fatalf("Error creating product: %v", err)
		}

		// Create stock in multiple stores
		stockA := testutil.CreateTestStock(product.ID, "STORE-A", func(s *domain.Stock) {
			s.Quantity = 50
		})
		if err := stockRepo.Create(ctx, stockA); err != nil {
			t.Fatalf("Error creating stock A: %v", err)
		}

		stockB := testutil.CreateTestStock(product.ID, "STORE-B", func(s *domain.Stock) {
			s.Quantity = 75
		})
		if err := stockRepo.Create(ctx, stockB); err != nil {
			t.Fatalf("Error creating stock B: %v", err)
		}

		stockC := testutil.CreateTestStock(product.ID, "STORE-C", func(s *domain.Stock) {
			s.Quantity = 100
		})
		if err := stockRepo.Create(ctx, stockC); err != nil {
			t.Fatalf("Error creating stock C: %v", err)
		}

		stocks, err := stockService.GetAllStockByProduct(ctx, product.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(stocks) != 3 {
			t.Errorf("Expected 3 stocks, got %d", len(stocks))
		}

		// Verify total quantity
		totalQuantity := 0
		for _, s := range stocks {
			totalQuantity += s.Quantity
		}

		expectedTotal := 225
		if totalQuantity != expectedTotal {
			t.Errorf("Expected total quantity %d, got %d", expectedTotal, totalQuantity)
		}
	})
}

func TestStockService_GetLowStockItems(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	stockRepo := repository.NewStockRepository(db)
	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	publisher := mocks.NewNoOpPublisher()
	stockService := service.NewStockService(stockRepo, productRepo, eventRepo, publisher)

	ctx := context.Background()

	t.Run("GetLowStockItems_FilterByThreshold", func(t *testing.T) {
		product1 := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "LOW-STOCK-001"
		})
		if err := productRepo.Create(ctx, product1); err != nil {
			t.Fatalf("Error creating product1: %v", err)
		}

		product2 := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "LOW-STOCK-002"
		})
		if err := productRepo.Create(ctx, product2); err != nil {
			t.Fatalf("Error creating product2: %v", err)
		}

		product3 := testutil.CreateTestProduct(func(p *domain.Product) {
			p.SKU = "LOW-STOCK-003"
		})
		if err := productRepo.Create(ctx, product3); err != nil {
			t.Fatalf("Error creating product3: %v", err)
		}

		// Create stocks with different quantities
		stock1 := testutil.CreateTestStock(product1.ID, "STORE-LOW", func(s *domain.Stock) {
			s.Quantity = 5 // Low
		})
		if err := stockRepo.Create(ctx, stock1); err != nil {
			t.Fatalf("Error creating stock1: %v", err)
		}

		stock2 := testutil.CreateTestStock(product2.ID, "STORE-LOW", func(s *domain.Stock) {
			s.Quantity = 50 // Normal
		})
		if err := stockRepo.Create(ctx, stock2); err != nil {
			t.Fatalf("Error creating stock2: %v", err)
		}

		stock3 := testutil.CreateTestStock(product3.ID, "STORE-LOW", func(s *domain.Stock) {
			s.Quantity = 3 // Low
		})
		if err := stockRepo.Create(ctx, stock3); err != nil {
			t.Fatalf("Error creating stock3: %v", err)
		}

		threshold := 10
		lowStocks, err := stockService.GetLowStockItems(ctx, threshold)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Since we can't filter by storeID, we check if at least 2 low stock items were found
		if len(lowStocks) < 2 {
			t.Errorf("Expected at least 2 low stock items, got %d", len(lowStocks))
		}

		// Verify all returned items are below threshold
		for _, stock := range lowStocks {
			if stock.Quantity >= threshold {
				t.Errorf("Expected quantity < %d, got %d", threshold, stock.Quantity)
			}
		}
	})
}
