# Plan de Implementación - Sistema de Inventario Distribuido

## 🎯 Alineación con Objetivos del Proyecto

Este plan implementa un **prototipo de sistema de gestión de inventario distribuido** que resuelve los problemas del sistema actual mediante:

1. **Arquitectura Event-Driven**: Reemplaza la sincronización periódica (cada 15 min) por actualizaciones en tiempo real basadas en eventos
2. **Consistencia Optimista**: Maneja concurrencia priorizando disponibilidad con consistencia eventual, usando optimistic locking para prevenir conflictos
3. **API REST bien diseñada**: Operaciones clave (Ver stock, Actualizar stock, Reservar producto) con endpoints RESTful
4. **Observabilidad**: Logging estructurado, métricas, y health checks para monitoreo en producción
5. **Seguridad básica**: Autenticación JWT, validación de inputs, y rate limiting
6. **Testing completo**: Unitarios e integración cubriendo escenarios de concurrencia

### Justificación de Decisiones Arquitectónicas

| Aspecto | Decisión | Justificación |
|---------|----------|---------------|
| **Modelo de Sincronización** | Event-Driven con NATS JetStream | ✅ Latencia <100ms vs 15 min actual<br>✅ Consistencia eventual garantizada<br>✅ Desacoplamiento entre tiendas<br>✅ Replay de eventos para recovery |
| **Patrón Arquitectónico** | CQRS simplificado | ✅ Separación lectura/escritura<br>✅ Optimización independiente<br>✅ Cache selectivo en lecturas |
| **Manejo de Concurrencia** | Optimistic Locking + Reservas TTL | ✅ Alta disponibilidad (lock-free)<br>✅ Previene overselling<br>✅ Auto-liberación de reservas expiradas |
| **Persistencia** | PostgreSQL (producción) + SQLite (dev/test) | ✅ ACID completo<br>✅ SQLite para testing sin infraestructura<br>✅ Migración simple entre entornos |
| **Message Broker** | NATS JetStream vs Kafka/RabbitMQ | ✅ Menor complejidad operacional<br>✅ At-least-once delivery<br>✅ Pull-based subscription<br>❌ Kafka: overhead excesivo para prototipo |
| **Resolución de Conflictos** | Last-Write-Wins (LWW) | ✅ Simple y predecible<br>✅ Timestamp-based<br>⚠️ Trade-off: puede perder actualizaciones concurrentes (aceptable para inventario) |

### Alternativas Consideradas y Descartadas

**❌ Sincronización Síncrona (HTTP API Central)**
- Problema: Single point of failure, alta latencia, no funciona offline
- Razón de descarte: Contradice requisito de tolerancia a fallos

**❌ Distributed Locks (Redis, etcd)**
- Problema: Contención alta, complejidad de timeout, deadlocks
- Razón de descarte: Afecta disponibilidad negativamente

**❌ Consistency fuerte (2PC, Raft)**
- Problema: Latencia alta (>500ms), complejidad excesiva
- Razón de descarte: Overkill para inventario (eventual consistency es suficiente)

## ✅ Respuestas a Preguntas Clave

### 1. **GetAllStores en StockRepository**
**✅ RESUELTO**: Se agregó el método `GetAllStores(ctx, productID) ([]*Stock, error)` en la Fase 4 del StockRepository. Este método obtiene el stock de todas las tiendas para un producto específico y es usado por el QueryHandler para el endpoint de availability.

### 2. **ReservationRepository en QueryHandler**
**✅ RESUELTO**: Se actualizó el constructor de QueryHandler para incluir `reservationRepo` como dependencia. Ahora el handler puede acceder correctamente a `GetByID()` para el endpoint `GET /api/v1/reservations/:id`.

### 3. **Worker de Limpieza**
**✅ DECISIÓN**: El worker de limpieza se implementa como **goroutine dentro del mismo proceso de la API**, no como servicio separado. 

**Razones**:
- ✅ Más simple de desplegar (un solo binario)
- ✅ Comparte la misma conexión a DB y configuración
- ✅ Menos overhead de recursos
- ✅ Graceful shutdown automático
- ✅ Suficiente para el alcance de este prototipo

El worker ejecuta cada 1 minuto, busca reservas expiradas y las cancela automáticamente.

### 4. **Variables de Entorno (.env.example)**
**✅ AGREGADO**: Se creó el archivo `.env.example` en la Fase 12.3 con todas las variables de configuración:
- Server (puerto, store ID)
- PostgreSQL (host, puerto, usuario, password, DB)
- Redis (host, puerto)
- NATS (URL)
- Business (TTL de reservas)
- Logging (nivel, formato)

La aplicación usa `godotenv` para cargar automáticamente el archivo `.env`.

### 5. **Logging Estructurado**
**✅ RECOMENDACIÓN**: Para el prototipo inicial, `log.Printf` es **aceptable**, pero se agregó una sección completa sobre **zerolog** como recomendación para producción.

**Razones para zerolog**:
- ✅ Performance superior (zero allocation)
- ✅ JSON estructurado nativo
- ✅ Niveles de log configurables
- ✅ Fácil integración con sistemas de agregación de logs

**Implementación**:
- **Fase inicial**: Usar `log.Printf` (más simple)
- **Producción**: Migrar a `zerolog` (código de ejemplo incluido en el plan)
- El middleware de logger puede ser actualizado fácilmente

## Información del Proyecto

**Nombre**: Sistema de Gestión de Inventario Distribuido  
**Lenguaje**: Go 1.21+ (elegido por: concurrencia nativa con goroutines, performance, deployment simple)  
**Framework Web**: Gin (gin-gonic/gin) - ligero, rápido, middleware ecosystem  
**Arquitectura**: Event-Driven con CQRS simplificado  
**Base de datos**: 
- **Producción**: PostgreSQL (ACID completo, JSON support, índices avanzados)
- **Desarrollo/Testing**: SQLite (in-memory, cero configuración, portabilidad)
**Cache**: Redis (stock availability, sessions)  
**Message Broker**: NATS JetStream (eventos, sincronización entre tiendas)  
**Observabilidad**: zerolog (structured logging), Prometheus (métricas)
**Seguridad**: JWT authentication, bcrypt, rate limiting

### Stack Tecnológico Completo

```
┌─────────────────────────────────────────────────┐
│              API Gateway (Gin)                  │
│  - REST endpoints                               │
│  - JWT Auth middleware                          │
│  - Rate limiting                                │
│  - Request validation                           │
└─────────────────────────────────────────────────┘
                      │
        ┌─────────────┼─────────────┐
        ▼             ▼             ▼
┌──────────────┐ ┌──────────┐ ┌─────────────┐
│   Services   │ │  Redis   │ │    NATS     │
│   Layer      │ │  Cache   │ │ JetStream   │
└──────────────┘ └──────────┘ └─────────────┘
        │
        ▼
┌──────────────────────────────────────────┐
│      Repository Layer                    │
│  - Optimistic Locking                    │
│  - Transaction management                │
└──────────────────────────────────────────┘
        │
        ▼
┌──────────────────────────────────────────┐
│   PostgreSQL / SQLite                    │
│  - Products, Stock, Reservations         │
│  - Events (Event Sourcing)               │
└──────────────────────────────────────────┘
```

---

## Información del Proyecto

## Estructura del Proyecto

```
inventory-system/
├── cmd/
│   └── api/
│       └── main.go                    # Entry point del API Gateway (incluye worker de cleanup)
├── internal/
│   ├── config/
│   │   └── config.go                  # Configuración de la aplicación
│   ├── domain/
│   │   ├── product.go                 # Modelo de producto
│   │   ├── stock.go                   # Modelo de stock
│   │   ├── reservation.go             # Modelo de reserva
│   │   └── event.go                   # Modelo de evento
│   ├── repository/
│   │   ├── product_repository.go      # Repositorio de productos
│   │   ├── stock_repository.go        # Repositorio de stock (con optimistic locking)
│   │   ├── reservation_repository.go  # Repositorio de reservas
│   │   └── event_repository.go        # Repositorio de eventos
│   ├── service/
│   │   ├── stock_service.go           # Lógica de negocio para stock
│   │   ├── reservation_service.go     # Lógica de negocio para reservas
│   │   ├── event_publisher.go         # Publicador de eventos
│   │   └── sync_service.go            # Servicio de sincronización
│   ├── handler/
│   │   ├── stock_handler.go           # HTTP handlers para stock
│   │   ├── reservation_handler.go     # HTTP handlers para reservas
│   │   ├── query_handler.go           # HTTP handlers para consultas
│   │   └── sync_handler.go            # HTTP handlers para sincronización
│   ├── eventbus/
│   │   └── nats_client.go             # Cliente NATS
│   ├── middleware/
│   │   ├── auth.go                    # Middleware de autenticación
│   │   ├── logger.go                  # Middleware de logging
│   │   └── error_handler.go           # Middleware de manejo de errores
│   └── database/
│       ├── postgres.go                # Cliente PostgreSQL
│       └── redis.go                   # Cliente Redis
├── migrations/
│   └── 001_initial_schema.sql         # Schema inicial de la base de datos
├── tests/
│   ├── integration/
│   │   ├── stock_test.go
│   │   └── reservation_test.go
│   └── unit/
│       ├── stock_service_test.go
│       └── reservation_service_test.go
├── docs/
│   ├── API.md                         # Documentación de la API
│   ├── ARCHITECTURE.md                # Documentación de arquitectura
│   └── run.md                         # Instrucciones de ejecución
├── docker-compose.yml                 # Configuración de Docker
├── Dockerfile                         # Dockerfile para la aplicación
├── Makefile                          # Comandos útiles
├── go.mod
├── go.sum
└── README.md
```

---

## Fase 1: Setup Inicial del Proyecto

### Tarea 1.1: Inicializar Proyecto Go

**Archivo**: Crear estructura base del proyecto

**Comandos**:
```bash
mkdir -p inventory-system
cd inventory-system
go mod init inventory-system
```

**Dependencias a instalar**:
```bash
# Core dependencies
go get github.com/gin-gonic/gin
go get github.com/gin-contrib/cors

# Database drivers (ambos para flexibilidad)
go get github.com/lib/pq                    # PostgreSQL
go get github.com/mattn/go-sqlite3          # SQLite

# Cache y Message Broker
go get github.com/redis/go-redis/v9
go get github.com/nats-io/nats.go

# Utilities
go get github.com/google/uuid               # ✅ INSTALADO - UUID v4
go get github.com/joho/godotenv

# Security
go get github.com/golang-jwt/jwt/v5         # ✅ INSTALADO - JWT authentication
go get golang.org/x/crypto/bcrypt           # ✅ INSTALADO - Password hashing

# Observability
go get github.com/rs/zerolog
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp

# Rate limiting
go get golang.org/x/time/rate

# Testing
go get github.com/stretchr/testify
```

