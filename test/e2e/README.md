# Tests End-to-End (E2E)

Este directorio contiene los tests de integración end-to-end para el sistema de inventario.

## Estructura

```
test/e2e/
├── setup_test.go         # Utilidades y helpers para tests E2E
├── health_test.go        # Tests del health check
├── products_test.go      # Tests de endpoints de productos
├── stock_test.go         # Tests de endpoints de stock
├── reservations_test.go  # Tests de endpoints de reservaciones
└── README.md            # Este archivo
```

## Requisitos Previos

1. **Servidor en ejecución**: Los tests E2E requieren que el servidor esté corriendo
2. **Base de datos limpia**: Se recomienda usar una base de datos de prueba
3. **Puerto 8080**: El servidor debe estar disponible en `http://localhost:8080`

## Ejecutar el Servidor para Tests

### Opción 1: Servidor con base de datos en memoria (SQLite)

```powershell
# Compilar
go build -o bin/inventory-api.exe cmd/api/main.go

# Ejecutar con SQLite en memoria
$env:DB_DRIVER="sqlite"; $env:DB_DSN=":memory:"; .\bin\inventory-api.exe
```

### Opción 2: Servidor con base de datos persistente

```powershell
# PostgreSQL
$env:DB_DRIVER="postgres"
$env:DB_DSN="postgresql://user:password@localhost:5432/inventory_test?sslmode=disable"
.\bin\inventory-api.exe

# O MySQL
$env:DB_DRIVER="mysql"
$env:DB_DSN="user:password@tcp(localhost:3306)/inventory_test?parseTime=true"
.\bin\inventory-api.exe
```

## Ejecutar los Tests E2E

### Ejecutar todos los tests

```powershell
cd test/e2e
go test -v
```

### Ejecutar un test específico

```powershell
# Test de health check
go test -v -run TestHealthCheck

# Test de productos
go test -v -run TestProductsE2E

# Test de stock
go test -v -run TestStockE2E

# Test de reservaciones
go test -v -run TestReservationsE2E
```

### Ejecutar tests con timeout

```powershell
go test -v -timeout 5m
```

### Ejecutar con más detalles

```powershell
go test -v -count=1  # Deshabilita cache
```

## Cobertura de Tests

### Health Check (1 endpoint)
- `GET /health` - Verificar estado del servidor

### Products (6 endpoints)
- `POST /api/v1/products` - Crear producto
- `GET /api/v1/products` - Listar productos
- `GET /api/v1/products/:id` - Obtener producto por ID
- `GET /api/v1/products/sku/:sku` - Obtener producto por SKU
- `PUT /api/v1/products/:id` - Actualizar producto
- `DELETE /api/v1/products/:id` - Eliminar producto

### Stock (9 endpoints)
- `POST /api/v1/stock` - Inicializar stock
- `GET /api/v1/stock/:productId/:storeId` - Obtener stock
- `GET /api/v1/stock/:productId/:storeId/availability` - Verificar disponibilidad
- `PUT /api/v1/stock/:productId/:storeId` - Actualizar stock
- `POST /api/v1/stock/:productId/:storeId/adjust` - Ajustar stock
- `POST /api/v1/stock/transfer` - Transferir stock entre tiendas
- `GET /api/v1/stock/product/:productId` - Stock por producto
- `GET /api/v1/stock/store/:storeId` - Stock por tienda
- `GET /api/v1/stock/low-stock` - Items con bajo stock

### Reservations (7 endpoints)
- `POST /api/v1/reservations` - Crear reserva
- `GET /api/v1/reservations/:id` - Obtener reserva
- `POST /api/v1/reservations/:id/confirm` - Confirmar reserva
- `POST /api/v1/reservations/:id/cancel` - Cancelar reserva
- `GET /api/v1/reservations/store/:storeId/pending` - Reservas pendientes
- `GET /api/v1/reservations/product/:productId/store/:storeId` - Reservas por producto
- `GET /api/v1/reservations/stats` - Estadísticas de reservas

