package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func programRequest556(r *http.Request) string {
	if musicRequest553(r) {
		return musicProgram553
	}
	if p := strings.TrimSpace(r.Header.Get("X-SGL-Program")); p != "" {
		return p
	}
	return strings.TrimSpace(r.URL.Query().Get("programId"))
}

func reviewRole556(role string) bool {
	return role == "admin" || role == "tutor" || role == "regional" || role == "departamental" || role == "secretaria" || role == "administrativo" || role == "coordinador_musica"
}
func canReview556(db map[string]any, actor, owner map[string]any, program string) bool {
	if program == musicProgram553 && len(arrayField(db, "modules")) > 0 {
		return musicReview560(db, actor, owner)
	}
	if actor == nil || owner == nil || stringValue(actor["id"]) == stringValue(owner["id"]) || !canReadOwner555(db, actor, owner, program) {
		return false
	}
	role := legacyRole(actor)
	if role == "admin" {
		return true
	}
	if role == "tutor" {
		return legacyTutorAssigned(owner, stringValue(actor["id"]), program)
	}
	return program == musicProgram553 && reviewRole556(role) && managementProgram555(db, actor, program)
}

// A small hash detects assignments and canonical changes in any module. The
// scoped snapshot contains metadata; evidence bytes stay in the file store.
func (s *Server) accountRoster556(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", 405)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	id, role, ok := s.accountActor556(r, r.URL.Query().Get("actorId"), &st)
	if !ok {
		legacyError(w, 403, "Sesión no autorizada.", nil)
		return
	}
	state := st.State
	state.Database = cloneAny(state.Database)
	snap := s.filteredLegacySnapshotMode(state, id, role, false)
	if musicRequest553(r) {
		snap["database"] = musicProject553(snap["database"], id, role)
	}
	db := legacyMap(snap["database"])
	people := []any{}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		row := map[string]any{}
		for _, k := range []string{"id", "fullName", "role", "active", "accountState", "region", "tutorByProgram", "tutorId", "tutorProgramIds", "regionalAssignedUserIds", "managementProgramIds"} {
			row[k] = u[k]
		}
		people = append(people, row)
	}
	b, _ := json.Marshal(map[string]any{"users": people, "modules": db["modules"], "workflow": mapField(legacyMap(db["music550"]), "workflow"), "generation": legacyCanonicalRestoreGeneration(st.State), "serverVersion": st.State.ServerVersion})
	h := sha256.Sum256(b)
	fingerprint := hex.EncodeToString(h[:])
	if r.URL.Query().Get("knownHash") == fingerprint {
		jsonOut(w, 200, map[string]any{"ok": true, "unchanged": true, "rosterHash": fingerprint})
		return
	}
	if _, remote := s.remoteSessionFromRequest(r); remote {
		stripRemoteCredentialFields508(snap)
	}
	snap["ok"] = true
	snap["rosterHash"] = fingerprint
	jsonOut(w, 200, snap)
}

