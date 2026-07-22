package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
	"github.com/nsilupu-30/go-process-supervisor/internal/logging"
)

// ProcessRunner se encarga del lanzamiento, monitoreo y recolección de los procesos hijos.
type ProcessRunner struct{}

// NewRunner crea un nuevo ejecutor de procesos.
func NewRunner() *ProcessRunner {
	return &ProcessRunner{}
}

// Run ejecuta el proceso especificado en cfg, captura su salida de forma asíncrona hacia logger
// y espera a su terminación liberando los recursos del sistema operativo (evitando procesos zombis).
func (r *ProcessRunner) Run(ctx context.Context, cfg config.ProcessConfig, logger logging.Logger) (*ExecutionResult, error) {
	if logger == nil {
		logger = logging.NewProcessLogger()
	}

	cmd := exec.CommandContext(ctx, cfg.Command, cfg.Args...)

	if cfg.WorkingDir != "" {
		cmd.Dir = cfg.WorkingDir
	}

	cmd.Env = mergeEnvironment(os.Environ(), cfg.Environment)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error al crear stdout pipe para %s: %w", cfg.Name, err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("error al crear stderr pipe para %s: %w", cfg.Name, err)
	}

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		logging.StreamPipe(stdoutPipe, cfg.Name, logger.LogStdout)
	}()

	go func() {
		defer wg.Done()
		logging.StreamPipe(stderrPipe, cfg.Name, logger.LogStderr)
	}()

	startedAt := time.Now()
	if err := cmd.Start(); err != nil {
		logger.LogError(cfg.Name, fmt.Sprintf("falló al iniciar el comando: %v", err))
		return &ExecutionResult{
			StartedAt: startedAt,
			ExitedAt:  time.Now(),
			Duration:  0,
			Error:     fmt.Errorf("error al iniciar proceso %s: %w", cfg.Name, err),
			ExitCode:  -1,
		}, nil
	}

	pid := cmd.Process.Pid
	logger.LogInfo(cfg.Name, fmt.Sprintf("proceso iniciado correctamente con PID %d", pid))

	// Invocación obligatoria de Wait() para recolectar el proceso y evitar zombis
	waitErr := cmd.Wait()
	exitedAt := time.Now()
	wg.Wait() // Esperar a que se terminen de procesar todos los logs

	result := &ExecutionResult{
		PID:       pid,
		StartedAt: startedAt,
		ExitedAt:  exitedAt,
		Duration:  exitedAt.Sub(startedAt),
	}

	if waitErr == nil {
		result.ExitCode = 0
		logger.LogInfo(cfg.Name, "proceso finalizó exitosamente (exit code 0)")
		return result, nil
	}

	if exitErr, ok := waitErr.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
		result.Error = exitErr

		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if status.Signaled() {
				result.TerminatedBySignal = true
				result.Signal = status.Signal().String()
				logger.LogInfo(cfg.Name, fmt.Sprintf("proceso finalizado por señal: %s (exit code %d)", result.Signal, result.ExitCode))
				return result, nil
			}
		}

		logger.LogInfo(cfg.Name, fmt.Sprintf("proceso finalizó con código de salida %d", result.ExitCode))
		return result, nil
	}

	result.Error = waitErr
	result.ExitCode = -1
	logger.LogError(cfg.Name, fmt.Sprintf("error inesperado durante la ejecución: %v", waitErr))
	return result, nil
}

// mergeEnvironment combina el entorno del host con las variables personalizadas del proceso.
func mergeEnvironment(hostEnv []string, customEnv map[string]string) []string {
	if len(customEnv) == 0 {
		return hostEnv
	}

	envMap := make(map[string]string)
	for _, e := range hostEnv {
		for i := 0; i < len(e); i++ {
			if e[i] == '=' {
				key := e[:i]
				val := e[i+1:]
				envMap[key] = val
				break
			}
		}
	}

	for k, v := range customEnv {
		envMap[k] = v
	}

	merged := make([]string, 0, len(envMap))
	for k, v := range envMap {
		merged = append(merged, fmt.Sprintf("%s=%s", k, v))
	}

	return merged
}
