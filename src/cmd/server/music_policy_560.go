package main

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AJUSTE FUNCIONES.xlsx, rows 3–11. These capabilities apply only to Music.
type musicPolicy560 struct {
	Review, Recommend, Final, Notices, Personal, CreateTutor, Tutors         bool
	Instructions, Profile, Registry, Events, Annual, Templates, InvestOthers bool
}

var musicPolicies560 = map[string]musicPolicy560{
	"admin":              {true, true, true, true, true, true, true, true, true, true, true, true, true, true},
	"departamental":      {true, false, true, true, true, true, true, true, true, true, true, true, true, true},
	"secretaria":         {false, true, false, true, true, true, true, true, true, true, true, true, true, true},
	"coordinador_musica": {true, true, false, true, true, true, true, true, true, true, true, false, true, true},
	"administrativo":     {true, true, false, true, true, true, true, false, false, false, true, false, false, false},
	"regional":           {true, true, false, true, true, true, true, false, false, false, false, false, false, false},
	"tutor":              {true, false, false, true, true, false, false, false, false, false, false, false, false, false},
	"pastor_aspirant":    {},
	"aspirant":           {},
}

func musicIssue560(db, owner map[string]any, actorID, kind, eventID, now string) {
	data := musicData550(db)
	certs := ensureMapField(data, "certificates")
	id := stringValue(owner["id"])
	key := id + ":" + kind + ":" + eventID
	if c := legacyMap(certs[key]); c != nil {
		c["fullName"] = owner["fullName"]
		return
	}
	c := map[string]any{"id": key, "token": randomMusicID550(), "ownerId": id, "fullName": owner["fullName"], "kind": kind, "title": certificateTypes550[kind], "eventId": eventID, "issuedAt": now, "issuedBy": actorID, "orientation": "landscape", "church": owner["church"], "mission": owner["mission"], "registration": cloneAny(owner["musicMinistryRegistration"]), "investiture": cloneAny(mapField(data, "investitures")[id]), "event": cloneAny(mapField(data, "events")[eventID])}
	if event := legacyMap(c["event"]); event != nil {
		delete(event, "participants")
	}
	certs[key] = c
}

func musicRegion560(v any) string {
	r := strings.ToUpper(strings.TrimSpace(stringValue(v)))
	r = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(r, "REGIÓN"), "REGION"))
	if r == "MBOS" {
		return r
	}
	n, err := strconv.Atoi(r)
	if err == nil && n >= 1 && n <= 12 {
		return strconv.Itoa(n)
	}
	return ""
}
func musicSameRegion560(a, b map[string]any) bool {
	r := musicRegion560(a["region"])
	return r != "" && r == musicRegion560(b["region"])
}
func musicTutor560(db map[string]any, tutor map[string]any) bool {
	if tutor == nil || !legacyActive(tutor) || tutor["musicAccessRemoved"] == true {
		return false
	}
	if reviewRole556(legacyRole(tutor)) && managementProgram555(db, tutor, musicProgram553) {
		return true
	}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if musicAssigned550(db, stringValue(u["id"])) && legacyActive(u) && legacyTutorAssigned(u, stringValue(tutor["id"]), musicProgram553) {
			return true
		}
	}
	return false
}
func musicSeeTutor560(db map[string]any, actor, tutor map[string]any) bool {
	if actor == nil || (!musicPolicies560[legacyRole(actor)].Tutors || !musicPermit602(db, actor, "tutors.view")) || !managementProgram555(db, actor, musicProgram553) || !musicTutor560(db, tutor) {
		return false
	}
	if legacyRole(actor) == "regional" {
		return stringValue(tutor["createdBy"]) == stringValue(actor["id"])
	}
	return true
}
func musicReview560(db map[string]any, actor, owner map[string]any) bool {
	return actor != nil && owner != nil && stringValue(actor["id"]) != stringValue(owner["id"]) && musicPolicies560[legacyRole(actor)].Review && musicPermit602(db, actor, "reports.view") && musicScope550(db, actor, owner)
}
func musicFinal560(db map[string]any, actor, owner map[string]any) bool {
	if actor == nil || owner == nil || stringValue(actor["id"]) == stringValue(owner["id"]) || !musicScope550(db, actor, owner) {
		return false
	}
	// Administrators have every management action; self approval remains blocked.
	return legacyRole(actor) == "departamental" || legacyRole(actor) == "admin"
}
func musicDelete560(db map[string]any, actor, owner map[string]any) bool {
	if !musicPermit602(db, actor, "people.remove") {
		return false
	}
	if actor == nil || owner == nil || !legacyActive(actor) || actor["musicAccessRemoved"] == true || stringValue(actor["id"]) == stringValue(owner["id"]) || !managementProgram555(db, actor, musicProgram553) {
		return false
	}
	r, target := legacyRole(actor), legacyRole(owner)
	if r == "admin" {
		return true
	}
	if target == "admin" || target == "departamental" {
		return false
	}
	if target == "secretaria" {
		return r == "departamental"
	}
	return r == "departamental" || r == "secretaria" || r == "coordinador_musica"

}

