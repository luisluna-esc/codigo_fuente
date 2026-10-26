package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
)

// These switches govern available operations inside the existing role and owner scope.
// A switch never assigns a role, reassigns a learner, or permits self approval.
var musicKeys602 = []string{"menu.music_ministries", "ministries.view", "ministries.create", "ministries.modalities", "ministries.edit", "ministries.delete", "ministries.link", "ministries.unlink", "ministries.export", "menu.aspirants", "menu.tutors", "menu.reviews", "menu.music_reports571", "menu.tasks", "menu.materials", "menu.music_module", "menu.music_certificates", "menu.music_events", "menu.profile", "menu.users", "whatsapp", "people.view", "people.create", "people.edit", "people.assign", "people.remove", "tutors.view", "tutors.create", "reports.view", "reports.submit", "reports.correct", "reports.approve", "reports.export", "reports.print", "events.view", "events.edit", "events.delete", "events.enroll", "events.attendance", "tasks.edit", "certificates.view", "certificates.issue", "certificates.revoke", "certificates.delete", "certificates.templates", "content.quotes", "content.quotes.bible", "content.quotes.egw", "content.quotes.music", "content.quotes.famous", "content.materials", "content.notices", "mentor.view", "mentor.write", "achievements", "birthdays.auto", "birthdays.calendar", "birthdays.all", "birthdays.reminders", "ui.reduceMotion", "ui.help", "content.photoReminder", "content.showNotices", "mentor.notify", "tasks.selfPlan", "events.external"}

