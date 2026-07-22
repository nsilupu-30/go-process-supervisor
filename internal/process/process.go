package process

import (
	"time"
)

// ExecutionResult representa el resultado detallado de la ejecución de un proceso hijo.
type ExecutionResult struct {
	PID                int           `json:"pid"`
	ExitCode           int           `json:"exit_code"`
	StartedAt          time.Time     `json:"started_at"`
	ExitedAt           time.Time     `json:"exited_at"`
	Duration           time.Duration `json:"duration"`
	Error              error         `json:"-"`
	TerminatedBySignal bool          `json:"terminated_by_signal"`
	Signal             string        `json:"signal,omitempty"`
}

// Success indica si el proceso finalizó con código de salida 0 y sin errores de inicio.
func (r *ExecutionResult) Success() bool {
	return r.Error == nil && r.ExitCode == 0 && !r.TerminatedBySignal
}
