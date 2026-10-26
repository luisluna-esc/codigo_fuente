$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$exe = Join-Path $root 'ServidorMusica.exe'
$config = Join-Path $root 'server-config.json'
if (!(Test-Path $config)) { throw 'Inicia Música una vez para crear la configuración antes de registrar el arranque.' }
$action = New-ScheduledTaskAction -Execute $exe -Argument ('--config "' + $config + '"') -WorkingDirectory $root
$trigger = New-ScheduledTaskTrigger -AtStartup
$principal = New-ScheduledTaskPrincipal -UserId 'SYSTEM' -LogonType ServiceAccount -RunLevel Highest
$settings = New-ScheduledTaskSettingsSet -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1) -ExecutionTimeLimit ([TimeSpan]::Zero) -MultipleInstances IgnoreNew
Register-ScheduledTask -TaskName 'MinisterioMusicaOnline' -Action $action -Trigger $trigger -Principal $principal -Settings $settings -Description 'Servidor independiente del Ministerio de la Música' -Force | Out-Null
Write-Host 'Arranque registrado. Guarda el paquete en una carpeta permanente accesible para SYSTEM.'
