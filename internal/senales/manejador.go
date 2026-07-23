package senales

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ManejadorSenales struct {
	gracePeriod time.Duration
	canal       chan os.Signal
	cancelar    context.CancelFunc
}

func NuevoManejadorSenales(gracePeriod time.Duration) *ManejadorSenales {
	return &ManejadorSenales{
		gracePeriod: gracePeriod,
		canal:       make(chan os.Signal, 1),
	}
}

func (m *ManejadorSenales) Iniciar(ctx context.Context) context.Context {
	ctx, cancelar := context.WithCancel(ctx)
	m.cancelar = cancelar

	signal.Notify(m.canal, syscall.SIGINT, syscall.SIGTERM)

	go m.esperarSenal(ctx)

	return ctx
}

func (m *ManejadorSenales) esperarSenal(ctx context.Context) {
	select {
	case <-m.canal:
		m.cancelar()
	case <-ctx.Done():
		signal.Stop(m.canal)
	}
}

func (m *ManejadorSenales) GracePeriod() time.Duration {
	return m.gracePeriod
}
