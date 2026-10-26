package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func musicArchiveIndex610(db map[string]any) []any {
	out := []any{}
	for year, raw := range mapField(musicData550(db), "archiveData610") {
		a := legacyMap(raw)
		out = append(out, map[string]any{"year": year, "closedAt": a["closedAt"], "closedBy": a["closedBy"]})
	}
	sort.Slice(out, func(i, j int) bool {
		return stringValue(legacyMap(out[i])["year"]) > stringValue(legacyMap(out[j])["year"])
	})
	return out
}

func (s *Server) musicArchiveView610(w http.ResponseWriter, r *http.Request, db, actor map[string]any) {
	if personalRequest555(r) || !musicHistoryAllowed611(actor) {
		legacyError(w, 403, "El historial de gestiones está disponible en la vista de gestión.", nil)
		return
	}
	a := legacyMap(mapField(musicData550(db), "archiveData610")[r.URL.Query().Get("year")])
	old := legacyMap(a["database"])
	if old == nil {
		legacyError(w, 404, "Gestión no encontrada.", nil)
		return
	}
	oldActor := musicArchiveActor610(old, actor)
	d := musicData550(old)
	people := []any{}
	for _, raw := range legacyUsers(old) {
		u := legacyMap(raw)
		if musicScope550(old, oldActor, u) {
			people = append(people, map[string]any{"id": u["id"], "fullName": u["fullName"], "church": u["church"], "role": u["role"], "progress": cloneAny(legacyProgressForRequirement(old, stringValue(u["id"]), musicProgram553, false))})
		}
	}
	certs := []any{}
	for _, raw := range mapField(d, "certificates") {
		c := legacyMap(raw)
		if musicScope550(old, oldActor, legacyUser(old, stringValue(c["ownerId"]))) || musicLeaders610(legacyRole(actor)) {
			copy := legacyMap(cloneAny(c))
			delete(copy, "templateSnapshot610")
			certs = append(certs, copy)
		}
	}
	ministries := []any{}
	for _, m := range mapField(d, "ministries610") {
		ministries = append(ministries, cloneAny(m))
	}
	events := cloneMap610(mapField(d, "events"))
	if !musicRoster610(legacyRole(actor)) {
		for _, raw := range events {
			ev := legacyMap(raw)
			mine := map[string]any{}
			if p := mapField(ev, "participants")[stringValue(actor["id"])]; p != nil {
				mine[stringValue(actor["id"])] = p
			}
			ev["participants"] = mine
			delete(ev, "externalParticipants")
		}
	}
	jsonOut(w, 200, map[string]any{"ok": true, "readOnly": true, "year": r.URL.Query().Get("year"), "people": people, "ministries": ministries, "certificates": certs, "events": events, "closedAt": a["closedAt"]})
}

func musicMaintenance610(actor map[string]any) bool {
	return legacyRole(actor) == "admin" || legacyRole(actor) == "secretaria"
}

