package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"sgl.local/sistema-gestion-liderazgo/internal/processutil"
)

type config struct {
	Port       int    `json:"port"`
	AdminToken string `json:"adminToken"`
}

var client = &http.Client{Timeout: 2 * time.Second}

func available(url string) bool {
	r, err := client.Get(url + "/api/installation/status")
	if err != nil {
		return false
	}
	defer r.Body.Close()
	var status struct {
		Version string `json:"version"`
	}
	return r.StatusCode == 200 && json.NewDecoder(r.Body).Decode(&status) == nil && strings.Contains(status.Version, "musica-online")
}

func browser(root, url string) error {
	candidates := []string{}
	for _, base := range []string{os.Getenv("LOCALAPPDATA"), os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)")} {
		if base != "" {
			candidates = append(candidates, filepath.Join(base, "Google", "Chrome", "Application", "chrome.exe"))
		}
	}
	for _, base := range []string{os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramFiles"), os.Getenv("LOCALAPPDATA")} {
		if base != "" {
			candidates = append(candidates, filepath.Join(base, "Microsoft", "Edge", "Application", "msedge.exe"))
		}
	}
	if runtime.GOOS != "windows" {
		for _, name := range []string{"google-chrome", "chromium", "microsoft-edge"} {
			if path, err := exec.LookPath(name); err == nil {
				candidates = append(candidates, path)
			}
		}
	}
	for _, path := range candidates {
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			continue
		}
		cmd := exec.Command(path, "--app="+url, "--user-data-dir="+filepath.Join(root, "perfil-musica"), "--no-first-run", "--disable-default-apps")
		processutil.Hidden(cmd)
		if err := cmd.Start(); err != nil {
			return err
		}
		return cmd.Process.Release()
	}
	return errors.New("No se encontró Chrome ni Microsoft Edge. Instala uno de estos navegadores para abrir la ventana de Música.")
}

func run(root string, stop bool) error {
	cfg := config{Port: 8796}
	configPath := filepath.Join(root, "server-config.json")
	if b, err := os.ReadFile(configPath); err == nil {
		if err = json.Unmarshal(b, &cfg); err != nil {
			return fmt.Errorf("Revisa server-config.json: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return errors.New("El puerto del servidor no es válido.")
	}
	url := "http://127.0.0.1:" + strconv.Itoa(cfg.Port)
	lock := filepath.Join(root, "musica-server.lock")
	if stop {
		if !available(url) {
			return processutil.ClearStaleLock(lock)
		}
		if cfg.AdminToken == "" {
			return errors.New("No se encontró la configuración de esta instalación.")
		}
		r, _ := http.NewRequest(http.MethodPost, url+"/api/server/stop", nil)
		r.Header.Set("Authorization", "Bearer "+cfg.AdminToken)
		response, err := client.Do(r)
		if err != nil {
			return err
		}
		response.Body.Close()
		if response.StatusCode != 200 {
			return errors.New("No se pudo detener esta versión. Cierra la consola de la versión anterior antes de actualizar.")
		}
		for i := 0; i < 60; i++ {
			if !available(url) {
				if _, err := os.Stat(lock); os.IsNotExist(err) {
					return nil
				}
			}
			time.Sleep(250 * time.Millisecond)
		}
		return errors.New("El servidor aún está terminando. Espera unos segundos antes de actualizar.")
	}
	if !available(url) {
		if err := processutil.ClearStaleLock(lock); err != nil {
			return err
		}
		exe := filepath.Join(root, "ServidorMusica.exe")
		if runtime.GOOS != "windows" {
			exe = filepath.Join(root, "ServidorMusica")
		}
		cmd := exec.Command(exe, "--config", configPath)
		cmd.Dir = root
		processutil.Hidden(cmd)
		log, err := os.OpenFile(filepath.Join(root, "ArranqueMusica.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		cmd.Stdout = log
		cmd.Stderr = log
		if err = cmd.Start(); err != nil {
			log.Close()
			return err
		}
		_ = cmd.Process.Release()
		log.Close()
		for i := 0; i < 100; i++ {
			if available(url) {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}
		if !available(url) {
			return errors.New("El servidor no respondió. Revisa ServidorMusica.log y ArranqueMusica.log en la carpeta de Música.")
		}
	}
	return browser(root, url)
}

func main() {
	stop := flag.Bool("stop", false, "detener el servidor de esta instalación")
	flag.Parse()
	exe, err := os.Executable()
	if err == nil {
		err = run(filepath.Dir(exe), *stop)
	}
	if err != nil {
		processutil.Message("Ministerio de la Música", err.Error())
		os.Exit(1)
	}
	if *stop {
		processutil.Message("Ministerio de la Música", "Servidor detenido. Ya puedes actualizar el programa.")
	}
}
