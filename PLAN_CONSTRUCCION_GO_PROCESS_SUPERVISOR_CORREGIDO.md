# Plan de Construcción — **Go Process Supervisor**

## Supervisor de Procesos y Job Scheduler en Go

**Nombre del proyecto:** Go Process Supervisor  
**Repositorio sugerido:** `go-process-supervisor`  
**Lenguaje:** Go  
**Tipo de aplicación:** CLI + API HTTP local  
**Enfoque:** Gestión concurrente del ciclo de vida de procesos hijos  
**Equipo propuesto:** 6 integrantes · 6 partes secuenciales e integrables  
**Sistema operativo objetivo:** Linux  
**Nivel esperado:** Medio–avanzado  

---

> **Objetivo general:** Construir un supervisor de procesos en Go capaz de leer una configuración, iniciar varios procesos hijos, capturar sus salidas, vigilar su ejecución, reiniciarlos según una política, aplicar backoff exponencial, responder a señales del sistema y apagarse sin dejar procesos zombis ni goroutines colgadas.

---

## Importante — Ubicación de los archivos del plan

Los siguientes archivos de gestión deben colocarse en la **raíz del proyecto**:

- `PLAN_CONSTRUCCION.md`
- `GUIA_ENTREGAS.md`
- `TEMPLATE_DOCUMENTACION.md`
- `BITACORA_DECISIONES.md`
- `DECLARACION_USO_IA.md`
- `README.md`

La carpeta `GRAFICOS/` se conserva como carpeta independiente para diagramas, capturas y evidencias.

El código fuente sí debe respetar la estructura de carpetas definida en este plan.

---

## Instrucciones para IA — Cómo usar este plan

Este archivo es el documento maestro del proyecto. Contiene el alcance, arquitectura, estructura, reglas técnicas, división del trabajo, criterios de aceptación y estándares del equipo.

Cuando un integrante solicite:

> “Realiza la parte [N] según `PLAN_CONSTRUCCION.md`”

la IA debe:

1. Leer el plan completo.
2. Ubicar la sección **PARTE [N]**.
3. Respetar la arquitectura y las interfaces ya definidas.
4. No reemplazar módulos existentes sin justificarlo.
5. Aplicar los estándares de código, commits y pruebas.
6. Crear o actualizar `Documentacion-parte[N].md`.
7. Indicar claramente qué código debe comprender y defender el integrante.
8. No generar el proyecto completo de una sola vez.
9. No tomar decisiones centrales sin explicarlas.
10. Priorizar el aprendizaje y la autoría del equipo.

### Uso permitido de IA

- Explicar errores de compilación o concurrencia.
- Revisar código escrito por el integrante.
- Proponer casos de prueba.
- Comparar alternativas como channels frente a mutex.
- Explicar `context`, `os/exec`, señales o backoff.
- Revisar estilo idiomático de Go.

### Uso no permitido

- Entregar todo el supervisor generado sin comprensión.
- Copiar un módulo central sin poder explicarlo.
- Ocultar el uso de IA.
- Usar código que el equipo no pueda modificar en la defensa.
- Delegar por completo el modelo de concurrencia o la máquina de estados.
- Usar IA durante la defensa oral o durante una evaluación de comprensión.
- Entregar conceptos, librerías o construcciones que el integrante no pueda explicar línea por línea.

## Controles obligatorios de autoría

La evaluación no se limita a comprobar que el programa funcione. También se verificará que el desarrollo sea progresivo, comprensible y defendible.

1. **Historia Git incremental:** commits pequeños, frecuentes y descriptivos. Un único commit masivo o una historia artificial se considera señal de alarma.
2. **Bitácora por hito:** cada hito debe registrar las decisiones técnicas relevantes, alternativas consideradas y justificación.
3. **Declaración de uso de IA:** debe indicar herramientas, finalidad, archivos afectados, partes de autoría íntegra y declaración firmada.
4. **Defensa oral:** duración obligatoria de 15 a 20 minutos.
5. **Modificación en vivo:** el docente podrá solicitar un cambio pequeño sobre el código durante la defensa.
6. **Preguntas de comprensión:** cada integrante debe justificar estructuras, casos borde, concurrencia, señales, pruebas y depuración.
7. **Consecuencia por falta de autoría:** una parte que no pueda explicarse o modificarse puede calificarse con cero y el proyecto no podrá superar el nivel «En desarrollo».

---

# 1. Alcance funcional

El sistema permitirá:

1. Leer procesos desde un archivo de configuración JSON.
2. Iniciar varios procesos hijos con `os/exec`.
3. Enviar argumentos, variables de entorno y directorio de trabajo.
4. Capturar `stdout` y `stderr`.
5. Detectar cuándo termina un proceso.
6. Aplicar políticas de reinicio:
   - `always`
   - `on-failure`
   - `never`
7. Aplicar backoff exponencial con límite máximo.
8. Mantener una máquina de estados por proceso.
9. Manejar `SIGINT`, `SIGTERM` y `SIGHUP`.
10. Realizar apagado ordenado con periodo de gracia.
11. Evitar procesos zombis y goroutines filtradas.
12. Recargar la configuración sin reiniciar todo el supervisor.
13. Exponer una API HTTP local para consultar y controlar procesos.
14. Ejecutar pruebas con `go test ./...` y `go test -race ./...`.

---

# 2. Fuera de alcance inicial

No se implementará en la primera versión:

- Interfaz web gráfica.
- Contenedores Docker.
- Gestión distribuida entre servidores.
- Base de datos externa.
- Autenticación de usuarios.
- Orquestación tipo Kubernetes.
- Programación cron completa.
- Reinicio automático del sistema operativo.
- Ejecución remota por SSH.

