package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadConfig abre y des-serializa un archivo de configuración JSON, aplicando la validación de reglas de negocio.
// Utiliza fmt.Errorf con %w para envolver los errores subyacentes y mantener el contexto de la ruta.
func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cargar configuración %q: %w", path, err)
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cargar configuración %q: JSON inválido: %w", path, err)
	}

	if err := ValidateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("cargar configuración %q: %w", path, err)
	}

	return &cfg, nil
}
