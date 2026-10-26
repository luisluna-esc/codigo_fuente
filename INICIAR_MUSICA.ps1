$ErrorActionPreference = 'Stop'
Start-Process -FilePath (Join-Path $PSScriptRoot 'Musica.exe') -WorkingDirectory $PSScriptRoot
