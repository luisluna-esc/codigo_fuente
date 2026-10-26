#!/bin/bash
# Actualiza Música en el servidor con lo último de GitHub.
# Uso, en el servidor:  bash /opt/musica/codigo/deploy.sh
# Antes de actualizar, haz un respaldo (ver DEPLOY.md).
set -euo pipefail

ROOT=/opt/musica

echo "== Bajando el código"
git -C "$ROOT/codigo" pull --ff-only
echo "   $(git -C "$ROOT/codigo" log --oneline -1)"

echo "== Construyendo la imagen (la versión actual sigue funcionando mientras tanto)"
docker build -q -t musica:latest "$ROOT/codigo" > /dev/null

echo "== Cambiando el contenedor"
# El servidor guarda los datos al cerrar: se le da tiempo para terminar.
docker stop -t 60 musica > /dev/null 2>&1 || true
docker rm musica > /dev/null 2>&1 || true

docker run -d --name musica --restart unless-stopped --stop-timeout 60 \
  -v "$ROOT/data:/data" \
  -p 127.0.0.1:5092:8796 musica:latest > /dev/null

docker image prune -f > /dev/null

echo "== Revisando"
sleep 5
docker ps --format "   {{.Names}}: {{.Status}}" | grep musica
echo "   sitio: $(curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:8092/)"
echo "Listo: http://148.230.73.226:8092"
