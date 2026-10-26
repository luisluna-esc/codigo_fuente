package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html"
	"image"
	_ "image/jpeg"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var musicRequirements550 = []string{"LM-PRE-1", "LM-PRE-2", "LM-PRE-3", "LM-PRE-4", "LM-PRE-5", "LM-MIS-1", "LM-MIS-2", "LM-MIS-3", "LM-MIS-4", "LM-MIS-5", "LM-MIS-6", "LM-REL-1", "LM-REL-2", "LM-REL-3", "LM-REL-4", "LM-REL-5", "LM-REL-6", "LM-NUT-1", "LM-NUT-2", "LM-NUT-3", "LM-NUT-4", "LM-NUT-5", "LM-NUT-6", "LM-NUT-7", "LM-NUT-8", "LM-NUT-9", "LM-TEM-1", "LM-TEM-2", "LM-TEM-3", "LM-TEM-4", "LM-TEM-5"}
var certificateTypes550 = map[string]string{"egw_reading": "Lectura libro Elena White", "year": "Año bíblico", "reavivados": "Reavivados por su Palabra", "forty_days": "40 días", "forty_dawns": "40 madrugadas", "bible_plan": "Bible Plan", "ministry": "Registro del Ministerio de la Música", "leadership": "10 horas de liderazgo", "beliefs": "Creencias fundamentales", "final": "Líder de Música"}

func musicManager550(role string) bool {
	return role == "admin" || role == "departamental" || role == "secretaria" || role == "coordinador_musica"
}
func musicEventManager551(role string) bool { return musicPolicies560[role].Events }
func musicEventKind551(event map[string]any) string {
	kind := stringValue(event["certificateKind"])
	if kind == "" {
		kind = stringValue(event["category"])
	}
	if kind == "leadership" || kind == "beliefs" || kind == "egw_reading" {
		return kind
	}
	return "general"
}
func musicAnnualChoice551(db map[string]any, owner string) string {
	return stringValue(musicDevotion550(db, owner, "LM-NUT-1")["choice"])
}
func musicAnnualReviewReady551(db map[string]any, owner string, submitted map[string]any) bool {
	if !musicAssigned550(db, owner) {
		return false
	}
	// A verified certificate is independent of the optional reading selection.
	// submitReview validates ownership, scope and the actual PDF bytes.
	if reviewSubmissionIsDocumentMode(submitted) {
		return strings.TrimSpace(stringValue(submitted["primaryDocumentId"])) != ""
	}
	choice := stringValue(legacyMap(submitted["bibleReadings"])["choice"])
	if choice == "" {
		choice = musicAnnualChoice551(db, owner)
	}
	if choice != "reavivados" && choice != "bible_plan" {
		return false
	}
	check := legacyMap(legacyMap(mapField(musicData550(db), "completions")[owner])[choice])
	validator := legacyUser(db, stringValue(check["validatedBy"]))
	role := legacyRole(validator)
	return check["concluded"] == true && stringValue(check["at"]) != "" && (role == "admin" || role == "departamental" || role == "secretaria")
}
func musicRecommend550(role string) bool {
	return musicPolicies560[role].Recommend
}
func musicApprove550(role string) bool {
	return role == "departamental" || role == "admin"
}
func musicAssigned550(db map[string]any, id string) bool {
	if legacyUser(db, id)["musicAccessRemoved"] == true {
		return false
	}
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if stringValue(m["id"]) == "music_ministry" && m["active"] != false && containsString(arrayField(m, "assignedUserIds"), id) {
			return true
		}
	}
	return false
}
func musicParticipant565(db map[string]any, user map[string]any) bool {
	return legacyActive(user) && user["musicAccessRemoved"] != true && (musicAssigned550(db, stringValue(user["id"])) || managementProgram555(db, user, musicProgram553))
}
func musicScope550(db map[string]any, actor, owner map[string]any) bool {
	if actor == nil || owner == nil || !legacyActive(actor) || actor["musicAccessRemoved"] == true || !legacyActive(owner) || !musicAssigned550(db, stringValue(owner["id"])) {
		return false
	}
	role := legacyRole(actor)
	if stringValue(actor["id"]) == stringValue(owner["id"]) {
		return true
	}
	if !musicPermit602(db, actor, "people.view") {
		return false
	}
	if musicManager550(role) {
		return managementProgram555(db, actor, musicProgram553)
	}
	if role == "regional" {
		return managementProgram555(db, actor, musicProgram553) && (legacyTutorAssigned(owner, stringValue(actor["id"]), musicProgram553) || legacyTutorForProgram(owner, musicProgram553) == "" && stringValue(owner["createdBy"]) == stringValue(actor["id"]))
	}
	if role == "tutor" || role == "administrativo" {
		return managementProgram555(db, actor, musicProgram553) && (legacyTutorAssigned(owner, stringValue(actor["id"]), "music_ministry") || role == "administrativo" && stringValue(owner["createdBy"]) == stringValue(actor["id"]))
	}
	return role == "aspirant" && stringValue(actor["id"]) == stringValue(owner["id"])
}
func musicComplete550(db map[string]any, owner string) bool {
	p := legacyProgressForRequirement(db, owner, "music_ministry", false)
	statuses := mapField(p, "status")
	for _, id := range musicRequirements550 {
		if stringValue(statuses[id]) != "approved" {
			return false
		}
	}
	return true
}
func musicData550(db map[string]any) map[string]any { return ensureMapField(db, "music550") }

