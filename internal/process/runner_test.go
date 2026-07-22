package process

import (
	"bytes"
	"context"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
	"github.com/nsilupu-30/go-process-supervisor/internal/logging"
)

func TestProcessRunner_SuccessfulExecution(t *testing.T) {
	runner := NewRunner()
	var buf bytes.Buffer
	logger := logging.NewProcessLoggerWithWriter(&buf)

	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd.exe"
		args = []string{"/c", "echo", "hola supervisor"}
	} else {
		cmd = "echo"
		args = []string{"hola supervisor"}
	}

	cfg := config.ProcessConfig{
		Name:    "worker-hola",
		Command: cmd,
		Args:    args,
	}

	ctx := context.Background()
	res, err := runner.Run(ctx, cfg, logger)

	if err != nil {
		t.Fatalf("se esperaba ejecución sin error, se obtuvo: %v", err)
	}

	if !res.Success() {
		t.Errorf("se esperaba resultado exitoso, se obtuvo ExitCode=%d, Error=%v", res.ExitCode, res.Error)
	}

	if res.PID <= 0 {
		t.Errorf("se esperaba PID válido > 0, se obtuvo: %d", res.PID)
	}

	output := buf.String()
	if !strings.Contains(output, "hola supervisor") {
		t.Errorf("se esperaba encontrar 'hola supervisor' en la salida de log, salida real:\n%s", output)
	}

	if !strings.Contains(output, "[worker-hola]") {
		t.Errorf("se esperaba prefijo de proceso '[worker-hola]' en los logs, salida real:\n%s", output)
	}
}

func TestProcessRunner_NonZeroExitCode(t *testing.T) {
	runner := NewRunner()
	var buf bytes.Buffer
	logger := logging.NewProcessLoggerWithWriter(&buf)

	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd.exe"
		args = []string{"/c", "exit 42"}
	} else {
		cmd = "sh"
		args = []string{"-c", "exit 42"}
	}

	cfg := config.ProcessConfig{
		Name:    "worker-falla",
		Command: cmd,
		Args:    args,
	}

	ctx := context.Background()
	res, err := runner.Run(ctx, cfg, logger)

	if err != nil {
		t.Fatalf("Run() no debe retornar error de función en fallo de proceso hijo, error: %v", err)
	}

	if res.ExitCode != 42 {
		t.Errorf("se esperaba ExitCode=42, se obtuvo: %d", res.ExitCode)
	}

	if res.Success() {
		t.Errorf("Success() debe ser false cuando ExitCode != 0")
	}
}

func TestProcessRunner_NonExistentCommand(t *testing.T) {
	runner := NewRunner()
	var buf bytes.Buffer
	logger := logging.NewProcessLoggerWithWriter(&buf)

	cfg := config.ProcessConfig{
		Name:    "worker-invalido",
		Command: "comando_que_definitivamente_no_existe_98765",
	}

	ctx := context.Background()
	res, err := runner.Run(ctx, cfg, logger)

	if err != nil {
		t.Fatalf("se esperaba que el error se capturara en ExecutionResult, se obtuvo err=%v", err)
	}

	if res.Error == nil {
		t.Fatalf("se esperaba error de comando inexistente en ExecutionResult")
	}

	if res.ExitCode != -1 {
		t.Errorf("se esperaba ExitCode=-1 para fallos de arranque, se obtuvo: %d", res.ExitCode)
	}
}

func TestProcessRunner_EnvironmentVariables(t *testing.T) {
	runner := NewRunner()
	var buf bytes.Buffer
	logger := logging.NewProcessLoggerWithWriter(&buf)

	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "cmd.exe"
		args = []string{"/c", "echo %MI_VARIABLE_TEST%"}
	} else {
		cmd = "sh"
		args = []string{"-c", "echo $MI_VARIABLE_TEST"}
	}

	cfg := config.ProcessConfig{
		Name:    "worker-env",
		Command: cmd,
		Args:    args,
		Environment: map[string]string{
			"MI_VARIABLE_TEST": "valor_de_prueba_123",
		},
	}

	ctx := context.Background()
	res, err := runner.Run(ctx, cfg, logger)

	if err != nil {
		t.Fatalf("error en Run(): %v", err)
	}

	if res.ExitCode != 0 {
		t.Errorf("se esperaba ExitCode=0, se obtuvo: %d", res.ExitCode)
	}

	output := buf.String()
	if !strings.Contains(output, "valor_de_prueba_123") {
		t.Errorf("la variable de entorno no se propagó correctamente al proceso hijo, salida:\n%s", output)
	}
}

func TestProcessRunner_ContextCancellation(t *testing.T) {
	runner := NewRunner()
	var buf bytes.Buffer
	logger := logging.NewProcessLoggerWithWriter(&buf)

	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		cmd = "ping"
		args = []string{"-n", "10", "127.0.0.1"}
	} else {
		cmd = "sleep"
		args = []string{"10"}
	}

	cfg := config.ProcessConfig{
		Name:    "worker-lento",
		Command: cmd,
		Args:    args,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	res, err := runner.Run(ctx, cfg, logger)

	if err != nil {
		t.Fatalf("error en Run(): %v", err)
	}

	if res.Success() {
		t.Errorf("un proceso cancelado por contexto no debe ser exitoso")
	}
}
