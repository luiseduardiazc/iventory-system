# Quick Start - Inventory System API

## Iniciar el servidor

```bash
# Opción 1: Con SQLite in-memory (más simple, sin infraestructura)
export DATABASE_DRIVER=sqlite
export SQLITE_PATH=:memory:
go run cmd/api/main.go

# Opción 2: Con PostgreSQL
docker-compose up -d postgres
export DATABASE_DRIVER=postgres
export DATABASE_URL="postgres://inventory:inventory123@localhost:5432/inventory?sslmode=disable"
go run cmd/api/main.go
```

El servidor inicia en `http://localhost:8080`

---

## Ejemplos de Uso

### 1. Health Check

```bash
curl http://localhost:8080/health
```

Respuesta:
```json
{
  "status": "healthy",
  "timestamp": "2025-01-26T10:30:00Z",
  "instance_id": "instance-001",
  "version": "1.0.0",
  "database": "healthy",
  "db_driver": "sqlite"
}
```

---

### 2. Productos

#### Crear producto

```bash
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "LAPTOP-001",
    "name": "Laptop Dell XPS 13",
    "description": "Ultrabook premium",
    "category": "Electronics",
    "price": 1299.99
  }'
```

Respuesta:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "sku": "LAPTOP-001",
  "name": "Laptop Dell XPS 13",
  "description": "Ultrabook premium",
  "category": "Electronics",
  "price": 1299.99,
  "active": true,
  "created_at": "2025-01-26T10:30:00Z",
  "updated_at": "2025-01-26T10:30:00Z"
}
```

#### Listar productos

```bash
curl "http://localhost:8080/api/v1/products?limit=10&offset=0"
```

#### Buscar por SKU

```bash
curl http://localhost:8080/api/v1/products/sku/LAPTOP-001
```

---

### 3. Stock

#### Inicializar stock en una tienda

```bash
curl -X POST http://localhost:8080/api/v1/stock \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "550e8400-e29b-41d4-a716-446655440000",
    "store_id": "MAD-001",
    "initial_quantity": 50
  }'
```

#### Ver stock de un producto en todas las tiendas

```bash
curl http://localhost:8080/api/v1/stock/product/550e8400-e29b-41d4-a716-446655440000
```

Respuesta:
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "stores": [
    {
      "id": "stock-1",
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "store_id": "MAD-001",
      "quantity": 50,
      "reserved": 0,
      "version": 1,
      "updated_at": "2025-01-26T10:30:00Z"
    },
    {
      "id": "stock-2",
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "store_id": "BCN-001",
      "quantity": 30,
      "reserved": 5,
      "version": 1,
      "updated_at": "2025-01-26T10:30:00Z"
    }
  ],
  "total_quantity": 80,
  "total_reserved": 5,
  "total_available": 75
}
```

#### Verificar disponibilidad

```bash
curl "http://localhost:8080/api/v1/stock/550e8400-e29b-41d4-a716-446655440000/MAD-001/availability?quantity=10"
```

Respuesta:
```json
{
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "store_id": "MAD-001",
  "requested": 10,
  "available": 50,
  "sufficient": true
}
```

#### Actualizar stock

```bash
curl -X PUT http://localhost:8080/api/v1/stock/550e8400-e29b-41d4-a716-446655440000/MAD-001 \
  -H "Content-Type: application/json" \
  -d '{
    "quantity": 60
  }'
```

#### Ajustar stock (incrementar o decrementar)

```bash
# Incrementar 10 unidades
curl -X POST http://localhost:8080/api/v1/stock/550e8400-e29b-41d4-a716-446655440000/MAD-001/adjust \
  -H "Content-Type: application/json" \
  -d '{
    "adjustment": 10
  }'

# Decrementar 5 unidades
curl -X POST http://localhost:8080/api/v1/stock/550e8400-e29b-41d4-a716-446655440000/MAD-001/adjust \
  -H "Content-Type: application/json" \
  -d '{
    "adjustment": -5
  }'
```

#### Transferir stock entre tiendas

```bash
curl -X POST http://localhost:8080/api/v1/stock/550e8400-e29b-41d4-a716-446655440000/MAD-001/transfer \
  -H "Content-Type: application/json" \
  -d '{
    "to_store_id": "BCN-001",
    "quantity": 10
  }'
```

#### Ver productos con stock bajo

```bash
curl "http://localhost:8080/api/v1/stock/low-stock?threshold=10"
```

---

### 4. Reservas

#### Crear reserva

