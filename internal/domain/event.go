package domain

import (
	"encoding/json"
	"time"
)

// Event representa un evento en el sistema (Event Sourcing)
type Event struct {
	ID            string     `json:"id"`
	EventType     string     `json:"event_type"`
	AggregateID   string     `json:"aggregate_id"`   // ID del producto/reserva afectado
	AggregateType string     `json:"aggregate_type"` // "product", "stock", "reservation"
	StoreID       string     `json:"store_id"`       // Origen del evento
	Payload       string     `json:"payload"`        // JSON string
	CreatedAt     time.Time  `json:"created_at"`
	Synced        bool       `json:"synced"`
	SyncedAt      *time.Time `json:"synced_at,omitempty"`
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

// Helper functions para crear eventos comunes

func NewStockUpdatedEvent(productID, storeID string, oldQuantity, newQuantity int) *Event {
	payload := map[string]interface{}{
		"product_id":   productID,
		"store_id":     storeID,
		"old_quantity": oldQuantity,
		"new_quantity": newQuantity,
	}
	payloadJSON, _ := json.Marshal(payload)

	return &Event{
		ID:            generateEventID(),
		EventType:     "stock.updated",
		AggregateID:   productID,
		AggregateType: "stock",
		StoreID:       storeID,
		Payload:       string(payloadJSON),
		CreatedAt:     time.Now(),
		Synced:        false,
	}
}

func NewStockCreatedEvent(productID, storeID string, initialQuantity int) *Event {
	payload := map[string]interface{}{
		"product_id":       productID,
		"store_id":         storeID,
		"initial_quantity": initialQuantity,
	}
	payloadJSON, _ := json.Marshal(payload)

	return &Event{
		ID:            generateEventID(),
		EventType:     "stock.created",
		AggregateID:   productID,
		AggregateType: "stock",
		StoreID:       storeID,
		Payload:       string(payloadJSON),
		CreatedAt:     time.Now(),
		Synced:        false,
	}
}

func NewStockTransferredEvent(productID, fromStoreID, toStoreID string, quantity int) *Event {
	payload := map[string]interface{}{
		"product_id":    productID,
		"from_store_id": fromStoreID,
		"to_store_id":   toStoreID,
		"quantity":      quantity,
	}
	payloadJSON, _ := json.Marshal(payload)

	return &Event{
		ID:            generateEventID(),
		EventType:     "stock.transferred",
		AggregateID:   productID,
		AggregateType: "stock",
		StoreID:       fromStoreID,
		Payload:       string(payloadJSON),
		CreatedAt:     time.Now(),
		Synced:        false,
	}
}

func NewReservationCreatedEvent(reservationID, productID, storeID string, quantity int) *Event {
	payload := map[string]interface{}{
		"reservation_id": reservationID,
		"product_id":     productID,
		"store_id":       storeID,
		"quantity":       quantity,
	}
	payloadJSON, _ := json.Marshal(payload)

	return &Event{
		ID:            generateEventID(),
		EventType:     "reservation.created",
		AggregateID:   reservationID,
		AggregateType: "reservation",
		StoreID:       storeID,
		Payload:       string(payloadJSON),
		CreatedAt:     time.Now(),
		Synced:        false,
	}
}

func NewReservationConfirmedEvent(reservationID, productID, storeID string, quantity int) *Event {
	payload := map[string]interface{}{
		"reservation_id": reservationID,
		"product_id":     productID,
		"store_id":       storeID,
		"quantity":       quantity,
	}
	payloadJSON, _ := json.Marshal(payload)

	return &Event{
		ID:            generateEventID(),
		EventType:     "reservation.confirmed",
		AggregateID:   reservationID,
		AggregateType: "reservation",
		StoreID:       storeID,
		Payload:       string(payloadJSON),
		CreatedAt:     time.Now(),
		Synced:        false,
	}
}

func NewReservationCancelledEvent(reservationID, productID, storeID string, quantity int) *Event {
	payload := map[string]interface{}{
		"reservation_id": reservationID,
		"product_id":     productID,
		"store_id":       storeID,
		"quantity":       quantity,
	}
	payloadJSON, _ := json.Marshal(payload)

	return &Event{
		ID:            generateEventID(),
		EventType:     "reservation.cancelled",
		AggregateID:   reservationID,
		AggregateType: "reservation",
		StoreID:       storeID,
		Payload:       string(payloadJSON),
		CreatedAt:     time.Now(),
		Synced:        false,
	}
}

func NewReservationExpiredEvent(reservationID, productID, storeID string, quantity int) *Event {
	payload := map[string]interface{}{
		"reservation_id": reservationID,
		"product_id":     productID,
		"store_id":       storeID,
		"quantity":       quantity,
	}
	payloadJSON, _ := json.Marshal(payload)

	return &Event{
		ID:            generateEventID(),
		EventType:     "reservation.expired",
		AggregateID:   reservationID,
		AggregateType: "reservation",
		StoreID:       storeID,
		Payload:       string(payloadJSON),
		CreatedAt:     time.Now(),
		Synced:        false,
	}
}

// TODO: Implementar generador UUID
func generateEventID() string {
	return "evt-" + time.Now().Format("20060102150405") + "-temp"
}
