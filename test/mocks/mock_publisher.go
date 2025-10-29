package mocks

import (
	"context"
	"errors"
	"log"
	"sync"

	"inventory-system/internal/domain"
)

// MockPublisher es una implementación falsa de EventPublisher para tests.
// Guarda los eventos en memoria en lugar de enviarlos a un broker real.
//
// Ejemplo de uso en tests:
//
//	mock := NewMockPublisher()
//	service := NewStockService(stockRepo, productRepo, eventRepo, mock)
//
//	// Ejecutar operación
//	service.UpdateStock(ctx, "prod-1", "MAD-001", 50)
//
//	// Verificar que se publicó el evento
//	assert.Equal(t, 1, len(mock.PublishedEvents))
//	assert.Equal(t, "stock.updated", mock.PublishedEvents[0].EventType)
type MockPublisher struct {
	mu              sync.Mutex
	PublishedEvents []*domain.Event
	ShouldFail      bool
	FailError       error
	PublishCount    int
	BatchCount      int
}

// NewMockPublisher crea una nueva instancia de MockPublisher.
func NewMockPublisher() *MockPublisher {
	return &MockPublisher{
		PublishedEvents: make([]*domain.Event, 0),
		ShouldFail:      false,
		FailError:       errors.New("mock publish failed"),
	}
}

// Publish simula la publicación de un evento.
// Si ShouldFail=true, retorna error.
func (m *MockPublisher) Publish(ctx context.Context, event *domain.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PublishCount++

	if m.ShouldFail {
		return m.FailError
	}

	// Guardar evento en memoria
	m.PublishedEvents = append(m.PublishedEvents, event)

	return nil
}

// PublishBatch simula la publicación de múltiples eventos.
func (m *MockPublisher) PublishBatch(ctx context.Context, events []*domain.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.BatchCount++

	if m.ShouldFail {
		return m.FailError
	}

	// Publicar uno por uno
	for _, event := range events {
		m.PublishedEvents = append(m.PublishedEvents, event)
	}

	return nil
}

// Close no hace nada en el mock.
func (m *MockPublisher) Close() error {
	return nil
}

// GetLastEvent retorna el último evento publicado o nil si no hay eventos.
func (m *MockPublisher) GetLastEvent() *domain.Event {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.PublishedEvents) == 0 {
		return nil
	}
	return m.PublishedEvents[len(m.PublishedEvents)-1]
}

// GetEventsByType filtra eventos por tipo.
func (m *MockPublisher) GetEventsByType(eventType string) []*domain.Event {
	m.mu.Lock()
	defer m.mu.Unlock()

	filtered := make([]*domain.Event, 0)
	for _, event := range m.PublishedEvents {
		if event.EventType == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// GetEventsByStore filtra eventos por tienda.
func (m *MockPublisher) GetEventsByStore(storeID string) []*domain.Event {
	m.mu.Lock()
	defer m.mu.Unlock()

	filtered := make([]*domain.Event, 0)
	for _, event := range m.PublishedEvents {
		if event.StoreID == storeID {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// Reset limpia todos los eventos guardados.
// Útil para limpiar estado entre tests.
func (m *MockPublisher) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PublishedEvents = make([]*domain.Event, 0)
	m.PublishCount = 0
	m.BatchCount = 0
	m.ShouldFail = false
}

// Count retorna el número total de eventos publicados.
func (m *MockPublisher) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.PublishedEvents)
}

// NoOpPublisher es una implementación de EventPublisher que no hace nada.
// Útil para testing y desarrollo cuando no se necesita publicación real de eventos.
type NoOpPublisher struct{}

// NewNoOpPublisher crea una nueva instancia de NoOpPublisher
func NewNoOpPublisher() *NoOpPublisher {
	return &NoOpPublisher{}
}

// Publish no hace nada, solo registra el evento en logs
func (p *NoOpPublisher) Publish(ctx context.Context, event *domain.Event) error {
	log.Printf("[NoOp] Event would be published: type=%s, store=%s, id=%s",
		event.EventType, event.StoreID, event.ID)
	return nil
}

// PublishBatch no hace nada para múltiples eventos
func (p *NoOpPublisher) PublishBatch(ctx context.Context, events []*domain.Event) error {
	log.Printf("[NoOp] Batch of %d events would be published", len(events))
	return nil
}

// Close no hace nada
func (p *NoOpPublisher) Close() error {
	return nil
}