**Estado**: ✅ COMPLETADO

### Tarea 1.2: Crear archivo de configuración

**Archivo**: `internal/config/config.go`

**Contenido**:
```go
package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	ServerPort string
	StoreID    string
	
	// Database
	DatabaseDriver   string // "postgres" o "sqlite"
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	SQLitePath       string // Para SQLite: ":memory:" o ruta a archivo
	
	// Redis
	RedisHost string
	RedisPort int
	
	// NATS
	NATSUrl string
	
	// Business
	ReservationTTL int // segundos
	
	// Security
	JWTSecret         string
	RateLimitRequests int // requests per minute
	
	// Observability
	LogLevel      string // debug, info, warn, error
	LogFormat     string // json, text
	EnableMetrics bool
}

func Load() *Config {
	postgresPort, _ := strconv.Atoi(getEnv("POSTGRES_PORT", "5432"))
	redisPort, _ := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	reservationTTL, _ := strconv.Atoi(getEnv("RESERVATION_TTL", "600"))
	rateLimitRequests, _ := strconv.Atoi(getEnv("RATE_LIMIT_REQUESTS", "100"))
	enableMetrics, _ := strconv.ParseBool(getEnv("ENABLE_METRICS", "true"))
	
	return &Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		StoreID:           getEnv("STORE_ID", "store-001"),
		DatabaseDriver:    getEnv("DATABASE_DRIVER", "postgres"), // "postgres" o "sqlite"
		PostgresHost:      getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:      postgresPort,
		PostgresUser:      getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword:  getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:        getEnv("POSTGRES_DB", "inventory"),
		SQLitePath:        getEnv("SQLITE_PATH", ":memory:"),
		RedisHost:         getEnv("REDIS_HOST", "localhost"),
		RedisPort:         redisPort,
		NATSUrl:           getEnv("NATS_URL", "nats://localhost:4222"),
		ReservationTTL:    reservationTTL,
		JWTSecret:         getEnv("JWT_SECRET", "change-me-in-production"),
		RateLimitRequests: rateLimitRequests,
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		LogFormat:         getEnv("LOG_FORMAT", "json"),
		EnableMetrics:     enableMetrics,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

### Tarea 1.3: Crear Docker Compose

**Archivo**: `docker-compose.yml`

**Contenido**:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: inventory
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  nats:
    image: nats:latest
    ports:
      - "4222:4222"
      - "8222:8222"
    command: 
      - "-js"
      - "-m"
      - "8222"
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8222/healthz"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

---

## Fase 2: Capa de Dominio (Domain Models)

### Tarea 2.1: Crear modelo Product

**Archivo**: `internal/domain/product.go`

**Contenido**:
```go
package domain

import "time"

type Product struct {
	ID          string    `json:"id" db:"id"`
	SKU         string    `json:"sku" db:"sku"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Category    string    `json:"category" db:"category"`
	Price       float64   `json:"price" db:"price"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}
```

### Tarea 2.2: Crear modelo Stock

**Archivo**: `internal/domain/stock.go`

**Contenido**:
```go
package domain

import "time"

type Stock struct {
	ID        string    `json:"id" db:"id"`
	ProductID string    `json:"productId" db:"product_id"`
	StoreID   string    `json:"storeId" db:"store_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Reserved  int       `json:"reserved" db:"reserved"`
	Version   int       `json:"version" db:"version"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Available calcula el stock disponible
func (s *Stock) Available() int {
	return s.Quantity - s.Reserved
}

// CanReserve verifica si hay suficiente stock para reservar
func (s *Stock) CanReserve(quantity int) bool {
	return s.Available() >= quantity
}
```

### Tarea 2.3: Crear modelo Reservation

**Archivo**: `internal/domain/reservation.go`

**Contenido**:
```go
package domain

import "time"

type ReservationStatus string

const (
	ReservationStatusPending   ReservationStatus = "PENDING"
	ReservationStatusConfirmed ReservationStatus = "CONFIRMED"
	ReservationStatusCancelled ReservationStatus = "CANCELLED"
	ReservationStatusExpired   ReservationStatus = "EXPIRED"
)

type Reservation struct {
	ID         string            `json:"id" db:"id"`
	ProductID  string            `json:"productId" db:"product_id"`
	StoreID    string            `json:"storeId" db:"store_id"`
	Quantity   int               `json:"quantity" db:"quantity"`
	CustomerID string            `json:"customerId" db:"customer_id"`
	Status     ReservationStatus `json:"status" db:"status"`
	ExpiresAt  time.Time         `json:"expiresAt" db:"expires_at"`
	CreatedAt  time.Time         `json:"createdAt" db:"created_at"`
}

// IsExpired verifica si la reserva ha expirado
func (r *Reservation) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// CanConfirm verifica si la reserva puede ser confirmada
func (r *Reservation) CanConfirm() bool {
	return r.Status == ReservationStatusPending && !r.IsExpired()
}

// CanCancel verifica si la reserva puede ser cancelada
func (r *Reservation) CanCancel() bool {
	return r.Status == ReservationStatusPending
}
```

### Tarea 2.4: Crear modelo Event

**Archivo**: `internal/domain/event.go`

**Contenido**:
```go
package domain

import "time"

type EventType string

const (
	EventTypeStockUpdated        EventType = "stock.updated"
	EventTypeStockReserved       EventType = "stock.reserved"
	EventTypeReservationConfirmed EventType = "reservation.confirmed"
	EventTypeReservationCancelled EventType = "reservation.cancelled"
	EventTypeReservationExpired   EventType = "reservation.expired"
)

type Event struct {
	ID          string                 `json:"id" db:"id"`
	Type        EventType              `json:"type" db:"event_type"`
	AggregateID string                 `json:"aggregateId" db:"aggregate_id"`
	StoreID     string                 `json:"storeId" db:"store_id"`
	Payload     map[string]interface{} `json:"payload" db:"payload"`
	Timestamp   time.Time              `json:"timestamp" db:"timestamp"`
	Version     int                    `json:"version" db:"version"`
	Synced      bool                   `json:"synced" db:"synced"`
}
```

---

## Fase 3: Capa de Base de Datos

### Tarea 3.1: Crear schema SQL

**Archivo**: `migrations/001_initial_schema.sql`

**Contenido**:
```sql
-- Products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Stock table
CREATE TABLE IF NOT EXISTS stock (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    store_id VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    reserved INTEGER NOT NULL DEFAULT 0,
    version INTEGER NOT NULL DEFAULT 1,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, store_id),
    CHECK (quantity >= 0),
    CHECK (reserved >= 0),
    CHECK (reserved <= quantity)
);

-- Reservations table
CREATE TABLE IF NOT EXISTS reservations (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    store_id VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL,
    customer_id VARCHAR(100),
    status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'CONFIRMED', 'CANCELLED', 'EXPIRED')),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Events table
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    aggregate_id UUID NOT NULL,
    store_id VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version INTEGER NOT NULL,
    synced BOOLEAN DEFAULT FALSE
);

-- Stock history table (para auditoría)
CREATE TABLE IF NOT EXISTS stock_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    store_id VARCHAR(50) NOT NULL,
    quantity_change INTEGER NOT NULL,
    reason VARCHAR(100),
    user_id VARCHAR(100),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes para optimizar consultas
CREATE INDEX idx_stock_product_store ON stock(product_id, store_id);
CREATE INDEX idx_reservations_product_store ON reservations(product_id, store_id);
CREATE INDEX idx_reservations_status ON reservations(status);
CREATE INDEX idx_reservations_expires_at ON reservations(expires_at);
CREATE INDEX idx_events_aggregate ON events(aggregate_id);
CREATE INDEX idx_events_store ON events(store_id);
CREATE INDEX idx_events_synced ON events(synced);
CREATE INDEX idx_events_timestamp ON events(timestamp);

-- Function para actualizar updated_at automáticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_stock_updated_at BEFORE UPDATE ON stock
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert sample data
INSERT INTO products (id, sku, name, description, category, price) VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'PROD-001', 'Laptop HP Pavilion 15', '15.6" FHD, Intel i5, 8GB RAM', 'electronics', 599.99),
    ('550e8400-e29b-41d4-a716-446655440001', 'PROD-002', 'Mouse Logitech M185', 'Wireless Mouse', 'accessories', 29.99),
    ('550e8400-e29b-41d4-a716-446655440002', 'PROD-003', 'Teclado Mecánico RGB', 'Switches Cherry MX Blue', 'accessories', 89.99);

INSERT INTO stock (product_id, store_id, quantity, reserved) VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'store-001', 10, 0),
    ('550e8400-e29b-41d4-a716-446655440000', 'store-002', 5, 0),
    ('550e8400-e29b-41d4-a716-446655440001', 'store-001', 50, 0),
    ('550e8400-e29b-41d4-a716-446655440002', 'store-001', 20, 0);
```

### Tarea 3.2: Crear cliente de Base de Datos (PostgreSQL/SQLite)

**Archivo**: `internal/database/database.go`

**Contenido**:
```go
package database

import (
	"database/sql"
	"fmt"
	"inventory-system/internal/config"
	
	_ "github.com/lib/pq"              // PostgreSQL driver
	_ "github.com/mattn/go-sqlite3"    // SQLite driver
)

// NewDatabaseClient crea un cliente de base de datos basado en la configuración
func NewDatabaseClient(cfg *config.Config) (*sql.DB, error) {
	switch cfg.DatabaseDriver {
	case "sqlite":
		return newSQLiteClient(cfg)
	case "postgres":
		return newPostgresClient(cfg)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.DatabaseDriver)
	}
}

func newPostgresClient(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresDB,
	)
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres database: %w", err)
	}
	
	// Configurar pool de conexiones
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	
	// Verificar conexión
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres database: %w", err)
	}
	
	return db, nil
}

