package installstate

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"sgl.local/sistema-gestion-liderazgo/internal/processutil"
)

// StopServer uses this installation's private token and loopback address. It
// never kills a process, follows a redirect, or sends credentials through a proxy.
func StopServer(root string) error {
	lock := filepath.Join(root, "musica-server.lock")
	if processutil.ClearStaleLock(lock) == nil {
		return nil
	}
	var cfg struct {
		Port       int    `json:"port"`
		AdminToken string `json:"adminToken"`
	}
	b, err := os.ReadFile(filepath.Join(root, "server-config.json"))
	if err != nil {
		return fmt.Errorf("detenga el servidor con DETENER_MUSICA.cmd: %w", err)
	}
	if err = json.Unmarshal(b, &cfg); err != nil {
		return err
	}
	if cfg.Port < 1 || cfg.Port > 65535 || cfg.AdminToken == "" {
		return errors.New("no se pudo identificar el servidor; deténgalo con DETENER_MUSICA.cmd y reintente")
	}
	client := &http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	defer client.CloseIdleConnections()
	r, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:"+strconv.Itoa(cfg.Port)+"/api/server/stop", nil)
	r.Header.Set("Authorization", "Bearer "+cfg.AdminToken)
	response, err := client.Do(r)
	if err != nil {
		return errors.New("no se pudo detener Música; cierre el servidor con DETENER_MUSICA.cmd y reintente")
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return errors.New("el servidor no autorizó el cierre; no se retiraron archivos del programa")
	}
	for deadline := time.Now().Add(30 * time.Second); time.Now().Before(deadline); {
		if processutil.ClearStaleLock(lock) == nil {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return errors.New("el servidor sigue cerrándose; espere y vuelva a ejecutar el desinstalador")
}
