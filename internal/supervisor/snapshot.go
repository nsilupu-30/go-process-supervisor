package supervisor

import (
	"sync"
	"time"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
	"github.com/nsilupu-30/go-process-supervisor/internal/process"
)

type SnapshotProceso struct {
	Nombre       string        `json:"name"`
	Estado       EstadoProceso `json:"state"`
	PID          int           `json:"pid,omitempty"`
	Reinicios    int           `json:"restart_count"`
	CodigoSalida int           `json:"exit_code,omitempty"`
	Error        string        `json:"error,omitempty"`
	Inicio       time.Time     `json:"started_at,omitempty"`
	Salida       time.Time     `json:"exited_at,omitempty"`
	Siguiente    *time.Time    `json:"next_retry_at,omitempty"`
}

type almacenSnapshots struct {
	mu            sync.RWMutex
	config        config.ProcessConfig
	estado        EstadoProceso
	pid           int
	reintentos    int
	salida        int
	mensajeError string
	inicio        time.Time
	salidaT       time.Time
	siguiente     *time.Time
}

func nuevoAlmacenSnapshots(cfg config.ProcessConfig) *almacenSnapshots {
	return &almacenSnapshots{config: cfg, estado: EstadoCreado}
}

func (a *almacenSnapshots) marcarIniciado(pid int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pid = pid
	a.estado = EstadoIniciando
	if a.inicio.IsZero() {
		a.inicio = time.Now()
	}
}

func (a *almacenSnapshots) marcarEjecutando() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.estado = EstadoEjecutando
}

func (a *almacenSnapshots) marcarEspera(siguiente *time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.estado = EstadoEspera
	a.siguiente = siguiente
}

func (a *almacenSnapshots) marcarDeteniendo() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.estado = EstadoDeteniendo
}

func (a *almacenSnapshots) marcarDetenido() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.estado = EstadoDetenido
	if a.salidaT.IsZero() {
		a.salidaT = time.Now()
	}
}

func (a *almacenSnapshots) marcarFallido() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.estado = EstadoFallido
	if a.salidaT.IsZero() {
		a.salidaT = time.Now()
	}
}

func (a *almacenSnapshots) actualizarDesdeResultado(r process.ExecutionResult) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pid = r.PID
	a.salida = r.ExitCode
	if r.Error != nil {
		a.mensajeError = r.Error.Error()
	}
	if r.ExitedAt.IsZero() {
		a.salidaT = time.Now()
	} else {
		a.salidaT = r.ExitedAt
	}
}

func (a *almacenSnapshots) obtenerSnapshot() SnapshotProceso {
	a.mu.RLock()
	defer a.mu.RUnlock()
	ss := SnapshotProceso{
		Nombre:       a.config.Name,
		Estado:       a.estado,
		PID:          a.pid,
		Reinicios:    a.reintentos,
		CodigoSalida: a.salida,
		Error:        a.mensajeError,
		Inicio:       a.inicio,
		Salida:       a.salidaT,
	}
	if a.siguiente != nil {
		ss.Siguiente = a.siguiente
	}
	return ss
}

func (a *almacenSnapshots) incrementarReintento() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.reintentos++
}

func (a *almacenSnapshots) contarReintentos() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.reintentos
}