## Estructura de un Test E2E

```go
func TestExample(t *testing.T) {
    client := NewTestClient()

    // 1. Preparar datos
    product := map[string]interface{}{
        "sku": RandomSKU("TEST"),
        "name": "Test Product",
        // ...
    }

    // 2. Ejecutar petición
    resp, body := client.POST(t, "/products", product)

    // 3. Verificar respuesta
    AssertStatusCode(t, http.StatusCreated, resp.StatusCode, body)
    
    // 4. Parsear y validar
    var result ProductResponse
    ParseJSON(t, body, &result)
    
    if result.Name != "Test Product" {
        t.Errorf("Expected name 'Test Product', got %s", result.Name)
    }

    // 5. Cleanup (opcional)
    client.DELETE(t, "/products/"+result.ID)
}
```

## Utilidades Disponibles

### TestClient
- `POST(t, path, body)` - Petición POST
- `GET(t, path)` - Petición GET
- `PUT(t, path, body)` - Petición PUT
- `DELETE(t, path)` - Petición DELETE

### Helpers
- `AssertStatusCode(t, expected, actual, body)` - Verificar código HTTP
- `AssertNoError(t, body)` - Verificar que no hay error en JSON
- `ParseJSON(t, body, v)` - Parsear respuesta JSON
- `WaitForServer(t, maxAttempts)` - Esperar a que el servidor esté listo
- `RandomSKU(prefix)` - Generar SKU único
- `RandomID(prefix)` - Generar ID único

## Escenarios de Test

### 1. Ciclo Completo de Producto
- Crear → Obtener → Actualizar → Listar → Eliminar

### 2. Gestión de Stock
- Inicializar → Actualizar → Ajustar → Transferir → Consultar

### 3. Flujo de Reserva
- Crear → Verificar stock reservado → Confirmar → Verificar stock actualizado

### 4. Cancelación de Reserva
- Crear → Verificar stock reservado → Cancelar → Verificar stock liberado

### 5. Validación de Errores
- Campos faltantes
- Recursos no encontrados
- Stock insuficiente
- IDs inválidos

## Troubleshooting

### Error: "connection refused"
- Verificar que el servidor esté corriendo en puerto 8080
- Ejecutar: `netstat -an | findstr 8080`

### Error: "timeout"
- Aumentar el timeout: `go test -v -timeout 10m`
- Verificar rendimiento del servidor

### Tests fallando aleatoriamente
- Puede haber race conditions
- Ejecutar con: `go test -v -race`

### Base de datos con datos antiguos
- Limpiar la base de datos antes de ejecutar tests
- O usar SQLite en memoria

## Mejores Prácticas

1. **Cleanup**: Siempre eliminar recursos creados en tests
2. **IDs únicos**: Usar `RandomSKU()` y `RandomID()` para evitar colisiones
3. **Verificación completa**: No solo verificar status code, validar el contenido
4. **Tests independientes**: Cada test debe poder ejecutarse solo
5. **Datos realistas**: Usar datos que reflejen casos de uso reales

## Integración Continua

Para ejecutar en CI/CD:

```yaml
# Ejemplo para GitHub Actions
- name: Run E2E Tests
  run: |
    # Iniciar servidor en background
    go build -o bin/server cmd/api/main.go
    DB_DRIVER=sqlite DB_DSN=:memory: ./bin/server &
    sleep 5
    
    # Ejecutar tests
    cd test/e2e
    go test -v -timeout 5m
    
    # Matar servidor
    pkill -f "./bin/server"
```

## Métricas

- **Total de tests**: ~15-20 escenarios
- **Cobertura**: 23 endpoints (100%)
- **Tiempo estimado**: 10-30 segundos
- **Dependencias**: Servidor en ejecución

## Contribuir

Al agregar nuevos endpoints:

1. Crear tests E2E correspondientes
2. Seguir la estructura existente
3. Documentar escenarios especiales
4. Actualizar este README
