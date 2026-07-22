# Documentación — Parte 1: Fundación, Estructura y Configuración

## 1. Objetivo de la Parte 1
Establecer la fundación del proyecto **Go Process Supervisor**, definiendo la estructura idiomática de paquetes en Go, el modelo declarativo de configuración (`AppConfig`, `ProcessConfig`, `BackoffConfig`, `RestartPolicy`), la lectura y des-serialización desde archivos JSON, y una validación exhaustiva de reglas de negocio sin recurrir a librerías de terceros ni llamadas a `panic`. Además, exponer el subcomando CLI `validate` para la comprobación previa de archivos de configuración sin iniciar procesos hijos.

---

## 2. Archivos Creados y Modificados
- `go.mod`: Módulo de Go inicializado con el path `github.com/nsilupu-30/go-process-supervisor`.
- `.gitignore`: Exclusión de binarios compilados y archivos temporales de prueba.
- `Makefile`: Automatización de tareas de compilación (`build`), formateo (`fmt`), análisis estático (`vet`) y pruebas (`test`, `race`).
- `internal/config/config.go`: Definición de tipos y estructuras de datos del modelo de configuración.
- `internal/config/loader.go`: Funcionalidad de lectura de archivos JSON y desempaquetado de estructuras.
- `internal/config/validator.go`: Lógica de validación de reglas de negocio de la configuración.
- `internal/config/loader_test.go`: Suite de pruebas unitarias basadas en tablas (*table-driven tests*).
- `cmd/supervisor/main.go`: Punto de entrada CLI con soporte para el subcomando `validate --config <path>`.
- `examples/config.example.json`: Archivo de configuración JSON de muestra con dos procesos de distintas políticas.
- `examples/workers/stable-worker.sh`: Script ejecutable de trabajador estable de prueba.
- `examples/workers/failing-worker.sh`: Script ejecutable de trabajador fallido de prueba.
- `BITACORA_DECISIONES.md`: Registro de la decisión técnica tomada en el Hito H1.
- `DECLARACION_USO_IA.md`: Declaración formal de autoría e implementación propia.
- `Documentacion-parte1.md`: Este documento de especificación técnica y evidencia.

---

## 3. Estructuras Implementadas

### `RestartPolicy`
Tipo de cadena enum que define los comportamientos de reinicio permitidos:
- `"always"`: Reinicia el proceso sin importar su código de salida.
- `"on-failure"`: Reinicia únicamente si el proceso finaliza con código de salida distinto de 0 o error.
- `"never"`: No aplica reinicios automáticos.

### `BackoffConfig`
Define los parámetros para el cálculo del incremento exponencial entre reintentos:
- `InitialSeconds` (int): Tiempo de espera inicial en segundos (`> 0`).
- `Factor` (float64): Multiplicador del tiempo de espera (`>= 1.0`).
- `MaxSeconds` (int): Límite superior máximo de la espera (`>= InitialSeconds`).

### `ProcessConfig`
Representa la declaración de un proceso individual a supervisar:
- `Name` (string): Identificador único obligatorio.
- `Command` (string): Ruta o nombre del ejecutable obligatorio.
- `Args` ([]string): Argumentos de línea de comandos.
- `WorkingDir` (string): Directorio de trabajo.
- `Environment` (map[string]string): Variables de entorno asociadas.
- `RestartPolicy` (`RestartPolicy`): Política de reinicio válida.
- `MaxRetries` (int): Cantidad máxima de reintentos acumulados (`>= 0`).
- `Backoff` (`BackoffConfig`): Configuración de retardo.

### `AppConfig`
Representa el contenedor global del supervisor:
- `GracePeriodSeconds` (int): Tiempo de espera antes de `SIGKILL` (`> 0`).
- `APIAddress` (string): Dirección host:puerto donde escuchará la API HTTP (ej. `127.0.0.1:8080`).
- `Processes` ([]ProcessConfig): Lista de procesos a supervisar (no vacía).

---

## 4. Reglas de Validación
La función `ValidateConfig` exige de forma estricta:
1. Archivo inexistente o JSON malformado: Retorna error envuelto con `fmt.Errorf`.
2. `GracePeriodSeconds`: Debe ser mayor a 0.
3. `APIAddress`: No vacío y con formato válido `host:port` (validado con `net.SplitHostPort`).
4. `Processes`: No puede estar vacío.
5. Nombre de proceso (`Name`): No vacío (tras eliminar espacios) y único en el archivo.
6. Comando (`Command`): No vacío.
7. Política (`RestartPolicy`): Únicamente `always`, `on-failure` o `never`.
8. `MaxRetries`: Mayor o igual a 0.
9. `Backoff.InitialSeconds`: Mayor a 0.
10. `Backoff.Factor`: Mayor o igual a 1.0.
11. `Backoff.MaxSeconds`: Mayor a 0 y mayor o igual a `InitialSeconds`.

---

## 5. Decisiones Técnicas
- **Uso exclusivo de estándar de Go:** No se introdujeron dependencias externas pesadas (`viper`, `cobra`) para mantener el binario ligero, sin deudas ocultas y 100% defendible en evaluación.
- **Envoltorio idiomático de errores (`%w`):** Todo error en la carga o lectura preserva la causa raíz y añade contexto legible sin `panic`.

---

## 6. Pruebas Realizadas
Se implementó una suite de pruebas *table-driven* en `internal/config/loader_test.go` cubriendo:
- Carga de configuración válida.
- Detección de JSON malformado.
- Manejo de ruta de archivo inexistente.
- Detección de lista de procesos vacía.
- Detección de nombres duplicados.
- Detección de comandos vacíos.
- Detección de políticas inválidas.
- Detección de parámetros de backoff fuera de rango (factor < 1.0, max_seconds < initial_seconds).
- Detección de grace period inválido (<= 0).
- Detección de dirección API malformada.

---

## 7. Comandos de Ejecución

### Validar archivo de configuración mediante CLI:
```bash
go run ./cmd/supervisor validate --config examples/config.example.json
```

### Ejecutar suite de pruebas:
```bash
go test -v ./internal/config/...
go test -race ./...
```

---

## 8. Limitaciones Actuales
- La aplicación únicamente lee y valida el archivo JSON.
- No se inician procesos hijos, goroutines ni escuchadores de señales.

---

## 9. Tareas No Implementadas (Reservadas para partes posteriores)
- Parte 2: Invocación de `os/exec.Command`, captura de `stdout`/`stderr` y recolección de PIDs.
- Parte 3: Goroutines de monitoreo, temporizadores de backoff y máquina de estados del proceso.
- Parte 4: Supervisor concurrente global, canal de eventos y snapshots thread-safe.
- Parte 5: Manejo de `SIGINT`, `SIGTERM`, `SIGHUP` y apagado con grace period.
- Parte 6: Servidor HTTP local con endpoints `/processes`, `/health`, etc.
