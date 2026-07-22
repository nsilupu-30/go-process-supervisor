package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "validate":
		runValidateCommand(os.Args[2:])
	case "version":
		fmt.Println("Go Process Supervisor v0.1.0 (Parte 1: Fundación y Configuración)")
	default:
		fmt.Fprintf(os.Stderr, "Error: comando desconocido %q\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// runValidateCommand implementa la ejecución del subcomando 'validate'.
// Analiza el flag --config, carga y valida el archivo JSON sin iniciar ningún proceso.
func runValidateCommand(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	configPath := fs.String("config", "", "Ruta al archivo de configuración JSON")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error al procesar argumentos de validate: %v\n", err)
		os.Exit(1)
	}

	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "Error: la opción --config es obligatoria.")
		fmt.Fprintln(os.Stderr, "Uso correcto: go run ./cmd/supervisor validate --config examples/config.example.json")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error de validación de configuración:\n  %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Configuración válida: %s (%d procesos configurados)\n", *configPath, len(cfg.Processes))
}

func printUsage() {
	fmt.Println("Go Process Supervisor CLI")
	fmt.Println()
	fmt.Println("Comandos disponibles:")
	fmt.Println("  validate --config <ruta_json>   Valida la sintaxis y reglas del archivo de configuración")
	fmt.Println("  version                         Muestra la versión del supervisor")
}