func (s *Server) musicSystem610(w http.ResponseWriter, r *http.Request, st *legacyStore, db, actor map[string]any, req musicRequest610) {
	if !musicMaintenance610(actor) {
		legacyError(w, 403, "Solo Admin y Secretaría pueden reiniciar el sistema o iniciar una gestión.", nil)
		return
	}
	d := musicData550(db)
	v := req.Value
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if req.Action == "system.backup" {
		// Export all records and attachment bytes once. Archives contain references,
		// not another copy of each PDF. The saved copy is outside the active data.
		backup := map[string]any{"format": "MinisterioMusicaOnlineBackup", "version": musicOnlineVersion, "exportedAt": now, "database": cloneAny(db), "files": cloneAny(st.State.Files), "store610": musicBackupExtras610(*st), "transport610": cloneAny(s.store)}
		b, err := json.Marshal(backup)
		if err != nil {
			legacyError(w, 500, "No se pudo preparar el respaldo.", nil)
			return
		}
		var verification map[string]any
		if json.Unmarshal(b, &verification) != nil || verification["database"] == nil {
			legacyError(w, 500, "El respaldo no pasó la comprobación de integridad.", nil)
			return
		}
		id := token(20)
		dir := filepath.Join(filepath.Dir(s.path), "respaldos")
		name := "Musica_respaldo_completo_" + time.Now().Format("20060102_150405") + "_" + id[:8] + ".json"
		if err = os.MkdirAll(dir, 0700); err == nil {
			err = os.WriteFile(filepath.Join(dir, name), b, 0600)
		}
		if err != nil {
			legacyError(w, 500, "No se pudo guardar el respaldo obligatorio. El reinicio permanece bloqueado.", nil)
			return
		}
		sum := sha256.Sum256(b)
		d["maintenanceBackup610"] = map[string]any{"id": id, "actorId": actor["id"], "file": name, "sha256": hex.EncodeToString(sum[:]), "version": st.State.ServerVersion, "generation": legacyCanonicalRestoreGeneration(st.State), "createdAt": now}
		st.State.Database = db
		if s.saveLegacy(*st) != nil {
			legacyError(w, 500, "No se pudo confirmar el respaldo.", nil)
			return
		}
		jsonOut(w, 200, map[string]any{"ok": true, "backup": backup, "backupToken": id, "fileName": name, "sha256": hex.EncodeToString(sum[:])})
		return
	}
	if req.Action == "system.reset" {
		proof := legacyMap(d["maintenanceBackup610"])
		if proof["actorId"] != actor["id"] || proof["id"] != v["backupToken"] || numericValue(proof["version"]) != float64(st.State.ServerVersion) || numericValue(proof["generation"]) != float64(legacyCanonicalRestoreGeneration(st.State)) {
			legacyError(w, 409, "Crea y descarga un respaldo completo actualizado antes de reiniciar.", nil)
			return
		}
		b, err := os.ReadFile(filepath.Join(filepath.Dir(s.path), "respaldos", filepath.Base(stringValue(proof["file"]))))
		sum := sha256.Sum256(b)
		if err != nil || len(b) == 0 || hex.EncodeToString(sum[:]) != proof["sha256"] {
			legacyError(w, 409, "El respaldo obligatorio no está disponible o cambió. No se eliminó ningún dato.", nil)
			return
		}
		if v["confirmation"] != "REINICIAR DESDE CERO" {
			legacyError(w, 400, "Escribe REINICIAR DESDE CERO para confirmar.", nil)
			return
		}
		key := token(24)
		if os.WriteFile(s.installationKeyPath(), []byte(key+"\n"), 0600) != nil {
			legacyError(w, 500, "No se pudo preparar la configuración inicial. No se eliminó ningún dato.", nil)
			return
		}
		designDB, designFonts := musicResetDesign619(db, st.State.Files)
		next := legacyStore{State: legacyState{ServerVersion: st.State.ServerVersion + 1, RestoreGeneration: legacyCanonicalRestoreGeneration(st.State) + 1, RestoreID: "reset_" + token(12), RestoredAt: now, UpdatedAt: now, Database: designDB, Files: designFonts}}
		oldTransport := s.store
		s.store.Entities = map[string]Entity{}
		s.store.Pending = map[string]Pending{}
		s.store.Audit = nil
		s.store.Revision++
		if err := s.save(); err != nil {
			s.store = oldTransport
			legacyError(w, 500, "No se pudo limpiar el almacenamiento activo. No se confirmó el reinicio.", nil)
			return
		}
		if s.saveLegacy(next) != nil {
			s.store = oldTransport
			_ = s.save()
			legacyError(w, 500, "No se pudo confirmar el reinicio. Conserva el respaldo.", nil)
			return
		}
		if s.legacyStorageMode == "sqlite" {
			if err := s.writeLegacyJSON(next); err != nil {
				_ = s.saveLegacy(*st)
				s.store = oldTransport
				_ = s.save()
				legacyError(w, 500, "No se pudo limpiar la copia de compatibilidad. Se conservaron los datos anteriores.", nil)
				return
			}
		}
		s.publishSystemRestoredControl("Sistema reiniciado. Configura la nueva cuenta administradora.", next.State, time.Now())
		s.invalidateMusicSessions610()
		_ = os.Remove(s.resetPath610())
		jsonOut(w, 200, map[string]any{"ok": true, "loginRequired": true, "setupKey": key, "backupFile": proof["file"]})
		return
	}
	if req.Action == "system.year" {
		year := int(numericValue(v["year"]))
		oldYear := int(numericValue(d["managementYear610"]))
		if oldYear == 0 {
			oldYear = time.Now().Year()
		}
		if year <= oldYear || year > oldYear+10 {
			legacyError(w, 400, "Selecciona una gestión posterior a la actual.", nil)
			return
		}
		if v["confirmation"] != "INICIAR GESTION "+strconv.Itoa(year) {
			legacyError(w, 400, "Confirma la nueva gestión con el texto indicado.", nil)
			return
		}
		archives := legacyMap(cloneAny(mapField(d, "archiveData610")))
		if archives == nil {
			archives = map[string]any{}
		}
		if archives[strconv.Itoa(oldYear)] != nil {
			legacyError(w, 409, "La gestión actual ya fue archivada. Actualiza la pantalla.", nil)
			return
		}
		previous := legacyMap(cloneAny(db))
		pd := musicData550(previous)
		delete(pd, "archiveData610")
		delete(pd, "templateVersions610")
		refs := map[string]any{}
		for id, raw := range mapField(pd, "templates") {
			t := legacyMap(raw)
			refs[id] = map[string]any{"name": t["name"], "version610": musicTemplateVersion610(d, t)}
		}
		pd["templates"] = refs
		delete(pd, "maintenanceBackup610")
		archives[strconv.Itoa(oldYear)] = map[string]any{"database": previous, "closedAt": now, "closedBy": actor["fullName"], "notifications": cloneAny(st.Notifications), "messages": cloneAny(st.Messages)}
		next := legacyMap(cloneAny(db))
		nd := map[string]any{"registrationSequence622": d["registrationSequence622"], "modalities622": cloneAny(d["modalities622"]), "schema610": true, "managementYear610": year, "publicQRURL613": d["publicQRURL613"], "archiveData610": archives, "updatedAt": now, "templateVersions610": cloneAny(mapField(d, "templateVersions610"))}
		if v["keepTemplates"] != false {
			nd["templates"] = cloneAny(mapField(d, "templates"))
		}
		if v["keepPermissions"] != false {
			nd["permissions602"] = cloneAny(mapField(d, "permissions602"))
		}
		if v["keepMinistries"] != false {
			nd["ministries610"] = cloneAny(mapField(d, "ministries610"))
		}
		next["music550"] = nd
		next["programProgress"] = map[string]any{}
		next["progress"] = map[string]any{}
		next["tasks"] = []any{}
		next["audit"] = []any{}
		next["updatedAt"] = now
		if v["keepRequirements"] == false {
			next["configs"] = map[string]any{}
			next["indicationRules"] = []any{}
			next["reportRequirementModes"] = map[string]any{}
		}
		users := []any{}
		selected := map[string]bool{}
		for _, id := range arrayField(v, "selectedUserIds") {
			selected[stringValue(id)] = true
		}
		for _, raw := range legacyUsers(next) {
			u := legacyMap(raw)
			if v["keepUsers"] == false && !musicLeaders610(legacyRole(u)) {
				continue
			}
			delete(u, "musicMinistryRegistration")
			users = append(users, u)
		}
		next["users"] = users
		for _, raw := range arrayField(next, "modules") {
			m := legacyMap(raw)
			ids := []any{}
			for _, id := range arrayField(m, "assignedUserIds") {
				sid := stringValue(id)
				if legacyUser(next, sid) != nil && (v["selectedUserIds"] == nil || selected[sid]) {
					ids = append(ids, id)
				}
			}
			m["assignedUserIds"] = ids
		}
		// Carry only explicitly selected, unfinished applications. Previous
		// approvals remain in history and are not copied as new approvals.
		for _, raw := range arrayField(v, "carryUserIds") {
			id := stringValue(raw)
			if legacyUser(next, id) == nil || !musicAssigned550(next, id) {
				continue
			}
			oldProgress := legacyProgressForRequirement(db, id, musicProgram553, false)
			p := reviewProgressContainer(next, id, musicProgram553)
			for rid, sub := range mapField(oldProgress, "submissions") {
				pSub := legacyMap(cloneAny(sub))
				if pSub == nil {
					continue
				}
				delete(pSub, "submittedAt")
				ensureMapField(p, "submissions")[rid] = pSub
				ensureMapField(p, "status")[rid] = "in_progress"
			}
		}
		musicEnsure610(next)
		st.State.Database = next
		st.State.ServerVersion++
		st.State.RestoreGeneration = legacyCanonicalRestoreGeneration(st.State) + 1
		st.State.RestoreID = "year_" + token(12)
		st.State.RestoredAt = now
		st.State.UpdatedAt = now
		st.Proposals = nil
		st.Notifications = nil
		st.Messages = nil
		st.AppliedOperations = nil
		st.UserChanges = nil
		st.Recoveries548 = nil
		if s.saveLegacy(*st) != nil {
			legacyError(w, 500, "No se pudo guardar la nueva gestión.", nil)
			return
		}
		s.publishSystemRestoredControl("Se inició una nueva gestión. Vuelve a ingresar para actualizar tus datos.", st.State, time.Now())
		s.invalidateMusicSessions610()
		jsonOut(w, 200, map[string]any{"ok": true, "loginRequired": true, "year": year})
		return
	}
	legacyError(w, 400, "Acción de mantenimiento no reconocida.", nil)
}

