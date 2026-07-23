# Documentación — Parte 6: API HTTP Local y Dashboard

## 1. Objetivo de la Parte 6
Exponer el estado del supervisor y controlar los procesos mediante una API HTTP local. El objetivo es permitir inspección y comandos remotos sobre el supervisor sin necesidad de usar directamente la CLI o los procesos internos.

## 2. Endpoints HTTP implementados
El módulo `internal/api` expone los siguientes endpoints:

- `GET /health`
  - Devuelve `200 OK` con `{ "message": "ok" }` cuando el supervisor está saludable.
  - Devuelve `503 Service Unavailable` si la evaluación de salud falla.

- `GET /processes`
  - Devuelve el listado completo de snapshots de procesos supervisados.
  - Cada elemento es un `SnapshotProceso` con campos JSON como `name`, `state`, `pid`, `restart_count`, `exit_code`, `error`, `started_at`, `exited_at` y `next_retry_at`.

- `GET /processes/{name}`
  - Devuelve el snapshot de un proceso en particular.
  - Si el proceso no existe, responde `404 Not Found` con un mensaje de error JSON.

- `POST /processes/{name}/start`
  - Encola el comando de inicio para el proceso indicado.
  - Responde `200 OK` con `{ "message": "command accepted" }`.

- `POST /processes/{name}/stop`
  - Encola la detención del proceso indicado.
  - Responde `200 OK` con `{ "message": "command accepted" }`.

- `POST /processes/{name}/restart`
  - Encola el reinicio del proceso indicado.
  - Responde `200 OK` con `{ "message": "command accepted" }`.

- `POST /reload`
  - Dispara la misma lógica de recarga de configuración usada por `SIGHUP`.
  - Responde `200 OK` con `{ "message": "reload triggered" }`.

## 3. Diseño de inyección de dependencias
El servidor HTTP usa un constructor `NewServer(address string, sup Supervisor) *Server`.

### Interfaz `Supervisor`
Para desacoplar la API del supervisor concreto se definió la interfaz `internal/api.Supervisor` con los métodos:

- `Health() bool`
- `ProcessSnapshots() []supervisor.SnapshotProceso`
- `ProcessSnapshot(name string) (supervisor.SnapshotProceso, bool)`
- `StartProcess(name string, ctx context.Context) error`
- `StopProcess(name string, ctx context.Context) error`
- `RestartProcess(name string, ctx context.Context) error`
- `Reload() error`

Esto permite:
- Probar los handlers con `httptest` usando un mock simple.
- Mantener bajo acoplamiento entre `internal/api` y `internal/supervisor`.
- Aislar la lógica de HTTP de la lógica de control de procesos.

## 4. Pruebas con `httptest`
Se implementaron pruebas unitarias en `internal/api/handlers_test.go`.

- Se creó un mock de `Supervisor` que implementa los métodos de la interfaz.
- Las pruebas realizan solicitudes HTTP simuladas con `httptest.NewRequest` y `httptest.NewRecorder`.
- Se valida el código HTTP y el JSON devuelto.
- Se prueba la ruta `/reload` y los comandos de proceso para garantizar que la API invoca el supervisor esperado.

## 5. Graceful shutdown del servidor HTTP
El servidor HTTP está encapsulado en la estructura `internal/api.Server`.

### Métodos principales
- `Start() error`
  - Arranca `http.Server.ListenAndServe()`.
  - Devuelve `nil` cuando el servidor se cierra de forma intencional (`http.ErrServerClosed` se ignora).

- `Stop(ctx context.Context) error`
  - Llama a `http.Server.Shutdown(ctx)` para permitir que las peticiones en curso finalicen.
  - Retorna error si el cierre ordenado falla.

### Integración con el flujo global
En `cmd/supervisor/main.go`:
- Se crea `api.NewServer(cfg.APIAddress, sup)`.
- Se arranca en goroutine paralela a `sup.Iniciar(ctx)`.
- Al recibir `SIGINT`/`SIGTERM`, se llama a `apiServer.Stop(ctxCancelado)` antes de detener el supervisor.

Esto garantiza que:
- Las peticiones HTTP abiertas terminan limpiamente.
- El servidor no queda escuchando mientras el supervisor cierra los procesos.
- El cierre del supervisor y del API ocurren de forma coordinada.

## 6. Resultados de calidad
Se generaron evidencias de calidad en la carpeta `EVIDENCIAS/` con:
- `build.txt`
- `vet.txt`
- `test.txt`
- `race.txt` (si está disponible en el entorno con CGO habilitado)

## 7. Limitaciones conocidas
- El endpoint `/reload` depende de la implementación de recarga en `Supervisor.Reload()` y del manejador de señales. Si la recarga no actualiza procesos activos, el endpoint aún responde, pero la lógica interna debe extenderse para aplicar cambios completos de configuración.
- El estado de salud actual es un chequeo simple (`Health() bool` siempre verdadero) y puede evolucionar hacia una validación más completa en el futuro.

## 8. Archivos clave de Parte 6
- `internal/api/server.go`
- `internal/api/handlers.go`
- `internal/api/handlers_test.go`
- `cmd/supervisor/main.go`
- `docs/partes/Documentacion-parte6.md`
