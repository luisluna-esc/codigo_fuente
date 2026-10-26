@echo off
cd /d "%~dp0"
echo Exporta primero un respaldo completo desde el Administrador anterior.
set /p "music_backup=Ruta del respaldo JSON o .cabackup: "
set /p "music_output=Ruta nueva de salida JSON (por ejemplo C:\Copias\Musica.json): "
ServidorMusica.exe --extraer-musica "%music_backup%" --salida "%music_output%"
pause
