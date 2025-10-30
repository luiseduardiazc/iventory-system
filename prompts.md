# 📝 Prompts Utilizados en la Construcción del Proyecto

Este documento recopila los prompts clave utilizados con IA (GitHub Copilot/Claude) durante el desarrollo del sistema de gestión de inventario distribuido.

## 🏗️ Fase 1: Arquitectura y Diseño

**Resultado**: Arquitectura con EventPublisher interface y doble persistencia (DB + Broker)

---

### Prompt: Estructura de Proyecto Go

```
Crea una estructura de proyecto Go siguiendo Clean Architecture con:
- Separación de capas (domain, service, repository, handler)
- Configuración por variables de entorno
- SQLite como base de datos
- Gin como framework web
- Soporte para testing con mocks
```

**Resultado**: Estructura de carpetas cmd/, internal/, pkg/, test/

---

## 📦 Fase 2: Implementación de Servicios

### Prompt: Servicio de Stock con Eventos

```
Implementa un StockService en Go que:
- Gestione operaciones CRUD de inventario
- Publique eventos (stock.created, stock.updated, stock.transferred)
- Use optimistic locking con campo version
- Inyecte EventPublisher interface para brokers intercambiables
- Incluya validaciones de negocio (cantidad >= 0)
```

**Resultado**: `internal/service/stock_service.go` con 400+ líneas

---

### Prompt: Sistema de Reservas con TTL

```
Crea un ReservationService que:
- Gestione reservas de stock con expiración automática
- Publique eventos (reservation.created, confirmed, cancelled, expired)
- Implemente un worker que expire reservas cada 10 segundos
- Use transacciones para garantizar consistencia
- Soporte cancelación manual y confirmación
```

**Resultado**: `internal/service/reservation_service.go` + worker background

---

### Prompt: EventPublisher Interface

```
Diseña una interface EventPublisher en Go que:
- Permita cambiar de Redis a Kafka sin modificar servicios
- Incluya métodos Publish(), PublishBatch(), Close()
- Tenga implementaciones: RedisPublisher, KafkaPublisher (futuro), MockPublisher, NoOpPublisher
- Soporte manejo de errores con retry
```

**Resultado**: `internal/infrastructure/event_publisher.go` con 4 implementaciones

---

## 🔄 Fase 3: Sistema de Resiliencia

### Prompt: Mecanismo de Retry Automático

```
Implementa un sistema de resiliencia que:
- Guarde eventos en DB con campo synced_at = NULL si el broker falla
- Tenga un EventSyncWorker que re-intente publicar eventos fallidos cada 10s
- Marque synced_at = NOW() cuando la publicación sea exitosa
- Garantice entrega eventual (at-least-once delivery)
- Sea observable (permitir consultar eventos pendientes)
```

**Resultado**: `internal/service/event_sync_service.go` + EventSyncWorker

---

### Prompt: Tests de Resiliencia

```
Genera tests unitarios para el sistema de retry que:
- Simulen fallos del broker con MockPublisher
- Validen que eventos fallidos se guardan en DB con synced_at = NULL
- Comprueben que el worker re-intenta publicaciones
- Verifiquen que synced_at se actualiza tras éxito
- Incluyan casos: fallo total, fallo parcial, éxito inmediato
```

**Resultado**: `test/unit/event_sync_retry_test.go` con 3 escenarios de prueba

---

## 🧪 Fase 4: Testing

### Prompt: Suite Completa de Tests Unitarios

```
Genera tests unitarios para todos los servicios con:
- Mocks in-memory para repositories y publishers
- Casos de prueba happy path y edge cases
- Validación de eventos publicados (tipo, payload)
- Tests de concurrencia para workers
- Cobertura de errores (DB down, broker down, validaciones)
```

**Resultado**: 60+ tests en `test/unit/` con 100% de éxito

---

### Prompt: Consolidación de Mocks

```
Consolida todos los mocks en test/mocks/ para:
- Eliminar duplicación de código
- Facilitar mantenimiento
- Permitir reutilización entre tests
- Incluir: MockRepository, MockPublisher, MockEventRepository
```

**Resultado**: `test/mocks/mock_*.go` con mocks reutilizables

---

## 📚 Fase 5: Documentación

### Prompt: README Completo

