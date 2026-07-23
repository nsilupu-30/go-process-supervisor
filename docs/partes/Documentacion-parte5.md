# Documentación — Parte 5: Señales, apagado ordenado y recarga dinámica

## 1. Objetivo de la Parte 5
Controlar señales del sistema operativo para garantizar un cierre limpio del supervisor y sus procesos hijos, además de permitir la recarga dinámica de la configuración sin reiniciar todo el sistema.

## 2. Archivos Creados y Modificados
- `internal/senales/manejador.go`: `ManejadorSenales` que escucha `SIGINT`, `SIGTERM` y `SIGHUP`, cancela contexto global, propaga señales a procesos hijos y recarga configuración.
- `internal/senales/manejador_test.go`: pruebas unitarias de cancelación, grace period y recarga inválida.
- `cmd/supervisor/main.go`: integración del `ManejadorSenales` en el comando `run`.
- `docs/partes/Documentacion-parte5.md`: este documento.

## 3. Tipos Exportados

### ManejadorSenales
Constructor: `NuevoManejadorSenales(gracePeriod time.Duration) *ManejadorSenales`

Métodos públicos:
- `Iniciar(ctx context.Context) context.Context`
- `IniciarConRecarga(ctx context.Context, rutaConfig string, sup *supervisor.Supervisor, logger logging.Logger) context.Context`
- `GracePeriod() time.Duration`

Campos internos:
- `gracePeriod`: duración del periodo de gracia antes de `SIGKILL`.
- `canal`: canal de señales del sistema (`os.Signal`).
- `cancelar`: función para cancelar el contexto global.
- `rutaConfig`: ruta al archivo JSON de configuración para recarga.
- `supervisor`: referencia al `Supervisor` para propagar detención.
- `logger`: logger de procesos.

## 4. Flujo de Apagado (SIGINT / SIGTERM)

```
Señal recibida
      ↓
Cancelar contexto global
      ↓
Enviar SIGTERM a cada proceso hijo
      ↓
Esperar grace_period_seconds
      ↓
Enviar SIGKILL a procesos que no terminaron
      ↓
Esperar salida de todos los procesos (cmd.Wait())
      ↓
Cerrar API y logs
      ↓
Salir
```

Reglas:
- El `ManejadorSenales` escucha en una goroutine dedicada.
- Al recibir `SIGINT` o `SIGTERM`: invoca `cancelar()` del contexto global.
- El `Supervisor` propaga `Detener(ctx)` a cada `AdministradorProceso`.
- Cada manager marca estado `deteniendo` y espera `GracePeriod()` antes de forzar terminación.
- No se permite reiniciar procesos durante el apagado.

## 5. Flujo de Recarga Dinámica (SIGHUP)

```
SIGHUP recibido
      ↓
Leer archivo de configuración
      ↓
Validar con ValidateConfig
      ↓
Comparar con configuración anterior
      ↓
Si inválida → registrar error y mantener anterior
      ↓
Si válida → actualizar lista de procesos
            → Agregar procesos nuevos
            → Reiniciar procesos modificados
            → Detener procesos eliminados
```

Reglas:
- Una configuración inválida no tumba el supervisor.
- La configuración anterior permanece activa hasta que la nueva sea válida.
- El método `recargarConfiguracion()` encapsula la lógica de lectura, validación y logging.
- Se compara por nombre de proceso para detectar agregados, eliminados y modificados.

## 6. Integración en main.go

El comando `run` ahora:
1. Crea `ManejadorSenales` con `GracePeriodSeconds` de la config.
2. Inicia `IniciarConRecarga(ctx, configPath, supervisor, logger)`.
3. Lanza `supervisor.Iniciar(ctx)` en goroutine.
4. Al recibir señal, propaga cancelación y detiene el supervisor ordenadamente.

## 7. Decisiones técnicas
1. Canal de señales con buffer 2: permite acumular `SIGINT`/`SIGTERM` y `SIGHUP` sin perder ninguna.
2. `context.WithCancel` como mecanismo principal de cancelación: evita señales duplicadas y race conditions.
3. Recarga en goroutine separada del ciclo de vida: no bloquea procesos en ejecución.
4. Grace period como `time.Duration` en el constructor: flexible para pruebas y producción.
5. `recargarConfiguracion()` loguea errores sin propagar: garantiza continuidad ante config inválida.

## 8. Pruebas y validación
Suite en `internal/senales/manejador_test.go`:
- `TestManejador_CancelacionContexto_DetieneSupervisor`
- `TestManejador_GracePeriodConfigurado`
- `TestManejador_RecargaConfigInvalida_NoRompeSupervisor`

Ejecución:
```bash
go test ./internal/senales/...
go test ./...
gofmt -w .
```

## 9. Limitaciones
- El diff real de procesos para agregar/eliminar/modificar durante recarga está planificado pero no completamente implementado en `recargarConfiguracion()`.
- El envío efectivo de `SIGTERM`/`SIGKILL` a PIDs hijos se coordina desde `AdministradorProceso` (Parte 4), no desde el `ManejadorSenales` directamente.
- En Windows, `SIGHUP` no está disponible nativamente; se simula en pruebas y el manejo queda condicionado.
