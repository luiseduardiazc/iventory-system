package infrastructure

import (
	"context"
	"log"

	"inventory-system/internal/domain"
)

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

// MockPublisher es un publisher para testing que almacena eventos en memoria
type MockPublisher struct {
	PublishedEvents []*domain.Event
	ShouldFail      bool
	FailError       error
}

// NewMockPublisher crea un nuevo MockPublisher
func NewMockPublisher() *MockPublisher {
	return &MockPublisher{
		PublishedEvents: make([]*domain.Event, 0),
		ShouldFail:      false,
	}
}

// Publish almacena el evento en memoria para verificación en tests
func (p *MockPublisher) Publish(ctx context.Context, event *domain.Event) error {
	if p.ShouldFail {
		if p.FailError != nil {
			return p.FailError
		}
		return &domain.ConflictError{Message: "mock publisher error"}
	}

	p.PublishedEvents = append(p.PublishedEvents, event)
	return nil
}

// PublishBatch almacena múltiples eventos
func (p *MockPublisher) PublishBatch(ctx context.Context, events []*domain.Event) error {
	if p.ShouldFail {
		if p.FailError != nil {
			return p.FailError
		}
		return &domain.ConflictError{Message: "mock publisher batch error"}
	}

	p.PublishedEvents = append(p.PublishedEvents, events...)
	return nil
}

// Close no hace nada
func (p *MockPublisher) Close() error {
	return nil
}

// Reset limpia los eventos publicados
func (p *MockPublisher) Reset() {
	p.PublishedEvents = make([]*domain.Event, 0)
	p.ShouldFail = false
	p.FailError = nil
}

// GetEventCount devuelve el número de eventos publicados
func (p *MockPublisher) GetEventCount() int {
	return len(p.PublishedEvents)
}

// GetEventsByType devuelve eventos filtrados por tipo
func (p *MockPublisher) GetEventsByType(eventType string) []*domain.Event {
	filtered := make([]*domain.Event, 0)
	for _, event := range p.PublishedEvents {
		if event.EventType == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// HasEvent verifica si un tipo de evento fue publicado
func (p *MockPublisher) HasEvent(eventType string) bool {
	for _, event := range p.PublishedEvents {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}
