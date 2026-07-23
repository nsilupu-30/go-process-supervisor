package supervisor

import (
	"context"
	"errors"
	"sync"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
	"github.com/nsilupu-30/go-process-supervisor/internal/logging"
)

type Supervisor struct {
	configuracion   config.AppConfig
	logger          logging.Logger
	administradores []*AdministradorProceso
}

var ErrProcessNotFound = errors.New("process not found")

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

func (s *Supervisor) Health() bool {
	return true
}

func (s *Supervisor) ProcessSnapshots() []SnapshotProceso {
	snapshots := make([]SnapshotProceso, len(s.administradores))
	for i, mgr := range s.administradores {
		snapshots[i] = mgr.ObtenerSnapshot()
	}
	return snapshots
}

func (s *Supervisor) ObtenerSnapshots() []SnapshotProceso {
	return s.ProcessSnapshots()
}

func (s *Supervisor) ProcessSnapshot(name string) (SnapshotProceso, bool) {
	mgr := s.findAdministrador(name)
	if mgr == nil {
		return SnapshotProceso{}, false
	}
	return mgr.ObtenerSnapshot(), true
}

func (s *Supervisor) StartProcess(name string, ctx context.Context) error {
	mgr := s.findAdministrador(name)
	if mgr == nil {
		return ErrProcessNotFound
	}
	mgr.EnviarComando(comandoIniciar, ctx)
	return nil
}

func (s *Supervisor) StopProcess(name string, ctx context.Context) error {
	mgr := s.findAdministrador(name)
	if mgr == nil {
		return ErrProcessNotFound
	}
	mgr.EnviarComando(comandoDetener, ctx)
	return nil
}

func (s *Supervisor) RestartProcess(name string, ctx context.Context) error {
	mgr := s.findAdministrador(name)
	if mgr == nil {
		return ErrProcessNotFound
	}
	mgr.EnviarComando(comandoReiniciar, ctx)
	return nil
}

func (s *Supervisor) Reload() error {
	return nil
}

func (s *Supervisor) findAdministrador(name string) *AdministradorProceso {
	for _, mgr := range s.administradores {
		if mgr.config.Name == name {
			return mgr
		}
	}
	return nil
}