func musicPermit602(db, actor map[string]any, key string) bool {
	if actor == nil || key == "whatsapp" || key == "achievements" || key == "menu.drive" || strings.HasPrefix(key, "drive.") {
		return false
	}
	if (legacyRole(actor) == "admin" && key != "ui.help" && key != "birthdays.reminders") || key == "events.view" || key == "menu.music_events" || key == "ministries.view" || key == "menu.music_ministries" {
		return true
	}
	cfg := mapField(musicData550(db), "permissions602")
	value := key != "birthdays.reminders" && key != "ui.help"
	if len(cfg) == 0 {
		if b, ok := mapField(musicData550(db), "effective602")[key].(bool); ok {
			value = b
		}
	}
	if b, ok := mapField(mapField(cfg, "roles"), legacyRole(actor))[key].(bool); ok {
		value = b
	}
	if b, ok := mapField(mapField(cfg, "users"), stringValue(actor["id"]))[key].(bool); ok {
		value = b
	}
	return value
}
func musicEffective602(db, actor map[string]any) map[string]any {
	out := map[string]any{}
	for _, k := range musicKeys602 {
		out[k] = musicPermit602(db, actor, k)
	}
	return out
}
func musicProjection602(db, actor, d map[string]any) {
	delete(d, "certificateDeletions618")
	delete(d, "ministryDeletions624")
	musicProjection610(db, actor, d)
	d["effective602"] = musicEffective602(db, actor)
	if legacyRole(actor) != "admin" {
		delete(d, "permissions602")
		delete(d, "permissionsAudit602")
	}
	notes := []any{}
	if musicPermit602(db, actor, "mentor.view") {
		for _, raw := range arrayField(d, "mentoring602") {
			n := legacyMap(raw)
			if musicScope550(db, actor, legacyUser(db, stringValue(n["ownerId"]))) {
				notes = append(notes, n)
			}
		}
	}
	d["mentoring602"] = notes
	milestones := map[string]any{}
	if musicPermit602(db, actor, "achievements") {
		for _, raw := range legacyUsers(db) {
			u := legacyMap(raw)
			id := stringValue(u["id"])
			if !musicScope550(db, actor, u) {
				continue
			}
			approved := 0
			p := legacyProgressForRequirement(db, id, musicProgram553, false)
			for _, rid := range musicRequirements550 {
				if stringValue(mapField(p, "status")[rid]) == "approved" {
					approved++
				}
			}
			milestones[id] = map[string]any{"approved": approved, "total": len(musicRequirements550)}
		}
	}
	d["milestones602"] = milestones
	if !musicPermit602(db, actor, "certificates.view") {
		d["certificates"] = map[string]any{}
	}
	if !musicPermit602(db, actor, "events.view") {
		d["events"] = map[string]any{}
	}
}
func musicMutation602(db, actor, owner map[string]any, action string, v map[string]any, now string, personal bool) string {
	d := musicData550(db)
	switch action {
	case "permissions602":
		if legacyRole(actor) != "admin" || personal {
			return "Solo el Administrador puede configurar permisos."
		}
		valid := map[string]bool{}
		for _, k := range musicKeys602 {
			valid[k] = true
		}
		cfg := map[string]any{}
		for _, scope := range []string{"roles", "users"} {
			out := map[string]any{}
			for id, raw := range mapField(v, scope) {
				if scope == "roles" {
					if _, ok := musicPolicies560[id]; !ok {
						return "Función desconocida."
					}
				} else if legacyUser(db, id) == nil {
					return "Usuario desconocido."
				}
				entry := map[string]any{}
				for key, value := range legacyMap(raw) {
					if !valid[key] {
						return "Permiso desconocido."
					}
					b, ok := value.(bool)
					if !ok {
						return "El permiso debe ser una casilla activada o desactivada."
					}
					entry[key] = b
				}
				out[id] = entry
			}
			cfg[scope] = out
		}
		message := strings.TrimSpace(stringValue(v["whatsappMessage"]))
		if len([]rune(message)) > 1000 {
			return "El mensaje inicial admite hasta 1000 caracteres."
		}
		cfg["whatsappMessage"] = message
		audit := arrayField(d, "permissionsAudit602")
		audit = append(audit, map[string]any{"actorId": actor["id"], "at": now, "before": cloneAny(d["permissions602"]), "after": cloneAny(cfg)})
		d["permissionsAudit602"] = audit
		d["permissions602"] = cfg
	case "mentor602":
		if personal || owner == nil || !musicPermit602(db, actor, "mentor.write") || !musicPermit602(db, actor, "mentor.view") || !legacyTutorAssigned(owner, stringValue(actor["id"]), musicProgram553) || !musicScope550(db, actor, owner) || stringValue(actor["id"]) == stringValue(owner["id"]) {
			return "Solo el tutor asignado puede escribir este acompañamiento."
		}
		message := strings.TrimSpace(stringValue(v["text"]))
		if message == "" || len([]rune(message)) > 1500 {
			return "Escribe un mensaje de hasta 1500 caracteres."
		}
		// Client-generated operation key makes a retry idempotent.
		id := stringValue(v["id"])
		if id == "" || len(id) > 100 {
			return "Identificador de mensaje inválido."
		}
		for _, raw := range arrayField(d, "mentoring602") {
			n := legacyMap(raw)
			if stringValue(n["id"]) == id {
				if n["actorId"] == actor["id"] && n["ownerId"] == owner["id"] && n["text"] == message {
					return ""
				}
				return "El mensaje ya existe."
			}
		}
		d["mentoring602"] = append(arrayField(d, "mentoring602"), map[string]any{"id": id, "ownerId": owner["id"], "actorId": actor["id"], "actorName": actor["fullName"], "text": message, "createdAt": now})
	case "revoke602":
		if personal || !musicPermit602(db, actor, "certificates.revoke") || !musicManager550(legacyRole(actor)) {
			return "No puedes revocar certificados."
		}
		c := legacyMap(mapField(d, "certificates")[stringValue(v["id"])])
		if c == nil || !musicScope550(db, actor, legacyUser(db, stringValue(c["ownerId"]))) {
			return "Certificado fuera de tu alcance."
		}
		c["revokedAt"] = now
		c["revokedBy"] = actor["id"]
	}
	return ""
}
func musicActionKey602(path, action string, v map[string]any) string {
	switch path {
	case "/api/music/contact":
		return "whatsapp"
	case "/api/accounts/tutor":
		return "tutors.create"
	case "/api/accounts/import":
		return "people.create"
	case "/api/accounts/music":
		return map[string]string{"create_aspirant": "people.create", "profile": "people.edit", "assign_tutor": "people.assign", "remove": "people.remove", "remove_empty": "people.remove", "assign_task": "tasks.edit"}[action]
	case "/api/music/module":
		return map[string]string{"event": "events.edit", "delete_event": "events.delete", "enroll": "events.enroll", "unenroll": "events.enroll", "external_enroll": "events.enroll", "external_remove": "events.enroll", "attendance": "events.attendance", "external_attendance": "events.attendance", "certificate": "certificates.issue", "certificates": "certificates.view", "template": "certificates.templates", "approve": "reports.approve", "recommend": "reports.approve", "correct": "reports.correct", "mentor602": "mentor.write", "revoke602": "certificates.revoke"}[action]
	case "/api/reviews/item":
		return "reports.view"
	case "/api/reviews/submit":
		return "reports.submit"
	case "/api/reviews/decision":
		if action == "approve" || action == "approved" {
			return "reports.approve"
		}
		return "reports.correct"
	case "/api/progress/reset":
		return "reports.correct"
	case "/api/google/upload", "/api/reports/drive/attempt":
		return "drive.upload"
	case "/api/google/file/trash":
		return "drive.delete"
	case "/api/google/file/status", "/api/reports/drive/verify", "/api/reports/drive/reopen":
		return "drive.verify"
	case "/api/google/file/share":
		return "drive.view"
	case "/api/indications":
		return "content.notices"
	}
	return ""
}
func (s *Server) musicGate602(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		v := map[string]any{}
		if r.Method == "POST" {
			raw, err := io.ReadAll(io.LimitReader(r.Body, 80<<20))
			if err != nil {
				legacyError(w, 400, "Datos inválidos.", nil)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(raw))
			_ = json.Unmarshal(raw, &v)
		}
		action := stringValue(v["action"])
		if action == "" {
			action = stringValue(v["decision"])
		}
		key := musicActionKey602(path, action, v)
		applies := strings.HasPrefix(path, "/api/music/") || path == "/api/accounts/music" || musicRequest553(r) || stringValue(v["programId"]) == musicProgram553 || r.URL.Query().Get("programId") == musicProgram553
		if key != "" && applies {
			s.mu.Lock()
			st, err := s.loadLegacy()
			if err != nil {
				s.mu.Unlock()
				legacyError(w, 500, "No se pudieron validar los permisos.", nil)
				return
			}
			id := stringValue(v["actorId"])
			if id == "" {
				id = r.URL.Query().Get("actorId")
			}
			actor, ok := s.musicActor550(r, legacyMap(st.State.Database), id)
			db := legacyMap(st.State.Database)
			allowed := ok && musicPermit602(db, actor, key)
			if strings.HasPrefix(action, "external_") {
				allowed = allowed && musicPermit602(db, actor, "events.external")
			}
			if strings.HasPrefix(key, "drive.") {
				allowed = allowed && mapField(musicData550(db), "interface598")["driveVisible"] == true
			}
			s.mu.Unlock()
			if !allowed {
				legacyError(w, 403, "Esta función de Música está desactivada para tu cuenta.", nil)
				return
			}
		}
		next(w, r)
	}
}

