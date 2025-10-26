package domain

import "time"

// ReservationStatus representa el estado de una reserva
type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "PENDING"   // Creada, esperando confirmación
	ReservationStatusConfirmed ReservationStatus = "CONFIRMED" // Confirmada, stock comprometido
	ReservationStatusCancelled ReservationStatus = "CANCELLED" // Cancelada manualmente
	ReservationStatusExpired   ReservationStatus = "EXPIRED"   // Expirada automáticamente
)

// Reservation representa una reserva temporal de stock
type Reservation struct {
	ID         string            `json:"id" db:"id"`
	ProductID  string            `json:"productId" db:"product_id"`
	StoreID    string            `json:"storeId" db:"store_id"`       // Tienda donde se reserva
	CustomerID string            `json:"customerId" db:"customer_id"` // Cliente que reserva
	Quantity   int               `json:"quantity" db:"quantity"`
	Status     ReservationStatus `json:"status" db:"status"`
	ExpiresAt  time.Time         `json:"expiresAt" db:"expires_at"`
	CreatedAt  time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt  *time.Time        `json:"updatedAt,omitempty" db:"updated_at"`
}

// IsExpired verifica si la reserva ha expirado
func (r *Reservation) IsExpired() bool {
	return time.Now().After(r.ExpiresAt) && r.Status == ReservationStatusPending
}

// CanConfirm verifica si la reserva puede ser confirmada
func (r *Reservation) CanConfirm() bool {
	return r.Status == ReservationStatusPending && !r.IsExpired()
}

// CanCancel verifica si la reserva puede ser cancelada
func (r *Reservation) CanCancel() bool {
	return r.Status == ReservationStatusPending
}

// TimeRemaining retorna el tiempo restante antes de expirar
func (r *Reservation) TimeRemaining() time.Duration {
	if r.Status != ReservationStatusPending {
		return 0
	}
	remaining := time.Until(r.ExpiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Validate verifica que la reserva tenga datos válidos
func (r *Reservation) Validate() error {
	if r.ProductID == "" {
		return &ValidationError{Field: "product_id", Message: "Product ID is required"}
	}
	if r.StoreID == "" {
		return &ValidationError{Field: "store_id", Message: "Store ID is required"}
	}
	if r.CustomerID == "" {
		return &ValidationError{Field: "customer_id", Message: "Customer ID is required"}
	}
	if r.Quantity <= 0 {
		return &ValidationError{Field: "quantity", Message: "Quantity must be positive"}
	}
	if r.ExpiresAt.Before(time.Now()) {
		return &ValidationError{Field: "expires_at", Message: "Expiration time must be in the future"}
	}
	return nil
}