func staffTutorCreator556(role string) bool {
	return role == "departamental" || role == "administrativo" || role == "secretaria"
}
func (s *Server) createTutor556(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	var req struct {
		ActorID    string         `json:"actorId"`
		ID         string         `json:"id"`
		ProgramID  string         `json:"programId"`
		Programs   []string       `json:"programIds"`
		Generation int64          `json:"restoreGeneration"`
		User       map[string]any `json:"user"`
	}
	if err := readJSONLimited(r, &req, 1<<20); err != nil {
		legacyError(w, 400, "Datos inválidos.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	id, _, ok := s.accountActor556(r, req.ActorID, &st)
	db := legacyMap(cloneAny(st.State.Database))
	actor := legacyUser(db, id)
	music := req.ProgramID == musicProgram553 || musicRequest553(r)
	creator := staffTutorCreator556(legacyRole(actor))
	if music {
		creator = musicPolicies560[legacyRole(actor)].CreateTutor
	}
	if !ok || personalRequest555(r) || !creator {
		legacyError(w, 403, "Tu función no permite crear tutores desde esta vista.", nil)
		return
	}
	if req.Generation != legacyCanonicalRestoreGeneration(st.State) {
		legacyError(w, 409, "Actualiza los datos antes de crear el tutor.", nil)
		return
	}
	if len(req.Programs) == 0 || len(req.Programs) > 5 || strings.TrimSpace(req.ID) == "" {
		legacyError(w, 400, "Selecciona los módulos del tutor.", nil)
		return
	}
	programs := []any{}
	seen := map[string]bool{}
	for _, p := range req.Programs {
		active := false
		for _, raw := range arrayField(db, "modules") {
			m := legacyMap(raw)
			if stringValue(m["id"]) == p && m["active"] != false {
				active = true
			}
		}
		if !active || !managementProgram555(db, actor, p) || (musicRequest553(r) && p != musicProgram553) {
			legacyError(w, 403, "No puedes asignar este módulo al tutor.", nil)
			return
		}
		if !seen[p] {
			programs = append(programs, p)
			seen[p] = true
		}
	}
	v := req.User
	name := strings.TrimSpace(stringValue(v["fullName"]))
	email := strings.ToLower(strings.TrimSpace(stringValue(v["email"])))
	username := strings.ToLower(strings.TrimSpace(stringValue(v["username"])))
	if name == "" || len(name) > 200 || !strings.Contains(email, "@") || len(email) > 250 || username == "" || len(username) > 100 {
		legacyError(w, 400, "Completa nombre, correo y usuario válidos.", nil)
		return
	}
	region := strings.TrimSpace(stringValue(v["region"]))
	if music {
		region = musicRegion560(region)
		if region == "" {
			legacyError(w, 400, "Selecciona una región: 1 a 12 o MBOS.", nil)
			return
		}
		if legacyRole(actor) == "regional" && region != musicRegion560(actor["region"]) {
			legacyError(w, 403, "Solo puedes crear tutores de tu región.", nil)
			return
		}
	} else if region != "" {
		n, e := strconv.Atoi(region)
		if e != nil || n < 1 || n > 12 {
			legacyError(w, 400, "La región debe estar entre 1 y 12.", nil)
			return
		}
		region = strconv.Itoa(n)
	}
	password := legacyMap(v["password"])
	salt, se := base64.StdEncoding.DecodeString(stringValue(password["salt"]))
	hash, he := base64.StdEncoding.DecodeString(stringValue(password["hash"]))
	if stringValue(password["algorithm"]) != "PBKDF2-SHA256" || numericValue(password["iterations"]) < 100000 || numericValue(password["iterations"]) > 1000000 || se != nil || he != nil || len(salt) != 16 || len(hash) != 32 {
		legacyError(w, 400, "La contraseña temporal no es válida.", nil)
		return
	}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if stringValue(u["id"]) == req.ID && legacyRole(u) == "tutor" && stringValue(u["createdBy"]) == id && strings.EqualFold(stringValue(u["email"]), email) {
			jsonOut(w, 200, map[string]any{"ok": true, "id": req.ID, "deduplicated": true})
			return
		}
		if stringValue(u["id"]) == req.ID || strings.EqualFold(stringValue(u["email"]), email) || strings.EqualFold(stringValue(u["username"]), username) {
			legacyError(w, 409, "El correo o usuario ya pertenece a una cuenta.", nil)
			return
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	u := map[string]any{"id": req.ID, "fullName": name, "email": email, "username": username, "role": "tutor", "active": true, "accountState": "active", "createdAt": now, "createdBy": id, "region": region, "password": cloneAny(password), "tutorProgramIds": programs, "theme": "classic_blue", "mustChangePassword": true}
	for _, k := range []string{"phone", "church", "mission", "union", "missionaryDistrict"} {
		if text := strings.TrimSpace(stringValue(v[k])); len(text) <= 250 {
			u[k] = text
		}
	}
	db["users"] = append(legacyUsers(db), u)
	st.State.Database = legacyPrepareCanonicalDB(db, st.State, "Tutor creado")
	st.State.ServerVersion++
	st.State.UpdatedAt = now
	for _, p := range programs {
		addLegacyNotification(&st, legacyNotification{UserID: id, OwnerID: req.ID, ActorID: id, Type: "tutor_created", Title: "Tutor creado", Message: name + " ya puede ingresar con su cuenta de tutor.", ProgramID: stringValue(p), CreatedAt: now})
	}
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "id": req.ID, "serverVersion": st.State.ServerVersion})
}

func stampAdminNotifications556(st *legacyStore, operation, program string) {
	if program == "" {
		program = "class_grouped"
	}
	for i := range st.Notifications {
		n := &st.Notifications[i]
		if n.ProgramID == "" && (n.SourceOperationID == operation || strings.HasPrefix(n.SourceOperationID, operation+":")) {
			n.ProgramID = program
		}
	}
}
func notificationProgram556(n legacyNotification) string {
	if n.ProgramID != "" {
		return n.ProgramID
	}
	if strings.HasPrefix(n.RequirementID, "LM-") {
		return musicProgram553
	}
	if n.Type == "admin_change_applied" || n.Type == "admin_change_pending" {
		return "class_grouped"
	}
	return ""
}
func musicWorkflowNotice556(st *legacyStore, db map[string]any, actor, owner map[string]any, action, now string) {
	if action != "recommend" && action != "approve" && action != "correct" {
		return
	}
	ids := []string{stringValue(owner["id"])}
	if action == "recommend" {
		for _, raw := range legacyUsers(db) {
			u := legacyMap(raw)
			if (legacyRole(u) == "departamental" || legacyRole(u) == "admin" && legacyRole(owner) == "departamental") && legacyActive(u) && managementProgram555(db, u, musicProgram553) {
				ids = append(ids, stringValue(u["id"]))
			}
		}
	}
	title := map[string]string{"recommend": "Expediente recomendado", "approve": "Expediente aprobado", "correct": "Observación en Música"}[action]
	for _, id := range ids {
		addLegacyNotification(st, legacyNotification{UserID: id, OwnerID: stringValue(owner["id"]), ActorID: stringValue(actor["id"]), ActorName: legacyName(actor), Type: "music_" + action, Title: title, Message: legacyName(owner) + " · " + title, ProgramID: musicProgram553, CreatedAt: now})
	}
}

func (s *Server) accountActor556(r *http.Request, requested string, st *legacyStore) (string, string, bool) {
	if sess, remote := s.remoteSessionFromRequest(r); remote {
		u := legacyUser(st.State.Database, sess.UserID)
		if (requested != "" && requested != sess.UserID) || !legacyActive(u) || requestRole555(r, u) != sess.Role {
			return "", "", false
		}
		return sess.UserID, requestRole555(r, u), true
	}
	scope, ok := s.auth(r)
	u := legacyUser(st.State.Database, requested)
	if !ok || !legacyActive(u) || (scope != "admin" && legacyRole(u) != "aspirant" && legacyRole(u) != "pastor_aspirant" && legacyRole(u) != "tutor") {
		return "", "", false
	}
	return requested, requestRole555(r, u), true
}
