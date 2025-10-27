package testutil

import (
	"testing"
	"time"

	"inventory-system/internal/domain"
)

// CreateTestProduct crea un producto de prueba con valores por defecto
func CreateTestProduct(overrides ...func(*domain.Product)) *domain.Product {
	p := &domain.Product{
		ID:          GenerateTestID(),
		SKU:         GenerateTestSKU(),
		Name:        "Test Product",
		Description: "Test description",
		Category:    "electronics",
		Price:       99.99,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	for _, override := range overrides {
		override(p)
	}

	return p
}

// CreateTestStock crea un stock de prueba con valores por defecto
func CreateTestStock(productID, storeID string, overrides ...func(*domain.Stock)) *domain.Stock {
	s := &domain.Stock{
		ID:        GenerateTestID(),
		ProductID: productID,
		StoreID:   storeID,
		Quantity:  100,
		Reserved:  0,
		Version:   1,
		UpdatedAt: time.Now(),
	}

	for _, override := range overrides {
		override(s)
	}

	return s
}

// CreateTestReservation crea una reserva de prueba con valores por defecto
func CreateTestReservation(productID, storeID string, overrides ...func(*domain.Reservation)) *domain.Reservation {
	now := time.Now()
	r := &domain.Reservation{
		ID:         GenerateTestID(),
		ProductID:  productID,
		StoreID:    storeID,
		CustomerID: "CUST-TEST-001",
		Quantity:   5,
		Status:     domain.ReservationStatusPending,
		ExpiresAt:  time.Now().Add(15 * time.Minute),
		CreatedAt:  time.Now(),
		UpdatedAt:  &now,
	}

	for _, override := range overrides {
		override(r)
	}

	return r
}

// CreateTestEvent crea un evento de prueba con valores por defecto
func CreateTestEvent(eventType, aggregateType, aggregateID, storeID string) *domain.Event {
	return &domain.Event{
		ID:            GenerateTestID(),
		EventType:     eventType,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		StoreID:       storeID,
		Payload:       `{"test": "data"}`,
		Synced:        false,
		CreatedAt:     time.Now(),
	}
}

// AssertProductEqual verifica que dos productos sean iguales
func AssertProductEqual(t *testing.T, expected, actual *domain.Product) {
	t.Helper()

	if actual.SKU != expected.SKU {
		t.Errorf("SKU mismatch: expected %s, got %s", expected.SKU, actual.SKU)
	}
	if actual.Name != expected.Name {
		t.Errorf("Name mismatch: expected %s, got %s", expected.Name, actual.Name)
	}
	if actual.Price != expected.Price {
		t.Errorf("Price mismatch: expected %.2f, got %.2f", expected.Price, actual.Price)
	}
}

// AssertStockEqual verifica que dos stocks sean iguales
func AssertStockEqual(t *testing.T, expected, actual *domain.Stock) {
	t.Helper()

	if actual.ProductID != expected.ProductID {
		t.Errorf("ProductID mismatch: expected %s, got %s", expected.ProductID, actual.ProductID)
	}
	if actual.StoreID != expected.StoreID {
		t.Errorf("StoreID mismatch: expected %s, got %s", expected.StoreID, actual.StoreID)
	}
	if actual.Quantity != expected.Quantity {
		t.Errorf("Quantity mismatch: expected %d, got %d", expected.Quantity, actual.Quantity)
	}
	if actual.Reserved != expected.Reserved {
		t.Errorf("Reserved mismatch: expected %d, got %d", expected.Reserved, actual.Reserved)
	}
}
