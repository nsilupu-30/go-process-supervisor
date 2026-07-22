package restart

import (
	"github.com/nsilupu-30/go-process-supervisor/internal/config"
)

// ShouldRestart evalúa si un proceso debe ser reiniciado basándose en la política configurada,
// el código de salida del proceso y el número de reintentos acumulados.
func ShouldRestart(policy config.RestartPolicy, exitCode int, currentRetries int, maxRetries int) bool {
	// Si se ha definido un máximo de reintentos (> 0) y ya se alcanzó o superó, no reiniciar.
	if maxRetries > 0 && currentRetries >= maxRetries {
		return false
	}

	switch policy {
	case config.RestartNever:
		return false

	case config.RestartAlways:
		return true

	case config.RestartOnFailure:
		return exitCode != 0

	default:
		return false
	}
}