Estas funciones podrán considerarse como mejoras futuras.

---

# 3. Arquitectura general

```text
┌─────────────────────────────┐
│          Usuario            │
│ CLI o cliente HTTP local    │
└──────────────┬──────────────┘
               │
               ▼
┌─────────────────────────────┐
│            CLI              │
│ start · status · stop       │
└──────────────┬──────────────┘
               │
               ▼
┌─────────────────────────────┐
│         Supervisor          │
│ coordinación y ciclo global │
└───────┬──────────┬──────────┘
        │          │
        │          └───────────────┐
        ▼                          ▼
┌──────────────────┐       ┌──────────────────┐
│ Process Manager  │       │ Signal Manager   │
│ start/stop/wait  │       │ INT/TERM/HUP     │
└────────┬─────────┘       └──────────────────┘
         │
         ▼
┌──────────────────┐
│ Process Runner   │
│ os/exec + logs   │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│ Procesos hijos   │
└──────────────────┘
```

---

# 4. Flujo principal

```text
Leer configuración
        ↓
Validar procesos
        ↓
Crear supervisor
        ↓
Crear un manager por proceso
        ↓
Iniciar cada proceso en su goroutine
        ↓
Esperar terminación
        ↓
Evaluar política de reinicio
        ↓
Aplicar backoff
        ↓
Reiniciar o detener
        ↓
Esperar señal global
        ↓
Apagado ordenado
```

---

# 5. Estructura del proyecto

```text
go-process-supervisor/
├── cmd/
│   └── supervisor/
│       └── main.go
│
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   ├── loader.go
│   │   ├── validator.go
│   │   └── loader_test.go
│   │
│   ├── process/
│   │   ├── process.go
│   │   ├── runner.go
│   │   ├── runner_test.go
│   │   ├── state.go
│   │   └── event.go
│   │
│   ├── restart/
│   │   ├── policy.go
│   │   ├── backoff.go
│   │   └── backoff_test.go
│   │
│   ├── supervisor/
│   │   ├── supervisor.go
│   │   ├── manager.go
│   │   ├── manager_test.go
│   │   └── snapshot.go
│   │
│   ├── signals/
│   │   └── handler.go
│   │
│   ├── api/
│   │   ├── server.go
│   │   ├── handlers.go
│   │   └── handlers_test.go
│   │
│   └── logging/
│       └── logger.go
│
├── examples/
│   ├── config.example.json
│   └── workers/
│       ├── stable-worker.sh
│       └── failing-worker.sh
│
├── scripts/
│   ├── build.sh
│   ├── test.sh
│   └── demo.sh
│
├── GRAFICOS/
│   ├── arquitectura.png
│   ├── maquina-estados.png
│   └── flujo-apagado.png
│
├── PLAN_CONSTRUCCION.md
├── GUIA_ENTREGAS.md
├── TEMPLATE_DOCUMENTACION.md
├── BITACORA_DECISIONES.md
├── DECLARACION_USO_IA.md
├── README.md
├── go.mod
├── go.sum
├── Makefile
└── .gitignore
```

---

# 6. Archivo de configuración

Archivo sugerido: `examples/config.example.json`

```json
{
  "grace_period_seconds": 8,
  "api_address": "127.0.0.1:8080",
  "processes": [
    {
      "name": "worker-estable",
      "command": "/bin/sh",
      "args": ["examples/workers/stable-worker.sh"],
      "working_dir": ".",
      "environment": {
        "APP_ENV": "development"
      },
      "restart_policy": "always",
      "max_retries": 5,
      "backoff": {
        "initial_seconds": 1,
        "factor": 2,
        "max_seconds": 30
      }
    },
    {
      "name": "worker-falla",
      "command": "/bin/sh",
      "args": ["examples/workers/failing-worker.sh"],
      "working_dir": ".",
      "environment": {},
      "restart_policy": "on-failure",
      "max_retries": 3,
      "backoff": {
        "initial_seconds": 1,
        "factor": 2,
        "max_seconds": 10
      }
    }
  ]
}
```

---

# 7. Modelos principales

## `ProcessConfig`

Representa la configuración declarativa de un proceso.

Campos:

- `Name`
- `Command`
- `Args`
- `WorkingDir`
- `Environment`
- `RestartPolicy`
- `MaxRetries`
- `Backoff`

## `ProcessState`

Estados permitidos:

```text
created
starting
running
backoff
stopping
stopped
failed
```

## `ProcessEvent`

Eventos internos:

```text
process_started
process_exited
process_failed
restart_scheduled
process_stopping
process_stopped
configuration_reloaded
shutdown_requested
```

## `ProcessSnapshot`

Datos consultables:

- nombre
- estado
- PID
- cantidad de reinicios
- último código de salida
- último error
- fecha de inicio
- fecha de salida
- próximo reintento

---

# 8. Máquina de estados

```text
CREATED
   │
   ▼
STARTING
   │
   ├── error al iniciar ─────────► FAILED
   │
   ▼
RUNNING
   │
   ├── stop solicitado ─────────► STOPPING ─► STOPPED
   │
   ├── salida exit 0 ───────────► evaluación de política
   │
   └── salida exit != 0 ────────► evaluación de política
                                      │
                                      ├── reiniciar ─► BACKOFF ─► STARTING
                                      └── no reiniciar ─────────► STOPPED/FAILED
```

Regla: ninguna transición debe hacerse directamente desde múltiples goroutines sin coordinación.

---

# 9. Políticas de reinicio

## `always`

Reinicia el proceso sin importar si terminó correctamente o con error.

## `on-failure`

Reinicia únicamente cuando el código de salida es distinto de cero o el proceso termina por error.

## `never`

No reinicia.

## Límite de reintentos

Cuando se alcanza `max_retries`, el proceso pasa a:

