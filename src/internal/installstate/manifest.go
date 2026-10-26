// Package installstate records the exact program files owned by an installation.
// User databases, configuration, backups and browser profiles are never owned.
package installstate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const ManifestName = "instalacion-musica.json"
const UninstallerName = "DESINSTALAR_MinisterioMusica.exe"
const product = "ministerio-musica-online"

type Manifest struct {
	Product string            `json:"product"`
	Version string            `json:"version"`
	Root    string            `json:"root"`
	Files   map[string]string `json:"files"` // slash path -> installed SHA-256
}

func CanonicalRoot(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// ProgramPath rejects traversal, alternate Windows data streams, protected user
// files, and paths through symlinks or directory junctions.
func ProgramPath(root, name string) (string, error) {
	if name == "" || strings.ContainsAny(name, "\\:\x00") || strings.HasPrefix(name, "/") {
		return "", fmt.Errorf("ruta de programa inválida: %q", name)
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(name)))
	if clean != name || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("ruta de programa inválida: %q", name)
	}
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if strings.TrimRight(part, " .") != part {
			return "", fmt.Errorf("ruta de programa inválida: %q", name)
		}
	}
	first := strings.ToLower(parts[0])
	if len(parts) > 1 {
		if first != "app" && first != "despliegue" {
			return "", fmt.Errorf("carpeta protegida: %s", name)
		}
	} else {
		allowed := map[string]bool{
			"servidormusica.exe": true, "musica.exe": true, "musica.ico": true,
			strings.ToLower(UninstallerName): true, "detener_musica.cmd": true,
			"iniciar_musica.cmd": true, "iniciar_musica.ps1": true, "extraer_solo_musica.cmd": true,
			"manual_uso.md": true, "leeme_primero.txt": true, "version.json": true, "validacion.txt": true,
			"licencias_terceros.txt": true, "fuentes_incluidas.txt": true, "autoria_y_uso.txt": true,
			"server-config.example.json": true, "acceso-servidor.example.json": true, "configuracion_acceso.md": true,
		}
		versionIcon := strings.HasPrefix(first, "musica-") && strings.HasSuffix(first, ".ico")
		if !allowed[first] && !versionIcon {
			return "", fmt.Errorf("archivo protegido: %s", name)
		}
	}
	target := root
	for _, part := range parts {
		target = filepath.Join(target, part)
		info, err := os.Lstat(target)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("no se admiten enlaces en %s", name)
		}
	}
	return target, nil
}

func FileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func New(root, version string) *Manifest {
	return &Manifest{Product: product, Root: root, Version: version, Files: map[string]string{}}
}

func Load(root string) (*Manifest, error) {
	f, err := os.Open(filepath.Join(root, ManifestName))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var m Manifest
	if err = json.NewDecoder(io.LimitReader(f, 2<<20)).Decode(&m); err != nil {
		return nil, err
	}
	if m.Product != product || m.Root != root || len(m.Files) == 0 {
		return nil, errors.New("no se reconoce esta instalación; vuelva a ejecutar el instalador en la misma carpeta")
	}
	for name, hash := range m.Files {
		if _, err = ProgramPath(root, name); err != nil {
			return nil, err
		}
		if b, e := hex.DecodeString(hash); e != nil || len(b) != sha256.Size {
			return nil, errors.New("registro de instalación inválido")
		}
	}
	return &m, nil
}

func (m *Manifest) Record(name string) error {
	path, err := ProgramPath(m.Root, name)
	if err != nil {
		return err
	}
	hash, err := FileHash(path)
	if err != nil {
		return err
	}
	m.Files[name] = hash
	return nil
}

func (m *Manifest) Save() error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(m.Root, ManifestName)
	tmp := path + ".installing"
	if err = os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	// Windows cannot atomically replace an existing destination via os.Rename.
	if err = os.Rename(tmp, path); err != nil {
		if e := os.Remove(path); e != nil && !os.IsNotExist(e) {
			return e
		}
		err = os.Rename(tmp, path)
	}
	return err
}

// RemoveProgram preserves changed files and every unrecorded file. The running
// uninstaller and its manifest are removed only after Windows registration is
// successfully cleared, so a failed uninstall can be retried.
func (m *Manifest) RemoveProgram() (preserved []string, err error) {
	names := make([]string, 0, len(m.Files))
	for name := range m.Files {
		if _, e := ProgramPath(m.Root, name); e != nil {
			return nil, e
		}
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if name == UninstallerName {
			continue
		}
		path, e := ProgramPath(m.Root, name)
		if e != nil {
			return preserved, e
		}
		hash, e := FileHash(path)
		if os.IsNotExist(e) {
			continue
		}
		if e != nil {
			err = errors.Join(err, e)
			continue
		}
		if hash != m.Files[name] {
			preserved = append(preserved, name)
			continue
		}
		if e = os.Remove(path); e != nil {
			err = errors.Join(err, e)
		}
	}
	return preserved, err
}

func (m *Manifest) Finish() error {
	if hash, ok := m.Files[UninstallerName]; ok {
		path, err := ProgramPath(m.Root, UninstallerName)
		if err != nil {
			return err
		}
		current, err := FileHash(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if current == hash {
			if err = os.Remove(path); err != nil {
				return err
			}
		}
	}
	if err := os.Remove(filepath.Join(m.Root, ManifestName)); err != nil {
		return err
	}
	// Remove only empty program directories, never their remaining contents.
	dirs := map[string]bool{}
	for name := range m.Files {
		for dir := filepath.Dir(filepath.FromSlash(name)); dir != "."; dir = filepath.Dir(dir) {
			dirs[dir] = true
		}
	}
	ordered := make([]string, 0, len(dirs))
	for dir := range dirs {
		ordered = append(ordered, dir)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ordered)))
	for _, dir := range ordered {
		_ = os.Remove(filepath.Join(m.Root, dir))
	}
	_ = os.Remove(m.Root) // succeeds only for a completely empty, unused installation
	return nil
}
