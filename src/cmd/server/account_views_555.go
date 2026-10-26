package main

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type personalContext555 struct{}

func approvalAt555(sub map[string]any) time.Time {
	approval := legacyMap(sub["approval"])
	for _, value := range []any{approval["at"], approval["approvedAt"], sub["approvedAt"], sub["reviewedAt"], sub["updatedAt"]} {
		if at, err := time.Parse(time.RFC3339, stringValue(value)); err == nil {
			return at
		}
	}
	return time.Time{}
}

func personalRequest555(r *http.Request) bool { return r.Context().Value(personalContext555{}) == true }
func requestRole555(r *http.Request, actor map[string]any) string {
	if personalRequest555(r) {
		return "aspirant"
	}
	return legacyRole(actor)
}
func assignedProgram555(db map[string]any, id, program string) bool {
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if stringValue(m["id"]) == program && m["active"] != false && containsString(arrayField(m, "assignedUserIds"), id) {
			return true
		}
	}
	return false
}
func leader555(db map[string]any, u map[string]any) bool {
	if u == nil || !legacyActive(u) {
		return false
	}
	if legacyRole(u) == "aspirant" {
		return true
	}
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if m["active"] != false && containsString(arrayField(m, "assignedUserIds"), stringValue(u["id"])) {
			return true
		}
	}
	return false
}
func managementProgram555(db map[string]any, actor map[string]any, program string) bool {
	if !legacyActive(actor) {
		return false
	}
	role := legacyRole(actor)
	if raw := stringValue(actor["role"]); (raw == "administrativo_sgm" || raw == "lider_musica") && program != musicProgram553 {
		return false
	}
	if program == musicProgram553 && actor["musicManagementRemoved597"] == true {
		return false
	}
	if role == "departamental" && program != musicProgram553 {
		return false
	}
	if program == musicProgram553 && actor["musicAccessRemoved"] == true {
		return false
	}
	if role == "admin" {
		return true
	}
	if role == "tutor" {
		return containsString(tutorPrograms548(db, actor), program)
	}
	if role == "regional" {
		if v, ok := actor["managementProgramIds"]; ok {
			return containsString(anySlice548(v), program)
		}
		return true
	}
	if viewAll548(role) {
		if v, ok := actor["managementProgramIds"]; ok {
			return containsString(anySlice548(v), program)
		}
		return role != "coordinador_musica" || program == musicProgram553
	}
	return false
}
func regionalAssigned555(actor, owner map[string]any, program string) bool {
	id := stringValue(owner["id"])
	if containsString(arrayField(actor, "regionalAssignedUserIds"), id) {
		return true
	}
	aid := stringValue(actor["id"])
	if stringValue(owner["regionalId"]) == aid {
		return true
	}
	return stringValue(legacyMap(owner["regionalByProgram"])[program]) == aid || (program == musicProgram553 && legacyTutorAssigned(owner, aid, program))
}
func canReadOwner555(db map[string]any, actor, owner map[string]any, program string) bool {
	if program == musicProgram553 && len(arrayField(db, "modules")) > 0 {
		return musicPermit602(db, actor, "reports.view") && musicScope550(db, actor, owner)
	}
	if !legacyActive(actor) || !leader555(db, owner) {
		return false
	}
	if legacyRole(actor) == "admin" {
		return true
	}
	if len(arrayField(db, "modules")) > 0 && !assignedProgram555(db, stringValue(owner["id"]), program) {
		return false
	}
	if stringValue(actor["id"]) == stringValue(owner["id"]) {
		return true
	}
	if legacyRole(actor) == "tutor" && len(arrayField(db, "modules")) == 0 {
		return legacyTutorAssigned(owner, stringValue(actor["id"]), program)
	}
	if !managementProgram555(db, actor, program) {
		return false
	}
	switch legacyRole(actor) {
	case "admin", "departamental", "secretaria", "administrativo", "coordinador_musica":
		return true
	case "tutor":
		return legacyTutorAssigned(owner, stringValue(actor["id"]), program)
	case "regional":
		return regionalAssigned555(actor, owner, program)
	}
	return false
}
func canReadFile555(db map[string]any, actor map[string]any, file map[string]any) bool {
	if legacyRole(actor) == "admin" {
		return true
	}
	pid := remoteFileProgram(file)
	for _, raw := range arrayField(db, "materials") {
		m := legacyMap(raw)
		if stringValue(m["fileId"]) == stringValue(file["id"]) {
			return managementProgram555(db, actor, remoteFileProgram(m)) || assignedProgram555(db, stringValue(actor["id"]), remoteFileProgram(m)) || len(arrayField(db, "modules")) == 0
		}
	}
	for _, raw := range arrayField(db, "bibleVersions") {
		if stringValue(legacyMap(raw)["fileId"]) == stringValue(file["id"]) {
			return true
		}
	}
	return canReadOwner555(db, actor, legacyUser(db, stringValue(file["ownerId"])), pid)
}
func scopeProject555(db map[string]any, actorID, role string) map[string]any {
	if role == "admin" || role == "aspirant" {
		return db
	}
	actor := legacyUser(db, actorID)
	if actor == nil {
		return db
	}
	if role != "regional" {
		if _, explicit := actor["managementProgramIds"]; !explicit {
			return db
		}
	}
	programs := map[string]bool{}
	owners := map[string]bool{actorID: true}
	allowed := map[string]map[string]bool{}
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		p := stringValue(m["id"])
		if m["active"] == false {
			continue
		}
		if managementProgram555(db, actor, p) || assignedProgram555(db, actorID, p) {
			programs[p] = true
		}
		for _, uid := range arrayField(m, "assignedUserIds") {
			id := stringValue(uid)
			if canReadOwner555(db, actor, legacyUser(db, id), p) {
				if allowed[id] == nil {
					allowed[id] = map[string]bool{}
				}
				allowed[id][p] = true
				owners[id] = true
			}
		}
	}
	mods := []any{}
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		p := stringValue(m["id"])
		if !programs[p] {
			continue
		}
		ids := []any{}
		for _, id := range arrayField(m, "assignedUserIds") {
			if allowed[stringValue(id)][p] {
				ids = append(ids, id)
			}
		}
		m["assignedUserIds"] = ids
		mods = append(mods, m)
	}
	db["modules"] = mods
	progress := map[string]any{}
	for id, v := range legacyMap(db["progress"]) {
		if allowed[id]["class_grouped"] {
			progress[id] = v
		}
	}
	db["progress"] = progress
	pp := map[string]any{}
	for id, v := range legacyMap(db["programProgress"]) {
		out := map[string]any{}
		for p, x := range legacyMap(v) {
			if allowed[id][p] {
				out[p] = x
			}
		}
		if len(out) > 0 {
			pp[id] = out
		}
	}
	db["programProgress"] = pp
	by := map[string]any{}
	for id, v := range legacyMap(db["bibleYear"]) {
		out := map[string]any{}
		for p, x := range legacyMap(v) {
			if allowed[id][p] {
				out[p] = x
			}
		}
		if len(out) > 0 {
			by[id] = out
		}
	}
	db["bibleYear"] = by
	users := []any{}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		id := stringValue(u["id"])
		keep := owners[id] || legacyRole(u) == "admin"
		if legacyRole(u) == "tutor" {
			for p := range programs {
				if containsString(tutorPrograms548(db, u), p) {
					keep = true
				}
			}
		}
		if keep {
			users = append(users, sanitizedUser(u, id == actorID))
		}
	}
	db["users"] = users
	for _, key := range []string{"materials", "tasks", "indicationRules"} {
		out := []any{}
		for _, raw := range arrayField(db, key) {
			m := legacyMap(raw)
			p := remoteFileProgram(m)
			if !programs[p] {
				continue
			}
			if key == "tasks" && !allowed[stringValue(m["assigneeId"])][p] {
				continue
			}
			out = append(out, raw)
		}
		db[key] = out
	}
	for _, key := range []string{"configs", "reportRequirementModes"} {
		out := map[string]any{}
		for k, v := range legacyMap(db[key]) {
			for p := range programs {
				if strings.HasPrefix(k, p+":") {
					out[k] = v
					break
				}
			}
		}
		db[key] = out
	}
	if settings := legacyMap(db["settings"]); settings != nil {
		settings["reportRequirementModes"] = cloneAny(db["reportRequirementModes"])
	}
	db["audit"] = []any{}
	db["errors"] = []any{}
	db["music550"] = musicView550(db, actor)
	return db
}