```text
failed
```

y no vuelve a reiniciarse automáticamente.

---

# 10. Backoff exponencial

Fórmula:

```text
espera = initial × factor ^ intento
```

Con límite:

```text
espera <= max_seconds
```

Ejemplo:

```text
Intento 1 → 1 s
Intento 2 → 2 s
Intento 3 → 4 s
Intento 4 → 8 s
Intento 5 → 16 s
Intento 6 → 30 s máximo
```

El backoff debe poder cancelarse mediante `context.Context`.

---

# 11. Modelo de concurrencia

Cada proceso tendrá una goroutine responsable de su propio ciclo de vida.

```text
Supervisor
├── goroutine manager proceso A
├── goroutine manager proceso B
├── goroutine manager proceso C
├── goroutine API HTTP
└── goroutine señales
```

Principios:

1. Cada manager controla un solo proceso.
2. El supervisor no modifica directamente el estado interno de un proceso.
3. Los eventos se comunican mediante channels.
4. Las lecturas de estado se realizan mediante snapshots seguros.
5. Todo channel debe tener propietario claro.
6. Toda goroutine debe tener una forma explícita de terminar.
7. Todo bloqueo debe poder cancelarse con `context`.
8. El proyecto debe pasar `go test -race ./...`.

---

# 12. Manejo de señales

## `SIGINT` y `SIGTERM`

Flujo:

```text
Señal recibida
      ↓
Cancelar contexto global
      ↓
Solicitar terminación a procesos hijos
      ↓
Esperar grace period
      ↓
Enviar SIGKILL a procesos que no terminaron
      ↓
Esperar todas las goroutines
      ↓
Cerrar API y logs
      ↓
Salir
```

## `SIGHUP`

Flujo:

```text
SIGHUP
  ↓
Leer nuevamente el archivo
  ↓
Validar configuración
  ↓
Comparar configuración anterior y nueva
  ↓
Agregar procesos nuevos
  ↓
Actualizar procesos modificados
  ↓
Detener procesos eliminados
```

Una configuración inválida no debe tumbar el supervisor. Debe conservar la configuración anterior y registrar el error.

---

# 13. API HTTP local

Dirección por defecto:

```text
127.0.0.1:8080
```

Endpoints mínimos:

| Método | Ruta | Función |
|---|---|---|
| `GET` | `/health` | Verificar que el supervisor está activo |
| `GET` | `/processes` | Listar procesos y estados |
| `GET` | `/processes/{name}` | Consultar un proceso |
| `POST` | `/processes/{name}/start` | Iniciar un proceso detenido |
| `POST` | `/processes/{name}/stop` | Detener un proceso |
| `POST` | `/processes/{name}/restart` | Reiniciar un proceso |
| `POST` | `/reload` | Recargar configuración |

Ejemplo:

```bash
curl http://127.0.0.1:8080/processes
```

Respuesta esperada:

```json
{
  "processes": [
    {
      "name": "worker-estable",
      "state": "running",
      "pid": 4242,
      "restart_count": 0
    }
  ]
}
```

---

# 14. CLI

Comandos mínimos:

```bash
go run ./cmd/supervisor start --config examples/config.example.json
go run ./cmd/supervisor validate --config examples/config.example.json
go run ./cmd/supervisor version
```

Opcionales:

```bash
go run ./cmd/supervisor status
go run ./cmd/supervisor stop
go run ./cmd/supervisor restart worker-estable
```

---

# 15. Correspondencia obligatoria con los hitos de la guía

Las partes del equipo deben integrarse hasta cumplir los cinco hitos oficiales. No basta con terminar una parte aislada.

| Hito oficial | Requisito evaluable | Evidencia mínima |
|---|---|---|
| **H1 — Configuración y arranque** | Leer JSON/TOML/YAML con comando, argumentos, entorno y directorio; iniciar procesos con `os/exec`; capturar `stdout` y `stderr` | Configuración de ejemplo, logs identificados y prueba de proceso iniciado |
| **H2 — Monitoreo y políticas** | Detectar terminación; aplicar `always`, `on-failure` y `never`; una goroutine coordinada por `context` por proceso | Demo de las tres políticas y pruebas automatizadas |
| **H3 — Backoff y estados** | Backoff exponencial, tope de reintentos y estados `running`, `backoff`, `stopped`, `failed`; evitar tormentas de reinicio | Logs de esperas progresivas, límite alcanzado y pruebas deterministas |
| **H4 — Señales y cierre limpio** | `SIGINT`/`SIGTERM`, periodo de gracia, `SIGKILL` de último recurso, espera de hijos, ausencia de zombis; `SIGHUP` para recarga sin reiniciar todo | Demo de Ctrl+C, recarga válida/inválida y evidencia de cierre completo |
| **H5 — Control externo** | Endpoint HTTP o socket Unix para consultar estado y controlar `start`, `stop` y `restart` | API funcional con pruebas `httptest` y comandos de demostración |

### Característica opcional de nivel superior

- Scheduling tipo cron para tareas periódicas.

No implementar cron no invalida el proyecto porque la guía lo considera opcional. Debe priorizarse primero la solidez de H1 a H5.

---

# 16. División del proyecto por partes

## PARTE 1 — Fundación, estructura y configuración

**Integrante:** Miembro 1

### Objetivo

Inicializar el proyecto y construir la carga segura del archivo de configuración.

### Tareas

1. Crear el módulo Go.
2. Crear la estructura de carpetas.
3. Implementar `ProcessConfig`, `BackoffConfig` y `AppConfig`.
4. Implementar carga desde JSON.
5. Validar:
   - nombres únicos;
   - comando obligatorio;
   - política válida;
   - tiempos positivos;
   - factor de backoff válido;
   - dirección HTTP válida.