// Ordinary full snapshots cannot replace this independently versioned resource.
// Explicit backup restoration still restores it through the restore endpoint.
func preserveMusicResource550(current, incoming any) any {
	if current == nil {
		return incoming
	}
	target := legacyMap(cloneAny(incoming))
	if target == nil {
		return incoming
	}
	if resource, exists := legacyMap(current)["music550"]; exists {
		target["music550"] = cloneAny(resource)
	} else {
		delete(target, "music550")
	}
	return target
}
func musicDevotion550(db map[string]any, owner, req string) map[string]any {
	return legacyMap(legacyMap(mapField(legacyProgressForRequirement(db, owner, "music_ministry", false), "submissions")[req])["bibleReadings"])
}
func musicFortyComplete550(db map[string]any, owner string) bool {
	p := legacyProgressForRequirement(db, owner, musicProgram553, false)
	sub := legacyMap(mapField(p, "submissions")["LM-TEM-1"])
	if stringValue(mapField(p, "status")["LM-TEM-1"]) == "approved" && strings.TrimSpace(stringValue(sub["primaryDocumentId"])) != "" && reviewSubmissionIsDocumentMode(sub) {
		return true
	}
	entries := mapField(musicDevotion550(db, owner, "LM-TEM-1"), "entries")
	dates := map[string]bool{}
	for i := 1; i <= 40; i++ {
		e := legacyMap(entries[strconv.Itoa(i)])
		if e["completed"] != true {
			return false
		}
		for _, key := range []string{"date", "reference", "favoriteVerses", "impactStory", "lifeLesson", "theme"} {
			if strings.TrimSpace(stringValue(e[key])) == "" {
				return false
			}
		}
		if strings.TrimSpace(stringValue(e["songTheme"])) == "" && (strings.TrimSpace(stringValue(e["song"])) == "" || strings.TrimSpace(stringValue(e["songRelation"])) == "") {
			return false
		}
		date := stringValue(e["date"])
		parsed, err := time.Parse("2006-01-02", date)
		if err != nil || parsed.After(time.Now()) || dates[date] {
			return false
		}
		dates[date] = true
	}
	return true
}
func musicYearComplete550(db map[string]any, owner string) bool {
	d := musicDevotion550(db, owner, "LM-NUT-1")
	if stringValue(d["choice"]) != "reavivados" {
		return false
	}
	start, err := time.Parse("2006-01-02", stringValue(d["startDate"]))
	if err != nil {
		return false
	}
	end := start.AddDate(1, 0, 0)
	if start.Month() == time.February && start.Day() == 29 {
		end = time.Date(end.Year(), time.February, 28, 0, 0, 0, 0, time.UTC)
	}
	entries := mapField(d, "entries")
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		if day.After(time.Now()) {
			return false
		}
		e := legacyMap(entries[day.Format("2006-01-02")])
		if e["completed"] != true {
			return false
		}
		for _, k := range []string{"date", "favoriteVerses", "impactStory", "lifeLesson"} {
			if strings.TrimSpace(stringValue(e[k])) == "" {
				return false
			}
		}
	}
	return true
}
func musicEligible550(db map[string]any, owner, kind, eventID string) bool {
	data := musicData550(db)
	checks := legacyMap(mapField(data, "completions")[owner])
	switch kind {
	case "ministry":
		return validMusicRegistration548(db, owner, legacyUser(db, owner)["musicMinistryRegistration"])
	case "year", "reavivados":
		return musicAnnualChoice551(db, owner) != "bible_plan" && (musicYearComplete550(db, owner) || legacyMap(checks["reavivados"])["concluded"] == true)
	case "bible_plan":
		return musicAnnualChoice551(db, owner) != "reavivados" && legacyMap(checks["bible_plan"])["concluded"] == true
	case "forty_days", "forty_dawns":
		return musicFortyComplete550(db, owner)
	case "final":
		return musicComplete550(db, owner) && musicApproval618(db, owner) != ""
	case "event":
		e := legacyMap(mapField(data, "events")[eventID])
		if !musicFrom2026_618(stringValue(e["date"])) {
			return false
		}
		return legacyMap(mapField(e, "participants")[owner])["attended"] == true && stringValue(e["templateId610"]) != ""
	case "leadership", "beliefs", "egw_reading":
		e := legacyMap(mapField(data, "events")[eventID])
		attendance := legacyMap(mapField(e, "participants")[owner])
		if !musicFrom2026_618(stringValue(e["date"])) {
			return false
		}
		if attendance["attended"] == true && musicEventKind551(e) == kind && stringValue(e["templateId610"]) != "" {
			return true
		}
		if attendance["attended"] != true || musicEventKind551(e) != kind || stringValue(e["date"]) > time.Now().Format("2006-01-02") {
			return false
		}
		if kind == "egw_reading" {
			return strings.TrimSpace(stringValue(e["bookTitle"])) != ""
		}
		return kind != "leadership" || numericValue(e["hours"]) >= 10
	}
	return false
}
func musicHasCertificate550(db map[string]any, owner string) bool {
	for kind := range certificateTypes550 {
		if eventCertificate557(kind) {
			for eventID := range mapField(musicData550(db), "events") {
				if musicEligible550(db, owner, kind, eventID) {
					return true
				}
			}
		} else if musicEligible550(db, owner, kind, "") {
			return true
		}
	}
	return false
}
func (s *Server) musicActor550(r *http.Request, db map[string]any, requested string) (map[string]any, bool) {
	if sess, ok := s.remoteSessionFromRequest(r); ok {
		if requested != "" && requested != sess.UserID {
			return nil, false
		}
		a := legacyUser(db, sess.UserID)
		if personalRequest555(r) && musicAssigned550(db, sess.UserID) {
			a = legacyMap(cloneAny(a))
			a["role"] = "aspirant"
		}
		return a, legacyActive(a) && (legacyRole(a) == canonicalRole548(sess.Role) || personalRequest555(r) && legacyRole(a) == "aspirant")
	}
	role, ok := s.auth(r)
	if !ok || (role != "admin" && !personalRequest555(r)) {
		return nil, false
	}
	a := legacyUser(db, requested)
	if personalRequest555(r) && musicAssigned550(db, requested) {
		a = legacyMap(cloneAny(a))
		a["role"] = "aspirant"
		return a, legacyActive(a)
	}
	return a, legacyActive(a) && legacyRole(a) == "admin"
}
func musicView550(db map[string]any, actor map[string]any) map[string]any {
	d := legacyMap(cloneAny(musicData550(db)))
	if d == nil {
		d = map[string]any{}
	}
	allowed := map[string]bool{}
	certificateOwners := map[string]bool{}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if musicScope550(db, actor, u) {
			allowed[stringValue(u["id"])] = true
		}
		if allowed[stringValue(u["id"])] || musicEventManager551(legacyRole(actor)) && legacyActive(u) {
			certificateOwners[stringValue(u["id"])] = true
		}
	}
	for _, field := range []string{"workflow", "completions", "investitures"} {
		out := map[string]any{}
		for id, v := range mapField(d, field) {
			if allowed[id] {
				out[id] = v
			}
		}
		d[field] = out
	}
	events := map[string]any{}
	for eventID, raw := range mapField(d, "events") {
		e := legacyMap(raw)
		participants := map[string]any{}
		for id, v := range mapField(e, "participants") {
			if allowed[id] || musicRoster610(legacyRole(actor)) {
				participants[id] = v
			}
		}
		e["participants"] = participants
		if !musicRoster610(legacyRole(actor)) {
			e["externalParticipants"] = map[string]any{}
		}
		events[eventID] = e
	}
	d["events"] = events
	certs := map[string]any{}
	for id, v := range mapField(d, "certificates") {
		c := legacyMap(v)
		if stringValue(c["revokedAt"]) == "" && certificateOwners[stringValue(c["ownerId"])] && musicEligible550(db, stringValue(c["ownerId"]), stringValue(c["kind"]), stringValue(c["eventId"])) {
			certs[id] = v
		}
	}
	d["certificates"] = certs
	d["canWhatsapp598"] = musicWhatsapp598(db, actor)
	if legacyRole(actor) != "admin" {
		ids := []any{}
		if d["canWhatsapp598"] == true {
			ids = append(ids, actor["id"])
		}
		d["interface598"] = map[string]any{"whatsappUsers": ids}
	}
	musicProjection602(db, actor, d)
	return d
}

