# Simplificación: Autenticación por API Key

## 📋 Cambios Realizados

Hemos simplificado la autenticación del sistema, eliminando la complejidad de JWT y usando **API Keys simples**.

## 🔑 Cómo Funciona

### Header Requerido

Todos los endpoints protegidos ahora requieren el header:

```
X-API-Key: <tu-api-key>
```

### API Keys por Defecto (Desarrollo)

```bash
# Tienda Madrid
X-API-Key: dev-key-store-001

# Tienda Barcelona  
X-API-Key: dev-key-store-002

# Admin
X-API-Key: dev-key-admin
```

## 🚀 Ejemplos de Uso

### PowerShell

```powershell
# Headers con API Key
$headers = @{
    "Content-Type" = "application/json"
    "X-API-Key" = "dev-key-store-001"
}

# Crear producto
$product = @{
    sku = "LAPTOP-001"
    name = "MacBook Pro 16"
    category = "Electronics"
    price = 2499.99
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/products" `
    -Method POST `
    -Headers $headers `
    -Body $product

# Inicializar stock
$stock = @{
    product_id = "uuid-del-producto"
    store_id = "MAD-001"
    initial_quantity = 100
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/stock" `
    -Method POST `
    -Headers $headers `
    -Body $stock

# Crear reserva
$reservation = @{
    product_id = "uuid-del-producto"
    store_id = "MAD-001"
    quantity = 2
    ttl_minutes = 15
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/reservations" `
    -Method POST `
    -Headers $headers `
    -Body $reservation
```

### Bash

```bash
# Crear producto con API Key
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-store-001" \
  -d '{
    "sku": "LAPTOP-001",
    "name": "MacBook Pro 16",
    "category": "Electronics",
    "price": 2499.99
  }'

# Inicializar stock
curl -X POST http://localhost:8080/api/v1/stock \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-store-001" \
  -d '{
    "product_id": "uuid-del-producto",
    "store_id": "MAD-001",
    "initial_quantity": 100
  }'
```

## 📊 Endpoints

### Públicos (sin API Key)

```bash
GET /health
GET /api/v1/products
GET /api/v1/products/:id
GET /api/v1/products/sku/:sku
```

### Protegidos (requieren X-API-Key)

```bash
# Products
POST   /api/v1/products
PUT    /api/v1/products/:id
DELETE /api/v1/products/:id

# Stock  
POST   /api/v1/stock
GET    /api/v1/stock/*
PUT    /api/v1/stock/*

# Reservations
POST   /api/v1/reservations
POST   /api/v1/reservations/:id/confirm
POST   /api/v1/reservations/:id/cancel
```

## ⚙️ Configuración de Producción

### Variables de Entorno

```bash
# Formato: key1:name1,key2:name2
API_KEYS=prod-madrid-abc123:Madrid,prod-bcn-def456:Barcelona,admin-xyz789:Admin
```

### Archivo .env

```env
API_KEYS=prod-madrid-abc123:Madrid,prod-bcn-def456:Barcelona
```

### Generar API Keys Seguras

```powershell
# PowerShell
[Convert]::ToBase64String((1..32 | ForEach-Object {Get-Random -Max 256})) -replace '[/+=]', ''
```

```bash
# Bash
openssl rand -base64 32 | tr -d '/+='
```

## 🔒 Ventajas vs JWT

- ✅ **Más simple**: No requiere registro ni login
- ✅ **Más rápido**: No hay procesamiento de tokens
- ✅ **Más fácil**: Solo agregar un header
- ✅ **Ideal para APIs**: Machine-to-machine communication
- ✅ **Sin expiración**: Keys válidas hasta que se revoquen

## 🚨 Errores Comunes

### 401 - Missing API Key

```json
{
  "error": "Unauthorized",
  "message": "missing X-API-Key header"
}
```

**Solución**: Agregar `-H "X-API-Key: dev-key-store-001"`

### 401 - Invalid API Key

```json
{
  "error": "Unauthorized",
  "message": "invalid API key"
}
```

**Solución**: Verificar que la key sea correcta

## 📝 Implementación Técnica

### Middleware Creado

```go
// internal/middleware/apikey.go
func APIKeyAuth(validAPIKeys map[string]string) gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-API-Key")
        
        if apiKey == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Unauthorized",
                "message": "missing X-API-Key header",
            })
            c.Abort()
            return
        }

        storeName, valid := validAPIKeys[apiKey]
        if !valid {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Unauthorized",
                "message": "invalid API key",
            })
            c.Abort()
            return
        }

        c.Set("api_key", apiKey)
        c.Set("store_name", storeName)
        c.Next()
    }
}
```

### Uso en Rutas

```go
// Proteger endpoints con API Key
products.POST("", middleware.APIKeyAuth(cfg.APIKeys), productHandler.CreateProduct)
stock.POST("", middleware.APIKeyAuth(cfg.APIKeys), stockHandler.InitializeStock)
```

## ✅ Migración desde JWT (si aplicaba)

**Antes**:
```bash
curl -H "Authorization: Bearer eyJhbGc..." http://localhost:8080/api/v1/products
```

**Ahora**:
```bash
curl -H "X-API-Key: dev-key-store-001" http://localhost:8080/api/v1/products
```

## 🎯 Próximos Pasos

1. ✅ Sistema simplificado con API Key
2. ⏭️ Compilar y probar
3. ⏭️ Actualizar tests E2E
4. ⏭️ Documentar en README principal

---

**Fecha**: 26 de Octubre, 2025  
**Versión**: 2.0.0 (Simplificado)