6. Crear el comando `validate`.
7. Crear `config.example.json`.
8. Crear trabajadores de ejemplo.
9. Crear pruebas table-driven del loader.
10. Crear `Documentacion-parte1.md`.

### Al terminar

- `go build ./...` funciona.
- `go test ./...` funciona.
- Configuración válida se carga correctamente.
- Configuración inválida devuelve errores descriptivos.
- No se usa `panic` para errores normales.
- El comando `validate` informa si el archivo es correcto.

### Commits sugeridos

```text
🔧 config: inicializar módulo y estructura del proyecto
✨ feat(config): agregar modelos de configuración
✨ feat(config): implementar carga de archivo json
✅ test(config): validar configuraciones correctas e inválidas
📝 docs(parte1): documentar fundación y configuración
```

---

## PARTE 2 — Ejecución de procesos y captura de logs

**Integrante:** Miembro 2

### Objetivo

Iniciar procesos hijos con `os/exec` y capturar correctamente su salida.

### Tareas

1. Crear `ProcessRunner`.
2. Ejecutar comando, argumentos y directorio.
3. Combinar variables de entorno del host y del proceso.
4. Capturar `stdout`.
5. Capturar `stderr`.
6. Registrar PID.
7. Esperar la terminación con `Wait()`.
8. Obtener exit code.
9. Diferenciar:
   - error al iniciar;
   - terminación exitosa;
   - terminación con error;
   - terminación por señal.
10. Evitar procesos zombis.
11. Crear pruebas con procesos controlados.
12. Crear `Documentacion-parte2.md`.

### Al terminar

- El supervisor puede ejecutar un proceso.
- Los logs aparecen identificados por nombre.
- Se conoce el PID.
- Se captura el exit code.
- `Wait()` se invoca correctamente.
- No quedan procesos zombis.

### Commits sugeridos

```text
✨ feat(process): implementar ejecución con os exec
✨ feat(logging): identificar stdout y stderr por proceso
✨ feat(process): capturar pid y código de salida
✅ test(process): probar procesos exitosos y fallidos
📝 docs(parte2): documentar ejecución de procesos
```

---

## PARTE 3 — Monitoreo, reinicios y backoff

**Integrante:** Miembro 3

### Objetivo

Monitorear procesos y aplicar políticas de reinicio sin generar tormentas.

### Tareas

1. Implementar `RestartPolicy`.
2. Implementar `always`.
3. Implementar `on-failure`.
4. Implementar `never`.
5. Implementar contador de reintentos.
6. Implementar backoff exponencial.
7. Limitar el backoff.
8. Cancelar el backoff mediante context.
9. Evitar reinicios cuando el supervisor se está apagando.
10. Cambiar el estado a `failed` cuando se alcance el máximo.
11. Crear pruebas deterministas del backoff.
12. Crear `Documentacion-parte3.md`.

### Al terminar

- Cada política funciona.
- El proceso no reinicia en bucle inmediato.
- Se respeta `max_retries`.
- El backoff tiene tope.
- El temporizador se cancela durante el apagado.
- Existen pruebas para todas las políticas.

### Commits sugeridos

```text
✨ feat(restart): agregar políticas de reinicio
✨ feat(restart): implementar backoff exponencial
🐛 fix(restart): evitar reinicio durante el apagado
✅ test(restart): probar políticas y límite de reintentos
📝 docs(parte3): documentar reinicios y backoff
```

---

## PARTE 4 — Supervisor concurrente y máquina de estados

**Integrante:** Miembro 4

### Objetivo

Coordinar varios procesos concurrentemente de forma segura.

### Tareas

1. Crear `ProcessManager`.
2. Crear `Supervisor`.
3. Ejecutar un manager por proceso.
4. Implementar eventos internos.
5. Implementar máquina de estados.
6. Crear snapshots de consulta.
7. Proteger el estado compartido.
8. Evitar condiciones de carrera.
9. Coordinar cierre con `sync.WaitGroup`.
10. Propagar `context.Context`.
11. Crear pruebas concurrentes.
12. Ejecutar `go test -race ./...`.
13. Crear `Documentacion-parte4.md`.

### Al terminar

- Se ejecutan varios procesos a la vez.
- Cada proceso mantiene su ciclo de vida independiente.
- No existen deadlocks conocidos.
- No hay goroutines filtradas.
- La máquina de estados rechaza transiciones inválidas.
- Todo el proyecto pasa `go test -race ./...`.

### Commits sugeridos

```text
✨ feat(supervisor): agregar coordinación de múltiples procesos
✨ feat(process): implementar máquina de estados
♻️ refactor(supervisor): centralizar eventos por channels
✅ test(supervisor): probar concurrencia y cancelación
📝 docs(parte4): documentar modelo concurrente
```

---

## PARTE 5 — Señales, apagado ordenado y recarga

**Integrante:** Miembro 5

### Objetivo

Controlar señales del sistema y garantizar un cierre limpio.

### Tareas

1. Escuchar `SIGINT`.
2. Escuchar `SIGTERM`.
3. Escuchar `SIGHUP`.
4. Cancelar el contexto global.
5. Enviar terminación a procesos hijos.
6. Implementar grace period.
7. Enviar `SIGKILL` cuando sea necesario.
8. Esperar salida de todos los procesos.
9. Recargar el archivo de configuración.
10. Comparar configuración anterior y nueva.
11. Mantener configuración anterior ante errores.
12. Crear pruebas de apagado.
13. Crear `Documentacion-parte5.md`.

### Al terminar

- `Ctrl+C` apaga el sistema ordenadamente.
- Los procesos reciben señal de terminación.
- Se respeta el periodo de gracia.
- No quedan hijos ejecutándose.
- `SIGHUP` recarga la configuración.
- Una recarga inválida no detiene el supervisor.

