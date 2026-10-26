package main

// Standalone online product. The former application's review, editor and music
// policy code is retained; public routes and persistent data are music-only.
import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const musicOnlineVersion = "6.2.6-musica-online"

// Removed integrations have no place in a standalone Music database. In
// particular, a remote provider's fileId is not a local attachment identifier.
func removedMusicIntegrationKey(key string) bool {
	lower := strings.ToLower(key)
	if strings.Contains(lower, "watermark") || strings.HasPrefix(lower, "interfacebackground") || lower == "interfacepanelopacity" {
		return true
	}
	for _, part := range []string{"drive", "google", "oauth", "tailscale", "zerotier"} {
		if strings.Contains(lower, part) {
			return true
		}
	}
	return false
}

func cleanMusicIntegrations(value any) {
	switch v := value.(type) {
	case map[string]any:
		for key, child := range v {
			if removedMusicIntegrationKey(key) {
				delete(v, key)
				continue
			}
			cleanMusicIntegrations(child)
		}
	case []any:
		for _, child := range v {
			cleanMusicIntegrations(child)
		}
	}
}

func musicRemovedAccount(user map[string]any) bool {
	state := strings.ToLower(stringValue(user["accountState"]))
	return user["musicAccessRemoved"] == true || user["deleted"] == true || stringValue(user["deletedAt"]) != "" || state == "deleted" || state == "removed"
}

// Elimination is different from temporarily deactivating an account. Preserve
// inactive members, but never resurrect an eliminated member through old progress.
func pruneRemovedMusicOwners(value any, removed map[string]bool) any {
	switch v := value.(type) {
	case map[string]any:
		for _, field := range []string{"ownerId", "assigneeId", "aspirantId", "userId"} {
			if removed[stringValue(v[field])] {
				return nil
			}
		}
		for key, child := range v {
			ownerKey := strings.SplitN(key, ":", 2)[0]
			if removed[ownerKey] {
				delete(v, key)
				continue
			}
			clean := pruneRemovedMusicOwners(child, removed)
			if clean == nil && child != nil {
				delete(v, key)
			} else {
				v[key] = clean
			}
		}
		return v
	case []any:
		out := []any{}
		for _, child := range v {
			if id, ok := child.(string); ok && removed[id] {
				continue
			}
			clean := pruneRemovedMusicOwners(child, removed)
			if clean != nil || child == nil {
				out = append(out, clean)
			}
		}
		return out
	default:
		return value
	}
}

func (s *Server) cleanStandaloneMusicStore(st *legacyStore) {
	db := legacyMap(st.State.Database)
	if !strings.HasPrefix(stringValue(mapField(db, "musicProduct")["id"]), "ministerio-musica") {
		return
	}
	clean, err := extractMusicDatabase(db)
	if err != nil {
		return
	}
	st.State.Database = clean
	ids := map[string]bool{}
	musicFileIDs553(clean, ids)
	files := []any{}
	for _, raw := range st.State.Files {
		file := legacyMap(raw)
		// Upload and report save are separate requests. Keep new files belonging
		// to existing members until they can be linked; exports still include
		// only referenced files. An eliminated owner's exclusive files are purged.
		owner := stringValue(file["ownerId"])
		pendingUpload := legacyUser(clean, owner) != nil && (stringValue(file["programId"]) == musicProgram553 || strings.HasPrefix(stringValue(file["requirementId"]), "LM-"))
		if ids[stringValue(file["id"])] || pendingUpload {
			cleanMusicIntegrations(file)
			files = append(files, raw)
		}
	}
	st.State.Files = files
}

func musicFileBelongs(db map[string]any, file map[string]any, referenced map[string]bool) bool {
	return referenced[stringValue(file["id"])]
}

func cleanMusicSettings(settings map[string]any) {
	for k := range settings {
		lower := strings.ToLower(k)
		if strings.Contains(lower, "google") || strings.Contains(lower, "drive") || strings.Contains(lower, "tailscale") || strings.Contains(lower, "zerotier") || strings.Contains(lower, "offline") || strings.Contains(lower, "recovery") || strings.Contains(lower, "remoteaccess") || k == "serverToken" || k == "serverUrl" || k == "nameAutocompleteHistory" {
			delete(settings, k)
		}
	}
	settings["autoSync"] = true // online write transport, never an offline option
}

