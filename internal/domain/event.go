package domain

import (
	"encoding/json"
	"time"
)

// EventType representa el tipo de evento
type EventType string

const (
	// Stock events
	EventTypeStockUpdated  EventType = "stock.updated"
	EventTypeStockReserved EventType = "stock.reserved"
	EventTypeStockReleased EventType = "stock.released"

	// Reservation events
	EventTypeReservationCreated   EventType = "reservation.created"
	EventTypeReservationConfirmed EventType = "reservation.confirmed"
	EventTypeReservationCancelled EventType = "reservation.cancelled"
	EventTypeReservationExpired   EventType = "reservation.expired"
)

// Event representa un evento en el sistema (Event Sourcing)
type Event struct {
	ID          string                 `json:"id" db:"id"`
	EventType   EventType              `json:"eventType" db:"event_type"`
	StoreID     string                 `json:"storeId" db:"store_id"`         // Origen del evento
	AggregateID string                 `json:"aggregateId" db:"aggregate_id"` // ID del producto/reserva afectado
	Payload     map[string]interface{} `json:"payload" db:"payload"`          // Datos del evento
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	Processed   bool                   `json:"processed" db:"processed"` // Para tracking de procesamiento
}

// PayloadAsJSON retorna el payload como JSON string (para guardar en DB)
func (e *Event) PayloadAsJSON() (string, error) {
	bytes, err := json.Marshal(e.Payload)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// SetPayloadFromJSON establece el payload desde un JSON string
func (e *Event) SetPayloadFromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), &e.Payload)
}

// Validate verifica que el evento tenga datos válidos
func (e *Event) Validate() error {
	if e.EventType == "" {
		return &ValidationError{Field: "event_type", Message: "Event type is required"}
	}
	if e.StoreID == "" {
		return &ValidationError{Field: "store_id", Message: "Store ID is required"}
	}
	if e.AggregateID == "" {
		return &ValidationError{Field: "aggregate_id", Message: "Aggregate ID is required"}
	}
	return nil
}

// StockUpdatedPayload estructura del payload para stock.updated
type StockUpdatedPayload struct {
	ProductID   string `json:"product_id"`
	StoreID     string `json:"store_id"`
	OldQuantity int    `json:"old_quantity"`
	NewQuantity int    `json:"new_quantity"`
	Reason      string `json:"reason"` // SALE, RESTOCK, ADJUSTMENT, etc.
	UserID      string `json:"user_id,omitempty"`
}

// ReservationCreatedPayload estructura del payload para reservation.created
type ReservationCreatedPayload struct {
	ReservationID string    `json:"reservation_id"`
	ProductID     string    `json:"product_id"`
	StoreID       string    `json:"store_id"`
	CustomerID    string    `json:"customer_id"`
	Quantity      int       `json:"quantity"`
	ExpiresAt     time.Time `json:"expires_at"`
}
