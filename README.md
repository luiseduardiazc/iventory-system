# Sistema de Gestión de Inventario Distribuido

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 🎯 Objetivo

Prototipo de sistema de gestión de inventario distribuido que optimiza la consistencia del inventario, reduce la latencia en las actualizaciones de stock y minimiza los costos operativos mediante una arquitectura event-driven.

## ✨ Características

- **Event-Driven Architecture**: Publicación de eventos en tiempo real con brokers intercambiables (Redis Streams, Kafka)
- **Message Broker Flexible**: Arquitectura desacoplada que permite cambiar de Redis a Kafka sin modificar código de negocio
- **Optimistic Locking**: Previene overselling manteniendo alta disponibilidad
- **Reservas con TTL**: Auto-expiración de reservas para liberar stock automáticamente
- **Event Sourcing**: Auditoría completa de eventos en base de datos + publicación en tiempo real
- **Arquitectura SOLID**: Dependency Inversion Principle para escalabilidad y mantenibilidad

## 🏗️ Arquitectura: Event-Driven Escalable

El sistema utiliza una **arquitectura event-driven** con **brokers intercambiables** siguiendo el **Dependency Inversion Principle**.

```
┌─────────────────────────────────────────────────────────────┐
│                    Servicios de Negocio                      │
│  (StockService, ReservationService, ProductService)         │
└────────────────┬────────────────────────────────────────────┘
                 │
                 │ dependen de
                 ▼
┌─────────────────────────────────────────────────────────────┐
│              EventPublisher (Interface)                      │
│  - Publish(event)                                           │
│  - PublishBatch(events)                                     │
│  - Close()                                                  │
└────────────────┬────────────────────────────────────────────┘
                 │
                 │ implementan
    ┌────────────┼────────────┬──────────┐
    ▼            ▼            ▼          ▼
┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
│ Redis   │ │ Kafka   │ │ Mock    │ │ NoOp    │
│Publisher│ │Publisher│ │Publisher│ │Publisher│
│   ✅    │ │   🔜    │ │   ✅    │ │   ✅    │
└─────────┘ └─────────┘ └─────────┘ └─────────┘
```

### Doble Persistencia: DB + Broker

```
┌──────────────┐
│   Evento     │
└──────┬───────┘
       │
   ┌───┴────┐
   │        │
   ▼        ▼
┌────┐  ┌──────┐
│ DB │  │Broker│
└────┘  └──────┘
  │        │
  │        └─────► Procesamiento en tiempo real
  │                Notificaciones
  │                Microservicios
  │
  └──────────────► Auditoría
                   Event Sourcing
                   Reconstrucción de estado
```

### 🔄 Mecanismo de Resiliencia: Retry Automático

El sistema implementa un **mecanismo de re-intentos automáticos** para garantizar la entrega eventual de eventos, incluso si el broker (Redis/Kafka) está temporalmente caído.

**Flujo Completo:**

```
┌────────────────────────────────────────────────────────────────┐
│ 1. Operación de Negocio (ej: UpdateStock)                     │
└──────────────────────┬─────────────────────────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────────────────┐
│ 2. Guardar Evento en DB (synced_at = NULL)                    │
│    ✅ SIEMPRE se persiste (auditoría garantizada)              │
└──────────────────────┬─────────────────────────────────────────┘
                       │
                       ▼
┌────────────────────────────────────────────────────────────────┐
│ 3. Publicar en Broker (Redis/Kafka)                           │
│    ✅ Éxito  → synced_at = NOW()                               │
│    ❌ Falla  → synced_at = NULL (queda pendiente)              │
└──────────────────────┬─────────────────────────────────────────┘
                       │
                       ▼ (si falló)
┌────────────────────────────────────────────────────────────────┐
│ 4. EventSyncWorker (cada 10 segundos)                         │
│    - Busca eventos con synced_at = NULL                       │
│    - RE-INTENTA publicar en el broker                         │
│    - Marca synced_at = NOW() si tiene éxito                   │
│    ✅ Garantiza entrega eventual                               │
└────────────────────────────────────────────────────────────────┘
```

**Componentes del Sistema de Resiliencia:**

| Componente | Responsabilidad | Frecuencia |
|------------|----------------|------------|
| **StockService** | Publicación directa (tiempo real) | Por operación |
| **ReservationService** | Publicación directa (tiempo real) | Por operación |
| **EventSyncService** | Re-intentos de eventos fallidos | Cada 10 segundos |
| **EventSyncWorker** | Ejecuta SyncPendingEvents() | Background (cada 10s) |
| **EventRepository** | Persistencia + tracking de synced_at | Por evento |

