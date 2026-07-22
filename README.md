# Go Process Supervisor 🚀

> **Supervisor de procesos y Job Scheduler concurrente desarrollado en Go.**

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Estado-Parte%202%20Completada-success)](docs/Documentacion-parte2.md)

---

## 📌 Descripción del Proyecto

**Go Process Supervisor** es un sistema de gestión del ciclo de vida de procesos externos en segundo plano. Permite declarar listas de ejecutables en formato JSON, supervisar su ejecución en paralelo, capturar de forma síncrona/asíncrona sus salidas `stdout` y `stderr` con formateo etiquetado, aplicar políticas de auto-reinicio con retardo exponencial (*exponential backoff*), y garantizar un apagado limpio (*graceful shutdown*) previniendo la acumulación de **procesos zombis** en el sistema operativo.

---

## 🚀 Estado del Avance por Partes

- [x] **Parte 1 — Fundación y Configuración:** Modelo declarativo (`AppConfig`, `ProcessConfig`), des-serializador JSON, validador de negocio y subcomando CLI `validate`.
- [x] **Parte 2 — Ejecución de Procesos y Logs:** Ejecutor `ProcessRunner` con `os/exec`, captura de `stdout`/`stderr` etiquetados, recolección de PID, control de códigos de salida y prevención de procesos zombis.
- [x] **Parte 3 — Monitoreo, Reinicios y Backoff:** Políticas `always`, `on-failure`, `never`, contadores de reintentos y cálculo exponencial.
- [ ] **Parte 4 — Supervisor Concurrente:** Máquina de estados segura por goroutines, canal de eventos y sincronización thread-safe.
- [ ] **Parte 5 — Manejo de Señales y Recarga:** Captura de `SIGINT`, `SIGTERM`, `SIGHUP` con periodo de gracia y *hot-reload*.
- [ ] **Parte 6 — API HTTP Local & Dashboard:** Endpoints REST (`/health`, `/processes`), panel web interactivo y entregables finales.

---

## 💻 Guía de Uso Rápido

### Requisitos Previos
* **Go** 1.21 o superior instalado.

### 1. Compilación y Verificación
```bash
# Formatear todo el código según el estándar Go
gofmt -w .

# Compilar todos los paquetes
go build ./...

# Ejecutar pruebas unitarias
go test -v ./...
```

### 2. Comandos CLI Disponibles

#### A. Validar un archivo de configuración (`validate`)
Verifica la sintaxis del JSON y cumple las reglas de negocio (nombres únicos, comandos requeridos, tiempos válidos) sin lanzar ningún proceso hijo:
```bash
go run ./cmd/supervisor validate --config examples/config.example.json
```

#### B. Ejecutar procesos con logs etiquetados en vivo (`run`)
Ejecuta de forma inmediata los procesos definidos en la configuración y muestra sus salidas formateadas en tiempo real:

```bash
# En Windows (PowerShell / CMD):
go run ./cmd/supervisor run --config examples/config.windows.json

# En Linux / macOS / POSIX:
go run ./cmd/supervisor run --config examples/config.example.json
```

---

## 🖥️ Ejemplo de Salida en Consola

```text
[18:11:54] [SUPERVISOR] [INFO] Iniciando ejecución de 2 proceso(s) desde examples/config.windows.json...
[18:11:54] [worker-falla] [INFO] proceso iniciado correctamente con PID 25112
[18:11:54] [worker-estable] [INFO] proceso iniciado correctamente con PID 19992
[18:11:54] [worker-falla] [STDOUT] [worker-falla] Simulando fallo... 
[18:11:54] [worker-falla] [INFO] proceso finalizó con código de salida 1
[18:11:54] [worker-falla] [INFO] Resumen de ejecución: PID=25112 | ExitCode=1 | Duración=29.9ms
[18:11:54] [worker-estable] [STDOUT] [worker-estable] Iniciando tarea... 
[18:11:55] [worker-estable] [STDOUT] [worker-estable] Tarea completada con exito.
[18:11:55] [worker-estable] [INFO] proceso finalizó exitosamente (exit code 0)
[18:11:55] [worker-estable] [INFO] Resumen de ejecución: PID=19992 | ExitCode=0 | Duración=1.05s
[18:11:55] [SUPERVISOR] [INFO] Todos los procesos han finalizado la ejecución de demostración de la Parte 2.
```

---

## 📁 Estructura del Proyecto

```text
go-process-supervisor/
├── cmd/
│   └── supervisor/
│       └── main.go               # Punto de entrada CLI (subcomandos validate, run, version)
├── internal/
│   ├── config/                   # Carga, validación y modelos JSON
│   ├── process/                  # Ejecutor de procesos (os/exec, PID, Wait, ExitCode)
│   ├── logging/                  # Formateador thread-safe de logs etiquetados
│   ├── restart/                  # Políticas de reinicio y retardo exponencial (Backoff)
│   ├── supervisor/               # (Parte 4) Coordinación concurrente y máquina de estados
│   ├── signals/                  # (Parte 5) Manejador de SIGINT, SIGTERM, SIGHUP
│   └── api/                      # (Parte 6) Servidor y endpoints HTTP
├── docs/
│   ├── partes/
│   │   ├── Documentacion-parte1.md   # Especificación técnica de la Parte 1
│   │   ├── Documentacion-parte2.md   # Especificación técnica de la Parte 2
│   │   └── Documentacion-parte3.md   # Especificación técnica de la Parte 3
│   ├── graficos/
│   │   └── arquitectura_y_flujos.md  # Diagramas de Arquitectura y Máquina de Estados
│   └── evidencias/
│       └── evidencias_hito1_hito2.md # Registros de pruebas y ejecución en vivo de los Hitos 1 y 2
├── examples/
│   ├── config.example.json       # Configuración de ejemplo para POSIX
│   ├── config.windows.json       # Configuración de ejemplo para Windows
│   └── workers/                  # Scripts ejecutables de prueba
├── BITACORA_DECISIONES.md        # Registro de decisiones de diseño del equipo
├── Makefile                      # Automatización de tareas
├── go.mod                        # Módulo Go
└── README.md                     # Documentación general del proyecto
```

---

## 📚 Documentación Técnica Detallada

* 📄 [Documentación de la Parte 1 — Fundación y Configuración](docs/partes/Documentacion-parte1.md)
* 📄 [Documentación de la Parte 2 — Ejecución de Procesos y Logs](docs/partes/Documentacion-parte2.md)
* 📄 [Documentación de la Parte 3 — Monitoreo, Reinicios y Backoff](docs/partes/Documentacion-parte3.md)
* 📊 [Diagramas de Arquitectura y Flujos](docs/graficos/arquitectura_y_flujos.md)
* 📸 [Evidencias de Ejecución y Pruebas](docs/evidencias/evidencias_hito1_hito2.md)
* 📓 [Bitácora de Decisiones Técnicas](BITACORA_DECISIONES.md)

---

## 👥 Equipo de Desarrollo — Grupo 7

* **Carrasco Millan Jose Manuel**
* **Silupu Becerra Nilson Jesus**
* **Vidarte Cruz Jose Junior**
* **Segundo Arteaga Karen Milenka**

---

## 📄 Licencia

Este proyecto está distribuido bajo la Licencia [MIT](LICENSE) © 2026 Grupo 7.

