package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type accessRecord571 struct {
	Drained        bool   `json:"drained"`
	Disabled       bool   `json:"disabled"`
	Manual         bool   `json:"manual"`
	Reason         string `json:"reason,omitempty"`
	BlockedAt      string `json:"blockedAt,omitempty"`
	ExtensionUsed  bool   `json:"extensionUsed"`
	ExtensionUntil string `json:"extensionUntil,omitempty"`
	ReleasedAt     string `json:"releasedAt,omitempty"`
	ChangedBy      string `json:"changedBy,omitempty"`
}
type accessFile571 map[string]map[string]*accessRecord571

func (s *Server) loadAccess571() (accessFile571, error) {
	data := accessFile571{}
	b, e := os.ReadFile(filepath.Join(filepath.Dir(s.path), "access-controls-571.json"))
	if os.IsNotExist(e) {
		return data, nil
	}
	if e != nil {
		return nil, e
	}
	e = json.Unmarshal(b, &data)
	return data, e
}
func (s *Server) saveAccess571(data accessFile571) error {
	p := filepath.Join(filepath.Dir(s.path), "access-controls-571.json")
	b, e := json.MarshalIndent(data, "", "  ")
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(p), 0755); e != nil {
		return e
	}
	if e = os.WriteFile(p+".tmp", b, 0600); e != nil {
		return e
	}
	return os.Rename(p+".tmp", p)
}
func accessEntry571(data accessFile571, scope, id string) *accessRecord571 {
	if data[scope] == nil {
		data[scope] = map[string]*accessRecord571{}
	}
	if data[scope][id] == nil {
		data[scope][id] = &accessRecord571{}
	}
	return data[scope][id]
}
func accessScope571(r *http.Request) string {
	if musicRequest553(r) || programRequest556(r) == musicProgram553 {
		return "music"
	}
	return "conquis"
}
func accessInScope571(db map[string]any, u map[string]any, scope string) bool {
	if scope == "music" {
		return musicAssigned550(db, stringValue(u["id"])) || managementProgram555(db, u, musicProgram553)
	}
	for _, p := range []string{"class_grouped", "guide_major", "guide_master", "guide_master_advanced"} {
		if assignedProgram555(db, stringValue(u["id"]), p) || managementProgram555(db, u, p) {
			return true
		}
	}
	return false
}
func accessAdmin571(db map[string]any, u map[string]any, scope string) bool {
	return legacyActive(u) && (legacyRole(u) == "admin" || scope == "music" && musicManager550(legacyRole(u)) && managementProgram555(db, u, musicProgram553))
}
func accessTime571(v any) time.Time {
	t, _ := time.Parse(time.RFC3339, stringValue(v))
	if t.IsZero() {
		t, _ = time.Parse("2006-01-02", stringValue(v))
	}
	return t
}
func musicTutorReady571(db map[string]any, u map[string]any) bool {
	id := legacyTutorForProgram(u, musicProgram553)
	t := legacyUser(db, id)
	return id != "" && id != stringValue(u["id"]) && legacyActive(t) && managementProgram555(db, t, musicProgram553)
}

