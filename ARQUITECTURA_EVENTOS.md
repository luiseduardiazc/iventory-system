# Arquitectura de Eventos - Sistema de Inventario

## 📋 Resumen

Este documento describe la arquitectura de publicación de eventos implementada en el sistema de inventario, que permite cambiar de broker de mensajes (Redis, Kafka) sin modificar el código de negocio.

## 🎯 Problema Resuelto

### Antes de la Refactorización

```
┌─────────────────┐
│  StockService   │
└────────┬────────┘
         │
         │ (dependencia concreta)
         ▼
┌─────────────────┐
│ EventRepository │──────► SQLite/PostgreSQL
└─────────────────┘

❌ Eventos solo en base de datos (no pub/sub)
❌ Para usar Kafka hay que modificar TODOS los servicios
❌ No se pueden intercambiar brokers
❌ Viola el Principio de Inversión de Dependencias (SOLID)
```

### Después de la Refactorización

```
┌─────────────────┐
│  StockService   │
└────────┬────────┘
         │
         │ (depende de abstracción)
         ▼
┌─────────────────────────┐
│  EventPublisher (IF)    │◄──────────┐
└──────────┬──────────────┘           │
           │                           │
           │ implementan              │
           ▼                           │
    ┌──────────────┐          ┌───────────────┐
    │ RedisPubl.   │          │ KafkaPubl.    │
    │ MockPubl.    │          │ NoOpPubl.     │
    └──────────────┘          └───────────────┘

✅ Eventos publicados a broker real (Redis Streams)
✅ Servicios no saben qué broker se usa
✅ Cambiar broker = cambiar 1 línea en main.go
✅ Cumple con Dependency Inversion Principle
```

## 🏗️ Componentes

### 1. EventPublisher Interface

**Ubicación**: `internal/domain/publisher.go`

```go
type EventPublisher interface {
    Publish(ctx context.Context, event *Event) error
    PublishBatch(ctx context.Context, events []*Event) error
    Close() error
}
```

**Propósito**: Abstracción que define el contrato para publicar eventos sin conocer la implementación concreta.

### 2. RedisPublisher Implementation

**Ubicación**: `internal/infrastructure/redis_publisher.go`

**Características**:
- Usa **Redis Streams** para pub/sub
- Persistencia automática de mensajes
- Consumer Groups para procesamiento distribuido
- Configuración:
  - `REDIS_HOST`, `REDIS_PORT`: Conexión
  - `REDIS_STREAM_NAME`: Nombre del stream (default: `inventory-events`)
  - `REDIS_MAX_LEN`: Retención máxima (default: 100,000 eventos)

**Ventajas**:
- Ligero y rápido
- No requiere infraestructura adicional si ya usas Redis
- Ideal para empezar con pub/sub

### 3. MockPublisher (Tests)

**Ubicación**: `test/mocks/mock_publisher.go`

**Propósito**: Implementación falsa para pruebas unitarias sin broker real.

**Capacidades**:
- Almacena eventos en memoria
- Permite verificar qué eventos se publicaron
- Simular errores de publicación
- Thread-safe

### 4. NoOpPublisher

**Ubicación**: `cmd/api/main.go`

**Propósito**: Implementación nula cuando `MESSAGE_BROKER=none`.

**Uso**: Para entornos donde no se necesita pub/sub (solo persistencia en DB).

## 🔄 Flujo de Eventos

### Publicación de Evento

```
1. Usuario → API Handler
              │
2. Handler → StockService.UpdateStock()
              │
3. StockService:
   ├─► eventRepo.Save(event)        // Persistir en DB (auditoría)
   └─► publisher.Publish(event)      // Publicar a broker (pub/sub)
              │
4. EventPublisher (interface)
              │
5. RedisPublisher.Publish():
   └─► Redis XADD inventory-events {...}
              │
6. ✅ Evento disponible para consumers
```

### Doble Persistencia

**¿Por qué guardamos eventos en DB Y en el broker?**

1. **Base de Datos (EventRepository)**:
   - Auditoría completa
   - Consultas históricas
   - Reconstrucción de estado
   - No expira

2. **Message Broker (EventPublisher)**:
   - Procesamiento en tiempo real
   - Notificaciones a otros servicios
   - Arquitectura event-driven
   - Puede expirar (según retención)

## 📝 Servicios Refactorizados

### StockService

**Métodos con publicación de eventos**:
- `UpdateStock()` → `stock.updated`
- `InitializeStock()` → `stock.created`
- `TransferStock()` → `stock.transferred`

**Patrón implementado**:
```go
// 1. Persistir evento (auditoría)
if err := s.eventRepo.Save(ctx, event); err != nil {
    return err
}

// 2. Publicar evento (pub/sub)
if err := s.publisher.Publish(ctx, event); err != nil {
    log.Printf("⚠️  Error publishing event: %v", err)
    // NO falla la operación si falla la publicación
}
```

### ReservationService

