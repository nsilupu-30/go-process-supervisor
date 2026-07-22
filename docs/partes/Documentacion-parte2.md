# Documentación — Parte 2: Ejecución de Procesos y Captura de Logs

## 1. Objetivo de la Parte 2
Implementar la capa de ejecución de procesos hijos utilizando el paquete estándar `os/exec` de Go. El módulo se encarga de:
1. Iniciar comandos externos configurando su directorio de trabajo y propagando variables de entorno.
2. Capturar asíncronamente las salidas estándar (`stdout`) y de error (`stderr`) mediante goroutines y tuberías (*pipes*), aplicando formateo etiquetado con prefijos `[timestamp] [proceso] [STDOUT|STDERR]`.
3. Registrar de forma precisa el PID (*Process ID*), código de salida (*exit code*) y duración del proceso.
4. Invocar la llamada obligatoria `cmd.Wait()` para la recolección del proceso en el sistema operativo, previniendo la acumulación de procesos zombis o recursos colgados.
5. Distinguir fallos de arranque, ejecuciones exitosas, salidas con código de error y terminaciones por señal del sistema.

---

## 2. Archivos Creados y Modificados
- `internal/logging/logger.go`: Definición de la interfaz `Logger`, la estructura `ProcessLogger` con sincronización `sync.Mutex` y el lector por tuberías `StreamPipe`.
- `internal/process/process.go`: Definición del modelo de resultado de ejecución `ExecutionResult`.
- `internal/process/runner.go`: Implementación de `ProcessRunner` con el método `Run()`.
- `internal/process/runner_test.go`: Suite de pruebas unitarias cubriendo escenarios exitosos, con errores, comandos inválidos, entorno y cancelación por contexto.
- `Documentacion-parte2.md`: Especificación técnica y evidencias de la Parte 2.

---

## 3. Estructuras e Interfaces Implementadas

### `Logger` (`internal/logging/logger.go`)
Interfaz thread-safe para la emisión de logs por proceso:
```go
type Logger interface {
    LogStdout(processName string, line string)
    LogStderr(processName string, line string)
    LogInfo(processName string, message string)
    LogError(processName string, message string)
}
```

### `ExecutionResult` (`internal/process/process.go`)
Representa el resultado final de la ejecución de un proceso hijo:
- `PID` (int): Identificador del proceso asignado por el Kernel del sistema operativo.
- `ExitCode` (int): Código de retorno del proceso (`0` para éxito, `-1` para fallos de arranque, `1-255` para errores del proceso).
- `StartedAt` (time.Time): Estampa de tiempo del inicio (`cmd.Start()`).
- `ExitedAt` (time.Time): Estampa de tiempo del fin de ejecución (`cmd.Wait()`).
- `Duration` (time.Duration): Tiempo total transcurrido.
- `Error` (error): Error devuelto por la llamada del sistema o nulo si finalizó correctamente.
- `TerminatedBySignal` (bool): Verdadero si el proceso fue terminado abruptamente por una señal del OS.
- `Signal` (string): Nombre de la señal de terminación (ej. `SIGKILL`, `SIGTERM`).

### `ProcessRunner` (`internal/process/runner.go`)
Componente responsable del ciclo de vida directo del proceso hijo:
```go
type ProcessRunner struct{}

func (r *ProcessRunner) Run(ctx context.Context, cfg config.ProcessConfig, logger logging.Logger) (*ExecutionResult, error)
```

---

## 4. Flujo de Ejecución y Prevención de Zombis

```text
               config.ProcessConfig
                        │
                        ▼
            exec.CommandContext(ctx)
                        │
       ┌────────────────┴────────────────┐
       ▼                                 ▼
 cmd.StdoutPipe()                  cmd.StderrPipe()
       │                                 │
       ▼ (goroutine)                     ▼ (goroutine)
StreamPipe(LogStdout)             StreamPipe(LogStderr)
       └────────────────┬────────────────┘
                        │
                        ▼
                   cmd.Start() ──► Registra PID y StartedAt
                        │
                        ▼
                   cmd.Wait()  ──► Libera recursos OS (Evita Zombis)
                        │
                        ▼
                wg.Wait() (Logs)
                        │
                        ▼
              ExecutionResult (ExitCode, Duration, Signal)
```

### Prevención de Procesos Zombis
En sistemas operativos tipo POSIX y Windows, cuando un proceso hijo finaliza, su entrada permanece en la tabla de procesos del Kernel hasta que el proceso padre lee su estado de salida. Si el padre omite esta llamada, el proceso se convierte en un **proceso zombi**.
`ProcessRunner` garantiza la eliminación de zombis mediante:
1. Invocación síncrona obligatoria de `cmd.Wait()` tras `cmd.Start()`.
2. Espera explícita con `sync.WaitGroup` a que se consuman todos los datos de las tuberías de salida antes de liberar el resultado.

---

## 5. Decisiones Técnicas
1. **Sincronización Thread-Safe de Logs:** Se protegen las escrituras concurrentes de múltiples goroutines sobre `io.Writer` con un `sync.Mutex` interno en `ProcessLogger`.
2. **Propagación del Entorno:** Se implementó `mergeEnvironment` para conservar las variables de entorno del sistema operativo host (`os.Environ()`) agregando u sobrescribiendo las definidas en `cfg.Environment`.
3. **Abstracción del Logger para Pruebas:** `NewProcessLoggerWithWriter` permite inyectar un `bytes.Buffer` en las pruebas unitarias para capturar y asertar la salida generada sin contaminar `os.Stdout`.

---

## 6. Pruebas Realizadas

La suite de pruebas en `internal/process/runner_test.go` valida:
- `TestProcessRunner_SuccessfulExecution`: Ejecución limpia con captura de PID y logs etiquetados.
- `TestProcessRunner_NonZeroExitCode`: Captura del código de salida devuelto por procesos fallidos (ej. código 42).
- `TestProcessRunner_NonExistentCommand`: Manejo seguro de ejecutables inexistentes sin provocar `panic`.
- `TestProcessRunner_EnvironmentVariables`: Inyección y lectura de variables de entorno personalizadas en el proceso hijo.
- `TestProcessRunner_ContextCancellation`: Cancelación inmediata del proceso ante un tiempo de espera excedido (*timeout*).

### Comandos de Ejecución de Pruebas:
```bash
go test -v ./internal/process/...
go test -race ./...
```
