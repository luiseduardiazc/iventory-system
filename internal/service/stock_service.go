package service

import (
	"context"
	"fmt"
	"time"
	
	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
)

// StockService maneja la lógica de negocio para stock
type StockService struct {
	stockRepo   *repository.StockRepository
	productRepo *repository.ProductRepository
	eventRepo   *repository.EventRepository
}

// NewStockService crea una nueva instancia del servicio
func NewStockService(
	stockRepo *repository.StockRepository,
	productRepo *repository.ProductRepository,
	eventRepo *repository.EventRepository,
) *StockService {
	return &StockService{
		stockRepo:   stockRepo,
		productRepo: productRepo,
		eventRepo:   eventRepo,
	}
}

// GetStockByProductAndStore obtiene el stock de un producto en una tienda
func (s *StockService) GetStockByProductAndStore(ctx context.Context, productID, storeID string) (*domain.Stock, error) {
	// Validar que el producto existe
	_, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	
	return s.stockRepo.GetByProductAndStore(ctx, productID, storeID)
}

// GetAllStockByProduct obtiene el stock de un producto en TODAS las tiendas
func (s *StockService) GetAllStockByProduct(ctx context.Context, productID string) ([]*domain.Stock, error) {
	// Validar que el producto existe
	_, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	
	return s.stockRepo.GetAllByProduct(ctx, productID)
}

// GetAllStockByStore obtiene todo el stock de una tienda
func (s *StockService) GetAllStockByStore(ctx context.Context, storeID string) ([]*domain.Stock, error) {
	return s.stockRepo.GetAllByStore(ctx, storeID)
}

// UpdateStock actualiza la cantidad de stock (con optimistic locking)
func (s *StockService) UpdateStock(ctx context.Context, productID, storeID string, newQuantity int) (*domain.Stock, error) {
	if newQuantity < 0 {
		return nil, &domain.ValidationError{
			Field:   "quantity",
			Message: "quantity cannot be negative",
		}
	}
	
	// Obtener stock actual
	stock, err := s.stockRepo.GetByProductAndStore(ctx, productID, storeID)
	if err != nil {
		return nil, err
	}
	
	// Validar que la nueva cantidad no sea menor que la cantidad reservada
	if newQuantity < stock.Reserved {
		return nil, &domain.ValidationError{
			Field:   "quantity",
			Message: fmt.Sprintf("new quantity (%d) cannot be less than reserved (%d)", newQuantity, stock.Reserved),
		}
	}
	
	// Actualizar cantidad
	oldQuantity := stock.Quantity
	stock.Quantity = newQuantity
	
	// Usar optimistic locking
	err = s.stockRepo.UpdateQuantity(ctx, stock)
	if err != nil {
		return nil, err
	}
	
	// Publicar evento de actualización de stock
	event := domain.NewStockUpdatedEvent(productID, storeID, oldQuantity, newQuantity)
	if err := s.eventRepo.Save(ctx, event); err != nil {
		// Log error pero no fallar la operación
		// TODO: implementar logging
		fmt.Printf("Warning: failed to save stock update event: %v\n", err)
	}
	
	// Retornar stock actualizado
	return s.stockRepo.GetByProductAndStore(ctx, productID, storeID)
}

// AdjustStock ajusta el stock (incrementa o decrementa)
func (s *StockService) AdjustStock(ctx context.Context, productID, storeID string, adjustment int) (*domain.Stock, error) {
	// Obtener stock actual
	stock, err := s.stockRepo.GetByProductAndStore(ctx, productID, storeID)
	if err != nil {
		return nil, err
	}
	
	newQuantity := stock.Quantity + adjustment
	
	// Validar que no quede negativo
	if newQuantity < 0 {
		return nil, &domain.ValidationError{
			Field:   "adjustment",
			Message: fmt.Sprintf("adjustment would result in negative stock (current: %d, adjustment: %d)", stock.Quantity, adjustment),
		}
	}
	
	// Validar que no sea menor que lo reservado
	if newQuantity < stock.Reserved {
		return nil, &domain.ValidationError{
			Field:   "adjustment",
			Message: fmt.Sprintf("new quantity (%d) cannot be less than reserved (%d)", newQuantity, stock.Reserved),
		}
	}
	
	return s.UpdateStock(ctx, productID, storeID, newQuantity)
}

