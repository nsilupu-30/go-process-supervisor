# Documentación — Parte 3: Monitoreo, Reinicios y Backoff Exponencial

## 1. Objetivo de la Parte 3
Implementar la lógica de decisión de reinicios automáticos y el cálculo de temporización por retardo exponencial (*exponential backoff*) dentro del paquete `internal/restart`. El módulo se encarga de:
1. Evaluar las políticas de reinicio declarativas (`always`, `on-failure`, `never`) junto al contador de intentos y el límite `max_retries`.
2. Calcular dinámicamente el tiempo de espera entre reintentos según la fórmula exponencial $espera = \min(initial \times factor^{(intento-1)}, max\_seconds)$.
3. Ofrecer un mecanismo de espera cancelable mediante `context.Context` para interrumpir los temporizadores durante el apagado del supervisor.
4. Prevenir tormentas de reinicios en bucle infinito ante fallos constantes.

---

## 2. Archivos Creados y Modificados
- `internal/restart/policy.go`: Función de decisión `ShouldRestart`.
- `internal/restart/backoff.go`: Cálculo exponencial `CalculateDelay` y temporizador cancelable `Wait`.
- `internal/restart/backoff_test.go`: Suite de pruebas unitarias basadas en tablas cubriendo políticas, tope de retardo y cancelación por contexto.
- `docs/Documentacion-parte3.md`: Especificación técnica y evidencias de la Parte 3.

---

## 3. Estructuras y Funciones Implementadas

### `ShouldRestart` (`internal/restart/policy.go`)
Firma de la función de evaluación:
```go
func ShouldRestart(policy config.RestartPolicy, exitCode int, currentRetries int, maxRetries int) bool
```

#### Reglas de Evaluación:
1. **Límite de Reintentos:** Si `maxRetries > 0` y `currentRetries >= maxRetries`, retorna `false` (pasa al estado `failed`).
2. **Política `never`:** Retorna siempre `false`.
3. **Política `always`:** Retorna siempre `true` (mientras no supere `maxRetries`).
4. **Política `on-failure`:** Retorna `true` únicamente si el código de salida es distinto de cero (`exitCode != 0`).

---

### `CalculateDelay` y `Wait` (`internal/restart/backoff.go`)

#### Fórmula de Retardo Exponencial:
$$\text{espera} = \min\left(\text{initial\_seconds} \times \text{factor}^{(\text{intento}-1)}, \text{max\_seconds}\right)$$

#### Ejemplo Práctico (para `initial=1s`, `factor=2.0`, `max=10s`):
- Intento 1: $1 \times 2^0 = 1\text{s}$
- Intento 2: $1 \times 2^1 = 2\text{s}$
- Intento 3: $1 \times 2^2 = 4\text{s}$
- Intento 4: $1 \times 2^3 = 8\text{s}$
- Intento 5: $1 \times 2^4 = 16\text{s} \rightarrow$ **Limitado al tope máximo de 10s**.

#### Cancelación por Contexto (`Wait`):
```go
func Wait(ctx context.Context, delay time.Duration) error
```
Utiliza un temporizador `time.NewTimer` escuchando en paralelo el canal `ctx.Done()`. Si el contexto global del supervisor se cancela, la espera se interrumpe inmediatamente y retorna `ctx.Err()`, evitando bloqueos de goroutines durante el apagado del sistema.

---

## 4. Decisiones Técnicas
1. **Separación de Responsabilidades:** `policy.go` se enfoca únicamente en la toma de decisiones pura (sin efectos secundarios ni llamadas a tiempo), mientras que `backoff.go` encapsula las matemáticas de tiempo y espera.
2. **Prevención de Tormentas de Reinicios:** La combinación del factor multiplicador y el límite de reintentos evita que un proceso que falla continuamente consuma el 100% de la CPU o llene los registros del sistema.

---

## 5. Pruebas Realizadas

La suite en `internal/restart/backoff_test.go` valida:
- `TestShouldRestart`: Pruebas de combinaciones de políticas `always`, `on-failure` y `never` con códigos 0 y 1, así como el agotamiento de reintentos.
- `TestCalculateDelay`: Verificación de valores de espera calculados y aplicación del tope máximo `MaxSeconds`.
- `TestWait_Completion`: Verificación del tiempo de espera normal.
- `TestWait_Cancellation`: Verificación de respuesta inmediata ante la cancelación del contexto.

### Comandos de Ejecución de Pruebas:
```bash
go test -v ./internal/restart/...
go test -v ./...
```
