package senales

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
	"github.com/nsilupu-30/go-process-supervisor/internal/logging"
	"github.com/nsilupu-30/go-process-supervisor/internal/supervisor"
)

type ManejadorSenales struct {
	gracePeriod time.Duration
	canal       chan os.Signal
	cancelar    context.CancelFunc
	rutaConfig  string
	supervisor  *supervisor.Supervisor
	logger      logging.Logger
}

func NuevoManejadorSenales(gracePeriod time.Duration) *ManejadorSenales {
	return &ManejadorSenales{
		gracePeriod: gracePeriod,
		canal:       make(chan os.Signal, 1),
	}
}

func (m *ManejadorSenales) IniciarConRecarga(ctx context.Context, rutaConfig string, sup *supervisor.Supervisor, logger logging.Logger) context.Context {
	ctx, cancelar := context.WithCancel(ctx)
	m.cancelar = cancelar
	m.rutaConfig = rutaConfig
	m.supervisor = sup
	m.logger = logger

	canalSenales := make(chan os.Signal, 2)
	signal.Notify(canalSenales, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	m.canal = canalSenales

	go m.esperarSenal(ctx)
	return ctx
}

func (m *ManejadorSenales) Iniciar(ctx context.Context) context.Context {
	ctx, cancelar := context.WithCancel(ctx)
	m.cancelar = cancelar

	signal.Notify(m.canal, syscall.SIGINT, syscall.SIGTERM)

	go m.esperarSenal(ctx)
	return ctx
}

func (m *ManejadorSenales) esperarSenal(ctx context.Context) {
	for {
		select {
		case <-m.canal:
			m.manejarSenal()
		case <-ctx.Done():
			signal.Stop(m.canal)
			return
		}
	}
}

func (m *ManejadorSenales) manejarSenal() {
	s := <-m.canal
	switch s {
	case syscall.SIGINT, syscall.SIGTERM:
		m.cancelar()
	}
}

func (m *ManejadorSenales) GracePeriod() time.Duration {
	return m.gracePeriod
}

func (m *ManejadorSenales) recargarConfiguracion() {
	if m.rutaConfig == "" || m.supervisor == nil {
		return
	}

	_, err := config.LoadConfig(m.rutaConfig)
	if err != nil {
		m.logger.LogError("SIGHUP", "configuracion invalida, se mantiene la anterior: "+err.Error())
		return
	}

	m.logger.LogInfo("SIGHUP", "configuracion recargada correctamente")
}
