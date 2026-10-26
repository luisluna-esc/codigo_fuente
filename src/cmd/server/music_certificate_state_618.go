package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func musicCanDeleteCertificate618(db, actor, c map[string]any) bool {
	role := legacyRole(actor)
	if role != "admin" && role != "departamental" && role != "coordinador_musica" && role != "secretaria" && role != "administrativo" {
		return false
	}
	eventID := stringValue(c["eventId"])
	ownerID := stringValue(c["ownerId"])
	return musicPermit602(db, actor, "certificates.delete") && ownerID != "" && eventID != "" && mapField(musicData550(db), "events")[eventID] == nil && (musicLeaders610(role) || musicScope550(db, actor, legacyUser(db, ownerID)))
}

func musicDeleteCertificate618(st *legacyStore, db, actor map[string]any, req musicRequest610) error {
	d := musicData550(db)
	id := stringValue(req.Value["id"])
	token := stringValue(req.Value["token"])
	c := legacyMap(mapField(d, "certificates")[id])
	if c != nil && c["token"] != token {
		c = nil
	}
	if c == nil {
		for _, raw := range arrayField(d, "certificateHistory610") {
			candidate := legacyMap(raw)
			if candidate["id"] == id && candidate["token"] == token {
				c = candidate
				break
			}
		}
	}
	if c == nil {
		return fmt.Errorf("El certificado ya no existe. Actualiza la lista.")
	}
	if !musicCanDeleteCertificate618(db, actor, c) {
		return fmt.Errorf("Solo los responsables autorizados pueden eliminar certificados de eventos eliminados dentro de su alcance.")
	}
	if current := legacyMap(mapField(d, "certificates")[id]); current != nil && current["token"] == token {
		delete(mapField(d, "certificates"), id)
	}
	history := []any{}
	for _, raw := range arrayField(d, "certificateHistory610") {
		if legacyMap(raw)["token"] != token {
			history = append(history, raw)
		}
	}
	d["certificateHistory610"] = history
	// Retain a small action log. Evidence referenced by an approved requirement
	// or a closed management year must not be deleted with this certificate.
	d["certificateDeletions618"] = append(arrayField(d, "certificateDeletions618"), map[string]any{"certificateId": id, "ownerId": c["ownerId"], "eventId": c["eventId"], "title": c["title"], "deletedBy": actor["id"], "deletedAt": time.Now().UTC().Format(time.RFC3339Nano)})
	refs := map[string]bool{}
	musicFileIDs553(db, refs)
	fileID := stringValue(c["fileId610"])
	if fileID != "" && !refs[fileID] {
		files := []any{}
		for _, raw := range st.State.Files {
			if legacyMap(raw)["id"] != fileID {
				files = append(files, raw)
			}
		}
		st.State.Files = files
	}
	return nil
}

func musicRefreshFinal618(st *legacyStore, db, c map[string]any) bool {
	date := musicApproval618(db, stringValue(c["ownerId"]))
	if date == "" || c["approvalDate618"] == date {
		return false
	}
	c["approvalDate618"] = date
	t := legacyMap(mapField(musicData550(db), "templateVersions610")[stringValue(c["templateVersion610"])])
	if t == nil {
		t = legacyMap(mapField(musicData550(db), "templates")["final"])
	}
	pdf := musicCertificateBytes610(c, musicHydrateFonts613(t, st.State.Files))
	for _, raw := range st.State.Files {
		f := legacyMap(raw)
		if f["id"] == c["fileId610"] {
			f["dataUrl"] = "data:application/pdf;base64," + base64.StdEncoding.EncodeToString(pdf)
			f["size"] = len(pdf)
			f["name"] = musicFilename618(c, false)
			sum := sha256.Sum256(pdf)
			c["sha256610"] = hex.EncodeToString(sum[:])
			return true
		}
	}
	return true
}

func musicFrom2026_618(value string) bool {
	if len(value) < 10 {
		return false
	}
	date, err := time.Parse("2006-01-02", value[:10])
	return err == nil && date.Year() >= 2026
}

