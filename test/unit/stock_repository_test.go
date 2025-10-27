package unit

import (
	"context"
	"testing"
	"time"

	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
	"inventory-system/test/testutil"
)

func TestStockRepository_GetByProductAndStore(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewStockRepository(db)
	ctx := context.Background()

	// Los datos de ejemplo tienen stock para 550e8400-e29b-41d4-a716-446655440000 en MAD-001 con quantity=10
	stock, err := repo.GetByProductAndStore(ctx, "550e8400-e29b-41d4-a716-446655440000", "MAD-001")
	if err != nil {
		t.Fatalf("Failed to get stock: %v", err)
	}

	if stock.ProductID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("Expected product ID 550e8400-e29b-41d4-a716-446655440000, got %s", stock.ProductID)
	}
	if stock.StoreID != "MAD-001" {
		t.Errorf("Expected store ID MAD-001, got %s", stock.StoreID)
	}
	if stock.Quantity != 10 {
		t.Errorf("Expected quantity 10, got %d", stock.Quantity)
	}
}

func TestStockRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewStockRepository(db)
	ctx := context.Background()

	stock := &domain.Stock{
		ID:        "test-stock-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440000",
		StoreID:   "TEST-001",
		Quantity:  100,
		Reserved:  0,
		Version:   1,
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, stock)
	if err != nil {
		t.Fatalf("Failed to create stock: %v", err)
	}

	// Verificar creación
	retrieved, err := repo.GetByProductAndStore(ctx, stock.ProductID, stock.StoreID)
	if err != nil {
		t.Fatalf("Failed to get created stock: %v", err)
	}

	if retrieved.Quantity != 100 {
		t.Errorf("Expected quantity 100, got %d", retrieved.Quantity)
	}
}

func TestStockRepository_UpdateQuantity_OptimisticLocking(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewStockRepository(db)
	ctx := context.Background()

	// Crear stock inicial
	stock := &domain.Stock{
		ID:        "lock-test-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440000",
		StoreID:   "LOCK-001",
		Quantity:  50,
		Reserved:  0,
		Version:   1,
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, stock)
	if err != nil {
		t.Fatalf("Failed to create stock: %v", err)
	}

	// Primera actualización (debería funcionar)
	stock.Quantity = 60
	err = repo.UpdateQuantity(ctx, stock)
	if err != nil {
		t.Fatalf("First update should succeed: %v", err)
	}

	// Segunda actualización con version antigua (debería fallar - optimistic lock)
	stock.Quantity = 70
	// NO incrementamos stock.Version, simulando una versión vieja
	err = repo.UpdateQuantity(ctx, stock)
	if err == nil {
		t.Fatal("Expected optimistic lock error, got nil")
	}

	if _, ok := err.(*domain.ConflictError); !ok {
		t.Errorf("Expected ConflictError, got %T: %v", err, err)
	}

	// Verificar que la cantidad no cambió
	current, err := repo.GetByProductAndStore(ctx, stock.ProductID, stock.StoreID)
	if err != nil {
		t.Fatalf("Failed to get current stock: %v", err)
	}

	if current.Quantity != 60 {
		t.Errorf("Expected quantity 60 (first update), got %d", current.Quantity)
	}
	if current.Version != 2 {
		t.Errorf("Expected version 2, got %d", current.Version)
	}
}

func TestStockRepository_ReserveStock(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewStockRepository(db)
	ctx := context.Background()

	// Crear stock
	stock := &domain.Stock{
		ID:        "reserve-test-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440000",
		StoreID:   "RES-001",
		Quantity:  100,
		Reserved:  0,
		Version:   1,
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, stock)
	if err != nil {
		t.Fatalf("Failed to create stock: %v", err)
	}

	// Reservar 10 unidades
	err = repo.ReserveStock(ctx, stock.ProductID, stock.StoreID, 10)
	if err != nil {
		t.Fatalf("Failed to reserve stock: %v", err)
	}

	// Verificar reserva
	updated, err := repo.GetByProductAndStore(ctx, stock.ProductID, stock.StoreID)
	if err != nil {
		t.Fatalf("Failed to get updated stock: %v", err)
	}

	if updated.Quantity != 100 {
		t.Errorf("Quantity should not change, expected 100, got %d", updated.Quantity)
	}
	if updated.Reserved != 10 {
		t.Errorf("Expected reserved 10, got %d", updated.Reserved)
	}
}