func newSQLiteClient(cfg *config.Config) (*sql.DB, error) {
	// SQLite path puede ser ":memory:" para in-memory o ruta a archivo
	db, err := sql.Open("sqlite3", cfg.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}
	
	// SQLite optimizations
	db.SetMaxOpenConns(1) // SQLite solo soporta 1 escritor
	
	// Verificar conexión
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}
	
	// Ejecutar pragmas para mejor performance
	pragmas := []string{
		"PRAGMA journal_mode=WAL",           // Write-Ahead Logging
		"PRAGMA synchronous=NORMAL",         // Balance entre safety y performance
		"PRAGMA cache_size=-64000",          // Cache de 64MB
		"PRAGMA foreign_keys=ON",            // Habilitar FKs
		"PRAGMA busy_timeout=5000",          // Timeout de 5s para locks
	}
	
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}
	
	return db, nil
}
```

**Nota**: Este enfoque permite alternar entre PostgreSQL (producción) y SQLite (desarrollo/testing) simplemente cambiando `DATABASE_DRIVER` en `.env`.

### Tarea 3.3: Crear cliente Redis

**Archivo**: `internal/database/redis.go`

**Contenido**:
```go
package database

import (
	"context"
	"fmt"
	"inventory-system/internal/config"
	
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
	})
	
	// Verificar conexión
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}
	
	return client, nil
}
```

---

## Fase 4: Capa de Repositorio

### Tarea 4.1: Crear StockRepository

**Archivo**: `internal/repository/stock_repository.go`

**Contenido**: Implementar todas las operaciones de stock con optimistic locking:
- `GetStock(ctx, productID, storeID) (*Stock, error)`
- `GetAllStores(ctx, productID) ([]*Stock, error)` - Obtener stock de todas las tiendas para un producto
- `UpdateStock(ctx, stock *Stock) error` - Con verificación de version
- `ReserveStock(ctx, productID, storeID, quantity) error` - Con SELECT FOR UPDATE
- `ReleaseReservation(ctx, productID, storeID, quantity) error`
- `ConfirmReservation(ctx, productID, storeID, quantity) error`
- Usar `sync.Map` para cache en memoria

**Ejemplo de GetAllStores**:
```go
func (r *StockRepository) GetAllStores(ctx context.Context, productID string) ([]*domain.Stock, error) {
	query := `
		SELECT id, product_id, store_id, quantity, reserved, version, updated_at
		FROM stock
		WHERE product_id = $1
		ORDER BY store_id
	`
	
	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var stocks []*domain.Stock
	for rows.Next() {
		var stock domain.Stock
		err := rows.Scan(
			&stock.ID,
			&stock.ProductID,
			&stock.StoreID,
			&stock.Quantity,
			&stock.Reserved,
			&stock.Version,
			&stock.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, &stock)
	}
	
	return stocks, rows.Err()
}
```

**Nota**: Este repositorio debe implementar el patrón de optimistic locking usando el campo `version`.

### Tarea 4.2: Crear ReservationRepository

**Archivo**: `internal/repository/reservation_repository.go`

**Contenido**: Implementar operaciones CRUD para reservas:
- `Create(ctx, reservation *Reservation) error`
- `GetByID(ctx, id string) (*Reservation, error)`
- `Update(ctx, reservation *Reservation) error`
- `Delete(ctx, id string) error`
- `GetPendingExpired(ctx) ([]*Reservation, error)` - Para cleanup automático

**Ejemplo de GetPendingExpired**:
```go
func (r *ReservationRepository) GetPendingExpired(ctx context.Context) ([]*domain.Reservation, error) {
	query := `
		SELECT id, product_id, store_id, quantity, customer_id, status, expires_at, created_at
		FROM reservations
		WHERE status = 'PENDING' AND expires_at < NOW()
		ORDER BY expires_at ASC
		LIMIT 100
	`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var reservations []*domain.Reservation
	for rows.Next() {
		var reservation domain.Reservation
		err := rows.Scan(
			&reservation.ID,
			&reservation.ProductID,
			&reservation.StoreID,
			&reservation.Quantity,
			&reservation.CustomerID,
			&reservation.Status,
			&reservation.ExpiresAt,
			&reservation.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, &reservation)
	}
	
	return reservations, rows.Err()
}
```

### Tarea 4.3: Crear EventRepository

**Archivo**: `internal/repository/event_repository.go`

**Contenido**: Implementar operaciones para eventos:
- `Save(ctx, event *Event) error`
- `GetPendingEvents(ctx, storeID, limit) ([]*Event, error)` - Eventos no sincronizados
- `MarkAsSynced(ctx, eventID) error`
- `GetEventsSince(ctx, lastEventID, limit) ([]*Event, error)`
- `GetLatestEvent(ctx, aggregateID) (*Event, error)`

### Tarea 4.4: Crear ProductRepository

**Archivo**: `internal/repository/product_repository.go`

**Contenido**: Implementar operaciones básicas para productos:
- `Create(ctx, product *Product) error`
- `GetByID(ctx, id string) (*Product, error)`
- `GetBySKU(ctx, sku string) (*Product, error)`
- `List(ctx, limit, offset) ([]*Product, error)`
- `Update(ctx, product *Product) error`

---

## Fase 5: Event Bus (NATS)

### Tarea 5.1: Crear cliente NATS

**Archivo**: `internal/eventbus/nats_client.go`

**Contenido**: Implementar cliente NATS con JetStream:
- `NewNATSClient(url) (*NATSClient, error)` - Conectar y crear streams
- `Publish(ctx, event *Event) error` - Publicar evento
- `Subscribe(subject, handler) error` - Subscribirse a eventos
- `SubscribeWithChannel(subject, eventChan) error` - Para usar con select
- `PublishBatch(events []*Event) error` - Publicar múltiples eventos
- `GetEventsSince(lastEventID, limit) ([]*Event, error)` - Pull de eventos
- `Close()` - Cerrar conexión

**Streams a crear**:
- `STOCK_EVENTS` - Subjects: `stock.*`, `reserv.*`
- `SYNC_EVENTS` - Subject: `sync.*`

---

## Fase 6: Capa de Servicio

### Tarea 6.1: Crear EventPublisher

**Archivo**: `internal/service/event_publisher.go`

**Contenido**: 
- Channel buffer de 1000 eventos
- Background worker que consume del channel y publica a NATS
- Método `Publish(event *Event)` - Non-blocking
- Método `Shutdown()` - Graceful shutdown

### Tarea 6.2: Crear StockService

**Archivo**: `internal/service/stock_service.go`

**Contenido**: Implementar lógica de negocio para stock:
- `UpdateStock(ctx, productID, storeID, quantity, reason) error`
- `ReserveStock(ctx, reservationID, productID, storeID, customerID, quantity, ttl) error`
- `ConfirmReservation(ctx, reservationID) error`
- `CancelReservation(ctx, reservationID) error`
- `scheduleExpiration(reservation *Reservation)` - Goroutine con timer

Cada método debe:
1. Realizar operación en repositorio
2. Crear evento
3. Publicar evento via EventPublisher

### Tarea 6.3: Crear SyncService

**Archivo**: `internal/service/sync_service.go`

**Contenido**: Implementar servicio de sincronización:
- Main loop con `select` multiplexing:
  - Channel de eventos remotos
  - Ticker para push (30s)
  - Ticker para pull (60s)
  - Ticker para heartbeat (30s)
  - Channel de shutdown
- `pushLocalEvents()` - Enviar eventos pendientes
- `pullRemoteEvents()` - Obtener eventos remotos
- `handleRemoteEvent(event)` - Procesar evento remoto
- `detectConflict(event) *Conflict` - Detectar conflictos
- `resolveConflict(conflict)` - Resolver usando Last-Write-Wins
- `sendHeartbeat()` - Enviar heartbeat a NATS

---

## Fase 7: Capa de Handlers (HTTP)

### Tarea 7.1: Crear StockHandler

**Archivo**: `internal/handler/stock_handler.go`

**Contenido**: Implementar endpoints HTTP con Gin:

```go
package handler

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
	"inventory-system/internal/service"
)

type StockHandler struct {
	stockService *service.StockService
}

func NewStockHandler(stockService *service.StockService) *StockHandler {
	return &StockHandler{
		stockService: stockService,
	}
}

type UpdateStockRequest struct {
	StoreID         string `json:"storeId" binding:"required"`
	Operation       string `json:"operation" binding:"required,oneof=add subtract set"`
	Quantity        int    `json:"quantity" binding:"required,min=0"`
	Reason          string `json:"reason" binding:"required"`
	ExpectedVersion int    `json:"expectedVersion"`
}

func (h *StockHandler) UpdateStock(c *gin.Context) {
	productID := c.Param("id")
	
	var req UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
			"details": err.Error(),
		})
		return
	}
	
	// Calcular nueva cantidad según operación
	var newQuantity int
	switch req.Operation {
	case "set":
		newQuantity = req.Quantity
	case "add", "subtract":
		// Obtener cantidad actual primero
		stock, err := h.stockService.GetStock(c.Request.Context(), productID, req.StoreID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Stock not found",
			})
			return
		}
		
		if req.Operation == "add" {
			newQuantity = stock.Quantity + req.Quantity
		} else {
			newQuantity = stock.Quantity - req.Quantity
		}
	}
	
	// Actualizar stock
	err := h.stockService.UpdateStock(
		c.Request.Context(),
		productID,
		req.StoreID,
		newQuantity,
		req.Reason,
	)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Stock update request accepted",
		"productId": productID,
		"storeId": req.StoreID,
	})
}
```

### Tarea 7.2: Crear ReservationHandler

**Archivo**: `internal/handler/reservation_handler.go`

**Contenido**: Implementar endpoints HTTP con Gin:

```go
package handler

import (
	"net/http"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"inventory-system/internal/service"
)

type ReservationHandler struct {
	stockService *service.StockService
}

func NewReservationHandler(stockService *service.StockService) *ReservationHandler {
	return &ReservationHandler{
		stockService: stockService,
	}
}

type CreateReservationRequest struct {
	ReservationID string `json:"reservationId"` // Cliente puede proporcionar ID (idempotencia)
	ProductID     string `json:"productId" binding:"required"`
	StoreID       string `json:"storeId" binding:"required"`
	Quantity      int    `json:"quantity" binding:"required,min=1"`
	CustomerID    string `json:"customerId"`
	TTL           int    `json:"ttl"` // Segundos, default 600 (10 min)
}

