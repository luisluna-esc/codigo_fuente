package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func musicFontManager613(db, actor map[string]any) bool {
	return musicLeaders610(legacyRole(actor)) && musicPermit602(db, actor, "certificates.templates")
}

func musicQRURL613(value string) (string, error) {
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	u, err := url.Parse(value)
	if err != nil || len(value) > 200 || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" || strings.ContainsAny(value, "\r\n\\") || !(u.Scheme == "https" || u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1")) {
		return "", fmt.Errorf("Escribe la dirección pública HTTPS, sin rutas: por ejemplo https://musica.ejemplo.org.")
	}
	return value, nil
}
func musicQRBase613(db map[string]any) string {
	if value, err := musicQRURL613(stringValue(musicData550(db)["publicQRURL613"])); err == nil {
		return value
	}
	return authBase610()
}
func musicQRManager613(actor map[string]any) bool {
	role := legacyRole(actor)
	return role == "admin" || role == "secretaria"
}
func musicFontFile613(files []any, id string) map[string]any {
	for _, raw := range files {
		f := legacyMap(raw)
		if f["id"] == id && f["kind"] == "music_font613" {
			return f
		}
	}
	return musicBundledAsset618(id)
}
func musicFontRaw613(f map[string]any) []byte {
	parts := strings.SplitN(stringValue(f["dataUrl"]), ",", 2)
	if len(parts) != 2 {
		return nil
	}
	data, _ := base64.StdEncoding.DecodeString(parts[1])
	return data
}

// Called before registry expansion so loading a font does not issue documents
// or transfer unrelated registries and certificate artwork.
func (s *Server) musicReadTools613(w http.ResponseWriter, r *http.Request, st legacyStore, db, actor map[string]any) bool {
	resource := r.URL.Query().Get("resource")
	if r.Method != http.MethodGet || resource != "font" && resource != "qr-settings" {
		return false
	}
	if resource == "qr-settings" {
		if !musicQRManager613(actor) || personalRequest555(r) {
			legacyError(w, 403, "Solo Administración y Secretaría pueden configurar la dirección pública.", nil)
			return true
		}
		jsonOut(w, 200, map[string]any{"ok": true, "value": musicData550(db)["publicQRURL613"], "effectiveURL": musicQRBase613(db), "restoreGeneration": legacyCanonicalRestoreGeneration(st.State)})
		return true
	}
	if !musicFontManager613(db, actor) || personalRequest555(r) {
		legacyError(w, 403, "No tienes permiso para cargar fuentes de las plantillas.", nil)
		return true
	}
	f := musicFontFile613(st.State.Files, r.URL.Query().Get("id"))
	data := musicFontRaw613(f)
	if len(data) == 0 {
		legacyError(w, 404, "Fuente no disponible. Restaura el respaldo completo o vuelve a cargarla.", nil)
		return true
	}
	w.Header().Set("Content-Type", stringValue(f["type"]))
	// IDs are SHA-256 hashes of immutable font bytes. Authentication is still
	// checked before conditional responses, including when using a browser cache.
	etag := `"` + r.URL.Query().Get("id") + `"`
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	w.WriteHeader(200)
	_, _ = w.Write(data)
	return true
}

func (s *Server) musicWriteTools613(w http.ResponseWriter, r *http.Request, st *legacyStore, db, actor map[string]any, req musicRequest610) bool {
	if req.Action != "font.save" && req.Action != "qr.configure" {
		return false
	}
	result := map[string]any{"ok": true, "restoreGeneration": legacyCanonicalRestoreGeneration(st.State)}
	if req.Action == "qr.configure" {
		if !musicQRManager613(actor) {
			legacyError(w, 403, "Solo Administración y Secretaría pueden configurar la dirección pública.", nil)
			return true
		}
		value := strings.TrimSpace(stringValue(req.Value["url"]))
		if value != "" {
			var err error
			value, err = musicQRURL613(value)
			if err != nil {
				legacyError(w, 400, err.Error(), nil)
				return true
			}
		}
		musicData550(db)["publicQRURL613"] = value
		result["effectiveURL"] = musicQRBase613(db)
		result["value"] = value
	} else {
		if !musicFontManager613(db, actor) {
			legacyError(w, 403, "No tienes permiso para administrar fuentes.", nil)
			return true
		}
		raw := musicFontRaw613(req.Value)
		font, err := parseMusicSFNT613(raw)
		if err != nil {
			legacyError(w, 400, err.Error(), nil)
			return true
		}
		hash := sha256.Sum256(raw)
		id := "font613_" + hex.EncodeToString(hash[:])
		result["id"] = id
		if musicFontFile613(st.State.Files, id) == nil {
			mime := "font/ttf"
			if font.openType {
				mime = "font/otf"
			}
			st.State.Files = append(st.State.Files, map[string]any{"id": id, "kind": "music_font613", "ownerId": actor["id"], "programId": musicProgram553, "name": font.Name, "size": len(raw), "type": mime, "dataUrl": "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(raw), "createdAt": time.Now().UTC().Format(time.RFC3339Nano)})
		}
	}
	st.State.Database = db
	st.State.ServerVersion++
	st.State.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err := s.saveLegacy(*st); err != nil {
		legacyError(w, 500, "No se pudo guardar. Intenta nuevamente.", nil)
		return true
	}
	jsonOut(w, 200, result)
	return true
}

func musicValidateFonts613(t map[string]any, files []any) error {
	ids := map[string]bool{}
	musicFileIDs553(t, ids)
	total, count := 0, 0
	for id := range ids {
		if !strings.HasPrefix(id, "font613_") {
			continue
		}
		f := musicFontFile613(files, id)
		if f == nil {
			return fmt.Errorf("La fuente seleccionada no está disponible. Vuelve a cargarla antes de guardar.")
		}
		total += len(stringValue(f["dataUrl"]))
		count++
	}
	if count > 12 || total > 24<<20 {
		return fmt.Errorf("Usa hasta 12 variantes de fuente por plantilla (18 MB en total).")
	}
	return nil
}

// Hydrated data is transient and is never duplicated in template history.
func musicHydrateFonts613(t map[string]any, files []any) map[string]any {
	copy := cloneMap610(t)
	musicBindFonts618(copy)
	ids := map[string]bool{}
	musicFileIDs553(copy, ids)
	assets := map[string]any{}
	for id := range ids {
		if f := musicFontFile613(files, id); f != nil {
			assets[id] = f["dataUrl"]
		}
	}
	copy["fontPrograms613"] = assets
	return copy
}
