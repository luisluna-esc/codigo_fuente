package main

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"net/http"
	"strings"
	"time"
)

// Only the narrow resource in the URL can be modified. Reader roles never submit
// a database snapshot, account fields, report edits, or configuration changes here.
func (s *Server) ministryResource548(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	authRole, ok := s.auth(r)
	if !ok {
		legacyError(w, 401, "Inicia sesión para guardar.", nil)
		return
	}
	var req struct {
		ActorID           string         `json:"actorId"`
		OwnerID           string         `json:"ownerId"`
		BaseUpdatedAt     string         `json:"baseUpdatedAt"`
		RestoreGeneration int64          `json:"restoreGeneration"`
		Value             map[string]any `json:"value"`
	}
	if err := readJSON(r, &req); err != nil || req.OwnerID == "" || req.Value == nil {
		legacyError(w, 400, "Datos inválidos.", nil)
		return
	}
	if sess, remote := s.remoteSessionFromRequest(r); remote {
		if sess.UserID != req.ActorID {
			legacyError(w, 403, "La sesión no corresponde al usuario.", nil)
			return
		}
	} else if authRole != "admin" {
		legacyError(w, 403, "Se requiere una sesión personal autenticada.", nil)
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
	actor := legacyUser(db, req.ActorID)
	owner := legacyUser(db, req.OwnerID)
	if !legacyActive(actor) || owner == nil {
		legacyError(w, 403, "Usuario no autorizado.", nil)
		return
	}
	generation := legacyCanonicalRestoreGeneration(st.State)
	if req.RestoreGeneration != generation && !(req.RestoreGeneration == 0 && generation == 1) {
		legacyError(w, 409, "Actualiza los datos después de la restauración.", map[string]any{"code": "RESTORE_REQUIRED"})
		return
	}
	role := legacyRole(actor)
	key := "musicMinistryRegistration"
	value := map[string]any{}
	if r.URL.Path == "/api/report-watermark" {
		key = "reportWatermarkPolicy"
		if role != "admin" {
			legacyError(w, 403, "Solo Administración configura la marca de agua.", nil)
			return
		}
		if legacyRole(owner) != "aspirant" {
			legacyError(w, 400, "Selecciona un aspirante.", nil)
			return
		}
		enabled, _ := req.Value["enabled"].(bool)
		dataURL := stringValue(req.Value["dataUrl"])
		if enabled || dataURL != "" {
			const prefix = "data:image/png;base64,"
			if !strings.HasPrefix(dataURL, prefix) || len(dataURL) > 280000 {
				legacyError(w, 400, "Usa una imagen PNG de hasta 200 KB.", nil)
				return
			}
			raw, e := base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURL, prefix))
			if e != nil {
				legacyError(w, 400, "PNG inválido.", nil)
				return
			}
			cfg, e := png.DecodeConfig(bytes.NewReader(raw))
			if e != nil || cfg.Width > 4096 || cfg.Height > 4096 {
				legacyError(w, 400, "PNG inválido o demasiado grande.", nil)
				return
			}
		}
		size, _ := req.Value["size"].(float64)
		opacity, _ := req.Value["opacity"].(float64)
		if size < 15 || size > 90 || opacity < 0.05 || opacity > 0.5 {
			legacyError(w, 400, "Tamaño u opacidad inválidos.", nil)
			return
		}
		value = map[string]any{"enabled": enabled, "dataUrl": dataURL, "size": size, "opacity": opacity, "allPages": true}
	} else {
		if !musicPolicies560[role].Registry || personalRequest555(r) || !managementProgram555(db, actor, musicProgram553) {
			legacyError(w, 403, "No puedes registrar este ministerio.", nil)
			return
		}
		if stringValue(req.Value["type"]) == "Otro" && strings.TrimSpace(stringValue(req.Value["otherType"])) == "" {
			legacyError(w, 400, "Especifica el otro tipo de ministerio.", nil)
			return
		}
		if !validMusicRegistration548(db, req.OwnerID, req.Value) {
			legacyError(w, 400, "Completa todos los campos de un aspirante asignado a Música.", nil)
			return
		}
		photo := stringValue(req.Value["photo"])
		if photo != "" && !validMusicPhoto592(photo) {
			legacyError(w, 400, "Foto de ministerio inválida.", nil)
			return
		}
		value["photo"] = photo
		for _, k := range []string{"type", "otherType", "name", "church", "district", "region", "mission", "union", "createdDate", "boardVote", "director", "directorPhone"} {
			value[k] = strings.TrimSpace(stringValue(req.Value[k]))
		}
	}
	if stringValue(legacyMap(owner[key])["updatedAt"]) != req.BaseUpdatedAt {
		legacyError(w, 409, "El registro cambió en otro equipo. Actualiza antes de guardar; tu formulario se conserva.", map[string]any{"code": "RESOURCE_CONFLICT"})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	value["updatedAt"] = now
	value["updatedBy"] = req.ActorID
	owner[key] = value
	st.State.ServerVersion++
	st.State.Database = legacyPrepareCanonicalDB(db, st.State, "Registro o configuración de Ministerio actualizado")
	st.State.UpdatedAt = now
	if err = s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "value": value, "serverVersion": st.State.ServerVersion})
}