func (h *ReservationHandler) CreateReservation(c *gin.Context) {
	var req CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
			"details": err.Error(),
		})
		return
	}
	
	// Generar ID si no se proporciona
	if req.ReservationID == "" {
		req.ReservationID = uuid.New().String()
	}
	
	// TTL por defecto
	if req.TTL == 0 {
		req.TTL = 600 // 10 minutos
	}
	
	// Crear reserva
	err := h.stockService.ReserveStock(
		c.Request.Context(),
		req.ReservationID,
		req.ProductID,
		req.StoreID,
		req.CustomerID,
		req.Quantity,
		time.Duration(req.TTL)*time.Second,
	)
	
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusAccepted, gin.H{
		"reservationId": req.ReservationID,
		"status": "pending",
		"expiresAt": time.Now().Add(time.Duration(req.TTL) * time.Second),
		"message": "Reservation accepted",
	})
}

type ConfirmReservationRequest struct {
	OrderID string `json:"orderId"`
}

func (h *ReservationHandler) ConfirmReservation(c *gin.Context) {
	reservationID := c.Param("id")
	
	var req ConfirmReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
		})
		return
	}
	
	err := h.stockService.ConfirmReservation(c.Request.Context(), reservationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"reservationId": reservationID,
		"status": "confirmed",
		"confirmedAt": time.Now(),
		"orderId": req.OrderID,
	})
}

func (h *ReservationHandler) CancelReservation(c *gin.Context) {
	reservationID := c.Param("id")
	
	err := h.stockService.CancelReservation(c.Request.Context(), reservationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.Status(http.StatusNoContent)
}
```

### Tarea 7.3: Crear QueryHandler

**Archivo**: `internal/handler/query_handler.go`

**Contenido**: Implementar endpoints de consulta con Gin:

```go
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"inventory-system/internal/domain"
	"inventory-system/internal/repository"
)

type QueryHandler struct {
	productRepo     *repository.ProductRepository
	stockRepo       *repository.StockRepository
	reservationRepo *repository.ReservationRepository
	redis           *redis.Client
}

func NewQueryHandler(
	productRepo *repository.ProductRepository,
	stockRepo *repository.StockRepository,
	reservationRepo *repository.ReservationRepository,
	redis *redis.Client,
) *QueryHandler {
	return &QueryHandler{
		productRepo:     productRepo,
		stockRepo:       stockRepo,
		reservationRepo: reservationRepo,
		redis:           redis,
	}
}

func (h *QueryHandler) ListProducts(c *gin.Context) {
	// Paginación
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if limit > 100 {
		limit = 100
	}
	
	offset := (page - 1) * limit
	
	products, err := h.productRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch products",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": products,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *QueryHandler) GetProduct(c *gin.Context) {
	productID := c.Param("id")
	
	product, err := h.productRepo.GetByID(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
		})
		return
	}
	
	c.JSON(http.StatusOK, product)
}

func (h *QueryHandler) GetAvailability(c *gin.Context) {
	productID := c.Param("id")
	storeID := c.Query("storeId")
	
	// Intentar desde cache
	cacheKey := fmt.Sprintf("availability:%s:%s", productID, storeID)
	
	cached, err := h.redis.Get(c.Request.Context(), cacheKey).Result()
	if err == nil {
		// Cache hit
		var result map[string]interface{}
		json.Unmarshal([]byte(cached), &result)
		result["cacheHit"] = true
		c.JSON(http.StatusOK, result)
		return
	}
	
	// Cache miss - consultar DB
	var stocks []*domain.Stock
	if storeID != "" {
		stock, err := h.stockRepo.GetStock(c.Request.Context(), productID, storeID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Stock not found",
			})
			return
		}
		stocks = []*domain.Stock{stock}
	} else {
		// Obtener de todas las tiendas
		stocks, err = h.stockRepo.GetAllStores(c.Request.Context(), productID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch stock",
			})
			return
		}
	}
	
	// Construir respuesta
	storeData := make([]gin.H, 0, len(stocks))
	totalAvailable := 0
	
	for _, stock := range stocks {
		available := stock.Available()
		totalAvailable += available
		
		storeData = append(storeData, gin.H{
			"storeId":     stock.StoreID,
			"quantity":    stock.Quantity,
			"reserved":    stock.Reserved,
			"available":   available,
			"lastUpdated": stock.UpdatedAt,
		})
	}
	
	result := gin.H{
		"productId":      productID,
		"stores":         storeData,
		"totalAvailable": totalAvailable,
		"cacheHit":       false,
		"cachedAt":       time.Now(),
	}
	
	// Guardar en cache (TTL: 30 segundos)
	resultJSON, _ := json.Marshal(result)
	h.redis.Set(c.Request.Context(), cacheKey, resultJSON, 30*time.Second)
	
	c.JSON(http.StatusOK, result)
}

func (h *QueryHandler) GetReservation(c *gin.Context) {
	reservationID := c.Param("id")
	
	reservation, err := h.reservationRepo.GetByID(c.Request.Context(), reservationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Reservation not found",
		})
		return
	}
	
	remainingSeconds := int(time.Until(reservation.ExpiresAt).Seconds())
	if remainingSeconds < 0 {
		remainingSeconds = 0
	}
	
	c.JSON(http.StatusOK, gin.H{
		"reservationId":    reservation.ID,
		"productId":        reservation.ProductID,
		"storeId":          reservation.StoreID,
		"quantity":         reservation.Quantity,
		"status":           reservation.Status,
		"createdAt":        reservation.CreatedAt,
		"expiresAt":        reservation.ExpiresAt,
		"remainingSeconds": remainingSeconds,
	})
}
```

### Tarea 7.4: Crear SyncHandler

**Archivo**: `internal/handler/sync_handler.go`

**Contenido**: Implementar endpoints de sincronización:
- `POST /api/v1/sync/pull` - Tienda solicita eventos
- `POST /api/v1/sync/push` - Tienda envía eventos
- `GET /api/v1/sync/status` - Estado de sincronización

---

## Fase 8: Middleware y Observabilidad

### Tarea 8.1: Crear Logger Middleware con Zerolog

**Archivo**: `internal/middleware/logger.go`

**Contenido**: 
```go
package middleware

import (
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Logger(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		// Procesar request
		c.Next()
		
		// Calcular latencia
		latency := time.Since(start)
		
		// Obtener información
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()
		bodySize := c.Writer.Size()
		
		if raw != "" {
			path = path + "?" + raw
		}
		
		// Log estructurado con zerolog
		event := logger.Info()
		if statusCode >= 500 {
			event = logger.Error()
		} else if statusCode >= 400 {
			event = logger.Warn()
		}
		
		event.
			Str("method", method).
			Str("path", path).
			Str("ip", clientIP).
			Int("status", statusCode).
			Dur("latency", latency).
			Int("body_size", bodySize).
			Str("user_agent", c.Request.UserAgent()).
			Str("error", errorMessage).
			Msg("HTTP request")
	}
}
```

### Tarea 8.2: Crear Metrics Middleware con Prometheus

**Archivo**: `internal/middleware/metrics.go`

**Contenido**:
```go
package middleware

import (
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latencies in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	
	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request sizes in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)
	
	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response sizes in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)
)

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Procesar request
		c.Next()
		
		// Registrar métricas
		duration := time.Since(start).Seconds()
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}
		
		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			endpoint,
			strconv.Itoa(c.Writer.Status()),
		).Inc()
		
		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			endpoint,
		).Observe(duration)
		
		httpRequestSize.WithLabelValues(
			c.Request.Method,
			endpoint,
		).Observe(float64(c.Request.ContentLength))
		
		httpResponseSize.WithLabelValues(
			c.Request.Method,
			endpoint,
		).Observe(float64(c.Writer.Size()))
	}
}
```

### Tarea 8.3: Crear Error Handler Middleware

**Archivo**: `internal/middleware/error_handler.go`

**Contenido**:
```go
package middleware

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
)

// ErrorHandler maneja errores de forma centralizada
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Verificar si hay errores
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			// Determinar código de estado
			statusCode := http.StatusInternalServerError
			
			switch err.Type {
			case gin.ErrorTypeBind:
				statusCode = http.StatusBadRequest
			case gin.ErrorTypePublic:
				statusCode = http.StatusBadRequest
			}
			
			// Responder con error
			c.JSON(statusCode, gin.H{
				"error": err.Error(),
			})
			return
		}
	}
}

// Custom error types
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	return "Resource not found: " + e.Resource
}

type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}
```

### Tarea 8.3: Crear Auth Middleware (Simple)

**Archivo**: `internal/middleware/auth.go`

**Contenido**:
```go
package middleware

import (
	"net/http"
	"strings"
	
	"github.com/gin-gonic/gin"
)

// Auth middleware simple para prototipo
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener header Authorization
		authHeader := c.GetHeader("Authorization")
		
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}
		
		// Verificar formato Bearer
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}
		
		token := parts[1]
		
		// Para prototipo: Solo verificar que el token existe
		// En producción: Validar JWT, verificar permisos, etc.
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}
		
		// Extraer storeID del token (simulado)
		// En producción: Decodificar JWT y extraer claims
		storeID := extractStoreIDFromToken(token)
		
		// Guardar en contexto
		c.Set("storeID", storeID)
		c.Set("token", token)
		
		c.Next()
	}
}

// extractStoreIDFromToken simula la extracción del storeID
// En producción: Decodificar JWT y extraer del claim
func extractStoreIDFromToken(token string) string {
	// Simulación simple
	// En producción usar: github.com/golang-jwt/jwt/v5
	return "store-001"
}
```

### Tarea 8.4: Crear Auth Middleware Mejorado con JWT

**Archivo**: `internal/middleware/auth_jwt.go`

**Contenido**:
```go
package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims estructura para el JWT
type Claims struct {
	StoreID  string `json:"store_id"`
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthJWT middleware con validación JWT completa
func AuthJWT(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}
		
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}
		
		tokenString := parts[1]
		
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}
		
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token expired or invalid",
			})
			c.Abort()
			return
		}
		
		c.Set("store_id", claims.StoreID)
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		
		c.Next()
	}
}

// GenerateToken genera un nuevo JWT (útil para testing)
func GenerateToken(storeID, userID, role, secret string) (string, error) {
	claims := Claims{
		StoreID: storeID,
		UserID:  userID,
		Role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
```

### Tarea 8.5: Crear Rate Limiting Middleware

**Archivo**: `internal/middleware/ratelimit.go`

**Contenido**:
```go
package middleware

import (
	"net/http"
	"sync"
	
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerMinute) / 60,
		burst:    requestsPerMinute / 10,
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}
	
	return limiter
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getLimiter(ip)
		
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}
```

---

## Fase 9: Main Application

### Tarea 9.1: Crear main de API

**Archivo**: `cmd/api/main.go`

**Contenido**:
```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"inventory-system/internal/config"
	"inventory-system/internal/database"
	"inventory-system/internal/eventbus"
	"inventory-system/internal/handler"
	"inventory-system/internal/middleware"
	"inventory-system/internal/repository"
	"inventory-system/internal/service"
)

