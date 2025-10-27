package mocks

import (
	"context"
	"database/sql"
	"inventory-system/internal/domain"
	"time"
)

// MockReservationRepository es un mock para ReservationRepository
type MockReservationRepository struct {
	CreateFunc               func(ctx context.Context, tx *sql.Tx, reservation *domain.Reservation) error
	GetByIDFunc              func(ctx context.Context, id string) (*domain.Reservation, error)
	UpdateStatusFunc         func(ctx context.Context, tx *sql.Tx, id string, status domain.ReservationStatus) error
	GetPendingExpiredFunc    func(ctx context.Context) ([]*domain.Reservation, error)
	GetByProductAndStoreFunc func(ctx context.Context, productID, storeID string, status domain.ReservationStatus) ([]*domain.Reservation, error)
	GetPendingByStoreFunc    func(ctx context.Context, storeID string) ([]*domain.Reservation, error)
	DeleteFunc               func(ctx context.Context, id string) error
	DeleteOldCompletedFunc   func(ctx context.Context, olderThan time.Time) (int, error)
	CountByStatusFunc        func(ctx context.Context, status domain.ReservationStatus) (int, error)
}

func (m *MockReservationRepository) Create(ctx context.Context, tx *sql.Tx, reservation *domain.Reservation) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, tx, reservation)
	}
	return nil
}

func (m *MockReservationRepository) GetByID(ctx context.Context, id string) (*domain.Reservation, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockReservationRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, id string, status domain.ReservationStatus) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, tx, id, status)
	}
	return nil
}

func (m *MockReservationRepository) GetPendingExpired(ctx context.Context) ([]*domain.Reservation, error) {
	if m.GetPendingExpiredFunc != nil {
		return m.GetPendingExpiredFunc(ctx)
	}
	return nil, nil
}

func (m *MockReservationRepository) GetByProductAndStore(ctx context.Context, productID, storeID string, status domain.ReservationStatus) ([]*domain.Reservation, error) {
	if m.GetByProductAndStoreFunc != nil {
		return m.GetByProductAndStoreFunc(ctx, productID, storeID, status)
	}
	return nil, nil
}

func (m *MockReservationRepository) GetPendingByStore(ctx context.Context, storeID string) ([]*domain.Reservation, error) {
	if m.GetPendingByStoreFunc != nil {
		return m.GetPendingByStoreFunc(ctx, storeID)
	}
	return nil, nil
}

func (m *MockReservationRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockReservationRepository) DeleteOldCompleted(ctx context.Context, olderThan time.Time) (int, error) {
	if m.DeleteOldCompletedFunc != nil {
		return m.DeleteOldCompletedFunc(ctx, olderThan)
	}
	return 0, nil
}

func (m *MockReservationRepository) CountByStatus(ctx context.Context, status domain.ReservationStatus) (int, error) {
	if m.CountByStatusFunc != nil {
		return m.CountByStatusFunc(ctx, status)
	}
	return 0, nil
}
