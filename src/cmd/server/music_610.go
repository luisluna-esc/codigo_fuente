package main

// Independent ministry records and explicit memberships. All writes are narrow,
// authenticated operations; a browser cannot assign permissions or approve itself.
import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

const ministryRequirement610 = "LM-PRE-5"

type musicRequest610 struct {
	ActorID           string         `json:"actorId"`
	OwnerID           string         `json:"ownerId"`
	Action            string         `json:"action"`
	RestoreGeneration int64          `json:"restoreGeneration"`
	Value             map[string]any `json:"value"`
}

func musicLeaders610(role string) bool {
	return role == "admin" || role == "departamental" || role == "secretaria" || role == "coordinador_musica"
}

func musicRoster610(role string) bool {
	return musicLeaders610(role) || role == "administrativo" || role == "regional"
}

func musicMinistryPermit610(db, actor map[string]any, key string) bool {
	if actor == nil || !legacyActive(actor) {
		return false
	}
	role := legacyRole(actor)
	if role == "admin" {
		return true
	}
	if key == "ministries.view" || key == "ministries.export" {
		return musicAllowed553(db, actor)
	}
	base := musicLeaders610(role) || (role == "administrativo" && (key == "ministries.create" || key == "ministries.link" || key == "ministries.unlink"))
	return base && managementProgram555(db, actor, musicProgram553) && musicPermit602(db, actor, key)
}

func ministrySnapshot610(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := legacyMap(cloneAny(m))
	delete(out, "history")
	delete(out, "photo") // the current photograph is stored once, not in every audit entry
	return out
}

func ministryHistory610(m, actor map[string]any, action, at string) {
	m["history"] = append(arrayField(m, "history"), map[string]any{
		"action": action, "at": at, "actorId": actor["id"], "actorName": actor["fullName"], "snapshot": ministrySnapshot610(m),
	})
}

func musicEnsure610(db map[string]any) bool {
	d := musicData550(db)
	changed := d["schema610"] != true
	if changed {
		d["schema610"] = true
		ministries := ensureMapField(d, "ministries610")
		links := ensureMapField(d, "ministryLinks610")
		for _, raw := range legacyUsers(db) {
			u := legacyMap(raw)
			reg := legacyMap(u["musicMinistryRegistration"])
			if strings.TrimSpace(stringValue(reg["name"])) == "" {
				continue
			}
			id := "legacy_" + stringValue(u["id"])
			if ministries[id] == nil {
				m := legacyMap(cloneAny(reg))
				m["id"] = id
				m["createdAt"] = reg["updatedAt"]
				m["history"] = []any{map[string]any{"action": "Registro conservado de la versión anterior", "at": reg["updatedAt"]}}
				ministries[id] = m
			}
			p := legacyProgressForRequirement(db, stringValue(u["id"]), musicProgram553, false)
			links[stringValue(u["id"])] = map[string]any{"ministryId": id, "linkedAt": reg["updatedAt"], "legacy": true, "previousEvidence": map[string]any{"submission": cloneAny(mapField(p, "submissions")[ministryRequirement610]), "status": mapField(p, "status")[ministryRequirement610]}}
		}
		if d["managementYear610"] == nil {
			d["managementYear610"] = time.Now().Year()
		}
	}
	if musicDefaults618(db) {
		changed = true
	}
	templates := ensureMapField(d, "templates")
	for kind, title := range certificateTypes550 {
		t := legacyMap(templates[kind])
		if t == nil {
			t = map[string]any{}
			templates[kind] = t
			changed = true
		}
		if t["name"] == nil {
			t["name"] = title
			changed = true
		}
		if t["certificateType"] == nil {
			t["certificateType"] = kind
			changed = true
		}
		if t["textOnly620"] != true && t["signatureDepartment"] == nil {
			t["signatureDepartment"] = "Firma Departamental"
			changed = true
		}
		if t["textOnly620"] != true && t["signatureCoordinator"] == nil {
			t["signatureCoordinator"] = "Firma Coord. General Música"
			changed = true
		}
	}
	for _, raw := range templates {
		t := legacyMap(raw)
		if musicTextOnly620(t) {
			changed = true
		}
		for _, field := range arrayField(t, "fields") {
			f := legacyMap(field)
			if f["text"] == "Emitido el {fecha}" && numericValue(f["y"]) == 81 {
				f["y"] = 89
				changed = true
			}
		}
	}
	if musicRegistryDefaults622(db) {
		changed = true
	}
	return changed
}

