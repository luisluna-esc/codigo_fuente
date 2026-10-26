from pathlib import Path
from zipfile import ZipFile, ZIP_DEFLATED

ROOT = Path(__file__).resolve().parents[1]
PAYLOAD = ROOT / 'src/cmd/installer/payload.zip'
RUNTIME_FILES = [
    'ServidorMusica.exe', 'Musica.exe', 'DESINSTALAR_MinisterioMusica.exe', 'Musica.ico', 'DETENER_MUSICA.cmd',
    'INICIAR_MUSICA.cmd', 'INICIAR_MUSICA.ps1', 'EXTRAER_SOLO_MUSICA.cmd',
    'MANUAL_USO.md', 'LEEME_PRIMERO.txt', 'VERSION.json', 'VALIDACION.txt',
    'LICENCIAS_TERCEROS.txt', 'FUENTES_INCLUIDAS.txt',
    'AUTORIA_Y_USO.txt',
    'server-config.example.json', 'acceso-servidor.example.json',
    'CONFIGURACION_ACCESO.md',
]
RUNTIME_DIRS = ['app', 'despliegue']
WEB_RUNTIME = {
    'music-runtime.js', 'music-runtime.css', 'auth-core.css',
    'auth-boot.js', 'music-context.js', 'inspiration-data.js',
    'music-transfer.js', 'music-media-612.js', 'photo-worker.js',
}

def runtime_file(path):
    if path.parts[0] != 'app':
        return True
    if len(path.parts) > 2:  # logos and other static assets
        return True
    name = path.name.removesuffix('.gz')
    return name in WEB_RUNTIME or Path(name).suffix in {'.html', '.json', '.docx', '.png', '.ico'}

def build_payload():
    for name in RUNTIME_FILES:
        if not (ROOT / name).is_file():
            raise SystemExit('Falta el archivo: ' + name)
    with ZipFile(PAYLOAD, 'w', compression=ZIP_DEFLATED, compresslevel=9) as z:
        for name in RUNTIME_FILES:
            z.write(ROOT / name, name)
        for directory in RUNTIME_DIRS:
            for p in sorted((ROOT / directory).rglob('*')):
                if p.is_file() and runtime_file(p.relative_to(ROOT)):
                    z.write(p, str(p.relative_to(ROOT)))
    print('Paquete interno de instalador listo:', PAYLOAD.name)

if __name__ == '__main__':
    build_payload()
