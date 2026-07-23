package api

import (
    "encoding/json"
    "net/http"
    "strings"

    "github.com/nsilupu-30/go-process-supervisor/internal/supervisor"
)

type apiError struct {
    Error string `json:"error"`
}

type apiMessage struct {
    Message string `json:"message"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        methodNotAllowed(w)
        return
    }

    if !s.supervisor.Health() {
        writeError(w, http.StatusServiceUnavailable, "supervisor unhealthy")
        return
    }

    writeJSON(w, http.StatusOK, apiMessage{Message: "ok"})
}

func (s *Server) handleProcesses(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        methodNotAllowed(w)
        return
    }

    writeJSON(w, http.StatusOK, s.supervisor.ProcessSnapshots())
}

func (s *Server) handleProcessAction(w http.ResponseWriter, r *http.Request) {
    if strings.TrimPrefix(r.URL.Path, "/processes") == "" {
        writeError(w, http.StatusNotFound, "endpoint not found")
        return
    }

    path := strings.TrimPrefix(r.URL.Path, "/processes/")
    path = strings.Trim(path, "/")
    if path == "" {
        writeError(w, http.StatusNotFound, "endpoint not found")
        return
    }

    parts := strings.Split(path, "/")
    name := parts[0]
    if name == "" {
        writeError(w, http.StatusNotFound, "process name is required")
        return
    }

    if len(parts) == 1 {
        if r.Method != http.MethodGet {
            methodNotAllowed(w)
            return
        }

        snapshot, ok := s.supervisor.ProcessSnapshot(name)
        if !ok {
            writeError(w, http.StatusNotFound, "process not found")
            return
        }

        writeJSON(w, http.StatusOK, snapshot)
        return
    }

    if len(parts) == 2 && r.Method == http.MethodPost {
        var err error
        switch parts[1] {
        case "start":
            err = s.supervisor.StartProcess(name, r.Context())
        case "stop":
            err = s.supervisor.StopProcess(name, r.Context())
        case "restart":
            err = s.supervisor.RestartProcess(name, r.Context())
        default:
            writeError(w, http.StatusNotFound, "endpoint not found")
            return
        }

        if err != nil {
            if err == supervisor.ErrProcessNotFound {
                writeError(w, http.StatusNotFound, err.Error())
                return
            }
            writeError(w, http.StatusInternalServerError, err.Error())
            return
        }

        writeJSON(w, http.StatusOK, apiMessage{Message: "command accepted"})
        return
    }

    writeError(w, http.StatusNotFound, "endpoint not found")
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        methodNotAllowed(w)
        return
    }

    if err := s.supervisor.Reload(); err != nil {
        writeError(w, http.StatusInternalServerError, err.Error())
        return
    }

    writeJSON(w, http.StatusOK, apiMessage{Message: "reload triggered"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
    writeJSON(w, status, apiError{Error: message})
}

func methodNotAllowed(w http.ResponseWriter) {
    writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}