// Preserve old clients while preventing a disabled operation via bulk synchronization.
func musicSyncGuard602(db, actor, candidate map[string]any) string {
	if actor == nil || candidate == nil {
		return ""
	}
	if reqSub := legacyMap(mapField(legacyProgressForRequirement(candidate, stringValue(actor["id"]), musicProgram553, false), "submissions")[ministryRequirement610]); !musicLinkedSubmissionGuard610(db, stringValue(actor["id"]), ministryRequirement610, reqSub) {
		return "El requisito vinculado utiliza el certificado oficial del ministerio. No se puede sustituir por otro archivo."
	}
	if !musicPermit602(db, actor, "tasks.selfPlan") {
		for _, raw := range arrayField(candidate, "tasks") {
			task := legacyMap(raw)
			if !musicItem553(task) || task["selfPlanned"] != true || stringValue(task["createdBy"]) != stringValue(actor["id"]) {
				continue
			}
			var old map[string]any
			for _, previous := range arrayField(db, "tasks") {
				item := legacyMap(previous)
				if item["id"] == task["id"] {
					old = item
					break
				}
			}
			if !reflect.DeepEqual(task, old) {
				return "La planificación propia está desactivada."
			}
		}
	}
	for _, pair := range [][2]string{{"tasks", "tasks.edit"}, {"materials", "content.materials"}, {"indicationRules", "content.notices"}} {
		if musicPermit602(db, actor, pair[1]) {
			continue
		}
		index := func(items []any) map[string]any {
			out := map[string]any{}
			for _, raw := range items {
				m := legacyMap(raw)
				if musicItem553(m) {
					out[stringValue(m["id"])] = m
				}
			}
			return out
		}
		if _, exists := candidate[pair[0]]; exists && !reflect.DeepEqual(index(arrayField(db, pair[0])), index(arrayField(candidate, pair[0]))) {
			return "Esta función de Música fue desactivada por el Administrador."
		}
	}
	for id, raw := range mapField(candidate, "programProgress") {
		incoming := legacyMap(legacyMap(raw)[musicProgram553])
		if incoming == nil {
			continue
		}
		old := legacyProgressForRequirement(db, id, musicProgram553, false)
		for rid, rawStatus := range mapField(incoming, "status") {
			status := stringValue(rawStatus)
			if status == stringValue(mapField(old, "status")[rid]) {
				continue
			}
			key := "reports.submit"
			if status == "approved" {
				key = "reports.approve"
			} else if status == "needs_changes" || stringValue(mapField(old, "status")[rid]) == "approved" {
				key = "reports.correct"
			}
			if !musicPermit602(db, actor, key) {
				return "La operación de informes está desactivada para tu cuenta."
			}
		}
	}
	for _, raw := range legacyUsers(candidate) {
		u := legacyMap(raw)
		old := legacyUser(db, stringValue(u["id"]))
		if old == nil {
			if !musicPermit602(db, actor, "people.create") && musicAllowed553(candidate, u) {
				return "Crear cuentas de Música está desactivado."
			}
			continue
		}
		if !musicAllowed553(db, old) {
			continue
		}
		for _, k := range []string{"fullName", "phone", "email", "region", "address", "church"} {
			if value, ok := u[k]; ok && !reflect.DeepEqual(value, old[k]) && !musicPermit602(db, actor, "people.edit") {
				return "Editar cuentas de Música está desactivado."
			}
		}
	}
	return ""
}