func main() {
	// 1. Cargar configuración
	cfg := config.Load()
	
	// 2. Conectar a base de datos
	db, err := database.NewPostgresClient(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	redis, err := database.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer redis.Close()
	
	// 3. Conectar a NATS
	natsClient, err := eventbus.NewNATSClient(cfg.NATSUrl)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer natsClient.Close()
	
	// 4. Inicializar repositorios
	stockRepo := repository.NewStockRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	eventRepo := repository.NewEventRepository(db)
	productRepo := repository.NewProductRepository(db)
	
	// 5. Inicializar servicios
	eventPublisher := service.NewEventPublisher(natsClient)
	defer eventPublisher.Shutdown()
	
	stockService := service.NewStockService(stockRepo, reservationRepo, eventPublisher)
	syncService := service.NewSyncService(cfg.StoreID, eventRepo, natsClient, stockService)
	syncService.Start()
	defer syncService.Shutdown()
	
	// 6. Inicializar handlers
	stockHandler := handler.NewStockHandler(stockService)
	reservationHandler := handler.NewReservationHandler(stockService)
	queryHandler := handler.NewQueryHandler(productRepo, stockRepo, reservationRepo, redis)
	syncHandler := handler.NewSyncHandler(syncService)
	
	// 7. Configurar modo Gin (release para producción)
	gin.SetMode(gin.DebugMode) // Cambiar a gin.ReleaseMode en producción
	
	// 8. Crear router Gin
	router := gin.New()
	
	// 9. Middlewares globales
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(middleware.ErrorHandler())
	
	// 10. Rutas públicas
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})
	
	// 11. Rutas de API
	api := router.Group("/api/v1")
	api.Use(middleware.Auth()) // Requiere autenticación
	
	// Command endpoints
	api.POST("/products/:id/stock", stockHandler.UpdateStock)
	api.POST("/reservations", reservationHandler.CreateReservation)
	api.POST("/reservations/:id/confirm", reservationHandler.ConfirmReservation)
	api.DELETE("/reservations/:id", reservationHandler.CancelReservation)
	
	// Query endpoints
	api.GET("/products", queryHandler.ListProducts)
	api.GET("/products/:id", queryHandler.GetProduct)
	api.GET("/products/:id/availability", queryHandler.GetAvailability)
	api.GET("/reservations/:id", queryHandler.GetReservation)
	
	// Sync endpoints
	api.POST("/sync/pull", syncHandler.Pull)
	api.POST("/sync/push", syncHandler.Push)
	api.GET("/sync/status", syncHandler.GetStatus)
	
	// 12. Configurar servidor HTTP
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	// 13. Iniciar servidor en goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// 14. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server exited")
}
```

### Tarea 9.2: Crear worker de limpieza integrado

**Archivo**: Integrado en `cmd/api/main.go`

**Contenido**: Agregar goroutine de limpieza de reservas expiradas dentro del proceso principal

Agregar después de inicializar los servicios (después de la línea del syncService):

```go
	// 5.5. Iniciar worker de limpieza de reservas expiradas
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		
		log.Println("[Worker] Reservation cleanup worker started")
		
		for {
			select {
			case <-ticker.C:
				// Obtener reservas expiradas
				ctx := context.Background()
				expired, err := reservationRepo.GetPendingExpired(ctx)
				if err != nil {
					log.Printf("[Worker] Error fetching expired reservations: %v", err)
					continue
				}
				
				if len(expired) == 0 {
					continue
				}
				
				log.Printf("[Worker] Found %d expired reservations to clean up", len(expired))
				
				// Cancelar cada reserva expirada
				for _, reservation := range expired {
					err := stockService.CancelReservation(ctx, reservation.ID)
					if err != nil {
						log.Printf("[Worker] Error cancelling reservation %s: %v", reservation.ID, err)
						continue
					}
					
					// Actualizar estado a EXPIRED
					reservation.Status = domain.ReservationStatusExpired
					if err := reservationRepo.Update(ctx, reservation); err != nil {
						log.Printf("[Worker] Error updating reservation status: %v", err)
					}
				}
				
			case <-quit:
				log.Println("[Worker] Cleanup worker shutting down")
				return
			}
		}
	}()
```

**Nota**: El worker se ejecuta como goroutine en el mismo proceso que la API. Esto simplifica el despliegue y no requiere un proceso separado. El worker se detendrá automáticamente cuando se reciba la señal de shutdown.

---

## Fase 10: Testing

### Tarea 10.1: Unit Tests para StockService

**Archivo**: `tests/unit/stock_service_test.go`

**Tests a implementar**:
```go
package unit

import (
	"context"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"inventory-system/internal/domain"
	"inventory-system/internal/service"
)

// Mock del StockRepository
type MockStockRepository struct {
	mock.Mock
}

func (m *MockStockRepository) GetStock(ctx context.Context, productID, storeID string) (*domain.Stock, error) {
	args := m.Called(ctx, productID, storeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Stock), args.Error(1)
}

func (m *MockStockRepository) UpdateStock(ctx context.Context, stock *domain.Stock) error {
	args := m.Called(ctx, stock)
	return args.Error(0)
}

// Test: Actualizar stock exitosamente
func TestUpdateStock_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockStockRepository)
	mockEventPub := new(MockEventPublisher)
	service := service.NewStockService(mockRepo, nil, mockEventPub)
	
	stock := &domain.Stock{
		ID:        "stock-1",
		ProductID: "prod-1",
		StoreID:   "store-1",
		Quantity:  10,
		Version:   1,
	}
	
	mockRepo.On("GetStock", mock.Anything, "prod-1", "store-1").Return(stock, nil)
	mockRepo.On("UpdateStock", mock.Anything, mock.Anything).Return(nil)
	mockEventPub.On("Publish", mock.Anything).Return()
	
	// Execute
	err := service.UpdateStock(context.Background(), "prod-1", "store-1", 20, "restock")
	
	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockEventPub.AssertExpectations(t)
}

// Test: Fallo de optimistic lock
func TestUpdateStock_OptimisticLockFailure(t *testing.T) {
	mockRepo := new(MockStockRepository)
	mockEventPub := new(MockEventPublisher)
	service := service.NewStockService(mockRepo, nil, mockEventPub)
	
	stock := &domain.Stock{
		ID:        "stock-1",
		ProductID: "prod-1",
		StoreID:   "store-1",
		Quantity:  10,
		Version:   1,
	}
	
	mockRepo.On("GetStock", mock.Anything, "prod-1", "store-1").Return(stock, nil)
	mockRepo.On("UpdateStock", mock.Anything, mock.Anything).Return(
		errors.New("optimistic lock failed: stock was modified by another transaction"),
	)
	
	err := service.UpdateStock(context.Background(), "prod-1", "store-1", 20, "restock")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "optimistic lock")
}

// Test: Reservar stock exitosamente
func TestReserveStock_Success(t *testing.T) {
	mockStockRepo := new(MockStockRepository)
	mockReservRepo := new(MockReservationRepository)
	mockEventPub := new(MockEventPublisher)
	service := service.NewStockService(mockStockRepo, mockReservRepo, mockEventPub)
	
	mockStockRepo.On("ReserveStock", mock.Anything, "prod-1", "store-1", 2).Return(nil)
	mockReservRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockEventPub.On("Publish", mock.Anything).Return()
	
	err := service.ReserveStock(
		context.Background(),
		"rsv-1",
		"prod-1",
		"store-1",
		"cust-1",
		2,
		10*time.Minute,
	)
	
	assert.NoError(t, err)
}

// Test: Stock insuficiente
func TestReserveStock_InsufficientStock(t *testing.T) {
	mockStockRepo := new(MockStockRepository)
	mockReservRepo := new(MockReservationRepository)
	mockEventPub := new(MockEventPublisher)
	service := service.NewStockService(mockStockRepo, mockReservRepo, mockEventPub)
	
	mockStockRepo.On("ReserveStock", mock.Anything, "prod-1", "store-1", 100).Return(
		errors.New("insufficient stock: available=5, requested=100"),
	)
	
	err := service.ReserveStock(
		context.Background(),
		"rsv-1",
		"prod-1",
		"store-1",
		"cust-1",
		100,
		10*time.Minute,
	)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient stock")
}
```

### Tarea 10.2: Integration Tests

**Archivo**: `tests/integration/stock_test.go`

**Tests a implementar**:
```go
package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"inventory-system/internal/handler"
	"inventory-system/internal/service"
)

// Test: Flujo completo de reserva y confirmación
func TestCreateReservationFlow(t *testing.T) {
	// Setup
	router := setupTestRouter()
	
	// 1. Crear reserva
	reqBody := map[string]interface{}{
		"productId": "prod-1",
		"storeId":   "store-1",
		"quantity":  2,
		"customerId": "cust-1",
	}
	
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/reservations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusAccepted, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	reservationID := response["reservationId"].(string)
	
	// 2. Confirmar reserva
	confirmBody := map[string]interface{}{
		"orderId": "order-123",
	}
	
	body, _ = json.Marshal(confirmBody)
	req = httptest.NewRequest(
		"POST",
		"/api/v1/reservations/"+reservationID+"/confirm",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

// Test: Reserva se expira automáticamente
func TestReservationExpiration(t *testing.T) {
	router := setupTestRouter()
	
	// Crear reserva con TTL corto (2 segundos)
	reqBody := map[string]interface{}{
		"productId": "prod-1",
		"storeId":   "store-1",
		"quantity":  2,
		"ttl":       2,
	}
	
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/reservations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	reservationID := response["reservationId"].(string)
	
	// Esperar a que expire
	time.Sleep(3 * time.Second)
	
	// Intentar confirmar (debe fallar)
	req = httptest.NewRequest(
		"POST",
		"/api/v1/reservations/"+reservationID+"/confirm",
		nil,
	)
	req.Header.Set("Authorization", "Bearer test-token")
	
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test: Múltiples goroutines intentan reservar el mismo stock
func TestConcurrentReservations(t *testing.T) {
	router := setupTestRouter()
	
	// 10 goroutines intentan reservar 2 unidades cada una
	// Stock disponible: 5 unidades
	// Solo 2 reservas deben tener éxito
	
	successCount := 0
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			reqBody := map[string]interface{}{
				"productId": "prod-1",
				"storeId":   "store-1",
				"quantity":  2,
			}
			
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/api/v1/reservations", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code == http.StatusAccepted {
				successCount++
			}
			
			done <- true
		}()
	}
	
	// Esperar a que terminen todas
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Solo 2 deben haber tenido éxito (5 unidades / 2 por reserva)
	assert.LessOrEqual(t, successCount, 2)
}

