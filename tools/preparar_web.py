"""Build ordered local bundles and reproducible precompressed web assets."""
from pathlib import Path
import gzip
import re
from empaquetar import runtime_file

ROOT = Path(__file__).resolve().parents[1]
APP = ROOT / "app"

def build_web():
    boot = (APP / "auth-boot.js").read_text(encoding="utf-8")
    for kind, prefix, output in [
        ("scripts", ["class-policy.js", "sgl-v4-migration.js"], "music-runtime.js"),
        ("styles", [], "music-runtime.css"),
    ]:
        match = re.search(r"const " + kind + r"=\[(.*?)\];", boot, re.S)
        if not match:
            raise ValueError("No se encontró el catálogo: " + kind)
        names = prefix + re.findall(r"'([^']+)'", match.group(1))
        parts = []
        for name in names:
            text = (APP / name).read_text(encoding="utf-8")
            parts.append("/* Archivo: " + name + " */\n" + text)
        separator = "\n;\n" if kind == "scripts" else "\n\n"
        (APP / output).write_text(separator.join(parts) + "\n", encoding="utf-8")
    for path in sorted(APP.rglob("*")):
        if path.is_file() and path.suffix in {".js", ".css", ".html", ".json", ".svg"}:
            compressed = path.with_name(path.name + ".gz")
            if runtime_file(Path("app") / path.relative_to(APP)):
                compressed.write_bytes(gzip.compress(path.read_bytes(), compresslevel=9, mtime=0))
            elif compressed.exists():
                compressed.unlink()
    print("Interfaz preparada: dos paquetes locales, sin dependencias externas.")

if __name__ == "__main__":
    build_web()