func TestStockRepository_ReserveStock_InsufficientStock(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewStockRepository(db)
	ctx := context.Background()

	// Crear stock con solo 5 unidades disponibles
	stock := &domain.Stock{
		ID:        "insufficient-test-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440000",
		StoreID:   "INS-001",
		Quantity:  10,
		Reserved:  5, // Solo 5 disponibles
		Version:   1,
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, stock)
	if err != nil {
		t.Fatalf("Failed to create stock: %v", err)
	}

	// Intentar reservar 10 (debería fallar)
	err = repo.ReserveStock(ctx, stock.ProductID, stock.StoreID, 10)
	if err == nil {
		t.Fatal("Expected InsufficientStockError, got nil")
	}

	if _, ok := err.(*domain.InsufficientStockError); !ok {
		t.Errorf("Expected InsufficientStockError, got %T: %v", err, err)
	}
}

func TestStockRepository_ReleaseReservedStock(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewStockRepository(db)
	ctx := context.Background()

	// Crear stock con reservas
	stock := &domain.Stock{
		ID:        "release-test-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440000",
		StoreID:   "REL-001",
		Quantity:  100,
		Reserved:  20,
		Version:   1,
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, stock)
	if err != nil {
		t.Fatalf("Failed to create stock: %v", err)
	}

	// Liberar 10 unidades
	err = repo.ReleaseReservedStock(ctx, stock.ProductID, stock.StoreID, 10)
	if err != nil {
		t.Fatalf("Failed to release stock: %v", err)
	}

	// Verificar
	updated, err := repo.GetByProductAndStore(ctx, stock.ProductID, stock.StoreID)
	if err != nil {
		t.Fatalf("Failed to get updated stock: %v", err)
	}

	if updated.Reserved != 10 {
		t.Errorf("Expected reserved 10, got %d", updated.Reserved)
	}
}

func TestStockRepository_ConfirmReservation(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewStockRepository(db)
	ctx := context.Background()

	// Crear stock con reservas
	stock := &domain.Stock{
		ID:        "confirm-test-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440000",
		StoreID:   "CONF-001",
		Quantity:  100,
		Reserved:  15,
		Version:   1,
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, stock)
	if err != nil {
		t.Fatalf("Failed to create stock: %v", err)
	}

	// Confirmar 15 unidades (debería decrementar quantity y reserved)
	err = repo.ConfirmReservation(ctx, stock.ProductID, stock.StoreID, 15)
	if err != nil {
		t.Fatalf("Failed to confirm reservation: %v", err)
	}

	// Verificar
	updated, err := repo.GetByProductAndStore(ctx, stock.ProductID, stock.StoreID)
	if err != nil {
		t.Fatalf("Failed to get updated stock: %v", err)
	}

	if updated.Quantity != 85 {
		t.Errorf("Expected quantity 85 (100-15), got %d", updated.Quantity)
	}
	if updated.Reserved != 0 {
		t.Errorf("Expected reserved 0 (15-15), got %d", updated.Reserved)
	}
}

func TestStockRepository_GetAllByProduct(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewStockRepository(db)
	ctx := context.Background()

	// Los datos de ejemplo tienen stock en 4 tiendas para cada producto
	stocks, err := repo.GetAllByProduct(ctx, "550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("Failed to get stocks by product: %v", err)
	}

	if len(stocks) != 4 {
		t.Errorf("Expected 4 stores, got %d", len(stocks))
	}

	// Verificar que todos son del mismo producto
	for _, s := range stocks {
		if s.ProductID != "550e8400-e29b-41d4-a716-446655440000" {
			t.Errorf("Expected product ID 550e8400-e29b-41d4-a716-446655440000, got %s", s.ProductID)
		}
	}
}

func TestStockRepository_GetLowStockItems(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewStockRepository(db)
	ctx := context.Background()

	// Crear stock con cantidad baja
	stock := &domain.Stock{
		ID:        "low-stock-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440000",
		StoreID:   "LOW-001",
		Quantity:  3,
		Reserved:  0,
		Version:   1,
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, stock)
	if err != nil {
		t.Fatalf("Failed to create stock: %v", err)
	}

	// Buscar items con stock bajo (threshold = 5)
	lowItems, err := repo.GetLowStockItems(ctx, 5)
	if err != nil {
		t.Fatalf("Failed to get low stock items: %v", err)
	}

	// Debería incluir nuestro item con 3 unidades
	found := false
	for _, item := range lowItems {
		if item.ID == stock.ID {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find low stock item in results")
	}
}

func TestStockRepository_GetAllByStore(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewStockRepository(db)
	ctx := context.Background()

	// Los datos de ejemplo tienen 5 productos en cada tienda
	stocks, err := repo.GetAllByStore(ctx, "MAD-001")
	if err != nil {
		t.Fatalf("Failed to get stocks by store: %v", err)
	}

	if len(stocks) != 5 {
		t.Errorf("Expected 5 products in store, got %d", len(stocks))
	}

	// Verificar que todos son de la misma tienda
	for _, s := range stocks {
		if s.StoreID != "MAD-001" {
			t.Errorf("Expected store ID MAD-001, got %s", s.StoreID)
		}
	}
}
