# Diagrama de Flujo de Apagado

```mermaid
flowchart TD
    A[Señal SIGINT/SIGTERM recibida] --> B[Cancelar contexto global]
    B --> C[Detener supervisor]
    C --> D[Enviar comando detener a cada proceso]
    D --> E[Permitir que procesos terminen limpiamente]
    E --> F[Esperar grace period]
    F --> G[Forzar terminación si el proceso no responde]
    G --> H[Detener servidor HTTP con Shutdown(ctx)]
    H --> I[Cerrar logs y salir]
```