// Filter Music's part of an ordinary SGL snapshot without changing any other program.
func musicScopeSnapshot560(db map[string]any, actorID, role string) map[string]any {
	hasMusic := false
	for _, raw := range arrayField(db, "modules") {
		if stringValue(legacyMap(raw)["id"]) == musicProgram553 {
			hasMusic = true
		}
	}
	if !hasMusic {
		return db
	}
	actor := legacyUser(db, actorID)
	if actor == nil {
		return db
	}
	if role == "aspirant" && legacyRole(actor) != "aspirant" {
		actor = legacyMap(cloneAny(actor))
		if legacyRole(actor) == "admin" {
			actor["musicAccessRemoved"] = true
		}
		actor["role"] = role
	}
	directory := []any{}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if musicSeeTutor560(db, actor, u) {
			directory = append(directory, map[string]any{"id": u["id"], "fullName": u["fullName"], "role": u["role"], "region": u["region"], "active": true, "accountState": "active", "musicIsTutor560": true, "createdBy": u["createdBy"]})
		}
	}
	db["musicTutorDirectory560"] = directory
	for id, raw := range legacyMap(db["programProgress"]) {
		if !musicPermit602(db, actor, "reports.view") || !musicScope550(db, actor, legacyUser(db, id)) {
			delete(legacyMap(raw), musicProgram553)
		}
	}
	if db["music550"] != nil {
		db["music550"] = musicView550(db, actor)
	}
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if stringValue(m["id"]) != musicProgram553 {
			continue
		}
		ids := []any{}
		for _, id := range arrayField(m, "assignedUserIds") {
			if musicScope550(db, actor, legacyUser(db, stringValue(id))) {
				ids = append(ids, id)
			}
		}
		m["assignedUserIds"] = ids
	}
	return db
}

// Keep personal-profile editing and mandatory regions authoritative on old sync
// paths too. Saving a report never forces a legacy user to invent a region.
func musicAccountPatch560(db map[string]any, actor map[string]any, candidate any) string {
	for _, raw := range legacyUsers(candidate) {
		u := legacyMap(raw)
		id := stringValue(u["id"])
		old := legacyUser(db, id)
		if old == nil {
			if musicRegion560(u["region"]) == "" {
				return "Selecciona una región: 1–12 o MBOS."
			}
			continue
		}
		for _, key := range []string{"fullName", "email", "username", "phone", "address", "church", "mission", "union", "missionaryDistrict", "birthDate", "gender", "documentNumber", "region", "churchPosition", "clubPosition"} {
			if value, exists := u[key]; exists && stringValue(value) != stringValue(old[key]) {
				if !musicPolicies560[legacyRole(actor)].Profile {
					return "Tu función no permite modificar datos personales en Música."
				}
				if key == "region" && musicRegion560(value) == "" {
					return "Selecciona una región: 1–12 o MBOS."
				}
			}
		}
	}
	return ""
}

// Metadata-only view changes: evidence is downloaded on demand, never at each switch.
func (s *Server) accountView560(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", 405)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacyLoginState()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	store := legacyStore{State: st}
	id, role, ok := s.accountActor556(r, r.URL.Query().Get("actorId"), &store)
	if !ok || musicRequest553(r) && !musicAllowed553(legacyMap(st.Database), legacyUser(st.Database, id)) {
		legacyError(w, 403, "Sesión no autorizada.", nil)
		return
	}
	snap := s.filteredLegacySnapshotMode(st, id, role, false)
	if musicRequest553(r) {
		snap["database"] = musicProject553(snap["database"], id, role)
	}
	if _, remote := s.remoteSessionFromRequest(r); remote {
		stripRemoteCredentialFields508(snap)
	}
	jsonOutRemote(w, r, 200, snap)
}

