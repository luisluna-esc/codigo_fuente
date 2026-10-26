FROM golang:1.23-alpine AS build
RUN apk add --no-cache python3
WORKDIR /src

COPY src/ src/
COPY app/ app/
COPY tools/ tools/
COPY despliegue/ despliegue/

# Arma music-runtime.js/.css y los .gz, igual que tools/compilar.py
RUN python3 tools/preparar_web.py
# Deja en app/ solo lo que lleva el instalador de Windows (tools/empaquetar.py)
RUN python3 -c "import sys; sys.path.insert(0, 'tools'); from pathlib import Path; from empaquetar import runtime_file; [p.unlink() for p in Path('app').rglob('*') if p.is_file() and not runtime_file(p)]"
RUN cd src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -buildvcs=false -trimpath -ldflags="-s -w" -o /out/ServidorMusica ./cmd/server

FROM alpine:3.20 AS runtime
RUN apk add --no-cache tzdata ca-certificates \
    && adduser -D -H -u 10001 musica
ENV TZ=America/La_Paz

WORKDIR /srv
COPY --from=build /out/ServidorMusica /srv/ServidorMusica
COPY --from=build /src/app /srv/app
COPY despliegue/docker-entrypoint.sh /srv/docker-entrypoint.sh
RUN chmod 755 /srv/docker-entrypoint.sh /srv/ServidorMusica

# Configuración, clave de instalación, datos, respaldos y log viven en /data (volumen)
VOLUME /data
USER musica
EXPOSE 8796

ENTRYPOINT ["/srv/docker-entrypoint.sh"]