```
Crea un README.md profesional que incluya:
- Badges de versión Go y licencia
- Objetivo del proyecto
- Características principales con bullets
- Diagramas de arquitectura ASCII
- Quick Start con comandos
- Documentación de API endpoints con ejemplos JSON
- Stack tecnológico justificado
- Roadmap futuro
```

**Resultado**: README.md de 500+ líneas con 11 secciones

---

### Prompt: Documentación de Resiliencia

```
Genera documentación técnica que explique:
- Cómo funciona el sistema de retry automático
- Flujo completo desde operación de negocio hasta evento publicado
- Componentes involucrados (StockService, EventSyncService, Worker)
- Ejemplos de logs (éxito, fallo, retry)
- Consultas SQL para monitorear eventos pendientes
- Configuración de variables de entorno
```

**Resultado**: `docs/EVENT_SYNC_RESILIENCE.md` con 350 líneas

---

### Prompt: Diagramas Mermaid

```
Crea diagramas Mermaid para:
1. Arquitectura general del sistema (componentes + flujos)
2. Diagrama ejecutivo simplificado para negocio
3. Flujo de resiliencia (secuencia de retry)
4. Customer journey (value flow)
5. Diagrama de clases para EventPublisher

Usa colores, estilos y anotaciones para claridad
```

**Resultado**: 5 archivos `.mmd` + 2 imágenes PNG

---

### Prompt: Guía de Ejecución

```
Escribe una guía paso a paso (run.md) que cubra:
- Prerrequisitos (Go, Docker, Redis)
- Instalación y configuración
- Ejecución en modo desarrollo (con/sin Redis)
- Ejecución en modo producción
- Troubleshooting de errores comunes
- Ejemplos de llamadas a la API con curl
- Comandos útiles (tests, build, logs)
```

**Resultado**: `docs/run.md` con 7 secciones y 30+ comandos

---

## 🔧 Fase 6: Refactoring y Optimización

### Prompt: Migración a EventPublisher Interface

```
Refactoriza el código para:
- Extraer EventPublisher interface de servicios
- Modificar constructores para inyectar publisher
- Actualizar main.go con dependency injection
- Migrar todos los tests a usar MockPublisher
- Garantizar 0 regresiones (todos los tests pasan)
```

**Resultado**: Refactoring completo sin errores, 74/74 tests OK

---

### Prompt: Documentación de Stack Tecnológico

```
Actualiza la tabla de stack tecnológico para reflejar:
- Uso real de Redis (broker, no cache)
- SQLite para Event Sourcing
- EventSyncWorker para resiliencia
- Dependency Inversion Principle
- Goroutines para workers background
- Testing con mocks in-memory
```

**Resultado**: Tabla actualizada con 9 tecnologías justificadas

---

## 📊 Fase 7: Diagramas para Diferentes Audiencias

### Prompt: Diagramas por Audiencia

```
Genera 3 versiones de diagramas de arquitectura:

1. **Técnico** (Desarrolladores):
   - Componentes detallados (servicios, repos, publishers)
   - Flujos de datos y eventos
   - Tecnologías específicas (Redis, SQLite, Gin)

2. **Ejecutivo** (CEOs, VPs):
   - Solo 6 bloques principales
   - Lenguaje de negocio (sin términos técnicos)
   - Valor y beneficios destacados

3. **Producto** (PMs, UX):
   - Customer journey
   - Decisiones y timeouts
   - Flujo de valor end-to-end
```

**Resultado**: 3 diagramas con niveles de abstracción diferenciados

---

### Prompt: Guía de Audiencias

```
Crea una guía (DIAGRAMAS_POR_AUDIENCIA.md) que:
- Liste todos los diagramas disponibles
- Recomiende cuál usar según la audiencia
- Explique qué muestra cada diagrama
- Incluya tabla comparativa de complejidad
```

**Resultado**: Guía de 400 líneas con matriz de audiencias

---

## 📝 Prompts de Mejora Continua

### Prompt: Corrección de Documentación

```
Revisa toda la documentación y corrige:
- Referencias a Redis como "opcional" (es REQUIRED por defecto)
- Comentarios obsoletos sobre funcionalidades no implementadas
- Links rotos entre documentos
- Inconsistencias entre README y docs/run.md
```

**Resultado**: 7 archivos corregidos sin inconsistencias

---

### Prompt: Simplificación de Quick Start