**Métodos con publicación de eventos**:
- `CreateReservation()` → `reservation.created`
- `ConfirmReservation()` → `reservation.confirmed`
- `CancelReservation()` → `reservation.cancelled`
- `ExpireReservation()` → `reservation.expired`

## ⚙️ Configuración

### Variables de Entorno

```bash
# Elegir broker
MESSAGE_BROKER=redis  # redis | kafka | none

# Configuración Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_STREAM_NAME=inventory-events
REDIS_MAX_LEN=100000
```

### Cambiar de Broker

**Redis → Kafka** (una vez implementado):
```bash
MESSAGE_BROKER=kafka
KAFKA_BROKERS=localhost:9092
```

**Sin broker**:
```bash
MESSAGE_BROKER=none
# Los eventos solo se guardan en la base de datos
```

## 🔌 Implementaciones Futuras

### Kafka Publisher (planificado)

```go
// internal/infrastructure/kafka_publisher.go
type KafkaPublisher struct {
    producer sarama.SyncProducer
}

func (p *KafkaPublisher) Publish(ctx context.Context, event *domain.Event) error {
    msg := &sarama.ProducerMessage{
        Topic: "inventory-events",
        Value: sarama.StringEncoder(event.ToJSON()),
    }
    _, _, err := p.producer.SendMessage(msg)
    return err
}
```

## 🧪 Testing

### Unit Tests

**Usar MockPublisher**:
```go
func TestStockService_UpdateStock(t *testing.T) {
    // Crear mock
    mockPublisher := mocks.NewMockPublisher()
    
    // Crear servicio con mock
    service := NewStockService(stockRepo, productRepo, eventRepo, mockPublisher)
    
    // Ejecutar operación
    service.UpdateStock(ctx, storeID, productID, 100, "restock")
    
    // Verificar que se publicó el evento
    require.Equal(t, 1, mockPublisher.Count())
    lastEvent := mockPublisher.GetLastEvent()
    require.Equal(t, "stock.updated", lastEvent.EventType)
}
```

### Integration Tests

Los tests E2E usan `MESSAGE_BROKER=none` para no depender de Redis:
```go
// test/e2e/setup.go
func setupTestServer() {
    os.Setenv("MESSAGE_BROKER", "none")
    // ...
}
```

## 📊 Monitoreo

### Logs de Publicación

```
2025-01-26T21:00:00Z [INFO] Event published to Redis: stock.updated (product: abc123)
2025-01-26T21:00:01Z [INFO] Event published to Redis: reservation.created (id: xyz789)
2025-01-26T21:00:02Z [WARN] Error publishing event to Redis: connection timeout
```

### Métricas Sugeridas

- `events_published_total{type="stock.updated", broker="redis"}`
- `events_publish_errors_total{broker="redis"}`
- `events_publish_duration_seconds{broker="redis"}`

## 🚀 Despliegue

### Con Docker Compose

```yaml
version: '3.8'

services:
  api:
    build: .
    environment:
      MESSAGE_BROKER: redis
      REDIS_HOST: redis
      REDIS_PORT: 6379
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

### Sin Broker (solo DB)

```yaml
services:
  api:
    build: .
    environment:
      MESSAGE_BROKER: none
      DATABASE_DRIVER: postgres
```

## 🎓 Principios SOLID Aplicados

### Dependency Inversion Principle (DIP)

✅ **Antes**: `StockService` dependía de `EventRepository` (implementación concreta)  
✅ **Ahora**: `StockService` depende de `EventPublisher` (abstracción)

**Beneficio**: Los módulos de alto nivel (servicios) no dependen de módulos de bajo nivel (infraestructura).

### Open/Closed Principle (OCP)

✅ **Extensible**: Podemos agregar `KafkaPublisher` sin modificar `StockService`  
✅ **Cerrado**: `EventPublisher` interface no cambia cuando agregamos implementaciones

### Single Responsibility Principle (SRP)

✅ `StockService`: Lógica de negocio de stock  
✅ `RedisPublisher`: Solo se encarga de publicar a Redis  
✅ `EventRepository`: Solo persiste eventos en DB

## 📚 Referencias

- **Redis Streams**: https://redis.io/docs/data-types/streams/
- **Apache Kafka**: https://kafka.apache.org/
- **Event-Driven Architecture**: https://martinfowler.com/articles/201701-event-driven.html
- **Dependency Inversion**: https://en.wikipedia.org/wiki/Dependency_inversion_principle

## ✅ Checklist de Implementación

- [x] EventPublisher interface creada
- [x] RedisPublisher implementado
- [x] MockPublisher para tests
- [x] StockService refactorizado
- [x] ReservationService refactorizado
- [x] Configuración MESSAGE_BROKER agregada
- [x] main.go con inicialización de publisher
- [x] Tests pasando (74/74)
- [x] Documentación actualizada
- [ ] KafkaPublisher implementado (futuro)
- [ ] Métricas de publicación (futuro)
- [ ] Consumer de eventos (futuro)

---

**Última actualización**: 2025-01-26  
**Autor**: Sistema de Inventario - Refactorización de Escalabilidad
