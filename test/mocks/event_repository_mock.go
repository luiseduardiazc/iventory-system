package mocks

import (
	"context"
	"inventory-system/internal/domain"
)

// MockEventRepository es un mock para EventRepository
type MockEventRepository struct {
	CreateFunc         func(ctx context.Context, event *domain.Event) error
	GetUnsyncedFunc    func(ctx context.Context, limit int) ([]*domain.Event, error)
	MarkSyncedFunc     func(ctx context.Context, eventID string) error
	GetByAggregateFunc func(ctx context.Context, aggregateID string) ([]*domain.Event, error)
	CountUnsyncedFunc  func(ctx context.Context) (int, error)
}

func (m *MockEventRepository) Create(ctx context.Context, event *domain.Event) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, event)
	}
	return nil
}

func (m *MockEventRepository) GetUnsynced(ctx context.Context, limit int) ([]*domain.Event, error) {
	if m.GetUnsyncedFunc != nil {
		return m.GetUnsyncedFunc(ctx, limit)
	}
	return nil, nil
}

func (m *MockEventRepository) MarkSynced(ctx context.Context, eventID string) error {
	if m.MarkSyncedFunc != nil {
		return m.MarkSyncedFunc(ctx, eventID)
	}
	return nil
}

func (m *MockEventRepository) GetByAggregate(ctx context.Context, aggregateID string) ([]*domain.Event, error) {
	if m.GetByAggregateFunc != nil {
		return m.GetByAggregateFunc(ctx, aggregateID)
	}
	return nil, nil
}

func (m *MockEventRepository) CountUnsynced(ctx context.Context) (int, error) {
	if m.CountUnsyncedFunc != nil {
		return m.CountUnsyncedFunc(ctx)
	}
	return 0, nil
}
