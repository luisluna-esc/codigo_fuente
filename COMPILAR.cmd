@echo off
setlocal
cd /d "%~dp0"
python tools\compilar.py --platform windows
if errorlevel 1 exit /b 1
pushd src
go test -timeout 90s ./...
set result=%errorlevel%
popd
exit /b %result%