// GetAvailableStock retorna la cantidad disponible (quantity - reserved)
func (s *StockService) GetAvailableStock(ctx context.Context, productID, storeID string) (int, error) {
	stock, err := s.stockRepo.GetByProductAndStore(ctx, productID, storeID)
	if err != nil {
		return 0, err
	}
	
	return stock.Available(), nil
}

// CheckAvailability verifica si hay stock suficiente disponible
func (s *StockService) CheckAvailability(ctx context.Context, productID, storeID string, quantity int) (bool, error) {
	available, err := s.GetAvailableStock(ctx, productID, storeID)
	if err != nil {
		return false, err
	}
	
	return available >= quantity, nil
}

// GetLowStockItems obtiene productos con stock bajo
func (s *StockService) GetLowStockItems(ctx context.Context, threshold int) ([]*domain.Stock, error) {
	if threshold < 0 {
		return nil, &domain.ValidationError{
			Field:   "threshold",
			Message: "threshold cannot be negative",
		}
	}
	
	return s.stockRepo.GetLowStockItems(ctx, threshold)
}

// InitializeStock crea stock inicial para un producto en una tienda
func (s *StockService) InitializeStock(ctx context.Context, productID, storeID string, initialQuantity int) (*domain.Stock, error) {
	if initialQuantity < 0 {
		return nil, &domain.ValidationError{
			Field:   "initialQuantity",
			Message: "initial quantity cannot be negative",
		}
	}
	
	// Validar que el producto existe
	_, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	
	// Crear stock
	stock := &domain.Stock{
		ID:        generateID(), // TODO: implementar generador de IDs
		ProductID: productID,
		StoreID:   storeID,
		Quantity:  initialQuantity,
		Reserved:  0,
		Version:   1,
	}
	
	err = s.stockRepo.Create(ctx, stock)
	if err != nil {
		return nil, err
	}
	
	// Publicar evento de creación de stock
	event := domain.NewStockCreatedEvent(productID, storeID, initialQuantity)
	if err := s.eventRepo.Save(ctx, event); err != nil {
		fmt.Printf("Warning: failed to save stock created event: %v\n", err)
	}
	
	return stock, nil
}

// TransferStock transfiere stock entre tiendas
func (s *StockService) TransferStock(ctx context.Context, productID, fromStoreID, toStoreID string, quantity int) error {
	if quantity <= 0 {
		return &domain.ValidationError{
			Field:   "quantity",
			Message: "transfer quantity must be positive",
		}
	}
	
	if fromStoreID == toStoreID {
		return &domain.ValidationError{
			Field:   "storeID",
			Message: "cannot transfer to the same store",
		}
	}
	
	// Verificar disponibilidad en tienda origen
	available, err := s.GetAvailableStock(ctx, productID, fromStoreID)
	if err != nil {
		return err
	}
	
	if available < quantity {
		return &domain.InsufficientStockError{
			ProductID: productID,
			StoreID:   fromStoreID,
			Available: available,
			Requested: quantity,
		}
	}
	
	// Decrementar en tienda origen
	_, err = s.AdjustStock(ctx, productID, fromStoreID, -quantity)
	if err != nil {
		return fmt.Errorf("failed to decrement stock from source store: %w", err)
	}
	
	// Incrementar en tienda destino
	_, err = s.AdjustStock(ctx, productID, toStoreID, quantity)
	if err != nil {
		// Intentar revertir (best effort)
		_, _ = s.AdjustStock(ctx, productID, fromStoreID, quantity)
		return fmt.Errorf("failed to increment stock in destination store: %w", err)
	}
	
	// Publicar evento de transferencia
	event := domain.NewStockTransferredEvent(productID, fromStoreID, toStoreID, quantity)
	if err := s.eventRepo.Save(ctx, event); err != nil {
		fmt.Printf("Warning: failed to save stock transfer event: %v\n", err)
	}
	
	return nil
}

// TODO: Implementar generador de IDs (UUID)
func generateID() string {
	return "temp-id-" + fmt.Sprint(time.Now().UnixNano())
}