func setupTestRouter() *gin.Engine {
	// Setup de test con base de datos de test
	// Usar testcontainers para PostgreSQL
	// ...
	return gin.Default()
}
```

Usar mocks para repositorios en unit tests, y base de datos real (Docker) para integration tests.

---

## Fase 11: Documentación

### Tarea 11.1: Crear README.md

**Archivo**: `README.md`

**Contenido**:
- Descripción del proyecto
- Arquitectura (diagrama Mermaid)
- Requisitos
- Quick start
- Variables de entorno

### Tarea 11.2: Crear run.md

**Archivo**: `docs/run.md`

**Contenido**:
```markdown
# Cómo Ejecutar el Sistema

## Prerrequisitos
- Go 1.21+
- Docker y Docker Compose
- Make (opcional, pero recomendado)

## Configuración Inicial

### 1. Clonar y preparar el proyecto
```bash
git clone <repository-url>
cd inventory-system
```

### 2. Configurar variables de entorno
```bash
# Copiar archivo de ejemplo
cp .env.example .env

# Editar según tu entorno (opcional, los defaults funcionan)
nano .env
```

### 3. Instalar dependencias
```bash
make deps
# O manualmente:
go mod download
go mod tidy
```

## Ejecución

### Método 1: Con Make (Recomendado)

```bash
# 1. Iniciar infraestructura (PostgreSQL, Redis, NATS)
make docker-up

# 2. Verificar que los servicios estén saludables
docker-compose ps
# Todos deben mostrar "healthy"

# 3. Compilar y ejecutar la aplicación
make run
```

### Método 2: Manual

```bash
# 1. Iniciar infraestructura
docker-compose up -d

# 2. Esperar a que estén listos (5-10 segundos)
sleep 10

# 3. Compilar
go build -o bin/api cmd/api/main.go

# 4. Ejecutar
./bin/api
```

## Verificación

### 1. Health Check
```bash
curl http://localhost:8080/health
```

Respuesta esperada:
```json
{"status":"healthy"}
```

### 2. Listar productos de ejemplo
```bash
curl http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer test-token"
```

### 3. Crear una reserva
```bash
curl -X POST http://localhost:8080/api/v1/reservations \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -d '{
    "productId": "550e8400-e29b-41d4-a716-446655440000",
    "storeId": "store-001",
    "quantity": 2
  }'
```

### 4. Verificar disponibilidad
```bash
curl http://localhost:8080/api/v1/products/550e8400-e29b-41d4-a716-446655440000/availability \
  -H "Authorization: Bearer test-token"
```

## Testing

### Ejecutar todos los tests
```bash
make test
```

### Tests con cobertura
```bash
make test-coverage
# Abre coverage.html en tu navegador
```

### Tests específicos
```bash
# Unit tests
go test -v ./tests/unit/...

# Integration tests
go test -v ./tests/integration/...
```

## Logs y Debugging

### Ver logs de todos los servicios
```bash
make docker-logs
```

### Ver logs solo de PostgreSQL
```bash
docker-compose logs -f postgres
```

### Ver logs solo de NATS
```bash
docker-compose logs -f nats
```

### Logs de la aplicación
Los logs se imprimen en stdout. Puedes redirigir a archivo:
```bash
./bin/api 2>&1 | tee app.log
```

## Detener el Sistema

### Detener la aplicación
```
Ctrl+C en la terminal donde corre la aplicación
```

### Detener infraestructura
```bash
make docker-down
# O manualmente:
docker-compose down
```

### Limpiar todo (incluyendo volúmenes)
```bash
docker-compose down -v
make clean
```

## Variables de Entorno

| Variable | Default | Descripción |
|----------|---------|-------------|
| `SERVER_PORT` | 8080 | Puerto del servidor HTTP |
| `STORE_ID` | store-001 | ID de la tienda (para multi-tienda) |
| `POSTGRES_HOST` | localhost | Host de PostgreSQL |
| `POSTGRES_PORT` | 5432 | Puerto de PostgreSQL |
| `POSTGRES_USER` | postgres | Usuario de PostgreSQL |
| `POSTGRES_PASSWORD` | postgres | Contraseña de PostgreSQL |
| `POSTGRES_DB` | inventory | Nombre de la base de datos |
| `REDIS_HOST` | localhost | Host de Redis |
| `REDIS_PORT` | 6379 | Puerto de Redis |
| `NATS_URL` | nats://localhost:4222 | URL de NATS |
| `RESERVATION_TTL` | 600 | TTL de reservas (segundos) |
| `LOG_LEVEL` | info | Nivel de log (debug/info/warn/error) |
| `LOG_FORMAT` | json | Formato de log (json/text) |

## Troubleshooting

### "Cannot connect to PostgreSQL"
```bash
# Verificar que Docker Compose esté corriendo
docker-compose ps

# Verificar logs de PostgreSQL
docker-compose logs postgres

# Reiniciar PostgreSQL
docker-compose restart postgres
```

### "NATS connection refused"
```bash
# NATS tarda ~5 segundos en iniciar
# Esperar y verificar logs
docker-compose logs nats

# Debe mostrar: "Server is ready"
```

### "Port 8080 already in use"
```bash
# Cambiar puerto en .env
echo "SERVER_PORT=8081" >> .env

# O matar proceso que usa el puerto
lsof -ti:8080 | xargs kill -9
```

### Tests fallan con race condition
```bash
# Ejecutar con detector de race conditions
go test -race ./tests/...
```

### Reset completo del sistema
```bash
# Detener todo
make docker-down

# Limpiar volúmenes
docker-compose down -v

# Limpiar builds
make clean

# Reiniciar
make docker-up
make run
```

## Desarrollo

### Watch mode (recompilación automática)
Instalar air:
```bash
go install github.com/cosmtrek/air@latest
```

Ejecutar con hot reload:
```bash
air
```

### Formato de código
```bash
go fmt ./...
```

### Linters
```bash
make lint
# O instalar golangci-lint:
# https://golangci-lint.run/usage/install/
```

## Notas Adicionales

- **Worker de limpieza**: Se ejecuta automáticamente como goroutine dentro del proceso principal (no requiere proceso separado)
- **Graceful shutdown**: La aplicación maneja señales SIGTERM/SIGINT correctamente
- **Sincronización**: El Sync Service se ejecuta automáticamente en background
- **Reservas**: Las reservas expiran automáticamente después del TTL configurado
```

### Tarea 11.3: Crear API.md

**Archivo**: `docs/API.md`

**Contenido**: Documentación completa de todos los endpoints

**Ejemplo de formato**:

```markdown
# API Documentation

## Authentication

Todos los endpoints (excepto `/health`) requieren autenticación mediante JWT:

```
Authorization: Bearer <token>
```

## Endpoints

### Health Check

**GET** `/health`

Verifica el estado del servidor.

**Response**:
```json
{
  "status": "healthy"
}
```

### Create Reservation

**POST** `/api/v1/reservations`

Crea una reserva temporal de stock.

**Request Body**:
```json
{
  "reservationId": "rsv-123",  // Opcional (cliente genera para idempotencia)
  "productId": "550e8400-e29b-41d4-a716-446655440000",
  "storeId": "store-001",
  "quantity": 2,
  "customerId": "cust-789",
  "ttl": 600  // Opcional (segundos, default: 600)
}
```

**Response** (202 Accepted):
```json
{
  "reservationId": "rsv-123",
  "status": "pending",
  "expiresAt": "2025-10-26T10:40:00Z",
  "message": "Reservation accepted"
}
```

**Error Response** (409 Conflict):
```json
{
  "error": "insufficient stock: available=1, requested=2"
}
```

**cURL Example**:
```bash
curl -X POST http://localhost:8080/api/v1/reservations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "productId": "550e8400-e29b-41d4-a716-446655440000",
    "storeId": "store-001",
    "quantity": 2
  }'
```

### Update Stock

**POST** `/api/v1/products/:id/stock`

Actualiza el stock de un producto.

**URL Parameters**:
- `id`: Product ID

**Request Body**:
```json
{
  "storeId": "store-001",
  "operation": "add",  // "add" | "subtract" | "set"
  "quantity": 20,
  "reason": "restock",
  "expectedVersion": 5  // Opcional (optimistic locking)
}
```

**Response** (202 Accepted):
```json
{
  "message": "Stock update request accepted",
  "productId": "550e8400-e29b-41d4-a716-446655440000",
  "storeId": "store-001"
}
```

... (continuar con todos los endpoints)
```

### Tarea 11.4: Crear ARCHITECTURE.md

**Archivo**: `docs/ARCHITECTURE.md`

**Contenido**:
- Decisiones arquitectónicas y justificación
- Diagrama de componentes (Mermaid)
- Diagrama de flujo de datos
- Explicación de Event-Driven Architecture
- Explicación de Optimistic Locking
- Explicación de Sync Service

---

## Fase 12: Makefile y Docker

### Tarea 12.1: Crear Makefile

**Archivo**: `Makefile`

**Contenido**:
```makefile
.PHONY: help build run test clean docker-up docker-down lint

help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linters"
	@echo "  docker-up   - Start Docker services"
	@echo "  docker-down - Stop Docker services"
	@echo "  clean       - Clean build artifacts"

build:
	@echo "Building API..."
	@go build -o bin/api cmd/api/main.go
	@echo "Build complete!"

run: build
	@echo "Starting API server..."
	@./bin/api

test:
	@echo "Running tests..."
	@go test -v -race ./tests/...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./tests/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	@echo "Running linters..."
	@golangci-lint run ./...

docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@docker-compose ps

docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down

docker-logs:
	@docker-compose logs -f

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

.DEFAULT_GOAL := help
```

### Tarea 12.2: Crear Dockerfile

**Archivo**: `Dockerfile`

**Contenido**:
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app/bin/api cmd/api/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bin/api .
EXPOSE 8080
CMD ["./api"]
```

### Tarea 12.3: Crear archivo .env.example

**Archivo**: `.env.example`