func musicRegistryState610(db, actor map[string]any, st legacyStore) map[string]any {
	return musicRegistryState624(db, actor, st, false)
}

func musicRegistryState624(db, actor map[string]any, st legacyStore, compact bool) map[string]any {
	d := musicData550(db)
	ministries := []any{}
	for _, raw := range mapField(d, "ministries610") {
		m := map[string]any{}
		if compact {
			source := legacyMap(raw)
			for _, key := range []string{"id", "name", "type", "otherType", "registrationCode622", "church", "deletedAt", "updatedAt", "createdDate"} {
				if source[key] != nil {
					m[key] = source[key]
				}
			}
		} else {
			m = cloneMap610(legacyMap(raw))
		}
		y, mo := musicAge622(stringValue(m["createdDate"]), time.Now().UTC().Format("2006-01-02"))
		m["years"], m["months"] = y, mo
		ministries = append(ministries, m)
	}
	sort.SliceStable(ministries, func(i, j int) bool {
		return strings.ToLower(stringValue(legacyMap(ministries[i])["name"])) < strings.ToLower(stringValue(legacyMap(ministries[j])["name"]))
	})
	people := []any{}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if legacyActive(u) && (musicAssigned550(db, stringValue(u["id"])) || musicAllowed553(db, u)) {
			people = append(people, map[string]any{"id": u["id"], "fullName": u["fullName"], "role": u["role"], "church": u["church"], "assigned": musicAssigned550(db, stringValue(u["id"]))})
		}
	}
	permissions := map[string]any{}
	for _, key := range []string{"ministries.view", "ministries.create", "ministries.edit", "ministries.delete", "ministries.link", "ministries.unlink", "ministries.export", "ministries.modalities"} {
		permissions[key] = musicMinistryPermit610(db, actor, key)
	}
	certs := musicCertificateList619(db, actor)
	publicLinks := map[string]any{}
	for id, raw := range mapField(d, "ministryLinks610") {
		copy := cloneMap610(legacyMap(raw))
		delete(copy, "previousEvidence")
		publicLinks[id] = copy
	}
	var templates any
	if compact {
		templates = musicTemplateIndex624(db)
	} else {
		templates = cloneAny(mapField(d, "templates"))
	}
	result := map[string]any{"ok": true, "registryFields622": musicRegistryFields622(), "modalities622": musicModalities622(d), "ministries": ministries, "links": publicLinks, "people": people, "templates": templates, "certificates": certs, "events": musicEvents624(db, actor), "permissions": permissions, "restoreGeneration": legacyCanonicalRestoreGeneration(st.State), "managementYear": d["managementYear610"], "archives": musicArchiveIndex610(db), "followup": musicFollowupView610(db, actor)}
	if !musicHistoryAllowed611(actor) {
		result["archives"] = []any{}
	}
	return result
}

