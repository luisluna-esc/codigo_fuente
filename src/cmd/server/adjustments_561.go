package main

import (
	"net/http"
	"net/mail"
	"strings"
	"time"
)

func musicAssign561(db map[string]any, actor map[string]any) bool {
	if !musicAllowed553(db, actor) || !managementProgram555(db, actor, musicProgram553) {
		return false
	}
	switch legacyRole(actor) {
	case "admin", "departamental", "secretaria", "coordinador_musica", "administrativo", "regional":
		return true
	}
	return false
}
func musicAssignmentPatch561(db map[string]any, actor map[string]any, candidate any) string {
	if db == nil || actor == nil {
		return ""
	}
	for _, raw := range legacyUsers(candidate) {
		u := legacyMap(raw)
		old := legacyUser(db, stringValue(u["id"]))
		if !musicAssigned550(db, stringValue(u["id"])) && !musicAssigned550(legacyMap(candidate), stringValue(u["id"])) {
			continue
		}
		_, explicit := legacyMap(u["tutorByProgram"])[musicProgram553]
		_, legacyExplicit := u["tutorId"]
		if !explicit && !legacyExplicit {
			continue
		}
		after := stringValue(legacyMap(u["tutorByProgram"])[musicProgram553])
		before := stringValue(legacyMap(old["tutorByProgram"])[musicProgram553])
		if _, exists := u["tutorByProgram"]; !exists {
			after = stringValue(u["tutorId"])
		}
		if _, exists := old["tutorByProgram"]; !exists {
			before = stringValue(old["tutorId"])
		}
		if after != before && after != "" && after == stringValue(u["id"]) {
			return "Nadie puede ser su propio tutor de Música."
		}
		if after != before && (!musicAssign561(db, actor) || (old != nil && !musicAssignOwner597(db, actor, old)) || (after != "" && !musicAssignTarget597(actor, legacyUser(db, after)))) {
			return "Tu función no permite asignar o cambiar tutores de Música."
		}
	}
	return ""
}
func musicRemove561(db, owner, value map[string]any, now string) string {
	id := stringValue(owner["id"])
	staff := reviewRole556(legacyRole(owner)) && legacyRole(owner) != "admin"
	keep := staff && musicAssigned550(db, id)
	if keep {
		region := musicRegion560(value["region"])
		if region == "" {
			return "Selecciona la región del aspirante antes de retirar su responsabilidad. Su tarjeta se conservará."
		}
		owner["previousMusicRole561"] = owner["role"]
		owner["musicManagementRemoved597"] = true
		owner["region"] = region
		for _, key := range []string{"managementProgramIds", "tutorProgramIds"} {
			prior := anySlice548(owner[key])
			if owner[key] == nil {
				prior = implicitProgramList559(db, owner, key)
			}
			kept := []any{}
			for _, p := range prior {
				if stringValue(p) != musicProgram553 {
					kept = append(kept, p)
				}
			}
			owner[key] = kept
		}
		owner["musicUnassigned561"] = true
		owner["musicAccessRemoved"] = false
		ensureMapField(owner, "tutorByProgram")[musicProgram553] = ""
	} else {
		owner["musicAccessRemoved"] = true
		owner["musicUnassigned561"] = false
		for _, raw := range arrayField(db, "modules") {
			m := legacyMap(raw)
			if stringValue(m["id"]) == musicProgram553 {
				ids := []any{}
				for _, uid := range arrayField(m, "assignedUserIds") {
					if stringValue(uid) != id {
						ids = append(ids, uid)
					}
				}
				m["assignedUserIds"] = ids
			}
		}
	}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if legacyTutorAssigned(u, id, musicProgram553) {
			ensureMapField(u, "tutorByProgram")[musicProgram553] = ""
			u["musicUnassigned561"] = true
			u["updatedAt"] = now
		}
	}
	return ""
}
func musicUnassignedMember561(db, u map[string]any) bool {
	return u != nil && legacyActive(u) && u["musicAccessRemoved"] != true && musicAssigned550(db, stringValue(u["id"])) && legacyTutorForProgram(u, musicProgram553) == ""
}

