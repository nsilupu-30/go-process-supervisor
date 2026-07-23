package supervisor

import "context"

type EstadoProceso string

const (
	EstadoCreado     EstadoProceso = "creado"
	EstadoIniciando  EstadoProceso = "iniciando"
	EstadoEjecutando EstadoProceso = "ejecutando"
	EstadoEspera     EstadoProceso = "espera"
	EstadoDeteniendo EstadoProceso = "deteniendo"
	EstadoDetenido   EstadoProceso = "detenido"
	EstadoFallido    EstadoProceso = "fallido"
)

type EventoProceso string

const (
	EventoProcesoIniciado    EventoProceso = "proceso_iniciado"
	EventoProcesoSalido      EventoProceso = "proceso_salido"
	EventoProcesoFallido     EventoProceso = "proceso_fallido"
	EventoReinicioProgramado EventoProceso = "reinicio_programado"
	EventoProcesoDeteniendo  EventoProceso = "proceso_deteniendo"
	EventoProcesoDetenido    EventoProceso = "proceso_detenido"
	EventoApagadoSolicitado  EventoProceso = "apagado_solicitado"
)

type estadoMaquina struct {
	estado             EstadoProceso
	cantidadReintentos int
}

func nuevoEstadoMaquina() *estadoMaquina {
	return &estadoMaquina{estado: EstadoCreado}
}

func (e *estadoMaquina) actual() EstadoProceso {
	return e.estado
}

func (e *estadoMaquina) contarReintentos() int {
	return e.cantidadReintentos
}

func (e *estadoMaquina) transicionar(ev EventoProceso) bool {
	if !e.transicionValida(ev) {
		return false
	}

	switch ev {
	case EventoProcesoIniciado:
		e.estado = EstadoIniciando
	case EventoProcesoSalido, EventoProcesoFallido:
		e.estado = EstadoEspera
		e.cantidadReintentos++
	case EventoReinicioProgramado:
		e.estado = EstadoEspera
	case EventoProcesoDeteniendo:
		e.estado = EstadoDeteniendo
	case EventoProcesoDetenido:
		e.estado = EstadoDetenido
	case EventoApagadoSolicitado:
		if e.estado != EstadoDetenido && e.estado != EstadoFallido {
			e.estado = EstadoDeteniendo
		}
	}
	return true
}

func (e *estadoMaquina) transicionValida(ev EventoProceso) bool {
	switch e.estado {
	case EstadoCreado:
		return ev == EventoProcesoIniciado
	case EstadoIniciando:
		return ev == EventoProcesoSalido || ev == EventoProcesoFallido
	case EstadoEjecutando:
		return ev == EventoProcesoDeteniendo || ev == EventoProcesoSalido || ev == EventoProcesoFallido
	case EstadoEspera:
		return ev == EventoProcesoIniciado
	case EstadoDeteniendo:
		return ev == EventoProcesoDetenido
	case EstadoDetenido, EstadoFallido:
		return false
	default:
		return false
	}
}

func (e *estadoMaquina) apagar(ctx context.Context) {
	e.transicionar(EventoApagadoSolicitado)
}
