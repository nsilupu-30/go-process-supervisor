# Diagramas y Gráficos Arquitectónicos del Proyecto

Este directorio contiene los diagramas arquitectónicos y de flujo del **Go Process Supervisor**.

---

## 1. Diagrama de Arquitectura General

```mermaid
graph TD
    CLI[Usuario / CLI<br/>validate | run | version] -->|Comando / Config JSON| SUP[Supervisor Global]
    
    subgraph Core System
        SUP --> CFG[Config Loader & Validator<br/>internal/config]
        SUP --> PM[Process Manager<br/>internal/supervisor]
        SUP --> SIG[Signal Handler<br/>internal/signals]
        SUP --> API[API Server HTTP<br/>internal/api]
    end
    
    subgraph Execution Layer
        PM --> PR1[Process Runner 1<br/>worker-estable]
        PM --> PR2[Process Runner 2<br/>worker-falla]
    end

    subgraph Logging Layer
        PR1 -->|Stdout / Stderr| LOG[Process Logger<br/>internal/logging]
        PR2 -->|Stdout / Stderr| LOG
    end

    subgraph Child Processes (OS)
        PR1 -->|os/exec| P1[Proceso Hijo 1<br/>PID 19992]
        PR2 -->|os/exec| P2[Proceso Hijo 2<br/>PID 25112]
    end
```

---

## 2. Diagrama de Flujo de Ejecución del Runner (Parte 2)

```mermaid
sequenceDiagram
    autonumber
    participant CLI as CLI / Main
    participant Runner as ProcessRunner (internal/process)
    participant Log as ProcessLogger (internal/logging)
    participant OS as Kernel (Sistema Operativo)

    CLI->>Runner: Run(ctx, ProcessConfig, Logger)
    Runner->>OS: exec.CommandContext(ctx, Command, Args)
    Runner->>OS: cmd.StdoutPipe() & cmd.StderrPipe()
    Runner->>Log: Iniciar goroutines StreamPipe()
    Runner->>OS: cmd.Start()
    OS-->>Runner: Retorna PID (ej. 19992)
    Runner->>Log: LogInfo("proceso iniciado correctamente con PID 19992")
    
    par Lectura de Logs en Tiempo Real
        OS-->>Log: Flujo Stdout / Stderr
        Log-->>Log: Formatear [HH:MM:SS] [nombre] [STDOUT|STDERR]
    and Espera y Recolección de Recursos (Prevención de Zombis)
        Runner->>OS: cmd.Wait()
        OS-->>Runner: Proceso finalizado (ExitCode 0 o != 0)
    end
    
    Runner->>CLI: Retorna ExecutionResult (PID, ExitCode, Duration)
```

---

## 3. Diagrama de Máquina de Estados por Proceso

```mermaid
stateDiagram-v2
    [*] --> CREATED: Carga de configuración
    CREATED --> STARTING: Invocación cmd.Start()
    
    STARTING --> RUNNING: Proceso iniciado con PID
    STARTING --> FAILED: Error al encontrar ejecutable
    
    RUNNING --> STOPPING: Solicitud de stop / SIGTERM
    STOPPING --> STOPPED: Invocación cmd.Wait() limpia
    
    RUNNING --> BACKOFF: Salida inesperada (ExitCode != 0)
    BACKOFF --> STARTING: Expiración de temporizador exponencial
    BACKOFF --> FAILED: Excedido max_retries
    
    STOPPED --> [*]
    FAILED --> [*]
```