func (s *Server) music610(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", 405)
		return
	}
	req := musicRequest610{ActorID: r.URL.Query().Get("actorId"), Value: map[string]any{}}
	if r.Method == http.MethodPost {
		if err := readJSONLimited(r, &req, 32<<20); err != nil {
			legacyError(w, 400, "Datos inválidos.", nil)
			return
		}
		if req.Value == nil {
			req.Value = map[string]any{}
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, "No se pudo abrir la base.", nil)
		return
	}
	db := legacyMap(st.State.Database)
	actor, ok := s.musicActor550(r, db, req.ActorID)
	if !ok || !musicAllowed553(db, actor) {
		legacyError(w, 403, "Inicia sesión con una cuenta habilitada para Música.", nil)
		return
	}
	if s.musicRead624(w, r, st, db, actor) {
		return
	}
	// Routine reminder polling must not clone or transfer certificate artwork,
	// ministry photos, history, or the complete registry for every signed-in user.
	if r.Method == http.MethodGet && r.URL.Query().Get("resource") == "followup" {
		jsonOut(w, 200, map[string]any{"ok": true, "followup": musicFollowupView610(db, actor), "restoreGeneration": legacyCanonicalRestoreGeneration(st.State)})
		return
	}
	if s.musicReadTools613(w, r, st, db, actor) {
		return
	}
	if s.musicReadFonts617(w, r, db, actor) || s.musicPreview617(w, r, st, db, actor, req) {
		return
	}
	// Stored PDFs are reconciled on writes. Reading one must not issue files,
	// clone artwork, or expand the complete registry.
	if r.Method == http.MethodGet && r.URL.Query().Get("resource") == "certificate" {
		s.musicCertificatePDF610(w, r, db, actor)
		return
	}
	db = legacyMap(cloneAny(st.State.Database))
	actor = legacyMap(cloneAny(actor)) // Keep the effective role of personal view.
	changed := musicEnsure610(db)
	st.State.Database = db
	if musicAutoCertificates610(&st) {
		changed = true
	}
	if r.Method == http.MethodGet {
		if changed {
			st.State.ServerVersion++
			st.State.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			if s.saveLegacy(st) != nil {
				legacyError(w, 500, "No se pudo preparar Música.", nil)
				return
			}
		}
		if s.musicRead624(w, r, st, db, actor) {
			return
		}
		resource := r.URL.Query().Get("resource")
		if resource == "certificates" {
			jsonOut(w, 200, map[string]any{"ok": true, "certificates": musicCertificateList619(db, actor), "restoreGeneration": legacyCanonicalRestoreGeneration(st.State)})
			return
		}
		if resource == "event-participants-csv" || resource == "event-participants-pdf" {
			s.musicEventExport611(w, r, db, actor)
			return
		}
		if resource == "linked-requirement" {
			if mapField(musicData550(db), "ministryLinks610")[stringValue(actor["id"])] == nil {
				legacyError(w, 404, "No hay un ministerio vinculado.", nil)
				return
			}
			p := legacyProgressForRequirement(db, stringValue(actor["id"]), musicProgram553, false)
			jsonOut(w, 200, map[string]any{"ok": true, "submission": cloneAny(mapField(p, "submissions")[ministryRequirement610]), "status": mapField(p, "status")[ministryRequirement610], "restoreGeneration": legacyCanonicalRestoreGeneration(st.State)})
			return
		}
		if resource == "certificate" {
			s.musicCertificatePDF610(w, r, db, actor)
			return
		}
		if resource == "ministries-pdf" {
			s.musicMinistriesPDF610(w, r, db)
			return
		}
		if resource == "archive" {
			s.musicArchiveView610(w, r, db, actor)
			return
		}
		jsonOut(w, 200, musicRegistryState610(db, actor, st))
		return
	}
	if req.RestoreGeneration != legacyCanonicalRestoreGeneration(st.State) {
		legacyError(w, 409, "Los datos cambiaron por una restauración o nueva gestión. Actualiza la pantalla.", nil)
		return
	}
	if personalRequest555(r) && req.Action != "certificate.issue" {
		legacyError(w, 403, "Cambia a la vista de gestión para realizar esta acción.", nil)
		return
	}
	if s.musicWriteTools613(w, r, &st, db, actor, req) {
		return
	}
	if strings.HasPrefix(req.Action, "system.") {
		s.musicSystem610(w, r, &st, db, actor, req)
		return
	}
	if req.Action == "certificate.delete" {
		if err = musicDeleteCertificate618(&st, db, actor, req); err != nil {
			legacyError(w, 403, err.Error(), nil)
			return
		}
	} else if req.Action == "certificate.issue" {
		kind, eventID := stringValue(req.Value["kind"]), stringValue(req.Value["eventId"])
		owner := legacyUser(db, req.OwnerID)
		if !musicPermit602(db, actor, "certificates.view") || !musicPermit602(db, actor, "certificates.issue") || !musicScope550(db, actor, owner) && !musicLeaders610(legacyRole(actor)) || !musicEligible550(db, req.OwnerID, kind, eventID) {
			legacyError(w, 403, "El certificado aún no está disponible o está fuera de tu alcance.", nil)
			return
		}
		mid := stringValue(legacyMap(mapField(musicData550(db), "ministryLinks610")[req.OwnerID])["ministryId"])
		tid := kind
		if eventID != "" {
			if configured := stringValue(legacyMap(mapField(musicData550(db), "events")[eventID])["templateId610"]); configured != "" {
				tid = configured
			}
		}
		musicIssue610(&st, db, req.OwnerID, kind, eventID, mid, tid, time.Now().UTC().Format(time.RFC3339Nano))
	} else if strings.HasPrefix(req.Action, "template.") {
		if err = musicValidateFonts613(req.Value, st.State.Files); err != nil {
			legacyError(w, 400, err.Error(), nil)
			return
		}
		if err = musicTemplateMutation610(db, actor, req); err != nil {
			legacyError(w, 403, err.Error(), nil)
			return
		}
	} else {
		if err = musicMinistryMutation610(&st, db, actor, req); err != nil {
			legacyError(w, 403, err.Error(), nil)
			return
		}
	}
	musicAutoCertificates610(&st)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	st.State.ServerVersion++
	st.State.UpdatedAt = now
	musicData550(db)["updatedAt"] = now
	st.State.Database = legacyPrepareCanonicalDB(db, st.State, "Ministerios y certificados actualizados")
	if err = s.saveLegacy(st); err != nil {
		legacyError(w, 500, "No se pudo confirmar el cambio. Actualiza antes de reintentar.", nil)
		return
	}
	jsonOut(w, 200, musicRegistryState624(db, actor, st, r.URL.Query().Get("compact624") == "1"))
}