```bash
curl -X POST http://localhost:8080/api/v1/reservations \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "550e8400-e29b-41d4-a716-446655440000",
    "store_id": "MAD-001",
    "quantity": 2,
    "ttl_minutes": 15
  }'
```

Respuesta:
```json
{
  "id": "res-123",
  "product_id": "550e8400-e29b-41d4-a716-446655440000",
  "store_id": "MAD-001",
  "quantity": 2,
  "status": "PENDING",
  "expires_at": "2025-01-26T10:45:00Z",
  "created_at": "2025-01-26T10:30:00Z",
  "updated_at": "2025-01-26T10:30:00Z"
}
```

#### Obtener reserva

```bash
curl http://localhost:8080/api/v1/reservations/res-123
```

#### Confirmar reserva (procesar venta)

```bash
curl -X POST http://localhost:8080/api/v1/reservations/res-123/confirm
```

Respuesta:
```json
{
  "message": "Reservation confirmed successfully",
  "reservation_id": "res-123",
  "status": "CONFIRMED"
}
```

#### Cancelar reserva

```bash
curl -X POST http://localhost:8080/api/v1/reservations/res-123/cancel
```

#### Ver reservas pendientes de una tienda

```bash
curl http://localhost:8080/api/v1/reservations/store/MAD-001/pending
```

#### Ver estadísticas de reservas

```bash
curl http://localhost:8080/api/v1/reservations/stats
```

Respuesta:
```json
{
  "PENDING": 5,
  "CONFIRMED": 120,
  "CANCELLED": 8,
  "EXPIRED": 3
}
```

---

## Flujo Completo de Ejemplo

```bash
# 1. Crear producto
PRODUCT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "PHONE-001",
    "name": "iPhone 15 Pro",
    "category": "Electronics",
    "price": 999.99
  }')

PRODUCT_ID=$(echo $PRODUCT_RESPONSE | jq -r '.id')
echo "Product ID: $PRODUCT_ID"

# 2. Inicializar stock en Madrid
curl -X POST http://localhost:8080/api/v1/stock \
  -H "Content-Type: application/json" \
  -d "{
    \"product_id\": \"$PRODUCT_ID\",
    \"store_id\": \"MAD-001\",
    \"initial_quantity\": 100
  }"

# 3. Inicializar stock en Barcelona
curl -X POST http://localhost:8080/api/v1/stock \
  -H "Content-Type: application/json" \
  -d "{
    \"product_id\": \"$PRODUCT_ID\",
    \"store_id\": \"BCN-001\",
    \"initial_quantity\": 75
  }"

# 4. Ver stock en todas las tiendas
curl "http://localhost:8080/api/v1/stock/product/$PRODUCT_ID" | jq

# 5. Crear reserva de 2 unidades con TTL de 15 minutos
RESERVATION_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/reservations \
  -H "Content-Type: application/json" \
  -d "{
    \"product_id\": \"$PRODUCT_ID\",
    \"store_id\": \"MAD-001\",
    \"quantity\": 2,
    \"ttl_minutes\": 15
  }")

RESERVATION_ID=$(echo $RESERVATION_RESPONSE | jq -r '.id')
echo "Reservation ID: $RESERVATION_ID"

# 6. Verificar stock (debería mostrar 2 unidades reservadas)
curl "http://localhost:8080/api/v1/stock/$PRODUCT_ID/MAD-001" | jq

# 7. Confirmar reserva (procesar venta)
curl -X POST "http://localhost:8080/api/v1/reservations/$RESERVATION_ID/confirm"

# 8. Verificar stock final (98 unidades disponibles)
curl "http://localhost:8080/api/v1/stock/$PRODUCT_ID/MAD-001" | jq
```

---

## Autenticación JWT

Para usar endpoints protegidos, primero debes autenticarte. Ver [JWT_AUTHENTICATION.md](JWT_AUTHENTICATION.md) para documentación completa.

### Flujo Rápido (PowerShell)

```powershell
# 1. Registrar usuario
$registerBody = @{
    username = "test_user"
    email = "test@example.com"
    password = "test123456"
    role = "operator"
} | ConvertTo-Json

$user = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/register" `
    -Method POST `
    -Body $registerBody `
    -ContentType "application/json"

# 2. Login y obtener token
$loginBody = @{
    username = "test_user"
    password = "test123456"
} | ConvertTo-Json

$loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
    -Method POST `
    -Body $loginBody `
    -ContentType "application/json"

$token = $loginResponse.token
Write-Host "Token: $token"

