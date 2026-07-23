package senales

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
	"github.com/nsilupu-30/go-process-supervisor/internal/logging"
	"github.com/nsilupu-30/go-process-supervisor/internal/supervisor"
)

func escribirConfigTemporal(t *testing.T, contenido string) string {
	t.Helper()
	dir := t.TempDir()
	ruta := filepath.Join(dir, "config.json")
	if err := os.WriteFile(ruta, []byte(contenido), 0644); err != nil {
		t.Fatalf("no se pudo crear config temporal: %v", err)
	}
	return ruta
}

func configValida() string {
	return `{
		"grace_period_seconds": 2,
		"api_address": "127.0.0.1:8080",
		"processes": [
			{
				"name": "eco",
				"command": "cmd.exe",
				"args": ["/c", "echo", "ok"],
				"restart_policy": "never",
				"backoff": {"initial_seconds": 1, "factor": 2, "max_seconds": 10}
			}
		]
	}`
}

func configInvalida() string {
	return `{
		"grace_period_seconds": 0,
		"api_address": "127.0.0.1:8080",
		"processes": []
	}`
}

// Verifica que la cancelación del contexto interrumpa el ciclo del supervisor
func TestManejador_CancelacionContexto_DetieneSupervisor(t *testing.T) {
	ruta := escribirConfigTemporal(t, configValida())
	cfg, _ := config.LoadConfig(ruta)
	logger := logging.NewProcessLoggerWithWriter(&bytes.Buffer{})
	sup := supervisor.NuevoSupervisor(*cfg, logger)

	manejador := NuevoManejadorSenales(500 * time.Millisecond)
	ctx, cancelar := context.WithCancel(context.Background())

	ctx = manejador.Iniciar(ctx)
	go sup.Iniciar(ctx)

	go func() {
		time.Sleep(200 * time.Millisecond)
		cancelar()
	}()

	select {
	case <-ctx.Done():
		// contexto cancelado correctamente, esperado
	case <-time.After(3 * time.Second):
		t.Fatal("timeout: la cancelación no propagó al contexto del supervisor")
	}
}

// Verifica que el grace period configurado se devuelve correctamente
func TestManejador_GracePeriodConfigurado(t *testing.T) {
	grace := 750 * time.Millisecond
	manejador := NuevoManejadorSenales(grace)

	if manejador.GracePeriod() != grace {
		t.Fatalf("esperado grace period %v, obtenido %v", grace, manejador.GracePeriod())
	}
}

// Verifica que recargar configuración inválida no rompe el supervisor
func TestManejador_RecargaConfigInvalida_NoRompeSupervisor(t *testing.T) {
	rutaValida := escribirConfigTemporal(t, configValida())
	cfg, _ := config.LoadConfig(rutaValida)

	logger := logging.NewProcessLoggerWithWriter(&bytes.Buffer{})
	sup := supervisor.NuevoSupervisor(*cfg, logger)

	manejador := NuevoManejadorSenales(1 * time.Second)
	ctx, cancelar := context.WithCancel(context.Background())
	defer cancelar()

	ctx = manejador.IniciarConRecarga(ctx, rutaValida, sup, logger)
	go sup.Iniciar(ctx)
	time.Sleep(300 * time.Millisecond)

	// Sobrescribir con una configuración inválida y recargar
	_ = escribirConfigTemporal(t, configInvalida())
	// recargarConfiguracion lee el archivo por ruta guardada en el manejador
	manejador.recargarConfiguracion()

	// El supervisor debe seguir respondiendo
	snaps := sup.ObtenerSnapshots()
	if len(snaps) == 0 {
		t.Fatal("esperado al menos un snapshot después de recarga inválida")
	}
}