**Ventajas de este Diseño:**

- ✅ **Auditoría garantizada**: Eventos SIEMPRE se guardan en DB, incluso si Redis cae
- ✅ **Resiliencia automática**: Worker re-intenta publicaciones fallidas sin intervención manual
- ✅ **Sin pérdida de datos**: Eventos pendientes se publican cuando el broker vuelve
- ✅ **Observabilidad**: Campo `synced_at` permite monitorear eventos pendientes
- ✅ **Idempotencia**: Re-publicar es seguro gracias a event IDs únicos

**Ejemplo de Logs:**

```bash
# Publicación exitosa (tiempo real)
✅ Event published to Redis: evt-20251028150405-001 (stock.updated)
✅ Event synced to DB: evt-20251028150405-001

# Redis caído (se guarda en DB, publicación falla)
✅ Event saved to DB: evt-20251028150406-002 (stock.created)
⚠️  Failed to publish to Redis: connection refused (will retry)

# Worker re-intenta 10 segundos después
📡 Event synchronization worker started
⚠️  Failed to sync event evt-20251028150406-002: connection refused (will retry later)

# Redis vuelve, evento se publica exitosamente
✅ Successfully synced 1 events (failed: 0)
✅ Event published to Redis: evt-20251028150406-002 (stock.created)
```

**Consultar Eventos Pendientes:**

```sql
-- Ver eventos que NO se han sincronizado con el broker
SELECT id, event_type, aggregate_id, store_id, created_at
FROM events
WHERE synced_at IS NULL
ORDER BY created_at DESC;

-- Contar eventos pendientes por tipo
SELECT event_type, COUNT(*) as pending_count
FROM events
WHERE synced_at IS NULL
GROUP BY event_type;
```

### Cambio de Broker en 1 Línea

```bash
# Redis Streams (actual)
MESSAGE_BROKER=redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Kafka (futuro - solo crear kafka_publisher.go)
MESSAGE_BROKER=kafka
KAFKA_BROKERS=localhost:9092

# Sin broker (solo DB)
MESSAGE_BROKER=none
```

**Beneficio Clave**: Cambiar de Redis a Kafka NO requiere modificar servicios, solo crear la implementación del publisher.

## 🚀 Quick Start

📖 **Para instrucciones detalladas de ejecución y troubleshooting, consulta [docs/run.md](docs/run.md)**

### Inicio Rápido

#### Prerrequisitos

