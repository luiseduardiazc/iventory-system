package unit

import (
	"context"
	"testing"
	"time"

	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
	"inventory-system/test/testutil"
)

func TestReservationRepository_Create(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewReservationRepository(db)
	ctx := context.Background()

	reservation := &domain.Reservation{
		ID:        "res-test-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440000",
		StoreID:   "MAD-001",
		Quantity:  5,
		Status:    domain.ReservationStatusPending,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
		UpdatedAt: testutil.PtrTime(time.Now()),
	}

	err := repo.Create(ctx, reservation)
	if err != nil {
		t.Fatalf("Failed to create reservation: %v", err)
	}

	// Verificar creación
	retrieved, err := repo.GetByID(ctx, reservation.ID)
	if err != nil {
		t.Fatalf("Failed to get reservation: %v", err)
	}

	if retrieved.ProductID != reservation.ProductID {
		t.Errorf("Expected product ID %s, got %s", reservation.ProductID, retrieved.ProductID)
	}
	if retrieved.Status != domain.ReservationStatusPending {
		t.Errorf("Expected status PENDING, got %s", retrieved.Status)
	}
}

func TestReservationRepository_GetByID_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewReservationRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "non-existent-reservation")
	if err == nil {
		t.Fatal("Expected error for non-existent reservation, got nil")
	}

	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Errorf("Expected NotFoundError, got %T", err)
	}
}

func TestReservationRepository_UpdateStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewReservationRepository(db)
	ctx := context.Background()

	reservation := &domain.Reservation{
		ID:        "status-test-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440000",
		StoreID:   "MAD-001",
		Quantity:  3,
		Status:    domain.ReservationStatusPending,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
		UpdatedAt: testutil.PtrTime(time.Now()),
	}

	err := repo.Create(ctx, reservation)
	if err != nil {
		t.Fatalf("Failed to create reservation: %v", err)
	}

	// Actualizar a CONFIRMED
	err = repo.UpdateStatus(ctx, reservation.ID, domain.ReservationStatusConfirmed)
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Verificar
	updated, err := repo.GetByID(ctx, reservation.ID)
	if err != nil {
		t.Fatalf("Failed to get updated reservation: %v", err)
	}

	if updated.Status != domain.ReservationStatusConfirmed {
		t.Errorf("Expected status CONFIRMED, got %s", updated.Status)
	}
}

func TestReservationRepository_GetPendingExpired(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewReservationRepository(db)
	ctx := context.Background()

	// Crear reserva ya expirada
	expiredReservation := &domain.Reservation{
		ID:        "expired-test-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440000",
		StoreID:   "MAD-001",
		Quantity:  2,
		Status:    domain.ReservationStatusPending,
		ExpiresAt: time.Now().Add(-10 * time.Minute), // Expirada hace 10 minutos
		CreatedAt: time.Now().Add(-20 * time.Minute),
		UpdatedAt: testutil.PtrTime(time.Now().Add(-20 * time.Minute)),
	}

	err := repo.Create(ctx, expiredReservation)
	if err != nil {
		t.Fatalf("Failed to create expired reservation: %v", err)
	}

	// Crear reserva válida (no expirada)
	validReservation := &domain.Reservation{
		ID:        "valid-test-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440000",
		StoreID:   "MAD-001",
		Quantity:  3,
		Status:    domain.ReservationStatusPending,
		ExpiresAt: time.Now().Add(15 * time.Minute), // Válida por 15 min
		CreatedAt: time.Now(),
		UpdatedAt: testutil.PtrTime(time.Now()),
	}

	err = repo.Create(ctx, validReservation)
	if err != nil {
		t.Fatalf("Failed to create valid reservation: %v", err)
	}

	// Obtener reservas pendientes expiradas
	expired, err := repo.GetPendingExpired(ctx)
	if err != nil {
		t.Fatalf("Failed to get pending expired: %v", err)
	}

	// Debería encontrar solo la expirada
	found := false
	for _, res := range expired {
		if res.ID == expiredReservation.ID {
			found = true
		}
		if res.ID == validReservation.ID {
			t.Error("Valid reservation should not be in expired list")
		}
	}

	if !found {
		t.Error("Expected to find expired reservation")
	}
}