// This is an export projection, not a role-filtered screen snapshot: inactive
// members and their historical evidence must survive migration too.
func extractMusicDatabase(source any) (map[string]any, error) {
	db := legacyMap(cloneAny(source))
	if db == nil {
		return nil, errors.New("El respaldo no contiene una base de datos válida.")
	}
	cleanMusicIntegrations(db)
	removed := map[string]bool{}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if musicRemovedAccount(u) {
			removed[stringValue(u["id"])] = true
		}
	}
	pruneRemovedMusicOwners(db, removed)
	keep := map[string]bool{}
	assigned := map[string]bool{}
	modules := []any{}
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if stringValue(m["id"]) == musicProgram553 {
			modules = append(modules, m)
			for _, id := range arrayField(m, "assignedUserIds") {
				if legacyUser(db, stringValue(id)) != nil && !removed[stringValue(id)] {
					keep[stringValue(id)] = true
					assigned[stringValue(id)] = true
				}
			}
		}
	}
	progress := map[string]any{}
	for id, raw := range legacyMap(db["programProgress"]) {
		if removed[id] || legacyUser(db, id) == nil {
			continue
		}
		if p, ok := legacyMap(raw)[musicProgram553]; ok {
			progress[id] = map[string]any{musicProgram553: p}
			keep[id] = true
			assigned[id] = true
		}
	}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		id := stringValue(u["id"])
		if !removed[id] && (legacyRole(u) == "admin" || musicAllowed553(db, u)) {
			keep[id] = true
		}
	}
	// Keep the exact tutoring chain for every current and former music member.
	for changed := true; changed; {
		changed = false
		for _, raw := range legacyUsers(db) {
			u := legacyMap(raw)
			if !keep[stringValue(u["id"])] {
				continue
			}
			v := legacyMap(u["tutorByProgram"])[musicProgram553]
			ids := anySlice548(v)
			if text, ok := v.(string); ok {
				ids = []any{text}
			}
			if v == nil && u["tutorId"] != nil {
				ids = append(ids, u["tutorId"])
			}
			for _, x := range ids {
				id := stringValue(x)
				if id != "" && !keep[id] && !removed[id] && legacyUser(db, id) != nil {
					keep[id] = true
					changed = true
				}
			}
		}
	}
	users := []any{}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		id := stringValue(u["id"])
		if !keep[id] {
			continue
		}
		for _, key := range []string{"tutorByProgram", "tutorPrograms", "programTutors"} {
			if m := legacyMap(u[key]); m != nil {
				out := map[string]any{}
				if value, ok := m[musicProgram553]; ok {
					out[musicProgram553] = value
				}
				u[key] = out
			}
		}
		for _, key := range []string{"tutorProgramIds", "programIds", "assignedProgramIds", "requirementViewPrograms", "regionalProgramIds", "managementProgramIds", "importedProgramIds"} {
			if value, ok := u[key]; ok {
				list := []any{}
				if containsString(anySlice548(value), musicProgram553) {
					list = append(list, musicProgram553)
				}
				u[key] = list
			}
		}
		for _, key := range []string{"clubId", "clubPosition", "classTeaching", "drive", "google", "googleId", "recoveryKeys"} {
			delete(u, key)
		}
		users = append(users, u)
	}
	if len(modules) == 0 {
		ids := []any{}
		for _, raw := range users {
			u := legacyMap(raw)
			if assigned[stringValue(u["id"])] {
				ids = append(ids, u["id"])
			}
		}
		modules = append(modules, map[string]any{"id": musicProgram553, "active": true, "assignedUserIds": ids})
	}
	db["users"] = users
	db["modules"] = modules
	db["programProgress"] = progress
	db["classes"] = []any{}
	db["clubs"] = []any{}
	db["progress"] = map[string]any{}
	db["bibleYear"] = map[string]any{}
	for _, key := range []string{"configs", "reportRequirementModes"} {
		out := map[string]any{}
		for k, v := range legacyMap(db[key]) {
			if musicKey553(k) {
				out[k] = v
			}
		}
		db[key] = out
	}
	for _, key := range []string{"tasks", "materials", "indicationRules"} {
		out := []any{}
		for _, raw := range arrayField(db, key) {
			if musicItem553(raw) {
				out = append(out, raw)
			}
		}
		db[key] = out
	}
	audit := []any{}
	for _, raw := range arrayField(db, "audit") {
		m := legacyMap(raw)
		if musicKey553(stringValue(m["entity"])) || stringValue(m["programId"]) == musicProgram553 {
			audit = append(audit, raw)
		}
	}
	db["audit"] = audit
	for _, key := range []string{"syncQueue", "conflicts", "errors", "recovery", "recovery548", "musicEmptyArchive597", "google", "drive", "__sglAdminBaseline"} {
		delete(db, key)
	}
	db["syncQueue"] = []any{}
	db["conflicts"] = []any{}
	db["errors"] = []any{}
	settings := ensureMapField(db, "settings")
	cleanMusicSettings(settings)
	if modes := legacyMap(settings["reportRequirementModes"]); modes != nil {
		for k := range modes {
			if !musicKey553(k) {
				delete(modes, k)
			}
		}
	}
	music := ensureMapField(db, "music550")
	// Disable all removed integration permissions, including imported overrides.
	if cfg := legacyMap(music["permissions602"]); cfg != nil {
		for _, scope := range []string{"roles", "users"} {
			for _, row := range legacyMap(cfg[scope]) {
				m := legacyMap(row)
				for key := range m {
					if strings.HasPrefix(key, "drive.") || key == "menu.drive" {
						m[key] = false
					}
				}
			}
		}
	}
	// Keep only the canonical music schema. Browser caches and unknown fields
	// from the multi-program product must not carry a hidden SGL snapshot.
	allowed := map[string]bool{}
	for _, key := range []string{"schemaVersion", "version", "revision", "serverVersion", "installationId", "createdAt", "updatedAt", "restoreGeneration", "restoreId", "restoredAt", "users", "modules", "programProgress", "progress", "classes", "clubs", "bibleYear", "settings", "configs", "reportRequirementModes", "indicationRules", "tasks", "materials", "bibleVersions", "audit", "music550", "musicProduct", "syncQueue", "conflicts", "errors"} {
		allowed[key] = true
	}
	for key := range db {
		if !allowed[key] {
			delete(db, key)
		}
	}
	cleanMusicIntegrations(db)
	db["version"] = musicOnlineVersion
	db["musicProduct"] = map[string]any{"id": "ministerio-musica-online", "version": musicOnlineVersion}
	return db, nil
}

