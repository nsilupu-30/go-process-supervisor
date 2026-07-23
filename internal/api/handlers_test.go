package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nsilupu-30/go-process-supervisor/internal/supervisor"
)

type mockSupervisor struct {
	health       bool
	snapshots    []supervisor.SnapshotProceso
	lastCommand  string
	lastName     string
	reloadCalled bool
}

func (m *mockSupervisor) Health() bool {
	return m.health
}

func (m *mockSupervisor) ProcessSnapshots() []supervisor.SnapshotProceso {
	return m.snapshots
}

func (m *mockSupervisor) ProcessSnapshot(name string) (supervisor.SnapshotProceso, bool) {
	for _, snap := range m.snapshots {
		if snap.Nombre == name {
			return snap, true
		}
	}
	return supervisor.SnapshotProceso{}, false
}

func (m *mockSupervisor) StartProcess(name string, ctx context.Context) error {
	m.lastCommand = "start"
	m.lastName = name
	return nil
}

func (m *mockSupervisor) StopProcess(name string, ctx context.Context) error {
	m.lastCommand = "stop"
	m.lastName = name
	return nil
}

func (m *mockSupervisor) RestartProcess(name string, ctx context.Context) error {
	m.lastCommand = "restart"
	m.lastName = name
	return nil
}

func (m *mockSupervisor) Reload() error {
	m.reloadCalled = true
	return nil
}

func executeRequest(t *testing.T, srv *Server, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rr, req)
	return rr
}

func TestHandleHealth(t *testing.T) {
	sup := &mockSupervisor{health: true}
	server := NewServer("127.0.0.1:0", sup)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := executeRequest(t, server, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload["message"] != "ok" {
		t.Fatalf("expected ok message, got %q", payload["message"])
	}
}

func TestHandleProcesses(t *testing.T) {
	sup := &mockSupervisor{snapshots: []supervisor.SnapshotProceso{{Nombre: "worker1", Estado: supervisor.EstadoCreado}}}
	server := NewServer("127.0.0.1:0", sup)

	req := httptest.NewRequest(http.MethodGet, "/processes", nil)
	rr := executeRequest(t, server, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload []supervisor.SnapshotProceso
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(payload) != 1 || payload[0].Nombre != "worker1" {
		t.Fatalf("unexpected snapshots response: %#v", payload)
	}
}

func TestHandleProcessByName(t *testing.T) {
	sup := &mockSupervisor{snapshots: []supervisor.SnapshotProceso{{Nombre: "worker1", Estado: supervisor.EstadoCreado}}}
	server := NewServer("127.0.0.1:0", sup)

	req := httptest.NewRequest(http.MethodGet, "/processes/worker1", nil)
	rr := executeRequest(t, server, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var payload supervisor.SnapshotProceso
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Nombre != "worker1" {
		t.Fatalf("expected process worker1, got %q", payload.Nombre)
	}
}

func TestHandleProcessNotFound(t *testing.T) {
	sup := &mockSupervisor{snapshots: []supervisor.SnapshotProceso{}}
	server := NewServer("127.0.0.1:0", sup)

	req := httptest.NewRequest(http.MethodGet, "/processes/missing", nil)
	rr := executeRequest(t, server, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}

func TestHandleProcessCommands(t *testing.T) {
	sup := &mockSupervisor{snapshots: []supervisor.SnapshotProceso{{Nombre: "worker1", Estado: supervisor.EstadoCreado}}}
	server := NewServer("127.0.0.1:0", sup)

	for _, tc := range []struct {
		path      string
		command   string
		expectCmd string
	}{
		{"/processes/worker1/start", "start", "start"},
		{"/processes/worker1/stop", "stop", "stop"},
		{"/processes/worker1/restart", "restart", "restart"},
	} {
		req := httptest.NewRequest(http.MethodPost, tc.path, nil)
		rr := executeRequest(t, server, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200 for %s, got %d", tc.path, rr.Code)
		}
		if sup.lastCommand != tc.expectCmd || sup.lastName != "worker1" {
			t.Fatalf("expected command %s for worker1, got %s on %s", tc.expectCmd, sup.lastCommand, sup.lastName)
		}
	}
}

func TestHandleReload(t *testing.T) {
	sup := &mockSupervisor{}
	server := NewServer("127.0.0.1:0", sup)

	req := httptest.NewRequest(http.MethodPost, "/reload", nil)
	rr := executeRequest(t, server, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if !sup.reloadCalled {
		t.Fatal("expected reload to be called")
	}
}
