package config

// RestartPolicy define las políticas de reinicio de procesos soportadas por el supervisor.
type RestartPolicy string

const (
	// RestartAlways indica que el proceso debe ser reiniciado siempre que termine.
	RestartAlways RestartPolicy = "always"

	// RestartOnFailure indica que el proceso sólo se reinicia si termina con un código de salida distinto de cero o error.
	RestartOnFailure RestartPolicy = "on-failure"

	// RestartNever indica que el proceso no debe ser reiniciado.
	RestartNever RestartPolicy = "never"
)

// BackoffConfig define los parámetros para la estrategia de retardo exponencial entre reintentos.
type BackoffConfig struct {
	InitialSeconds int     `json:"initial_seconds"`
	Factor         float64 `json:"factor"`
	MaxSeconds     int     `json:"max_seconds"`
}

// ProcessConfig especifica la configuración declarativa de un proceso supervisado.
type ProcessConfig struct {
	Name          string            `json:"name"`
	Command       string            `json:"command"`
	Args          []string          `json:"args"`
	WorkingDir    string            `json:"working_dir"`
	Environment   map[string]string `json:"environment"`
	RestartPolicy RestartPolicy     `json:"restart_policy"`
	MaxRetries    int               `json:"max_retries"`
	Backoff       BackoffConfig     `json:"backoff"`
}

// AppConfig representa la configuración global de la aplicación cargada desde JSON.
type AppConfig struct {
	GracePeriodSeconds int             `json:"grace_period_seconds"`
	APIAddress         string          `json:"api_address"`
	Processes          []ProcessConfig `json:"processes"`
}
