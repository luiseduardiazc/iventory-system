# Test script for PostgreSQL database abstraction
Write-Host "🧪 Testing Database Abstraction Layer with PostgreSQL" -ForegroundColor Cyan
Write-Host ""

# Step 1: Build the application
Write-Host "📦 Building application..." -ForegroundColor Yellow
go build -o bin/inventory-api.exe cmd/api/main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Build successful" -ForegroundColor Green
Write-Host ""

# Step 2: Test placeholder conversion
Write-Host "🔍 Testing placeholder conversion logic..." -ForegroundColor Yellow
go test ./internal/database/... -v -run TestPlaceholderConversion
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Placeholder conversion test failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Placeholder conversion tests passed" -ForegroundColor Green
Write-Host ""

# Step 3: Verify PostgreSQL is running
Write-Host "🐘 Checking PostgreSQL container..." -ForegroundColor Yellow
$pgStatus = docker ps --filter "name=inventory-postgres" --format "{{.Status}}"
if ($pgStatus -notlike "*healthy*") {
    Write-Host "❌ PostgreSQL container is not healthy!" -ForegroundColor Red
    Write-Host "   Status: $pgStatus" -ForegroundColor Yellow
    Write-Host "   Run: docker-compose up -d" -ForegroundColor Yellow
    exit 1
}
Write-Host "✅ PostgreSQL is healthy" -ForegroundColor Green
Write-Host ""

# Step 4: Start server with PostgreSQL
Write-Host "🚀 Starting server with PostgreSQL..." -ForegroundColor Yellow
$env:DATABASE_DRIVER="postgres"
$env:POSTGRES_HOST="localhost"
$env:POSTGRES_PORT="5432"
$env:POSTGRES_USER="inventory"
$env:POSTGRES_PASSWORD="inventory123"
$env:POSTGRES_DB="inventory_test_db"
$env:MESSAGE_BROKER="none"
$env:API_KEYS="test-key-12345"
$env:SERVER_PORT="8082"

# Start server in background
$job = Start-Job -ScriptBlock {
    param($binPath, $envVars)
    foreach ($key in $envVars.Keys) {
        Set-Item -Path "env:$key" -Value $envVars[$key]
    }
    & $binPath
} -ArgumentList @(
    (Resolve-Path ".\bin\inventory-api.exe"),
    @{
        DATABASE_DRIVER="postgres"
        POSTGRES_HOST="localhost"
        POSTGRES_PORT="5432"
        POSTGRES_USER="inventory"
        POSTGRES_PASSWORD="inventory123"
        POSTGRES_DB="inventory_test_db"
        MESSAGE_BROKER="none"
        API_KEYS="test-key-12345"
        SERVER_PORT="8082"
    }
)

# Wait for server to start
Write-Host "⏳ Waiting for server to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 6

# Step 5: Test health endpoint
Write-Host "🏥 Testing health endpoint..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8082/health" -ErrorAction Stop
    Write-Host "✅ Health check passed" -ForegroundColor Green
    Write-Host "   Database: $($health.database)" -ForegroundColor Cyan
    Write-Host "   Driver: $($health.db_driver)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Health check failed: $_" -ForegroundColor Red
    Stop-Job -Job $job
    Remove-Job -Job $job
    exit 1
}
Write-Host ""

# Step 6: Test product creation (verifies placeholder conversion)
Write-Host "📝 Testing product creation with PostgreSQL..." -ForegroundColor Yellow
try {
    $productData = @{
        sku = "TEST-PG-001"
        name = "PostgreSQL Test Product"
        description = "Testing database abstraction layer"
        category = "test"
        price = 99.99
    } | ConvertTo-Json

    $product = Invoke-RestMethod `
        -Uri "http://localhost:8082/api/v1/products" `
        -Method POST `
        -Body $productData `
        -ContentType "application/json" `
        -Headers @{"X-API-Key"="test-key-12345"} `
        -ErrorAction Stop

    Write-Host "✅ Product created successfully" -ForegroundColor Green
    Write-Host "   ID: $($product.id)" -ForegroundColor Cyan
    Write-Host "   SKU: $($product.sku)" -ForegroundColor Cyan
    Write-Host "   Name: $($product.name)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Product creation failed: $_" -ForegroundColor Red
    Stop-Job -Job $job
    Remove-Job -Job $job
    exit 1
}
Write-Host ""

# Step 7: Retrieve product (verifies SELECT queries)
Write-Host "🔍 Testing product retrieval..." -ForegroundColor Yellow
try {
    $retrieved = Invoke-RestMethod `
        -Uri "http://localhost:8082/api/v1/products/$($product.id)" `
        -ErrorAction Stop

    if ($retrieved.sku -eq "TEST-PG-001") {
        Write-Host "✅ Product retrieved successfully" -ForegroundColor Green
    } else {
        Write-Host "❌ Retrieved product doesn't match!" -ForegroundColor Red
        Stop-Job -Job $job
        Remove-Job -Job $job
        exit 1
    }
} catch {
    Write-Host "❌ Product retrieval failed: $_" -ForegroundColor Red
    Stop-Job -Job $job
    Remove-Job -Job $job
    exit 1
}
Write-Host ""

# Cleanup
Write-Host "🧹 Cleaning up..." -ForegroundColor Yellow
Stop-Job -Job $job -ErrorAction SilentlyContinue
Remove-Job -Job $job -ErrorAction SilentlyContinue
Write-Host ""

# Success!
Write-Host "═══════════════════════════════════════════" -ForegroundColor Green
Write-Host "✨ All tests passed successfully!" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════" -ForegroundColor Green
Write-Host ""
Write-Host "Database Abstraction Layer Status:" -ForegroundColor Cyan
Write-Host "  ✅ Placeholder conversion (? → \$1, \$2, \$3)" -ForegroundColor Green
Write-Host "  ✅ PostgreSQL connection" -ForegroundColor Green
Write-Host "  ✅ INSERT queries" -ForegroundColor Green
Write-Host "  ✅ SELECT queries" -ForegroundColor Green
Write-Host "  ✅ Multi-database support ready" -ForegroundColor Green
Write-Host ""
