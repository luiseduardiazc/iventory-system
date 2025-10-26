package service

import (
	"context"
	"fmt"
	"time"

	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
)

// ReservationService maneja la lógica de negocio para reservas
type ReservationService struct {
	reservationRepo *repository.ReservationRepository
	stockRepo       *repository.StockRepository
	productRepo     *repository.ProductRepository
	eventRepo       *repository.EventRepository
}

// NewReservationService crea una nueva instancia del servicio
func NewReservationService(
	reservationRepo *repository.ReservationRepository,
	stockRepo *repository.StockRepository,
	productRepo *repository.ProductRepository,
	eventRepo *repository.EventRepository,
) *ReservationService {
	return &ReservationService{
		reservationRepo: reservationRepo,
		stockRepo:       stockRepo,
		productRepo:     productRepo,
		eventRepo:       eventRepo,
	}
}

// CreateReservation crea una nueva reserva de stock
func (s *ReservationService) CreateReservation(ctx context.Context, productID, storeID string, quantity int, ttlMinutes int) (*domain.Reservation, error) {
	// Validaciones
	if quantity <= 0 {
		return nil, &domain.ValidationError{
			Field:   "quantity",
			Message: "quantity must be positive",
		}
	}

	if ttlMinutes <= 0 {
		return nil, &domain.ValidationError{
			Field:   "ttlMinutes",
			Message: "TTL must be positive",
		}
	}

	// Validar que el producto existe
	_, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Reservar stock (usa transacción interna con lock)
	err = s.stockRepo.ReserveStock(ctx, productID, storeID, quantity)
	if err != nil {
		return nil, err
	}

	// Crear reserva
	expiresAt := time.Now().Add(time.Duration(ttlMinutes) * time.Minute)
	reservation := &domain.Reservation{
		ID:        generateID(),
		ProductID: productID,
		StoreID:   storeID,
		Quantity:  quantity,
		Status:    domain.ReservationStatusPending,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	err = s.reservationRepo.Create(ctx, reservation)
	if err != nil {
		// Revertir reserva de stock
		_ = s.stockRepo.ReleaseReservedStock(ctx, productID, storeID, quantity)
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	// Publicar evento
	event := domain.NewReservationCreatedEvent(reservation.ID, productID, storeID, quantity)
	if err := s.eventRepo.Save(ctx, event); err != nil {
		fmt.Printf("Warning: failed to save reservation created event: %v\n", err)
	}

	return reservation, nil
}

// GetReservation obtiene una reserva por ID
func (s *ReservationService) GetReservation(ctx context.Context, id string) (*domain.Reservation, error) {
	return s.reservationRepo.GetByID(ctx, id)
}

// ConfirmReservation confirma una reserva (decrementa stock real)
func (s *ReservationService) ConfirmReservation(ctx context.Context, reservationID string) error {
	// Obtener reserva
	reservation, err := s.reservationRepo.GetByID(ctx, reservationID)
	if err != nil {
		return err
	}

	// Validar estado
	if reservation.Status != domain.ReservationStatusPending {
		return &domain.ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("cannot confirm reservation with status: %s", reservation.Status),
		}
	}

	// Validar expiración
	if reservation.IsExpired() {
		return &domain.ValidationError{
			Field:   "expiresAt",
			Message: "reservation has expired",
		}
	}

	// Confirmar en stock (decrementa quantity y reserved)
	err = s.stockRepo.ConfirmReservation(ctx, reservation.ProductID, reservation.StoreID, reservation.Quantity)
	if err != nil {
		return fmt.Errorf("failed to confirm in stock: %w", err)
	}

	// Actualizar estado de reserva
	err = s.reservationRepo.UpdateStatus(ctx, reservationID, domain.ReservationStatusConfirmed)
	if err != nil {
		// Intentar revertir
		_ = s.stockRepo.ReleaseReservedStock(ctx, reservation.ProductID, reservation.StoreID, reservation.Quantity)
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	// Publicar evento
	event := domain.NewReservationConfirmedEvent(reservationID, reservation.ProductID, reservation.StoreID, reservation.Quantity)
	if err := s.eventRepo.Save(ctx, event); err != nil {
		fmt.Printf("Warning: failed to save reservation confirmed event: %v\n", err)
	}

	return nil
}

// CancelReservation cancela una reserva (libera stock reservado)
func (s *ReservationService) CancelReservation(ctx context.Context, reservationID string) error {
	// Obtener reserva
	reservation, err := s.reservationRepo.GetByID(ctx, reservationID)
	if err != nil {
		return err
	}

	// Validar estado (solo se puede cancelar si está pending)
	if reservation.Status != domain.ReservationStatusPending {
		return &domain.ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("cannot cancel reservation with status: %s", reservation.Status),
		}
	}

	// Liberar stock reservado
	err = s.stockRepo.ReleaseReservedStock(ctx, reservation.ProductID, reservation.StoreID, reservation.Quantity)
	if err != nil {
		return fmt.Errorf("failed to release reserved stock: %w", err)
	}

	// Actualizar estado
	err = s.reservationRepo.UpdateStatus(ctx, reservationID, domain.ReservationStatusCancelled)
	if err != nil {
		// Intentar revertir
		_ = s.stockRepo.ReserveStock(ctx, reservation.ProductID, reservation.StoreID, reservation.Quantity)
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	// Publicar evento
	event := domain.NewReservationCancelledEvent(reservationID, reservation.ProductID, reservation.StoreID, reservation.Quantity)
	if err := s.eventRepo.Save(ctx, event); err != nil {
		fmt.Printf("Warning: failed to save reservation cancelled event: %v\n", err)
	}

	return nil
}

// ExpireReservation expira una reserva (similar a cancelar pero por TTL)
func (s *ReservationService) ExpireReservation(ctx context.Context, reservationID string) error {
	// Obtener reserva
	reservation, err := s.reservationRepo.GetByID(ctx, reservationID)
	if err != nil {
		return err
	}

	// Validar que realmente esté expirada
	if !reservation.IsExpired() {
		return &domain.ValidationError{
			Field:   "expiresAt",
			Message: "reservation has not expired yet",
		}
	}

	// Validar estado
	if reservation.Status != domain.ReservationStatusPending {
		return nil // Ya fue procesada
	}

	// Liberar stock
	err = s.stockRepo.ReleaseReservedStock(ctx, reservation.ProductID, reservation.StoreID, reservation.Quantity)
	if err != nil {
		return fmt.Errorf("failed to release reserved stock: %w", err)
	}

	// Marcar como expirada
	err = s.reservationRepo.UpdateStatus(ctx, reservationID, domain.ReservationStatusExpired)
	if err != nil {
		_ = s.stockRepo.ReserveStock(ctx, reservation.ProductID, reservation.StoreID, reservation.Quantity)
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	// Publicar evento
	event := domain.NewReservationExpiredEvent(reservationID, reservation.ProductID, reservation.StoreID, reservation.Quantity)
	if err := s.eventRepo.Save(ctx, event); err != nil {
		fmt.Printf("Warning: failed to save reservation expired event: %v\n", err)
	}

	return nil
}

// ProcessExpiredReservations procesa todas las reservas expiradas (llamado por worker)
func (s *ReservationService) ProcessExpiredReservations(ctx context.Context) (int, error) {
	// Obtener reservas expiradas
	expired, err := s.reservationRepo.GetPendingExpired(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired reservations: %w", err)
	}

	processedCount := 0
	for _, reservation := range expired {
		err := s.ExpireReservation(ctx, reservation.ID)
		if err != nil {
			// Log error pero continuar con las demás
			fmt.Printf("Error expiring reservation %s: %v\n", reservation.ID, err)
			continue
		}
		processedCount++
	}

	return processedCount, nil
}

// GetPendingByStore obtiene reservas pendientes de una tienda
func (s *ReservationService) GetPendingByStore(ctx context.Context, storeID string) ([]*domain.Reservation, error) {
	return s.reservationRepo.GetPendingByStore(ctx, storeID)
}

// GetReservationsByProduct obtiene reservas de un producto en una tienda
func (s *ReservationService) GetReservationsByProduct(ctx context.Context, productID, storeID string, status *domain.ReservationStatus) ([]*domain.Reservation, error) {
	return s.reservationRepo.GetByProductAndStore(ctx, productID, storeID, status)
}

// CleanupOldReservations elimina reservas completadas/canceladas antiguas
func (s *ReservationService) CleanupOldReservations(ctx context.Context, daysOld int) (int64, error) {
	if daysOld <= 0 {
		return 0, &domain.ValidationError{
			Field:   "daysOld",
			Message: "daysOld must be positive",
		}
	}

	olderThan := time.Now().AddDate(0, 0, -daysOld)
	return s.reservationRepo.DeleteOldCompleted(ctx, olderThan)
}

// GetReservationStats obtiene estadísticas de reservas
func (s *ReservationService) GetReservationStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	// Contar por estado
	statuses := []domain.ReservationStatus{
		domain.ReservationStatusPending,
		domain.ReservationStatusConfirmed,
		domain.ReservationStatusCancelled,
		domain.ReservationStatusExpired,
	}

	for _, status := range statuses {
		count, err := s.reservationRepo.CountByStatus(ctx, status)
		if err != nil {
			return nil, fmt.Errorf("failed to count reservations: %w", err)
		}
		stats[string(status)] = count
	}

	return stats, nil
}
