#!/bin/sh
set -eu
cd "$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
python3 tools/compilar.py --platform linux
python3 tools/compilar.py --platform windows
cd src
go test -timeout 90s ./...
