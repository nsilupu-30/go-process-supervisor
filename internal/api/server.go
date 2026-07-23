package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nsilupu-30/go-process-supervisor/internal/supervisor"
)

type Supervisor interface {
	Health() bool
	ProcessSnapshots() []supervisor.SnapshotProceso
	ProcessSnapshot(name string) (supervisor.SnapshotProceso, bool)
	StartProcess(name string, ctx context.Context) error
	StopProcess(name string, ctx context.Context) error
	RestartProcess(name string, ctx context.Context) error
	Reload() error
}

type Server struct {
	httpServer *http.Server
	supervisor Supervisor
}

func NewServer(address string, sup Supervisor) *Server {
	mux := http.NewServeMux()

	srv := &Server{
		supervisor: sup,
		httpServer: &http.Server{
			Addr:    address,
			Handler: mux,
		},
	}

	srv.registerRoutes(mux)
	return srv
}

func (s *Server) Start() error {
	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error en servidor http: %w", err)
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("error cerrando el servidor http: %w", err)
	}
	return nil
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/processes", s.handleProcesses)
	mux.HandleFunc("/processes/", s.handleProcessAction)
	mux.HandleFunc("/reload", s.handleReload)
}
