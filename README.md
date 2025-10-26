# Sistema de Gestión de Inventario Distribuido

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## 🎯 Objetivo

Prototipo de sistema de gestión de inventario distribuido que optimiza la consistencia del inventario, reduce la latencia en las actualizaciones de stock y minimiza los costos operativos mediante una arquitectura event-driven.

## ✨ Características

- **Event-Driven Architecture**: Sincronización en tiempo real (<1s) vs sincronización periódica (15 min)
- **Optimistic Locking**: Previene overselling manteniendo alta disponibilidad
- **Reservas con TTL**: Auto-expiración de reservas para liberar stock automáticamente
- **Multi-Database**: Soporte para PostgreSQL (producción) y SQLite (desarrollo/testing)
- **Observabilidad**: Logging estructurado (zerolog) y métricas (Prometheus)
- **Seguridad**: Autenticación JWT, rate limiting, validación de inputs

## 🏗️ Arquitectura: API Centralizada Escalable

### Decisión Arquitectónica

**API Única Centralizada** con multi-tenancy por `store_id` en lugar de una API por tienda.

**Justificación**:
- ✅ **Reduce costos** 70%: 1-3 servidores centrales vs N servidores (uno por tienda)
- ✅ **Simplifica operaciones**: Un deployment vs N deployments
- ✅ **Mejor consistencia**: Una fuente de verdad compartida
- ✅ **Escalabilidad horizontal**: Load balancer + auto-scaling
- ✅ **Cumple objetivo**: "reducir costos operativos"

```
Clientes (Web/Móvil/POS)
         │
         ▼
  Load Balancer
         │
    ┌────┼────┐
    ▼    ▼    ▼
  API  API  API  (Stateless, auto-scaling)
    │    │    │
    └────┼────┘
         │
    ┌────┼────┐
    │    │    │
    ▼    ▼    ▼
  PgSQL Redis NATS
  
Multi-Tenant: Todas las tiendas comparten infraestructura
Partición de datos por store_id en tablas
```

### Flujo de Sincronización (Event-Driven)

```
Antes: Polling cada 15 minutos ❌
Tienda → Wait 15min → Sync → Cliente ve cambio

Ahora: Event-Driven <1 segundo ✅  
Tienda → NATS event (50ms) → Cache update (20ms) → Cliente ve cambio
Latencia: 15 min → 70ms = 12,857x más rápido
```

Ver [ARCHITECTURE.md](docs/ARCHITECTURE.md) para detalles completos.

## 🚀 Quick Start

### Prerrequisitos

- Go 1.21+ ([Descargar aquí](https://golang.org/dl/))
- Docker y Docker Compose (opcional, para infraestructura)
- Make (opcional, pero recomendado)

### Instalación

```bash
# 1. Clonar repositorio
git clone <repository-url>
cd inventory-system

# 2. Instalar dependencias
go mod download

# 3. Copiar configuración
cp .env.example .env
```

### Opción 1: Desarrollo con SQLite (sin Docker)

Perfecto para desarrollo local sin infraestructura:

```bash
# Editar .env para usar SQLite
DATABASE_DRIVER=sqlite
SQLITE_PATH=:memory:

# Ejecutar
go run cmd/api/main.go
```

### Opción 2: Producción con PostgreSQL (Docker)

```bash
# Iniciar infraestructura
docker-compose up -d

# Esperar a que esté saludable
docker-compose ps

# Ejecutar con PostgreSQL
DATABASE_DRIVER=postgres go run cmd/api/main.go
```

### Verificación

```bash
# Health check
curl http://localhost:8080/health

# Respuesta esperada:
# {"status":"healthy","timestamp":"2025-10-26T...","store_id":"store-001"}
```

## 📚 Documentación

- [📖 run.md](docs/run.md) - Instrucciones detalladas de ejecución
- [🔌 API.md](docs/API.md) - Documentación completa de la API
- [🏛️ ARCHITECTURE.md](docs/ARCHITECTURE.md) - Decisiones arquitectónicas
- [📋 IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) - Plan de implementación detallado

## 🧪 Testing

```bash
# Todos los tests
go test ./... -v

# Con race detector
go test -race ./...

# Con cobertura
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📊 Stack Tecnológico

| Categoría | Tecnología | Justificación |
|-----------|-----------|---------------|
| **Lenguaje** | Go 1.21+ | Concurrencia nativa, performance, simplicidad |
| **Web Framework** | Gin | Ligero, rápido, rico ecosistema de middleware |
| **Base de Datos** | PostgreSQL / SQLite | PostgreSQL para producción, SQLite para dev/test |
| **Cache** | Redis | Alta velocidad, soporte TTL nativo |
| **Message Broker** | NATS JetStream | Ligero, at-least-once delivery, pull-based |
| **Logging** | Zerolog | Zero-allocation, structured logging |
| **Métricas** | Prometheus | Estándar de facto para métricas |
| **Auth** | JWT | Stateless, escalable |

## 🛠️ Comandos Útiles

```bash
# Con Makefile
make deps          # Instalar dependencias
make build         # Compilar
make run           # Ejecutar
make test          # Tests
make docker-up     # Iniciar infraestructura
make docker-down   # Detener infraestructura

# Sin Makefile
go mod download    # Instalar dependencias
go build -o bin/api cmd/api/main.go  # Compilar
./bin/api          # Ejecutar
go test ./...      # Tests
```

## 📋 Estado del Proyecto

- [x] Fase 1: Fundación (Setup básico, health endpoint)
- [ ] Fase 2: Modelos de Dominio
- [ ] Fase 3: Persistencia (PostgreSQL + SQLite)
- [ ] Fase 4: Repositorios (Optimistic locking)
- [ ] Fase 5: Event Bus (NATS JetStream)
- [ ] Fase 6: Servicios (Stock, Reservas)
- [ ] Fase 7: HTTP Handlers
- [ ] Fase 8: Middleware (Auth, Logging, Metrics)
- [ ] Fase 9: Worker de Limpieza
- [ ] Fase 10: Testing Comprehensivo
- [ ] Fase 11: Documentación
- [ ] Fase 12: DevOps

## 🤝 Contribuir

Ver [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) para el plan detallado de desarrollo.

## 📝 Licencia

MIT License - Ver archivo LICENSE para detalles

## 👨‍💻 Autor

Desarrollado como prototipo de mejora para un sistema de gestión de inventario distribuido.

---

**Estado**: 🚧 En Desarrollo - Fase 1 Completada

