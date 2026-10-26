package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Existing templates retain their exact A4 coordinates. Carta and Oficio use
// the dimensions requested by the ministry, in landscape orientation.
func musicPaper617(t map[string]any) (float64, float64) {
	switch stringValue(t["paperSize617"]) {
	case "a4":
		return 842, 595
	case "oficio":
		return 330 * 72 / 25.4, 216 * 72 / 25.4
	default:
		return 280 * 72 / 25.4, 216 * 72 / 25.4
	}
}

func musicPaperSettings617(v, out map[string]any) error {
	size := stringValue(v["paperSize617"])
	if size == "" {
		size = "carta"
	}
	if size != "a4" && size != "carta" && size != "oficio" {
		return fmt.Errorf("Selecciona A4, Carta u Oficio.")
	}
	out["paperSize617"] = size
	if v["showQR617"] != nil && v["showQR617"] != true && v["showQR617"] != false {
		return fmt.Errorf("Revisa la opción Mostrar QR.")
	}
	out["showQR617"] = v["showQR617"] != false
	return nil
}

func musicPreviewCertificate617(t map[string]any) map[string]any {
	now := time.Now().UTC()
	day := now.Format("2006-01-02")
	c := map[string]any{
		"kind": stringValue(t["certificateType"]), "title": stringValue(t["name"]),
		"fullName": "Nombre del aspirante", "issuedAt": now.Format(time.RFC3339), "eventId": "",
		"church": "Iglesia", "mission": "Misión", "token": "VISTA-PREVIA-SIN-EMISION",
		"registration": map[string]any{"registrationCode622": "MBOS-MMUS-0001", "type": "Dúo", "name": "Emanuel", "boardVote": "001", "church": "Iglesia", "district": "Centro", "region": "1", "mission": "Misión", "union": "Unión", "director": "Nombre del director", "createdDate": day},
		"event":        map[string]any{"name": "Evento de ejemplo", "place": "Lugar del certificado", "date": day},
		"investiture":  map[string]any{"date": day},
		// A sample QR verifies the layout without issuing a live certificate.
		"verificationURL610": "https://example.org/vista-previa",
	}
	if c["kind"] == "event" {
		c["eventId"] = "vista-previa-evento"
	}
	if c["kind"] == "ministry" {
		c["fullName"] = "Emanuel"
	}
	return c
}

// Preview reads do not save templates, issue certificates, or modify progress.
func (s *Server) musicPreview617(w http.ResponseWriter, r *http.Request, st legacyStore, db, actor map[string]any, req musicRequest610) bool {
	get := r.Method == http.MethodGet && r.URL.Query().Get("resource") == "template-preview"
	post := r.Method == http.MethodPost && req.Action == "template.preview"
	if !get && !post {
		return false
	}
	if !musicFontManager613(db, actor) || personalRequest555(r) {
		legacyError(w, 403, "No tienes permiso para previsualizar plantillas.", nil)
		return true
	}
	var template map[string]any
	if get {
		template = legacyMap(mapField(musicData550(db), "templates")[r.URL.Query().Get("id")])
		if template == nil {
			legacyError(w, 404, "Plantilla no encontrada.", nil)
			return true
		}
	} else {
		if req.RestoreGeneration != legacyCanonicalRestoreGeneration(st.State) {
			legacyError(w, 409, "Actualiza la pantalla después de la restauración o nueva gestión.", nil)
			return true
		}
		// Validate in a small detached registry. Unrelated users, reports,
		// archived years and other template artwork are never cloned.
		copyDB := map[string]any{}
		for key, item := range db {
			copyDB[key] = item
		}
		copyMusic := map[string]any{}
		for key, item := range mapField(db, "music550") {
			copyMusic[key] = item
		}
		copyMusic["templates"] = map[string]any{}
		copyDB["music550"] = copyMusic
		value := cloneMap610(req.Value)
		value["id"] = ""
		if strings.TrimSpace(stringValue(value["name"])) == "" {
			value["name"] = "Vista previa de la plantilla"
		}
		if err := musicValidateFonts613(value, st.State.Files); err != nil {
			legacyError(w, 400, err.Error(), nil)
			return true
		}
		if err := musicTemplateMutation610(copyDB, actor, musicRequest610{Action: "template.save", Value: value}); err != nil {
			legacyError(w, 400, err.Error(), nil)
			return true
		}
		for _, raw := range mapField(musicData550(copyDB), "templates") {
			template = legacyMap(raw)
			break
		}
	}
	pdf := musicCertificateBytes610(musicPreviewCertificate617(template), musicHydrateFonts613(template, st.State.Files))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", musicDisposition618(musicFilename618(template, true)))
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(200)
	_, _ = w.Write(pdf)
	return true
}
