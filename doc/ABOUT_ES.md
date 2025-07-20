# Acerca del Proyecto budva43

## Descripción General

**budva43** es un sistema inteligente de reenvío automático de mensajes de Telegram escrito en Go. Este proyecto implementa un enfoque empresarial utilizando principios UNIX-way y Clean Architecture para crear resúmenes temáticos a partir de mensajes de varios canales y grupos.

## Funcionalidades Principales

### Reenvío Automático de Mensajes
- **Forward** — Envío de mensajes preservando el autor original
- **Send Copy** — Creación de copias de mensajes sin mostrar la fuente original
- **Soporte de Álbumes Multimedia** — Procesamiento de imágenes y archivos agrupados

### Sistema de Filtrado
- **Filtros de Exclusión** — Expresiones regulares para bloquear contenido no deseado
- **Filtros de Inclusión** — Reglas para permitir solo mensajes relevantes
- **Filtros de Subcadenas** — Filtrado preciso de subcadenas usando grupos de regex
- **Respuestas Automáticas** — Respuestas automáticas a mensajes con teclados
- **Chats Especiales** — Envío automático de mensajes filtrados a canales check/other

### Transformación de Contenido
- **Reemplazo de Enlaces** — Sustitución automática de enlaces a mensajes propios en chats objetivo
- **Eliminación de Enlaces Externos** — Limpieza de enlaces a fuentes externas
- **Reemplazo de Fragmentos de Texto** — Transformaciones de texto personalizables para diferentes destinatarios
- **Firma de Fuente** — Visualización de la fuente del mensaje original
- **Generación de Enlaces de Fuente** — Adición automática de enlaces al mensaje original

### Gestión del Ciclo de Vida de Mensajes
- **Copy Once** — Envío único sin sincronización durante ediciones
- **Indelible** — Protección contra eliminación de mensajes cuando se borra el original
- **Sincronización de Ediciones** — Actualización automática de mensajes copiados cuando cambia el original
- **Sincronización de Eliminaciones** — Eliminación de copias cuando se elimina el mensaje fuente

### Funciones Adicionales
- **Limitación de Velocidad** — Control de velocidad de envío para prevenir bloqueos
- **Procesamiento de Mensajes del Sistema** — Eliminación automática de notificaciones de servicio

## Arquitectura

### Estructura de Microservicios
El proyecto se divide en dos servicios principales:

#### Engine (cmd/engine)
- **Propósito**: Ejecutar el reenvío de mensajes
- **Restricción**: Prohibido enviar nuevos mensajes a chats de envío
- **Componentes**: Manejadores de actualizaciones de Telegram, servicios de reenvío y filtrado

#### Facade (cmd/facade)
- **Propósito**: Proporcionar APIs (GraphQL, gRPC, REST)
- **Funcionalidad**: Acceso completo a funciones de envío de mensajes
- **Interfaces**: Interfaz web, API gRPC, interfaz de terminal

### Arquitectura por Capas

```
Transport Layer    → HTTP, gRPC, Terminal, Telegram Bot API
Service Layer      → Lógica de negocio, procesamiento de reglas de reenvío
Repository Layer   → TDLib, Storage, Queue
Domain Layer       → Modelos de datos, reglas de reenvío
```

### Patrones de Diseño
- **Clean Architecture** — Separación clara de capas de responsabilidad
- **Inyección de Dependencias** — Sistema personalizado de inyección de dependencias
- **Patrón Repository** — Abstracción de acceso a datos
- **Patrón Observer** — Procesamiento de actualizaciones de Telegram

## Stack Tecnológico

### Tecnologías Principales
- **Go 1.24** — Lenguaje principal de desarrollo con soporte para genéricos
- **TDLib** — Biblioteca oficial de Telegram para aplicaciones cliente
- **BadgerDB** — Base de datos NoSQL embebida

### APIs y Transporte
- **gRPC** — API de alto rendimiento para integraciones
- **GraphQL** — API flexible para clientes web
- **REST** — API HTTP clásica
- **Telegram Client API** — Interacción directa con Telegram
- **Interfaz de Terminal** — Interfaz interactiva de terminal

### Desarrollo y Pruebas
- **Docker y DevContainers** — Contenerización del desarrollo
- **Testcontainers** — Pruebas de integración (incluyendo pruebas de conexión Redis)
- **Mockery** — Generación de objetos mock
- **Godog (BDD)** — Pruebas dirigidas por comportamiento
- **GitHub Actions CI** — Integración continua automatizada

### Monitoreo y Observabilidad
- **Logging Estructurado** — Registro estructurado usando slog
- **Grafana + Loki** — Logging centralizado y monitoreo
- **pplog** — Logs JSON legibles para desarrollo
- **spylog** — Interceptación de logs en pruebas