func musicMinistryMutation610(st *legacyStore, db, actor map[string]any, req musicRequest610) error {
	d := musicData550(db)
	records := ensureMapField(d, "ministries610")
	links := ensureMapField(d, "ministryLinks610")
	v := req.Value
	id := strings.TrimSpace(stringValue(v["id"]))
	m := legacyMap(records[id])
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if req.Action == "ministry.save" {
		create := m == nil
		key := "ministries.edit"
		if create {
			key = "ministries.create"
			if id != "" {
				return fmt.Errorf("El ministerio ya no existe. Actualiza.")
			}
		}
		if !musicMinistryPermit610(db, actor, key) {
			return fmt.Errorf("Para editar un ministerio registrado contacta al equipo de gestión.")
		}
		if !create && (m["deletedAt"] != nil || stringValue(v["expectedUpdatedAt"]) != stringValue(m["updatedAt"])) {
			return fmt.Errorf("El registro cambió. Actualiza antes de guardar.")
		}
		out := map[string]any{}
		for _, k := range []string{"type", "otherType", "name", "church", "district", "region", "mission", "union", "createdDate", "boardVote", "director", "directorPhone", "biography"} {
			x := strings.TrimSpace(stringValue(v[k]))
			limit := 500
			if k == "biography" {
				limit = 3000
			}
			if len([]rune(x)) > limit {
				return fmt.Errorf("El campo %s es demasiado extenso.", k)
			}
			out[k] = x
		}
		for _, k := range []string{"type", "name", "church", "district", "region", "mission", "union", "createdDate", "boardVote", "director", "directorPhone"} {
			if out[k] == "" {
				return fmt.Errorf("Completa todos los datos institucionales y del director.")
			}
		}
		if out["type"] != "Otro" {
			value := strings.TrimSpace(stringValue(out["type"]))
			if len([]rune(value)) > 80 {
				return fmt.Errorf("La modalidad admite hasta 80 caracteres.")
			}
			found := false
			for _, raw := range musicModalities622(d) {
				if strings.EqualFold(strings.Join(strings.Fields(stringValue(raw)), " "), strings.Join(strings.Fields(value), " ")) {
					out["type"] = raw
					found = true
					break
				}
			}
			if !found {
				if req.Value["addModality622"] != true || !musicMinistryPermit610(db, actor, "ministries.modalities") {
					return fmt.Errorf("No tienes permiso para añadir esta modalidad.")
				}
				out["type"] = strings.Join(strings.Fields(value), " ")
			}
		}
		if out["type"] == "Otro" && out["otherType"] == "" {
			return fmt.Errorf("Especifica el tipo de ministerio.")
		}
		date, e := time.Parse("2006-01-02", stringValue(out["createdDate"]))
		if e != nil || date.After(time.Now()) {
			return fmt.Errorf("Indica una fecha de creación válida.")
		}
		for _, k := range []string{"years", "months"} {
			n := numericValue(v[k])
			max := 200.
			if k == "months" {
				max = 11
			}
			if n < 0 || n > max || n != float64(int(n)) {
				return fmt.Errorf("La trayectoria debe indicar años y meses válidos.")
			}
			out[k] = int(n)
		}
		photo := stringValue(v["photo"])
		if photo != "" && !validMusicPhoto592(photo) {
			return fmt.Errorf("Usa una fotografía PNG/JPG válida de hasta 2 MB.")
		}
		if out["type"] != "Otro" && !containsString(musicModalities622(d), stringValue(out["type"])) {
			d["modalities622"] = append(musicModalities622(d), out["type"])
		}
		years, months := musicAge622(stringValue(out["createdDate"]), now)
		out["years"], out["months"] = years, months
		out["photo"] = photo
		if create {
			id = "min_" + token(12)
			out["createdAt"] = now
			out["createdBy"] = actor["id"]
			out["history"] = []any{}
		} else {
			out["createdAt"] = m["createdAt"]
			out["createdBy"] = m["createdBy"]
			out["history"] = cloneAny(arrayField(m, "history"))
			out["history"] = append(arrayField(out, "history"), map[string]any{"action": "Datos anteriores a la edición", "at": now, "actorId": actor["id"], "actorName": actor["fullName"], "snapshot": ministrySnapshot610(m)})
		}
		if create {
			out["registrationCode622"] = musicNextCode622(d)
		} else {
			out["registrationCode622"] = m["registrationCode622"]
		}
		out["id"] = id
		out["updatedAt"] = now
		out["updatedBy"] = actor["id"]
		ministryHistory610(out, actor, map[bool]string{true: "Ministerio registrado", false: "Datos actualizados"}[create], now)
		records[id] = out
		if c := legacyMap(mapField(d, "certificates")["ministry-registration:"+id]); c != nil {
			c["revokedAt"] = now
			c["revokedReason"] = "Registro actualizado"
		}
		musicIssue610(st, db, "", "ministry", "", id, "ministry", now)
		return nil
	}
	if req.Action == "ministry.purge624" {
		return musicPurgeMinistry624(st, db, actor, req)
	}
	if req.Action == "ministry.restore623" {
		if !musicMinistryPermit610(db, actor, "ministries.delete") {
			return fmt.Errorf("No tienes permiso para restaurar ministerios.")
		}
		if m == nil || m["deletedAt"] == nil {
			return fmt.Errorf("El ministerio ya no está en la papelera.")
		}
		delete(m, "deletedAt")
		delete(m, "deletedBy")
		m["updatedAt"] = now
		m["updatedBy"] = actor["id"]
		ministryHistory610(m, actor, "Ministerio restaurado de la papelera", now)
		musicIssue610(st, db, "", "ministry", "", id, "ministry", now)
		return nil
	}
	if m == nil || m["deletedAt"] != nil {
		return fmt.Errorf("Ministerio no disponible.")
	}
	switch req.Action {
	case "ministry.reissue622":
		if !musicMinistryPermit610(db, actor, "ministries.edit") || !musicPermit602(db, actor, "certificates.issue") {
			return fmt.Errorf("No tienes permiso para emitir certificados.")
		}
		if c := legacyMap(mapField(d, "certificates")["ministry-registration:"+id]); c != nil {
			c["revokedAt"] = now
			c["revokedReason"] = "Nueva emisión del registro"
		}
		musicIssue610(st, db, "", "ministry", "", id, "ministry", now)
		ministryHistory610(m, actor, "Certificado de registro emitido", now)
		return nil
	case "ministry.link":
		if !musicMinistryPermit610(db, actor, "ministries.link") {
			return fmt.Errorf("No tienes permiso para vincular aspirantes.")
		}
		owner := legacyUser(db, req.OwnerID)
		if !legacyActive(owner) || !musicAssigned550(db, req.OwnerID) {
			return fmt.Errorf("Selecciona un aspirante habilitado para Música.")
		}
		if old := legacyMap(links[req.OwnerID]); old != nil {
			if old["ministryId"] == id && old["legacy"] != true {
				return nil
			}
			if old["ministryId"] == id {
				delete(links, req.OwnerID)
			} else {
				return fmt.Errorf("El aspirante ya pertenece a otro ministerio. Desvincúlalo primero.")
			}
		}
		p := reviewProgressContainer(db, req.OwnerID, musicProgram553)
		saved := map[string]any{"submission": cloneAny(mapField(p, "submissions")[ministryRequirement610]), "status": mapField(p, "status")[ministryRequirement610], "at": now}
		links[req.OwnerID] = map[string]any{"ministryId": id, "linkedAt": now, "linkedBy": actor["id"], "previousEvidence": saved}
		owner["musicMinistryRegistration"] = ministrySnapshot610(m)
		c := musicIssue610(st, db, req.OwnerID, "ministry", "", id, "ministry", now)
		if c == nil {
			return fmt.Errorf("No se pudo generar el certificado.")
		}
		fileID := stringValue(c["fileId610"])
		ensureMapField(p, "submissions")[ministryRequirement610] = map[string]any{"submissionMode": "document", "deliveryMode": "document", "documentMode": true, "documentFormat": "pdf", "primaryDocumentId": fileID, "attachmentIds": []any{fileID}, "documentPageCount": 1, "documentSource": "ministry_link", "documentOriginalName": "Certificado_de_ministerio.pdf", "linkedMinistryId610": id, "completionDate": now[:10], "updatedAt": now}
		ensureMapField(p, "status")[ministryRequirement610] = "in_progress"
		ensureMapField(db, "reportRequirementModes")[musicProgram553+":"+ministryRequirement610] = "document"
		ministryHistory610(m, actor, "Aspirante vinculado: "+stringValue(owner["fullName"]), now)
		addLegacyNotification(st, legacyNotification{UserID: req.OwnerID, Type: "ministry_link610", Title: "Certificado de ministerio disponible", Message: "Tu ministerio fue vinculado. El certificado PDF está listo para enviar el requisito a revisión.", ProgramID: musicProgram553, ActorID: stringValue(actor["id"]), CreatedAt: now})
		return nil
	case "ministry.unlink":
		if !musicMinistryPermit610(db, actor, "ministries.unlink") {
			return fmt.Errorf("No tienes permiso para desvincular aspirantes.")
		}
		if legacyMap(links[req.OwnerID])["ministryId"] != id {
			return fmt.Errorf("El aspirante ya no está vinculado a este ministerio.")
		}
		musicUnlink610(st, db, actor, req.OwnerID, now)
		return nil
	case "ministry.delete":
		if !musicMinistryPermit610(db, actor, "ministries.delete") {
			return fmt.Errorf("Solo el equipo de gestión puede eliminar ministerios.")
		}
		if stringValue(v["confirmation"]) != "ELIMINAR MINISTERIO" {
			return fmt.Errorf("Confirma la eliminación del ministerio.")
		}
		for ownerID, raw := range links {
			if legacyMap(raw)["ministryId"] == id {
				musicUnlink610(st, db, actor, ownerID, now)
			}
		}
		ministryHistory610(m, actor, "Ministerio movido a papelera", now)
		m["deletedAt"] = now
		m["deletedBy"] = actor["id"]
		if c := legacyMap(mapField(d, "certificates")["ministry-registration:"+id]); c != nil {
			c["revokedAt"] = now
			c["revokedReason"] = "Ministerio eliminado del registro vigente"
		}
		return nil
	}
	return fmt.Errorf("Acción no reconocida.")
}