**Contenido**:
```bash
# Server Configuration
SERVER_PORT=8080
STORE_ID=store-001

# Database Configuration
DATABASE_DRIVER=postgres  # postgres | sqlite
SQLITE_PATH=:memory:      # Para SQLite: ":memory:" o ruta como "./data/inventory.db"

# PostgreSQL Configuration (solo si DATABASE_DRIVER=postgres)
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=inventory

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# NATS Configuration
NATS_URL=nats://localhost:4222

# Business Configuration
RESERVATION_TTL=600  # Tiempo de expiración de reservas en segundos (default: 10 minutos)

# Security Configuration
JWT_SECRET=change-me-in-production-use-strong-random-string
RATE_LIMIT_REQUESTS=100  # Requests por minuto por IP

# Observability Configuration
LOG_LEVEL=info      # debug | info | warn | error
LOG_FORMAT=json     # json | text
ENABLE_METRICS=true # true | false (Prometheus metrics)
```

**Instrucciones de uso**:
```bash
# Copiar el archivo de ejemplo
cp .env.example .env

# Editar con tus valores
nano .env

# La aplicación cargará automáticamente las variables usando godotenv
```

**Nota sobre DATABASE_DRIVER**:
- Para desarrollo/testing local sin infraestructura: `DATABASE_DRIVER=sqlite` y `SQLITE_PATH=:memory:`
- Para testing con persistencia: `DATABASE_DRIVER=sqlite` y `SQLITE_PATH=./data/test.db`
- Para producción: `DATABASE_DRIVER=postgres` con configuración completa de PostgreSQL

### Tarea 12.4: Actualizar config.go para cargar .env

Agregar al inicio de `internal/config/config.go`:

```go
package config

import (
	"os"
	"strconv"
	
	"github.com/joho/godotenv"
)

func Load() *Config {
	// Cargar .env si existe (ignora error si no existe)
	_ = godotenv.Load()
	
	// ... resto del código existente
}
```

---

## 📋 Orden de Implementación Incremental

Esta sección define el orden exacto de implementación para construir el sistema **fase por fase**, donde cada fase es ejecutable y validable independientemente.

### Principios de Implementación Incremental

1. **Cada fase debe ser ejecutable**: Al completar una fase, debes poder ejecutar y probar la funcionalidad implementada
2. **Testing continuo**: Escribir tests mientras implementas, no al final
3. **Validación incremental**: Validar cada componente antes de pasar al siguiente
4. **Commits frecuentes**: Commit después de completar cada tarea

---

### 🔧 Fase 1: Fundación (Ejecutable: Hello World API)

**Objetivo**: Tener un servidor HTTP básico corriendo

**Tareas**:
1. ✅ Inicializar proyecto Go (`go mod init`)
2. ✅ Instalar dependencias core (Gin, config)
3. ✅ Crear `internal/config/config.go` con variables de entorno
4. ✅ Crear `.env.example` y `.env`
5. ✅ Crear `cmd/api/main.go` con servidor HTTP básico
6. ✅ Agregar endpoint `/health`

**Validación**:
```bash
go run cmd/api/main.go
curl http://localhost:8080/health
# Debe retornar: {"status":"healthy"}
```

**Commit**: `feat: setup básico del proyecto con health endpoint`

---

### 📦 Fase 2: Modelos de Dominio (Ejecutable: Tests unitarios)

**Objetivo**: Definir las estructuras de datos core

**Tareas**:
1. ✅ Crear `internal/domain/product.go`
2. ✅ Crear `internal/domain/stock.go` (con métodos `Available()`, `CanReserve()`)
3. ✅ Crear `internal/domain/reservation.go` (con métodos `IsExpired()`, `CanConfirm()`)
4. ✅ Crear `internal/domain/event.go`
5. ✅ Crear `tests/unit/domain_test.go` para testear lógica de métodos

**Validación**:
```bash
go test ./internal/domain/... -v
# Todos los tests deben pasar
```

**Commit**: `feat: agregar modelos de dominio con lógica de negocio`

---

### 🗄️ Fase 3: Persistencia (Ejecutable: Queries SQL)

**Objetivo**: Conectar a base de datos y ejecutar queries básicas

**Tareas**:
1. ✅ Crear `migrations/001_initial_schema.sql` (SQLite compatible)
2. ✅ Crear `internal/database/database.go` (soporte PostgreSQL y SQLite)
3. ✅ Iniciar Docker Compose (`make docker-up`)
4. ✅ Aplicar migración a PostgreSQL
5. ✅ Crear script de migración para SQLite
6. ✅ Agregar health check de DB en `/health`

**Validación**:
```bash
# PostgreSQL
make docker-up
docker exec -it <postgres_container> psql -U postgres -d inventory -c "SELECT * FROM products;"

# SQLite (para testing)
DATABASE_DRIVER=sqlite SQLITE_PATH=./test.db go run cmd/api/main.go
sqlite3 test.db "SELECT * FROM products;"
```

**Commit**: `feat: agregar soporte para PostgreSQL y SQLite con migraciones`

---

### 🔌 Fase 4: Repositorios (Ejecutable: Integration tests con DB)

**Objetivo**: Implementar operaciones CRUD con la base de datos

**Tareas**:
1. ✅ Crear `internal/repository/product_repository.go`
2. ✅ Crear `internal/repository/stock_repository.go` (con optimistic locking)
3. ✅ Crear `internal/repository/reservation_repository.go`
4. ✅ Crear `internal/repository/event_repository.go`
5. ✅ Crear `tests/integration/repository_test.go`

**Validación**:
```bash
# Integration tests con SQLite in-memory
DATABASE_DRIVER=sqlite SQLITE_PATH=:memory: go test ./tests/integration/... -v
```

**Commit**: `feat: implementar repositorios con optimistic locking`

---

### 📡 Fase 5: Event Bus (Ejecutable: Pub/Sub test)

**Objetivo**: Conectar a NATS y publicar/consumir eventos

**Tareas**:
1. ✅ Crear `internal/eventbus/nats_client.go`
2. ✅ Crear streams en NATS (STOCK_EVENTS, SYNC_EVENTS)
3. ✅ Implementar `Publish()` y `Subscribe()`
4. ✅ Crear `tests/integration/nats_test.go`

**Validación**:
```bash
# Iniciar NATS
docker-compose up -d nats

# Test de pub/sub
go test ./tests/integration/nats_test.go -v
```

**Commit**: `feat: integrar NATS JetStream para eventos`

---

### 🎯 Fase 6: Servicios (Ejecutable: Operaciones de negocio)

**Objetivo**: Implementar lógica de negocio core

**Tareas**:
1. ✅ Crear `internal/service/event_publisher.go` (buffered channel)
2. ✅ Crear `internal/service/stock_service.go`
   - `UpdateStock()`
   - `ReserveStock()`
   - `ConfirmReservation()`
   - `CancelReservation()`
3. ✅ Crear `tests/unit/stock_service_test.go` (con mocks)
4. ✅ Crear `internal/service/sync_service.go` (background worker)

**Validación**:
```bash
# Unit tests con mocks
go test ./tests/unit/stock_service_test.go -v

# Integration test de flujo completo
go test ./tests/integration/stock_flow_test.go -v
```

**Commit**: `feat: implementar servicios de negocio con event publishing`

---

### 🌐 Fase 7: HTTP Handlers (Ejecutable: API REST funcional)

**Objetivo**: Exponer funcionalidad via REST API

**Tareas**:
1. ✅ Crear `internal/handler/stock_handler.go`
2. ✅ Crear `internal/handler/reservation_handler.go`
3. ✅ Crear `internal/handler/query_handler.go`
4. ✅ Agregar rutas en `main.go`
5. ✅ Crear `tests/integration/api_test.go`

**Validación**:
```bash
# Iniciar servidor
go run cmd/api/main.go

# Test manual con cURL
curl -X POST http://localhost:8080/api/v1/reservations \
  -H "Content-Type: application/json" \
  -d '{"productId":"550e8400-e29b-41d4-a716-446655440000","storeId":"store-001","customerId":"cust-001","quantity":1}'

# Integration tests
go test ./tests/integration/api_test.go -v
```

**Commit**: `feat: agregar HTTP handlers para operaciones de inventario`

---

### 🛡️ Fase 8: Middleware y Observabilidad (Ejecutable: API con seguridad)

**Objetivo**: Agregar autenticación, rate limiting, logging y métricas

**Tareas**:
1. ✅ Crear `internal/middleware/logger.go` (zerolog)
2. ✅ Crear `internal/middleware/metrics.go` (Prometheus)
3. ✅ Crear `internal/middleware/error_handler.go`
4. ✅ Crear `internal/middleware/auth_jwt.go`
5. ✅ Crear `internal/middleware/ratelimit.go`
6. ✅ Agregar endpoint `/metrics`
7. ✅ Aplicar middlewares en `main.go`

**Validación**:
```bash
# Generar token JWT para testing
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/token \
  -d '{"storeId":"store-001","userId":"user-001"}' | jq -r '.token')

# Usar token en request
curl http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer $TOKEN"

# Ver métricas de Prometheus
curl http://localhost:8080/metrics

# Test de rate limiting (enviar 101 requests en 1 minuto)
for i in {1..101}; do curl http://localhost:8080/health; done
```

**Commit**: `feat: agregar seguridad, logging y métricas`

---

### 🔄 Fase 9: Worker de Limpieza (Ejecutable: Reservas expiran automáticamente)

**Objetivo**: Implementar cleanup automático de reservas expiradas

**Tareas**:
1. ✅ Agregar goroutine de cleanup en `main.go`
2. ✅ Implementar `cleanupExpiredReservations()` que usa `reservationRepo.GetPendingExpired()`
3. ✅ Agregar graceful shutdown
4. ✅ Crear test de expiración automática

**Validación**:
```bash
# Crear reserva con TTL corto
curl -X POST http://localhost:8080/api/v1/reservations \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"productId":"...","storeId":"store-001","customerId":"cust-001","quantity":1,"ttl":10}'

# Esperar 15 segundos y verificar que expiró
sleep 15
curl http://localhost:8080/api/v1/reservations/{id} -H "Authorization: Bearer $TOKEN"
# Debe mostrar status: EXPIRED
```

**Commit**: `feat: agregar worker de limpieza de reservas expiradas`

---

### 🧪 Fase 10: Testing Comprehensivo (Ejecutable: Suite completa de tests)

**Objetivo**: Cobertura > 70% con tests de concurrencia

