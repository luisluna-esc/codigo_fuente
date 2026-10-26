package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Screen indexes contain metadata only. Artwork and audit snapshots are read
// on demand; GET must never issue PDFs or copy the complete database.
func musicTemplateIndex624(db map[string]any) map[string]any {
	out := map[string]any{}
	for id, raw := range mapField(musicData550(db), "templates") {
		t := legacyMap(raw)
		row := map[string]any{"id": id}
		for _, key := range []string{"name", "certificateType", "updatedAt", "paperSize617"} {
			if t[key] != nil {
				row[key] = t[key]
			}
		}
		out[id] = row
	}
	return out
}

func musicEvents624(db, actor map[string]any) map[string]any {
	out := map[string]any{}
	if !musicPermit602(db, actor, "events.view") {
		return out
	}
	for id, raw := range mapField(musicData550(db), "events") {
		e := cloneMap610(legacyMap(raw))
		if !musicRoster610(legacyRole(actor)) {
			people := map[string]any{}
			for owner, value := range mapField(e, "participants") {
				if musicScope550(db, actor, legacyUser(db, owner)) {
					people[owner] = value
				}
			}
			e["participants"] = people
			e["externalParticipants"] = map[string]any{}
		}
		out[id] = e
	}
	return out
}

func (s *Server) musicRead624(w http.ResponseWriter, r *http.Request, st legacyStore, db, actor map[string]any) bool {
	if r.Method != http.MethodGet || musicData550(db)["schema610"] != true {
		return false
	}
	resource := r.URL.Query().Get("resource")
	base := map[string]any{"ok": true, "restoreGeneration": legacyCanonicalRestoreGeneration(st.State)}
	switch resource {
	case "templates", "template":
		if personalRequest555(r) || !musicFontManager613(db, actor) {
			legacyError(w, 403, "No tienes permiso para administrar plantillas.", nil)
			return true
		}
		if resource == "templates" {
			base["templates"] = musicTemplateIndex624(db)
			base["registryFields622"] = musicRegistryFields622()
		} else {
			t := mapField(musicData550(db), "templates")[r.URL.Query().Get("id")]
			if t == nil {
				legacyError(w, 404, "La plantilla ya no existe.", nil)
				return true
			}
			base["template"] = cloneAny(t)
		}
	case "registry":
		jsonOut(w, 200, musicRegistryState624(db, actor, st, true))
		return true
	case "certificates":
		base["certificates"] = musicCertificateList619(db, actor)
	case "ministry":
		if !musicMinistryPermit610(db, actor, "ministries.view") {
			legacyError(w, 403, "No tienes permiso para consultar ministerios.", nil)
			return true
		}
		id := r.URL.Query().Get("id")
		m := legacyMap(mapField(musicData550(db), "ministries610")[id])
		if m == nil {
			legacyError(w, 404, "El ministerio ya no existe.", nil)
			return true
		}
		copy := cloneMap610(m)
		copy["years"], copy["months"] = musicAge622(stringValue(m["createdDate"]), time.Now().UTC().Format("2006-01-02"))
		base["ministry"] = copy
		certs := []any{}
		for _, raw := range musicCertificateList619(db, actor) {
			c := legacyMap(raw)
			if c["ministryId"] == id && stringValue(c["ownerId"]) == "" {
				certs = append(certs, c)
			}
		}
		sort.SliceStable(certs, func(i, j int) bool {
			return stringValue(legacyMap(certs[i])["issuedAt"]) > stringValue(legacyMap(certs[j])["issuedAt"])
		})
		base["certificates"] = certs
	default:
		return false
	}
	jsonOut(w, 200, base)
	return true
}

// Purging is allowed only from Trash, with the same management permission as
// deletion and a version guard. Closed years and submitted evidence retain
// their independent copies. Registration numbers are never reused.
func musicPurgeMinistry624(st *legacyStore, db, actor map[string]any, req musicRequest610) error {
	d := musicData550(db)
	id := stringValue(req.Value["id"])
	m := legacyMap(mapField(d, "ministries610")[id])
	if !musicMinistryPermit610(db, actor, "ministries.delete") {
		return fmt.Errorf("No tienes permiso para eliminar definitivamente.")
	}
	if m == nil || m["deletedAt"] == nil {
		return fmt.Errorf("Solo se pueden eliminar definitivamente los ministerios de la papelera.")
	}
	if req.Value["confirmation"] != "ELIMINAR DEFINITIVAMENTE" {
		return fmt.Errorf("Escribe ELIMINAR DEFINITIVAMENTE para confirmar.")
	}
	if stringValue(req.Value["expectedUpdatedAt"]) != stringValue(m["updatedAt"]) {
		return fmt.Errorf("El registro cambió. Actualiza la papelera.")
	}
	for owner, raw := range mapField(d, "ministryLinks610") {
		if legacyMap(raw)["ministryId"] == id {
			musicUnlink610(st, db, actor, owner, time.Now().UTC().Format(time.RFC3339Nano))
		}
	}
	fileIDs := map[string]bool{}
	for key, raw := range mapField(d, "certificates") {
		c := legacyMap(raw)
		if c["ministryId"] == id {
			fileIDs[stringValue(c["fileId610"])] = true
			delete(mapField(d, "certificates"), key)
		}
	}
	history := []any{}
	for _, raw := range arrayField(d, "certificateHistory610") {
		c := legacyMap(raw)
		if c["ministryId"] == id {
			fileIDs[stringValue(c["fileId610"])] = true
		} else {
			history = append(history, raw)
		}
	}
	d["certificateHistory610"] = history
	links := []any{}
	for _, raw := range arrayField(d, "ministryLinkHistory610") {
		if legacyMap(raw)["ministryId"] != id {
			links = append(links, raw)
		}
	}
	d["ministryLinkHistory610"] = links
	delete(mapField(d, "ministries610"), id)
	d["ministryDeletions624"] = append(arrayField(d, "ministryDeletions624"), map[string]any{"id": id, "registrationCode622": m["registrationCode622"], "deletedBy": actor["id"], "deletedAt": time.Now().UTC().Format(time.RFC3339Nano)})
	refs := map[string]bool{}
	musicFileIDs553(db, refs)
	files := []any{}
	for _, raw := range st.State.Files {
		f := legacyMap(raw)
		fid := stringValue(f["id"])
		if !fileIDs[fid] || refs[fid] || !strings.HasPrefix(fid, "cert610_") {
			files = append(files, raw)
		}
	}
	st.State.Files = files
	return nil
}