func musicProjection610(db, actor, d map[string]any) {
	delete(d, "archiveData610")
	delete(d, "templateVersions610")
	delete(d, "maintenanceBackup610")
	delete(d, "followupSent610")
	for _, raw := range mapField(d, "ministryLinks610") {
		delete(legacyMap(raw), "previousEvidence")
	}
	for _, raw := range mapField(d, "certificates") {
		delete(legacyMap(raw), "templateSnapshot610")
	}
	if !musicLeaders610(legacyRole(actor)) {
		delete(d, "certificateHistory610")
	}
}

func musicReviewRows610(db map[string]any, actor map[string]any) []any {
	rows := []any{}
	now := time.Now()
	all := musicLeaders610(legacyRole(actor)) || legacyRole(actor) == "administrativo"
	for _, raw := range legacyUsers(db) {
		owner := legacyMap(raw)
		id := stringValue(owner["id"])
		if !musicAssigned550(db, id) || !legacyActive(owner) {
			continue
		}
		if !all && !musicReview560(db, actor, owner) {
			continue
		}
		p := legacyProgressForRequirement(db, id, musicProgram553, false)
		for rid, state := range mapField(p, "status") {
			if state != "in_review" {
				continue
			}
			sub := legacyMap(mapField(p, "submissions")[rid])
			at, e := time.Parse(time.RFC3339Nano, stringValue(sub["submittedAt"]))
			if e != nil {
				continue
			}
			days := int(now.Sub(at).Hours() / 24)
			if days < 3 {
				continue
			}
			rows = append(rows, map[string]any{"ownerId": id, "fullName": owner["fullName"], "requirementId": rid, "submittedAt": sub["submittedAt"], "days": days, "report": days >= 5})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		return numericValue(legacyMap(rows[i])["days"]) > numericValue(legacyMap(rows[j])["days"])
	})
	return rows
}
func musicFollowupView610(db, actor map[string]any) map[string]any {
	rows := musicReviewRows610(db, actor)
	pending := []any{}
	for _, raw := range rows {
		row := legacyMap(raw)
		owner := legacyUser(db, stringValue(row["ownerId"]))
		if musicAssignedReview611(db, actor, owner) {
			pending = append(pending, raw)
		}
	}
	return map[string]any{"pendingCount": len(pending), "pendingRows": pending, "showReminder": !musicLeaders610(legacyRole(actor)), "rows": rows, "canReport": musicLeaders610(legacyRole(actor)) || legacyRole(actor) == "administrativo"}
}
func musicFollowupNotifications610(st *legacyStore) bool {
	db := legacyMap(st.State.Database)
	if db == nil || musicData550(db)["schema610"] != true {
		return false
	}
	changed := false
	now := time.Now().UTC().Format(time.RFC3339Nano)
	sent := ensureMapField(musicData550(db), "followupSent610")
	seen := map[string]bool{}
	for _, n := range st.Notifications {
		seen[n.UserID+"|"+n.SourceOperationID] = true
	}
	for _, raw := range legacyUsers(db) {
		actor := legacyMap(raw)
		if !legacyActive(actor) || !musicAllowed553(db, actor) {
			continue
		}
		rows := musicReviewRows610(db, actor)
		for _, rr := range rows {
			row := legacyMap(rr)
			owner := legacyUser(db, stringValue(row["ownerId"]))
			recipient := stringValue(actor["id"])
			reviewer := musicAssignedReview611(db, actor, owner)
			report := row["report"] == true && (musicLeaders610(legacyRole(actor)) || legacyRole(actor) == "administrativo")
			if !reviewer && !report {
				continue
			}
			stage := "recordatorio"
			message := "Hay requisitos que llevan 3 días o más esperando tu revisión. Si permanecen pendientes al cumplirse 5 días desde su envío, se notificará al equipo de gestión para dar seguimiento al proceso."
			title := "Tienes requisitos pendientes de revisión"
			if report {
				stage = "reporte"
				title = "Seguimiento de requisitos pendientes"
				message = stringValue(owner["fullName"]) + " · " + stringValue(row["requirementId"]) + " · " + fmt.Sprint(row["days"]) + " días pendiente de revisión."
			}
			key := "revision610:" + stage + ":" + stringValue(row["ownerId"]) + ":" + stringValue(row["requirementId"]) + ":" + stringValue(row["submittedAt"])
			bound := recipient + ":" + stringValue(row["ownerId"]) + ":" + stringValue(row["requirementId"]) + ":" + stage
			if sent[bound] == row["submittedAt"] {
				continue
			}
			if seen[recipient+"|"+key] {
				sent[bound] = row["submittedAt"]
				changed = true
				continue
			}
			if !seen[recipient+"|"+key] {
				sent[bound] = row["submittedAt"]
				changed = true
				addLegacyNotification(st, legacyNotification{UserID: recipient, Type: "revision610", Title: title, Message: message, ProgramID: musicProgram553, OwnerID: stringValue(row["ownerId"]), RequirementID: stringValue(row["requirementId"]), CreatedAt: now, SourceOperationID: key})
				seen[recipient+"|"+key] = true
			}
		}
	}
	return changed
}

