package unit

import (
	"context"
	"testing"
	"time"

	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
	"inventory-system/test/testutil"
)

func TestProductRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	product := &domain.Product{
		ID:          "test-product-1",
		SKU:         "TEST-001",
		Name:        "Test Product",
		Description: "Test Description",
		Category:    "Test Category",
		Price:       99.99,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, product)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Verificar que se creó
	retrieved, err := repo.GetByID(ctx, product.ID)
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}

	if retrieved.SKU != product.SKU {
		t.Errorf("Expected SKU %s, got %s", product.SKU, retrieved.SKU)
	}
	if retrieved.Name != product.Name {
		t.Errorf("Expected Name %s, got %s", product.Name, retrieved.Name)
	}
	if retrieved.Price != product.Price {
		t.Errorf("Expected Price %.2f, got %.2f", product.Price, retrieved.Price)
	}
}

func TestProductRepository_GetByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "non-existent-id")
	if err == nil {
		t.Fatal("Expected error for non-existent product, got nil")
	}

	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestProductRepository_GetBySKU(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	product := &domain.Product{
		ID:       "test-product-2",
		SKU:      "TEST-SKU-002",
		Name:     "Product by SKU",
		Category: "Test",
		Price:    49.99,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, product)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Buscar por SKU
	retrieved, err := repo.GetBySKU(ctx, product.SKU)
	if err != nil {
		t.Fatalf("Failed to get product by SKU: %v", err)
	}

	if retrieved.ID != product.ID {
		t.Errorf("Expected ID %s, got %s", product.ID, retrieved.ID)
	}
}

func TestProductRepository_List(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	// Crear múltiples productos
	products := []*domain.Product{
		{
			ID:       "prod-1",
			SKU:      "SKU-001",
			Name:     "Product 1",
			Category: "Electronics",
			Price:    100.00,

			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:       "prod-2",
			SKU:      "SKU-002",
			Name:     "Product 2",
			Category: "Electronics",
			Price:    200.00,

			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:       "prod-3",
			SKU:      "SKU-003",
			Name:     "Product 3",
			Category: "Books",
			Price:    30.00,

			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, p := range products {
		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("Failed to create product: %v", err)
		}
	}

	// Listar todos (incluyendo los 5 del schema inicial)
	list, err := repo.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("Failed to list products: %v", err)
	}

	if len(list) < 3 {
		t.Errorf("Expected at least 3 products, got %d", len(list))
	}

	// Test paginación
	page1, err := repo.List(ctx, 2, 0)
	if err != nil {
		t.Fatalf("Failed to get page 1: %v", err)
	}
	if len(page1) != 2 {
		t.Errorf("Expected 2 products in page 1, got %d", len(page1))
	}

	page2, err := repo.List(ctx, 2, 2)
	if err != nil {
		t.Fatalf("Failed to get page 2: %v", err)
	}
	if len(page2) == 0 {
		t.Error("Expected products in page 2, got 0")
	}
}

func TestProductRepository_ListByCategory(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	// Los datos de ejemplo ya tienen productos de electronics (minúscula)
	list, err := repo.ListByCategory(ctx, "electronics", 10, 0)
	if err != nil {
		t.Fatalf("Failed to list by category: %v", err)
	}

	if len(list) < 2 {
		t.Errorf("Expected at least 2 electronics products from sample data, got %d", len(list))
	}

	// Verificar que todos son de la categoría correcta
	for _, p := range list {
		if p.Category != "electronics" {
			t.Errorf("Expected category electronics, got %s", p.Category)
		}
	}
}

func TestProductRepository_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	// Crear producto
	product := &domain.Product{
		ID:       "update-test-1",
		SKU:      "UPD-001",
		Name:     "Original Name",
		Category: "Test",
		Price:    99.99,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, product)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Actualizar
	product.Name = "Updated Name"
	product.Price = 149.99

	err = repo.Update(ctx, product)
	if err != nil {
		t.Fatalf("Failed to update product: %v", err)
	}

	// Verificar actualización
	updated, err := repo.GetByID(ctx, product.ID)
	if err != nil {
		t.Fatalf("Failed to get updated product: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updated.Name)
	}
	if updated.Price != 149.99 {
		t.Errorf("Expected price 149.99, got %.2f", updated.Price)
	}
}

func TestProductRepository_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	// Crear producto
	product := &domain.Product{
		ID:       "delete-test-1",
		SKU:      "DEL-001",
		Name:     "To Delete",
		Category: "Test",
		Price:    99.99,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, product)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Eliminar
	err = repo.Delete(ctx, product.ID)
	if err != nil {
		t.Fatalf("Failed to delete product: %v", err)
	}

	// Verificar que no existe
	_, err = repo.GetByID(ctx, product.ID)
	if err == nil {
		t.Fatal("Expected error when getting deleted product, got nil")
	}

	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestProductRepository_Count(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewProductRepository(db)
	ctx := context.Background()

	// Contar (debería tener los 5 productos del schema)
	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count products: %v", err)
	}

	if count < 5 {
		t.Errorf("Expected at least 5 products from sample data, got %d", count)
	}

	// Agregar uno más
	product := &domain.Product{
		ID:       "count-test-1",
		SKU:      "CNT-001",
		Name:     "Count Test",
		Category: "Test",
		Price:    99.99,

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, product)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	// Contar nuevamente
	newCount, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count products: %v", err)
	}

	if newCount != count+1 {
		t.Errorf("Expected count %d, got %d", count+1, newCount)
	}
}
