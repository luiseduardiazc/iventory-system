package mocks

import (
	"context"
	"database/sql"
	"inventory-system/internal/domain"
)

// MockStockRepository es un mock para StockRepository
type MockStockRepository struct {
	CreateFunc               func(ctx context.Context, stock *domain.Stock) error
	GetByProductAndStoreFunc func(ctx context.Context, productID, storeID string) (*domain.Stock, error)
	UpdateQuantityFunc       func(ctx context.Context, productID, storeID string, newQuantity, expectedVersion int) error
	ReserveStockFunc         func(ctx context.Context, tx *sql.Tx, productID, storeID string, quantity int) error
	ReleaseReservedStockFunc func(ctx context.Context, tx *sql.Tx, productID, storeID string, quantity int) error
	ConfirmReservationFunc   func(ctx context.Context, tx *sql.Tx, productID, storeID string, quantity int) error
	GetAllByProductFunc      func(ctx context.Context, productID string) ([]*domain.Stock, error)
	GetAllByStoreFunc        func(ctx context.Context, storeID string) ([]*domain.Stock, error)
	GetLowStockItemsFunc     func(ctx context.Context, storeID string, threshold int) ([]*domain.Stock, error)
}

func (m *MockStockRepository) Create(ctx context.Context, stock *domain.Stock) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, stock)
	}
	return nil
}

func (m *MockStockRepository) GetByProductAndStore(ctx context.Context, productID, storeID string) (*domain.Stock, error) {
	if m.GetByProductAndStoreFunc != nil {
		return m.GetByProductAndStoreFunc(ctx, productID, storeID)
	}
	return nil, nil
}

func (m *MockStockRepository) UpdateQuantity(ctx context.Context, productID, storeID string, newQuantity, expectedVersion int) error {
	if m.UpdateQuantityFunc != nil {
		return m.UpdateQuantityFunc(ctx, productID, storeID, newQuantity, expectedVersion)
	}
	return nil
}

func (m *MockStockRepository) ReserveStock(ctx context.Context, tx *sql.Tx, productID, storeID string, quantity int) error {
	if m.ReserveStockFunc != nil {
		return m.ReserveStockFunc(ctx, tx, productID, storeID, quantity)
	}
	return nil
}

func (m *MockStockRepository) ReleaseReservedStock(ctx context.Context, tx *sql.Tx, productID, storeID string, quantity int) error {
	if m.ReleaseReservedStockFunc != nil {
		return m.ReleaseReservedStockFunc(ctx, tx, productID, storeID, quantity)
	}
	return nil
}

func (m *MockStockRepository) ConfirmReservation(ctx context.Context, tx *sql.Tx, productID, storeID string, quantity int) error {
	if m.ConfirmReservationFunc != nil {
		return m.ConfirmReservationFunc(ctx, tx, productID, storeID, quantity)
	}
	return nil
}

func (m *MockStockRepository) GetAllByProduct(ctx context.Context, productID string) ([]*domain.Stock, error) {
	if m.GetAllByProductFunc != nil {
		return m.GetAllByProductFunc(ctx, productID)
	}
	return nil, nil
}

func (m *MockStockRepository) GetAllByStore(ctx context.Context, storeID string) ([]*domain.Stock, error) {
	if m.GetAllByStoreFunc != nil {
		return m.GetAllByStoreFunc(ctx, storeID)
	}
	return nil, nil
}

func (m *MockStockRepository) GetLowStockItems(ctx context.Context, storeID string, threshold int) ([]*domain.Stock, error) {
	if m.GetLowStockItemsFunc != nil {
		return m.GetLowStockItemsFunc(ctx, storeID, threshold)
	}
	return nil, nil
}