### Commits sugeridos

```text
✨ feat(signals): manejar sigint y sigterm
✨ feat(supervisor): implementar apagado ordenado
✨ feat(config): recargar configuración con sighup
✅ test(signals): probar cancelación y cierre limpio
📝 docs(parte5): documentar señales y apagado
```

---

## PARTE 6 — API HTTP, pruebas finales y documentación

**Integrante:** Miembro 6

### Objetivo

Agregar control externo y preparar la entrega final.

### Tareas

1. Crear servidor HTTP local.
2. Implementar `/health`.
3. Implementar listado y detalle de procesos.
4. Implementar start, stop y restart.
5. Implementar reload.
6. Validar nombres inexistentes.
7. Manejar respuestas JSON y códigos HTTP.
8. Crear pruebas con `httptest`.
9. Integrar comandos de build y test.
10. Crear `Makefile`.
11. Completar `README.md`.
12. Crear diagramas en `GRAFICOS/`.
13. Revisar `gofmt`.
14. Ejecutar `go vet ./...`.
15. Ejecutar `go test ./...`.
16. Ejecutar `go test -race ./...`.
17. Preparar demostración.
18. Crear `Documentacion-parte6.md`.

### Al terminar

- API local funcional.
- Se puede consultar y controlar procesos.
- Todos los tests pasan.
- No hay carreras detectadas.
- README completo.
- Bitácora y declaración de IA completas.
- Proyecto preparado para defensa.

### Commits sugeridos

```text
✨ feat(api): agregar consulta de estados por http
✨ feat(api): agregar start stop y restart
✅ test(api): probar endpoints con httptest
📝 docs(readme): agregar instalación uso y demostración
📝 docs(parte6): documentar integración final
```

---

# 17. Orden de integración

```text
P1 Fundación
      ↓
P2 Ejecución de procesos
      ↓
P3 Reinicios y backoff
      ↓
P4 Concurrencia y estados
      ↓
P5 Señales y recarga
      ↓
P6 API y cierre final
```

Dependencias:

- P1 bloquea a todas las demás.
- P2 es necesaria para P3, P4 y P5.
- P3 y P4 deben integrarse cuidadosamente.
- P5 depende del modelo de cancelación definido en P4.
- P6 se integra cuando el supervisor ya es estable.

---

# 18. Convención de ramas

```text
main
├── parte1/configuracion
├── parte2/ejecucion-procesos
├── parte3/reinicio-backoff
├── parte4/concurrencia-estados
├── parte5/senales-apagado
└── parte6/api-documentacion
```

Flujo:

1. Cada integrante actualiza su rama desde `main`.
2. Trabaja únicamente en su parte.
3. Realiza commits pequeños.
4. Ejecuta pruebas.
5. Crea Pull Request.
6. Otro integrante revisa.
7. Se corrigen observaciones.
8. Se aprueba y fusiona.

---

# 19. Convención de commits

Formato:

```text
emoji tipo(scope): descripción
```

| Emoji | Tipo | Uso |
|---|---|---|
| ✨ | `feat` | Nueva funcionalidad |
| 🐛 | `fix` | Corrección |
| ♻️ | `refactor` | Mejora interna |
| ✅ | `test` | Pruebas |
| 📝 | `docs` | Documentación |
| 🔧 | `config` | Configuración |
| 🔒 | `security` | Seguridad |
| 🚀 | `build` | Compilación o automatización |

Ejemplos:

```text
✨ feat(process): ejecutar procesos con os exec
🐛 fix(restart): detener reintentos al cancelar contexto
✅ test(supervisor): validar cierre de goroutines
📝 docs(parte4): explicar modelo de concurrencia
```

Reglas:

- Descripción en español.
- Minúsculas.
- Sin punto final.
- Máximo 72 caracteres.
- Un commit debe representar una sola intención.
- No usar mensajes como `cambios`, `avance`, `listo` o `final`.

---

# 20. Estándares de código Go

## Formato

```bash
gofmt -w .
```

## Análisis estático

```bash
go vet ./...
```

## Pruebas

```bash
go test ./...
go test -race ./...
```

## Errores

- No usar `panic` en flujos normales.
- Envolver errores con contexto:

```go
return fmt.Errorf("iniciar proceso %q: %w", name, err)
```

## Context

- Toda operación bloqueante debe poder cancelarse.
- No guardar context dentro de estructuras permanentes.
- Pasar `context.Context` como primer parámetro.

## Interfaces

Crear interfaces pequeñas:

```go
type Runner interface {
    Start(ctx context.Context, cfg config.ProcessConfig) (*RunningProcess, error)
}
```

No crear interfaces grandes con demasiadas responsabilidades.

## Logs

Cada mensaje debe identificar el proceso:

```text
[worker-estable] proceso iniciado pid=4242
[worker-falla] salida detectada exit_code=1
```

## Comentarios

Los comentarios explican el porqué, no repiten el código.

```go
// El timer se crea por intento para permitir cancelarlo inmediatamente
// cuando el supervisor recibe una señal de apagado.
timer := time.NewTimer(delay)
```

## Recursos

- Detener timers.
- Cerrar servidores.
- Esperar procesos.
- Cerrar channels solo desde su propietario.
- No cerrar un channel desde el receptor.

---

# 21. Reglas técnicas obligatorias

