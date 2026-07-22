package config

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// ValidateConfig comprueba que todas las reglas de negocio de AppConfig se cumplan.
// Devuelve un error descriptivo si encuentra alguna falla de validación.
func ValidateConfig(cfg *AppConfig) error {
	if cfg == nil {
		return errors.New("la configuración no puede ser nula")
	}

	// El periodo de gracia debe ser un número positivo de segundos para permitir el cierre ordenado.
	if cfg.GracePeriodSeconds <= 0 {
		return fmt.Errorf("grace_period_seconds debe ser positivo, se obtuvo: %d", cfg.GracePeriodSeconds)
	}

	// La dirección de la API HTTP debe especificarse y tener un formato host:puerto válido.
	trimmedAPI := strings.TrimSpace(cfg.APIAddress)
	if trimmedAPI == "" {
		return errors.New("api_address es obligatorio")
	}
	if _, _, err := net.SplitHostPort(trimmedAPI); err != nil {
		return fmt.Errorf("api_address inválido %q: debe tener el formato host:puerto (ej. 127.0.0.1:8080): %w", cfg.APIAddress, err)
	}

	// Debe configurarse al menos un proceso para que el supervisor tenga tareas que gestionar.
	if len(cfg.Processes) == 0 {
		return errors.New("la lista de procesos ('processes') no puede estar vacía")
	}

	seenNames := make(map[string]bool)

	for i, proc := range cfg.Processes {
		procName := strings.TrimSpace(proc.Name)
		if procName == "" {
			return fmt.Errorf("proceso en índice %d: el nombre ('name') es obligatorio", i)
		}

		// Se exige que los nombres sean únicos para identificar sin ambigüedad cada proceso en logs y API.
		if seenNames[procName] {
			return fmt.Errorf("nombre de proceso duplicado %q en el índice %d", procName, i)
		}
		seenNames[procName] = true

		if strings.TrimSpace(proc.Command) == "" {
			return fmt.Errorf("proceso %q: el comando ('command') es obligatorio", procName)
		}

		switch proc.RestartPolicy {
		case RestartAlways, RestartOnFailure, RestartNever:
			// Política de reinicio permitida.
		default:
			return fmt.Errorf("proceso %q: política de reinicio inválida %q (permitidas: 'always', 'on-failure', 'never')", procName, proc.RestartPolicy)
		}

		if proc.MaxRetries < 0 {
			return fmt.Errorf("proceso %q: max_retries no puede ser negativo, se obtuvo: %d", procName, proc.MaxRetries)
		}

		if err := validateBackoff(procName, proc.Backoff); err != nil {
			return err
		}
	}

	return nil
}

// validateBackoff verifica las propiedades de la estrategia de backoff de un proceso.
func validateBackoff(procName string, b BackoffConfig) error {
	if b.InitialSeconds <= 0 {
		return fmt.Errorf("proceso %q: backoff.initial_seconds debe ser positivo, se obtuvo: %d", procName, b.InitialSeconds)
	}

	if b.Factor < 1.0 {
		return fmt.Errorf("proceso %q: backoff.factor debe ser mayor o igual a 1.0, se obtuvo: %f", procName, b.Factor)
	}

	if b.MaxSeconds <= 0 {
		return fmt.Errorf("proceso %q: backoff.max_seconds debe ser positivo, se obtuvo: %d", procName, b.MaxSeconds)
	}

	if b.MaxSeconds < b.InitialSeconds {
		return fmt.Errorf("proceso %q: backoff.max_seconds (%d) debe ser mayor o igual a initial_seconds (%d)", procName, b.MaxSeconds, b.InitialSeconds)
	}

	return nil
}
