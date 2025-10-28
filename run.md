# 🚀 Guía de Ejecución - Sistema de Inventario

Esta guía explica cómo ejecutar el proyecto en **cualquier sistema operativo** (Windows, Linux, macOS).

---

## 📋 Prerrequisitos

Antes de ejecutar el proyecto, asegúrate de tener instalado:

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

---

## 🔧 Instalación

### Paso 1: Extraer el Archivo ZIP

Extrae el contenido del archivo `inventory-system.zip` en cualquier directorio de tu sistema.

#### Windows
- Clic derecho en el archivo ZIP → "Extraer todo..."
- O usar PowerShell:
  ```powershell
  Expand-Archive -Path inventory-system.zip -DestinationPath C:\Projects\
  ```

#### Linux / macOS
```bash
unzip inventory-system.zip
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

### Opción 1: Ejecución Directa (Modo Desarrollo)

La forma más rápida de ejecutar el proyecto:

#### Windows (PowerShell)
```powershell
go run cmd/api/main.go
```

#### Linux / macOS (Bash)
```bash
go run cmd/api/main.go
```

**Salida esperada**:
```
2025/10/26 15:30:00 ✅ Connected to SQLite database: :memory:
2025/10/26 15:30:00 📊 Applying database migrations...
2025/10/26 15:30:00 ✅ Database migrations applied successfully
2025/10/26 15:30:00 ⏰ Reservation expiration worker started
2025/10/26 15:30:00 📡 Event synchronization worker started
2025/10/26 15:30:00 🚀 Server starting on port 8080 (instance: api-001)
2025/10/26 15:30:00 🔒 API Keys loaded: 3
2025/10/26 15:30:00 🌐 API available at http://localhost:8080/api/v1
```

---

### Opción 2: Compilar y Ejecutar (Modo Producción)

Compila el proyecto en un binario ejecutable:

#### Windows (PowerShell)
```powershell
# Compilar
go build -o bin/inventory-api.exe cmd/api/main.go

# Ejecutar
.\bin\inventory-api.exe
```

#### Linux / macOS (Bash)
```bash
# Compilar
go build -o bin/inventory-api cmd/api/main.go

# Ejecutar
./bin/inventory-api
```

**Ventajas**:
- ✅ Más rápido (sin recompilación)
- ✅ Binario portable
- ✅ Listo para despliegue

---

## ✅ Verificación

### 1. Health Check

Una vez el servidor esté corriendo, verifica que funciona:

#### Windows (PowerShell)
```powershell
Invoke-RestMethod -Uri http://localhost:8080/health
```

#### Linux / macOS (curl)
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

#### Windows (PowerShell)
```powershell
$headers = @{
    "Content-Type" = "application/json"
    "X-API-Key" = "dev-key-store-001"
}

$body = @{
    sku = "LAPTOP-001"
    name = "MacBook Pro 16"
    description = "Apple M3 Max, 32GB RAM"
    category = "electronics"
    price = 2499.99
} | ConvertTo-Json

Invoke-RestMethod -Uri http://localhost:8080/api/v1/products -Method POST -Headers $headers -Body $body
```

#### Linux / macOS (curl)
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

### 3. Ejecutar con PostgreSQL (Avanzado)

Si prefieres usar PostgreSQL:

1. Iniciar PostgreSQL con Docker:
   ```bash
   docker-compose up -d postgres
   ```

2. Configurar `.env`:
   ```bash
   DATABASE_DRIVER=postgres
   DATABASE_URL=postgres://inventory:inventory123@localhost:5432/inventory?sslmode=disable
   ```

3. Ejecutar:
   ```bash
   go run cmd/api/main.go
   ```

---

## 🛑 Detener el Servidor

### Modo Desarrollo (go run)
Presiona `Ctrl + C` en la terminal

### Modo Producción (binario)

#### Windows
```powershell
# Encontrar proceso
Get-Process -Name "inventory-api"

# Detener
Stop-Process -Name "inventory-api"
```

#### Linux / macOS
```bash
# Encontrar proceso
ps aux | grep inventory-api

# Detener
pkill inventory-api
```

---

## 🐛 Troubleshooting

### Error: "Port 8080 already in use"

**Causa**: Otro proceso está usando el puerto 8080

**Solución**:

#### Windows
```powershell
# Encontrar proceso en puerto 8080
netstat -ano | findstr :8080

# Matar proceso (usar PID del comando anterior)
taskkill /PID <PID> /F
```

#### Linux / macOS
```bash
# Encontrar proceso
lsof -i :8080

# Matar proceso
kill -9 <PID>
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
   echo $PATH  # Linux/macOS
   $env:PATH   # Windows PowerShell
   ```

---

### Error: "failed to connect to database"

**Causa**: Problema con la base de datos

**Solución**:
1. Verificar configuración en `.env`
2. Para SQLite, asegurar que el directorio existe
3. Para PostgreSQL, verificar que el contenedor esté corriendo:
   ```bash
   docker-compose ps
   ```

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

## 📚 Documentación Adicional

- **API Endpoints**: Ver `docs/QUICKSTART.md`
- **Arquitectura**: Ver `docs/ARCHITECTURE.md`
- **Autenticación**: Ver `docs/API_KEY_AUTH.md`
- **Ejemplos PowerShell**: Ver `docs/QUICK_TEST.md`

---

## 🎯 Resumen de Comandos

| Acción | Windows | Linux/macOS |
|--------|---------|-------------|
| **Extraer ZIP** | `Expand-Archive inventory-system.zip` | `unzip inventory-system.zip` |
| **Navegar** | `cd inventory-system` | `cd inventory-system` |
| **Dependencias** | `go mod download` | `go mod download` |
| **Ejecutar** | `go run cmd/api/main.go` | `go run cmd/api/main.go` |
| **Compilar** | `go build -o bin/inventory-api.exe cmd/api/main.go` | `go build -o bin/inventory-api cmd/api/main.go` |
| **Tests** | `go test ./... -v` | `go test ./... -v` |
| **Health Check** | `Invoke-RestMethod http://localhost:8080/health` | `curl http://localhost:8080/health` |
| **Detener** | `Ctrl + C` | `Ctrl + C` |

---

## ✅ Verificación de Ejecución Exitosa

Si ves estos mensajes, el proyecto está funcionando correctamente:

```
✅ Connected to SQLite database
✅ Database migrations applied successfully
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

# 4. Ejecutar
go run cmd/api/main.go

# 5. Verificar (en otra terminal)
curl http://localhost:8080/health
```

**Tiempo total**: ~3-5 minutos (incluyendo descarga de dependencias)

---

**¿Problemas?** Revisa la sección [Troubleshooting](#-troubleshooting) o consulta la documentación en `docs/`.
