# Evidencias de Ejecución y Pruebas — Hitos 1 y 2

Este archivo contiene los registros de pruebas automatizadas y evidencias de ejecución en vivo del **Go Process Supervisor**.

---

## 📸 Evidencia 1: Validación de Configuración (Hito 1 / Parte 1)

### Comando Ejecutado:
```powershell
go run ./cmd/supervisor validate --config examples/config.example.json
```

### Resultado de Salida:
```text
✓ Configuración válida: examples/config.example.json (2 procesos configurados)
```

---

## 📸 Evidencia 2: Ejecución de Procesos con Logs Etiquetados (Hito 2 / Parte 2)

### Comando Ejecutado:
```powershell
go run ./cmd/supervisor run --config examples/config.windows.json
```

### Registro en Vivo de Consola:
```text
[18:11:54] [SUPERVISOR] [INFO] Iniciando ejecución de 2 proceso(s) desde examples/config.windows.json...
[18:11:54] [worker-falla] [INFO] proceso iniciado correctamente con PID 25112
[18:11:54] [worker-estable] [INFO] proceso iniciado correctamente con PID 19992
[18:11:54] [worker-falla] [STDOUT] [worker-falla] Simulando fallo... 
[18:11:54] [worker-falla] [INFO] proceso finalizó con código de salida 1
[18:11:54] [worker-falla] [INFO] Resumen de ejecución: PID=25112 | ExitCode=1 | Duración=29.9046ms
[18:11:54] [worker-estable] [STDOUT] [worker-estable] Iniciando tarea... 
[18:11:55] [worker-estable] [STDOUT] [worker-estable] Tarea completada con exito.
[18:11:55] [worker-estable] [INFO] proceso finalizó exitosamente (exit code 0)
[18:11:55] [worker-estable] [INFO] Resumen de ejecución: PID=19992 | ExitCode=0 | Duración=1.0575002s
[18:11:55] [SUPERVISOR] [INFO] Todos los procesos han finalizado la ejecución de demostración de la Parte 2.
```

---

## 📸 Evidencia 3: Ejecución de la Suite Completa de Pruebas Unitarias

### Comando Ejecutado:
```powershell
go test -v ./...
```

### Resultado de Pruebas:
```text
?   	github.com/nsilupu-30/go-process-supervisor/cmd/supervisor	[no test files]
=== RUN   TestLoadConfigTableDriven
=== RUN   TestLoadConfigTableDriven/Configuración_válida
=== RUN   TestLoadConfigTableDriven/JSON_malformado
=== RUN   TestLoadConfigTableDriven/Archivo_de_configuración_inexistente
=== RUN   TestLoadConfigTableDriven/Lista_de_procesos_vacía
=== RUN   TestLoadConfigTableDriven/Nombres_de_procesos_duplicados
=== RUN   TestLoadConfigTableDriven/Comando_obligatorio_vacío
=== RUN   TestLoadConfigTableDriven/Política_de_reinicio_no_válida
=== RUN   TestLoadConfigTableDriven/Factor_de_backoff_menor_a_1.0
=== RUN   TestLoadConfigTableDriven/Max_seconds_de_backoff_menor_a_initial_seconds
=== RUN   TestLoadConfigTableDriven/Grace_period_no_positivo
=== RUN   TestLoadConfigTableDriven/Dirección_API_sin_puerto
--- PASS: TestLoadConfigTableDriven (0.16s)
PASS
ok  	github.com/nsilupu-30/go-process-supervisor/internal/config	0.16s

=== RUN   TestProcessRunner_SuccessfulExecution
--- PASS: TestProcessRunner_SuccessfulExecution (0.05s)
=== RUN   TestProcessRunner_NonZeroExitCode
--- PASS: TestProcessRunner_NonZeroExitCode (0.03s)
=== RUN   TestProcessRunner_NonExistentCommand
--- PASS: TestProcessRunner_NonExistentCommand (0.00s)
=== RUN   TestProcessRunner_EnvironmentVariables
--- PASS: TestProcessRunner_EnvironmentVariables (0.04s)
=== RUN   TestProcessRunner_ContextCancellation
--- PASS: TestProcessRunner_ContextCancellation (0.20s)
PASS
ok  	github.com/nsilupu-30/go-process-supervisor/internal/process	2.35s
```
