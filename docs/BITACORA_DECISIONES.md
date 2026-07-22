# Bitácora de Decisiones de Diseño — Go Process Supervisor

Este documento registra cronológicamente las decisiones técnicas y de arquitectura tomadas durante el desarrollo del proyecto.

---

## Decisión H1-01 — Estructura del Modelo de Configuración y Mecanismo de Validación

- **Fecha:** 2026-07-22
- **Hito:** H1 — Configuración y estructura inicial
- **Parte / Integrante:** Parte 1 — Fundación, estructura y configuración
- **Problema:** Se requería definir cómo cargar, representar y validar la configuración del supervisor desde un archivo JSON, garantizando que los datos sean semánticamente correctos antes de iniciar cualquier goroutine o proceso hijo, sin comprometer la sencillez ni la capacidad de defender el código.
- **Decisión:** Implementar un paquete independiente `internal/config` usando la librería estándar de Go (`encoding/json`), separando el modelo de datos (`config.go`), la des-serialización de archivos (`loader.go`) y la validación semántica exhaustiva (`validator.go`).
- **Alternativas consideradas:**
  1. *Librería externa de terceos (Viper / Cobra):* Descartada porque añade dependencias pesadas innecesarias y reduce el control directo exigido en las especificaciones del curso.
  2. *Validación implícita mediante unmarshal de JSON únicamente:* Descartada porque JSON solo valida sintaxis básica y tipos de datos, pero no reglas de negocio (ej. nombres duplicados, tiempos positivos, factor de backoff >= 1.0).
- **Justificación técnica:** La separación en `loader.go` y `validator.go` mantiene el principio de responsabilidad única. La validación explícita mediante un recorrido determinista permite retornar errores detallados que envuelven la causa (`fmt.Errorf("...: %w", err)`), evitando panics y garantizando la robustez requerida.
- **Consecuencias y limitaciones:** Toda nueva propiedad agregada al archivo de configuración requerirá actualizar explícitamente `config.go` y la función de validación en `validator.go`. Como beneficio, la aplicación rechaza configuraciones inválidas en tiempo de inicio.
- **Evidencia:**
  - Archivos: `internal/config/config.go`, `internal/config/loader.go`, `internal/config/validator.go`, `internal/config/loader_test.go`
  - Pruebas automatizadas ejecutan escenarios de prueba exitosos y de error.
- **Uso de IA:** Ninguno. Diseño e implementación realizados 100% por el estudiante utilizando la documentación oficial de Go.