func musicUnlink610(st *legacyStore, db, actor map[string]any, ownerID, now string) {
	d := musicData550(db)
	links := ensureMapField(d, "ministryLinks610")
	link := legacyMap(links[ownerID])
	id := stringValue(link["ministryId"])
	if m := legacyMap(mapField(d, "ministries610")[id]); m != nil {
		ministryHistory610(m, actor, "Aspirante desvinculado: "+stringValue(legacyUser(db, ownerID)["fullName"]), now)
	}
	c := legacyMap(mapField(d, "certificates")[ownerID+":ministry:"])
	if c != nil {
		c["revokedAt"] = now
		c["revokedReason"] = "Vínculo con el ministerio finalizado"
	}
	p := reviewProgressContainer(db, ownerID, musicProgram553)
	old := legacyMap(cloneAny(legacyMap(link["previousEvidence"])["submission"]))
	if old == nil {
		old = map[string]any{}
	}
	delete(old, "linkedMinistryId610")
	old["updatedAt"] = now
	ensureMapField(p, "submissions")[ministryRequirement610] = old
	ensureMapField(p, "status")[ministryRequirement610] = "in_progress"
	delete(legacyUser(db, ownerID), "musicMinistryRegistration")
	d["ministryLinkHistory610"] = append(arrayField(d, "ministryLinkHistory610"), map[string]any{"ownerId": ownerID, "ministryId": id, "linkedAt": link["linkedAt"], "unlinkedAt": now, "unlinkedBy": actor["id"]})
	delete(links, ownerID)
}