1. Los nombres de procesos deben ser únicos.
2. No se permite un comando vacío.
3. Una política desconocida invalida la configuración.
4. El supervisor no debe reiniciar procesos durante el apagado.
5. Todo proceso iniciado debe ser esperado con `Wait()`.
6. Ninguna goroutine puede quedar sin ruta de salida.
7. El backoff debe ser cancelable.
8. La API debe escuchar solo en localhost por defecto.
9. Una recarga inválida conserva la configuración anterior.
10. El estado visible debe provenir de snapshots seguros.
11. Todo acceso concurrente debe estar coordinado.
12. El proyecto debe pasar el detector de carreras.
13. El shutdown debe esperar a los procesos hijos.
14. No se deben registrar secretos ni variables sensibles.
15. No se debe usar `exec.Command("sh", "-c", inputUsuario)` con entrada no validada.

---

# 22. Pruebas mínimas

## Configuración

- Archivo correcto.
- JSON inválido.
- Proceso sin nombre.
- Nombre duplicado.
- Política inválida.
- Backoff inválido.

## Ejecución

- Proceso exit 0.
- Proceso exit distinto de 0.
- Comando inexistente.
- Captura de stdout.
- Captura de stderr.
- Cancelación.

## Reinicio

- `always`.
- `on-failure`.
- `never`.
- Máximo de reintentos.
- Tope del backoff.
- Cancelación durante backoff.

## Supervisor

- Varios procesos simultáneos.
- Estados correctos.
- Stop manual.
- Restart manual.
- Apagado global.
- Sin carreras.

## Señales

- SIGTERM.
- SIGINT.
- Grace period.
- Proceso que no responde.
- Recarga válida.
- Recarga inválida.

## API

- Health.
- Lista.
- Detalle.
- Proceso inexistente.
- Start.
- Stop.
- Restart.
- Reload.

---

# 23. Comandos obligatorios de calidad

Todos deben ejecutarse desde la raíz del repositorio y quedar registrados como evidencia.

```bash
go mod tidy
gofmt -w .
go build ./...
go vet ./...
go test ./...
go test -race ./...
```

### Construcción del binario entregable

Linux:

```bash
mkdir -p bin
go build -o bin/supervisor ./cmd/supervisor
```

Windows, únicamente para desarrollo:

```powershell
New-Item -ItemType Directory -Force bin
go build -o bin/supervisor.exe ./cmd/supervisor
```

La defensa principal debe realizarse en Linux porque el proyecto usa señales del sistema como `SIGHUP`, `SIGTERM` y `SIGINT`.

### Ejecución funcional

```bash
./bin/supervisor validate --config examples/config.example.json
./bin/supervisor start --config examples/config.example.json
```

Durante el desarrollo también se permite:

```bash
go run ./cmd/supervisor validate --config examples/config.example.json
go run ./cmd/supervisor start --config examples/config.example.json
```

### Comandos con Makefile

```bash
make fmt
make build
make vet
make test
make race
make run
make demo
```

## Evidencias obligatorias

Crear la carpeta:

```text
EVIDENCIAS/
├── build.txt
├── vet.txt
├── test.txt
├── race.txt
├── demo-reinicio.txt
├── demo-shutdown.txt
├── commits.png
└── pull-requests.png
```

Generación sugerida:

```bash
mkdir -p EVIDENCIAS
go build ./... 2>&1 | tee EVIDENCIAS/build.txt
go vet ./... 2>&1 | tee EVIDENCIAS/vet.txt
go test ./... 2>&1 | tee EVIDENCIAS/test.txt
go test -race ./... 2>&1 | tee EVIDENCIAS/race.txt
```

`gofmt -w .` debe ejecutarse antes de la entrega. Después de aplicarlo, `git status` no debe mostrar cambios de formato pendientes.

---

# 24. Entregables finales obligatorios

La entrega se considera incompleta si falta cualquiera de los elementos comunes exigidos por la guía.

## Repositorio y código

- Repositorio Git con historia incremental, coherente y verificable.
- Commits pequeños, frecuentes y descriptivos.
- Pull Requests revisados antes de fusionar a `main`.
- Código fuente completo en Go.
- Código formateado con `gofmt`.
- `go build ./...` exitoso.
- `go vet ./...` sin observaciones.
- `go test ./...` exitoso.
- `go test -race ./...` exitoso.
- Binario ejecutable:
  - `bin/supervisor` en Linux.
  - `bin/supervisor.exe` solo como apoyo en Windows.

## Archivos funcionales

- `examples/config.example.json`.
- Varios procesos de ejemplo.
- Worker estable.
- Worker que termina con éxito.
- Worker que falla.
- Worker que ignora inicialmente la señal de terminación para probar el periodo de gracia.
- Scripts o comandos reproducibles para la demostración.

## Documentación

- `README.md` con:
  - descripción;
  - requisitos;
  - compilación;
  - configuración;
  - ejecución;
  - API;
  - políticas;
  - señales;
  - pruebas;
  - demostración;
  - limitaciones.
- Documento de la máquina de estados.
- Documento de políticas de reinicio y backoff:
  - valor inicial;
  - factor;
  - tope;
  - máximo de reintentos.
- Explicación del modelo de concurrencia:
  - channels o mutex;
  - propietario del estado;
  - cancelación;
  - cierre de goroutines.
- `Documentacion-parte1.md` hasta `Documentacion-parte6.md`.
- Diagramas y capturas en `GRAFICOS/`.
- Evidencias en `EVIDENCIAS/`.

## Autoría

- `BITACORA_DECISIONES.md` completada durante el desarrollo.
- `DECLARACION_USO_IA.md` completada y firmada por los integrantes.
- Defensa oral de 15 a 20 minutos.
- Modificación de código en vivo.
- Todos los integrantes deben comprender el código que presentan.

---

# 25. Historia Git obligatoria

No se permite entregar el sistema mediante un único commit final.

El repositorio debe mostrar:

