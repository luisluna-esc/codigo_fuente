package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMusicOnlineExtractOnlyMusicAndInactiveHistory(t *testing.T) {
	s, _ := musicFixture553(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	b := legacyUser(db, "b")
	b["active"] = false
	b["accountState"] = "inactive"
	legacyMap(legacyMap(db["programProgress"])["a"])[musicProgram553] = map[string]any{"status": map[string]any{"LM-PRE-1": "approved"}, "submissions": map[string]any{"LM-PRE-1": map[string]any{"text": "Evidencia completa", "attachmentIds": []any{"music-a"}}}}
	settings := ensureMapField(db, "settings")
	settings["googleClientId"] = "secret"
	settings["driveToken"] = "secret"
	settings["serverToken"] = "secret"
	db["syncQueue"] = []any{map[string]any{"programId": "guide_major", "patch": map[string]any{"private": "other module"}}}
	data, _ := json.Marshal(map[string]any{"database": db, "files": st.State.Files})
	out, files, err := decodeMusicBackup(data, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(arrayField(out, "modules")) != 1 || stringValue(legacyMap(arrayField(out, "modules")[0])["id"]) != musicProgram553 {
		t.Fatal("other module retained")
	}
	for _, id := range []string{"c", "tg"} {
		if legacyUser(out, id) != nil {
			t.Fatalf("unrelated account %s retained", id)
		}
	}
	if legacyUser(out, "b") == nil {
		t.Fatal("inactive member lost")
	}
	if legacyUser(out, "tm") == nil {
		t.Fatal("tutor lost")
	}
	p := legacyMap(legacyMap(legacyMap(out["programProgress"])["a"])[musicProgram553])
	if stringValue(legacyMap(p["status"])["LM-PRE-1"]) != "approved" {
		t.Fatal("approval lost")
	}
	if len(files) != 1 || stringValue(legacyMap(files[0])["id"]) != "music-a" {
		t.Fatalf("incorrect files %#v", files)
	}
	bdata, _ := json.Marshal(out["settings"])
	if strings.Contains(string(bdata), "secret") || len(arrayField(out, "syncQueue")) != 0 {
		t.Fatal("integration secrets or offline queue retained")
	}
}

func TestMusicOnlineEncryptedBackup(t *testing.T) {
	s, _ := musicFixture553(t)
	st, _ := s.loadLegacy()
	plain, _ := json.Marshal(map[string]any{"format": "SistemaGestionLiderazgoBackup", "database": st.State.Database, "files": st.State.Files})
	salt := []byte("0123456789abcdef")
	iv := []byte("123456789012")
	block, _ := aes.NewCipher(pbkdf2SHA256([]byte("password-test"), salt, 1000, 32))
	gcm, _ := cipher.NewGCM(block)
	env := map[string]any{"salt": base64.StdEncoding.EncodeToString(salt), "iv": base64.StdEncoding.EncodeToString(iv), "ciphertext": base64.StdEncoding.EncodeToString(gcm.Seal(nil, iv, plain, nil)), "iterations": 1000}
	data, _ := json.Marshal(env)
	if _, _, err := decodeMusicBackup(data, "wrong"); err == nil {
		t.Fatal("wrong password accepted")
	}
	if db, _, err := decodeMusicBackup(data, "password-test"); err != nil || legacyUser(db, "admin") == nil {
		t.Fatalf("encrypted import: %v", err)
	}
}

func TestMusicOnlineRemovedEndpointsAndProgramBoundary(t *testing.T) {
	s, h := musicFixture553(t)
	h = s.onlineProduct(h)
	for _, path := range []string{"/api/google/config", "/api/musica/api/google/start", "/api/admin/recovery", "/api/recovery/status", "/api/accounts/reconcile598", "/api/sync/diagnostics", "/sw.js", "/offline-assets.json", "/manifest.webmanifest"} {
		r := httptest.NewRequest("GET", path, nil)
		r.Header.Set("Authorization", "Bearer fixture-admin")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 404 {
			t.Errorf("%s returned %d", path, w.Code)
		}
	}
	r := httptest.NewRequest("GET", "/api/music/module?programId=guide_major", nil)
	r.Header.Set("Authorization", "Bearer fixture-admin")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatalf("other program accepted: %d", w.Code)
	}
}

func TestMusicOnlineImportPreviewConfirmationAndSafety(t *testing.T) {
	s, h := musicFixture553(t)
	s.configPath = filepath.Join(filepath.Dir(s.path), "server-config.json")
	h = s.onlineProduct(h)
	st, _ := s.loadLegacy()
	before, _ := json.Marshal(st.State.Database)
	payload := map[string]any{"database": st.State.Database, "files": st.State.Files}
	call := func(token string, body any) (int, map[string]any) {
		data, _ := json.Marshal(body)
		r := httptest.NewRequest(http.MethodPost, "/api/music-data/import", bytes.NewReader(data))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	if code, _ := call("", map[string]any{"payload": payload, "preview": true}); code != 403 {
		t.Fatal("anonymous import allowed")
	}
	if code, out := call("fixture-admin", map[string]any{"payload": payload, "preview": true}); code != 200 || out["summary"] == nil {
		t.Fatalf("preview=%d %#v", code, out)
	}
	current, _ := s.loadLegacy()
	after, _ := json.Marshal(current.State.Database)
	if !bytes.Equal(before, after) {
		t.Fatal("preview mutated data")
	}
	if code, _ := call("fixture-admin", map[string]any{"payload": payload}); code != 400 {
		t.Fatal("confirmation not required")
	}
	if code, out := call("fixture-admin", map[string]any{"payload": payload, "confirmation": "IMPORTAR MUSICA"}); code != 200 {
		t.Fatalf("import %d %#v", code, out)
	}
	current, _ = s.loadLegacy()
	if len(arrayField(legacyMap(current.State.Database), "modules")) != 1 {
		t.Fatal("import retained other programs")
	}
	files, _ := filepath.Glob(filepath.Join(filepath.Dir(s.legacyPath()), "sgl-server-compat.pre-restore-*.json"))
	if len(files) != 1 {
		t.Fatal("missing pre-import safety backup")
	}
}

func TestMusicOnlineSetupIsOneTimeAndRequiresServerKey(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{PublicWeb: true, AdminToken: "admin", ClientToken: "client"}, path: filepath.Join(dir, "data", "musica-server-data.json"), configPath: filepath.Join(dir, "server-config.json")}
	if err := s.prepareInstallation(); err != nil {
		t.Fatal(err)
	}
	key, _ := os.ReadFile(s.installationKeyPath())
	call := func(key string) int {
		data, _ := json.Marshal(map[string]any{"key": key, "fullName": "Admin Música", "username": "admin", "password": "ClaveSegura123!"})
		w := httptest.NewRecorder()
		s.installationEndpoint(w, httptest.NewRequest("POST", "/api/installation/create", bytes.NewReader(data)), "/api/installation/create")
		return w.Code
	}
	if call("wrong") != 403 {
		t.Fatal("setup without key")
	}
	if code := call(strings.TrimSpace(string(key))); code != 200 {
		t.Fatalf("setup=%d", code)
	}
	if call(strings.TrimSpace(string(key))) != 409 {
		t.Fatal("setup allowed twice")
	}
	st, _ := s.loadLegacy()
	admin := legacyMap(legacyUsers(st.State.Database)[0])
	if !verifyRemotePassword("ClaveSegura123!", admin) {
		t.Fatal("admin password not valid")
	}
}

func TestMusicOnlineNoRemoteReferencesDeletedUsersOrOldVersions(t *testing.T) {
	s, _ := musicFixture553(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	p := legacyMap(legacyMap(db["programProgress"])["a"])[musicProgram553].(map[string]any)
	ensureMapField(p, "submissions")["LM-PRE-1"] = map[string]any{"text": "Informe vigente", "attachmentIds": []any{"music-a"}, "drive": map[string]any{"fileId": "remote-provider-id", "webViewLink": "https://drive.google.com/old"}}
	legacyUser(db, "b")["musicAccessRemoved"] = true
	legacyMap(db["programProgress"])["b"] = map[string]any{musicProgram553: map[string]any{"status": map[string]any{"LM-PRE-1": "approved"}, "submissions": map[string]any{"LM-PRE-1": map[string]any{"attachmentIds": []any{"deleted-evidence"}}}}}
	music := ensureMapField(db, "music550")
	music["certificates"] = map[string]any{"b:final:": map[string]any{"ownerId": "b", "fullName": "Eliminado"}}
	music["workflow"] = map[string]any{"b": map[string]any{"state": "approved"}, "a": map[string]any{"state": "correction"}}
	ensureMapField(db, "settings")["nested"] = map[string]any{"googleOAuth": map[string]any{"fileId": "another-remote-id"}, "watermarkFileId": "music-a"}
	st.State.Files = append(st.State.Files, map[string]any{"id": "deleted-evidence", "ownerId": "b", "programId": musicProgram553}, map[string]any{"id": "old-version", "ownerId": "a", "programId": musicProgram553})
	input, _ := json.Marshal(map[string]any{"database": db, "files": st.State.Files})
	out, files, err := decodeMusicBackup(input, "")
	if err != nil {
		t.Fatal(err)
	}
	if legacyUser(out, "b") != nil || legacyMap(out["programProgress"])["b"] != nil {
		t.Fatal("eliminated account resurrected")
	}
	if len(files) != 1 || stringValue(legacyMap(files[0])["id"]) != "music-a" {
		t.Fatal("deleted or obsolete files migrated")
	}
	if len(mapField(legacyMap(out["music550"]), "certificates")) != 0 || mapField(legacyMap(out["music550"]), "workflow")["b"] != nil {
		t.Fatal("eliminated user's records retained")
	}
	ids := map[string]bool{}
	musicFileIDs553(out, ids)
	if len(ids) != 1 || !ids["music-a"] {
		t.Fatalf("remote identifiers treated as attachments: %#v", ids)
	}
	encoded, _ := json.Marshal(out)
	if strings.Contains(strings.ToLower(string(encoded)), "drive") || strings.Contains(strings.ToLower(string(encoded)), "oauth") {
		t.Fatal("removed integration retained")
	}
	if legacyUser(db, "b") == nil || mapField(mapField(p, "submissions"), "LM-PRE-1")["drive"] == nil {
		t.Fatal("source backup modified")
	}
}

func TestMusicOnlineRealMissingAttachmentStillBlocksImport(t *testing.T) {
	s, h := musicFixture553(t)
	h = s.onlineProduct(h)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	p := legacyMap(legacyMap(db["programProgress"])["a"])[musicProgram553].(map[string]any)
	ensureMapField(p, "submissions")["LM-PRE-1"] = map[string]any{"attachmentIds": []any{"actual-missing-file"}, "drive": map[string]any{"fileId": "remote-id"}}
	payload := map[string]any{"database": db, "files": st.State.Files}
	call := func(preview bool) (int, map[string]any) {
		body, _ := json.Marshal(map[string]any{"payload": payload, "preview": preview, "confirmation": "IMPORTAR MUSICA"})
		r := httptest.NewRequest("POST", "/api/music-data/import", bytes.NewReader(body))
		r.Header.Set("Authorization", "Bearer fixture-admin")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		out := map[string]any{}
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	code, out := call(true)
	if code != 200 || len(arrayField(out, "missingFiles")) != 1 || arrayField(out, "missingFiles")[0] != "actual-missing-file" {
		t.Fatalf("preview: %d %#v", code, out)
	}
	if code, _ = call(false); code != 400 {
		t.Fatal("import accepted missing local evidence")
	}
	payload["files"] = append(st.State.Files, map[string]any{"id": "actual-missing-file", "ownerId": "a", "programId": musicProgram553})
	if code, out = call(true); code != 200 || len(arrayField(out, "missingFiles")) != 1 {
		t.Fatal("empty file metadata accepted as evidence", code, out)
	}
}

func TestMusicOnlinePendingUploadSurvivesUntilReportSave(t *testing.T) {
	s, _ := musicFixture553(t)
	st, _ := s.loadLegacy()
	db, _ := extractMusicDatabase(st.State.Database)
	st.State.Database = db
	st.State.Files = append(st.State.Files, map[string]any{"id": "new-upload", "ownerId": "a", "programId": musicProgram553})
	s.cleanStandaloneMusicStore(&st)
	found := false
	for _, raw := range st.State.Files {
		if stringValue(legacyMap(raw)["id"]) == "new-upload" {
			found = true
		}
	}
	if !found {
		t.Fatal("new upload removed before report could link it")
	}
	legacyUser(legacyMap(st.State.Database), "a")["musicAccessRemoved"] = true
	s.cleanStandaloneMusicStore(&st)
	for _, raw := range st.State.Files {
		if stringValue(legacyMap(raw)["id"]) == "new-upload" {
			t.Fatal("eliminated owner's exclusive files retained")
		}
	}
}

func TestMusicOnlineStopRequiresLocalConfigurationSecret(t *testing.T) {
	s, h := musicFixture553(t)
	s.stop = make(chan struct{}, 1)
	h = s.onlineProduct(h)
	cases := []struct {
		addr, token, origin string
		want                int
	}{{"127.0.0.1:12000", "", "", 404}, {"198.51.100.2:12000", "fixture-admin", "", 404}, {"127.0.0.1:12000", "fixture-admin", "https://example.test", 404}, {"127.0.0.1:12000", "fixture-admin", "", 200}}
	for _, c := range cases {
		r := httptest.NewRequest("POST", "/api/server/stop", nil)
		r.RemoteAddr = c.addr
		r.Header.Set("Authorization", "Bearer "+c.token)
		if c.origin != "" {
			r.Header.Set("Origin", c.origin)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != c.want {
			t.Fatalf("stop returned %d, want %d", w.Code, c.want)
		}
	}
	if len(s.stop) != 1 {
		t.Fatal("authorized stop did not reach shutdown channel")
	}
}