- Go 1.21+ ([Descargar aquí](https://golang.org/dl/))
- Docker y Docker Compose (para Redis - ver [run.md](docs/run.md))

#### Instalación

```bash
# 1. Clonar repositorio
git clone <repository-url>
cd inventory-system

# 2. Instalar dependencias
go mod download

# 3. Copiar configuración
cp .env.example .env
```

#### Redis

```bash
# Iniciar Redis
docker-compose up -d redis

# Configurar .env
MESSAGE_BROKER=redis

# Ejecutar
go run cmd/api/main.go
```

#### Verificación

```bash
# Health check
curl http://localhost:8080/health
```

📋 **Más comandos, ejemplos de API y solución de problemas en [docs/run.md](docs/run.md)**

## 📚 Documentación

### 📖 Guías de Usuario

- **[Guía de Ejecución](docs/run.md)** - Instrucciones completas para ejecutar el sistema

### 🔄 Sistema de Resiliencia

- **[Event Sync Resilience](docs/EVENT_SYNC_RESILIENCE.md)** - Guía completa del mecanismo de re-intentos automáticos
- **[Implementación Event Sync](docs/IMPLEMENTACION_EVENT_SYNC_COMPLETA.md)** - Resumen de la implementación del sistema de resiliencia

### 📊 Diagramas de Arquitectura

#### Diagramas Principales
- **[Diagrama de Arquitectura](docs/architecture-diagram.mmd)** - Vista completa del sistema (técnico y negocio)
- **[Diagrama de Negocio](docs/business-architecture-diagram.mmd)** - Vista ejecutiva simplificada
- **[Diagrama de Flujo de Valor](docs/value-flow-diagram.mmd)** - Customer journey y procesos

#### Diagramas de Resiliencia
- **[Flujo de Resiliencia](docs/resilience-flow-diagram.mmd)** - Secuencia de re-intentos automáticos ([PNG](docs/resilience-flow-diagram.png))
- **[Flujo de Valor](docs/value-flow-diagram.mmd)** - Journey del cliente ([PNG](docs/value-flow-diagram.png))

> 💡 **Tip**: Los archivos `.mmd` se pueden visualizar en [mermaid.live](https://mermaid.live/) o con extensiones de VS Code

### 🎯 Para Comenzar

1. **Primera vez**: Lee [run.md](docs/run.md) para configurar el entorno

**Implementado:**
- ✅ **EventPublisher Interface** - Abstracción para brokers intercambiables
- ✅ **RedisPublisher** - Implementación con Redis Streams (176 líneas)
- ✅ **MockPublisher** - Para tests sin dependencias externas (100 líneas)
- ✅ **StockService** refactorizado - Publica eventos stock.updated, stock.created, stock.transferred
- ✅ **ReservationService** refactorizado - Publica eventos reservation.*
- ✅ **74/74 tests pasando** - Sin regresiones
- ✅ **Arquitectura SOLID** - Dependency Inversion Principle aplicado
- ✅ **Compilación exitosa** - Sin errores ni warnings

## 📡 API Endpoints

Base URL: `http://localhost:8080/api/v1`

### 🏥 Health Check

| Método | Endpoint | Descripción | Auth | Event |
|--------|----------|-------------|------|---------|
| `GET` | `/health` | Estado del servidor y base de datos | No | ❌ |

### 📦 Products (Productos)

| Método | Endpoint | Descripción | Auth | Event |
|--------|----------|-------------|------|---------|
| `GET` | `/products` | Listar todos los productos (paginado) | No | ❌ |
| `GET` | `/products/:id` | Obtener producto por ID | No | ❌ |
| `GET` | `/products/sku/:sku` | Obtener producto por SKU | No | ❌ |
| `POST` | `/products` | Crear nuevo producto | ✅ API Key | ❌ |
| `PUT` | `/products/:id` | Actualizar producto existente | ✅ API Key | ❌ |
| `DELETE` | `/products/:id` | Eliminar producto | ✅ API Key | ❌ |

**Nota**: Los productos NO generan eventos pub/sub (solo operaciones CRUD simples).

---

### 📊 Stock (Inventario)

Todos los endpoints de stock requieren **API Key** authentication.

| Método | Endpoint | Descripción | Event |
|--------|----------|-------------|---------------|
| `POST` | `/stock` | Inicializar stock para producto/tienda | ✅ `stock.created` |
| `GET` | `/stock/product/:productId` | Obtener stock de un producto en todas las tiendas | ❌ |
| `GET` | `/stock/store/:storeId` | Obtener todo el stock de una tienda | ❌ |
| `GET` | `/stock/low-stock` | Obtener productos con stock bajo | ❌ |
| `GET` | `/stock/:productId/:storeId` | Obtener stock específico producto/tienda | ❌ |
| `GET` | `/stock/:productId/:storeId/availability` | Verificar disponibilidad | ❌ |
| `PUT` | `/stock/:productId/:storeId` | Actualizar stock (restock/ajuste) | ✅ `stock.updated` |
| `POST` | `/stock/:productId/:storeId/adjust` | Ajustar stock (incremento/decremento) | ✅ `stock.updated` |
| `POST` | `/stock/transfer` | Transferir stock entre tiendas | ✅ `stock.transferred` |

**Eventos Publicados:**

```json
// stock.created
{
  "event_type": "stock.created",
  "aggregate_id": "product-123",
  "store_id": "MAD-001",
  "payload": {
    "product_id": "product-123",
    "store_id": "MAD-001",
    "quantity": 100,
    "reason": "initial_stock"
  }
}

// stock.updated
{
  "event_type": "stock.updated",
  "aggregate_id": "product-123",
  "store_id": "MAD-001",
  "payload": {
    "product_id": "product-123",
    "store_id": "MAD-001",
    "previous_quantity": 100,
    "new_quantity": 150,
    "change": 50,
    "reason": "restock"
  }
}

// stock.transferred
{
  "event_type": "stock.transferred",
  "aggregate_id": "product-123",
  "payload": {
    "product_id": "product-123",
    "from_store": "MAD-001",
    "to_store": "BCN-001",
    "quantity": 20,
    "reason": "transfer"
  }
}
```

---

### 🎫 Reservations (Reservas)

Todos los endpoints de reservations requieren **API Key** authentication.

| Método | Endpoint | Descripción | Event |
|--------|----------|-------------|---------------|
| `POST` | `/reservations` | Crear nueva reserva | ✅ `reservation.created` |
| `GET` | `/reservations/:id` | Obtener reserva por ID | ❌ |
| `POST` | `/reservations/:id/confirm` | Confirmar reserva (finalizar venta) | ✅ `reservation.confirmed` |
| `POST` | `/reservations/:id/cancel` | Cancelar reserva (liberar stock) | ✅ `reservation.cancelled` |
| `GET` | `/reservations/store/:storeId/pending` | Listar reservas pendientes de una tienda | ❌ |
| `GET` | `/reservations/product/:productId/store/:storeId` | Listar reservas de un producto | ❌ |
| `GET` | `/reservations/stats` | Obtener estadísticas de reservas | ❌ |

**Eventos Publicados:**

```json
// reservation.created
{
  "event_type": "reservation.created",
  "aggregate_id": "reservation-456",
  "store_id": "MAD-001",
  "payload": {
    "reservation_id": "reservation-456",
    "product_id": "product-123",
    "store_id": "MAD-001",
    "quantity": 5,
    "customer_id": "customer-789",
    "expires_at": "2025-10-26T22:00:00Z"
  }
}

// reservation.confirmed
{
  "event_type": "reservation.confirmed",
  "aggregate_id": "reservation-456",
  "store_id": "MAD-001",
  "payload": {
    "reservation_id": "reservation-456",
    "product_id": "product-123",
    "quantity": 5,
    "confirmed_at": "2025-10-26T21:45:00Z"
  }
}

// reservation.cancelled
{
  "event_type": "reservation.cancelled",
  "aggregate_id": "reservation-456",
  "store_id": "MAD-001",
  "payload": {
    "reservation_id": "reservation-456",
    "product_id": "product-123",
    "quantity": 5,
    "reason": "manual_cancellation"
  }
}

// reservation.expired (generado automáticamente por worker)
{
  "event_type": "reservation.expired",
  "aggregate_id": "reservation-456",
  "store_id": "MAD-001",
  "payload": {
    "reservation_id": "reservation-456",
    "product_id": "product-123",
    "quantity": 5,
    "expired_at": "2025-10-26T22:00:00Z"
  }
}
```

---

### 📊 Resumen de Eventos

**Total de Endpoints**: 29  
**Endpoints que generan eventos**: 7 (24%)

| Evento | Trigger | Propósito |
|--------|---------|-----------|
| `stock.created` | POST `/stock` | Notificar inicialización de inventario |
| `stock.updated` | PUT/POST `/stock/...` | Notificar cambios de cantidad en stock |
| `stock.transferred` | POST `/stock/transfer` | Notificar transferencias entre tiendas |
| `reservation.created` | POST `/reservations` | Notificar nueva reserva de stock |
| `reservation.confirmed` | POST `/reservations/:id/confirm` | Notificar venta completada |
| `reservation.cancelled` | POST `/reservations/:id/cancel` | Notificar cancelación manual |
| `reservation.expired` | Worker automático | Notificar expiración por TTL |

**Consumo de Eventos**: Los eventos se pueden consumir desde:
- **Redis Streams** (actual): `XREAD` sobre stream `inventory-events`
- **Kafka** (futuro): Topic `inventory-events`
- **Base de datos**: Tabla `events` para auditoría

## 📊 Stack Tecnológico

| Categoría | Tecnología | Justificación |
|-----------|-----------|---------------|
| **Lenguaje** | Go 1.21+ | Concurrencia nativa, performance, simplicidad |
| **Web Framework** | Gin | Ligero, rápido, rico ecosistema de middleware |
| **Base de Datos** | SQLite | Ligera, embebida, sin configuración, persistencia de eventos (Event Sourcing) |
| **Message Broker** | Redis Streams ✅ | Tiempo real para eventos (intercambiable con Kafka 🔜 u otro Broker) |
| **Event Store** | SQLite (tabla `events`) | Auditoría completa, tracking de sincronización (`synced_at`) |
| **Resiliencia** | EventSyncWorker | Re-intentos automáticos cada 10s para eventos fallidos |
| **Patrón Arquitectónico** | Event-Driven + DIP | Publisher interface para brokers intercambiables |
| **Concurrencia** | Goroutines + Workers | Background workers para expiración de reservas y sync de eventos |
| **Testing** | Go testing + Mocks | 60+ tests unitarios, mocks in-memory, cobertura completa |


## 📋 Roadmap Futuro 
- [ ] Consumer de eventos (microservicio separado)
- [ ] Métricas de publicación (Prometheus)
- [ ] Dashboard de monitoreo (Grafana)

