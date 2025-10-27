# API Key Authentication - Simple Guide

## 🔑 Overview

Este sistema usa autenticación simple mediante **API Keys** enviadas en el header `X-API-Key`.

## 🚀 Quick Start

### 1. API Keys por Defecto (Desarrollo)

```bash
# Madrid Store
X-API-Key: dev-key-store-001

# Barcelona Store  
X-API-Key: dev-key-store-002

# Admin
X-API-Key: dev-key-admin
```

### 2. Ejemplo de Uso

#### PowerShell

```powershell
# Crear producto con API Key
$headers = @{
    "Content-Type" = "application/json"
    "X-API-Key" = "dev-key-store-001"
}

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
```

#### Bash

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
```

---

## 📋 Endpoints Protegidos vs Públicos

### Endpoints Públicos (sin API Key)

```bash
# Listar productos
GET /api/v1/products

# Ver producto por ID
GET /api/v1/products/:id

# Buscar por SKU
GET /api/v1/products/sku/:sku

# Health check
GET /health
```

### Endpoints Protegidos (requieren X-API-Key)

```bash
# Crear producto
POST /api/v1/products

# Actualizar producto
PUT /api/v1/products/:id

# Eliminar producto
DELETE /api/v1/products/:id

# Todas las operaciones de stock
POST /api/v1/stock
GET /api/v1/stock/*
PUT /api/v1/stock/*

# Todas las operaciones de reservas
POST /api/v1/reservations
POST /api/v1/reservations/:id/confirm
POST /api/v1/reservations/:id/cancel
```

---

## ⚙️ Configuración

### Variables de Entorno

```bash
# API Keys personalizadas (formato: key1:name1,key2:name2)
API_KEYS=production-key-madrid:Madrid Store,production-key-bcn:Barcelona Store,admin-secret-key:Admin
```

### Archivo .env

```env
# Development
API_KEYS=dev-key-store-001:Madrid,dev-key-store-002:Barcelona,dev-key-admin:Admin

# Production
API_KEYS=prod-abc123xyz:Madrid,prod-def456uvw:Barcelona,prod-admin-789:Admin
```

---

## 🔒 Generar API Keys Seguras

### PowerShell

```powershell
# Generar API Key aleatoria
$bytes = New-Object Byte[] 32
[Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
$apiKey = [Convert]::ToBase64String($bytes) -replace '[/+=]', ''
Write-Host "New API Key: $apiKey"
```

### Bash

```bash
# Generar API Key aleatoria
openssl rand -base64 32 | tr -d '/+=' | cut -c1-40
```

---

## 📊 Ejemplos Completos

### Flujo Completo de Producto

```powershell
$headers = @{
    "Content-Type" = "application/json"
    "X-API-Key" = "dev-key-store-001"
}

# 1. Crear producto
$product = @{
    sku = "PHONE-001"
    name = "iPhone 15 Pro"
    category = "Electronics"
    price = 999.99
} | ConvertTo-Json

$newProduct = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/products" `
    -Method POST `
    -Headers $headers `
    -Body $product

$productId = $newProduct.id
Write-Host "Product created: $productId"

# 2. Inicializar stock
$stock = @{
    product_id = $productId
    store_id = "MAD-001"
    initial_quantity = 100
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/stock" `
    -Method POST `
    -Headers $headers `
    -Body $stock

# 3. Crear reserva
$reservation = @{
    product_id = $productId
    store_id = "MAD-001"
    quantity = 2
    ttl_minutes = 15
} | ConvertTo-Json

$newReservation = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/reservations" `
    -Method POST `
    -Headers $headers `
    -Body $reservation

Write-Host "Reservation created: $($newReservation.id)"

# 4. Ver productos (sin API Key, es público)
$products = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/products"
$products | Format-Table
```

---

## 🚨 Errores Comunes

### 401 Unauthorized - Missing API Key

```json
{
  "error": "Unauthorized",
  "message": "missing X-API-Key header"
}
```

**Solución**: Agregar header `X-API-Key` con una key válida.

### 401 Unauthorized - Invalid API Key

```json
{
  "error": "Unauthorized",
  "message": "invalid API key"
}
```

**Solución**: Verificar que la API key sea correcta y esté en la lista de keys válidas.

---

## 🔐 Mejores Prácticas

### Desarrollo

```bash
# Usar keys descriptivas para identificar fácilmente
dev-key-store-001  # Tienda Madrid
dev-key-store-002  # Tienda Barcelona
dev-key-admin      # Admin
```

### Producción

```bash
# Usar keys aleatorias y largas
prod-a8f3b9c2d1e4f5g6h7i8j9k0l1m2n3o4
prod-b9g4c0d2e5f6g7h8i9j0k1l2m3n4o5p6
admin-x7y8z9a0b1c2d3e4f5g6h7i8j9k0l1m2
```

### Seguridad

1. **Nunca** commitear API keys en Git
2. Usar `.env` file y agregarlo a `.gitignore`
3. Rotar keys periódicamente (cada 3-6 meses)
4. Usar keys diferentes por ambiente (dev, staging, prod)
5. Usar HTTPS en producción
6. Implementar rate limiting por API key
7. Monitorear uso de cada API key
8. Revocar keys comprometidas inmediatamente

---

## 📈 Monitoreo

### Logs de Acceso

El sistema registra automáticamente cada request con API key:

```json
{
  "timestamp": "2025-10-26T10:30:00Z",
  "method": "POST",
  "path": "/api/v1/products",
  "store_name": "Madrid",
  "api_key": "dev-***-001",
  "status": 201,
  "latency": "15ms"
}
```

---

## 🔄 Migración desde JWT

Si anteriormente usabas JWT, simplemente:

1. Elimina header `Authorization: Bearer <token>`
2. Agrega header `X-API-Key: <your-api-key>`
3. Listo!

**Antes (JWT)**:
```bash
curl -H "Authorization: Bearer eyJhbGc..." http://localhost:8080/api/v1/products
```

**Ahora (API Key)**:
```bash
curl -H "X-API-Key: dev-key-store-001" http://localhost:8080/api/v1/products
```

---

## ✅ Checklist de Producción

- [ ] Generar API keys seguras (32+ caracteres)
- [ ] Configurar `API_KEYS` en variables de entorno
- [ ] Eliminar keys de desarrollo
- [ ] Documentar qué key usa cada tienda/servicio
- [ ] Configurar HTTPS
- [ ] Implementar rate limiting
- [ ] Configurar logging de accesos
- [ ] Establecer proceso de rotación de keys
- [ ] Definir proceso de revocación de keys comprometidas
- [ ] Monitorear uso anómalo de keys

---

**Ventajas de API Key vs JWT**:
- ✅ Más simple de implementar
- ✅ Más fácil de usar
- ✅ No requiere login/registro
- ✅ No expira (hasta que se revoque)
- ✅ Ideal para servicios y APIs machine-to-machine
- ✅ Menos overhead computacional

**Fecha**: 26 de Octubre, 2025  
**Versión**: 1.0.0