**Tareas**:
1. ✅ Completar unit tests para todos los servicios
2. ✅ Agregar tests de concurrencia (race conditions)
3. ✅ Agregar tests de flujo end-to-end
4. ✅ Configurar coverage report

**Validación**:
```bash
# Tests con race detector
go test -race ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
# Verificar que cobertura > 70%

# Benchmark de performance
go test -bench=. -benchmem ./tests/benchmark/...
```

**Commit**: `test: agregar suite completa de tests con >70% cobertura`

---

### 📚 Fase 11: Documentación (Ejecutable: Docs navegables)

**Objetivo**: Documentación completa para usuarios y desarrolladores

**Tareas**:
1. ✅ Crear `README.md` (overview, quick start, arquitectura)
2. ✅ Crear `docs/API.md` (todos los endpoints con ejemplos cURL)
3. ✅ Crear `docs/run.md` (instrucciones paso a paso)
4. ✅ Crear `docs/ARCHITECTURE.md` (decisiones técnicas justificadas)
5. ✅ Agregar diagramas Mermaid

**Validación**:
```bash
# Seguir instrucciones en run.md desde cero
# Verificar que funciona sin conocimiento previo del proyecto
```

**Commit**: `docs: agregar documentación completa del sistema`

---

### 🚀 Fase 12: DevOps (Ejecutable: Deploy con un comando)

**Objetivo**: Makefile, Docker, CI/CD básico

**Tareas**:
1. ✅ Crear `Makefile` con targets útiles
2. ✅ Crear `Dockerfile` multi-stage
3. ✅ Crear `docker-compose.yml` completo
4. ✅ Crear `.env.example`
5. ✅ (Opcional) Crear `.github/workflows/ci.yml`

**Validación**:
```bash
# Build y run con un comando
make build
make run

# Build Docker image
docker build -t inventory-system:latest .

# Run con Docker Compose
docker-compose up -d
curl http://localhost:8080/health
```

**Commit**: `ci: agregar Makefile, Dockerfile y docker-compose`

---

## ✅ Checklist de Validación Final

Antes de considerar el proyecto completo, verificar:

### Funcionalidad Core
- [ ] Cliente puede crear una reserva
- [ ] Reserva expira automáticamente después del TTL
- [ ] Cliente puede confirmar una reserva
- [ ] Cliente puede cancelar una reserva
- [ ] Actualización de stock se sincroniza via eventos
- [ ] Sistema previene overselling (optimistic locking funciona)
- [ ] Cache de Redis funciona correctamente
- [ ] NATS está publicando y consumiendo eventos

### Performance y Concurrencia
- [ ] API responde < 100ms para GET
- [ ] API responde < 200ms para POST/PUT
- [ ] Tests con `-race` pasan sin errores
- [ ] Sistema soporta 100 requests concurrentes
- [ ] Worker de cleanup no afecta performance

### Seguridad y Observabilidad
- [ ] JWT authentication funciona
- [ ] Rate limiting bloquea requests excesivos
- [ ] Logs estructurados con zerolog
- [ ] Métricas de Prometheus expuestas en `/metrics`
- [ ] Health check retorna estado de dependencias

### Testing
- [ ] Cobertura de tests > 70%
- [ ] Unit tests para servicios
- [ ] Integration tests con DB y NATS
- [ ] Tests de concurrencia
- [ ] Tests end-to-end de flujos críticos

### Documentación
- [ ] README.md con quick start funcional
- [ ] run.md con instrucciones paso a paso
- [ ] API.md con todos los endpoints documentados
- [ ] ARCHITECTURE.md con decisiones justificadas
- [ ] Diagramas de arquitectura (Mermaid)
- [ ] Comentarios en código para lógica compleja

### DevOps
- [ ] `make run` inicia la aplicación
- [ ] `make test` ejecuta todos los tests
- [ ] `make docker-up` inicia infraestructura
- [ ] Docker Compose funciona correctamente
- [ ] `.env.example` está completo
- [ ] Dockerfile build exitosamente

---

## Orden de Implementación Recomendado (Resumen)

1. **Fase 1** (Setup): Estructura del proyecto, dependencias, Docker Compose
2. **Fase 2** (Domain): Modelos de dominio
3. **Fase 3** (Database): Clientes de DB, schema SQL
4. **Fase 4** (Repository): Implementar repositorios con optimistic locking
5. **Fase 5** (Event Bus): Cliente NATS
6. **Fase 6** (Service): EventPublisher → StockService → SyncService
7. **Fase 7** (Handlers): HTTP handlers
8. **Fase 8** (Middleware): Logger, Metrics, ErrorHandler, Auth, RateLimit
9. **Fase 9** (Worker): Cleanup de reservas expiradas
10. **Fase 10** (Testing): Unit tests → Integration tests → E2E tests
11. **Fase 11** (Docs): Documentación completa
12. **Fase 12** (DevOps): Makefile, Dockerfile, CI/CD

---

## Criterios de Aceptación

### Funcionalidad
- ✅ Cliente puede reservar producto online
- ✅ Reserva expira automáticamente después de 10 minutos
- ✅ Venta en tienda se refleja en web en < 2 segundos
- ✅ No hay overselling (dos clientes no pueden comprar la última unidad)
- ✅ Sistema se recupera automáticamente de desconexiones temporales

### Performance
- ✅ API responde < 100ms (GET)
- ✅ API responde < 200ms (POST/PUT)
- ✅ Latencia de sincronización < 1 segundo
- ✅ Soporta 100 requests concurrentes sin errores

### Código
- ✅ Cobertura de tests > 70%
- ✅ Sin race conditions (verificar con `go test -race`)
- ✅ Código documentado con comentarios
- ✅ Manejo de errores consistente

### Documentación
- ✅ README con quick start
- ✅ run.md con instrucciones paso a paso
- ✅ API.md con todos los endpoints documentados
- ✅ ARCHITECTURE.md con decisiones técnicas justificadas

---

## Notas Importantes

### Logging Estructurado

Para producción se recomienda usar **zerolog** en lugar de `log.Printf` por las siguientes razones:

1. **Performance**: zerolog es el logger más rápido de Go (zero allocation)
2. **Formato estructurado**: JSON nativo para agregación de logs
3. **Niveles de log**: Debug, Info, Warn, Error
4. **Context awareness**: Fácil agregar campos a cada log

**Instalación**:
```bash
go get github.com/rs/zerolog
```

**Ejemplo de uso**:
```go
package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func init() {
	// Configuración para desarrollo (pretty print)
	if os.Getenv("LOG_FORMAT") != "json" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	
	// Nivel de log desde variable de entorno
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// Uso en el código
log.Info().
	Str("productId", productID).
	Str("storeId", storeID).
	Int("quantity", quantity).
	Msg("Stock updated successfully")

log.Error().
	Err(err).
	Str("operation", "ReserveStock").
	Msg("Failed to reserve stock")
```

**Para este prototipo**: 
- Puedes usar `log.Printf` para simplicidad inicial
- Para producción: Migrar a zerolog es altamente recomendado
- El middleware de logger puede ser actualizado fácilmente para usar zerolog

**Actualizar middleware/logger.go para zerolog** (opcional):
```go
package middleware

import (
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		c.Next()
		
		latency := time.Since(start)
		
		if raw != "" {
			path = path + "?" + raw
		}
		
		// Log estructurado con zerolog
		log.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Str("ip", c.ClientIP()).
			Int("status", c.Writer.Status()).
			Dur("latency", latency).
			Str("user_agent", c.Request.UserAgent()).
			Msg("HTTP request")
	}
}
```

### Por qué Gin en lugar de Fiber

Este proyecto usa **Gin** (github.com/gin-gonic/gin) como framework web por las siguientes razones:

1. **Madurez**: Gin tiene 10 años en producción, es extremadamente estable
2. **Compatibilidad**: Usa `net/http` estándar de Go, compatible con todo el ecosistema
3. **Testing**: Funciona perfectamente con `httptest` de la biblioteca estándar
4. **Comunidad**: 75k+ estrellas, usado por Google, Uber, Alibaba
5. **Middleware**: Ecosistema rico de middleware de terceros (gin-contrib/*)
6. **Documentación**: Abundante documentación y ejemplos
7. **Estabilidad de API**: Pocas breaking changes desde v1.0

Aunque Fiber es más rápido (~50% en benchmarks), para este sistema de inventario la diferencia es imperceptible y no justifica los trade-offs en compatibilidad y estabilidad.

### Optimistic Locking
El campo `version` en la tabla `stock` es CRÍTICO para evitar race conditions. Cada UPDATE debe verificar que la version no haya cambiado:

```sql
UPDATE stock 
SET quantity = $1, version = version + 1 
WHERE product_id = $2 AND store_id = $3 AND version = $4
```

Si el UPDATE afecta 0 filas, significa que otro proceso modificó el stock (conflicto).

### Event Ordering
Los eventos en NATS deben procesarse en orden para cada agregado (producto). Usar JetStream con "Consumer Groups" garantiza esto.

### Idempotencia
Todos los endpoints de comando (POST, PUT, DELETE) deben ser idempotentes usando:
- Client-generated IDs para reservations
- Header `X-Idempotency-Key`
- Verificar duplicados antes de procesar

### Graceful Shutdown
El sistema debe:
1. Dejar de aceptar nuevas requests
2. Completar requests en progreso
3. Cerrar conexiones de DB/NATS limpiamente
4. Esperar máximo 30 segundos

---

## Troubleshooting

### "No se puede conectar a PostgreSQL"
- Verificar que Docker Compose esté corriendo: `docker-compose ps`
- Verificar que el puerto no esté ocupado: `lsof -i :5432`

### "NATS connection refused"
- NATS tarda ~5 segundos en iniciar, esperar un momento
- Verificar logs: `docker-compose logs nats`

### "Tests fallan con race condition"
- Ejecutar con: `go test -race ./...`
- Revisar uso de `sync.Map` en repositorios

### "Eventos no se sincronizan"
- Verificar que Sync Service esté corriendo
- Revisar logs del Sync Service
- Verificar tabla `events` que `synced = true`

---

## Extensiones Futuras (Fuera de Scope del Prototipo)

- Autenticación real con JWT
- Rate limiting por usuario
- Webhook notifications cuando stock es bajo
- Dashboard web en tiempo real
- Métricas con Prometheus
- Distributed tracing con OpenTelemetry
- Replicación multi-región

---

**Fin del Plan de Implementación**

Este archivo debe ser suficiente para que Claude Code pueda implementar el sistema completo siguiendo las instrucciones paso a paso.
