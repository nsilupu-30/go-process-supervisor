# Go Process Supervisor

> **Supervisor de procesos concurrente desarrollado en Go.**

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Estado-Parte%206%20Completada-success)](docs/partes/Documentacion-parte6.md)

---

## 📌 Descripción del Proyecto

**Go Process Supervisor** es un sistema de gestión del ciclo de vida de procesos externos. Permite declarar ejecutables en JSON, supervisar su ejecución en paralelo, capturar `stdout` y `stderr` con formateo etiquetado, aplicar políticas de auto-reinicio con retardo exponencial y garantizar un apagado limpio previniendo procesos zombis.

Ahora incluye coordinación concurrente (`Parte 4`) y manejo de señales del sistema con recarga dinámica (`Parte 5`).

---

## 🚀 Estado del Avance por Partes

- [x] **Parte 1 — Fundación y Configuración:** Modelo declarativo (`AppConfig`, `ProcessConfig`), des-serializador JSON, validador de negocio y subcomando CLI `validate`.
- [x] **Parte 2 — Ejecución de Procesos y Logs:** Ejecutor `ProcessRunner` con `os/exec`, captura de `stdout`/`stderr` etiquetados, recolección de PID, control de códigos de salida y prevención de procesos zombis.
- [x] **Parte 3 — Monitoreo, Reinicios y Backoff:** Políticas `always`, `on-failure`, `never`, contadores de reintentos y cálculo exponencial.
- [x] **Parte 4 — Supervisor Concurrente:** Máquina de estados segura por goroutines, canal de comandos y sincronización thread-safe.
- [x] **Parte 5 — Manejo de Señales y Recarga:** Captura de `SIGINT`, `SIGTERM`, `SIGHUP` con periodo de gracia y recarga dinámica.
- [x] **Parte 6 — API HTTP Local & Dashboard:** Endpoints REST (`/health`, `/processes`) y entregables finales.

---

## 💻 Guía de Uso Rápido

### Requisitos Previos
* **Go** 1.21 o superior instalado.

### 1. Compilación y Verificación
```bash
make build
make test
make race
```

### 2. Comandos CLI Disponibles

#### A. Validar configuración (`validate`)
```bash
go run ./cmd/supervisor validate --config examples/config.example.json
```

#### B. Ejecutar procesos (`run`)
```bash
# Windows
go run ./cmd/supervisor run --config examples/config.windows.json

# Linux / macOS
go run ./cmd/supervisor run --config examples/config.example.json
```

#### C. Apagado ordenado
Presionar `Ctrl+C` envía `SIGINT`. El supervisor propaga la señal a los procesos hijos, espera el `grace_period_seconds` y fuerza `SIGKILL` si es necesario.

#### D. Recarga dinámica (`SIGHUP`)
Enviar `SIGHUP` al proceso del supervisor recarga la configuración sin reiniciar los procesos en ejecución. Si la nueva configuración es inválida, se conserva la anterior.

#### E. API HTTP local
El supervisor expone un servidor HTTP en la dirección configurada en `api_address`.

Ejemplos:
```bash
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/processes
curl http://127.0.0.1:8080/processes/<name>
curl -X POST http://127.0.0.1:8080/processes/<name>/restart
curl -X POST http://127.0.0.1:8080/reload
```

---

## 🖥️ Ejemplo de Salida en Consola

```text
[18:11:54] [SUPERVISOR] [INFO] Iniciando ejecución de 2 proceso(s) desde examples/config.example.json...
[18:11:54] [worker-estable] [INFO] proceso iniciado correctamente con PID 19992
[18:11:54] [worker-falla] [INFO] proceso iniciado correctamente con PID 25112
[18:11:54] [worker-estable] [STDOUT] Iniciando tarea...
[18:11:54] [worker-falla] [STDOUT] Simulando fallo...
[18:11:55] [worker-falla] [INFO] proceso finalizó con código de salida 1
[18:11:55] [worker-falla] [INFO] reinicio programado en 1s
[18:11:55] [worker-estable] [STDOUT] Tarea completada con éxito.
[18:11:56] [worker-estable] [INFO] proceso finalizó exitosamente (exit code 0)
[18:11:56] [worker-falla] [INFO] proceso iniciado correctamente con PID 28440
[18:12:00] [SUPERVISOR] [INFO] Recibida señal de apagado, propagando a procesos hijos...
[18:12:00] [worker-estable] [INFO] proceso detenido correctamente
[18:12:00] [worker-falla] [INFO] proceso detenido correctamente
[18:12:00] [SUPERVISOR] [INFO] Todos los procesos han finalizado. Apagado completo.
```

---

## 📁 Estructura del Proyecto

```text
go-process-supervisor/
├── cmd/supervisor/main.go          # CLI: validate, run, version
├── internal/
│   ├── config/                     # Carga y validación JSON
│   ├── process/                    # ProcessRunner + ExecutionResult
│   ├── logging/                    # Logger thread-safe etiquetado
│   ├── restart/                    # Políticas + backoff exponencial
│   ├── supervisor/                 # Supervisor, AdministradorProceso, estados
│   └── senales/                    # SIGINT, SIGTERM, SIGHUP + recarga
├── docs/partes/
│   ├── Documentacion-parte1.md
│   ├── Documentacion-parte2.md
│   ├── Documentacion-parte3.md
│   ├── Documentacion-parte4.md
│   └── Documentacion-parte5.md
├── examples/
│   ├── config.example.json
│   ├── config.windows.json
│   └── workers/
├── Makefile
├── go.mod
└── README.md
```

---

## 📚 Documentación Técnica

* 📄 [Parte 1 — Fundación y Configuración](docs/partes/Documentacion-parte1.md)
* 📄 [Parte 2 — Ejecución de Procesos y Logs](docs/partes/Documentacion-parte2.md)
* 📄 [Parte 3 — Monitoreo, Reinicios y Backoff](docs/partes/Documentacion-parte3.md)
* 📄 [Parte 4 — Supervisor Concurrente](docs/partes/Documentacion-parte4.md)
* 📄 [Parte 5 — Señales y Recarga Dinámica](docs/partes/Documentacion-parte5.md)
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
