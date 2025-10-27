Objetivo General
Diseñar y crear un prototipo de mejora para un sistema de gestión de inventario existente que opera en un entorno distribuido.

El objetivo es optimizar la consistencia del inventario, reducir la latencia en las actualizaciones de stock y reducir los costos operativos, asegurando al mismo tiempo la seguridad y la observabilidad.

💡 Contexto Actual
Tu empresa mantiene un sistema de gestión de inventario para una cadena de tiendas minoristas.

Sincronización: Actualmente, cada tienda tiene una base de datos local que se sincroniza periódicamente (cada 15 minutos) con una base de datos central.

Problemas: Los clientes ven el stock en línea, pero las inconsistencias y la latencia en las actualizaciones han provocado problemas de experiencia de usuario y pérdidas de ventas debido a discrepancias de stock.

Arquitectura Actual: El sistema tiene un backend monolítico y el frontend es una aplicación web legacy.

📐 Diseño Técnico y Arquitectónico
1. Diseño de Arquitectura Distribuida
Proponer una arquitectura distribuida que aborde los problemas de consistencia y latencia del sistema actual. (Se requiere un cambio del modelo de sincronización basado en tiempo a un modelo más reactivo o de eventos).

Justificar la elección arquitectónica, explicando por qué es la más adecuada para este escenario distribuido.

2. Diseño de API
Diseñar la API para las operaciones clave de inventario (ej. Ver stock, Actualizar stock, Reservar producto).

Justificar las decisiones de diseño de la API (ej. verbos HTTP, endpoints, estructuras de datos), explicando por qué son adecuadas para el escenario distribuido propuesto.

3. Implementación de Backend
Persistencia: Simular la persistencia de datos utilizando archivos locales JSON/CSV o una base de datos en memoria (ej. SQLite, Base de datos H2) para representar el inventario. No se requiere una base de datos real.

Tolerancia a Fallos: Implementar los mecanismos básicos de tolerancia a fallos que se consideren necesarios.

Manejo de Concurrencia: Incluir la lógica para manejar las actualizaciones de stock en un entorno concurrente, priorizando la consistencia sobre la disponibilidad (o viceversa), justificando la elección realizada.

✅ Requisitos No Funcionales
Se dará especial consideración a las buenas prácticas en:

Manejo de errores

Documentación

Pruebas (Unitarias/Integración, etc.)

Cualquier otro aspecto no funcional relevante que elijas demostrar (ej. Logging, Métrica, Seguridad básica).

📄 Documentación y Estrategia
1. Estrategia Técnica
Detallar la pila tecnológica (stack) elegida para el backend (lenguaje, frameworks, herramientas de mensajería/eventos, etc.).

2. Documentación de Apoyo
Incluir un README o un Diagrama (opcional) que explique:

Diseño de la API y endpoints principales.

Instrucciones de configuración (setup).

Cualquier decisión arquitectónica clave tomada durante el desarrollo.

3. Ejecución del Proyecto
Debe contener un archivo run.md explicando cómo ejecutar el proyecto