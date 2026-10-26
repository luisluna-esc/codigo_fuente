#!/bin/sh
# Arranque del contenedor de Música.
set -eu

CONFIG=/data/server-config.json

# Primer arranque: la configuración por defecto del servidor escucha en 127.0.0.1 y busca
# la interfaz junto a la configuración; dentro de Docker debe escuchar en 0.0.0.0 y usar /srv/app.
if [ ! -f "$CONFIG" ]; then
  gen() { tr -dc 'A-Za-z0-9' </dev/urandom | head -c 32; }
  umask 077
  cat > "$CONFIG" <<EOF
{
  "bindAddress": "0.0.0.0",
  "port": 8796,
  "adminToken": "$(gen)",
  "clientToken": "$(gen)",
  "dataFile": "data/musica-server-data.json",
  "serverName": "Ministerio de la Música",
  "webRoot": "/srv/app",
  "publicWeb": true,
  "allowedOrigins": []
}
EOF
  echo "Configuración creada en $CONFIG"
fi

# El bloqueo guarda el PID del servidor. Tras una caída del contenedor el proceso nuevo
# vuelve a tener el mismo PID y el servidor creería que sigue activo. Este contenedor es el
# único dueño de /data, así que un bloqueo al arrancar siempre es de una ejecución anterior.
rm -f /data/musica-server.lock

exec /srv/ServidorMusica --config "$CONFIG"
