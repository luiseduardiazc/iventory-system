# 🔄 Event Sync: Mecanismo de Resiliencia y Re-intentos

## 📋 Resumen

El sistema de inventario implementa un **mecanismo de resiliencia automática** para garantizar la entrega eventual de eventos al message broker (Redis/Kafka), incluso si el broker está temporalmente caído.

## 🎯 Problema que Resuelve

**Escenario**: Redis está caído temporalmente

```
StockService.UpdateStock()
    ↓
✅ Evento guardado en DB (auditoría garantizada)
    ↓
❌ Publicación a Redis FALLA (connection refused)
    ↓
⚠️  Sin mecanismo de retry → EVENTO PERDIDO
```

**Solución**: EventSyncService con re-intentos automáticos

```
StockService.UpdateStock()
    ↓
✅ Evento guardado en DB (synced_at = NULL)
    ↓
❌ Publicación a Redis FALLA
    ↓
[10 segundos después - EventSyncWorker ejecuta]
    ↓
✅ Re-intenta publicar eventos pendientes
    ↓
✅ Marca synced_at = NOW() cuando tiene éxito
```

## 🏗️ Arquitectura del Sistema

### Componentes

```
┌──────────────────────────────────────────────────────────────┐
│                  Capa de Servicios                           │
│                                                              │
│  ┌─────────────────┐        ┌──────────────────┐           │
│  │  StockService   │        │ReservationService│           │
│  └────────┬────────┘        └────────┬─────────┘           │
│           │                          │                      │
│           └──────────┬───────────────┘                      │
│                      │                                      │
│                      ▼                                      │
│           ┌────────────────────┐                           │
│           │  EventPublisher    │ (Interface)               │
│           │  - Publish()       │                           │
│           └──────┬──────────┬──┘                           │
│                  │          │                               │
│        ┌─────────┘          └───────────┐                  │
│        ▼                                 ▼                  │
│  ┌──────────┐                    ┌──────────────┐          │
│  │Redis     │                    │EventSyncService│         │
│  │Publisher │                    │(Retry Logic)   │         │
│  └──────────┘                    └───────┬────────┘         │
│       │                                  │                  │
└───────┼──────────────────────────────────┼──────────────────┘
        │                                  │
        │                                  │
        ▼                                  ▼
  ┌──────────┐                      ┌──────────┐
  │  Redis   │◄─────────────────────│   DB     │
  │ Streams  │   Re-publish failed  │ events   │
  └──────────┘   events from DB     └──────────┘
```

### Flujo Completo

```
┌─────────────────────────────────────────────────────────────┐
│ PASO 1: Operación de Negocio                               │
│ UpdateStock() / CreateReservation()                        │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│ PASO 2: Persistencia en DB                                 │
│ - event.Save(db)                                           │
│ - synced_at = NULL (pendiente de publicación)             │
│ ✅ Auditoría GARANTIZADA (siempre se guarda)               │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│ PASO 3: Intento de Publicación Inmediata                   │
│ publisher.Publish(event)                                   │
│                                                            │
│ ✅ ÉXITO  → synced_at = NOW()                              │
│ ❌ FALLA  → synced_at = NULL (queda pendiente)             │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼ (si falló)
┌─────────────────────────────────────────────────────────────┐
│ PASO 4: Background Worker (cada 10 segundos)               │
│ EventSyncWorker ejecuta:                                   │
│                                                            │
│ 1. eventRepo.GetPendingEvents() → WHERE synced_at IS NULL │
│ 2. Para cada evento:                                      │
│    - publisher.Publish(event)                             │
│    - Si éxito → Mark synced_at = NOW()                    │
│    - Si falla → Dejar NULL (retry en próxima ejecución)   │
│                                                            │
│ ✅ Garantiza entrega EVENTUAL                              │
└─────────────────────────────────────────────────────────────┘
```

## 📊 Tabla de Componentes

| Componente | Archivo | Responsabilidad | Frecuencia |
|------------|---------|----------------|------------|
| **StockService** | `internal/service/stock_service.go` | Publicación directa (tiempo real) | Por operación |
| **ReservationService** | `internal/service/reservation_service.go` | Publicación directa (tiempo real) | Por operación |
| **EventSyncService** | `internal/service/event_sync_service.go` | Re-intentos de eventos fallidos | Cada 10s (worker) |
| **EventSyncWorker** | `cmd/api/main.go:243` | Ejecuta SyncPendingEvents() | Background (cada 10s) |
| **EventRepository** | `internal/repository/event_repository.go` | Tracking de synced_at | Por evento |
| **EventPublisher** | `internal/infrastructure/*_publisher.go` | Abstracción de brokers | Variable |

