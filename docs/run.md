# 🚀 Guía de Ejecución y Consumo de API - Sistema de Inventario

Esta guía explica cómo ejecutar el proyecto y consumir todos los endpoints de la API REST en **cualquier sistema operativo** (Windows, Linux, macOS).

---

## 📋 Prerrequisitos

Antes de ejecutar el proyecto, asegúrate de tener instalado:

### Requerido

**Go 1.21 o superior**

- **Windows**: Descargar desde [golang.org/dl](https://golang.org/dl/)

- **Linux (Ubuntu/Debian)**: 
  ```bash
  # Remover versión antigua si existe
  sudo apt remove golang-go
  
  # Descargar Go 1.21+ (verificar última versión en golang.org/dl)
  wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
  
  # Extraer en /usr/local
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
  
  # Agregar al PATH (añadir a ~/.bashrc o ~/.profile)
  echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
  source ~/.bashrc
  ```

- **macOS**: `brew install go`

Verificar instalación:
```bash
go version
```

**Salida esperada**: `go version go1.21.0` (o superior)

### Requerido (para Event Publishing)

**Docker** (necesario para ejecutar Redis, que es el message broker por defecto)

- **Windows/macOS**: Descargar [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- **Linux (Ubuntu/Debian)**:
  ```bash
  # Instalar Docker
  curl -fsSL https://get.docker.com -o get-docker.sh
  sudo sh get-docker.sh
  
  # Agregar usuario al grupo docker
  sudo usermod -aG docker $USER
  ```

Verificar instalación:
```bash
docker --version
```

**Salida esperada**: `Docker version 24.0.0` (o superior)

**Nota importante**: El sistema está configurado para usar Redis por defecto (`MESSAGE_BROKER=redis`). Si no quieres usar Redis, debes configurar explícitamente `MESSAGE_BROKER=none` en las variables de entorno (ver sección "Ejecutar sin Redis").

---

## 🔧 Instalación

### Paso 1: Extraer el Archivo ZIP

Extrae el contenido del archivo `inventory-system.zip` en cualquier directorio de tu sistema.

```bash
# Linux/macOS
unzip inventory-system.zip

# Windows (PowerShell)
Expand-Archive -Path inventory-system.zip -DestinationPath C:\Projects\
```

### Paso 2: Navegar al Directorio

```bash
cd inventory-system
```

### Paso 3: Descargar Dependencias

```bash
go mod download
```

**Tiempo estimado**: 1-2 minutos

**Nota**: Requiere conexión a Internet para descargar las dependencias de Go.

---

## ▶️ Ejecución del Proyecto

### ⚠️ IMPORTANTE: Iniciar Redis primero

El sistema usa Redis como message broker por defecto. **Debes iniciar Redis antes de ejecutar el servidor**:

```bash
# Iniciar Redis con Docker (desde el directorio del proyecto)
docker-compose up -d redis
```

**Verificar que Redis está corriendo:**
```bash
docker ps
```

**Deberías ver**:
```
CONTAINER ID   IMAGE            STATUS         PORTS                    NAMES
abc123def456   redis:7-alpine   Up 10 seconds  0.0.0.0:6379->6379/tcp   inventory-redis
```

### Opción 1: Ejecución Directa (Modo Desarrollo)

Una vez Redis está corriendo, ejecuta el proyecto:

```bash
go run cmd/api/main.go
```

**Salida esperada (con Redis)**:
```
2025/10/28 15:30:00 ✅ Connected to SQLite database: :memory:
2025/10/28 15:30:00 📊 Applying database migrations...
2025/10/28 15:30:00 ✅ Database migrations applied successfully
2025/10/28 15:30:00 ✅ Connected to Redis at localhost:6379 (stream: inventory-events)
2025/10/28 15:30:00 ✅ Using Redis Streams as message broker (localhost:6379)
2025/10/28 15:30:00 ⏰ Reservation expiration worker started
2025/10/28 15:30:00 📡 Event synchronization worker started
2025/10/28 15:30:00 🚀 Server starting on port 8080 (instance: api-001)
2025/10/28 15:30:00 🔒 API Keys loaded: 3
2025/10/28 15:30:00 🌐 API available at http://localhost:8080/api/v1
```

---

### Opción 2: Compilar y Ejecutar

Compila el proyecto en un binario ejecutable:

```bash
# Compilar
go build -o bin/inventory-api cmd/api/main.go       # Linux/macOS
go build -o bin/inventory-api.exe cmd/api/main.go   # Windows

# Ejecutar
./bin/inventory-api      # Linux/macOS
.\bin\inventory-api.exe  # Windows
```

**Ventajas**:
- ✅ Más rápido (sin recompilación)
- ✅ Binario portable
- ✅ Listo para despliegue

---

## ✅ Verificación

### 1. Health Check

Una vez el servidor esté corriendo, verifica que funciona:

```bash
curl http://localhost:8080/health
```

**Respuesta esperada**:
```json
{
  "status": "healthy",
  "database": "healthy",
  "db_driver": "sqlite",
  "instance_id": "api-001",
  "timestamp": "2025-10-26T15:30:00-05:00",
  "version": "1.0.0"
}
```

### 2. Crear un Producto (Ejemplo)

```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-store-001" \
  -d '{
    "sku": "LAPTOP-001",
    "name": "MacBook Pro 16",
    "description": "Apple M3 Max, 32GB RAM",
    "category": "electronics",
    "price": 2499.99
  }'
```

**Respuesta esperada**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "sku": "LAPTOP-001",
  "name": "MacBook Pro 16",
  "description": "Apple M3 Max, 32GB RAM",
  "category": "electronics",
  "price": 2499.99,
  "created_at": "2025-10-26T15:30:00Z"
}
```

---

## 🔑 Autenticación

El API requiere una **API Key** en el header `X-API-Key` para endpoints protegidos.

### API Keys por Defecto (Desarrollo)

```
dev-key-store-001  →  Store Madrid
dev-key-store-002  →  Store Barcelona
dev-key-admin      →  Admin
```

### Endpoints Públicos (sin API Key)
- `GET /health`
- `GET /api/v1/products`
- `GET /api/v1/products/:id`
- `GET /api/v1/products/sku/:sku`

### Endpoints Protegidos (requieren API Key)
- `POST /api/v1/products`
- `PUT /api/v1/products/:id`
- `DELETE /api/v1/products/:id`
- Todos los endpoints de `/api/v1/stock/*`
- Todos los endpoints de `/api/v1/reservations/*`

---

## 🧪 Ejecutar Tests

### Tests Unitarios

```bash
go test ./test/unit/... -v
```

**Salida esperada**: `PASS` en 27 tests

### Tests E2E (Requiere Servidor Corriendo)

#### Terminal 1: Iniciar Servidor
```bash
go run cmd/api/main.go
```

#### Terminal 2: Ejecutar Tests
```bash
go test ./test/e2e/... -v
```

**Salida esperada**: `PASS` en 47 tests

### Todos los Tests
```bash
go test ./... -v
```

### Con Cobertura
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## ⚙️ Configuración (Opcional)

El proyecto funciona con configuración por defecto (SQLite in-memory), pero puedes personalizar:

### 1. Copiar Archivo de Configuración

```bash
cp .env.example .env
```

### 2. Editar `.env`

```bash
# Base de Datos (SQLite por defecto)
DATABASE_DRIVER=sqlite
SQLITE_PATH=:memory:

# API Key personalizado (opcional)
API_KEYS=my-key-1:Store_A,my-key-2:Store_B

# Puerto (por defecto: 8080)
SERVER_PORT=8080

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## 📡 Event Publishing con Redis Streams

El sistema utiliza una **arquitectura event-driven** con Redis Streams como message broker para publicar eventos de inventario en tiempo real.

### ¿Qué es Redis Streams?

**Redis Streams** es un sistema de mensajería tipo log que permite:
- **Publicar eventos** de forma persistente (los eventos se almacenan en Redis)
- **Consumir eventos** mediante consumer groups (múltiples servicios pueden procesar eventos)
- **Auditoría y replay** de eventos históricos
- **Baja latencia** (~1-5ms) y alto throughput

### 🎯 Propósito en el Sistema

El sistema sigue el patrón **Event Sourcing + Message Broker**:

```
┌──────────────┐
│   Operación  │  (Ej: Crear reserva, ajustar stock)
└──────┬───────┘
       │
   ┌───┴────┐
   │        │
   ▼        ▼
┌────┐  ┌──────────┐
│ DB │  │ Redis    │
│    │  │ Streams  │
└────┘  └──────────┘
  │          │
  │          └─────► Consumidores externos:
  │                  • Notificaciones push
  │                  • Analytics en tiempo real
  │                  • Microservicios (ej: facturación)
  │                  • Dashboard en tiempo real
  │
  └──────────────► Persistencia y auditoría
```

**Eventos publicados:**
- `product.created`, `product.updated`, `product.deleted`
- `stock.initialized`, `stock.updated`, `stock.adjusted`, `stock.transferred`
- `reservation.created`, `reservation.confirmed`, `reservation.cancelled`, `reservation.expired`

### ✅ Beneficios

- � **Integración con servicios externos** sin acoplar código
- 🔔 **Notificaciones en tiempo real** (push notifications, websockets)
- 📊 **Analytics y reporting** en tiempo real
- � **Arquitectura desacoplada** (microservicios pueden consumir eventos)
- � **Event replay** (reconstruir estado desde eventos)

### 🚀 Iniciar Redis con Docker

**Prerrequisito:** Tener Docker instalado
- **Windows/macOS**: [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- **Linux**: Docker Engine

#### Paso 1: Verificar Docker

```bash
docker --version
```

**Salida esperada**: `Docker version 24.0.0` (o superior)

#### Paso 2: Iniciar Redis

Desde el directorio del proyecto:

```bash
# Iniciar Redis en segundo plano
docker-compose up -d redis
```

**Salida esperada**:
```
[+] Running 2/2
 ✔ Network inventory-system_inventory-network  Created
 ✔ Container inventory-redis                   Started
```

#### Paso 3: Verificar Redis

```bash
# Verificar que Redis está corriendo
docker ps
```

**Deberías ver**:
```
CONTAINER ID   IMAGE            STATUS         PORTS                    NAMES
abc123def456   redis:7-alpine   Up 10 seconds  0.0.0.0:6379->6379/tcp   inventory-redis
```

#### Paso 4: Configurar Variables de Entorno

Editar `.env` (o crear si no existe):

```bash
# Habilitar publicación de eventos con Redis
MESSAGE_BROKER=redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Base de datos (SQLite por defecto)
DATABASE_DRIVER=sqlite
SQLITE_PATH=./inventory.db
```

#### Paso 5: Ejecutar el Servidor

```bash
go run cmd/api/main.go
```

**Salida esperada (con Redis habilitado)**:
```
2025/10/27 15:30:00 ✅ Connected to SQLite database: ./inventory.db
2025/10/27 15:30:00 📊 Applying database migrations...
2025/10/27 15:30:00 ✅ Database migrations applied successfully
2025/10/27 15:30:00 ✅ Connected to Redis at localhost:6379 (stream: inventory-events)
2025/10/27 15:30:00 ✅ Using Redis Streams as message broker (localhost:6379)
2025/10/27 15:30:00 ⏰ Reservation expiration worker started
2025/10/27 15:30:00 📡 Event synchronization worker started
2025/10/27 15:30:00 🚀 Server starting on port 8080 (instance: api-001)
2025/10/27 15:30:00 🔒 API Keys loaded: 3
2025/10/27 15:30:00 🌐 API available at http://localhost:8080/api/v1
```

### 🧪 Probar Event Publishing

Ejecuta una operación y verifica que el evento se publica a Redis:

```bash
# Crear un producto
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-store-001" \
  -d '{
    "sku": "TEST-001",
    "name": "Test Product",
    "category": "test",
    "price": 99.99
  }' | jq
```

**Verás en los logs del servidor:**
```
📤 Event published to Redis: type=product.created, store=MAD-001, id=abc123...
```

### 📊 Eventos Publicados

El sistema publica eventos para:

| Evento | Descripción |
|--------|-------------|
| `product.created` | Nuevo producto creado |
| `product.updated` | Producto actualizado |
| `product.deleted` | Producto eliminado |
| `stock.initialized` | Stock inicializado |
| `stock.updated` | Cantidad de stock actualizada |
| `stock.adjusted` | Ajuste manual de stock |
| `stock.transferred` | Transferencia entre tiendas |
| `reservation.created` | Nueva reserva |
| `reservation.confirmed` | Reserva confirmada |
| `reservation.cancelled` | Reserva cancelada |
| `reservation.expired` | Reserva expirada |

### 🔍 Monitorear Eventos en Redis Streams

Puedes ver eventos publicados en tiempo real:

```bash
# Conectar a Redis CLI
docker exec -it inventory-redis redis-cli

# Leer últimos eventos del stream
XREAD COUNT 10 STREAMS inventory-events 0
```

**Salida esperada:**
```
1) 1) "inventory-events"
   2) 1) 1) "1698765432100-0"
         2) 1) "event_type"
            2) "product.created"
            3) "store_id"
            4) "MAD-001"
            5) "payload"
            6) "{\"id\":\"abc123\",\"sku\":\"TEST-001\"...}"
```

### 🔄 Consumir Eventos (Servicio Externo)

Ejemplo de consumer en Go que lee eventos del stream:

```go
package main

import (
    "context"
    "fmt"
    "github.com/redis/go-redis/v9"
)

func main() {
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    ctx := context.Background()
    lastID := "0" // Leer desde el inicio

    for {
        // Leer nuevos eventos
        streams, err := client.XRead(ctx, &redis.XReadArgs{
            Streams: []string{"inventory-events", lastID},
            Count:   10,
            Block:   0, // Esperar nuevos eventos
        }).Result()

        if err != nil {
            panic(err)
        }

        for _, stream := range streams {
            for _, message := range stream.Messages {
                eventType := message.Values["event_type"]
                payload := message.Values["payload"]
                
                fmt.Printf("Evento recibido: %s\n", eventType)
                fmt.Printf("Payload: %s\n", payload)
                
                // Procesar evento...
                // - Enviar notificación push
                // - Actualizar dashboard
                // - Registrar en analytics
                
                lastID = message.ID
            }
        }
    }
}
```

### 🛑 Detener Redis

```bash
# Detener Redis
docker-compose down redis

# O detener todos los servicios
docker-compose down
```

### ⚙️ Configuración Avanzada

Variables de entorno disponibles para Redis Streams:

```bash
# Message Broker (redis, kafka, none)
MESSAGE_BROKER=redis         # Default: none

# Redis Connection
REDIS_HOST=localhost         # Default: localhost
REDIS_PORT=6379             # Default: 6379

# Stream Configuration
REDIS_STREAM_NAME=inventory-events  # Nombre del stream
REDIS_MAX_LEN=100000                # Retener últimos 100k eventos
```

### ❌ Ejecutar sin Redis (Modo Standalone)

Si no quieres usar Redis/Docker, puedes ejecutar sin el message broker:

**IMPORTANTE**: Debes configurar explícitamente `MESSAGE_BROKER=none`, de lo contrario el servidor intentará conectar a Redis y fallará.

#### Opción 1: Variable de entorno (recomendado)

```bash
# Windows PowerShell
$env:MESSAGE_BROKER="none"; go run cmd/api/main.go

# Linux/macOS
MESSAGE_BROKER=none go run cmd/api/main.go
```

#### Opción 2: Archivo .env

Crear o editar `.env`:
```bash
MESSAGE_BROKER=none
```

Luego ejecutar:
```bash
go run cmd/api/main.go
```

**Salida esperada (sin Redis)**:
```
2025/10/28 15:30:00 ✅ Connected to SQLite database: :memory:
2025/10/28 15:30:00 📊 Applying database migrations...
2025/10/28 15:30:00 ✅ Database migrations applied successfully
2025/10/28 15:30:00 ⚠️  No message broker configured (MESSAGE_BROKER=none)
2025/10/28 15:30:00 ⏰ Reservation expiration worker started
2025/10/28 15:30:00 📡 Event synchronization worker started
2025/10/28 15:30:00 🚀 Server starting on port 8080 (instance: api-001)
2025/10/28 15:30:00 🔒 API Keys loaded: 3
2025/10/28 15:30:00 🌐 API available at http://localhost:8080/api/v1
```

**Nota**: Sin Redis, el sistema funciona perfectamente pero NO publica eventos a servicios externos. Los eventos se siguen guardando en la base de datos para auditoría.

### 🔄 Arquitectura Flexible: Cambiar a Kafka

La arquitectura soporta múltiples brokers sin cambiar código de negocio:

```bash
# Redis Streams (actual - implementado)
MESSAGE_BROKER=redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Apache Kafka (futuro - requiere implementar kafka_publisher.go)
MESSAGE_BROKER=kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=inventory-events

# Sin broker
MESSAGE_BROKER=none
```

**Ventaja clave**: Cambiar de Redis a Kafka solo requiere implementar `KafkaPublisher` sin modificar servicios de negocio (Dependency Inversion Principle).

---

---

## �🛑 Detener el Servidor

### Modo Desarrollo (go run)
Presiona `Ctrl + C` en la terminal

### Modo Producción (binario)

```bash
# Encontrar proceso
ps aux | grep inventory-api    # Linux/macOS
Get-Process -Name "inventory-api"  # Windows

# Detener
pkill inventory-api            # Linux/macOS
Stop-Process -Name "inventory-api"  # Windows
```

---

## 🐛 Troubleshooting

### Error: "redis: connection pool: failed to dial"

**Error completo**:
```
redis: connection pool: failed to dial after 5 attempts: dial tcp [::1]:6379: connectex: No connection could be made because the target machine actively refused it.
Failed to initialize event publisher: failed to create Redis publisher
```

**Causa**: El servidor está configurado para usar Redis (`MESSAGE_BROKER=redis` por defecto) pero Redis no está corriendo.

**Soluciones**:

1. **Opción A - Iniciar Redis** (recomendado):
   ```bash
   docker-compose up -d redis
   ```

2. **Opción B - Deshabilitar Redis**:
   ```bash
   # Windows PowerShell
   $env:MESSAGE_BROKER="none"; go run cmd/api/main.go
   
   # Linux/macOS
   MESSAGE_BROKER=none go run cmd/api/main.go
   ```

3. **Opción C - Configurar en .env**:
   Crear `.env`:
   ```
   MESSAGE_BROKER=none
   ```

---

### Error: "Port 8080 already in use"

**Causa**: Otro proceso está usando el puerto 8080

**Solución**:

```bash
# Encontrar proceso en puerto 8080
lsof -i :8080             # Linux/macOS
netstat -ano | findstr :8080  # Windows

# Matar proceso
kill -9 <PID>             # Linux/macOS
taskkill /PID <PID> /F    # Windows
```

**O cambiar puerto en `.env`**:
```bash
SERVER_PORT=8081
```

---

### Error: "go: command not found"

**Causa**: Go no está instalado o no está en PATH

**Solución**:
1. Instalar Go desde [golang.org/dl](https://golang.org/dl/)
2. Verificar PATH:
   ```bash
   echo $PATH      # Linux/macOS
   $env:PATH       # Windows
   ```

---

### Error: "failed to connect to database"

**Causa**: Problema con la base de datos

**Solución**:
1. Verificar configuración en `.env`
2. Para SQLite, asegurar que el directorio existe

---

### Tests E2E fallan con "connection refused"

**Causa**: Servidor no está corriendo

**Solución**:
1. Iniciar servidor en terminal separada:
   ```bash
   go run cmd/api/main.go
   ```
2. Esperar a ver mensaje "Server starting on port 8080"
3. Ejecutar tests en otra terminal

---
---

## 🎯 Resumen de Comandos

| Acción | Windows | Linux/macOS |
|--------|---------|-------------|
| **Extraer ZIP** | `Expand-Archive inventory-system.zip` | `unzip inventory-system.zip` |
| **Navegar** | `cd inventory-system` | `cd inventory-system` |
| **Dependencias** | `go mod download` | `go mod download` |
| **Iniciar Redis** ⚠️ | `docker-compose up -d redis` | `docker-compose up -d redis` |
| **Ejecutar** | `go run cmd/api/main.go` | `go run cmd/api/main.go` |
| **Ejecutar sin Redis** | `$env:MESSAGE_BROKER="none"; go run cmd/api/main.go` | `MESSAGE_BROKER=none go run cmd/api/main.go` |
| **Compilar** | `go build -o bin/inventory-api.exe cmd/api/main.go` | `go build -o bin/inventory-api cmd/api/main.go` |
| **Tests** | `go test ./... -v` | `go test ./... -v` |
| **Health Check** | `Invoke-RestMethod http://localhost:8080/health` | `curl http://localhost:8080/health` |
| **Detener Servidor** | `Ctrl + C` | `Ctrl + C` |
| **Detener Redis** | `docker-compose down` | `docker-compose down` |

⚠️ **Redis es REQUERIDO por defecto**. Si no quieres usarlo, configura `MESSAGE_BROKER=none`.

---

## ✅ Verificación de Ejecución Exitosa

Si ves estos mensajes, el proyecto está funcionando correctamente:

**Con Redis (configuración por defecto):**
```
✅ Connected to SQLite database
✅ Database migrations applied successfully
✅ Connected to Redis at localhost:6379 (stream: inventory-events)
✅ Using Redis Streams as message broker (localhost:6379)
⏰ Reservation expiration worker started
📡 Event synchronization worker started
🚀 Server starting on port 8080
🔒 API Keys loaded: 3
🌐 API available at http://localhost:8080/api/v1
```

**Sin Redis (MESSAGE_BROKER=none):**
```
✅ Connected to SQLite database
✅ Database migrations applied successfully
⚠️  No message broker configured (MESSAGE_BROKER=none)
⏰ Reservation expiration worker started
📡 Event synchronization worker started
🚀 Server starting on port 8080
🔒 API Keys loaded: 3
🌐 API available at http://localhost:8080/api/v1
```

**El servidor está listo para recibir peticiones** 🎉

---

## 📦 Contenido del ZIP

El archivo `inventory-system.zip` contiene:

```
inventory-system/
├── cmd/                    # Punto de entrada de la aplicación
├── internal/              # Código fuente (domain, handlers, services, repos)
├── test/                  # Tests unitarios y E2E
├── docs/                  # Documentación adicional
├── migrations/            # Migraciones de base de datos (si aplica)
├── go.mod                 # Dependencias de Go
├── go.sum                 # Checksums de dependencias
├── .env.example          # Ejemplo de configuración
├── docker-compose.yml    # Infraestructura Docker (opcional)
├── README.md             # Documentación principal
├── run.md                # Esta guía
└── ANALISIS_CUMPLIMIENTO.md  # Análisis de cumplimiento de requisitos
```

---

## 🚀 Quick Start (Inicio Rápido)

Para evaluar el proyecto rápidamente:

```bash
# 1. Extraer ZIP
unzip inventory-system.zip    # Linux/macOS
# o Expand-Archive en Windows

# 2. Navegar
cd inventory-system

# 3. Descargar dependencias
go mod download

# 4. Iniciar Redis (REQUERIDO por defecto)
docker-compose up -d redis

# 5. Ejecutar
go run cmd/api/main.go

# 6. Verificar (en otra terminal)
curl http://localhost:8080/health
```

**Tiempo total**: ~3-5 minutos (incluyendo descarga de dependencias)

**Alternativa sin Docker/Redis**: Si no quieres usar Docker, configura `MESSAGE_BROKER=none`:
```bash
# Windows PowerShell
$env:MESSAGE_BROKER="none"; go run cmd/api/main.go

# Linux/macOS
MESSAGE_BROKER=none go run cmd/api/main.go
```

---

## 📖 Documentación Completa de Endpoints

A continuación se encuentra la documentación detallada de todos los endpoints de la API con ejemplos de request y response que puedes ejecutar manualmente.

### 🔑 Autenticación

**Todas las peticiones protegidas requieren el header:**
```
X-API-Key: dev-key-store-001
```

**API Keys Disponibles:**
```
dev-key-store-001  →  Store Madrid (MAD-001)
dev-key-store-002  →  Store Barcelona (BCN-001)
dev-key-admin      →  Admin (acceso completo)
```

---

## 1️⃣ Health Check

### GET /health
Verifica el estado del servidor y sus dependencias.

**Endpoint Público** (No requiere autenticación)

#### Request
```bash
curl http://localhost:8080/health | jq
```

#### Response (200 OK)
```json
{
  "status": "healthy",
  "database": "healthy",
  "db_driver": "sqlite",
  "instance_id": "api-001",
  "timestamp": "2025-10-27T21:57:22-05:00",
  "version": "1.0.0"
}
```

---

## 2️⃣ Products (Productos)

### 2.1 POST /api/v1/products
Crea un nuevo producto en el catálogo.

**Requiere:** `X-API-Key`

#### Request
```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-store-001" \
  -d '{
    "sku": "LAPTOP-HP-2024",
    "name": "Laptop HP Pavilion 15",
    "description": "15.6 FHD, Intel i5-1135G7, 8GB RAM, 256GB SSD",
    "category": "electronics",
    "price": 599.99
  }' | jq
```

#### Request Body
```json
{
  "sku": "LAPTOP-HP-2024",
  "name": "Laptop HP Pavilion 15",
  "description": "15.6 FHD, Intel i5-1135G7, 8GB RAM, 256GB SSD",
  "category": "electronics",
  "price": 599.99
}
```

#### Response (201 Created)
```json
{
  "id": "8dd21a53-d570-4d64-a973-a8d8cfb70bd6",
  "sku": "LAPTOP-HP-2024",
  "name": "Laptop HP Pavilion 15",
  "description": "15.6 FHD, Intel i5-1135G7, 8GB RAM, 256GB SSD",
  "category": "electronics",
  "price": 599.99,
  "created_at": "2025-10-27T22:00:00Z",
  "updated_at": null
}
```

---

### 2.2 GET /api/v1/products/{id}
Obtiene un producto por su ID.

**Endpoint Público** (No requiere autenticación)

#### Request
```bash
curl http://localhost:8080/api/v1/products/550e8400-e29b-41d4-a716-446655440000 | jq
```

#### Response (200 OK)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "sku": "PROD-001",
  "name": "Laptop HP Pavilion 15",
  "description": "15.6\" FHD, Intel i5-1135G7, 8GB RAM, 256GB SSD",
  "category": "electronics",
  "price": 599.99,
  "created_at": "2025-10-27T22:00:00Z",
  "updated_at": null
}
```

---

### 2.3 GET /api/v1/products/sku/{sku}
Obtiene un producto por su SKU.

**Endpoint Público** (No requiere autenticación)

#### Request
```bash
curl http://localhost:8080/api/v1/products/sku/PROD-001 | jq
```

#### Response (200 OK)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "sku": "PROD-001",
  "name": "Laptop HP Pavilion 15",
  "description": "15.6\" FHD, Intel i5-1135G7, 8GB RAM, 256GB SSD",
  "category": "electronics",
  "price": 599.99,
  "created_at": "2025-10-27T22:00:00Z",
  "updated_at": null
}
```

---

### 2.4 GET /api/v1/products
Lista todos los productos con paginación y filtros.

**Endpoint Público** (No requiere autenticación)

**Query Parameters:**
- `page` (opcional): Número de página (default: 1)
- `page_size` (opcional): Tamaño de página (default: 10)
- `category` (opcional): Filtrar por categoría

#### Request
```bash
# Listar con paginación
curl "http://localhost:8080/api/v1/products?page=1&page_size=10" | jq

# Filtrar por categoría
curl "http://localhost:8080/api/v1/products?category=electronics" | jq
```

#### Response (200 OK)
```json
{
  "products": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "sku": "PROD-001",
      "name": "Laptop HP Pavilion 15",
      "description": "15.6\" FHD, Intel i5-1135G7, 8GB RAM, 256GB SSD",
      "category": "electronics",
      "price": 599.99,
      "created_at": "2025-10-27T22:00:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "sku": "PROD-002",
      "name": "Mouse Logitech MX Master 3",
      "description": "Wireless, ergonómico, 7 botones programables",
      "category": "accessories",
      "price": 99.99,
      "created_at": "2025-10-27T22:00:00Z"
    }
  ],
  "total": 6,
  "page": 1,
  "page_size": 10,
  "total_pages": 1
}
```

---

### 2.5 PUT /api/v1/products/{id}
Actualiza un producto existente.

**Requiere:** `X-API-Key`

#### Request
```bash
curl -X PUT http://localhost:8080/api/v1/products/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-store-001" \
  -d '{
    "sku": "PROD-001",
    "name": "Laptop HP Pavilion 15 - UPDATED",
    "description": "15.6 FHD, Intel i7, 16GB RAM, 512GB SSD",
    "category": "electronics",
    "price": 799.99
  }' | jq
```

#### Request Body
```json
{
  "sku": "PROD-001",
  "name": "Laptop HP Pavilion 15 - UPDATED",
  "description": "15.6 FHD, Intel i7, 16GB RAM, 512GB SSD",
  "category": "electronics",
  "price": 799.99
}
```

#### Response (200 OK)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "sku": "PROD-001",
  "name": "Laptop HP Pavilion 15 - UPDATED",
  "description": "15.6 FHD, Intel i7, 16GB RAM, 512GB SSD",
  "category": "electronics",
  "price": 799.99,
  "created_at": "2025-10-27T22:00:00Z",
  "updated_at": "2025-10-27T22:15:00Z"
}
```

---

### 2.6 DELETE /api/v1/products/{id}
Elimina un producto del catálogo.

**Requiere:** `X-API-Key`

#### Request
```bash
curl -X DELETE http://localhost:8080/api/v1/products/550e8400-e29b-41d4-a716-446655440000 \
  -H "X-API-Key: dev-key-store-001"
```

#### Response (204 No Content)
```
(sin contenido - eliminación exitosa)
```

---

## 3️⃣ Stock (Inventario)

### 3.1 POST /api/v1/stock
Inicializa el stock de un producto en una tienda.

**Requiere:** `X-API-Key`

#### Request
```bash
curl -X POST http://localhost:8080/api/v1/stock \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-store-001" \
  -d '{
    "product_id": "550e8400-e29b-41d4-a716-446655440000",
    "store_id": "MAD-001",
    "initial_quantity": 100
  }' | jq
```

#### Request Body
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "store_id": "MAD-001",
  "initial_quantity": 100
}
```

#### Response (201 Created)
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "store_id": "MAD-001",
  "quantity": 100,
  "reserved_qty": 0,
  "available_qty": 100,
  "min_stock": 0,
  "max_stock": 0,
  "reorder_point": 0,
  "reorder_qty": 0,
  "version": 1,
  "last_updated": "2025-10-27T22:00:00Z"
}
```

---

### 3.2 GET /api/v1/stock/{product_id}/{store_id}
Obtiene el stock de un producto en una tienda específica.

**Requiere:** `X-API-Key`

#### Request
```bash
curl http://localhost:8080/api/v1/stock/550e8400-e29b-41d4-a716-446655440000/MAD-001 \
  -H "X-API-Key: dev-key-store-001" | jq
```

#### Response (200 OK)
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "store_id": "MAD-001",
  "quantity": 10,
  "reserved_qty": 0,
  "available_qty": 10,
  "min_stock": 0,
  "max_stock": 0,
  "reorder_point": 0,
  "reorder_qty": 0,
  "version": 1,
  "last_updated": "2025-10-27T22:00:00Z",
  "product_name": "Laptop HP Pavilion 15",
  "product_sku": "PROD-001"
}
```

---

### 3.3 PUT /api/v1/stock/{product_id}/{store_id}
Actualiza la configuración de stock.

**Requiere:** `X-API-Key`

#### Request
```bash
curl -X PUT http://localhost:8080/api/v1/stock/550e8400-e29b-41d4-a716-446655440000/MAD-001 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-store-001" \
  -d '{
    "quantity": 150,
    "min_stock": 15,
    "max_stock": 250,
    "reorder_point": 25,
    "reorder_qty": 75
  }' | jq
```

#### Request Body
```json
{
  "quantity": 150,
  "min_stock": 15,
  "max_stock": 250,
  "reorder_point": 25,
  "reorder_qty": 75
}
```

#### Response (200 OK)
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "store_id": "MAD-001",
  "quantity": 150,
  "reserved_qty": 0,
  "available_qty": 150,
  "min_stock": 15,
  "max_stock": 250,
  "reorder_point": 25,
  "reorder_qty": 75,
  "version": 2,
  "last_updated": "2025-10-27T22:10:00Z"
}
```

---

### 3.4 POST /api/v1/stock/{product_id}/{store_id}/adjust
Ajusta el stock con un incremento o decremento.

**Requiere:** `X-API-Key`

#### Request
```bash
curl -X POST http://localhost:8080/api/v1/stock/550e8400-e29b-41d4-a716-446655440000/MAD-001/adjust \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-store-001" \
  -d '{
    "adjustment": -20,
    "reason": "Devolución de productos defectuosos"
  }' | jq
```

#### Request Body
```json
{
  "adjustment": -20,
  "reason": "Devolución de productos defectuosos"
}
```

#### Response (200 OK)
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "store_id": "MAD-001",
  "quantity": 130,
  "reserved_qty": 0,
  "available_qty": 130,
  "version": 3,
  "last_updated": "2025-10-27T22:15:00Z"
}
```

---

### 3.5 GET /api/v1/stock/{product_id}/{store_id}/availability
Verifica si hay suficiente stock disponible.

**Requiere:** `X-API-Key`

**Query Parameters:**
- `quantity` (requerido): Cantidad a verificar

#### Request
```bash
curl "http://localhost:8080/api/v1/stock/550e8400-e29b-41d4-a716-446655440000/MAD-001/availability?quantity=50" \
  -H "X-API-Key: dev-key-store-001" | jq
```

#### Response (200 OK)
```json
{
  "sufficient": true,
  "available": 130,
  "requested": 50
}
```

---

### 3.6 GET /api/v1/stock/product/{product_id}
Obtiene el stock de un producto en todas las tiendas.

**Requiere:** `X-API-Key`

#### Request
```bash
curl http://localhost:8080/api/v1/stock/product/550e8400-e29b-41d4-a716-446655440000 \
  -H "X-API-Key: dev-key-store-001" | jq
```

#### Response (200 OK)
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "stores": [
    {
      "store_id": "MAD-001",
      "quantity": 10,
      "reserved_qty": 0,
      "available_qty": 10
    },
    {
      "store_id": "BCN-001",
      "quantity": 15,
      "reserved_qty": 2,
      "available_qty": 13
    },
    {
      "store_id": "VAL-001",
      "quantity": 5,
      "reserved_qty": 1,
      "available_qty": 4
    }
  ],
  "total_quantity": 30,
  "total_reserved": 3,
  "total_available": 27
}
```

---

### 3.7 GET /api/v1/stock/store/{store_id}
Obtiene todo el inventario de una tienda.

**Requiere:** `X-API-Key`

#### Request
```bash
curl http://localhost:8080/api/v1/stock/store/MAD-001 \
  -H "X-API-Key: dev-key-store-001" | jq
```

#### Response (200 OK)
```json
{
  "store_id": "MAD-001",
  "items": [
    {
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "product_name": "Laptop HP Pavilion 15",
      "product_sku": "PROD-001",
      "quantity": 10,
      "reserved_qty": 0,
      "available_qty": 10
    },
    {
      "product_id": "550e8400-e29b-41d4-a716-446655440001",
      "product_name": "Mouse Logitech MX Master 3",
      "product_sku": "PROD-002",
      "quantity": 50,
      "reserved_qty": 5,
      "available_qty": 45
    }
  ],
  "count": 2
}
```

---

### 3.8 GET /api/v1/stock/low-stock
Obtiene items con stock bajo.

**Requiere:** `X-API-Key`

**Query Parameters:**
- `threshold` (opcional): Umbral de bajo stock (default: 10)

#### Request
```bash
curl "http://localhost:8080/api/v1/stock/low-stock?threshold=20" \
  -H "X-API-Key: dev-key-store-001" | jq
```

#### Response (200 OK)
```json
{
  "threshold": 20,
  "items": [
    {
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "product_name": "Laptop HP Pavilion 15",
      "product_sku": "PROD-001",
      "store_id": "MAD-001",
      "quantity": 10,
      "available_qty": 10
    },
    {
      "product_id": "550e8400-e29b-41d4-a716-446655440003",
      "product_name": "Monitor LG 27\" 4K",
      "product_sku": "PROD-004",
      "store_id": "MAD-001",
      "quantity": 5,
      "available_qty": 4
    }
  ],
  "count": 2
}
```

---

### 3.9 POST /api/v1/stock/transfer
Transfiere stock entre tiendas.

**Requiere:** `X-API-Key`

#### Request
```bash
curl -X POST http://localhost:8080/api/v1/stock/transfer \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-admin" \
  -d '{
    "product_id": "550e8400-e29b-41d4-a716-446655440000",
    "from_store_id": "MAD-001",
    "to_store_id": "BCN-001",
    "quantity": 30,
    "reason": "Rebalanceo de inventario"
  }' | jq
```

#### Request Body
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "from_store_id": "MAD-001",
  "to_store_id": "BCN-001",
  "quantity": 30,
  "reason": "Rebalanceo de inventario"
}
```

#### Response (200 OK)
```json
{
  "message": "stock transferred successfully",
  "from_store": {
    "store_id": "MAD-001",
    "product_id": "550e8400-e29b-41d4-a716-446655440000",
    "quantity": 70
  },
  "to_store": {
    "store_id": "BCN-001",
    "product_id": "550e8400-e29b-41d4-a716-446655440000",
    "quantity": 45
  }
}
```

---

## 4️⃣ Reservations (Reservas)

### 4.1 POST /api/v1/reservations
Crea una nueva reserva de producto.

**Requiere:** `X-API-Key`

#### Request
```bash
curl -X POST http://localhost:8080/api/v1/reservations \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-store-001" \
  -d '{
    "product_id": "550e8400-e29b-41d4-a716-446655440001",
    "store_id": "MAD-001",
    "customer_id": "CUST-12345",
    "quantity": 5,
    "ttl_minutes": 30
  }' | jq
```

#### Request Body
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440001",
  "store_id": "MAD-001",
  "customer_id": "CUST-12345",
  "quantity": 5,
  "ttl_minutes": 30
}
```

#### Response (201 Created)
```json
{
  "id": "7ed6c7c8-630d-4077-a18e-61564072fdb4",
  "product_id": "550e8400-e29b-41d4-a716-446655440001",
  "store_id": "MAD-001",
  "customer_id": "CUST-12345",
  "quantity": 5,
  "status": "PENDING",
  "expires_at": "2025-10-27T22:30:00Z",
  "created_at": "2025-10-27T22:00:00Z",
  "updated_at": null
}
```

---

### 4.2 GET /api/v1/reservations/{id}
Obtiene una reserva por su ID.

**Requiere:** `X-API-Key`

#### Request
```bash
curl http://localhost:8080/api/v1/reservations/7ed6c7c8-630d-4077-a18e-61564072fdb4 \
  -H "X-API-Key: dev-key-store-001" | jq
```

#### Response (200 OK)
```json
{
  "id": "7ed6c7c8-630d-4077-a18e-61564072fdb4",
  "product_id": "550e8400-e29b-41d4-a716-446655440001",
  "store_id": "MAD-001",
  "customer_id": "CUST-12345",
  "quantity": 5,
  "status": "PENDING",
  "expires_at": "2025-10-27T22:30:00Z",
  "created_at": "2025-10-27T22:00:00Z",
  "updated_at": null
}
```

---

### 4.3 POST /api/v1/reservations/{id}/confirm
Confirma una reserva (convierte stock reservado en venta).

**Requiere:** `X-API-Key`

#### Request
```bash
curl -X POST http://localhost:8080/api/v1/reservations/7ed6c7c8-630d-4077-a18e-61564072fdb4/confirm \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-store-001" | jq
```

#### Response (200 OK)
```json
{
  "id": "7ed6c7c8-630d-4077-a18e-61564072fdb4",
  "product_id": "550e8400-e29b-41d4-a716-446655440001",
  "store_id": "MAD-001",
  "customer_id": "CUST-12345",
  "quantity": 5,
  "status": "CONFIRMED",
  "expires_at": "2025-10-27T22:30:00Z",
  "created_at": "2025-10-27T22:00:00Z",
  "updated_at": "2025-10-27T22:10:00Z"
}
```

---

### 4.4 POST /api/v1/reservations/{id}/cancel
Cancela una reserva (libera stock reservado).

**Requiere:** `X-API-Key`

#### Request
```bash
curl -X POST http://localhost:8080/api/v1/reservations/7ed6c7c8-630d-4077-a18e-61564072fdb4/cancel \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-store-001" | jq
```

#### Response (200 OK)
```json
{
  "id": "7ed6c7c8-630d-4077-a18e-61564072fdb4",
  "product_id": "550e8400-e29b-41d4-a716-446655440001",
  "store_id": "MAD-001",
  "customer_id": "CUST-12345",
  "quantity": 5,
  "status": "CANCELLED",
  "expires_at": "2025-10-27T22:30:00Z",
  "created_at": "2025-10-27T22:00:00Z",
  "updated_at": "2025-10-27T22:12:00Z"
}
```

---

### 4.5 GET /api/v1/reservations/store/{store_id}/pending
Obtiene todas las reservas pendientes de una tienda.

**Requiere:** `X-API-Key`

#### Request
```bash
curl http://localhost:8080/api/v1/reservations/store/MAD-001/pending \
  -H "X-API-Key: dev-key-store-001" | jq
```

#### Response (200 OK)
```json
{
  "store_id": "MAD-001",
  "reservations": [
    {
      "id": "7ed6c7c8-630d-4077-a18e-61564072fdb4",
      "product_id": "550e8400-e29b-41d4-a716-446655440001",
      "customer_id": "CUST-12345",
      "quantity": 5,
      "status": "PENDING",
      "expires_at": "2025-10-27T22:30:00Z",
      "created_at": "2025-10-27T22:00:00Z"
    }
  ],
  "count": 1
}
```

---

### 4.6 GET /api/v1/reservations/product/{product_id}/store/{store_id}
Obtiene todas las reservas de un producto en una tienda.

**Requiere:** `X-API-Key`

**Query Parameters:**
- `status` (opcional): Filtrar por estado (PENDING, CONFIRMED, CANCELLED, EXPIRED)

#### Request
```bash
# Todas las reservas
curl http://localhost:8080/api/v1/reservations/product/550e8400-e29b-41d4-a716-446655440001/store/MAD-001 \
  -H "X-API-Key: dev-key-store-001" | jq

# Filtrar por estado
curl "http://localhost:8080/api/v1/reservations/product/550e8400-e29b-41d4-a716-446655440001/store/MAD-001?status=PENDING" \
  -H "X-API-Key: dev-key-store-001" | jq
```

#### Response (200 OK)
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440001",
  "store_id": "MAD-001",
  "status": "",
  "reservations": [
    {
      "id": "7ed6c7c8-630d-4077-a18e-61564072fdb4",
      "customer_id": "CUST-12345",
      "quantity": 5,
      "status": "PENDING",
      "expires_at": "2025-10-27T22:30:00Z",
      "created_at": "2025-10-27T22:00:00Z"
    }
  ],
  "count": 1
}
```

---

### 4.7 GET /api/v1/reservations/stats
Obtiene estadísticas generales de reservas.

**Requiere:** `X-API-Key` (admin)

#### Request
```bash
curl http://localhost:8080/api/v1/reservations/stats \
  -H "X-API-Key: dev-key-admin" | jq
```

#### Response (200 OK)
```json
{
  "total_reservations": 0,
  "pending_reservations": 0,
  "confirmed_reservations": 0,
  "cancelled_reservations": 0,
  "by_status": {}
}
```

---

## 📊 Resumen de Endpoints

### Total de Endpoints: 27

**Públicos (4):**
- ✅ GET /health
- ✅ GET /api/v1/products
- ✅ GET /api/v1/products/{id}
- ✅ GET /api/v1/products/sku/{sku}

**Protegidos - Products (3):**
- 🔐 POST /api/v1/products
- 🔐 PUT /api/v1/products/{id}
- 🔐 DELETE /api/v1/products/{id}

**Protegidos - Stock (9):**
- 🔐 POST /api/v1/stock
- 🔐 GET /api/v1/stock/{product_id}/{store_id}
- 🔐 PUT /api/v1/stock/{product_id}/{store_id}
- 🔐 POST /api/v1/stock/{product_id}/{store_id}/adjust
- 🔐 GET /api/v1/stock/{product_id}/{store_id}/availability
- 🔐 GET /api/v1/stock/product/{product_id}
- 🔐 GET /api/v1/stock/store/{store_id}
- 🔐 GET /api/v1/stock/low-stock
- 🔐 POST /api/v1/stock/transfer

**Protegidos - Reservations (7):**
- 🔐 POST /api/v1/reservations
- 🔐 GET /api/v1/reservations/{id}
- 🔐 POST /api/v1/reservations/{id}/confirm
- 🔐 POST /api/v1/reservations/{id}/cancel
- 🔐 GET /api/v1/reservations/store/{store_id}/pending
- 🔐 GET /api/v1/reservations/product/{product_id}/store/{store_id}
- 🔐 GET /api/v1/reservations/stats

---

## 🔐 Manejo de Errores

### 400 Bad Request
```json
{
  "error": "validation error",
  "message": "SKU is required"
}
```

### 401 Unauthorized
```json
{
  "error": "unauthorized",
  "message": "missing X-API-Key header"
}
```

### 404 Not Found
```json
{
  "error": "not found",
  "message": "product not found"
}
```

### 409 Conflict
```json
{
  "error": "conflict",
  "message": "product with SKU 'LAPTOP-HP-2024' already exists"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error",
  "message": "database connection failed"
}
```

---

**¿Problemas?** Revisa la sección [Troubleshooting](#-troubleshooting) o consulta la documentación en `docs/`.
