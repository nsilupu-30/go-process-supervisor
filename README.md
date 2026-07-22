# Go Process Supervisor

Supervisor de procesos y job scheduler concurrentes en Go.

## Descripción del Proyecto
Herramienta de supervisión de procesos de sistema capaz de gestionar el ciclo de vida de ejecutables mediante archivos de configuración declarativos en JSON, políticas de reinicio automático (`always`, `on-failure`, `never`), retardo exponencial (*exponential backoff*), y apagado ordenado ante señales de sistema (`SIGINT`, `SIGTERM`, `SIGHUP`).

---

## Estado del Proyecto — Parte 1 Completada
- [x] **Parte 1:** Fundación, estructura y configuración (`internal/config`, comando `validate`)
- [ ] **Parte 2:** Ejecución de procesos y captura de logs (`internal/process`)
- [ ] **Parte 3:** Monitoreo, reinicios y backoff exponencial (`internal/restart`)
- [ ] **Parte 4:** Supervisor concurrente y máquina de estados (`internal/supervisor`)
- [ ] **Parte 5:** Manejo de señales y apagado ordenado (`internal/signals`)
- [ ] **Parte 6:** API HTTP local y documentación final (`internal/api`)

---

## Instrucciones de Uso (Parte 1)

### Requisitos
- Go 1.21 o superior.

### Compilación y Verificación
```bash
# Formatear código
gofmt -w .

# Actualizar y verificar dependencias
go mod tidy

# Compilar el proyecto
go build ./...

# Análisis estático
go vet ./...

# Pruebas unitarias
go test ./...

# Pruebas con detector de condiciones de carrera
go test -race ./...
```

### Probar Comando CLI `validate`

1. **Validación exitosa (archivo de ejemplo válido):**
```bash
go run ./cmd/supervisor validate --config examples/config.example.json
```
*Salida esperada:*
```text
✓ Configuración válida: examples/config.example.json (2 procesos configurados)
```

2. **Validación fallida (ejemplo con parámetro inválido):**
Si el archivo contiene errores de sintaxis JSON o incumple alguna regla de negocio (ej. nombres duplicados o comandos vacíos), se muestra un mensaje explicativo y finaliza con código de salida distinto de cero (`exit status 1`).

---

## Estructura del Repositorio

```text
go-process-supervisor/
├── cmd/
│   └── supervisor/
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   ├── loader.go
│   │   ├── validator.go
│   │   └── loader_test.go
│   ├── process/
│   ├── restart/
│   ├── supervisor/
│   ├── signals/
│   ├── api/
│   └── logging/
├── examples/
│   ├── config.example.json
│   └── workers/
│       ├── stable-worker.sh
│       └── failing-worker.sh
├── Documentacion-parte1.md
├── BITACORA_DECISIONES.md
├── DECLARACION_USO_IA.md
├── Makefile
├── go.mod
└── .gitignore
```
