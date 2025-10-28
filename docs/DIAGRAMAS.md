# 📊 Diagramas de Arquitectura del Sistema de Inventario

Este directorio contiene los diagramas de arquitectura del sistema en formato **Mermaid**, listos para importar en [MermaidChart.com](https://www.mermaidchart.com/).

## 📁 Diagramas Disponibles

### 1. **architecture-diagram.mmd** - Diagrama General de Arquitectura
**Tipo**: Graph (Componentes y Relaciones)  
**Descripción**: Vista completa del sistema mostrando todas las capas, componentes y sus relaciones.

**Muestra**:
- Client Layer (API Clients)
- API Gateway (Gin HTTP Server)
- Middleware Layer (Auth, Logger, Rate Limiter)
- Handler Layer (Product, Stock, Reservation)
- Service Layer (Business Logic)
- Event Publishing (Interface + Implementations)
- Repository Layer (Data Access)
- Database Layer (SQLite + Tables)
- Message Broker (Redis Streams)
- Background Workers (Expiration, Sync)

**Colores**:
- 🟢 Verde: Implementado
- 🟡 Amarillo: Planificado
- 🔵 Azul: Infraestructura
- 🟣 Púrpura: Base de datos
- 🔴 Rojo: Workers

---

### 2. **event-flow-diagram.mmd** - Diagrama de Flujo de Eventos
**Tipo**: Sequence Diagram  
**Descripción**: Flujo detallado de eventos en el sistema, mostrando la interacción entre componentes en escenarios de negocio.

**Escenarios cubiertos**:
1. ✅ **Create Stock** (Inventario Inicial)
   - Request → Handler → Service → Repository → Database
   - Event persistence (DB + Redis)
   - Response flow

2. ✅ **Create Reservation** (Reservar Stock)
   - Optimistic locking en acción
   - Check de disponibilidad
   - Actualización de stock reservado
   - Publicación de eventos

3. ✅ **Confirm Reservation** (Completar Compra)
   - Validaciones (status, expiration)
   - Reducción de inventario
   - Liberación de reserva
   - Double persistence

4. ✅ **Background Worker - Expiration**
   - Auto-cancelación de reservas expiradas
   - Liberación automática de stock
   - Event publishing

**Notas especiales**:
- Muestra la estrategia de doble persistencia (DB + Broker)
- Detalla el optimistic locking con versiones
- Ilustra el patrón de eventos de dominio

---

### 3. **layered-architecture-diagram.mmd** - Arquitectura en Capas
**Tipo**: Graph (Layered)  
**Descripción**: Arquitectura del sistema organizada en 7 capas, mostrando la separación de responsabilidades.

**Capas**:
1. **Layer 1 - Presentation**: HTTP Endpoints, Middleware, DTOs
2. **Layer 2 - Application**: Handlers (Product, Stock, Reservation, Health)
3. **Layer 3 - Domain/Business**: Services con lógica de negocio
4. **Layer 4 - Infrastructure**: Event Publishing, Background Workers
5. **Layer 5 - Data Access**: Repositories con CRUD
6. **Layer 6 - Database**: SQLite + Tablas
7. **Layer 7 - External Systems**: Redis, Kafka (futuro), Config

**Principios**:
- ✅ Dependency Rule: Capas internas no dependen de capas externas
- ✅ Dependency Inversion: Services dependen de interfaces
- ✅ Single Responsibility: Cada capa tiene una responsabilidad clara

**Colores**:
- 🟢 Presentation Layer
- 🔵 Application Layer
- 🟡 Domain Layer
- 🌸 Infrastructure Layer
- 🟣 Data Access Layer
- 🐚 Database Layer
- 🟠 External Systems

---

### 4. **event-publisher-class-diagram.mmd** - Diagrama de Clases (Event Publisher)
**Tipo**: Class Diagram  
**Descripción**: Diseño detallado del patrón Event Publisher implementado en el sistema.

**Componentes**:

**Interface**:
- `EventPublisher` (abstracción)
  - `Publish(event Event) error`
  - `PublishBatch(events []Event) error`
  - `Close() error`

**Implementaciones**:
- ✅ `RedisPublisher` - Usa Redis Streams (XADD)
- 🔜 `KafkaPublisher` - Usa Kafka Producer API (planificado)
- ✅ `MockPublisher` - Captura eventos en memoria (testing)
- ✅ `NoOpPublisher` - No-op pattern (desarrollo)

**Servicios que lo usan**:
- `StockService`
- `ReservationService`
- `ProductService`

**Repositorio relacionado**:
- `EventRepository` - Persiste eventos en SQLite

**Sistemas externos**:
- `RedisStreams` - Stream real-time
- `KafkaTopic` - Topic para eventos (futuro)
- `SQLiteDatabase` - Tabla de eventos

**Patrón demostrado**:
- ✅ Dependency Inversion Principle (SOLID)
- ✅ Strategy Pattern
- ✅ Adapter Pattern

---

### 5. **deployment-diagram.mmd** - Diagrama de Despliegue
**Tipo**: Graph (Deployment)  
**Descripción**: Opciones de despliegue y configuración del sistema en diferentes entornos.

**Entornos**:

1. **Development Environment**
   - Source code + Go compiler
   - Binary compilation (19 MB)
   - Docker Compose (Redis container)
   - SQLite file (./inventory.db)
   - .env configuration

2. **Testing Environment**
   - Test suite (74 tests)
   - Mock publisher (in-memory)
   - SQLite :memory: (ephemeral)

3. **Production Environment**
   - Application server (multiple instances)
   - SQLite persistent volume
   - Redis cluster (master + replicas)
   - Monitoring & observability

**Deployment Options**:
- Option 1: Single binary + SQLite + Redis
- Option 2: Docker container
- Option 3: Kubernetes (Deployment + Service + ConfigMap)

**Configuration Files**:
- `docker-compose.yml`
- `.env.example`
- `migrations/001_initial_schema.sql`

---

## 🚀 Cómo Usar estos Diagramas

### Opción 1: MermaidChart.com (Recomendado)

1. Ve a [https://www.mermaidchart.com/](https://www.mermaidchart.com/)
2. Crea una cuenta gratuita o inicia sesión
3. Click en "New Diagram" o "Import"
4. Selecciona "From Text"
5. Copia el contenido de cualquier archivo `.mmd`
6. Pega en el editor
7. El diagrama se renderizará automáticamente
8. Puedes editar, exportar (PNG, SVG, PDF) y compartir

### Opción 2: Visual Studio Code

1. Instala la extensión "Markdown Preview Mermaid Support"
2. Abre cualquier archivo `.mmd`
3. Click derecho → "Open Preview"
4. El diagrama se mostrará renderizado

### Opción 3: GitHub (Automático)

GitHub renderiza automáticamente los diagramas Mermaid en archivos `.md`:

```markdown
```mermaid
[contenido del diagrama]
\```
```

### Opción 4: Mermaid Live Editor

1. Ve a [https://mermaid.live/](https://mermaid.live/)
2. Pega el contenido del archivo `.mmd`
3. Visualiza y exporta

---

## 📋 Guía Rápida de Archivos

| Archivo | Tipo | Propósito | Recomendado para |
|---------|------|-----------|------------------|
| `architecture-diagram.mmd` | Graph | Vista general completa | Presentaciones, documentación |
| `event-flow-diagram.mmd` | Sequence | Flujos de negocio detallados | Onboarding de desarrolladores |
| `layered-architecture-diagram.mmd` | Graph | Estructura en capas | Diseño de software, reviews |
| `event-publisher-class-diagram.mmd` | Class | Patrón de diseño | Documentación técnica |
| `deployment-diagram.mmd` | Graph | Infraestructura y despliegue | DevOps, deployment |

---

## 🎨 Personalización

Todos los diagramas usan clases CSS para colores. Puedes personalizar los estilos:

```mermaid
classDef implemented fill:#90EE90,stroke:#006400,stroke-width:2px
classDef planned fill:#FFD700,stroke:#FF8C00,stroke-width:2px
class REDIS_PUB,MOCK_PUB implemented
class KAFKA_PUB planned
```

---

## 📚 Referencias

- **Mermaid Documentation**: https://mermaid.js.org/
- **MermaidChart**: https://www.mermaidchart.com/
- **Mermaid Live Editor**: https://mermaid.live/
- **GitHub Mermaid Support**: https://github.blog/2022-02-14-include-diagrams-markdown-files-mermaid/

---

## ✅ Validación

Todos los diagramas han sido validados en:
- ✅ MermaidChart.com
- ✅ Mermaid Live Editor
- ✅ VS Code con extensión Mermaid
- ✅ Sintaxis Mermaid 10.0+

---

## 🔄 Actualización

Los diagramas deben actualizarse cuando:
- Se añaden nuevos componentes al sistema
- Cambia la arquitectura de capas
- Se implementan features planificados (Kafka, Metrics)
- Se modifican los flujos de negocio

**Última actualización**: Octubre 27, 2025  
**Versión del sistema**: 1.0.0  
**Estado**: Producción-Ready
