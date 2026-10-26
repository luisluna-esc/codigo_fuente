package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type adminRestoreRequest struct {
	Database any    `json:"database"`
	Files    []any  `json:"files"`
	Message  string `json:"message,omitempty"`
}

func restoredDatabaseHasActiveAdmin(db any) bool {
	for _, raw := range legacyUsers(db) {
		user := legacyMap(raw)
		if legacyActive(user) && legacyRole(user) == "admin" {
			return true
		}
	}
	return false
}

func (s *Server) savePreRestoreSafety(st legacyStore, now time.Time) error {
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.legacyPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	name := "sgl-server-compat.pre-restore-" + now.UTC().Format("20060102T150405.000000000Z") + ".json"
	return os.WriteFile(filepath.Join(dir, name), b, 0600)
}

func (s *Server) publishSystemRestoredControl(message string, state legacyState, now time.Time) controlState {
	message = strings.TrimSpace(message)
	if message == "" {
		message = "El sistema fue restaurado por el Administrador. Se cargará la copia vigente del servidor y se cerrará tu sesión."
	}
	if len(message) > 260 {
		message = message[:260]
	}
	s.remoteMu.Lock()
	defer s.remoteMu.Unlock()
	next := s.control.Revision + 1
	if next <= 0 {
		next = now.UnixNano()
	}
	s.control = controlState{
		Revision:          next,
		Kind:              "force-relogin",
		Message:           message,
		IssuedAt:          now.UTC().Format(time.RFC3339Nano),
		IssuedBy:          "admin",
		RestoreGeneration: state.RestoreGeneration,
		RestoreID:         state.RestoreID,
		ServerVersion:     state.ServerVersion,
	}
	s.saveControlStateLocked()
	return s.control
}

func (s *Server) adminRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	role, ok := s.auth(r)
	if !ok || role != "admin" {
		legacyError(w, http.StatusForbidden, "Esta acción requiere Administrador.", nil)
		return
	}
	var req adminRestoreRequest
	if err := readJSONLimited(r, &req, 256<<20); err != nil || req.Database == nil {
		legacyError(w, http.StatusBadRequest, "La copia a restaurar no contiene una base válida.", nil)
		return
	}
	if !restoredDatabaseHasActiveAdmin(req.Database) {
		legacyError(w, http.StatusBadRequest, "La copia restaurada debe contener al menos un Administrador activo.", nil)
		return
	}
	if req.Files == nil {
		req.Files = []any{}
	}

	now := time.Now().UTC()
	s.mu.Lock()
	st, err := s.loadLegacy()
	if err != nil {
		s.mu.Unlock()
		legacyError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if err := s.savePreRestoreSafety(st, now); err != nil {
		s.mu.Unlock()
		legacyError(w, http.StatusInternalServerError, "No se pudo crear el punto de seguridad previo a la restauración.", nil)
		return
	}
	currentGeneration := st.State.RestoreGeneration
	if currentGeneration < 1 {
		currentGeneration = 1
	}
	nextVersion := st.State.ServerVersion + 1
	incomingDB := preserveLegacyCredentials(st.State.Database, req.Database)
	st.State.ServerVersion = nextVersion
	st.State.RestoreGeneration = currentGeneration + 1
	st.State.RestoreID = "restore_" + token(16)
	st.State.RestoredAt = now.Format(time.RFC3339Nano)
	st.State.Database = legacyPrepareCanonicalDB(incomingDB, st.State, "Sistema restaurado por el Administrador")
	st.State.Files = cloneAny(req.Files).([]any)
	st.State.UpdatedAt = now.Format(time.RFC3339Nano)
	st.Proposals = []legacyProposal{}
	st.UserChanges = []legacyUserChangeRequest{}
	st.AppliedOperations = []legacyAppliedOperation{}
	if err := s.saveLegacy(st); err != nil {
		s.mu.Unlock()
		legacyError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	restoredState := st.State
	snapshot := legacySnapshot(restoredState)
	s.mu.Unlock()

	control := s.publishSystemRestoredControl(req.Message, restoredState, now)
	snapshot["control"] = control
	snapshot["restored"] = true
	jsonOut(w, http.StatusOK, snapshot)
}