func TestReservationRepository_GetByProductAndStore(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewReservationRepository(db)
	ctx := context.Background()

	productID := "550e8400-e29b-41d4-a716-446655440000"
	storeID := "BCN-001"

	// Crear varias reservas
	reservations := []*domain.Reservation{
		{
			ID:        "filter-test-1",
			ProductID: productID,
			StoreID:   storeID,
			Quantity:  2,
			Status:    domain.ReservationStatusPending,
			ExpiresAt: time.Now().Add(15 * time.Minute),
			CreatedAt: time.Now(),
			UpdatedAt: testutil.PtrTime(time.Now()),
		},
		{
			ID:        "filter-test-2",
			ProductID: productID,
			StoreID:   storeID,
			Quantity:  3,
			Status:    domain.ReservationStatusConfirmed,
			ExpiresAt: time.Now().Add(15 * time.Minute),
			CreatedAt: time.Now(),
			UpdatedAt: testutil.PtrTime(time.Now()),
		},
		{
			ID:        "filter-test-3",
			ProductID: productID,
			StoreID:   "MAD-001", // Diferente tienda
			Quantity:  1,
			Status:    domain.ReservationStatusPending,
			ExpiresAt: time.Now().Add(15 * time.Minute),
			CreatedAt: time.Now(),
			UpdatedAt: testutil.PtrTime(time.Now()),
		},
	}

	for _, r := range reservations {
		if err := repo.Create(ctx, r); err != nil {
			t.Fatalf("Failed to create reservation: %v", err)
		}
	}

	// Buscar todas las reservas del producto en BCN-001
	results, err := repo.GetByProductAndStore(ctx, productID, storeID, nil)
	if err != nil {
		t.Fatalf("Failed to get by product and store: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 reservations for BCN-001, got %d", len(results))
	}

	// Filtrar solo PENDING
	status := domain.ReservationStatusPending
	pendingResults, err := repo.GetByProductAndStore(ctx, productID, storeID, &status)
	if err != nil {
		t.Fatalf("Failed to get pending reservations: %v", err)
	}

	if len(pendingResults) != 1 {
		t.Errorf("Expected 1 pending reservation, got %d", len(pendingResults))
	}
}

func TestReservationRepository_GetPendingByStore(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewReservationRepository(db)
	ctx := context.Background()

	storeID := "VAL-001"

	// Crear reservas pendientes
	pending := &domain.Reservation{
		ID:        "pending-store-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440000",
		StoreID:   storeID,
		Quantity:  5,
		Status:    domain.ReservationStatusPending,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
		UpdatedAt: testutil.PtrTime(time.Now()),
	}

	confirmed := &domain.Reservation{
		ID:        "confirmed-store-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440001",
		StoreID:   storeID,
		Quantity:  3,
		Status:    domain.ReservationStatusConfirmed,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		CreatedAt: time.Now(),
		UpdatedAt: testutil.PtrTime(time.Now()),
	}

	for _, r := range []*domain.Reservation{pending, confirmed} {
		if err := repo.Create(ctx, r); err != nil {
			t.Fatalf("Failed to create reservation: %v", err)
		}
	}

	// Obtener solo pendientes
	results, err := repo.GetPendingByStore(ctx, storeID)
	if err != nil {
		t.Fatalf("Failed to get pending by store: %v", err)
	}

	// Verificar que solo hay pending
	for _, r := range results {
		if r.Status != domain.ReservationStatusPending {
			t.Errorf("Expected only PENDING reservations, got %s", r.Status)
		}
	}
}

func TestReservationRepository_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewReservationRepository(db)
	ctx := context.Background()

	reservation := &domain.Reservation{
		ID:        "delete-test-1",
		ProductID: "550e8400-e29b-41d4-a716-446655440000",
		StoreID:   "MAD-001",
		Quantity:  1,
		Status:    domain.ReservationStatusCancelled,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
		UpdatedAt: testutil.PtrTime(time.Now()),
	}

	err := repo.Create(ctx, reservation)
	if err != nil {
		t.Fatalf("Failed to create reservation: %v", err)
	}

	// Eliminar
	err = repo.Delete(ctx, reservation.ID)
	if err != nil {
		t.Fatalf("Failed to delete reservation: %v", err)
	}

	// Verificar que no existe
	_, err = repo.GetByID(ctx, reservation.ID)
	if err == nil {
		t.Fatal("Expected error when getting deleted reservation, got nil")
	}
}