// Scoped Music account operations; role, credentials and other SGL programs are untouched.
func (s *Server) musicAccounts560(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	var req struct {
		ActorID    string         `json:"actorId"`
		OwnerID    string         `json:"ownerId"`
		Action     string         `json:"action"`
		Generation int64          `json:"restoreGeneration"`
		ExpectedAt string         `json:"expectedUpdatedAt"`
		Value      map[string]any `json:"value"`
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
	actor, owner := legacyUser(db, id), legacyUser(db, req.OwnerID)
	if key := musicActionKey602("/api/accounts/music", req.Action, req.Value); key != "" && !musicPermit602(db, actor, key) {
		legacyError(w, 403, "Función desactivada.", nil)
		return
	}
	if ok && !personalRequest555(r) && req.Action == "create_aspirant" && musicAllowed553(db, actor) && musicPolicies560[legacyRole(actor)].Notices {
		if req.Generation != legacyCanonicalRestoreGeneration(st.State) {
			legacyError(w, 409, "Actualiza los datos antes de crear la cuenta.", nil)
			return
		}
		user, message := musicNewAspirant560(db, actor, req.Value)
		if message != "" {
			legacyError(w, 400, message, nil)
			return
		}
		db["users"] = append(legacyUsers(db), user)
		for _, raw := range arrayField(db, "modules") {
			m := legacyMap(raw)
			if stringValue(m["id"]) == musicProgram553 && m["active"] != false {
				m["assignedUserIds"] = append(arrayField(m, "assignedUserIds"), user["id"])
			}
		}
		st.State.Database = legacyPrepareCanonicalDB(db, st.State, "Aspirante de Música creado")
		st.State.ServerVersion++
		st.State.UpdatedAt = stringValue(user["createdAt"])
		if err := s.saveLegacy(st); err != nil {
			legacyError(w, 500, err.Error(), nil)
			return
		}
		jsonOut(w, 200, map[string]any{"ok": true, "id": user["id"], "serverVersion": st.State.ServerVersion})
		return
	}
	if !ok || (personalRequest555(r) && req.Action != "self_photo" && !(req.Action == "profile" && id == req.OwnerID)) || actor == nil || owner == nil || !musicAllowed553(db, actor) || !musicAllowed553(db, owner) {
		legacyError(w, 403, "Cuenta fuera del alcance de Música.", nil)
		return
	}
	if req.Generation != legacyCanonicalRestoreGeneration(st.State) || req.ExpectedAt != stringValue(owner["updatedAt"]) {
		legacyError(w, 409, "La cuenta cambió. Actualiza antes de guardar.", nil)
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	switch req.Action {
	case "assign_task":
		if personalRequest555(r) || !musicPolicies560[legacyRole(actor)].Notices || !musicAssignOwner597(db, actor, owner) || !musicAssigned550(db, req.OwnerID) {
			legacyError(w, 403, "Aspirante fuera de tu alcance.", nil)
			return
		}
		rid, taskID := stringValue(req.Value["requirementId"]), stringValue(req.Value["id"])
		valid := false
		for _, x := range musicRequirements550 {
			if x == rid {
				valid = true
			}
		}
		start, e1 := time.Parse("2006-01-02", stringValue(req.Value["plannedDate"]))
		end, e2 := time.Parse("2006-01-02", stringValue(req.Value["dueDate"]))
		if !valid || taskID == "" || len(taskID) > 100 || e1 != nil || e2 != nil || end.Before(start) || len(stringValue(req.Value["instructions"])) > 4000 {
			legacyError(w, 400, "Completa requisito, fechas válidas e indicaciones de hasta 4000 caracteres.", nil)
			return
		}
		progress := reviewProgressContainer(db, req.OwnerID, musicProgram553)
		status := stringValue(mapField(progress, "status")[rid])
		if status == "approved" || status == "in_review" {
			legacyError(w, 409, "El requisito ya fue enviado o revisado.", nil)
			return
		}
		for _, raw := range arrayField(db, "tasks") {
			task := legacyMap(raw)
			if stringValue(task["id"]) == taskID || (musicItem553(task) && stringValue(task["assigneeId"]) == req.OwnerID && stringValue(task["requirementId"]) == rid && task["active"] != false && stringValue(task["deletedAt"]) == "" && stringValue(task["presentedAt"]) == "") {
				legacyError(w, 409, "Ya existe una tarea activa para este requisito. Actualiza la lista.", nil)
				return
			}
		}
		task := map[string]any{"id": taskID, "programId": musicProgram553, "assigneeId": req.OwnerID, "requirementId": rid, "plannedDate": req.Value["plannedDate"], "dueDate": req.Value["dueDate"], "instructions": req.Value["instructions"], "createdAt": now, "updatedAt": now, "createdBy": id, "createdByRole": actor["role"], "selfPlanned": false, "active": true}
		db["tasks"] = append(arrayField(db, "tasks"), task)
		if status == "" || status == "pending" {
			ensureMapField(progress, "status")[rid] = "assigned"
		}
		addLegacyNotification(&st, legacyNotification{UserID: req.OwnerID, Type: "task_assigned", Title: "Tarea de Música asignada", Message: rid, ProgramID: musicProgram553, RequirementID: rid, ActorID: id, ActorName: stringValue(actor["fullName"]), CreatedAt: now, SourceOperationID: taskID})
	case "remove_empty":
		if legacyRole(actor) != "admin" {
			legacyError(w, 403, "Solo el Administrador puede depurar registros incompletos.", nil)
			return
		}
		if reason := musicEmptyReason597(db, owner, st.State.Files); reason != "" {
			legacyError(w, 409, reason, nil)
			return
		}
		// Preserve the exact record in a private audit archive before removing its empty enrollment.
		db["musicEmptyArchive597"] = append(arrayField(db, "musicEmptyArchive597"), map[string]any{"user": cloneAny(owner), "removedBy": id, "removedAt": now})
		kept := []any{}
		for _, raw := range legacyUsers(db) {
			if stringValue(legacyMap(raw)["id"]) != req.OwnerID {
				kept = append(kept, raw)
			}
		}
		db["users"] = kept
		for _, raw := range arrayField(db, "modules") {
			m := legacyMap(raw)
			ids := []any{}
			for _, uid := range arrayField(m, "assignedUserIds") {
				if stringValue(uid) != req.OwnerID {
					ids = append(ids, uid)
				}
			}
			m["assignedUserIds"] = ids
		}
		delete(legacyMap(db["programProgress"]), req.OwnerID)
		delete(legacyMap(db["progress"]), req.OwnerID)
	case "self_photo":
		if req.ActorID != req.OwnerID {
			legacyError(w, 403, "Solo puedes cambiar tu propia foto.", nil)
			return
		}
		photo := stringValue(req.Value["photo"])
		if !validMusicPhoto592(photo) {
			legacyError(w, 400, "Usa una foto PNG/JPG válida de hasta 2 MB.", nil)
			return
		}
		owner["photo"] = photo
	case "profile":
		if id != req.OwnerID && (!musicPolicies560[legacyRole(actor)].Profile || !managementProgram555(db, actor, musicProgram553)) {
			legacyError(w, 403, "Tu función no permite editar datos personales.", nil)
			return
		}
		if email, exists := req.Value["email"]; exists && !strings.EqualFold(strings.TrimSpace(stringValue(email)), stringValue(owner["email"])) {
			if req.Value["emailChangeConfirmed"] != true {
				legacyError(w, 400, "Confirma el cambio de correo antes de guardar.", nil)
				return
			}
			for _, raw := range legacyUsers(db) {
				other := legacyMap(raw)
				if stringValue(other["id"]) != req.OwnerID && strings.EqualFold(strings.TrimSpace(stringValue(email)), stringValue(other["email"])) {
					legacyError(w, 409, "El correo ya pertenece a otra cuenta.", nil)
					return
				}
			}
		}
		if email, ok := req.Value["email"]; ok && !strings.EqualFold(strings.TrimSpace(stringValue(email)), stringValue(owner["email"])) {
			delete(owner, "identitySubject610")
		}
		region := musicRegion560(req.Value["region"])
		if id == req.OwnerID {
			region = musicRegion560(owner["region"])
		} // Region affects regional access: personal edits cannot move authority boundaries.
		if region == "" && id != req.OwnerID {
			legacyError(w, 400, "Selecciona una región: 1 a 12 o MBOS.", nil)
			return
		}
		if strings.TrimSpace(stringValue(req.Value["fullName"])) == "" {
			legacyError(w, 400, "Completa el nombre.", nil)
			return
		}
		for _, key := range []string{"fullName", "email", "phone", "address", "church", "mission", "union", "missionaryDistrict", "birthDate", "gender", "documentNumber"} {
			if v, exists := req.Value[key]; exists {
				text := strings.TrimSpace(stringValue(v))
				if len(text) > 250 {
					legacyError(w, 400, "Texto demasiado largo.", nil)
					return
				}
				owner[key] = text
			}
		}
		owner["region"] = region
	case "assign_self":
		legacyError(w, 403, "Usa Asignar / cambiar tutor para confirmar el tutor y la región.", nil)
		return
	case "assign_tutor":
		if !musicAssign561(db, actor) || !musicAssignOwner597(db, actor, owner) {
			legacyError(w, 403, "Tu función no permite asignar tutores.", nil)
			return
		}
		tutorID := stringValue(req.Value["tutorId"])
		tutor := legacyUser(db, tutorID)
		if tutorID == "" || tutorID == req.OwnerID || !musicTutor560(db, tutor) || !musicAssignTarget597(actor, tutor) {
			legacyError(w, 400, "Selecciona una autoridad o tutor activo de Música distinto del aspirante.", nil)
			return
		}
		if regionValue, exists := req.Value["region"]; exists {
			region := musicRegion560(regionValue)
			if region == "" {
				legacyError(w, 400, "Selecciona una región válida.", nil)
				return
			}
			owner["region"] = region
		}
		previousTutor := legacyTutorForProgram(owner, musicProgram553)
		ensureMapField(owner, "tutorByProgram")[musicProgram553] = tutorID
		musicTutorNotifications597(&st, actor, owner, previousTutor, tutorID, now)
		owner["musicUnassigned561"] = tutorID == ""
		delete(owner, "musicRegionRequests561")
	case "remove":
		if !musicDelete560(db, actor, owner) {
			legacyError(w, 403, "Tu función no permite eliminar esta cuenta de Música.", nil)
			return
		}
		if message := musicRemove561(db, owner, req.Value, now); message != "" {
			legacyError(w, 400, message, nil)
			return
		}

	default:
		legacyError(w, 400, "Operación inválida.", nil)
		return
	}
	owner["updatedAt"] = now
	st.State.Database = legacyPrepareCanonicalDB(db, st.State, "Cuenta de Música actualizada")
	st.State.ServerVersion++
	st.State.UpdatedAt = now
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "serverVersion": st.State.ServerVersion})
}