func (s *Server) musicFollowupLoop610() {
	timer := time.NewTicker(time.Minute)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
		}
		s.mu.Lock()
		st, err := s.loadLegacy()
		if err == nil {
			notified := musicFollowupNotifications610(&st)
			generated := musicAutoCertificates610(&st)
			if notified || generated {
				st.State.ServerVersion++
				st.State.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
				_ = s.saveLegacy(st)
			}
		}
		s.mu.Unlock()
	}
}

func musicLinkedSubmissionGuard610(db map[string]any, ownerID, reqID string, sub map[string]any) bool {
	if reqID != ministryRequirement610 {
		return true
	}
	link := legacyMap(mapField(musicData550(db), "ministryLinks610")[ownerID])
	if link == nil {
		return true
	}
	canonical := legacyMap(mapField(legacyProgressForRequirement(db, ownerID, musicProgram553, false), "submissions")[reqID])
	if link["legacy"] == true {
		return sub["primaryDocumentId"] == canonical["primaryDocumentId"] && stringValue(sub["text"]) == stringValue(canonical["text"])
	}
	ids := arrayField(sub, "attachmentIds")
	return sub["primaryDocumentId"] == canonical["primaryDocumentId"] && sub["primaryDocumentId"] != nil && reviewSubmissionIsDocumentMode(sub) && len(ids) == 1 && ids[0] == canonical["primaryDocumentId"]
}

func parseMusicYear610(s string) int { n, _ := strconv.Atoi(strings.TrimSpace(s)); return n }

func (s *Server) invalidateMusicSessions610() {
	s.remoteMu.Lock()
	defer s.remoteMu.Unlock()
	s.remoteSessions = map[string]remoteSession{}
	s.saveRemoteSessionsLocked()
}

func musicArchiveActor610(old, current map[string]any) map[string]any {
	u := cloneMap610(legacyUser(old, stringValue(current["id"])))
	if len(u) == 0 {
		u = cloneMap610(current)
	}
	u["role"] = current["role"]
	u["active"] = true
	u["accountState"] = "active"
	return u
}