// The queue deliberately exposes only directory metadata, never foreign cards.
func (s *Server) musicUnassigned561(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "method", 405)
		return
	}
	var req struct {
		ActorID    string `json:"actorId"`
		OwnerID    string `json:"ownerId"`
		Generation int64  `json:"restoreGeneration"`
	}
	req.ActorID = r.URL.Query().Get("actorId")
	if r.Method == "POST" {
		if err := readJSONLimited(r, &req, 1<<20); err != nil {
			legacyError(w, 400, "Datos inválidos.", nil)
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
	id, _, ok := s.accountActor556(r, req.ActorID, &st)
	db := legacyMap(cloneAny(st.State.Database))
	actor := legacyUser(db, id)
	if !ok || personalRequest555(r) || !reviewRole556(legacyRole(actor)) || !musicAllowed553(db, actor) || !managementProgram555(db, actor, musicProgram553) {
		legacyError(w, 403, "Lista reservada a responsables activos de Música.", nil)
		return
	}
	if r.Method == "POST" {
		owner := legacyUser(db, req.OwnerID)
		if legacyRole(actor) != "regional" || !musicUnassignedMember561(db, owner) || !musicScope550(db, actor, owner) || musicRegion560(actor["region"]) == "" {
			legacyError(w, 403, "Solicitud no permitida.", nil)
			return
		}
		if req.Generation != legacyCanonicalRestoreGeneration(st.State) {
			legacyError(w, 409, "Actualiza la lista antes de solicitar.", nil)
			return
		}
		ensureMapField(owner, "musicRegionRequests561")[id] = map[string]any{"region": actor["region"], "name": actor["fullName"], "at": time.Now().UTC().Format(time.RFC3339Nano)}
		owner["updatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
		st.State.Database = legacyPrepareCanonicalDB(db, st.State, "Solicitud regional de asignación")
		st.State.ServerVersion++
		st.State.UpdatedAt = stringValue(owner["updatedAt"])
		if err := s.saveLegacy(st); err != nil {
			legacyError(w, 500, err.Error(), nil)
			return
		}
	}
	rows := []any{}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if musicUnassignedMember561(db, u) && stringValue(u["id"]) != id && musicScope550(db, actor, u) {
			rows = append(rows, map[string]any{"id": u["id"], "fullName": u["fullName"], "region": u["region"], "role": u["role"], "updatedAt": u["updatedAt"], "requests": u["musicRegionRequests561"]})
		}
	}
	jsonOut(w, 200, map[string]any{"ok": true, "rows": rows})
}

// Admin-only atomic batch registration. Preview is performed against current
// server data; commit repeats validation so duplicates cannot race the preview.
func (s *Server) accountsImport561(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method", 405)
		return
	}
	var req struct {
		ActorID    string           `json:"actorId"`
		Action     string           `json:"action"`
		BatchID    string           `json:"batchId"`
		Generation int64            `json:"restoreGeneration"`
		Rows       []map[string]any `json:"rows"`
	}
	if err := readJSONLimited(r, &req, 8<<20); err != nil {
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
	id, role, ok := s.accountActor556(r, req.ActorID, &st)
	if !ok || role != "admin" || personalRequest555(r) {
		legacyError(w, 403, "Solo el Administrador puede importar usuarios.", nil)
		return
	}
	if req.Action != "preview" && req.Action != "commit" {
		legacyError(w, 400, "Operación inválida.", nil)
		return
	}
	db := legacyMap(cloneAny(st.State.Database))
	receipts := ensureMapField(db, "importBatches561")
	if req.Action == "commit" && req.BatchID != "" && receipts[req.BatchID] != nil {
		jsonOut(w, 200, map[string]any{"ok": true, "alreadyImported": true})
		return
	}
	if len(req.Rows) == 0 || len(req.Rows) > 500 || req.Generation != legacyCanonicalRestoreGeneration(st.State) {
		legacyError(w, 400, "Importa entre 1 y 500 filas y actualiza los datos del servidor.", nil)
		return
	}
	users := []any{}
	issues := []any{}
	actor := legacyUser(db, id)
	allowedRoles := map[string]bool{"coordinador_sgm": true, "aspirant": true, "pastor_aspirant": true, "departamental": true, "secretaria": true, "coordinador_musica": true, "administrativo": true, "regional": true, "tutor": true, "administrativo_sgm": true, "lider_musica": true}
	for i, v := range req.Rows {
		name := strings.TrimSpace(stringValue(v["fullName"]))
		email := strings.ToLower(strings.TrimSpace(stringValue(v["email"])))
		rRole := stringValue(v["role"])
		program := stringValue(v["programId"])
		issue := ""
		address, e := mail.ParseAddress(email)
		if name == "" || len(name) > 200 || e != nil || address.Address != email {
			issue = "Completa nombres y correo válido."
		}
		region := musicRegion560(v["region"])
		if region == "" {
			issue = "Región inválida: usa 1–12 o MBOS."
		}
		if !allowedRoles[rRole] {
			issue = "Selecciona una función válida."
		}
		activeProgram := false
		for _, raw := range arrayField(db, "modules") {
			m := legacyMap(raw)
			if stringValue(m["id"]) == program && m["active"] != false {
				activeProgram = true
			}
		}
		if !activeProgram || musicRequest553(r) && program != musicProgram553 {
			issue = "Módulo no disponible."
		}
		if (rRole == "departamental" || rRole == "coordinador_musica" || rRole == "administrativo_sgm" || rRole == "lider_musica") && program != musicProgram553 {
			issue = "Esta función corresponde solo a Música."
		}
		for _, raw := range legacyUsers(db) {
			u := legacyMap(raw)
			if strings.EqualFold(stringValue(u["email"]), email) || strings.EqualFold(stringValue(u["username"]), email) {
				issue = "Correo o usuario duplicado; no se sobrescribirá."
			}
		}
		birth := strings.TrimSpace(stringValue(v["birthDate"]))
		if birth != "" {
			date, e := time.Parse("2006-01-02", birth)
			if e != nil || date.After(time.Now()) || date.Year() < 1900 {
				issue = "Fecha inválida: usa AAAA-MM-DD."
			}
		}
		user := map[string]any{"id": stringValue(v["id"]), "fullName": name, "email": email, "username": email, "role": rRole, "region": region, "birthDate": birth, "mission": "Misión Boliviana Occidental Sur", "union": "Unión Boliviana", "active": true, "accountState": "active", "theme": "classic_blue", "mustChangePassword": true, "createdBy": actor["id"], "createdAt": time.Now().UTC().Format(time.RFC3339Nano), "tutorByProgram": map[string]any{program: ""}, "managementProgramIds": []any{}, "tutorProgramIds": []any{}}
		for _, k := range []string{"phone", "address", "church", "missionaryDistrict"} {
			text := strings.TrimSpace(stringValue(v[k]))
			if len(text) > 250 {
				issue = "Texto demasiado largo."
			}
			user[k] = text
		}
		if rRole == "tutor" || rRole == "lider_musica" {
			user["tutorProgramIds"] = []any{program}
		} else if rRole != "aspirant" && rRole != "pastor_aspirant" {
			user["managementProgramIds"] = []any{program}
		}
		if program == musicProgram553 && (rRole == "aspirant" || rRole == "pastor_aspirant" || v["training"] == true) {
			tutorID := stringValue(v["tutorId"])
			if tutorID == "" || tutorID == stringValue(user["id"]) || !musicTutor560(db, legacyUser(db, tutorID)) {
				issue = "Selecciona un tutor de Música para el aspirante."
			} else {
				ensureMapField(user, "tutorByProgram")[program] = tutorID
			}
		}
		if req.Action == "commit" {
			user["password"] = v["password"]
			// Reuse the bounded identity/password validator; Music trainees require a tutor.
			probe := legacyMap(cloneAny(user))
			probe["role"] = "aspirant"
			probe["tutorId"] = actor["id"]
			if program == musicProgram553 && (rRole == "aspirant" || rRole == "pastor_aspirant" || v["training"] == true) {
				probe["tutorId"] = v["tutorId"]
			}
			if _, message := musicNewAspirant560(db, actor, probe); message != "" {
				issue = message
			}
		}
		if issue != "" {
			issues = append(issues, map[string]any{"row": i + 1, "error": issue})
		}
		users = append(users, user)
		db["users"] = append(legacyUsers(db), user)
		if rRole == "aspirant" || rRole == "pastor_aspirant" || v["training"] == true {
			for _, raw := range arrayField(db, "modules") {
				m := legacyMap(raw)
				if stringValue(m["id"]) == program {
					m["assignedUserIds"] = append(arrayField(m, "assignedUserIds"), user["id"])
				}
			}
			if program == musicProgram553 {
				user["musicUnassigned561"] = legacyTutorForProgram(user, musicProgram553) == ""
			}
		}
	}
	if len(issues) > 0 {
		jsonOut(w, 200, map[string]any{"ok": false, "issues": issues, "error": "Corrige las filas señaladas antes de registrar."})
		return
	}
	if req.Action == "preview" {
		jsonOut(w, 200, map[string]any{"ok": true, "count": len(users)})
		return
	}
	if req.BatchID == "" || len(req.BatchID) > 100 {
		legacyError(w, 400, "Falta el identificador del lote.", nil)
		return
	}
	receipts[req.BatchID] = map[string]any{"count": len(users), "actorId": id, "at": time.Now().UTC().Format(time.RFC3339Nano)}
	st.State.Database = legacyPrepareCanonicalDB(db, st.State, "Importación de usuarios confirmada")
	st.State.ServerVersion++
	st.State.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "count": len(users), "serverVersion": st.State.ServerVersion})
}

func (s *Server) accountDirectory561(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
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
	_, role, ok := s.accountActor556(r, r.URL.Query().Get("actorId"), &st)
	if !ok || role != "admin" || personalRequest555(r) {
		legacyError(w, 403, "Solo el Administrador puede consultar todas las cuentas.", nil)
		return
	}
	db := legacyMap(st.State.Database)
	rows := []any{}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		programs := []any{}
		for _, rawModule := range arrayField(db, "modules") {
			m := legacyMap(rawModule)
			p := stringValue(m["id"])
			if assignedProgram555(db, stringValue(u["id"]), p) || managementProgram555(db, u, p) {
				programs = append(programs, p)
			}
		}
		rows = append(rows, map[string]any{"id": u["id"], "fullName": u["fullName"], "role": u["role"], "email": u["email"], "region": u["region"], "musicAccessRemoved": u["musicAccessRemoved"], "directoryPrograms561": programs})
	}
	jsonOut(w, 200, map[string]any{"ok": true, "rows": rows})
}
