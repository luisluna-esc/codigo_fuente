from pathlib import Path
import argparse, os, subprocess, sys
root = Path(__file__).resolve().parents[1]
subprocess.run([sys.executable, str(root / "tools/preparar_web.py")], check=True)
parser = argparse.ArgumentParser(description="Compilar Ministerio de Música")
parser.add_argument("--platform", choices=["windows", "linux"], default="windows")
platform = parser.parse_args().platform
base_env = os.environ.copy()
env = {**base_env, "CGO_ENABLED": "0", "GOOS": platform, "GOARCH": "amd64"}
def go(args, build_env=env):
    subprocess.run(["go", *args], cwd=root / "src", env=build_env, check=True)
output = "ServidorMusica.exe" if platform == "windows" else "ServidorMusica"
go(["build", "-buildvcs=false", "-trimpath", "-ldflags=-s -w", "-o", str(root / output), "./cmd/server"])
if platform == "windows":
    go(["build", "-buildvcs=false", "-trimpath", "-ldflags=-s -w -H=windowsgui", "-o", str(root / "Musica.exe"), "./cmd/launcher"])
    go(["run", "./tools/iconpatch", "../Musica.ico", "../Musica.exe"], base_env)
    go(["build", "-buildvcs=false", "-trimpath", "-ldflags=-s -w -H=windowsgui", "-o", str(root / "DESINSTALAR_MinisterioMusica.exe"), "./cmd/uninstaller"])
    go(["run", "./tools/iconpatch", "../Musica.ico", "../DESINSTALAR_MinisterioMusica.exe"], base_env)
    subprocess.run([sys.executable, str(root / "tools/empaquetar.py")], check=True)
    go(["build", "-buildvcs=false", "-trimpath", "-ldflags=-s -w", "-o", str(root / "INSTALAR_MinisterioMusica.exe"), "./cmd/installer"])
    go(["run", "./tools/iconpatch", "../Musica.ico", "../INSTALAR_MinisterioMusica.exe"], base_env)
print("Compilación terminada:", platform)