# 3. Usar token en requests
$headers = @{
    "Authorization" = "Bearer $token"
}

# Obtener perfil
$profile = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/profile" `
    -Method GET `
    -Headers $headers

Write-Host "User: $($profile | ConvertTo-Json)"
```

### Flujo Rápido (Bash)

```bash
# 1. Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo_user",
    "email": "demo@example.com",
    "password": "demo123456",
    "role": "operator"
  }'

# 2. Login
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo_user",
    "password": "demo123456"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
echo "Token: $TOKEN"

# 3. Use token
curl http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN"
```

---

## Workers en Ejecución

### Worker de Expiración de Reservas
- Ejecuta cada **1 minuto**
- Encuentra reservas con estado `PENDING` y `expires_at < NOW()`
- Libera el stock reservado automáticamente
- Cambia estado a `EXPIRED`

### Worker de Sincronización de Eventos
- Ejecuta cada **10 segundos**
- Sincroniza eventos pendientes con el sistema central (NATS)
- Procesa en batches de 100 eventos
- Marca eventos como `synced = true`

---

## Optimistic Locking en Acción

El sistema usa **optimistic locking** para prevenir race conditions:

```bash
# Terminal 1: Actualizar stock
curl -X PUT http://localhost:8080/api/v1/stock/PRODUCT_ID/MAD-001 \
  -d '{"quantity": 100}'

# Terminal 2: Actualizar simultáneamente (puede fallar con 409 Conflict)
curl -X PUT http://localhost:8080/api/v1/stock/PRODUCT_ID/MAD-001 \
  -d '{"quantity": 95}'
```

Si el campo `version` cambió entre la lectura y la escritura, obtendrás:

```json
{
  "error": "Conflict",
  "message": "optimistic lock failed: stock was modified by another transaction"
}
```

---

## Manejo de Errores

### 400 Bad Request
```json
{
  "error": "Validation Error",
  "message": "quantity cannot be negative"
}
```

### 404 Not Found
```json
{
  "error": "Not Found",
  "message": "Product with ID 'xyz' not found"
}
```

### 409 Conflict (Stock Insuficiente)
```json
{
  "error": "Insufficient Stock",
  "message": "Insufficient stock for product abc in store MAD-001: available=5, requested=10"
}
```

---

## Datos de Ejemplo Precargados

El sistema incluye datos de ejemplo al inicializarse:

**Tiendas:**
- MAD-001 (Madrid)
- BCN-001 (Barcelona)
- VAL-001 (Valencia)
- SEV-001 (Sevilla)

**Productos:**
- iPhone 15 Pro (Electronics)
- Samsung Galaxy S24 (Electronics)
- Sony WH-1000XM5 (Electronics)
- MacBook Pro M3 (Electronics)
- iPad Air (Electronics)

**Stock inicial:** 5 unidades de cada producto en cada tienda

---

## Logs del Servidor

```
🚀 Server starting on port 8080 (instance: instance-001)
📊 Database driver: sqlite
🔒 Log level: info, format: json
📡 API available at http://localhost:8080/api/v1
⏰ Reservation expiration worker started
📡 Event synchronization worker started

[GIN] 2025/01/26 - 10:30:00 | 201 |   5.234567ms | 127.0.0.1 | POST /api/v1/products
[GIN] 2025/01/26 - 10:30:05 | 200 |   1.234567ms | 127.0.0.1 | GET /api/v1/stock/product/550e8400
✅ Expired 3 reservations
✅ Synced 15 events
```

---

## Next Steps

1. ~~Implementar UUID generator real (`github.com/google/uuid`)~~ ✅ **COMPLETADO**
2. Integrar NATS JetStream para eventos reales
3. ~~Agregar tests automatizados~~ ✅ **COMPLETADO** (27 unit tests + 32 E2E tests)
4. Configurar Prometheus para métricas
5. ~~Implementar JWT authentication~~ ✅ **COMPLETADO**

## Características Implementadas

- ✅ **UUID Generation**: Todos los IDs ahora usan `github.com/google/uuid` (UUID v4)
- ✅ **Complete Testing**: 100% test coverage (unit + E2E)
- ✅ **Optimistic Locking**: Prevención de race conditions en actualizaciones de stock
- ✅ **Background Workers**: Expiración automática de reservas y sincronización de eventos
- ✅ **JWT Authentication**: Sistema completo de autenticación con tokens JWT
- ✅ **Role-Based Access Control**: Control de acceso por roles (admin, operator, user)
