package supervisor

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
	"github.com/nsilupu-30/go-process-supervisor/internal/logging"
)

func escribirScript(t *testing.T, nombre, contenido string) string {
	t.Helper()
	dir := t.TempDir()
	ruta := filepath.Join(dir, nombre)
	if err := os.WriteFile(ruta, []byte(contenido), 0755); err != nil {
		t.Fatalf("no se pudo crear script %s: %v", nombre, err)
	}
	return ruta
}

func configEcoLinux(t *testing.T) config.ProcessConfig {
	ruta := escribirScript(t, "eco.sh", "#!/bin/sh\necho ok\n")
	return config.ProcessConfig{
		Name:          "eco",
		Command:       ruta,
		RestartPolicy: config.RestartNever,
		Backoff:       config.BackoffConfig{InitialSeconds: 1, Factor: 2, MaxSeconds: 10},
	}
}

func configEsperaLarga(t *testing.T) config.ProcessConfig {
	ruta := escribirScript(t, "largo.sh", "#!/bin/sh\nsleep 5\n")
	return config.ProcessConfig{
		Name:          "largo",
		Command:       ruta,
		RestartPolicy: config.RestartNever,
		Backoff:       config.BackoffConfig{InitialSeconds: 1, Factor: 2, MaxSeconds: 10},
	}
}

func TestAdministradorProceso_ComandosNoSePierden(t *testing.T) {
	cfg := configEsperaLarga(t)
	logger := logging.NewProcessLoggerWithWriter(&bytes.Buffer{})
	mgr := NuevoAdministradorProceso(cfg, logger)

	mgr.EnviarComando(comandoIniciar, context.Background())
	mgr.EnviarComando(comandoDetener, context.Background())

	cancelable, cancelar := context.WithCancel(context.Background())
	go mgr.IniciarCicloVida(cancelable)
	time.Sleep(500 * time.Millisecond)
	cancelar()

	snap := mgr.ObtenerSnapshot()
	if snap.Estado != EstadoDetenido && snap.Estado != EstadoDeteniendo {
		t.Fatalf("esperado estado detenido o deteniendo, obtenido %q", snap.Estado)
	}
}

func TestAdministradorProceso_ComandoInexistenteNoRompe(t *testing.T) {
	cfg := config.ProcessConfig{
		Name:          "inexistente",
		Command:       "/ruta/inexistente/seguro",
		RestartPolicy: config.RestartNever,
		Backoff:       config.BackoffConfig{InitialSeconds: 1, Factor: 2, MaxSeconds: 10},
	}
	logger := logging.NewProcessLoggerWithWriter(&bytes.Buffer{})
	mgr := NuevoAdministradorProceso(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	hecho := make(chan struct{})
	go func() {
		mgr.IniciarCicloVida(ctx)
		close(hecho)
	}()

	select {
	case <-hecho:
		snap := mgr.ObtenerSnapshot()
		if snap.Estado != EstadoFallido && snap.Estado != EstadoDetenido && snap.Estado != EstadoDeteniendo {
			t.Fatalf("esperado estado fallido, detenido o deteniendo, obtenido %q", snap.Estado)
		}
	case <-time.After(11 * time.Second):
		t.Fatal("timeout esperando fin del ciclo de vida")
	}
}

func TestAdministradorProceso_CancelacionDetieneCiclo(t *testing.T) {
	cfg := configEsperaLarga(t)
	logger := logging.NewProcessLoggerWithWriter(&bytes.Buffer{})
	mgr := NuevoAdministradorProceso(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go mgr.IniciarCicloVida(ctx)
	time.Sleep(300 * time.Millisecond)
	cancel()

	select {
	case <-time.After(3 * time.Second):
		t.Fatal("timeout: cancelación no detuvo el ciclo de vida")
	default:
	}
}
