package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"inventory-system/internal/domain"

	"github.com/redis/go-redis/v9"
)

// RedisPublisher implementa EventPublisher usando Redis Streams.
// Redis Streams es una opción ligera y rápida para pub/sub con persistencia.
type RedisPublisher struct {
	client     *redis.Client
	streamName string
	maxLen     int64 // Máximo de eventos a retener
}

// RedisPublisherConfig configuración para RedisPublisher
type RedisPublisherConfig struct {
	Addr       string // "localhost:6379"
	Password   string // "" para sin password
	DB         int    // 0 por defecto
	StreamName string // Nombre del stream (ej: "inventory-events")
	MaxLen     int64  // Máximo eventos a retener (0 = ilimitado)
}

// NewRedisPublisher crea una nueva instancia de RedisPublisher.
//
// Ejemplo:
//
//	publisher := NewRedisPublisher(RedisPublisherConfig{
//	    Addr:       "localhost:6379",
//	    StreamName: "inventory-events",
//	    MaxLen:     100000, // Retener últimos 100k eventos
//	})
//	defer publisher.Close()
func NewRedisPublisher(cfg RedisPublisherConfig) (*RedisPublisher, error) {
	// Valores por defecto
	if cfg.StreamName == "" {
		cfg.StreamName = "inventory-events"
	}
	if cfg.MaxLen == 0 {
		cfg.MaxLen = 100000 // Default: 100k eventos
	}

	// Crear cliente Redis
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("✅ Connected to Redis at %s (stream: %s)", cfg.Addr, cfg.StreamName)

	return &RedisPublisher{
		client:     client,
		streamName: cfg.StreamName,
		maxLen:     cfg.MaxLen,
	}, nil
}

// Publish publica un evento a Redis Streams.
//
// El evento se serializa a JSON y se añade al stream con los siguientes campos:
//   - id: ID del evento
//   - event_type: Tipo de evento (stock.updated, reservation.created, etc.)
//   - store_id: Tienda origen del evento
//   - aggregate_id: ID del agregado (producto, reserva, etc.)
//   - payload: JSON completo del evento
//   - timestamp: Unix timestamp
func (p *RedisPublisher) Publish(ctx context.Context, event *domain.Event) error {
	// Serializar evento completo a JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publicar a Redis Stream
	args := &redis.XAddArgs{
		Stream: p.streamName,
		MaxLen: p.maxLen,
		Approx: true, // ~MaxLen (más eficiente que exacto)
		Values: map[string]interface{}{
			"id":           event.ID,
			"event_type":   event.EventType,
			"store_id":     event.StoreID,
			"aggregate_id": event.AggregateID,
			"payload":      string(eventJSON),
			"timestamp":    event.CreatedAt.Unix(),
		},
	}

	if err := p.client.XAdd(ctx, args).Err(); err != nil {
		return fmt.Errorf("failed to publish event to Redis: %w", err)
	}

	log.Printf("📤 Event published to Redis: type=%s, store=%s, id=%s",
		event.EventType, event.StoreID, event.ID)

	return nil
}

// PublishBatch publica múltiples eventos usando Redis Pipeline.
// Esto es más eficiente que llamar Publish() múltiples veces.
func (p *RedisPublisher) PublishBatch(ctx context.Context, events []*domain.Event) error {
	if len(events) == 0 {
		return nil
	}

	// Usar pipeline para batch
	pipe := p.client.Pipeline()

	for _, event := range events {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event %s: %w", event.ID, err)
		}

		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: p.streamName,
			MaxLen: p.maxLen,
			Approx: true,
			Values: map[string]interface{}{
				"id":           event.ID,
				"event_type":   event.EventType,
				"store_id":     event.StoreID,
				"aggregate_id": event.AggregateID,
				"payload":      string(eventJSON),
				"timestamp":    event.CreatedAt.Unix(),
			},
		})
	}

	// Ejecutar pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to publish batch to Redis: %w", err)
	}

	log.Printf("📤 Batch published to Redis: %d events", len(events))

	return nil
}

// Close cierra la conexión a Redis.
func (p *RedisPublisher) Close() error {
	if p.client != nil {
		log.Printf("🔌 Closing Redis connection")
		return p.client.Close()
	}
	return nil
}