func decodeMusicBackup(data []byte, password string) (map[string]any, []any, error) {
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, nil, errors.New("Selecciona un respaldo JSON o una copia protegida .cabackup; el ZIP del programa no contiene los registros.")
	}
	if root["ciphertext"] != nil {
		if password == "" {
			return nil, nil, errors.New("Escribe la contraseña del respaldo protegido.")
		}
		salt, e1 := base64.StdEncoding.DecodeString(stringValue(root["salt"]))
		iv, e2 := base64.StdEncoding.DecodeString(stringValue(root["iv"]))
		encrypted, e3 := base64.StdEncoding.DecodeString(stringValue(root["ciphertext"]))
		iterations := 180000
		if n, ok := root["iterations"].(float64); ok {
			iterations = int(n)
		}
		if e1 != nil || e2 != nil || e3 != nil || iterations < 1000 || iterations > 2000000 || len(salt) < 8 {
			return nil, nil, errors.New("El respaldo cifrado está dañado o usa un formato no compatible.")
		}
		block, _ := aes.NewCipher(pbkdf2SHA256([]byte(password), salt, iterations, 32))
		gcm, _ := cipher.NewGCM(block)
		if len(iv) != gcm.NonceSize() {
			return nil, nil, errors.New("El respaldo cifrado está dañado.")
		}
		plain, err := gcm.Open(nil, iv, encrypted, nil)
		if err != nil {
			return nil, nil, errors.New("La contraseña del respaldo no es correcta o el archivo está dañado.")
		}
		return decodeMusicBackup(plain, "")
	}
	candidate := root
	files := arrayField(root, "files")
	if state := legacyMap(root["state"]); state != nil {
		candidate = state
		files = arrayField(state, "files")
	}
	if db := legacyMap(candidate["database"]); db != nil {
		candidate = db
	}
	if _, ok := candidate["users"]; !ok {
		return nil, nil, errors.New("La copia no contiene usuarios. Exporta un respaldo completo desde la cuenta del Administrador del sistema anterior.")
	}
	db, err := extractMusicDatabase(candidate)
	if err != nil {
		return nil, nil, err
	}
	ids := map[string]bool{}
	musicFileIDs553(db, ids)
	out := []any{}
	seen := map[string]bool{}
	for _, raw := range files {
		f := legacyMap(raw)
		id := stringValue(f["id"])
		// Only current references are migrated. Old versions and orphan files
		// cannot silently reappear just because their former owner still exists.
		if id != "" && ids[id] && !seen[id] {
			cleanMusicIntegrations(f)
			out = append(out, raw)
			seen[id] = true
		}
	}
	return db, out, nil
}

