#!/bin/sh
# Worker de ejemplo fallido para pruebas futuras.
# No se ejecuta automáticamente en la Parte 1.
echo "[failing-worker] Iniciando trabajador que fallará..."
sleep 1
echo "[failing-worker] Error intencional detectado."
exit 1