func TestReservationRepository_DeleteOldCompleted(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewReservationRepository(db)
	ctx := context.Background()

	// Crear reserva antigua confirmada
	old := &domain.Reservation{
		ID:         "old-completed-1",
		ProductID:  "550e8400-e29b-41d4-a716-446655440000",
		StoreID:    "MAD-001",
		CustomerID: "CUST-OLD-001",
		Quantity:   1,
		Status:     domain.ReservationStatusConfirmed,
		ExpiresAt:  time.Now().Add(-50 * 24 * time.Hour), // Hace 50 días
		CreatedAt:  time.Now().Add(-50 * 24 * time.Hour),
		UpdatedAt:  testutil.PtrTime(time.Now().Add(-50 * 24 * time.Hour)),
	}

	// Crear reserva reciente confirmada
	recent := &domain.Reservation{
		ID:         "recent-completed-1",
		ProductID:  "550e8400-e29b-41d4-a716-446655440000",
		StoreID:    "MAD-001",
		CustomerID: "CUST-RECENT-001",
		Quantity:   1,
		Status:     domain.ReservationStatusConfirmed,
		ExpiresAt:  time.Now().Add(-1 * 24 * time.Hour), // Hace 1 día
		CreatedAt:  time.Now().Add(-1 * 24 * time.Hour),
		UpdatedAt:  testutil.PtrTime(time.Now().Add(-1 * 24 * time.Hour)),
	}

	for _, r := range []*domain.Reservation{old, recent} {
		if err := repo.Create(ctx, r); err != nil {
			t.Fatalf("Failed to create reservation: %v", err)
		}
	}

	// Eliminar completadas con más de 30 días
	deleted, err := repo.DeleteOldCompleted(ctx, time.Now().Add(-30*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to delete old completed: %v", err)
	}

	if deleted < 1 {
		t.Errorf("Expected at least 1 deletion, got %d", deleted)
	}

	// Verificar que la antigua fue eliminada
	_, err = repo.GetByID(ctx, old.ID)
	if err == nil {
		t.Error("Old reservation should have been deleted")
	}

	// Verificar que la reciente sigue existiendo
	_, err = repo.GetByID(ctx, recent.ID)
	if err != nil {
		t.Error("Recent reservation should still exist")
	}
}

func TestReservationRepository_CountByStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, db)

	repo := repository.NewReservationRepository(db)
	ctx := context.Background()

	// Crear varias reservas con diferentes estados
	statuses := []domain.ReservationStatus{
		domain.ReservationStatusPending,
		domain.ReservationStatusPending,
		domain.ReservationStatusConfirmed,
		domain.ReservationStatusCancelled,
	}

	for i, status := range statuses {
		r := &domain.Reservation{
			ID:        string(rune('a'+i)) + "-count-test",
			ProductID: "550e8400-e29b-41d4-a716-446655440000",
			StoreID:   "MAD-001",
			Quantity:  1,
			Status:    status,
			ExpiresAt: time.Now().Add(15 * time.Minute),
			CreatedAt: time.Now(),
			UpdatedAt: testutil.PtrTime(time.Now()),
		}
		if err := repo.Create(ctx, r); err != nil {
			t.Fatalf("Failed to create reservation: %v", err)
		}
	}

	// Contar pending (debería ser 2)
	count, err := repo.CountByStatus(ctx, domain.ReservationStatusPending)
	if err != nil {
		t.Fatalf("Failed to count pending: %v", err)
	}

	if count < 2 {
		t.Errorf("Expected at least 2 pending reservations, got %d", count)
	}

	// Contar confirmed (debería ser 1)
	count, err = repo.CountByStatus(ctx, domain.ReservationStatusConfirmed)
	if err != nil {
		t.Fatalf("Failed to count confirmed: %v", err)
	}

	if count < 1 {
		t.Errorf("Expected at least 1 confirmed reservation, got %d", count)
	}
}