func musicIssue610(st *legacyStore, db map[string]any, ownerID, kind, eventID, ministryID, templateID, now string) map[string]any {
	d := musicData550(db)
	certs := ensureMapField(d, "certificates")
	key := ownerID + ":" + kind + ":" + eventID
	if ownerID == "" {
		key = "ministry-registration:" + ministryID
	}
	if old := legacyMap(certs[key]); old != nil {
		if old["revokedAt"] == nil && old["fileId610"] != nil {
			return old
		}
		d["certificateHistory610"] = append(arrayField(d, "certificateHistory610"), cloneAny(old))
	}
	u := legacyUser(db, ownerID)
	t := legacyMap(mapField(d, "templates")[templateID])
	if t == nil {
		t = legacyMap(mapField(d, "templates")[kind])
	}
	title := certificateTypes550[kind]
	if t["name"] != nil {
		title = stringValue(t["name"])
	}
	c := map[string]any{"id": key, "token": randomMusicID550(), "ownerId": ownerID, "fullName": u["fullName"], "kind": kind, "title": title, "eventId": eventID, "ministryId": ministryID, "issuedAt": now, "church": u["church"], "mission": u["mission"], "orientation": "landscape", "layoutVersion618": musicTemplateLayout618(t), "templateId610": templateID, "templateVersion610": musicTemplateVersion610(d, t), "event": cloneAny(mapField(d, "events")[eventID]), "registration": ministrySnapshot610(legacyMap(mapField(d, "ministries610")[ministryID]))}
	if ev := legacyMap(c["event"]); ev != nil {
		delete(ev, "participants")
		delete(ev, "externalParticipants")
	}
	if ownerID == "" {
		c["fullName"] = legacyMap(c["registration"])["name"]
	}
	if c["registration"] == nil && kind == "ministry" {
		c["registration"] = ministrySnapshot610(legacyMap(u["musicMinistryRegistration"]))
	}
	if kind == "final" {
		c["approvalDate618"] = musicApproval618(db, ownerID)
	}
	if base := musicQRBase613(db); base != "" {
		c["verificationURL610"] = base + "/api/music/certificate?token=" + stringValue(c["token"])
	}
	pdf := musicCertificateBytes610(c, musicHydrateFonts613(t, st.State.Files))
	fileID := "cert610_" + token(12)
	sum := sha256.Sum256(pdf)
	c["fileId610"] = fileID
	c["sha256610"] = hex.EncodeToString(sum[:])
	certs[key] = c
	st.State.Files = append(st.State.Files, map[string]any{"id": fileID, "ownerId": ownerID, "programId": musicProgram553, "requirementId": map[bool]string{true: ministryRequirement610, false: ""}[kind == "ministry"], "kind": "requirement_document", "type": "application/pdf", "name": musicFilename618(c, false), "size": len(pdf), "dataUrl": "data:application/pdf;base64," + base64.StdEncoding.EncodeToString(pdf), "createdAt": now, "generated610": true, "certificateId610": key})
	return c
}