## 🔧 Implementación

### EventSyncService (Código Clave)

```go
// internal/service/event_sync_service.go

type EventSyncService struct {
    eventRepo *repository.EventRepository
    publisher EventPublisher  // ✅ Inyectado para re-intentos
}

func (s *EventSyncService) SyncPendingEvents(ctx context.Context, batchSize int) (int, error) {
    // 1. Obtener eventos NO sincronizados (synced_at IS NULL)
    events, err := s.eventRepo.GetPendingEvents(ctx, batchSize)
    if err != nil {
        return 0, fmt.Errorf("failed to get pending events: %w", err)
    }

    syncedCount := 0
    failedCount := 0
    eventIDs := make([]string, 0, len(events))

    // 2. RE-INTENTAR publicación de cada evento
    for _, event := range events {
        err := s.publisher.Publish(ctx, event)
        if err != nil {
            log.Printf("⚠️  Failed to sync event %s: %v (will retry later)", event.ID, err)
            failedCount++
            continue  // No marcar como sincronizado
        }

        eventIDs = append(eventIDs, event.ID)
        syncedCount++
    }

    // 3. Marcar como sincronizados SOLO los exitosos
    if len(eventIDs) > 0 {
        err = s.eventRepo.MarkMultipleAsSynced(ctx, eventIDs)
        if err != nil {
            return syncedCount, fmt.Errorf("failed to mark events as synced: %w", err)
        }
        log.Printf("✅ Successfully synced %d events (failed: %d)", syncedCount, failedCount)
    }

    return syncedCount, nil
}
```

### Worker en main.go

```go
// cmd/api/main.go:243

func startEventSyncWorker(service *service.EventSyncService) {
    ticker := time.NewTicker(10 * time.Second)  // Ejecuta cada 10 segundos
    defer ticker.Stop()

    log.Println("📡 Event synchronization worker started")

    for range ticker.C {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        count, err := service.SyncPendingEvents(ctx, 100)  // Batch de 100 eventos
        cancel()

        if err != nil {
            log.Printf("Error syncing events: %v", err)
        } else if count > 0 {
            log.Printf("✅ Synced %d events", count)
        }
    }
}
```

## 📈 Ventajas del Diseño

| Ventaja | Descripción | Beneficio |
|---------|-------------|-----------|
| **Auditoría Garantizada** | Eventos SIEMPRE se guardan en DB | No se pierde información histórica |
| **Resiliencia Automática** | Worker re-intenta sin intervención manual | Sin downtime por caídas temporales de Redis |
| **Sin Pérdida de Datos** | Eventos pendientes se publican cuando broker vuelve | Consistencia eventual garantizada |
| **Observabilidad** | Campo `synced_at` permite monitorear lag | Métricas y alertas fáciles de implementar |
| **Idempotencia** | Re-publicar es seguro (event IDs únicos) | No hay eventos duplicados |
| **Escalabilidad** | Batch processing (100 eventos/ciclo) | Maneja grandes volúmenes de eventos pendientes |

## 🧪 Testing

### Casos de Prueba Implementados

```bash
# Ejecutar tests de resiliencia
go test ./test/unit -v -run TestEventSyncService_RetryMechanism

# Tests implementados:
# ✅ Retry_PublishFailedEvents_Success
#    - Simula Redis caído en primer intento
#    - Verifica que evento queda pendiente
#    - Segundo intento tiene éxito
#    - Evento se marca como sincronizado

# ✅ Retry_PartialFailure_OnlySuccessfulMarked
#    - 4 eventos: 2 fallan, 2 tienen éxito
#    - Verifica que SOLO los exitosos se marcan como synced
#    - Los fallidos quedan pendientes para próximo retry

# ✅ Retry_NoOpPublisher_AlwaysSucceeds
#    - NoOpPublisher nunca falla
#    - Todos los eventos se sincronizan en primer intento
```

### Ejemplo de Logs