// A submission and its associated task represent one delivery, not two.
func autoReasons571(db map[string]any, u map[string]any, now time.Time, released string) []string {
	if legacyRole(u) == "admin" {
		return nil
	}
	since := now.AddDate(0, 0, -14)
	joined := accessTime571(u["createdAt"])
	if joined.After(since) || accessTime571(released).After(since) {
		return nil
	}
	reasons := []string{}
	for program, ids := range requirements571 {
		if !assignedProgram555(db, stringValue(u["id"]), program) {
			continue
		}
		p := legacyProgressForRequirement(db, stringValue(u["id"]), program, false)
		statuses := mapField(p, "status")
		subs := mapField(p, "submissions")
		presented := 0
		weeks := [2]map[string]bool{{}, {}}
		deliveries := map[string]time.Time{}
		add := func(key string, at time.Time) {
			if key != "" && at.After(deliveries[key]) {
				deliveries[key] = at
			}
		}

		for _, id := range ids {
			status := stringValue(statuses[id])
			sub := legacyMap(subs[id])
			if status == "in_review" || status == "approved" || status == "needs_changes" {
				presented++
				at := accessTime571(sub["submittedAt"])
				if at.IsZero() {
					at = accessTime571(sub["sentAt"])
				}
				if at.IsZero() {
					at = accessTime571(sub["reviewRequestedAt"])
				}
				if at.IsZero() {
					at = accessTime571(sub["updatedAt"])
				}
				add(id, at)
			}
		}
		if len(ids) == 0 || float64(presented)/float64(len(ids)) >= .9 {
			continue
		}
		total, done := 0, 0
		for _, raw := range arrayField(db, "tasks") {
			task := legacyMap(raw)
			pid := stringValue(task["programId"])
			if pid == "" {
				pid = "class_grouped"
			}
			if pid != program || stringValue(task["assigneeId"]) != stringValue(u["id"]) || task["active"] == false || stringValue(task["deletedAt"]) != "" {
				continue
			}
			id := stringValue(task["requirementId"])
			at := accessTime571(task["presentedAt"])
			add(id, at)
			due := accessTime571(task["dueDate"])
			if due.IsZero() || due.Before(since) || due.After(now) {
				continue
			}
			total++
			status := stringValue(statuses[id])
			if !at.IsZero() || status == "in_review" || status == "approved" {
				done++
			}
		}
		for key, at := range deliveries {
			if at.Before(since) || at.After(now) || at.IsZero() {
				continue
			}
			i := 0
			if !at.Before(now.AddDate(0, 0, -7)) {
				i = 1
			}
			weeks[i][key] = true
		}
		if len(weeks[0]) < 2 || len(weeks[1]) < 2 {
			reasons = append(reasons, program+": menos de dos entregas en alguna de las últimas dos semanas (las tareas presentadas cuentan).")
		}
		if total > 0 && float64(done)/float64(total) < .75 {
			reasons = append(reasons, program+": menos del 75 % de tareas con vencimiento en las últimas dos semanas cumplidas.")
		}
	}
	return reasons
}
func accessStatus571(db map[string]any, u map[string]any, scope string, record *accessRecord571, now time.Time) map[string]any {
	reasons := []string{}
	if record.Disabled {
		reasons = append(reasons, "Aspirante inhabilitado. Comunícate con el administrador.")
	} else if record.Manual {
		reasons = append(reasons, record.Reason)
		if record.Reason == "" {
			reasons = []string{"Bloqueo temporal dispuesto por la administración."}
		}
	}
	noTutor := scope == "music" && legacyRole(u) == "aspirant" && !musicTutorReady571(db, u)
	if noTutor {
		reasons = append(reasons, "Necesitas un tutor de Música asignado y activo para ingresar.")
	}
	extension := scope == "conquis" && record.Manual && !record.Disabled && !noTutor && now.Before(accessTime571(record.ExtensionUntil))
	blocked := len(reasons) > 0 && !extension
	if blocked && record.BlockedAt == "" {
		record.BlockedAt = now.UTC().Format(time.RFC3339Nano)
	}
	if !blocked && record.BlockedAt != "" {
		record.BlockedAt = ""
		record.Drained = false
	}
	return map[string]any{"scope": scope, "followup": followup572(db, u, scope, now), "blocked": blocked, "disabled": record.Disabled, "reasons": reasons, "blockedAt": record.BlockedAt, "canExtend": blocked && scope == "conquis" && record.Manual && !record.Disabled && !record.ExtensionUsed, "extensionUntil": record.ExtensionUntil, "needsTutor": noTutor, "trainingNeedsTutor": scope == "music" && !musicTutorReady571(db, u)}
}
func (s *Server) access571(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "method", 405)
		return
	}
	var req struct {
		ActorID string   `json:"actorId"`
		Action  string   `json:"action"`
		IDs     []string `json:"ids"`
		Reason  string   `json:"reason"`
		All     bool     `json:"all"`
	}
	req.ActorID = r.URL.Query().Get("actorId")
	if r.Method == "POST" {
		if readJSONLimited(r, &req, 1<<20) != nil {
			legacyError(w, 400, "Solicitud inválida", nil)
			return
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, e := s.loadLegacy()
	if e != nil {
		legacyError(w, 500, e.Error(), nil)
		return
	}
	id, _, ok := s.accountActor556(r, req.ActorID, &st)
	db := legacyMap(st.State.Database)
	u := legacyUser(db, id)
	if !ok || !legacyActive(u) {
		legacyError(w, 403, "No autorizado", nil)
		return
	}
	scope := accessScope571(r)
	data, e := s.loadAccess571()
	if e != nil {
		legacyError(w, 500, e.Error(), nil)
		return
	}
	now := time.Now().UTC()
	record := accessEntry571(data, scope, id)
	manager := accessAdmin571(db, u, scope) && !personalRequest555(r)
	if r.Method == "POST" {
		if req.Action == "extend" {
			status := accessStatus571(db, u, scope, record, now)
			if status["canExtend"] != true {
				legacyError(w, 403, "El plazo no está disponible. Comunícate con el administrador.", nil)
				return
			}
			record.ExtensionUsed = true
			record.ExtensionUntil = now.Add(72 * time.Hour).Format(time.RFC3339Nano)
			record.BlockedAt = ""
		} else if req.Action == "drained" { // UI confirms its queued changes reached the server.
			if record.BlockedAt != "" {
				record.Drained = true
			}
		} else {
			if !manager {
				legacyError(w, 403, "No autorizado para gestionar bloqueos", nil)
				return
			}
			if req.Action != "block" && req.Action != "release" && req.Action != "disable" && req.Action != "enable" {
				legacyError(w, 400, "Acción inválida", nil)
				return
			}
			ids := req.IDs
			if req.All {
				ids = nil
				for _, raw := range legacyUsers(db) {
					target := legacyMap(raw)
					if accessInScope571(db, target, scope) && legacyRole(target) != "admin" {
						ids = append(ids, stringValue(target["id"]))
					}
				}
			}
			for _, uid := range ids {
				target := legacyUser(db, uid)
				if target == nil || uid == id || legacyRole(target) == "admin" || !accessInScope571(db, target, scope) {
					legacyError(w, 403, "La selección contiene usuarios fuera del alcance o administradores protegidos.", nil)
					return
				}
			}
			for _, uid := range ids {
				rec := accessEntry571(data, scope, uid)
				rec.ChangedBy = id
				switch req.Action {
				case "block":
					rec.Manual = true
					if scope == "conquis" {
						rec.ExtensionUsed = false
						rec.ExtensionUntil = ""
					}
					rec.Drained = false
					rec.Reason = strings.TrimSpace(req.Reason)
					rec.BlockedAt = now.Format(time.RFC3339Nano)
				case "disable":
					rec.Disabled = true
					rec.BlockedAt = now.Format(time.RFC3339Nano)
				case "enable":
					rec.Disabled = false
					rec.Manual = false
					rec.BlockedAt = ""
					rec.ReleasedAt = now.Format(time.RFC3339Nano)
				case "release":
					rec.Manual = false
					rec.BlockedAt = ""
					rec.ExtensionUsed = false
					rec.ExtensionUntil = ""
					rec.ReleasedAt = now.Format(time.RFC3339Nano)
				}
			}
		}
	}
	status := accessStatus571(db, u, scope, record, now)
	rows := []any{}
	if manager && r.URL.Query().Get("manage") == "1" {
		for _, raw := range legacyUsers(db) {
			target := legacyMap(raw)
			if !accessInScope571(db, target, scope) {
				continue
			}
			uid := stringValue(target["id"])
			entry := accessStatus571(db, target, scope, accessEntry571(data, scope, uid), now)
			entry["id"] = uid
			entry["fullName"] = target["fullName"]
			entry["role"] = target["role"]
			rows = append(rows, entry)
		}
	}
	if e = s.saveAccess571(data); e != nil {
		legacyError(w, 500, e.Error(), nil)
		return
	}
	disabledIDs := []string{}
	for uid, entry := range data[scope] {
		if entry.Disabled {
			disabledIDs = append(disabledIDs, uid)
		}
	}
	jsonOut(w, 200, map[string]any{"ok": true, "disabledIds": disabledIDs, "status": status, "rows": rows, "manager": manager, "scope": scope})
}

// Enforce remote access on the server, including sessions opened before a block.
func (s *Server) accessGuard571(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if !strings.HasPrefix(path, "/api/") || path == "/api/access571" || path == "/api/remote/login" || path == "/api/remote/logout" || path == "/api/health" || path == "/api/remote/info" || path == "/api/remote/session" {
			next.ServeHTTP(w, r)
			return
		}
		sess, ok := s.remoteSessionFromRequest(r)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		s.mu.Lock()
		st, e := s.loadLegacy()
		if e != nil {
			s.mu.Unlock()
			legacyError(w, 503, "No se pudo comprobar el acceso", nil)
			return
		}
		db := legacyMap(st.State.Database)
		u := legacyUser(db, sess.UserID)
		scope := accessScope571(r)
		data, e := s.loadAccess571()
		if e != nil {
			s.mu.Unlock()
			legacyError(w, 503, "No se pudo comprobar el acceso", nil)
			return
		}
		before, _ := json.Marshal(data)
		rec := accessEntry571(data, scope, sess.UserID)
		status := accessStatus571(db, u, scope, rec, time.Now().UTC())
		if !legacyActive(u) {
			s.mu.Unlock()
			legacyError(w, 403, "Cuenta inactiva", nil)
			return
		}
		blocked := status["blocked"] == true
		// Existing sessions may finish saving for up to two minutes. New sessions never get this window.
		drain := blocked && !rec.Drained && !rec.Disabled && accessTime571(sess.CreatedAt).Before(accessTime571(rec.BlockedAt)) && time.Since(accessTime571(rec.BlockedAt)) < 2*time.Minute
		after, _ := json.Marshal(data)
		if !bytes.Equal(before, after) {
			e = s.saveAccess571(data)
		}
		s.mu.Unlock()
		if e != nil {
			legacyError(w, 503, "No se pudo registrar el estado de acceso", nil)
			return
		}
		if blocked && !drain {
			legacyError(w, 423, "BLOQUEO TEMPORAL", map[string]any{"access571": status})
			return
		}
		if scope == "music" && !musicTutorReady571(db, u) && (personalRequest555(r) || legacyRole(u) == "aspirant") && r.Method == "POST" && path != "/api/remote/heartbeat" {
			legacyError(w, 403, "Necesitas un tutor de Música para registrar tu avance personal.", map[string]any{"access571": status})
			return
		}
		next.ServeHTTP(w, r)
	})
}