func musicAutoCertificates610(st *legacyStore) bool {
	db := legacyMap(st.State.Database)
	if db == nil {
		return false
	}
	d := musicData550(db)
	if d["schema610"] != true {
		return false
	}
	changed := musicInstallAssets618(st)
	for _, raw := range mapField(d, "templates") {
		if musicTextOnly620(legacyMap(raw)) {
			changed = true
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		id := stringValue(u["id"])
		if !legacyActive(u) || !musicAssigned550(db, id) {
			continue
		}
		if !musicComplete550(db, id) {
			if c := legacyMap(mapField(d, "certificates")[id+":final:"]); c != nil && c["revokedAt"] == nil {
				c["revokedAt"] = now
				c["revokedReason"] = "Requisitos pendientes de revisión"
				changed = true
			}
		}
		if musicComplete550(db, id) && musicApproval618(db, id) != "" {
			old := legacyMap(mapField(d, "certificates")[id+":final:"])
			if old == nil || old["fileId610"] == nil || old["revokedAt"] != nil && old["revokedReason"] == "Requisitos pendientes de revisión" {
				musicIssue610(st, db, id, "final", "", "", "final", now)
				changed = true
			} else if musicRefreshFinal618(st, db, old) {
				changed = true
			}
		}
	}
	for mid, raw := range mapField(d, "ministries610") {
		m := legacyMap(raw)
		if m["deletedAt"] == nil && mapField(d, "certificates")["ministry-registration:"+mid] == nil {
			musicIssue610(st, db, "", "ministry", "", mid, "ministry", now)
			changed = true
		}
	}
	for ownerID, raw := range mapField(d, "ministryLinks610") {
		link := legacyMap(raw)
		mid := stringValue(link["ministryId"])
		m := legacyMap(mapField(d, "ministries610")[mid])
		if m == nil || m["deletedAt"] != nil {
			continue
		}
		c := legacyMap(mapField(d, "certificates")[ownerID+":ministry:"])
		if c == nil || c["fileId610"] == nil {
			c = musicIssue610(st, db, ownerID, "ministry", "", mid, "ministry", now)
			changed = true
		}
		if link["legacy"] == true {
			p := reviewProgressContainer(db, ownerID, musicProgram553)
			state := stringValue(mapField(p, "status")[ministryRequirement610])
			if state != "approved" && state != "in_review" {
				fileID := stringValue(c["fileId610"])
				ensureMapField(p, "submissions")[ministryRequirement610] = map[string]any{"submissionMode": "document", "deliveryMode": "document", "documentMode": true, "documentFormat": "pdf", "primaryDocumentId": fileID, "attachmentIds": []any{fileID}, "documentPageCount": 1, "documentSource": "ministry_link", "documentOriginalName": "Certificado_de_ministerio.pdf", "completionDate": now[:10], "updatedAt": now}
				ensureMapField(p, "status")[ministryRequirement610] = "in_progress"
				ensureMapField(db, "reportRequirementModes")[musicProgram553+":"+ministryRequirement610] = "document"
				link["legacy"] = false
				changed = true
			}
		}
	}

	for eventID, raw := range mapField(d, "events") {
		ev := legacyMap(raw)
		templateID := stringValue(ev["templateId610"])
		if templateID == "" {
			if ev["templateId610"] == nil && musicEventKind551(ev) != "general" {
				templateID = musicEventKind551(ev)
				ev["templateId610"] = templateID
				changed = true
			} else {
				continue
			}
		}
		kind := musicEventKind551(ev)
		if kind == "general" {
			kind = "event"
		}
		for ownerID, p := range mapField(ev, "participants") {
			if !musicFrom2026_618(stringValue(ev["date"])) {
				continue
			}
			if legacyMap(p)["attended"] != true {
				continue
			}
			old := legacyMap(mapField(d, "certificates")[ownerID+":"+kind+":"+eventID])
			if old == nil || old["revokedAt"] != nil && strings.HasPrefix(stringValue(old["revokedReason"]), "Asistencia") {
				musicIssue610(st, db, ownerID, kind, eventID, "", templateID, now)
				changed = true
			}
		}
	}
	if musicRefreshLinked618(st, db) {
		changed = true
	}
	return changed
}

func cloneMap610(m map[string]any) map[string]any {
	c := legacyMap(cloneAny(m))
	if c == nil {
		return map[string]any{}
	}
	return c
}
func musicTemplateVersion610(d, t map[string]any) string {
	b, _ := json.Marshal(t)
	sum := sha256.Sum256(b)
	key := hex.EncodeToString(sum[:])
	ensureMapField(d, "templateVersions610")[key] = cloneAny(t)
	return key
}
func musicBackupExtras610(st legacyStore) map[string]any {
	b, _ := json.Marshal(st)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	delete(m, "state")
	return m
}

func musicRevokeEvent610(data map[string]any, eventID, ownerID, reason, at string) {
	for _, raw := range mapField(data, "certificates") {
		c := legacyMap(raw)
		if c["eventId"] == eventID && (ownerID == "" || c["ownerId"] == ownerID) && c["revokedAt"] == nil {
			c["revokedAt"] = at
			c["revokedReason"] = reason
		}
	}
}
