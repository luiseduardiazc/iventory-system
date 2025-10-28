package unit

import (
	"context"
	"testing"

	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
	"inventory-system/internal/service"
	"inventory-system/test/testutil"
)

func TestProductService_CreateProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	productService := service.NewProductService(productRepo, eventRepo)

	ctx := context.Background()

	t.Run("CreateProduct_Success", func(t *testing.T) {
		product := &domain.Product{
			SKU:         "TEST-001",
			Name:        "Test Product",
			Description: "Test Description",
			Category:    "electronics",
			Price:       99.99,
		}

		created, err := productService.CreateProduct(ctx, product)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if created.ID == "" {
			t.Error("Expected product ID to be generated")
		}

		if created.SKU != product.SKU {
			t.Errorf("Expected SKU %s, got %s", product.SKU, created.SKU)
		}
	})

	t.Run("CreateProduct_DuplicateSKU", func(t *testing.T) {
		product1 := &domain.Product{
			SKU:      "DUP-001",
			Name:     "First Product",
			Category: "electronics",
			Price:    50.00,
		}

		_, err := productService.CreateProduct(ctx, product1)
		if err != nil {
			t.Fatalf("First product creation failed: %v", err)
		}

		// Try to create duplicate
		product2 := &domain.Product{
			SKU:      "DUP-001", // Same SKU
			Name:     "Second Product",
			Category: "electronics",
			Price:    60.00,
		}

		_, err = productService.CreateProduct(ctx, product2)
		if err == nil {
			t.Error("Expected error for duplicate SKU, got nil")
		}

		// Verify it's a ConflictError
		if _, ok := err.(*domain.ConflictError); !ok {
			t.Errorf("Expected ConflictError, got %T", err)
		}
	})

	t.Run("CreateProduct_InvalidData", func(t *testing.T) {
		product := &domain.Product{
			// Missing required fields
			SKU:      "",
			Name:     "",
			Price:    -10.00, // Invalid price
			Category: "electronics",
		}

		_, err := productService.CreateProduct(ctx, product)
		if err == nil {
			t.Error("Expected validation error, got nil")
		}
	})
}

func TestProductService_GetProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	productService := service.NewProductService(productRepo, eventRepo)

	ctx := context.Background()

	t.Run("GetProduct_Success", func(t *testing.T) {
		// Create a product first
		product := &domain.Product{
			SKU:      "GET-001",
			Name:     "Get Test Product",
			Category: "electronics",
			Price:    75.00,
		}

		created, err := productService.CreateProduct(ctx, product)
		if err != nil {
			t.Fatalf("Failed to create product: %v", err)
		}

		// Get the product
		found, err := productService.GetProduct(ctx, created.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if found.ID != created.ID {
			t.Errorf("Expected ID %s, got %s", created.ID, found.ID)
		}

		if found.SKU != product.SKU {
			t.Errorf("Expected SKU %s, got %s", product.SKU, found.SKU)
		}
	})

	t.Run("GetProduct_NotFound", func(t *testing.T) {
		_, err := productService.GetProduct(ctx, "non-existent-id")
		if err == nil {
			t.Error("Expected error for non-existent product, got nil")
		}
	})
}

func TestProductService_GetProductBySKU(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	productService := service.NewProductService(productRepo, eventRepo)

	ctx := context.Background()

	t.Run("GetProductBySKU_Success", func(t *testing.T) {
		product := &domain.Product{
			SKU:      "SKU-TEST-001",
			Name:     "SKU Test Product",
			Category: "electronics",
			Price:    125.00,
		}

		_, err := productService.CreateProduct(ctx, product)
		if err != nil {
			t.Fatalf("Failed to create product: %v", err)
		}

		// Get by SKU
		found, err := productService.GetProductBySKU(ctx, "SKU-TEST-001")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if found.SKU != "SKU-TEST-001" {
			t.Errorf("Expected SKU SKU-TEST-001, got %s", found.SKU)
		}
	})

	t.Run("GetProductBySKU_NotFound", func(t *testing.T) {
		_, err := productService.GetProductBySKU(ctx, "NON-EXISTENT-SKU")
		if err == nil {
			t.Error("Expected error for non-existent SKU, got nil")
		}
	})
}

func TestProductService_ListProducts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	productService := service.NewProductService(productRepo, eventRepo)

	ctx := context.Background()

	t.Run("ListProducts_Success", func(t *testing.T) {
		// Create multiple products
		for i := 1; i <= 5; i++ {
			product := &domain.Product{
				SKU:      testutil.GenerateSKU(),
				Name:     "List Test Product",
				Category: "electronics",
				Price:    float64(i * 10),
			}
			_, err := productService.CreateProduct(ctx, product)
			if err != nil {
				t.Fatalf("Failed to create product %d: %v", i, err)
			}
		}

		// List products
		products, err := productService.ListProducts(ctx, 10, 0)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(products) < 5 {
			t.Errorf("Expected at least 5 products, got %d", len(products))
		}
	})

	t.Run("ListProducts_Pagination", func(t *testing.T) {
		products, err := productService.ListProducts(ctx, 2, 0)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(products) > 2 {
			t.Errorf("Expected max 2 products, got %d", len(products))
		}
	})
}

func TestProductService_UpdateProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	productService := service.NewProductService(productRepo, eventRepo)

	ctx := context.Background()

	t.Run("UpdateProduct_Success", func(t *testing.T) {
		// Create product
		product := &domain.Product{
			SKU:      "UPDATE-001",
			Name:     "Original Name",
			Category: "electronics",
			Price:    100.00,
		}

		created, err := productService.CreateProduct(ctx, product)
		if err != nil {
			t.Fatalf("Failed to create product: %v", err)
		}

		// Update product
		created.Name = "Updated Name"
		created.Price = 150.00

		updated, err := productService.UpdateProduct(ctx, created)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if updated.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got %s", updated.Name)
		}

		if updated.Price != 150.00 {
			t.Errorf("Expected price 150.00, got %.2f", updated.Price)
		}
	})

	t.Run("UpdateProduct_NotFound", func(t *testing.T) {
		product := &domain.Product{
			ID:       "non-existent-id",
			SKU:      "UPDATE-002",
			Name:     "Test",
			Category: "electronics",
			Price:    50.00,
		}

		_, err := productService.UpdateProduct(ctx, product)
		if err == nil {
			t.Error("Expected error for non-existent product, got nil")
		}
	})
}

func TestProductService_DeleteProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer db.Close()

	productRepo := repository.NewProductRepository(db)
	eventRepo := repository.NewEventRepository(db)
	productService := service.NewProductService(productRepo, eventRepo)

	ctx := context.Background()

	t.Run("DeleteProduct_Success", func(t *testing.T) {
		// Create product
		product := &domain.Product{
			SKU:      "DELETE-001",
			Name:     "To Be Deleted",
			Category: "electronics",
			Price:    75.00,
		}

		created, err := productService.CreateProduct(ctx, product)
		if err != nil {
			t.Fatalf("Failed to create product: %v", err)
		}

		// Delete product
		err = productService.DeleteProduct(ctx, created.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify deletion
		_, err = productService.GetProduct(ctx, created.ID)
		if err == nil {
			t.Error("Expected error when getting deleted product, got nil")
		}
	})

	t.Run("DeleteProduct_NotFound", func(t *testing.T) {
		err := productService.DeleteProduct(ctx, "non-existent-id")
		if err == nil {
			t.Error("Expected error for non-existent product, got nil")
		}
	})
}
