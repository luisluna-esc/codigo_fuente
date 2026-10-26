package main

import (
	"net/http"
	"strings"
	"time"
)

// Readers receive the complete progress scope, never credentials or Drive tokens.
func sanitizeReadAll548(db map[string]any, actorID string) {
	for _, raw := range legacyUsers(db) {
		user := legacyMap(raw)
		delete(user, "password")
		delete(user, "recoveryKeys")
		delete(user, "drive")
	}
	for _, key := range []string{"googleClientSecret", "googleClientId", "serverToken", "clientToken", "adminToken", "recoveryKeys", "adminRecovery"} {
		delete(legacyMap(db["settings"]), key)
	}
}

type indicationEdit548 struct {
	ID            string `json:"id"`
	ProgramID     string `json:"programId"`
	RequirementID string `json:"requirementId"`
	Instructions  string `json:"instructions"`
	BaseUpdatedAt string `json:"baseUpdatedAt"`
}

// This endpoint accepts only instruction text. It cannot change modes, scopes,
// assignments, account data or reports. Each rule has an optimistic version guard.
func (s *Server) indications548(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	if _, ok := s.auth(r); !ok {
		legacyError(w, 401, "Sesión no válida.", nil)
		return
	}
	var req struct {
		ActorID           string              `json:"actorId"`
		RestoreGeneration int64               `json:"restoreGeneration"`
		Edits             []indicationEdit548 `json:"edits"`
	}
	if err := readJSON(r, &req); err != nil || len(req.Edits) == 0 || len(req.Edits) > 500 {
		legacyError(w, 400, "Indicaciones inválidas.", nil)
		return
	}
	if sess, ok := s.remoteSessionFromRequest(r); ok && sess.UserID != req.ActorID {
		legacyError(w, 403, "La sesión no corresponde al usuario.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	actor := legacyUser(st.State.Database, req.ActorID)
	allowed := legacyRole(actor) == "departamental"
	allMusic := true
	for _, edit := range req.Edits {
		if edit.ProgramID != musicProgram553 {
			allMusic = false
		}
	}
	if allMusic {
		allowed = musicPolicies560[legacyRole(actor)].Instructions && !personalRequest555(r)
	}
	if !legacyActive(actor) || !allowed {
		legacyError(w, 403, "Solo Departamental puede editar indicaciones desde esta vista.", nil)
		return
	}
	generation := legacyCanonicalRestoreGeneration(st.State)
	if req.RestoreGeneration != generation && !(req.RestoreGeneration == 0 && generation == 1) {
		legacyError(w, 409, "Actualiza la copia después de la restauración.", map[string]any{"code": "RESTORE_REQUIRED"})
		return
	}
	db := legacyMap(cloneAny(st.State.Database))
	rules := arrayField(db, "indicationRules")
	now := time.Now().UTC().Format(time.RFC3339Nano)
	seen := map[string]bool{}
	for _, edit := range req.Edits {
		if !managementProgram555(legacyMap(st.State.Database), actor, edit.ProgramID) {
			legacyError(w, 403, "Este módulo no está habilitado para tu función.", nil)
			return
		}
		if !containsString([]any{"class_grouped", "guide_major", "guide_master", "guide_master_advanced", "music_ministry"}, edit.ProgramID) || strings.TrimSpace(edit.RequirementID) == "" || strings.TrimSpace(edit.ID) == "" || len(edit.Instructions) > 30000 || seen[edit.ID] {
			legacyError(w, 400, "Indicación inválida.", nil)
			return
		}
		seen[edit.ID] = true
		var target map[string]any
		for _, raw := range rules {
			rule := legacyMap(raw)
			if stringValue(rule["id"]) == edit.ID {
				target = rule
				break
			}
		}
		if target != nil {
			if stringValue(target["programId"]) != edit.ProgramID || stringValue(target["requirementId"]) != edit.RequirementID || stringValue(target["scope"]) != "all" {
				legacyError(w, 403, "Solo puedes modificar la indicación general.", nil)
				return
			}
			if stringValue(target["updatedAt"]) != edit.BaseUpdatedAt {
				if stringValue(legacyMap(target["config"])["instructions"]) == edit.Instructions {
					continue
				} // retry after a lost acknowledgement
				legacyError(w, 409, "Otra persona cambió esta indicación. Actualiza y revisa tu texto antes de guardar.", map[string]any{"code": "INDICATION_CONFLICT", "requirementId": edit.RequirementID})
				return
			}
		} else {
			if edit.BaseUpdatedAt != "" {
				legacyError(w, 409, "La indicación cambió o fue eliminada.", nil)
				return
			}
			// A concurrent general rule must not be duplicated under a different ID.
			for _, raw := range rules {
				rule := legacyMap(raw)
				if stringValue(rule["programId"]) == edit.ProgramID && stringValue(rule["requirementId"]) == edit.RequirementID && stringValue(rule["scope"]) == "all" && rule["active"] != false {
					legacyError(w, 409, "Ya existe una indicación general nueva. Actualiza la vista.", map[string]any{"code": "INDICATION_CONFLICT"})
					return
				}
			}
			target = map[string]any{"id": edit.ID, "programId": edit.ProgramID, "requirementId": edit.RequirementID, "scope": "all", "scopeId": "", "active": true, "createdAt": now, "createdBy": req.ActorID, "config": map[string]any{}}
			rules = append(rules, target)
		}
		ensureMapField(target, "config")["instructions"] = edit.Instructions
		target["updatedAt"] = now
		target["updatedBy"] = req.ActorID
	}
	db["indicationRules"] = rules
	st.State.ServerVersion++
	st.State.Database = legacyPrepareCanonicalDB(db, st.State, "Indicaciones actualizadas por Departamental")
	st.State.UpdatedAt = now
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "indicationRules": rules, "serverVersion": st.State.ServerVersion})
}

// A ministry registration belongs only to an account assigned to Music.
func validMusicRegistration548(db map[string]any, ownerID string, value any) bool {
	assigned := false
	for _, raw := range arrayField(db, "modules") {
		module := legacyMap(raw)
		if stringValue(module["id"]) == "music_ministry" && module["active"] != false && containsString(arrayField(module, "assignedUserIds"), ownerID) {
			assigned = true
		}
	}
	if !assigned {
		return false
	}
	reg := legacyMap(value)
	if reg == nil {
		return false
	}
	if !containsString([]any{"Coro", "Solista", "Conjunto", "Dúo", "Trío", "Conjunto Instrumental", "Conjunto Vocal", "Cuarteto", "Otro"}, stringValue(reg["type"])) {
		return false
	}
	if stringValue(reg["type"]) == "Otro" && strings.TrimSpace(stringValue(reg["otherType"])) == "" && stringValue(reg["updatedAt"]) == "" {
		return false
	}
	if len(stringValue(reg["otherType"])) > 500 {
		return false
	}
	for _, key := range []string{"name", "church", "district", "region", "mission", "union", "createdDate", "boardVote", "director", "directorPhone"} {
		val, ok := reg[key].(string)
		if !ok || strings.TrimSpace(val) == "" || len(val) > 500 {
			return false
		}
	}
	_, err := time.Parse("2006-01-02", stringValue(reg["createdDate"]))
	return err == nil
}