- Inicio del módulo.
- Configuración y validación.
- Ejecución de procesos.
- Captura de salida.
- Políticas de reinicio.
- Backoff.
- Máquina de estados.
- Concurrencia.
- Señales.
- Apagado.
- API.
- Pruebas.
- Documentación.

Ejemplo de progresión creíble:

```text
🔧 config: inicializar módulo y estructura
✨ feat(config): cargar configuración json
✅ test(config): validar errores de configuración
✨ feat(process): iniciar y esperar procesos
✨ feat(logging): capturar stdout y stderr
✨ feat(restart): agregar políticas de reinicio
✨ feat(restart): implementar backoff exponencial
✨ feat(supervisor): coordinar managers concurrentes
✨ feat(signals): implementar apagado ordenado
✨ feat(config): recargar mediante sighup
✨ feat(api): exponer estado y controles
✅ test(supervisor): verificar cierre y detector de carreras
📝 docs(readme): documentar compilación y demostración
```

Reglas:

1. No reescribir artificialmente la historia antes de entregar.
2. No compartir una sola cuenta de GitHub.
3. Cada integrante debe usar su propia cuenta.
4. Cada Pull Request debe describir lo implementado y cómo se verificó.
5. La persona que creó el código no debe aprobar su propio Pull Request cuando se requiera una aprobación.
6. Un commit de corrección debe explicar el error resuelto.
7. Los commits deben corresponder con decisiones registradas en la bitácora.

---

# 26. Bitácora obligatoria por hito

La bitácora debe escribirse a medida que avanza el proyecto, no reconstruirse únicamente al final.

Debe existir al menos una decisión relevante por cada hito oficial:

- **H1:** formato de configuración y estrategia de ejecución.
- **H2:** goroutine por proceso y evaluación de políticas.
- **H3:** fórmula, cancelación y límites del backoff.
- **H4:** propagación de señales, grace period y cierre.
- **H5:** diseño de API o socket de control.

Cada entrada debe incluir:

| Campo | Contenido obligatorio |
|---|---|
| Fecha | Día en que se tomó la decisión |
| Hito | H1, H2, H3, H4 o H5 |
| Parte/integrante | Responsable |
| Problema | Qué debía resolverse |
| Decisión | Alternativa elegida |
| Alternativas evaluadas | Opciones consideradas |
| Justificación técnica | Por qué se eligió |
| Consecuencia | Ventaja, costo o limitación |
| Evidencia | Commit, PR, prueba o archivo |
| Uso de IA | Consulta realizada y decisión de aceptar, adaptar o descartar |

### Plantilla de entrada

```markdown
## Decisión H3-01 — Coordinación del backoff

**Fecha:**  
**Integrante:**  
**Problema:**  
**Decisión:**  
**Alternativas consideradas:**  
**Justificación técnica:**  
**Consecuencias y limitaciones:**  
**Evidencia (commit/PR/prueba):**  
**Uso de IA:**  
```

Decisiones mínimas que deben aparecer:

- Channels frente a mutex.
- Propietario de cada channel.
- Estrategia para consultar snapshots.
- Uso de `exec.Command` o `exec.CommandContext`.
- Forma de obtener el exit code.
- Política para detener procesos.
- Grace period y uso de `SIGKILL`.
- Estrategia para evitar reinicios durante shutdown.
- Recarga de configuración.
- Prevención de goroutines filtradas.

---

# 27. Declaración obligatoria de uso de IA

El archivo `DECLARACION_USO_IA.md` debe incluir, por integrante:

```markdown
# Declaración de uso de inteligencia artificial

## Datos del integrante

**Nombre completo:**  
**Parte desarrollada:**  
**Fecha:**  

## Herramientas utilizadas

Indicar el nombre de cada asistente o herramienta de IA utilizada.

## Finalidad del uso

Marcar y explicar:

- [ ] Consulta de conceptos
- [ ] Explicación de errores
- [ ] Discusión de alternativas de diseño
- [ ] Generación o adaptación de pruebas
- [ ] Revisión de estilo
- [ ] Revisión de código ya escrito
- [ ] Otro:

## Archivos o módulos influenciados

Indicar exactamente en qué archivos influyó el uso de IA y de qué manera.

## Partes de autoría íntegra

Indicar qué diseño y qué lógica fueron desarrollados directamente por el integrante.

## Validación de comprensión

Explicar cómo verificó, probó y adaptó las respuestas recibidas.

## Declaración

Declaro que soy autor del diseño y de la lógica central correspondiente
a mi participación en este proyecto, que comprendo todo el código que
presento y que puedo explicarlo y modificarlo durante la defensa oral.

**Firma o nombre completo:**  
```

Declarar honestamente el uso responsable de IA no penaliza. Ocultarlo o usarlo para sustituir la autoría sí puede generar la penalización establecida por la institución.

---

# 28. Defensa oral obligatoria

## Duración

```text
15 a 20 minutos
```

## Contenido mínimo

1. Problema que resuelve el supervisor.
2. Arquitectura y responsabilidades.
3. Lectura y validación de configuración.
4. Ejecución con `os/exec`.
5. Captura de `stdout` y `stderr`.
6. Detección de salida y exit code.
7. Políticas `always`, `on-failure` y `never`.
8. Backoff exponencial y máximo de reintentos.
9. Máquina de estados.
10. Modelo concurrente.
11. Uso de `context`.
12. Manejo de `SIGINT`, `SIGTERM` y `SIGHUP`.
13. Apagado ordenado.
14. API HTTP.
15. Pruebas y detector de carreras.
16. Bitácora y declaración de IA.

## Demostraciones obligatorias