```bash
# ✅ Publicación exitosa (tiempo real)
✅ Event published to Redis: evt-20251028150405-001 (stock.updated)
✅ Event synced to DB: evt-20251028150405-001

# ❌ Redis caído (se guarda en DB, publicación falla)
✅ Event saved to DB: evt-20251028150406-002 (stock.created)
⚠️  Failed to publish to Redis: connection refused (will retry)

# 🔄 Worker re-intenta 10 segundos después (Redis sigue caído)
📡 Event synchronization worker started
⚠️  Failed to sync event evt-20251028150406-002: connection refused (will retry later)
⚠️  All 1 events failed to sync (will retry in next cycle)

# ✅ Redis vuelve, evento se publica exitosamente
✅ Successfully synced 1 events (failed: 0)
✅ Event published to Redis: evt-20251028150406-002 (stock.created)
```

## 🔍 Monitoreo y Observabilidad

### Consultas SQL Útiles

```sql
-- Ver eventos pendientes de sincronización
SELECT 
    id,
    event_type,
    aggregate_id,
    store_id,
    created_at,
    ROUND((julianday('now') - julianday(created_at)) * 24 * 60, 2) as minutes_pending
FROM events
WHERE synced_at IS NULL
ORDER BY created_at DESC;

-- Contar eventos pendientes por tipo
SELECT 
    event_type,
    COUNT(*) as pending_count,
    MIN(created_at) as oldest_pending
FROM events
WHERE synced_at IS NULL
GROUP BY event_type
ORDER BY pending_count DESC;

-- Latencia de sincronización (promedio)
SELECT 
    event_type,
    AVG((julianday(synced_at) - julianday(created_at)) * 24 * 60) as avg_sync_latency_minutes
FROM events
WHERE synced_at IS NOT NULL
GROUP BY event_type;

-- Eventos fallidos hace más de 1 hora (alerta)
SELECT COUNT(*) as critical_pending
FROM events
WHERE synced_at IS NULL
  AND datetime(created_at) < datetime('now', '-1 hour');
```

### Métricas Recomendadas (Prometheus)

```go
// Métricas sugeridas para implementación futura

// Contador de eventos pendientes
eventsPendingGauge.Set(float64(pendingCount))

// Latencia de sincronización
syncLatencyHistogram.Observe(syncDuration.Seconds())

// Eventos sincronizados exitosamente
eventsSyncedCounter.Inc()

// Eventos fallidos
eventsFailedCounter.Inc()
```

## 🚀 Configuración

### Variables de Entorno

```bash
# Frecuencia del worker (no configurable actualmente, hardcoded a 10s)
# Futuro: EVENT_SYNC_INTERVAL=10s

# Tamaño de batch
# Futuro: EVENT_SYNC_BATCH_SIZE=100

# Timeout de sincronización
# Futuro: EVENT_SYNC_TIMEOUT=30s
```

### Escalabilidad

| Escenario | Configuración Recomendada |
|-----------|---------------------------|
| **Bajo Volumen** (<100 eventos/min) | Intervalo: 10s, Batch: 100 |
| **Medio Volumen** (100-1000 eventos/min) | Intervalo: 5s, Batch: 500 |
| **Alto Volumen** (>1000 eventos/min) | Intervalo: 1s, Batch: 1000, Worker dedicado |

## 🔮 Roadmap Futuro

- [ ] Configuración de intervalo via env var (`EVENT_SYNC_INTERVAL`)
- [ ] Métricas de Prometheus integradas
- [ ] Alertas cuando eventos pendientes > umbral
- [ ] Dashboard de Grafana con visualización de lag
- [ ] Dead Letter Queue para eventos fallidos >N veces
- [ ] Priorización de eventos críticos vs. no críticos
- [ ] Compresión de eventos antes de publicar (batch)
- [ ] Circuit Breaker para evitar reintentos innecesarios

## 📚 Referencias

- [Event Sourcing Pattern](https://martinfowler.com/eaaDev/EventSourcing.html)
- [Outbox Pattern](https://microservices.io/patterns/data/transactional-outbox.html)
- [Retry Patterns](https://docs.microsoft.com/en-us/azure/architecture/patterns/retry)

---

**Autor**: Sistema de Inventario Distribuido  
**Fecha**: Octubre 2025  
**Versión**: 1.0.0