func musicSummary(db map[string]any, files []any) map[string]any {
	owners := len(legacyMap(db["programProgress"]))
	approved, inReview, withContent := 0, 0, 0
	for _, raw := range legacyMap(db["programProgress"]) {
		p := legacyMap(legacyMap(raw)[musicProgram553])
		hasContent := false
		for _, state := range legacyMap(p["status"]) {
			value := stringValue(state)
			if value == "in_review" {
				inReview++
			}
			if value != "" && value != "pending" {
				hasContent = true
			}
			if stringValue(state) == "approved" {
				approved++
			}
		}
		if !hasContent {
			for _, rawSub := range mapField(p, "submissions") {
				sub := legacyMap(rawSub)
				if strings.TrimSpace(stringValue(sub["text"])) != "" || len(arrayField(sub, "attachmentIds")) > 0 || stringValue(sub["primaryDocumentId"]) != "" {
					hasContent = true
					break
				}
			}
		}
		if hasContent {
			withContent++
		}
	}
	approvedOwners := []any{}
	for id, raw := range legacyMap(db["programProgress"]) {
		count := 0
		for _, status := range mapField(legacyMap(legacyMap(raw)[musicProgram553]), "status") {
			if stringValue(status) == "approved" {
				count++
			}
		}
		if count > 0 {
			u := legacyUser(db, id)
			approvedOwners = append(approvedOwners, map[string]any{"name": stringValue(u["fullName"]), "requirements": count})
		}
	}
	return map[string]any{"users": len(legacyUsers(db)), "records": owners, "files": len(files), "approvedRequirements": approved, "approvedOwners": approvedOwners, "inReviewRequirements": inReview, "recordsWithContent": withContent, "recordsOnlyPending": owners - withContent, "usersWithoutRecord": len(legacyUsers(db)) - owners}
}

