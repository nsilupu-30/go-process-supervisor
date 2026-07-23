package supervisor

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
	"github.com/nsilupu-30/go-process-supervisor/internal/logging"
)

func configEco() config.ProcessConfig {
	return config.ProcessConfig{
		Name:          "eco",
		Command:       "cmd.exe",
		Args:          []string{"/c", "echo", "ok"},
		RestartPolicy: config.RestartNever,
		Backoff:       config.BackoffConfig{InitialSeconds: 1, Factor: 2, MaxSeconds: 10},
	}
}

func configFallo() config.ProcessConfig {
	return config.ProcessConfig{
		Name:          "fallo",
		Command:       "cmd.exe",
		Args:          []string{"/c", "exit", "1"},
		RestartPolicy: config.RestartOnFailure,
		MaxRetries:    1,
		Backoff:       config.BackoffConfig{InitialSeconds: 1, Factor: 2, MaxSeconds: 10},
	}
}

func TestAdministradorProceso_SnapshotInicial(t *testing.T) {
	cfg := configEco()
	logger := logging.NewProcessLoggerWithWriter(&bytes.Buffer{})
	mgr := NuevoAdministradorProceso(cfg, logger)

	snap := mgr.ObtenerSnapshot()
	if snap.Nombre != cfg.Name {
		t.Fatalf("esperado nombre %q, obtenido %q", cfg.Name, snap.Nombre)
	}
	if snap.Estado != EstadoCreado {
		t.Fatalf("esperado estado %q, obtenido %q", EstadoCreado, snap.Estado)
	}
}

func TestAdministradorProceso_ProcesoExitosoNoReinicia(t *testing.T) {
	cfg := configEco()
	var out bytes.Buffer
	logger := logging.NewProcessLoggerWithWriter(&out)
	mgr := NuevoAdministradorProceso(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go mgr.IniciarCicloVida(ctx)
	time.Sleep(600 * time.Millisecond)

	snap := mgr.ObtenerSnapshot()
	if snap.Reinicios != 0 {
		t.Fatalf("esperado 0 reintentos, obtenido %d", snap.Reinicios)
	}
	if strings.Contains(out.String(), "reinicio programado") {
		t.Logf("logs: %s", out.String())
		t.Fatalf("no debe programarse reinicio para proceso exitoso con RestartNever")
	}
}

func TestAdministradorProceso_CancelacionInterrumpeEspera(t *testing.T) {
	cfg := configFallo()
	logger := logging.NewProcessLoggerWithWriter(&bytes.Buffer{})
	mgr := NuevoAdministradorProceso(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go mgr.IniciarCicloVida(ctx)

	select {
	case <-time.After(3 * time.Second):
		t.Fatal("timeout: cancelación no detuvo el ciclo de vida")
	case <-ctx.Done():
	}

	snap := mgr.ObtenerSnapshot()
	if snap.Estado == EstadoEspera && snap.Siguiente != nil {
		t.Fatalf("esperado estado posterior a espera, obtenido %q", snap.Estado)
	}
}

func TestSupervisor_VariosProcesosSimultaneos(t *testing.T) {
	cfg := config.AppConfig{
		GracePeriodSeconds: 5,
		APIAddress:         "127.0.0.1:8080",
		Processes: []config.ProcessConfig{
			{Name: "p1", Command: "cmd.exe", Args: []string{"/c", "echo", "p1"}, RestartPolicy: config.RestartNever, Backoff: config.BackoffConfig{InitialSeconds: 1, Factor: 2, MaxSeconds: 10}},
			{Name: "p2", Command: "cmd.exe", Args: []string{"/c", "echo", "p2"}, RestartPolicy: config.RestartNever, Backoff: config.BackoffConfig{InitialSeconds: 1, Factor: 2, MaxSeconds: 10}},
		},
	}

	logger := logging.NewProcessLoggerWithWriter(&bytes.Buffer{})
	sup := NuevoSupervisor(cfg, logger)

	go sup.Iniciar(context.Background())
	time.Sleep(400 * time.Millisecond)

	snaps := sup.ObtenerSnapshots()
	if len(snaps) != 2 {
		t.Fatalf("esperados 2 snapshots, obtenidos %d", len(snaps))
	}
	nombres := map[string]bool{"p1": true, "p2": true}
	for _, s := range snaps {
		if !nombres[s.Nombre] {
			t.Fatalf("nombre inesperado %q", s.Nombre)
		}
	}
}

func TestSupervisor_ApagadoOrdenado(t *testing.T) {
	cfg := config.AppConfig{
		GracePeriodSeconds: 1,
		APIAddress:         "127.0.0.1:8080",
		Processes: []config.ProcessConfig{
			{Name: "lento", Command: "cmd.exe", Args: []string{"/c", "timeout", "10"}, RestartPolicy: config.RestartNever, Backoff: config.BackoffConfig{InitialSeconds: 1, Factor: 2, MaxSeconds: 10}},
		},
	}

	logger := logging.NewProcessLoggerWithWriter(&bytes.Buffer{})
	sup := NuevoSupervisor(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	go sup.Iniciar(ctx)
	time.Sleep(300 * time.Millisecond)
	sup.Detener(ctx)

	snaps := sup.ObtenerSnapshots()
	if len(snaps) != 1 {
		t.Fatalf("esperado 1 snapshot, obtenidos %d", len(snaps))
	}
}

func TestAdministradorProceso_ComandosNoDuplican(t *testing.T) {
	cfg := configEco()
	logger := logging.NewProcessLoggerWithWriter(&bytes.Buffer{})
	mgr := NuevoAdministradorProceso(cfg, logger)

	ctx := context.Background()
	mgr.EnviarComando(comandoIniciar, ctx)
	mgr.EnviarComando(comandoIniciar, ctx)
	mgr.EnviarComando(comandoDetener, ctx)

	cancelable, cancelar := context.WithCancel(ctx)
	go mgr.IniciarCicloVida(cancelable)
	time.Sleep(2 * time.Second)
	cancelar()

	snap := mgr.ObtenerSnapshot()
	if snap.Reinicios != 0 {
		t.Fatalf("sin reintentos, obtenido %d", snap.Reinicios)
	}
}