### Herramientas de Build y Desarrollo
- **Task** — Alternativa a Make para automatización de tareas
- **golangci-lint** — Verificación integral de calidad de código
- **Linters Personalizados** — "error-log-or-return" y "unused-interface-methods"
- **protobuf** — Generación de interfaces gRPC
- **jq** — Visualización de logs en tiempo real con filtrado
- **EditorConfig** — Consistencia en configuraciones de editor

## Principios de Desarrollo

### Principios Arquitectónicos
- **SOLID** — Aplicación de los 5 principios OOP
- **DRY** — Evitar duplicación de código sin fanatismo
- **KISS** — Preferir soluciones simples sobre complejas
- **YAGNI** — Implementar solo funcionalidad necesaria

### Enfoques Específicos de Go
- **CSP (Communicating Sequential Processes)** — Uso de canales en lugar de mutex
- **Segregación de Interfaces** — Interfaces locales en módulos consumidores
- **Accept Interfaces, Return Structs** — Trabajo idiomático con interfaces
- **Early Return** — Reducción de anidamiento de código
- **Table-Driven Tests** — Pruebas estructuradas

### Convenciones de Manejo de Errores
- **Errores Estructurados** — Errores estructurados con callstack automático
- **Log o Return** — Decisión entre loggear errores o pasarlos hacia arriba
- **Wrapping Mínimo** — Envolver errores solo al agregar contexto

## Configuración

### Jerarquía de Configuración
```
defaultConfig() → config.yml → .env
```

### Tipos de Configuración
- **Configuración Estática** — Configuraciones básicas de aplicación
- **Configuración Dinámica** — Reglas de reenvío con hot reload
- **Datos Secretos** — Claves API y tokens a través de variables de entorno

### Ejemplo de Configuración de Reenvío
```yaml
forward_rules:
  rule1:
    from: 1001234567890
    to: [1009876543210, 1001111111111]
    send_copy: true
    exclude: "EXCLUDE|spam"
    include: "IMPORTANT|urgent"
    copy_once: false
    indelible: true
```

## Pruebas

### Pruebas Multicapa
- **Pruebas Unitarias** — Pruebas de componentes aislados con mocks
- **Pruebas de Integración** — Pruebas de interacción entre componentes
- **Pruebas E2E** — Escenarios completos de usuario a través de API gRPC
- **Pruebas BDD** — Descripción de comportamiento en lenguaje natural
- **Pruebas de Snapshot** — Pruebas con snapshots de referencia

### Técnicas Especiales
- **Sync Testing** — Pruebas de tiempo y concurrencia
- **Call-Driven Testing** — Pruebas tabulares con funciones de preparación
- **Spy Logging** — Interceptación y verificación de logs en pruebas

### Cobertura de Pruebas
- **Integración con Codecov.io** — Seguimiento automático de cobertura de código
- **Cobertura de Pruebas de Integración** — Herramientas dedicadas
- **Cobertura Funcional** — Todos los escenarios principales de uso
- **Cobertura Técnica** — Funciones internas y casos extremos
- **Escenarios BDD** — Historias de usuario y reglas de negocio

## Despliegue y Operaciones

### Opciones de Inicio
- **Desarrollo Local** — Instalación directa de TDLib en máquina host
- **DevContainer** — Entorno de desarrollo completamente aislado
- **Producción** — Despliegue contenerizado

### Monitoreo y Depuración
- **Logs Estructurados** — Logs JSON para procesamiento automático
- **Logs Legibles** — pplog para desarrollo
- **Health Checks** — Verificación del estado de servicios
- **Graceful Shutdown** — Finalización correcta del trabajo

### Integraciones
- **Cliente Telegram** — Cliente completo con autenticación
- **APIs Externas** — Integración a través de GraphQL, gRPC, REST
- **Cola de Mensajes** — Procesamiento asíncrono de mensajes

## Contribución al Desarrollo

### Filosofía del Proyecto
Budva43 se posiciona como "mi mejor proyecto de aprendizaje para aplicar tecnologías — desde MVP hasta nivel empresarial". Este proyecto demuestra enfoques modernos de desarrollo Go, incluyendo las últimas características del lenguaje y mejores prácticas de la industria.

### Características Únicas
- **Enfoque Experimental** — Uso de características de Go de vanguardia
- **Pruebas Integrales** — Espectro completo de técnicas de testing
- **Calidad Lista para Producción** — Preparado para uso industrial
- **Carácter Educativo** — Documentación rica y ejemplos

El proyecto se desarrolla activamente y sirve como demostración de las capacidades de Go moderno en desarrollo empresarial. 