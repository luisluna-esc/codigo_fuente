#!/bin/bash
# Mantiene las tablas relacionales de la base musica al día con lo que guarda el sistema.
# El sistema guarda su estado en el esquema legacy; este script lo exporta y, sólo si cambió,
# reconstruye las tablas en una transacción (si algo falla, quedan como estaban).
# Lo ejecuta cron cada 10 minutos (ver DEPLOY.md). Uso manual: bash actualizar-tablas.sh
set -euo pipefail

SCRIPS=${SCRIPS:-/root/musica-scrips}
CONFIG=/opt/musica/data/server-config.json
STATE=/var/lib/musica-tablas
mkdir -p "$STATE"
exec 9>"$STATE/lock"
flock -n 9 || exit 0

source /root/musica-db.env
export PGPASSWORD="$DB_PASSWORD"
TOKEN=$(grep -o '"adminToken": *"[^"]*"' "$CONFIG" | sed 's/.*"\([^"]*\)"$/\1/')

umask 077
EXPORT=$(mktemp /dev/shm/musica-export.XXXXXX)
trap 'rm -f "$EXPORT"' EXIT

curl -sSf --max-time 300 -H "X-SGL-Token: $TOKEN" -o "$EXPORT" http://127.0.0.1:5092/api/music-data/export
HASH=$(python3 "$SCRIPS/importarRespaldo.py" --huella "$EXPORT")
if [ -f "$STATE/huella" ] && [ "$(cat "$STATE/huella")" = "$HASH" ]; then
  exit 0
fi

python3 "$SCRIPS/importarRespaldo.py" --reemplazar "$EXPORT" 2>>"$STATE/avisos.log" \
  | psql -v ON_ERROR_STOP=1 -q -h 127.0.0.1 -U musica -d musica > /dev/null
echo "$HASH" > "$STATE/huella"
echo "$(date '+%F %T') tablas actualizadas" >> "$STATE/actualizaciones.log"
