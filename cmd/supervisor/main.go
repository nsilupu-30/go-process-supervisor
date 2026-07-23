package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
	"github.com/nsilupu-30/go-process-supervisor/internal/logging"
	"github.com/nsilupu-30/go-process-supervisor/internal/supervisor"
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
	case "run", "start":
		runStartCommand(os.Args[2:])
	case "version":
		fmt.Println("Go Process Supervisor v0.2.0 (Parte 2: Ejecución de Procesos)")
	default:
		fmt.Fprintf(os.Stderr, "Error: comando desconocido %q\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// runValidateCommand implementa la ejecución del subcomando 'validate'.
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

// runStartCommand ejecuta los procesos definidos en la configuración usando ProcessRunner.
func runStartCommand(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	configPath := fs.String("config", "", "Ruta al archivo de configuración JSON")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error al procesar argumentos de run: %v\n", err)
		os.Exit(1)
	}

	if *configPath == "" {
		fmt.Fprintln(os.Stderr, "Error: la opción --config es obligatoria.")
		fmt.Fprintln(os.Stderr, "Uso correcto: go run ./cmd/supervisor run --config examples/config.windows.json")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al cargar la configuración:\n  %v\n", err)
		os.Exit(1)
	}

	logger := logging.NewProcessLogger()
	sup := supervisor.NuevoSupervisor(*cfg, logger)
	ctx := context.Background()

	logger.LogInfo("SUPERVISOR", fmt.Sprintf("Iniciando ejecución de %d proceso(s) desde %s...", len(cfg.Processes), *configPath))

	go sup.Iniciar(ctx)

	// Esperar Ctrl+C
	esperarSenal()

	ctxCancelado, cancel := context.WithCancel(ctx)
	defer cancel()
	sup.Detener(ctxCancelado)
	logger.LogInfo("SUPERVISOR", "Supervisor detenido correctamente.")
}

func printUsage() {
	fmt.Println("Go Process Supervisor CLI")
	fmt.Println()
	fmt.Println("Comandos disponibles:")
	fmt.Println("  validate --config <ruta_json>   Valida la sintaxis y reglas del archivo de configuración")
	fmt.Println("  run --config <ruta_json>        Ejecuta los procesos de la configuración y muestra sus logs")
	fmt.Println("  version                         Muestra la versión del supervisor")
}

func esperarSenal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
}
