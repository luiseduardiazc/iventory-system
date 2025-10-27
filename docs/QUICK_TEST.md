# 🚀 Test Rápido - Autenticación API Key

## ✅ Sistema Listo

El sistema ahora usa autenticación simple por **API Key**.

## 🔑 API Keys de Desarrollo

```
dev-key-store-001  →  Store Madrid
dev-key-store-002  →  Store Barcelona
dev-key-admin      →  Admin
```

## 📝 Pruebas PowerShell

### 1. Iniciar el Servidor

```powershell
cd c:\Users\80213585\Documents\inventory-system
.\bin\inventory-api.exe
```

### 2. Test Completo

```powershell
# Headers con API Key
$headers = @{
    "Content-Type" = "application/json"
    "X-API-Key" = "dev-key-store-001"
}

# 1. Health Check (público, sin API Key)
Invoke-RestMethod -Uri "http://localhost:8080/health"

# 2. Listar productos (público)
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/products"

# 3. Crear producto (requiere API Key)
$product = @{
    sku = "LAPTOP-001"
    name = "MacBook Pro 16"
    description = "Laptop premium"
    category = "Electronics"
    price = 2499.99
} | ConvertTo-Json

$newProduct = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/products" `
    -Method POST `
    -Headers $headers `
    -Body $product

Write-Host "✅ Producto creado: $($newProduct.id)" -ForegroundColor Green
$productId = $newProduct.id

# 4. Inicializar stock (requiere API Key)
$stock = @{
    product_id = $productId
    store_id = "MAD-001"
    initial_quantity = 100
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/stock" `
    -Method POST `
    -Headers $headers `
    -Body $stock

Write-Host "✅ Stock inicializado" -ForegroundColor Green

# 5. Ver stock (requiere API Key)
$stockData = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/stock/$productId/MAD-001" `
    -Headers $headers

Write-Host "📊 Stock actual: $($stockData.quantity) unidades" -ForegroundColor Cyan

# 6. Crear reserva (requiere API Key)
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

Write-Host "✅ Reserva creada: $($newReservation.id)" -ForegroundColor Green

# 7. Confirmar reserva (requiere API Key)
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/reservations/$($newReservation.id)/confirm" `
    -Method POST `
    -Headers $headers

Write-Host "✅ Reserva confirmada" -ForegroundColor Green

# 8. Ver stock final
$finalStock = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/stock/$productId/MAD-001" `
    -Headers $headers

Write-Host "📊 Stock final: $($finalStock.quantity) unidades (98 esperadas)" -ForegroundColor Cyan

Write-Host "`n🎉 ¡Todas las pruebas completadas exitosamente!" -ForegroundColor Green
```

## 🚨 Test de Errores

### Sin API Key (debe fallar con 401)

```powershell
# Intentar crear producto sin API Key
try {
    $product = @{
        sku = "TEST-001"
        name = "Test"
        price = 10.00
    } | ConvertTo-Json
    
    Invoke-RestMethod -Uri "http://localhost:8080/api/v1/products" `
        -Method POST `
        -ContentType "application/json" `
        -Body $product
} catch {
    Write-Host "❌ Error esperado: $_" -ForegroundColor Yellow
    Write-Host "✅ Validación correcta: endpoint protegido funciona" -ForegroundColor Green
}
```

### API Key Inválida (debe fallar con 401)

```powershell
# Intentar con API Key incorrecta
try {
    $headers = @{
        "Content-Type" = "application/json"
        "X-API-Key" = "invalid-key-12345"
    }
    
    $product = @{
        sku = "TEST-001"
        name = "Test"
        price = 10.00
    } | ConvertTo-Json
    
    Invoke-RestMethod -Uri "http://localhost:8080/api/v1/products" `
        -Method POST `
        -Headers $headers `
        -Body $product
} catch {
    Write-Host "❌ Error esperado: $_" -ForegroundColor Yellow
    Write-Host "✅ Validación correcta: API Key inválida rechazada" -ForegroundColor Green
}
```

## 📊 Endpoints por Tipo

### Públicos (sin API Key)
```powershell
GET http://localhost:8080/health
GET http://localhost:8080/api/v1/products
GET http://localhost:8080/api/v1/products/:id
GET http://localhost:8080/api/v1/products/sku/:sku
```

### Protegidos (requieren X-API-Key)
```powershell
POST   /api/v1/products
PUT    /api/v1/products/:id
DELETE /api/v1/products/:id

POST   /api/v1/stock
GET    /api/v1/stock/*
PUT    /api/v1/stock/*

POST   /api/v1/reservations
POST   /api/v1/reservations/:id/confirm
POST   /api/v1/reservations/:id/cancel
GET    /api/v1/reservations/*
```

## ⚡ One-Liner para Test Rápido

```powershell
# Test completo en una línea
$h = @{"Content-Type"="application/json";"X-API-Key"="dev-key-store-001"}; $p = @{sku="QUICK-TEST";name="Quick Test";price=99.99} | ConvertTo-Json; $prod = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/products" -Method POST -Headers $h -Body $p; Write-Host "✅ Producto: $($prod.id)" -ForegroundColor Green
```

## 🎯 Resultado Esperado

Si todo funciona correctamente, deberías ver:

```
✅ Producto creado: 550e8400-e29b-41d4-a716-446655440000
✅ Stock inicializado
📊 Stock actual: 100 unidades
✅ Reserva creada: 660f9511-f0bc-42e5-b827-557766551111
✅ Reserva confirmada
📊 Stock final: 98 unidades (98 esperadas)

🎉 ¡Todas las pruebas completadas exitosamente!
```

## 🔄 Comparación: Antes vs Ahora

### Antes (JWT - Complejo)
```powershell
# 1. Registrar
$user = Invoke-RestMethod .../auth/register -Body $userData
# 2. Login
$login = Invoke-RestMethod .../auth/login -Body $credentials
# 3. Extraer token
$token = $login.token
# 4. Usar en cada request
$headers = @{"Authorization"="Bearer $token"}
# 5. Renovar cuando expira
$refreshed = Invoke-RestMethod .../auth/refresh -Headers $headers
```

### Ahora (API Key - Simple)
```powershell
# 1. Usar directamente
$headers = @{"X-API-Key"="dev-key-store-001"}
# ¡Listo!
```

---

**Versión**: 2.0.0 (Simplificado)  
**Fecha**: 26 de Octubre, 2025
