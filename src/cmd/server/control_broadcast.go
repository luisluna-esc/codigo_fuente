package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type controlState struct {
	Revision          int64  `json:"revision"`
	Kind              string `json:"kind,omitempty"`
	Message           string `json:"message,omitempty"`
	IssuedAt          string `json:"issuedAt,omitempty"`
	IssuedBy          string `json:"issuedBy,omitempty"`
	RestoreGeneration int64  `json:"restoreGeneration,omitempty"`
	RestoreID         string `json:"restoreId,omitempty"`
	ServerVersion     int64  `json:"serverVersion,omitempty"`
}

func (s *Server) controlPath() string {
	return filepath.Join(filepath.Dir(s.path), "sgl-control-state.json")
}

func (s *Server) loadControlState() {
	b, err := os.ReadFile(s.controlPath())
	if err != nil {
		return
	}
	var st controlState
	if json.Unmarshal(b, &st) == nil && st.Revision > 0 {
		if st.Kind == "system-restored" && st.RestoreGeneration > 0 {
			st.Kind = "force-relogin"
			s.control = st
			s.saveControlStateLocked()
			return
		}
		s.control = st
	}
}

func (s *Server) saveControlStateLocked() {
	b, _ := json.MarshalIndent(s.control, "", "  ")
	_ = os.MkdirAll(filepath.Dir(s.controlPath()), 0755)
	tmp := s.controlPath() + ".tmp"
	if os.WriteFile(tmp, b, 0600) == nil {
		if runtimeIsWindows() {
			_ = os.Remove(s.controlPath())
		}
		_ = os.Rename(tmp, s.controlPath())
	}
}

func (s *Server) currentControlState() controlState {
	s.remoteMu.Lock()
	defer s.remoteMu.Unlock()
	return s.control
}

func (s *Server) adminControlState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	role, ok := s.auth(r)
	if !ok || role != "admin" {
		legacyError(w, http.StatusForbidden, "Esta acción requiere Administrador.", nil)
		return
	}
	jsonOut(w, http.StatusOK, map[string]any{"ok": true, "control": s.currentControlState()})
}

func (s *Server) adminForceRelogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	role, ok := s.auth(r)
	if !ok || role != "admin" {
		legacyError(w, http.StatusForbidden, "Esta acción requiere Administrador.", nil)
		return
	}
	var req struct {
		Message string `json:"message"`
	}
	if r.Body != nil {
		defer r.Body.Close()
		if r.ContentLength != 0 {
			decErr := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req)
			if decErr != nil && decErr != io.EOF {
				legacyError(w, http.StatusBadRequest, "Solicitud inválida.", nil)
				return
			}
		}
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		message = "El Administrador actualizó el sistema. Vuelve a iniciar sesión para cargar la versión más reciente."
	}
	if len(message) > 260 {
		message = message[:260]
	}
	now := time.Now().UTC()
	s.remoteMu.Lock()
	next := s.control.Revision + 1
	if next <= 0 {
		next = now.UnixNano()
	}
	s.control = controlState{Revision: next, Kind: "force-relogin", Message: message, IssuedAt: now.Format(time.RFC3339Nano), IssuedBy: "admin"}
	s.saveControlStateLocked()
	control := s.control
	s.remoteMu.Unlock()
	jsonOut(w, http.StatusOK, map[string]any{"ok": true, "control": control})
}
