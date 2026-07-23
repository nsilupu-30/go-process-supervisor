# Diagrama de Arquitectura

```mermaid
graph TD
    CLI[Usuario / CLI<br/>validate | run | version]
    SUP[Supervisor Global<br/>cmd/supervisor/main.go]
    CFG[Config Loader & Validator<br/>internal/config]
    PM[Process Manager<br/>internal/supervisor]
    SIG[Manejador de Señales<br/>internal/senales]
    API[API Server HTTP<br/>internal/api]
    PR1[Process Runner 1<br/>internal/process]
    PR2[Process Runner 2<br/>internal/process]
    LOG[Logger de Procesos<br/>internal/logging]
    OS1[Proceso Hijo 1]
    OS2[Proceso Hijo 2]

    CLI -->|Carga config JSON| SUP
    SUP --> CFG
    SUP --> PM
    SUP --> SIG
    SUP --> API
    PM --> PR1
    PM --> PR2
    PR1 -->|stdout/stderr| LOG
    PR2 -->|stdout/stderr| LOG
    PR1 -->|os/exec| OS1
    PR2 -->|os/exec| OS2
```