package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
)

func TestLoadConfigTableDriven(t *testing.T) {
	tempDir := t.TempDir()

	validJSON := `{
		"grace_period_seconds": 10,
		"api_address": "127.0.0.1:8080",
		"processes": [
			{
				"name": "proc-1",
				"command": "/bin/echo",
				"args": ["hello"],
				"working_dir": ".",
				"restart_policy": "always",
				"max_retries": 5,
				"backoff": {
					"initial_seconds": 1,
					"factor": 2.0,
					"max_seconds": 30
				}
			},
			{
				"name": "proc-2",
				"command": "/bin/ls",
				"args": ["-la"],
				"working_dir": "/tmp",
				"restart_policy": "on-failure",
				"max_retries": 0,
				"backoff": {
					"initial_seconds": 2,
					"factor": 1.5,
					"max_seconds": 10
				}
			}
		]
	}`

	invalidJSON := `{ "grace_period_seconds": 10, "processes": [ `

	emptyProcessesJSON := `{
		"grace_period_seconds": 5,
		"api_address": "127.0.0.1:8080",
		"processes": []
	}`

	duplicateNameJSON := `{
		"grace_period_seconds": 5,
		"api_address": "127.0.0.1:8080",
		"processes": [
			{
				"name": "worker-1",
				"command": "/bin/sleep",
				"restart_policy": "always",
				"max_retries": 1,
				"backoff": {"initial_seconds": 1, "factor": 2, "max_seconds": 10}
			},
			{
				"name": "worker-1",
				"command": "/bin/ls",
				"restart_policy": "never",
				"max_retries": 1,
				"backoff": {"initial_seconds": 1, "factor": 2, "max_seconds": 10}
			}
		]
	}`

	emptyCommandJSON := `{
		"grace_period_seconds": 5,
		"api_address": "127.0.0.1:8080",
		"processes": [
			{
				"name": "worker-1",
				"command": "   ",
				"restart_policy": "always",
				"max_retries": 1,
				"backoff": {"initial_seconds": 1, "factor": 2, "max_seconds": 10}
			}
		]
	}`

	invalidPolicyJSON := `{
		"grace_period_seconds": 5,
		"api_address": "127.0.0.1:8080",
		"processes": [
			{
				"name": "worker-1",
				"command": "/bin/sleep",
				"restart_policy": "sometimes",
				"max_retries": 1,
				"backoff": {"initial_seconds": 1, "factor": 2, "max_seconds": 10}
			}
		]
	}`

	invalidBackoffFactorJSON := `{
		"grace_period_seconds": 5,
		"api_address": "127.0.0.1:8080",
		"processes": [
			{
				"name": "worker-1",
				"command": "/bin/sleep",
				"restart_policy": "always",
				"max_retries": 1,
				"backoff": {"initial_seconds": 1, "factor": 0.5, "max_seconds": 10}
			}
		]
	}`

	invalidBackoffMaxSecJSON := `{
		"grace_period_seconds": 5,
		"api_address": "127.0.0.1:8080",
		"processes": [
			{
				"name": "worker-1",
				"command": "/bin/sleep",
				"restart_policy": "always",
				"max_retries": 1,
				"backoff": {"initial_seconds": 5, "factor": 2, "max_seconds": 2}
			}
		]
	}`

	invalidGracePeriodJSON := `{
		"grace_period_seconds": 0,
		"api_address": "127.0.0.1:8080",
		"processes": [
			{
				"name": "worker-1",
				"command": "/bin/sleep",
				"restart_policy": "always",
				"max_retries": 1,
				"backoff": {"initial_seconds": 1, "factor": 2, "max_seconds": 10}
			}
		]
	}`

	invalidAPIAddressJSON := `{
		"grace_period_seconds": 5,
		"api_address": "invalid-address-without-port",
		"processes": [
			{
				"name": "worker-1",
				"command": "/bin/sleep",
				"restart_policy": "always",
				"max_retries": 1,
				"backoff": {"initial_seconds": 1, "factor": 2, "max_seconds": 10}
			}
		]
	}`

	createTempFile := func(name, content string) string {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("error creando archivo temporal de prueba %s: %v", name, err)
		}
		return path
	}

	tests := []struct {
		name          string
		filePath      string
		fileContent   string
		isNonExistent bool
		wantErr       bool
		errSubstring  string
	}{
		{
			name:        "Configuración válida",
			filePath:    createTempFile("valid.json", validJSON),
			fileContent: validJSON,
			wantErr:     false,
		},
		{
			name:         "JSON malformado",
			filePath:     createTempFile("invalid.json", invalidJSON),
			fileContent:  invalidJSON,
			wantErr:      true,
			errSubstring: "JSON inválido",
		},
		{
			name:          "Archivo de configuración inexistente",
			filePath:      filepath.Join(tempDir, "no_existe.json"),
			isNonExistent: true,
			wantErr:       true,
			errSubstring:  "cargar configuración",
		},
		{
			name:         "Lista de procesos vacía",
			filePath:     createTempFile("empty_proc.json", emptyProcessesJSON),
			wantErr:      true,
			errSubstring: "no puede estar vacía",
		},
		{
			name:         "Nombres de procesos duplicados",
			filePath:     createTempFile("duplicate.json", duplicateNameJSON),
			wantErr:      true,
			errSubstring: "duplicado",
		},
		{
			name:         "Comando obligatorio vacío",
			filePath:     createTempFile("empty_cmd.json", emptyCommandJSON),
			wantErr:      true,
			errSubstring: "el comando ('command') es obligatorio",
		},
		{
			name:         "Política de reinicio no válida",
			filePath:     createTempFile("invalid_policy.json", invalidPolicyJSON),
			wantErr:      true,
			errSubstring: "política de reinicio inválida",
		},
		{
			name:         "Factor de backoff menor a 1.0",
			filePath:     createTempFile("invalid_factor.json", invalidBackoffFactorJSON),
			wantErr:      true,
			errSubstring: "backoff.factor debe ser mayor o igual a 1.0",
		},
		{
			name:         "Max seconds de backoff menor a initial seconds",
			filePath:     createTempFile("invalid_maxsec.json", invalidBackoffMaxSecJSON),
			wantErr:      true,
			errSubstring: "debe ser mayor o igual a initial_seconds",
		},
		{
			name:         "Grace period no positivo",
			filePath:     createTempFile("invalid_grace.json", invalidGracePeriodJSON),
			wantErr:      true,
			errSubstring: "grace_period_seconds debe ser positivo",
		},
		{
			name:         "Dirección API sin puerto",
			filePath:     createTempFile("invalid_api.json", invalidAPIAddressJSON),
			wantErr:      true,
			errSubstring: "api_address inválido",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.LoadConfig(tt.filePath)
			if tt.wantErr {
				if err == nil {
					t.Errorf("se esperaba un error pero LoadConfig devolvió nil")
					return
				}
				if tt.errSubstring != "" && !strings.Contains(err.Error(), tt.errSubstring) {
					t.Errorf("error devuelto = %q, se esperaba subcadena %q", err.Error(), tt.errSubstring)
				}
			} else {
				if err != nil {
					t.Errorf("no se esperaba error pero se obtuvo: %v", err)
					return
				}
				if cfg == nil {
					t.Errorf("se esperaba AppConfig no nulo")
					return
				}
				if len(cfg.Processes) != 2 {
					t.Errorf("se esperaban 2 procesos cargados, se obtuvo: %d", len(cfg.Processes))
				}
			}
		})
	}
}