func musicNewAspirant560(db map[string]any, actor, value map[string]any) (map[string]any, string) {
	id, name := strings.TrimSpace(stringValue(value["id"])), strings.TrimSpace(stringValue(value["fullName"]))
	email, username := strings.ToLower(strings.TrimSpace(stringValue(value["email"]))), strings.ToLower(strings.TrimSpace(stringValue(value["username"])))
	region := musicRegion560(value["region"])
	if id == "" || len(id) > 100 || name == "" || len(name) > 200 || !strings.Contains(email, "@") || len(email) > 250 || username == "" || len(username) > 100 || region == "" {
		return nil, "Completa nombre, correo, usuario y región (1–12 o MBOS)."
	}
	if legacyRole(actor) == "regional" && region != musicRegion560(actor["region"]) {
		return nil, "Solo puedes registrar aspirantes en tu propia región."
	}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if stringValue(u["id"]) == id || strings.EqualFold(stringValue(u["email"]), email) || strings.EqualFold(stringValue(u["username"]), username) {
			return nil, "El usuario o correo ya pertenece a otra cuenta."
		}
	}
	p := legacyMap(value["password"])
	salt, se := base64.StdEncoding.DecodeString(stringValue(p["salt"]))
	hash, he := base64.StdEncoding.DecodeString(stringValue(p["hash"]))
	if stringValue(p["algorithm"]) != "PBKDF2-SHA256" || numericValue(p["iterations"]) < 100000 || numericValue(p["iterations"]) > 1000000 || se != nil || he != nil || len(salt) != 16 || len(hash) != 32 {
		return nil, "La contraseña temporal no es válida."
	}
	role := stringValue(value["role"])
	if role != "pastor_aspirant" {
		role = "aspirant"
	}
	tutor := stringValue(value["tutorId"])
	if tutor == "" && legacyRole(actor) != "regional" {
		return nil, "Debes asignar un tutor para crear el aspirante."
	}
	if tutor != "" && !musicAssign561(db, actor) {
		return nil, "Tu función puede crear aspirantes, pero no asignar tutores."
	}
	if tutor != "" && (tutor == id || !musicTutor560(db, legacyUser(db, tutor))) {
		return nil, "Selecciona un tutor habilitado para Música."
	}
	if tutor != "" && !musicAssignTarget597(actor, legacyUser(db, tutor)) {
		return nil, "Selecciona un administrativo, regional, tutor o Líder de Música."
	}
	user := map[string]any{"id": id, "fullName": name, "email": email, "username": username, "region": region, "role": role, "active": true, "accountState": "active", "theme": "classic_blue", "mustChangePassword": true, "createdBy": actor["id"], "createdAt": time.Now().UTC().Format(time.RFC3339Nano), "password": cloneAny(p), "tutorByProgram": map[string]any{musicProgram553: tutor}}
	for _, k := range []string{"phone", "address", "church", "mission", "union", "missionaryDistrict", "birthDate"} {
		text := strings.TrimSpace(stringValue(value[k]))
		if len(text) > 250 {
			return nil, "Texto demasiado largo."
		}
		user[k] = text
	}
	return user, ""
}
