package main

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"sgl.local/sistema-gestion-liderazgo/internal/installstate"
	"sgl.local/sistema-gestion-liderazgo/internal/processutil"
	"sgl.local/sistema-gestion-liderazgo/internal/release"
)

//go:embed payload.zip
var payload []byte

//go:embed obsolete624.txt
var obsolete624 string

func main() {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base, _ = os.UserHomeDir()
	}
	destination := flag.String("dest", filepath.Join(base, "MinisterioMusicaOnline"), "carpeta de instalación o actualización")
	noStart := flag.Bool("no-start", false, "instalar sin iniciar el servidor")
	flag.Parse()
	if err := install(*destination, *noStart); err != nil {
		fmt.Fprintln(os.Stderr, "No se completó la instalación:", err)
		if runtime.GOOS == "windows" {
			processutil.Message("Instalación de Música", err.Error())
		}
		os.Exit(1)
	}
}

func install(destination string, noStart bool) error {
	root, err := filepath.Abs(destination)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(root, 0755); err != nil {
		return err
	}
	root, err = installstate.CanonicalRoot(root)
	if err != nil {
		return err
	}
	manifest := installstate.New(root, release.Version)
	if previous, e := installstate.Load(root); e == nil {
		manifest.Files = previous.Files
	} else if !os.IsNotExist(e) {
		return e
	}
	_, installedErr := os.Stat(filepath.Join(root, "Musica.exe"))
	wasInstalled := installedErr == nil
	if err = processutil.ClearStaleLock(filepath.Join(root, "musica-server.lock")); err != nil {
		return fmt.Errorf("detén el servidor antes de actualizar: %w. Usa DETENER_MUSICA.cmd; en la versión anterior cierra la ventana ServidorMusica.exe. No se modificaron los datos", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return err
	}
	fmt.Printf("Ministerio de la Música Online\nInstalando en: %s\n", root)
	fmt.Println("Uso autorizado sólo para el entorno de la Misión Boliviana Occidental Sur.")
	for _, entry := range zr.File {
		name := filepath.Clean(filepath.FromSlash(entry.Name))
		target := filepath.Join(root, name)
		if name == ".." || filepath.IsAbs(name) || !strings.HasPrefix(target, root+string(filepath.Separator)) {
			return fmt.Errorf("ruta de paquete inválida")
		}
		// User databases, activation key and server configuration are never packaged.
		if strings.EqualFold(strings.Split(name, string(filepath.Separator))[0], "data") || strings.EqualFold(name, "server-config.json") {
			continue
		}
		if entry.FileInfo().IsDir() {
			continue
		}
		if _, err = installstate.ProgramPath(root, filepath.ToSlash(name)); err != nil {
			return err
		}
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		in, e := entry.Open()
		if e != nil {
			return e
		}
		tmp := target + ".installing"
		out, e := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, entry.Mode())
		if e != nil {
			in.Close()
			return e
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		in.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if runtime.GOOS == "windows" {
			_ = os.Remove(target)
		}
		if err = os.Rename(tmp, target); err != nil {
			return err
		}
		if err = manifest.Record(filepath.ToSlash(name)); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(root, "app", "music-runtime.js")); err == nil {
		// Remove only known superseded release files, after all new files were
		// installed. Never sweep directories or touch databases/configuration.
		for _, name := range strings.Fields(obsolete624) {
			clean := filepath.Clean(filepath.FromSlash(name))
			if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
				continue
			}
			_ = os.Remove(filepath.Join(root, clean))
		}
	}
	if wasInstalled {
		fmt.Println("Actualización completada. Se conservaron los datos y la configuración.")
	} else {
		fmt.Println("Instalación completada.")
	}
	if runtime.GOOS == "windows" {
		// A new icon filename avoids reusing Explorer's cached icon after updates.
		iconPath := filepath.Join(root, "Musica-"+strings.TrimSuffix(release.Version, "-musica-online")+".ico")
		icon, e := os.ReadFile(filepath.Join(root, "Musica.ico"))
		if e != nil {
			return e
		}
		if e = os.WriteFile(iconPath, icon, 0644); e != nil {
			return e
		}
		if err = manifest.Record(filepath.Base(iconPath)); err != nil {
			return err
		}
		if err = manifest.Save(); err != nil {
			return err
		}
		if err = installstate.Register(root, release.Version, iconPath); err != nil {
			return fmt.Errorf("el programa está instalado, pero no se pudieron registrar los accesos o el desinstalador: %w", err)
		}
	} else {
		if err = manifest.Save(); err != nil {
			return err
		}
	}
	if !noStart {
		if runtime.GOOS == "windows" {
			cmd := exec.Command(filepath.Join(root, "Musica.exe"))
			cmd.Dir = root
			processutil.Hidden(cmd)
			return cmd.Start()
		}
		fmt.Println("Para iniciar: ./ServidorMusica --config server-config.json")
	}
	return nil
}
