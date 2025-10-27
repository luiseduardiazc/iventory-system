package mocks

import (
	"context"
	"inventory-system/internal/domain"
)

// MockProductRepository es un mock para ProductRepository
type MockProductRepository struct {
	CreateFunc         func(ctx context.Context, product *domain.Product) error
	GetByIDFunc        func(ctx context.Context, id string) (*domain.Product, error)
	GetBySKUFunc       func(ctx context.Context, sku string) (*domain.Product, error)
	ListFunc           func(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*domain.Product, error)
	UpdateFunc         func(ctx context.Context, product *domain.Product) error
	DeleteFunc         func(ctx context.Context, id string) error
	ListByCategoryFunc func(ctx context.Context, category string, limit, offset int) ([]*domain.Product, error)
	CountFunc          func(ctx context.Context, filters map[string]interface{}) (int, error)
}

func (m *MockProductRepository) Create(ctx context.Context, product *domain.Product) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, product)
	}
	return nil
}

func (m *MockProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockProductRepository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	if m.GetBySKUFunc != nil {
		return m.GetBySKUFunc(ctx, sku)
	}
	return nil, nil
}

func (m *MockProductRepository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*domain.Product, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filters, limit, offset)
	}
	return nil, nil
}

func (m *MockProductRepository) Update(ctx context.Context, product *domain.Product) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, product)
	}
	return nil
}

func (m *MockProductRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockProductRepository) ListByCategory(ctx context.Context, category string, limit, offset int) ([]*domain.Product, error) {
	if m.ListByCategoryFunc != nil {
		return m.ListByCategoryFunc(ctx, category, limit, offset)
	}
	return nil, nil
}

func (m *MockProductRepository) Count(ctx context.Context, filters map[string]interface{}) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, filters)
	}
	return 0, nil
}
