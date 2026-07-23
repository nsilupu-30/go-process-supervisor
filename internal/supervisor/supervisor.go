package supervisor

import (
	"context"
	"sync"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
	"github.com/nsilupu-30/go-process-supervisor/internal/logging"
)

type Supervisor struct {
	configuracion config.AppConfig
	logger        logging.Logger
	administradores []*AdministradorProceso
}

func NuevoSupervisor(cfg config.AppConfig, logger logging.Logger) *Supervisor {
	return &Supervisor{configuracion: cfg, logger: logger}
}

func (s *Supervisor) Iniciar(ctx context.Context) {
	s.administradores = make([]*AdministradorProceso, len(s.configuracion.Processes))
	var espera sync.WaitGroup
	for i, proc := range s.configuracion.Processes {
		s.administradores[i] = NuevoAdministradorProceso(proc, s.logger)
		espera.Add(1)
		go func(mgr *AdministradorProceso) {
			defer espera.Done()
			mgr.IniciarCicloVida(ctx)
		}(s.administradores[i])
	}
	espera.Wait()
}

func (s *Supervisor) Detener(ctx context.Context) {
	for _, mgr := range s.administradores {
		mgr.EnviarComando(comandoDetener, ctx)
	}
}

func (s *Supervisor) ObtenerSnapshots() []SnapshotProceso {
	snapshots := make([]SnapshotProceso, len(s.administradores))
	for i, mgr := range s.administradores {
		snapshots[i] = mgr.ObtenerSnapshot()
	}
	return snapshots
}