func musicTemplateLayout618(t map[string]any) string {
	layout := map[string]any{"renderer618": "6.2.0-ordinary-text-boxes"}
	for key, value := range t {
		if key != "updatedAt" && key != "updatedBy" && key != "id" {
			layout[key] = value
		}
	}
	raw, _ := json.Marshal(layout)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func musicUpdateCertificateFile618(st *legacyStore, c map[string]any, pdf []byte, archiveRefs map[string]bool) {
	id := stringValue(c["fileId610"])
	if id == "" || archiveRefs[id] {
		id = "cert610_" + token(12)
		c["fileId610"] = id
	}
	files := append([]any{}, st.State.Files...)
	found := false
	for at, raw := range files {
		old := legacyMap(raw)
		if old["id"] != id {
			continue
		}
		file := map[string]any{}
		for k, v := range old {
			file[k] = v
		}
		file["dataUrl"] = "data:application/pdf;base64," + base64.StdEncoding.EncodeToString(pdf)
		file["name"] = musicFilename618(c, false)
		file["size"] = len(pdf)
		files[at] = file
		found = true
		break
	}
	if !found {
		files = append(files, map[string]any{"id": id, "ownerId": c["ownerId"], "programId": musicProgram553, "kind": "requirement_document", "type": "application/pdf", "name": musicFilename618(c, false), "size": len(pdf), "dataUrl": "data:application/pdf;base64," + base64.StdEncoding.EncodeToString(pdf), "generated610": true, "certificateId610": c["id"], "createdAt": c["issuedAt"]})
	}
	st.State.Files = files
	sum := sha256.Sum256(pdf)
	c["sha256610"] = hex.EncodeToString(sum[:])
}

// Reconcile only changed designs. Tokens, QR URLs, approval dates and original
// issue dates stay stable. Closed management years keep their own PDF files.
func musicRefreshLinked618(st *legacyStore, db map[string]any) bool {
	d := musicData550(db)
	templates := mapField(d, "templates")
	keys := map[string]string{}
	for id, raw := range templates {
		keys[id] = musicTemplateLayout618(legacyMap(raw))
	}
	archiveRefs := map[string]bool{}
	musicFileIDs553(mapField(d, "archiveData610"), archiveRefs)
	certs := append([]any{}, arrayField(d, "certificateHistory610")...)
	for _, raw := range mapField(d, "certificates") {
		certs = append(certs, raw)
	}
	changed := false
	for _, raw := range certs {
		c := legacyMap(raw)
		if c == nil {
			continue
		}
		validYear := musicFrom2026_618(stringValue(c["issuedAt"]))
		if eid := stringValue(c["eventId"]); eid != "" {
			event := legacyMap(mapField(d, "events")[eid])
			if event == nil {
				event = legacyMap(c["event"])
			}
			validYear = validYear && musicFrom2026_618(stringValue(event["date"]))
		}
		if !validYear {
			if c["revokedAt"] == nil {
				c["revokedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
				c["revokedReason"] = "Fecha anterior a la gestión 2026"
				changed = true
			}
			continue
		}
		if c["revokedReason"] == "Fecha anterior a la gestión 2026" {
			delete(c, "revokedAt")
			delete(c, "revokedReason")
			c["layoutVersion618"] = "actualizar-fecha"
			if ev := mapField(d, "events")[stringValue(c["eventId"])]; ev != nil {
				c["event"] = cloneAny(ev)
				delete(legacyMap(c["event"]), "participants")
				delete(legacyMap(c["event"]), "externalParticipants")
			}
			changed = true
		}
		id := stringValue(c["templateId610"])
		if id == "" {
			id = stringValue(c["kind"])
		}
		t := legacyMap(templates[id])
		if t == nil {
			continue
		}
		key := keys[id]
		qrChanged := musicSyncQR619(db, c)
		if c["layoutVersion618"] == key && !qrChanged {
			continue
		}

		c["templateId610"] = id
		c["templateVersion610"] = musicTemplateVersion610(d, t)
		c["layoutVersion618"] = key
		if name := strings.TrimSpace(stringValue(t["name"])); name != "" {
			c["title"] = name
		}
		pdf := musicCertificateBytes610(c, musicHydrateFonts613(t, st.State.Files))
		musicUpdateCertificateFile618(st, c, pdf, archiveRefs)
		changed = true
	}
	return changed
}