```
Simplifica la sección Quick Start del README:
- Reducir de 50 a 30 líneas
- Referenciar docs/run.md para detalles
- Mostrar solo configuración esencial (MESSAGE_BROKER)
- Agregar 3 links explícitos a run.md
```

**Resultado**: Quick Start conciso con buena UX

## 💡 Lecciones sobre Prompting Efectivo

### ✅ Buenas Prácticas

1. **Ser específico con constraints**: "60+ tests", "3,000+ líneas", "cada 10 segundos"
2. **Nombrar patrones**: DIP, Event Sourcing, Optimistic Locking
3. **Listar casos de uso**: "stock.created, updated, transferred"
4. **Definir métricas de éxito**: "74/74 tests pasando", "0 regresiones"
5. **Especificar audiencia**: "técnico vs ejecutivo vs producto"

### ❌ Anti-Patrones

1. ❌ "Crea un sistema de inventario" (demasiado vago)
2. ❌ "Haz que funcione" (sin criterio de aceptación)
3. ❌ "Genera código" (sin especificar patrones ni estructura)
4. ❌ "Documenta esto" (sin definir audiencia ni formato)
5. ❌ "Arregla los tests" (sin explicar qué está roto)

---

## 🔄 Evolución de Prompts

### Iteración 1: Vago
```
"Crea un sistema de inventario con eventos"
```
**Resultado**: Código básico sin patrones claros

---

### Iteración 2: Específico
```
"Crea StockService que publique eventos a Redis usando patrón Publisher"
```
**Resultado**: Mejor, pero acoplado a Redis

---

### Iteración 3: Arquitectónicamente Sólido
```
"Crea StockService que inyecte EventPublisher interface, permitiendo 
cambiar de Redis a Kafka sin modificar código de negocio"
```
**Resultado**: ✅ Código desacoplado, extensible y mantenible

---
## 🎓 Conclusión

Los prompts más efectivos son aquellos que:

1. ✅ **Especifican patrones arquitectónicos** (DIP, Event Sourcing, etc.)
2. ✅ **Definen métricas de éxito** (tests pasando, cobertura, performance)
3. ✅ **Listan casos de uso concretos** (eventos, endpoints, flujos)
4. ✅ **Incluyen constraints** (tecnologías, límites, requisitos no funcionales)
5. ✅ **Definen audiencia** (técnico, ejecutivo, documentación)

> 💡 **Regla de Oro**: Un prompt bien estructurado ahorra 3-5 iteraciones de refinamiento.

---

## 🧠 Reflexión Final: IA como Herramienta, No como Reemplazo

### El Rol Real de la IA en Este Proyecto

Este proyecto demuestra que la **IA generativa es un acelerador poderoso**, pero no elimina la necesidad de:

1. **Conocimiento Técnico **
   - La IA generó código, pero yo decidí la arquitectura event-driven

2. **Pensamiento Crítico**
   - La IA propuso soluciones, yo evalué trade-offs
   - Detecté cuando el código generado tenía acoplamiento innecesario
   - Refiné prompts basándome en experiencias previas

3. **Visión Arquitectónica**
   - Identifiqué la necesidad de retry automático por experiencia
   - Planifiqué la extensibilidad (Redis → Kafka) anticipando evolución

### Lo Que la IA Hizo Excepcionalmente Bien

✅ **Código Boilerplate**: Repositories, handlers, DTOs (90% útil)  
✅ **Tests Unitarios**: Generó 60+ tests con mocks correctos (95% útil)  
✅ **Documentación**: README, diagramas, guías (85% útil)  
✅ **Refactoring Mecánico**: Renombrar, mover, reestructurar (100% útil)  

### Lo Que Requirió Intervención Humana Constante

⚠️ **Lógica de Negocio Compleja**: Validaciones, edge cases, transacciones  
⚠️ **Decisiones Arquitectónicas**: Patrones, abstracciones, extensibilidad  
⚠️ **Trade-offs**: Performance vs mantenibilidad, simplicidad vs flexibilidad  
⚠️ **Contexto del Dominio**: Qué eventos publicar, cuándo, por qué  

### La Metodología Que Funcionó

```
1. YO defino arquitectura y patrones
   ↓
2. IA genera implementación base
   ↓
3. YO reviso, refino y valido
   ↓
4. IA genera tests
   ↓
5. YO ejecuto tests y corrijo
   ↓
6. IA genera documentación
   ↓
7. YO ajusto para audiencia correcta
```
