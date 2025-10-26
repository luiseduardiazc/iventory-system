package domain

import "time"

// Stock representa el inventario de un producto en una tienda específica
type Stock struct {
	ID        string    `json:"id" db:"id"`
	ProductID string    `json:"productId" db:"product_id"`
	StoreID   string    `json:"storeId" db:"store_id"`  // Identificador de la tienda
	Quantity  int       `json:"quantity" db:"quantity"` // Cantidad total
	Reserved  int       `json:"reserved" db:"reserved"` // Cantidad reservada (pendiente)
	Version   int       `json:"version" db:"version"`   // Para optimistic locking
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Available calcula el stock disponible (no reservado)
func (s *Stock) Available() int {
	return s.Quantity - s.Reserved
}

// CanReserve verifica si hay suficiente stock disponible para reservar
func (s *Stock) CanReserve(quantity int) bool {
	return s.Available() >= quantity
}

// CanFulfill verifica si hay suficiente cantidad total
func (s *Stock) CanFulfill(quantity int) bool {
	return s.Quantity >= quantity
}

// Validate verifica que el stock tenga datos válidos
func (s *Stock) Validate() error {
	if s.ProductID == "" {
		return &ValidationError{Field: "product_id", Message: "Product ID is required"}
	}
	if s.StoreID == "" {
		return &ValidationError{Field: "store_id", Message: "Store ID is required"}
	}
	if s.Quantity < 0 {
		return &ValidationError{Field: "quantity", Message: "Quantity cannot be negative"}
	}
	if s.Reserved < 0 {
		return &ValidationError{Field: "reserved", Message: "Reserved cannot be negative"}
	}
	if s.Reserved > s.Quantity {
		return &ValidationError{Field: "reserved", Message: "Reserved cannot exceed quantity"}
	}
	return nil
}

// StockAvailability representa la disponibilidad de un producto en todas las tiendas
type StockAvailability struct {
	ProductID      string                     `json:"productId"`
	ProductName    string                     `json:"productName"`
	TotalAvailable int                        `json:"totalAvailable"`
	Stores         []StockAvailabilityByStore `json:"stores"`
}

// StockAvailabilityByStore representa la disponibilidad en una tienda específica
type StockAvailabilityByStore struct {
	StoreID   string `json:"storeId"`
	StoreName string `json:"storeName"`
	Quantity  int    `json:"quantity"`
	Reserved  int    `json:"reserved"`
	Available int    `json:"available"`
}
