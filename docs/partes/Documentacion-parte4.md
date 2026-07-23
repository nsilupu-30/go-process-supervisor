# Documentación — Parte 4: Supervisor concurrente y máquina de estados

## 1. Objetivo de la Parte 4
Coordinar múltiples procesos hijos de forma concurrente, thread-safe y sin deadlocks. Esta parte introduce:
1. Una máquina de estados pura por proceso.
2. Un administrador de proceso que ejecuta un ciclo de vida en su propia goroutine.
3. Un supervisor orquestador que crea y espera administradores concurrentemente.
4. Snapshots consultables sin mutar estado compartido.
5. Canal de comandos con propietario único para evitar condiciones de carrera.

## 2. Archivos Creados y Modificados
- `internal/supervisor/supervisor.go`: `Supervisor` y `Supervisar` públicos.
- `internal/supervisor/manager.go`: `AdministradorProceso`, canal de comandos, ciclo de vida y reinicio con backoff.
- `internal/supervisor/state.go`: Tipos `EstadoProceso`, `EventoProceso`, máquina de estados pura y transiciones válidas.
- `internal/supervisor/snapshot.go`: `SnapshotProceso`, `almacenSnapshots` con `sync.RWMutex` para lecturas concurrentes sin data races.
- `cmd/supervisor/main.go`: integración del `Supervisor` en el comando `run` y espera de señales en main.
- `internal/supervisor/supervisor_test.go`: pruebas concurrentes y de cierre.
- `Documentacion-parte4.md`: este documento.

## 3. Tipos Exportados

### EstadoProceso
Estados válidos serializados en string:
- `EstadoCreado` = "creado"
- `EstadoIniciando` = "iniciando"
- `EstadoEjecutando` = "ejecutando"
- `EstadoEspera` = "espera"
- `EstadoDeteniendo` = "deteniendo"
- `EstadoDetenido` = "detenido"
- `EstadoFallido` = "fallido"

### EventoProceso
Eventos internos del sistema:
- `EventoProcesoIniciado` = "proceso_iniciado"
- `EventoProcesoSalido` = "proceso_salido"
- `EventoProcesoFallido` = "proceso_fallido"
- `EventoReinicioProgramado` = "reinicio_programado"
- `EventoProcesoDeteniendo` = "proceso_deteniendo"
- `EventoProcesoDetenido` = "proceso_detenido"
- `EventoApagadoSolicitado` = "apagado_solicitado"

### SnapshotProceso
Datos consultables de un proceso:
- `Nombre`, `Estado`, `PID`, `Reinicios`, `CodigoSalida`, `Error`, `Inicio`, `Salida`, `Siguiente`

### AdministradorProceso
Constructor: `NuevoAdministradorProceso(cfg, logger) *AdministradorProceso`

Métodos públicos:
- `EnviarComando(tipo comandoManager, ctx context.Context)`
- `ObtenerSnapshot() SnapshotProceso`
- `IniciarCicloVida(ctx context.Context)`

### Supervisor
Constructor: `NuevoSupervisor(cfg config.AppConfig, logger logging.Logger) *Supervisor`

Métodos públicos:
- `Iniciar(ctx context.Context)`
- `Detener(ctx context.Context)`
- `ObtenerSnapshots() []SnapshotProceso`

## 4. Máquina de estados

```
CREADO
  → INICIANDO   (EventoProcesoIniciado)

INICIANDO
  → ESPERA   (EventoProcesoSalido | EventoProcesoFallido)
  → ESPERA   (EventoProcesoFallido en fallo de ejecución)

ESPERA
  → INICIANDO   (EventoProcesoIniciado tras backoff)

EJECUTANDO
  → DETENIENDO   (EventoProcesoDeteniendo)

DETENIENDO
  → DETENIDO   (EventoProcesoDetenido)

APAGADO_SOLICITADO
  → DETENIENDO   (si no está detenido ni fallido)
```

Reglas:
- Cada transición valida el estado actual antes de aplicarse.
- `transicionar` retorna `false` si la transición no es válida.
- `apagar` aplica `EventoApagadoSolicitado` solo si aún no está detenido o fallido.
- La máquina no conoce políticas de reinicio; eso vive exclusivamente en `AdministradorProceso`.

## 5. Modelo de concurrencia

### Propietarios de canales
- Cada `AdministradorProceso` es el único escritor y lector de su propio `canal comando` (buffered de 1).
- El `Supervisor` solo envía comandos iniciales; la goroutine del manager los consume.
- El método `EnviarComando` usa `select` con `default` para no bloquear si el canal está lleno.

### Thread-safety
- `sync.Mutex` protege el flag `detenido` y la escritura en el canal del manager.
- `sync.RWMutex` en `almacenSnapshots`: escrituras exclusivas, lecturas concurrentes.
- El `Supervisor` coordina goroutines con `sync.WaitGroup`.
- Todo bloqueo que pueda cancelarse usa `context.Context`.

### Cierre limpio
- `IniciarCicloVida` cierra su defer: marca `detenido = true` y cierra el canal.
- `ctx.Done()` interrumpe backoff y el loop principal.
- El comando `run` en main escucha `SIGINT`/`SIGTERM`, crea un `context.WithCancel` y propaga cancelación.

## 6. Decisiones técnicas
1. Máquina de estados pura: sin lógica de tiempo ni E/S, ni dependencias externas.
2. Snapshots como proyección inmutable de estado interno: la API lee snapshots, nunca el estado mutable.
3. Canal de comandos con buffer 1 y non-blocking send: evita deadlocks si el manager está ocupado.
4. Reinicio recursivo: `ejecutarProceso` se llama a sí misma tras backoff, cancelable por `ctx`.
5. No se usa `sync.WaitGroup` dentro del manager para esperar el proceso; el propio `ctx` orquesta la cancelación.
6. El `Supervisor.Iniciar` usa `sync.WaitGroup` para esperar a todos los managers, no `time.Sleep` ni canales extra.

## 7. Pruebas y validación
Suite en `internal/supervisor/supervisor_test.go`:
- `TestAdministradorProceso_SnapshotInicial`
- `TestAdministradorProceso_ProcesoExitosoNoReinicia`
- `TestAdministradorProceso_CancelacionInterrumpeEspera`
- `TestSupervisor_VariosProcesosSimultaneos`
- `TestSupervisor_ApagadoOrdenado`
- `TestAdministradorProceso_ComandosNoDuplican`

Ejecución:
```bash
go test ./internal/supervisor/...
go test ./...
gofmt -w .
```

## 8. Limitaciones
- La recarga dinámica (`SIGHUP`) aún no está implementada (Parte 5).
- La API HTTP no expone aún los nuevos endpoints (Parte 6).
- `BackoffConfig.MaxRetries` se evalúa en `manager.go`, no en la máquina de estados.
- En Windows algunos comandos pueden requerir `cmd.exe /c` explícito.