// Narrow operations avoid exposing generic database writes to staff roles.
func (s *Server) music550(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	var req struct {
		ActorID           string         `json:"actorId"`
		OwnerID           string         `json:"ownerId"`
		Action            string         `json:"action"`
		BaseUpdatedAt     string         `json:"baseUpdatedAt"`
		RestoreGeneration int64          `json:"restoreGeneration"`
		Value             map[string]any `json:"value"`
	}
	if r.Method == http.MethodGet {
		req.ActorID = r.URL.Query().Get("actorId")
	} else if err := readJSONLimited(r, &req, 5<<20); err != nil {
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
	db := legacyMap(cloneAny(st.State.Database))
	actor, ok := s.musicActor550(r, db, req.ActorID)
	if !ok {
		legacyError(w, 403, "Sesión no autorizada.", nil)
		return
	}
	if legacyRole(actor) != "admin" && !managementProgram555(db, actor, musicProgram553) && musicAssigned550(db, stringValue(actor["id"])) {
		actor = legacyMap(cloneAny(actor))
		actor["role"] = "aspirant"
	}
	role := legacyRole(actor)
	if !musicAssigned550(db, stringValue(actor["id"])) && !managementProgram555(db, actor, musicProgram553) {
		legacyError(w, 403, "Música no está habilitada para tu cuenta.", nil)
		return
	}
	if !(musicManager550(role) || musicRecommend550(role) || role == "tutor" || musicAssigned550(db, stringValue(actor["id"]))) {
		legacyError(w, 403, "Módulo de Música no asignado.", nil)
		return
	}
	data := musicData550(db)
	if r.Method == http.MethodGet {
		jsonOut(w, 200, map[string]any{"ok": true, "data": musicView550(db, actor)})
		return
	}
	generation := legacyCanonicalRestoreGeneration(st.State)
	if req.RestoreGeneration != generation && !(req.RestoreGeneration == 0 && generation == 1) {
		legacyError(w, 409, "Actualiza los datos después de la restauración.", nil)
		return
	}
	if req.BaseUpdatedAt != stringValue(data["updatedAt"]) {
		legacyError(w, 409, "Música cambió en otro equipo. Actualiza; tu formulario se conserva.", map[string]any{"code": "MUSIC_CONFLICT"})
		return
	}
	value := req.Value
	if value == nil {
		value = map[string]any{}
	}
	owner := legacyUser(db, req.OwnerID)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	eventParticipantAction := req.Action == "enroll" || req.Action == "attendance" || req.Action == "unenroll"
	certificateAction := req.Action == "certificates" || req.Action == "certificate"
	eventCertificateAction := certificateAction && musicEventManager551(role) && legacyActive(owner)
	if req.Action == "certificate" {
		kind, eventID := stringValue(value["kind"]), stringValue(value["eventId"])
		eventCertificateAction = eventCertificateAction && certificateTypes550[kind] != "" && musicEligible550(db, req.OwnerID, kind, eventID)
	} else if req.Action == "certificates" {
		eventCertificateAction = eventCertificateAction && musicHasCertificate550(db, req.OwnerID)
	}
	if req.OwnerID != "" && !musicScope550(db, actor, owner) && !(eventParticipantAction && musicEventManager551(role) && owner != nil) && !eventCertificateAction {
		legacyError(w, 403, "Miembro aspirante fuera de tu alcance.", nil)
		return
	}
	if key := musicActionKey602("/api/music/module", req.Action, req.Value); key != "" && !musicPermit602(db, actor, key) {
		legacyError(w, 403, "Función desactivada.", nil)
		return
	}
	fail := func(message string) { legacyError(w, 400, message, nil) }
	switch req.Action {
	case "permissions602", "mentor602", "revoke602":
		notesBefore602 := len(arrayField(data, "mentoring602"))
		if reason := musicMutation602(db, actor, owner, req.Action, value, now, personalRequest555(r)); reason != "" {
			legacyError(w, 403, reason, nil)
			return
		}
		if req.Action == "mentor602" && len(arrayField(data, "mentoring602")) > notesBefore602 && musicPermit602(db, owner, "mentor.notify") {
			addLegacyNotification(&st, legacyNotification{UserID: req.OwnerID, Type: "music_mentoring", Title: "Acompañamiento de tu tutor", Message: "Tu tutor ha dejado un mensaje en Acompañamiento.", ProgramID: musicProgram553, ActorID: req.ActorID, ActorName: stringValue(actor["fullName"]), CreatedAt: now, SourceOperationID: stringValue(value["id"])})
		}
	case "interface_settings":
		if role != "admin" || personalRequest555(r) {
			legacyError(w, 403, "Solo el Administrador puede modificar estas opciones.", nil)
			return
		}
		data["interface598"] = map[string]any{"whatsappUsers": []any{}}
	case "delete_event":
		if !(role == "admin" || role == "departamental" || role == "secretaria" || role == "coordinador_musica") {
			legacyError(w, 403, "Tu función no permite eliminar eventos.", nil)
			return
		}
		id := strings.TrimSpace(stringValue(value["eventId"]))
		events := mapField(data, "events")
		if id == "" || legacyMap(events[id]) == nil {
			fail("Evento no encontrado.")
			return
		}
		musicRevokeEvent610(data, id, "", "Evento eliminado", now)
		delete(events, id)
	case "event":
		if !musicEventManager551(role) {
			legacyError(w, 403, "No puedes administrar eventos.", nil)
			return
		}
		if !musicFrom2026_618(stringValue(value["date"])) {
			fail("La fecha del evento debe corresponder a la gestión 2026 o posterior.")
			return
		}
		for _, key := range []string{"name", "date", "place", "time", "category"} {
			if strings.TrimSpace(stringValue(value[key])) == "" {
				fail("Completa nombre, fecha, lugar, horario y categoría.")
				return
			}
		}
		if _, err := time.Parse("2006-01-02", stringValue(value["date"])); err != nil {
			fail("Fecha inválida.")
			return
		}
		if _, err := time.Parse("15:04", stringValue(value["time"])); err != nil {
			fail("Horario inválido.")
			return
		}
		category := stringValue(value["category"])
		if category != "music_ministry" && category != "general" && category != "leadership" && category != "beliefs" && category != "egw_reading" {
			fail("Categoría inválida.")
			return
		}
		certificateKind := musicEventKind551(value)
		if raw := stringValue(value["certificateKind"]); raw != "" && raw != "general" && raw != "leadership" && raw != "beliefs" && raw != "egw_reading" {
			fail("Certificado asociado inválido.")
			return
		}
		if certificateKind == "egw_reading" && strings.TrimSpace(stringValue(value["bookTitle"])) == "" {
			fail("Indica el libro de Elena White asociado al evento.")
			return
		}
		hours := numericValue(value["hours"])
		if hours < 0 || hours > 1000 {
			fail("Horas inválidas.")
			return
		}
		templateID := strings.TrimSpace(stringValue(value["templateId610"]))
		if templateID != "" && mapField(data, "templates")[templateID] == nil {
			fail("Selecciona una plantilla existente.")
			return
		}
		id := strings.TrimSpace(stringValue(value["id"]))
		events := ensureMapField(data, "events")
		old := legacyMap(events[id])
		if id == "" {
			id = "event_" + randomMusicID550()
		}
		if old != nil && (len(mapField(old, "participants")) > 0 || len(mapField(old, "externalParticipants")) > 0) && (musicEventKind551(old) != certificateKind || numericValue(old["hours"]) != hours || strings.TrimSpace(stringValue(old["bookTitle"])) != strings.TrimSpace(stringValue(value["bookTitle"]))) {
			fail("Con participantes inscritos no puedes cambiar la categoría ni las horas.")
			return
		}
		e := map[string]any{"id": id, "hours": hours, "category": "music_ministry", "certificateKind": certificateKind, "templateId610": templateID, "participants": cloneAny(mapField(old, "participants")), "externalParticipants": cloneAny(mapField(old, "externalParticipants")), "updatedAt": now}
		for _, key := range []string{"name", "date", "place", "time", "notes", "description", "bookTitle"} {
			text := strings.TrimSpace(stringValue(value[key]))
			if len(text) > 1000 {
				fail("Texto demasiado largo.")
				return
			}
			e[key] = text
		}
		if stringValue(e["description"]) == "" {
			e["description"] = e["notes"]
		}
		e["notes"] = e["description"]
		events[id] = e
	case "enroll", "attendance", "unenroll":
		if !musicEventManager551(role) || owner == nil {
			legacyError(w, 403, "No puedes inscribir participantes.", nil)
			return
		}
		e := legacyMap(mapField(data, "events")[stringValue(value["eventId"])])
		if e == nil {
			fail("Evento no encontrado.")
			return
		}
		participants := ensureMapField(e, "participants")
		p := legacyMap(participants[req.OwnerID])
		if req.Action == "enroll" {
			if !musicParticipant565(db, owner) {
				legacyError(w, 403, "Solo puedes inscribir cuentas activas de Música.", nil)
				return
			}
			if p != nil {
				fail("Ya está inscrito.")
				return
			}
			participants[req.OwnerID] = map[string]any{"enrolledAt": now, "attended": false}
		} else if req.Action == "unenroll" {
			if p == nil {
				fail("Participante no inscrito.")
				return
			}
			musicRevokeEvent610(data, stringValue(value["eventId"]), req.OwnerID, "Inscripción finalizada", now)
			delete(participants, req.OwnerID)
		} else {
			if p == nil {
				fail("Inscribe al participante primero.")
				return
			}
			if value["attended"] != true {
				musicRevokeEvent610(data, stringValue(value["eventId"]), req.OwnerID, "Asistencia retirada", now)
			}
			p["attended"] = value["attended"] == true
			p["checkedAt"] = now
			p["checkedBy"] = req.ActorID
		}
		e["updatedAt"] = now
	case "external_enroll", "external_attendance", "external_remove":
		if !musicEventManager551(role) {
			legacyError(w, 403, "No puedes administrar participantes externos.", nil)
			return
		}
		e := legacyMap(mapField(data, "events")[stringValue(value["eventId"])])
		if e == nil {
			fail("Evento no encontrado.")
			return
		}
		guests := ensureMapField(e, "externalParticipants")
		guestID := strings.TrimSpace(stringValue(value["participantId"]))
		guest := legacyMap(guests[guestID])
		switch req.Action {
		case "external_enroll":
			name := strings.Join(strings.Fields(stringValue(value["fullName"])), " ")
			if len([]rune(name)) < 3 || len([]rune(name)) > 120 {
				fail("Escribe el nombre completo del participante (3 a 120 caracteres).")
				return
			}
			for _, existing := range guests {
				if strings.EqualFold(strings.TrimSpace(stringValue(legacyMap(existing)["fullName"])), name) {
					fail("Esta persona ya está en la lista de participantes externos.")
					return
				}
			}
			guestID = "guest_" + randomMusicID550()
			guests[guestID] = map[string]any{"id": guestID, "fullName": name, "enrolledAt": now, "attended": false}
		case "external_attendance":
			if guest == nil {
				fail("Participante externo no encontrado.")
				return
			}
			guest["attended"] = value["attended"] == true
			guest["checkedAt"] = now
			guest["checkedBy"] = req.ActorID
		case "external_remove":
			if guest == nil {
				fail("Participante externo no encontrado.")
				return
			}
			delete(guests, guestID)
		}
		e["updatedAt"] = now
	case "conclude":
		if req.OwnerID == stringValue(actor["id"]) {
			legacyError(w, 403, "Otra autoridad debe validar tu formación personal.", nil)
			return
		}
		if !(role == "admin" || role == "departamental" || role == "secretaria") || owner == nil {
			legacyError(w, 403, "Solo Administración, Departamental y Secretaría validan la conclusión.", nil)
			return
		}
		kind := stringValue(value["kind"])
		if kind != "reavivados" && kind != "bible_plan" {
			fail("Modalidad inválida.")
			return
		}
		if musicAnnualChoice551(db, req.OwnerID) != kind {
			fail("Solo puedes validar la modalidad elegida por el aspirante.")
			return
		}
		progress := reviewProgressContainer(db, req.OwnerID, "music_ministry")
		status := ensureMapField(progress, "status")
		if stringValue(status["LM-NUT-1"]) == "approved" || stringValue(status["LM-NUT-1"]) == "in_review" {
			fail("El requisito ya fue revisado. Solicita corrección para cambiarlo.")
			return
		}
		checks := ensureMapField(ensureMapField(data, "completions"), req.OwnerID)
		checks[kind] = map[string]any{"concluded": true, "validatedBy": req.ActorID, "at": now}
		sub := ensureMapField(ensureMapField(progress, "submissions"), "LM-NUT-1")
		if kind == "bible_plan" || strings.TrimSpace(stringValue(sub["text"])) == "" {
			sub["text"] = "Lectura anual concluida: " + kind + "\nValidada por: " + stringValue(actor["fullName"]) + "\nFecha de conclusión: " + now
			sub["richText"] = "<p>" + strings.ReplaceAll(html.EscapeString(stringValue(sub["text"])), "\n", "</p><p>") + "</p>"
		}
		sub["submissionMode"] = "written"
		sub["deliveryMode"] = "written"
		sub["documentMode"] = false
		sub["draftOnly"] = false
		sub["completionDate"] = now[:10]
		sub["updatedAt"] = now
		status["LM-NUT-1"] = "in_progress"
		progress["updatedAt"] = now
	case "recommend", "approve", "correct":
		if req.OwnerID == stringValue(actor["id"]) {
			legacyError(w, 403, "Otra autoridad debe revisar tu tarjeta personal.", nil)
			return
		}
		if owner == nil {
			fail("Selecciona un miembro aspirante.")
			return
		}
		workflow := ensureMapField(ensureMapField(data, "workflow"), req.OwnerID)
		if req.Action == "correct" {
			if !musicReview560(db, actor, owner) {
				legacyError(w, 403, "No puedes devolver requisitos.", nil)
				return
			}
			id := stringValue(value["requirementId"])
			if !containsString(stringSliceToAny550(musicRequirements550), id) || strings.TrimSpace(stringValue(value["feedback"])) == "" {
				fail("Selecciona un requisito y explica la corrección.")
				return
			}
			p := reviewProgressContainer(db, req.OwnerID, "music_ministry")
			status := ensureMapField(p, "status")
			if stringValue(status[id]) != "approved" && stringValue(status[id]) != "in_review" {
				fail("Este requisito no está disponible para corrección.")
				return
			}
			status[id] = "needs_changes"
			sub := ensureMapField(ensureMapField(p, "submissions"), id)
			sub["feedback"] = strings.TrimSpace(stringValue(value["feedback"]))
			sub["updatedAt"] = now
			delete(sub, "approval")
			p["updatedAt"] = now
			workflow["state"] = "correction"
			workflow["requirementId"] = id
			workflow["changedBy"] = req.ActorID
			workflow["at"] = now
		} else {
			if !musicComplete550(db, req.OwnerID) {
				fail("Todos los requisitos deben estar revisados (100%).")
				return
			}
			if req.Action == "recommend" {
				if !musicRecommend550(role) {
					legacyError(w, 403, "Tu función no incluye recomendar.", nil)
					return
				}
				if stringValue(workflow["state"]) == "approved" {
					fail("El expediente ya fue aprobado.")
					return
				}
				workflow["state"] = "recommended"
				workflow["recommendedBy"] = req.ActorID
				workflow["recommendedAt"] = now
			} else {
				if !musicFinal560(db, actor, owner) {
					legacyError(w, 403, "Tu función no incluye aprobar.", nil)
					return
				}
				if stringValue(workflow["state"]) != "recommended" {
					fail("Primero debe existir una recomendación.")
					return
				}
				workflow["state"] = "approved"
				workflow["approvedBy"] = req.ActorID
				workflow["approvedAt"] = now
			}
		}
	case "investiture":
		if owner == nil || !musicComplete550(db, req.OwnerID) || stringValue(legacyMap(mapField(data, "workflow")[req.OwnerID])["state"]) != "approved" {
			fail("La investidura requiere 100% revisado y aprobación final del Departamental.")
			return
		}
		if !(musicManager550(role) || stringValue(actor["id"]) == req.OwnerID) {
			legacyError(w, 403, "No puedes fijar esta investidura.", nil)
			return
		}
		for _, key := range []string{"place", "date", "time", "investedBy"} {
			if strings.TrimSpace(stringValue(value[key])) == "" {
				fail("Completa lugar, fecha, hora y quién investirá.")
				return
			}
		}
		if _, err := time.Parse("2006-01-02", stringValue(value["date"])); err != nil {
			fail("Fecha inválida.")
			return
		}
		if _, err := time.Parse("15:04", stringValue(value["time"])); err != nil {
			fail("Horario inválido.")
			return
		}
		v := map[string]any{"ownerId": req.OwnerID, "fullName": owner["fullName"], "updatedAt": now}
		for _, key := range []string{"place", "date", "time", "investedBy"} {
			v[key] = strings.TrimSpace(stringValue(value[key]))
		}
		ensureMapField(data, "investitures")[req.OwnerID] = v
	case "template":
		if !musicManager550(role) {
			legacyError(w, 403, "No puedes subir plantillas.", nil)
			return
		}
		kind := stringValue(value["kind"])
		if certificateTypes550[kind] == "" {
			fail("Tipo de certificado inválido.")
			return
		}
		source := stringValue(value["dataUrl"])
		if !(strings.HasPrefix(source, "data:image/png;base64,") || strings.HasPrefix(source, "data:image/jpeg;base64,")) || len(source) > 2800000 {
			fail("Usa una plantilla PNG/JPG horizontal de hasta 2 MB.")
			return
		}
		encoded := strings.SplitN(source, ",", 2)
		raw, err := base64.StdEncoding.DecodeString(encoded[1])
		if err != nil || len(raw) > 2*1024*1024 {
			fail("Imagen inválida o demasiado grande.")
			return
		}
		cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
		if err != nil || (format != "png" && format != "jpeg") || cfg.Width > 6000 || cfg.Height > 6000 {
			fail("Usa una imagen PNG/JPG válida de hasta 6000 píxeles.")
			return
		}
		width, height := float64(cfg.Width), float64(cfg.Height)
		if width <= height || height < 100 {
			fail("La plantilla debe ser horizontal.")
			return
		}
		template := map[string]any{"dataUrl": source, "width": width, "height": height, "updatedAt": now}
		for _, k := range []string{"nameX", "nameY", "titleY", "detailsY", "qrX", "qrY"} {
			n := numericValue(value[k])
			if n < 0 || n > 100 {
				fail("Las posiciones deben estar entre 0 y 100.")
				return
			}
			template[k] = n
		}
		if rawFields, exists := value["fields"]; exists {
			fields, valid := musicTemplateFields592(rawFields)
			if !valid {
				fail("Revisa los textos, fuentes, tamaños y posiciones de la plantilla.")
				return
			}
			template["fields"] = fields
		}
		ensureMapField(data, "templates")[kind] = template
	case "certificates":
		if owner == nil {
			fail("Selecciona un aspirante.")
			return
		}
		changed := false
		issue := func(kind, eventID string) {
			c := legacyMap(mapField(data, "certificates")[req.OwnerID+":"+kind+":"+eventID])
			if c == nil || stringValue(c["fullName"]) != stringValue(owner["fullName"]) {
				changed = true
				musicIssue560(db, owner, req.ActorID, kind, eventID, now)
			}
		}
		for kind := range certificateTypes550 {
			if eventCertificate557(kind) {
				for eventID := range mapField(data, "events") {
					if musicEligible550(db, req.OwnerID, kind, eventID) {
						issue(kind, eventID)
					}
				}
			} else if musicEligible550(db, req.OwnerID, kind, "") {
				issue(kind, "")
			}
		}
		if !changed {
			jsonOut(w, 200, map[string]any{"ok": true, "data": musicView550(db, actor), "serverVersion": st.State.ServerVersion})
			return
		}
	case "certificate":
		kind, eventID := stringValue(value["kind"]), stringValue(value["eventId"])
		if owner == nil || certificateTypes550[kind] == "" || !musicEligible550(db, req.OwnerID, kind, eventID) {
			fail("El certificado aún no está habilitado para este miembro aspirante.")
			return
		}
		if !musicScope550(db, actor, owner) && !eventCertificateAction {
			legacyError(w, 403, "No puedes emitir este certificado.", nil)
			return
		}
		musicIssue560(db, owner, req.ActorID, kind, eventID, now)
	default:
		fail("Operación no reconocida.")
		return
	}
	musicWorkflowNotice556(&st, db, actor, owner, req.Action, now)
	data["updatedAt"] = now
	data["updatedBy"] = req.ActorID
	st.State.Database = legacyPrepareCanonicalDB(db, st.State, "Ministerio de Música actualizado")
	st.State.ServerVersion++
	st.State.UpdatedAt = now
	if err = s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "data": musicView550(db, actor), "serverVersion": st.State.ServerVersion, "progress": cloneAny(legacyProgressForRequirement(db, req.OwnerID, "music_ministry", false))})
}
func randomMusicID550() string {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		panic(err)
	}
	return hex.EncodeToString(raw)
}
func stringSliceToAny550(values []string) []any {
	out := make([]any, len(values))
	for i, v := range values {
		out[i] = v
	}
	return out
}