// The selector changes this request's capabilities; the stored account and login never change.
func (s *Server) personalAccess555(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-SGL-Personal") != "1" {
			next.ServeHTTP(w, r)
			return
		}
		session, remote := s.remoteSessionFromRequest(r)
		actorID := session.UserID
		if !remote {
			if _, ok := s.auth(r); !ok {
				legacyError(w, 401, "Inicia sesión.", nil)
				return
			}
			actorID = r.Header.Get("X-SGL-Actor")
		}
		s.mu.Lock()
		st, err := s.loadLegacyLoginState()
		s.mu.Unlock()
		db := legacyMap(st.Database)
		actor := legacyUser(db, actorID)
		if err != nil || actorID == "" || !leader555(db, actor) {
			legacyError(w, 403, "El Administrador debe asignar tu tarjeta personal.", nil)
			return
		}
		path := r.URL.Path
		if legacyRole(actor) == "admin" && (strings.HasPrefix(path, musicAPI553) || programRequest556(r) == musicProgram553) {
			legacyError(w, 403, "El Administrador gestiona Música; no cursa una tarjeta personal de Música.", nil)
			return
		}
		if strings.HasPrefix(path, musicAPI553) {
			path = strings.TrimPrefix(path, musicAPI553)
		}
		blocked := strings.HasPrefix(path, "/api/admin/") && path != "/api/admin/control" || path == "/api/reviews/decision" || path == "/api/indications" || path == "/api/report-config" && r.Method != http.MethodGet || path == "/api/report-watermark" || path == "/api/music-registration" || strings.Contains(path, "/restore") || strings.HasPrefix(path, "/api/recovery/")
		if blocked {
			legacyError(w, 403, "Esta operación pertenece a Gestión. Cambia de vista.", nil)
			return
		}
		request := r.Clone(context.WithValue(r.Context(), personalContext555{}, true))
		next.ServeHTTP(w, request)
	})
}
func mergePersonal555(serverDB any, proposal *legacyProposal) (any, map[string]bool, error) {
	db := legacyMap(cloneAny(serverDB))
	source := legacyMap(cloneAny(proposal.Database))
	id := proposal.ActorID
	for pid := range legacyMap(legacyMap(source["programProgress"])[id]) {
		if !assignedProgram555(db, id, pid) {
			delete(legacyMap(legacyMap(source["programProgress"])[id]), pid)
		}
	}
	if !assignedProgram555(db, id, "class_grouped") {
		delete(legacyMap(source["progress"]), id)
	}
	copyProposal := *proposal
	copyProposal.Database = source
	proposal = &copyProposal
	actual := legacyUser(db, proposal.ActorID)
	role := actual["role"]
	actual["role"] = "aspirant"
	merged, allowed, err := mergeLegacyProposalOnServer(db, proposal)
	if err != nil {
		return nil, nil, err
	}
	legacyUser(merged, proposal.ActorID)["role"] = role
	return merged, allowed, nil
}
func (s *Server) driveAttempt555(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	if _, ok := s.auth(r); !ok {
		legacyError(w, 401, "Inicia sesión.", nil)
		return
	}
	var req struct {
		ActorID         string `json:"actorId"`
		OwnerID         string `json:"ownerId"`
		ProgramID       string `json:"programId"`
		RequirementID   string `json:"requirementId"`
		Error           string `json:"error"`
		SourceUpdatedAt string `json:"sourceUpdatedAt"`
	}
	if readJSON(r, &req) != nil || req.Error == "" {
		legacyError(w, 400, "Intento de Drive inválido.", nil)
		return
	}
	if sess, ok := s.remoteSessionFromRequest(r); ok {
		if req.ActorID != sess.UserID {
			legacyError(w, 403, "Cuenta incorrecta.", nil)
			return
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	db := legacyMap(st.State.Database)
	actor := legacyUser(db, req.ActorID)
	owner := legacyUser(db, req.OwnerID)
	if !canReadOwner555(db, actor, owner, req.ProgramID) || personalRequest555(r) && req.ActorID != req.OwnerID {
		legacyError(w, 403, "Informe fuera de tu alcance.", nil)
		return
	}
	p := legacyProgressForRequirement(db, req.OwnerID, req.ProgramID, false)
	sub := legacyMap(mapField(p, "submissions")[req.RequirementID])
	if sub == nil || stringValue(mapField(p, "status")[req.RequirementID]) != "approved" {
		legacyError(w, 409, "Solo se envían informes revisados.", nil)
		return
	}
	if req.ActorID != req.OwnerID {
		at := approvalAt555(sub)
		if at.IsZero() || time.Since(at) < 6*24*time.Hour {
			legacyError(w, 403, "Envío delegado aún no habilitado.", nil)
			return
		}
	}
	if req.SourceUpdatedAt != "" && req.SourceUpdatedAt != stringValue(sub["updatedAt"]) {
		legacyError(w, 409, "El informe cambió. Actualiza la vista.", nil)
		return
	}
	d := ensureMapField(sub, "drive")
	if stringValue(d["status"]) == "sent" || stringValue(d["status"]) == "verified" {
		legacyError(w, 409, "El archivo ya fue enviado. Comprueba Drive.", nil)
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	d["status"] = "error"
	d["error"] = req.Error
	d["errorAt"] = now
	d["errorBy"] = req.ActorID
	sub["updatedAt"] = now
	p["updatedAt"] = now
	st.State.ServerVersion++
	st.State.UpdatedAt = now
	if err = s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "drive": d, "updatedAt": now, "serverVersion": st.State.ServerVersion})
}
