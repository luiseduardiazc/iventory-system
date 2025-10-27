package domain

import "context"

// EventPublisher define el contrato para publicar eventos a un message broker.
// Esta interfaz permite cambiar entre diferentes implementaciones (NATS, Kafka, Redis)
// sin modificar la lógica de negocio (Dependency Inversion Principle).
//
// Implementaciones disponibles:
//   - NATSPublisher: Para NATS JetStream (baja latencia, distribuido)
//   - KafkaPublisher: Para Apache Kafka (alto throughput, retención larga)
//   - RedisPublisher: Para Redis Streams (simple, rápido setup)
//   - MockPublisher: Para tests unitarios
type EventPublisher interface {
	// Publish publica un evento al message broker.
	// El evento se envía de forma asíncrona a todos los subscribers interesados.
	//
	// Parámetros:
	//   - ctx: Context para cancelación y timeout
	//   - event: Evento a publicar
	//
	// Retorna error si:
	//   - La conexión al broker falla
	//   - El evento no puede ser serializado
	//   - El broker rechaza el mensaje
	Publish(ctx context.Context, event *Event) error

	// PublishBatch publica múltiples eventos en una sola operación.
	// Útil para optimizar rendimiento cuando se tienen muchos eventos.
	//
	// Parámetros:
	//   - ctx: Context para cancelación y timeout
	//   - events: Slice de eventos a publicar
	//
	// Retorna error si algún evento falla en publicarse.
	PublishBatch(ctx context.Context, events []*Event) error

	// Close cierra la conexión al message broker de forma ordenada.
	// Debe ser llamado al finalizar la aplicación (típicamente con defer).
	Close() error
}