- Iniciar varios procesos.
- Mostrar logs identificados.
- Mostrar un proceso que termina correctamente.
- Mostrar un proceso que falla.
- Demostrar las políticas de reinicio.
- Demostrar esperas progresivas del backoff.
- Mostrar el estado `failed` después del máximo de reintentos.
- Consultar estados por API.
- Ejecutar start/stop/restart.
- Recargar configuración con `SIGHUP`.
- Detener con `Ctrl+C`.
- Comprobar que no quedan procesos hijos.
- Ejecutar `go test -race ./...`.

## Modificación en vivo

El docente podrá solicitar una modificación pequeña, por ejemplo:

- Agregar el endpoint `/version`.
- Cambiar el máximo del backoff.
- Agregar el estado `paused`.
- Mostrar el tiempo de ejecución.
- Agregar una política `unless-stopped`.
- Cambiar el grace period.
- Agregar un campo a la respuesta de estado.
- Incorporar un nuevo proceso en la configuración.

Cada integrante debe poder realizar cambios en los archivos que afirma haber desarrollado.

## Preguntas probables

- ¿Por qué se eligieron channels o mutex?
- ¿Quién escribe y quién cierra cada channel?
- ¿Qué ocurre si un proceso falla al iniciar?
- ¿Cómo se obtiene el código de salida?
- ¿Cómo se evita un proceso zombi?
- ¿Cómo se evita una tormenta de reinicios?
- ¿Cómo se cancela un timer de backoff?
- ¿Qué pasa si llega SIGTERM durante el backoff?
- ¿Cómo se impide una condición de carrera?
- ¿Qué detecta `go test -race`?
- ¿Cómo se demuestra que no hay goroutines filtradas?
- ¿Qué ocurre si la nueva configuración es inválida?
- ¿Por qué la API escucha en localhost?

---

# 29. Demostración final sugerida

1. Mostrar el historial de commits y Pull Requests.
2. Ejecutar:

```bash
gofmt -w .
go build ./...
go vet ./...
go test ./...
go test -race ./...
```

3. Construir el binario:

```bash
go build -o bin/supervisor ./cmd/supervisor
```

4. Validar la configuración.
5. Iniciar al menos dos procesos.
6. Mostrar PID, stdout y stderr.
7. Hacer fallar un worker.
8. Mostrar backoff de 1, 2 y 4 segundos.
9. Mostrar límite de reintentos y estado `failed`.
10. Consultar `/processes`.
11. Detener y reiniciar por API.
12. Modificar la configuración.
13. Enviar `SIGHUP`.
14. Mostrar que una recarga inválida conserva la configuración anterior.
15. Presionar `Ctrl+C`.
16. Mostrar el periodo de gracia.
17. Comprobar con `ps` que no quedan procesos hijos.
18. Mostrar bitácora y declaración de IA.
19. Realizar el cambio solicitado en vivo.

---

# 30. Matriz de cumplimiento de la rúbrica oficial

| Criterio | Peso | Evidencia para nivel Competente |
|---|---:|---|
| Gestión de procesos | **20 %** | Arranque, logs, detección de terminación y reinicio fiables mediante `os/exec` |
| Concurrencia segura | **25 %** | `go test -race` aprobado; estado coordinado; ausencia de deadlocks y goroutines filtradas |
| Señales y cierre limpio | **20 %** | SIGTERM/SIGINT, periodo de gracia, espera de hijos, ausencia de zombis y recarga con SIGHUP |
| Reinicio, backoff y estados | **15 %** | Tres políticas correctas; backoff exponencial con tope; estados claros; sin tormenta de reinicios |
| Calidad y pruebas | **10 %** | Código idiomático, `gofmt`, `go vet`, pruebas de reinicio, backoff y cancelación |
| Defensa y autoría | **10 %** | Explicación del modelo, flujo de cancelación, ejecución de `-race`, cambio en vivo y bitácora coherente |
| **Total** | **100 %** | Cumplimiento conjunto de H1 a H5 y entregables comunes |

## Condiciones para aspirar a Sobresaliente

Además del nivel Competente:

- Casos borde robustos.
- Errores claros y contextualizados.
- Pruebas deterministas y amplias.
- API bien diseñada.
- Recarga diferencial segura.
- Métricas o scheduling tipo cron opcional.
- Explicación técnica sólida durante la defensa.
- Modificación en vivo correcta sin romper pruebas.

---

# 31. Criterio de proyecto terminado

El proyecto solo se considera terminado cuando:

- [ ] Se cumplen H1, H2, H3, H4 y H5.
- [ ] Compila con `go build ./...`.
- [ ] Existe `bin/supervisor`.
- [ ] `gofmt` no deja cambios pendientes.
- [ ] `go vet ./...` no reporta observaciones.
- [ ] `go test ./...` pasa.
- [ ] `go test -race ./...` pasa.
- [ ] Administra varios procesos.
- [ ] Captura stdout y stderr.
- [ ] Aplica las tres políticas.
- [ ] Aplica backoff exponencial con tope.
- [ ] Mantiene estados correctos.
- [ ] Responde a SIGINT y SIGTERM.
- [ ] Recarga con SIGHUP.
- [ ] Se apaga sin procesos zombis.
- [ ] Se apaga sin goroutines filtradas.
- [ ] La API permite status/start/stop/restart.
- [ ] La configuración de ejemplo funciona.
- [ ] README permite reproducir la instalación y la demo.
- [ ] La historia Git es incremental y creíble.
- [ ] La bitácora tiene entradas por hito.
- [ ] La declaración de IA está completa y firmada.
- [ ] Las evidencias están guardadas.
- [ ] La defensa dura 15–20 minutos.
- [ ] El equipo realiza una modificación en vivo.
- [ ] Cada integrante explica y modifica su propia parte.

---

> **Nombre oficial:** **Go Process Supervisor** — Supervisor concurrente de procesos y planificador de tareas desarrollado en Go.