// Public QR page is intentionally data-only; it never returns reports, contacts or account records.
func (s *Server) musicCertificate550(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet {
		http.Error(w, "method", 405)
		return
	}
	token := r.URL.Query().Get("token")
	if len(token) != 32 {
		http.NotFound(w, r)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		http.Error(w, "No disponible", 503)
		return
	}
	var cert map[string]any
	db := legacyMap(st.State.Database)
	if db == nil {
		http.NotFound(w, r)
		return
	}
	for _, v := range mapField(musicData550(db), "certificates") {
		c := legacyMap(v)
		if stringValue(c["token"]) == token {
			cert = c
			break
		}
	}
	if cert == nil {
		for _, raw := range arrayField(musicData550(db), "certificateHistory610") {
			c := legacyMap(raw)
			if c["token"] == token {
				cert = c
				break
			}
		}
	}
	historicalYear := ""
	if cert == nil {
		for year, raw := range mapField(musicData550(db), "archiveData610") {
			old := legacyMap(legacyMap(raw)["database"])
			for _, raw := range mapField(musicData550(old), "certificates") {
				c := legacyMap(raw)
				if c["token"] == token {
					cert = c
					db = old
					historicalYear = year
					break
				}
			}
			if cert != nil {
				break
			}
		}
	}
	if cert == nil {
		http.NotFound(w, r)
		return
	}
	valid := musicEligible550(db, stringValue(cert["ownerId"]), stringValue(cert["kind"]), stringValue(cert["eventId"]))
	if stringValue(cert["ownerId"]) == "" && stringValue(cert["ministryId"]) != "" {
		m := legacyMap(mapField(musicData550(db), "ministries610")[stringValue(cert["ministryId"])])
		valid = m != nil && m["deletedAt"] == nil
	}
	state := "Vigente"
	if !valid {
		state = "Pendiente de nueva validación"
	}
	if stringValue(cert["revokedAt"]) != "" {
		if strings.HasPrefix(stringValue(cert["revokedReason"]), "Asistencia") || strings.HasPrefix(stringValue(cert["revokedReason"]), "Inscripción") || strings.HasPrefix(stringValue(cert["revokedReason"]), "Evento") {
			state = "Pendiente de nueva validación"
		} else {
			state = "Revocado"
		}
	}
	if historicalYear != "" {
		state = "Historial de la gestión " + historicalYear + " · " + state
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	fmt.Fprintf(w, "<!doctype html><html lang=es><meta name=viewport content='width=device-width,initial-scale=1'><title>Datos del certificado</title><body style='font:18px system-ui;max-width:700px;margin:40px auto;padding:24px'><h1>Certificado · Ministerio de la Música</h1><dl><dt>Nombre completo</dt><dd>%s</dd><dt>Tipo</dt><dd>%s</dd><dt>Fecha de emisión</dt><dd>%s</dd><dt>Iglesia</dt><dd>%s</dd><dt>Misión</dt><dd>%s</dd><dt>Estado</dt><dd>%s</dd><dt>Identificador</dt><dd>%s</dd></dl></body></html>", html.EscapeString(stringValue(cert["fullName"])), html.EscapeString(stringValue(cert["title"])), html.EscapeString(stringValue(cert["issuedAt"])), html.EscapeString(stringValue(cert["church"])), html.EscapeString(stringValue(cert["mission"])), state, token)
}
