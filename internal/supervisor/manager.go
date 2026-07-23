package supervisor

import (
	"context"
	"sync"
	"time"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
	"github.com/nsilupu-30/go-process-supervisor/internal/logging"
	"github.com/nsilupu-30/go-process-supervisor/internal/process"
	"github.com/nsilupu-30/go-process-supervisor/internal/restart"
)

type comandoManager int

const (
	comandoIniciar comandoManager = iota
	comandoDetener
	comandoReiniciar
)

type comando struct {
	tipo comandoManager
	ctx  context.Context
}
type AdministradorProceso struct {
	config   config.ProcessConfig
	logger   logging.Logger
	corredor *process.ProcessRunner
	almacen  *almacenSnapshots
	maquina  *estadoMaquina
	canal    chan comando
	detenido bool
	mu       sync.Mutex
}

func NuevoAdministradorProceso(cfg config.ProcessConfig, logger logging.Logger) *AdministradorProceso {
	return &AdministradorProceso{
		config:   cfg,
		logger:   logger,
		corredor: process.NewRunner(),
		almacen:  nuevoAlmacenSnapshots(cfg),
		maquina:  nuevoEstadoMaquina(),
		canal:    make(chan comando, 1),
	}
}
func (a *AdministradorProceso) EnviarComando(tipo comandoManager, ctx context.Context) {
	a.mu.Lock()
	if a.detenido {
		a.mu.Unlock()
		return
	}
	a.mu.Unlock()
	select {
	case a.canal <- comando{tipo: tipo, ctx: ctx}:
	default:
	}
}
func (a *AdministradorProceso) ObtenerSnapshot() SnapshotProceso {
	return a.almacen.obtenerSnapshot()
}
func (a *AdministradorProceso) IniciarCicloVida(ctx context.Context) {
	defer func() {
		a.mu.Lock()
		a.detenido = true
		a.mu.Unlock()
		close(a.canal)
	}()
	a.EnviarComando(comandoIniciar, ctx)
	for {
		select {
		case cmd, ok := <-a.canal:
			if !ok {
				return
			}
			switch cmd.tipo {
			case comandoDetener:
				a.almacen.marcarDeteniendo()
				a.maquina.transicionar(EventoProcesoDeteniendo)
			case comandoReiniciar:
				a.almacen.marcarDeteniendo()
				a.maquina.transicionar(EventoProcesoDeteniendo)
				a.maquina = nuevoEstadoMaquina()
				a.ejecutarProceso(cmd.ctx)
			case comandoIniciar:
				if a.maquina.actual() == EstadoCreado {
					a.ejecutarProceso(cmd.ctx)
				}
			}
		case <-ctx.Done():
			a.almacen.marcarDeteniendo()
			a.maquina.apagar(ctx)
			return
		}
	}
}
func (a *AdministradorProceso) ejecutarProceso(ctx context.Context) {
	a.maquina.transicionar(EventoProcesoIniciado)
	a.logger.LogInfo(a.config.Name, "iniciando proceso...")
	resultado, err := a.corredor.Run(ctx, a.config, a.logger)
	if err != nil {
		a.logger.LogError(a.config.Name, "error al ejecutar: "+err.Error())
		a.almacen.marcarFallido()
		a.maquina.transicionar(EventoProcesoFallido)
		return
	}
	a.almacen.actualizarDesdeResultado(*resultado)
	a.almacen.marcarIniciado(resultado.PID)
	exito := resultado.Success() && !resultado.TerminatedBySignal
	if exito {
		a.almacen.marcarEjecutando()
		a.maquina.transicionar(EventoProcesoSalido)
	} else {
		a.almacen.marcarFallido()
		a.maquina.transicionar(EventoProcesoFallido)
	}
	codigo := resultado.ExitCode
	reiniciar := false
	switch a.config.RestartPolicy {
	case config.RestartAlways:
		reiniciar = true
	case config.RestartOnFailure:
		reiniciar = codigo != 0
	}
	if !reiniciar {
		a.almacen.marcarDetenido()
		return
	}
	if a.config.MaxRetries > 0 && a.maquina.contarReintentos() >= a.config.MaxRetries {
		a.almacen.marcarDetenido()
		return
	}
	a.almacen.incrementarReintento()
	retardo := restart.CalculateDelay(a.maquina.contarReintentos(), a.config.Backoff)
	tiempo := time.Now().Add(retardo)
	a.almacen.marcarEspera(&tiempo)
	a.maquina.transicionar(EventoReinicioProgramado)
	a.logger.LogInfo(a.config.Name, "reinicio programado en "+retardo.String())
	if err := restart.Wait(ctx, retardo); err != nil {
		a.logger.LogInfo(a.config.Name, "espera cancelada")
		return
	}
	a.ejecutarProceso(ctx)
}