func musicMissingFiles(db map[string]any, files []any) []string {
	ids := map[string]bool{}
	musicFileIDs553(db, ids)
	for _, raw := range files {
		f := legacyMap(raw)
		// The Bible library stores structured verses instead of a data URL.
		present := len(legacyMap(f["verses"])) > 0
		data := stringValue(f["dataUrl"])
		if strings.HasPrefix(data, "data:") {
			if at := strings.Index(data, ","); at > 0 && strings.Contains(data[:at], ";base64") {
				decoded, err := base64.StdEncoding.DecodeString(data[at+1:])
				present = err == nil && len(decoded) > 0
			}
		}
		if present {
			delete(ids, stringValue(f["id"]))
		}
	}
	missing := []string{}
	for id := range ids {
		if id != "" {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	return missing
}

func (s *Server) installationKeyPath() string {
	return filepath.Join(filepath.Dir(s.configPath), "CLAVE_INSTALACION.txt")
}
func (s *Server) prepareInstallation() error {
	s.mu.Lock()
	st, err := s.loadLegacy()
	s.mu.Unlock()
	if err != nil {
		return err
	}
	if len(legacyUsers(st.State.Database)) > 0 {
		return nil
	}
	if _, err = os.Stat(s.installationKeyPath()); os.IsNotExist(err) {
		return os.WriteFile(s.installationKeyPath(), []byte(token(24)+"\n"), 0600)
	}
	return err
}
func (s *Server) setupAuthorized(r *http.Request, key string) bool {
	b, err := os.ReadFile(s.installationKeyPath())
	return err == nil && key != "" && subtle.ConstantTimeCompare([]byte(strings.TrimSpace(string(b))), []byte(strings.TrimSpace(key))) == 1
}
func (s *Server) onlineProduct(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		if (r.Method == http.MethodGet || r.Method == http.MethodHead) && r.URL.Query().Get("v") == musicOnlineVersion && (strings.HasSuffix(r.URL.Path, ".js") || strings.HasSuffix(r.URL.Path, ".css")) {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		path := strings.TrimPrefix(r.URL.Path, musicAPI553)
		if path == "/api/server/stop" {
			host, _, _ := net.SplitHostPort(r.RemoteAddr)
			ip := net.ParseIP(host)
			secret := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if r.Method != http.MethodPost || ip == nil || !ip.IsLoopback() || r.Header.Get("Origin") != "" || s.stop == nil || s.cfg.AdminToken == "" || subtle.ConstantTimeCompare([]byte(secret), []byte(s.cfg.AdminToken)) != 1 {
				http.NotFound(w, r)
				return
			}
			jsonOut(w, 200, map[string]any{"ok": true})
			select {
			case s.stop <- struct{}{}:
			default:
			}
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Header.Get("Origin") != "" && !s.originAllowed(r, r.Header.Get("Origin")) {
			legacyError(w, 403, "Origen no autorizado.", nil)
			return
		}
		if path == "/api/report-watermark" {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(path, "/api/auth610/") {
			s.musicAuth610(w, r, path)
			return
		}
		if strings.HasPrefix(path, "/api/google/") || strings.HasPrefix(path, "/api/reports/drive/") || strings.Contains(path, "recovery") || strings.Contains(path, "reconcile598") || strings.HasPrefix(path, "/api/v4/") || path == "/api/sync/diagnostics" || path == "/sw.js" || path == "/offline-assets.json" || path == "/manifest.webmanifest" {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(path, "/api/installation/") {
			s.installationEndpoint(w, r, path)
			return
		}
		if path == "/api/music-data/import" || path == "/api/music-data/export" || path == "/api/music-data/photos" {
			s.musicDataEndpoint(w, r, path)
			return
		}
		if r.URL.Path == "/musica" || strings.HasPrefix(r.URL.Path, "/musica/") {
			target := strings.TrimPrefix(r.URL.Path, "/musica")
			if target == "" {
				target = "/"
			}
			if r.URL.RawQuery != "" {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusTemporaryRedirect)
			return
		}
		// Every API request belongs to the music portal; there is no SGL entry.
		if strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, musicAPI553+"/") {
			r = r.Clone(r.Context())
			urlCopy := *r.URL
			r.URL = &urlCopy
			r.URL.Path = musicAPI553 + r.URL.Path
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) installationEndpoint(w http.ResponseWriter, r *http.Request, path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, "No se pudo abrir la base de música.", nil)
		return
	}
	empty := len(legacyUsers(st.State.Database)) == 0
	if path == "/api/installation/status" && r.Method == http.MethodGet {
		result := map[string]any{"ok": true, "setupRequired": empty, "version": musicOnlineVersion}
		if empty {
			if key := s.localSetupKey624(r); key != "" {
				result["localSetupKey624"] = key
			}
		}
		jsonOut(w, 200, result)
		return
	}
	if path != "/api/installation/create" || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if !empty {
		legacyError(w, 409, "La instalación ya tiene usuarios. Ingresa con tu cuenta.", nil)
		return
	}
	var req struct{ Key, FullName, Username, Email, Password string }
	if readJSON(r, &req) != nil || !s.setupAuthorized(r, req.Key) {
		legacyError(w, 403, "La clave de instalación no es válida. Consúltala en CLAVE_INSTALACION.txt, en el servidor.", nil)
		return
	}
	if len(req.Password) < 10 || len(req.Password) > 256 || strings.TrimSpace(req.FullName) == "" || len(req.Username) < 3 {
		legacyError(w, 400, "Completa el nombre, un usuario de al menos 3 caracteres y una contraseña de 10 a 256 caracteres.", nil)
		return
	}
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	rec := map[string]any{"salt": base64.StdEncoding.EncodeToString(salt), "hash": base64.StdEncoding.EncodeToString(pbkdf2SHA256([]byte(req.Password), salt, 150000, 32)), "iterations": 150000, "algorithm": "PBKDF2-SHA256"}
	db := map[string]any{"schemaVersion": 6, "version": musicOnlineVersion, "revision": 1, "installationId": "music_" + token(12), "createdAt": time.Now().UTC().Format(time.RFC3339Nano), "users": []any{map[string]any{"id": "admin_" + token(10), "fullName": strings.TrimSpace(req.FullName), "username": strings.ToLower(strings.TrimSpace(req.Username)), "email": strings.ToLower(strings.TrimSpace(req.Email)), "role": "admin", "active": true, "accountState": "active", "password": rec, "managementProgramIds": []any{musicProgram553}}}, "modules": []any{map[string]any{"id": musicProgram553, "active": true, "assignedUserIds": []any{}}}, "programProgress": map[string]any{}, "settings": map[string]any{}, "tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "bibleVersions": []any{}, "audit": []any{}}
	db, _ = extractMusicDatabase(db)
	designDB, designFonts := musicResetDesign619(legacyMap(st.State.Database), st.State.Files)
	if len(mapField(musicData550(designDB), "templates")) > 0 {
		db["music550"] = designDB["music550"]
	}
	musicEnsure610(db)
	st.State = legacyState{ServerVersion: st.State.ServerVersion + 1, RestoreGeneration: legacyCanonicalRestoreGeneration(st.State), RestoreID: st.State.RestoreID, RestoredAt: st.State.RestoredAt, Database: db, Files: designFonts, UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	if err = s.saveLegacy(st); err != nil {
		legacyError(w, 500, "No se pudo guardar la cuenta.", nil)
		return
	}
	_ = os.Remove(s.installationKeyPath())
	jsonOut(w, 200, map[string]any{"ok": true})
}

func (s *Server) musicDataEndpoint(w http.ResponseWriter, r *http.Request, path string) {
	if path == "/api/music-data/export" {
		if r.Method != http.MethodGet {
			http.Error(w, "Método no permitido", 405)
			return
		}
		role, ok := s.auth(r)
		if !ok || role != "admin" {
			legacyError(w, 403, "Solo el Administrador puede exportar la base completa de música.", nil)
			return
		}
		s.mu.Lock()
		st, err := s.loadLegacy()
		if err != nil {
			s.mu.Unlock()
			legacyError(w, 500, "No se pudo abrir el respaldo.", nil)
			return
		}
		db, _ := extractMusicDatabase(st.State.Database)
		ids := map[string]bool{}
		musicFileIDs553(db, ids)
		files := []any{}
		for _, f := range st.State.Files {
			if musicFileBelongs(db, legacyMap(f), ids) {
				files = append(files, f)
			}
		}
		s.mu.Unlock()
		w.Header().Set("Content-Disposition", "attachment; filename=Musica_respaldo_"+time.Now().Format("2006-01-02")+".json")
		jsonOut(w, 200, map[string]any{"format": "MinisterioMusicaOnlineBackup", "version": musicOnlineVersion, "exportedAt": time.Now().UTC().Format(time.RFC3339Nano), "database": db, "files": files})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", 405)
		return
	}
	if path == "/api/music-data/photos" {
		s.optimizeMusicPhotos(w, r)
		return
	}
	req, err := readMusicImport625(w, r, musicImportFileLimit625)
	if err != nil {
		legacyError(w, 400, err.Error(), nil)
		return
	}
	role, auth := s.auth(r)
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, "No se pudo abrir la base.", nil)
		return
	}
	empty := len(legacyUsers(st.State.Database)) == 0
	if !(auth && role == "admin") && !(empty && s.setupAuthorized(r, req.Key)) {
		legacyError(w, 403, "Ingresa como Administrador para importar. En una instalación nueva utiliza su clave de instalación.", nil)
		return
	}
	db, files, err := decodeMusicBackup(req.Payload, req.Password)
	if err != nil {
		legacyError(w, 400, err.Error(), nil)
		return
	}
	// A partial personal backup must never replace the server database.
	if !restoredDatabaseHasActiveAdmin(db) {
		legacyError(w, 400, "La copia debe ser completa e incluir un Administrador activo con contraseña. Exporta desde el Administrador del sistema anterior.", nil)
		return
	}
	validAdmin := false
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if legacyRole(u) == "admin" && legacyActive(u) && legacyMap(u["password"])["hash"] != nil {
			validAdmin = true
		}
	}
	if !validAdmin {
		legacyError(w, 400, "El respaldo no conserva la contraseña del Administrador. Usa un respaldo completo del servidor o crea primero una cuenta y exporta una copia completa.", nil)
		return
	}
	summary := musicSummary(db, files)
	missing := musicMissingFiles(db, files)
	if req.Preview {
		jsonOut(w, 200, map[string]any{"ok": true, "summary": summary, "missingFiles": missing, "replacesCurrent": !empty})
		return
	}
	if req.Confirmation != "IMPORTAR MUSICA" {
		legacyError(w, 400, "Confirma escribiendo IMPORTAR MUSICA.", nil)
		return
	}
	if len(missing) > 0 {
		legacyError(w, 400, "Faltan archivos referenciados en la copia. Crea un respaldo completo con los adjuntos antes de importar.", map[string]any{"missingFiles": missing})
		return
	}
	if !empty {
		if err = s.savePreRestoreSafety(st, time.Now()); err != nil {
			legacyError(w, 500, "No se pudo crear el respaldo de seguridad previo. No se modificó la base.", nil)
			return
		}
	}
	generation := legacyCanonicalRestoreGeneration(st.State) + 1
	st.State = legacyState{ServerVersion: st.State.ServerVersion + 1, RestoreGeneration: generation, RestoreID: "music_import_" + token(12), RestoredAt: time.Now().UTC().Format(time.RFC3339Nano), Database: db, Files: files, UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	st.Proposals = []legacyProposal{}
	st.Messages = []legacyMessage{}
	st.Notifications = []legacyNotification{}
	st.UserChanges = []legacyUserChangeRequest{}
	st.AppliedOperations = []legacyAppliedOperation{}
	st.Recoveries548 = nil
	var complete struct {
		Store     json.RawMessage `json:"store610"`
		Transport *Store          `json:"transport610"`
	}
	if json.Unmarshal(req.Payload, &complete) == nil && len(complete.Store) > 0 {
		var extra legacyStore
		if json.Unmarshal(complete.Store, &extra) == nil {
			st.Proposals = extra.Proposals
			st.Messages = extra.Messages
			st.Notifications = extra.Notifications
			st.UserChanges = extra.UserChanges
			st.AppliedOperations = extra.AppliedOperations
			st.Recoveries548 = extra.Recoveries548
			if complete.Transport != nil {
				s.store.Entities = complete.Transport.Entities
				s.store.Pending = complete.Transport.Pending
				s.store.Audit = complete.Transport.Audit
				s.store.Revision++
				if err = s.save(); err != nil {
					legacyError(w, 500, "No se pudo restaurar el almacenamiento auxiliar.", nil)
					return
				}
			}
		}
	}
	if err = s.saveLegacy(st); err != nil {
		legacyError(w, 500, "No se pudo guardar la importación.", nil)
		return
	}
	s.publishSystemRestoredControl("Se importó la base de música. Vuelve a ingresar con las cuentas importadas.", st.State, time.Now())
	_ = os.Remove(s.installationKeyPath())
	jsonOut(w, 200, map[string]any{"ok": true, "summary": summary, "loginRequired": true})
}

func (s *Server) optimizeMusicPhotos(w http.ResponseWriter, r *http.Request) {
	role, ok := s.auth(r)
	if !ok || role != "admin" {
		legacyError(w, 403, "Solo el Administrador puede reducir fotos.", nil)
		return
	}
	var req struct {
		Files []map[string]any `json:"files"`
	}
	if readJSONLimited(r, &req, 128<<20) != nil || len(req.Files) > 300 {
		legacyError(w, 400, "Carga inválida; usa lotes de hasta 300 fotos.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, "No se pudo abrir la base.", nil)
		return
	}
	ids := map[string]bool{}
	musicFileIDs553(st.State.Database, ids)
	for _, raw := range st.State.Files {
		file := legacyMap(raw)
		if musicFileBelongs(legacyMap(st.State.Database), file, ids) {
			ids[stringValue(file["id"])] = true
		}
	}
	replacements := map[string]map[string]any{}
	for _, f := range req.Files {
		id := stringValue(f["id"])
		url := stringValue(f["dataUrl"])
		if !ids[id] || !strings.HasPrefix(url, "data:image/jpeg;base64,") {
			legacyError(w, 400, "Foto no válida o ajena a Música.", nil)
			return
		}
		data, e := base64.StdEncoding.DecodeString(strings.TrimPrefix(url, "data:image/jpeg;base64,"))
		if e != nil {
			legacyError(w, 400, "Foto JPEG no válida.", nil)
			return
		}
		cfg, _, e := image.DecodeConfig(bytes.NewReader(data))
		if e != nil || cfg.Width > 2400 || cfg.Height > 2400 || cfg.Width < 1 || cfg.Height < 1 {
			legacyError(w, 400, "Dimensiones de foto no válidas.", nil)
			return
		}
		f["size"] = len(data)
		f["width"] = cfg.Width
		f["height"] = cfg.Height
		replacements[id] = f
	}
	files := cloneAny(st.State.Files).([]any)
	count := 0
	for i, raw := range files {
		old := legacyMap(raw)
		next := replacements[stringValue(old["id"])]
		if next == nil {
			continue
		}
		if stringValue(old["kind"]) != "photo" {
			legacyError(w, 400, "Solo se pueden comprimir evidencias fotográficas.", nil)
			return
		}
		if len(stringValue(next["dataUrl"])) >= len(stringValue(old["dataUrl"])) {
			continue
		}
		for _, key := range []string{"dataUrl", "size", "width", "height"} {
			old[key] = next[key]
		}
		old["type"] = "image/jpeg"
		old["updatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
		files[i] = old
		count++
	}
	if count > 0 {
		if err = s.savePreRestoreSafety(st, time.Now()); err != nil {
			legacyError(w, 500, "No se pudo crear el respaldo de fotos originales.", nil)
			return
		}
		st.State.Files = files
		st.State.ServerVersion++
		st.State.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		if err = s.saveLegacy(st); err != nil {
			legacyError(w, 500, "No se pudo guardar la optimización.", nil)
			return
		}
	}
	jsonOut(w, 200, map[string]any{"ok": true, "optimized": count, "serverVersion": st.State.ServerVersion})
}

func extractMusicCLI(input, output string) error {
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	var root map[string]any
	if json.Unmarshal(data, &root) != nil {
		return errors.New("Selecciona un respaldo JSON o .cabackup, no el ZIP del programa.")
	}
	password := ""
	if root["ciphertext"] != nil {
		fmt.Print("Contraseña de la copia (se mostrará en esta consola): ")
		password, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		password = strings.TrimSpace(password)
	}
	db, files, err := decodeMusicBackup(data, password)
	if err != nil {
		return err
	}
	if missing := musicMissingFiles(db, files); len(missing) > 0 {
		return fmt.Errorf("faltan %d adjuntos locales o sus contenidos en el respaldo; no se creó una copia incompleta", len(missing))
	}
	out := map[string]any{"format": "MinisterioMusicaOnlineBackup", "version": musicOnlineVersion, "database": db, "files": files, "exportedAt": time.Now().UTC().Format(time.RFC3339Nano)}
	b, _ := json.MarshalIndent(out, "", "  ")
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewReader(b))
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	summary, _ := json.Marshal(musicSummary(db, files))
	fmt.Printf("Copia exclusiva de Música creada: %s\n%s\n", output, summary)
	return nil
}
