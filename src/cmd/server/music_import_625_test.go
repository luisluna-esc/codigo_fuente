package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func upload625(t *testing.T, payload []byte, preview, confirmation string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	m := multipart.NewWriter(&body)
	for _, field := range [][2]string{{"key", ""}, {"password", "Contraseña con ñ y espacios"}, {"preview", preview}, {"confirmation", confirmation}} {
		if err := m.WriteField(field[0], field[1]); err != nil {
			t.Fatal(err)
		}
	}
	f, err := m.CreateFormFile("backup", "copia anterior.json")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.Write(payload)
	_ = m.Close()
	r := httptest.NewRequest(http.MethodPost, "/api/music-data/import", &body)
	r.Header.Set("Content-Type", m.FormDataContentType())
	return r
}

func Test625NativeUploadPreviewImportAuthorizationAndSafety(t *testing.T) {
	s, h := musicFixture553(t)
	h = s.onlineProduct(h)
	s.configPath = filepath.Join(filepath.Dir(s.path), "server-config.json")
	st, _ := s.loadLegacy()
	before, _ := json.Marshal(st.State)
	data, _ := json.Marshal(map[string]any{"database": st.State.Database, "files": st.State.Files})
	call := func(auth, preview, confirmation string) (int, map[string]any) {
		r := upload625(t, data, preview, confirmation)
		if auth != "" {
			r.Header.Set("Authorization", "Bearer "+auth)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		var out map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}
	if code, _ := call("", "true", ""); code != 403 {
		t.Fatal("anonymous upload accepted", code)
	}
	if code, out := call("fixture-admin", "true", ""); code != 200 || out["summary"] == nil {
		t.Fatal("preview", code, out)
	}
	st, _ = s.loadLegacy()
	after, _ := json.Marshal(st.State)
	if !bytes.Equal(before, after) {
		t.Fatal("preview changed stored data")
	}
	if code, _ := call("fixture-admin", "false", ""); code != 400 {
		t.Fatal("import without confirmation", code)
	}
	if code, out := call("fixture-admin", "false", "IMPORTAR MUSICA"); code != 200 {
		t.Fatal("restore", code, out)
	}
	backups, _ := filepath.Glob(filepath.Join(filepath.Dir(s.legacyPath()), "sgl-server-compat.pre-restore-*.json"))
	if len(backups) != 1 {
		t.Fatal("missing safety backup")
	}
}

func Test625NativeUploadBoundariesAndUnicodeMetadata(t *testing.T) {
	data := []byte(`{"database":{}}`)
	req, err := readMusicImport625(httptest.NewRecorder(), upload625(t, data, "true", "IMPORTAR MUSICA"), int64(len(data)))
	if err != nil || !bytes.Equal(req.Payload, data) || req.Password != "Contraseña con ñ y espacios" || !req.Preview || req.Confirmation != "IMPORTAR MUSICA" {
		t.Fatal("metadata or file changed", err)
	}
	if _, err = readMusicImport625(httptest.NewRecorder(), upload625(t, data, "true", ""), int64(len(data)-1)); err == nil {
		t.Fatal("oversized upload accepted")
	}
	if _, err = readMusicImport625(httptest.NewRecorder(), upload625(t, nil, "true", ""), 100); err == nil {
		t.Fatal("empty upload accepted")
	}
	if _, err = readMusicImport625(httptest.NewRecorder(), upload625(t, data, "unexpected", ""), 100); err == nil {
		t.Fatal("invalid preview accepted")
	}
	var body bytes.Buffer
	m := multipart.NewWriter(&body)
	_ = m.WriteField("preview", "true")
	_ = m.WriteField("preview", "false")
	_ = m.Close()
	r := httptest.NewRequest("POST", "/api/music-data/import", &body)
	r.Header.Set("Content-Type", m.FormDataContentType())
	if _, err = readMusicImport625(httptest.NewRecorder(), r, 100); err == nil {
		t.Fatal("duplicate fields accepted")
	}
}

func Test625NativeUploadRejectsTruncatedBody(t *testing.T) {
	r := upload625(t, []byte(`{"database":{}}`), "true", "")
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b[:len(b)-10]))
	if _, err := readMusicImport625(httptest.NewRecorder(), r, 1000); err == nil {
		t.Fatal("truncated upload accepted")
	}
}

// Optional private fixture: the source release never contains the user's backup.
func Test625ActualBackupUploadPreview(t *testing.T) {
	path := os.Getenv("MUSIC_IMPORT_BACKUP")
	if path == "" {
		t.Skip("set MUSIC_IMPORT_BACKUP for private large-backup validation")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s, h := musicFixture553(t)
	h = s.onlineProduct(h)
	r := upload625(t, data, "true", "")
	r.Header.Set("Authorization", "Bearer fixture-admin")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("large-backup preview returned %d: %s", w.Code, strings.TrimSpace(w.Body.String()))
	}
}
