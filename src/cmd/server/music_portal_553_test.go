package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func musicFixture553(t *testing.T) (*Server, http.Handler) {
	t.Helper()
	s := &Server{cfg: Config{PublicWeb: true, AdminToken: "fixture-admin", ClientToken: "fixture-client", ServerName: "SGL de prueba", WebRoot: filepath.Clean("../../../app")}, path: filepath.Join(t.TempDir(), "sgl-server-data.json"), remoteSessions: map[string]remoteSession{}, loginAttempts: map[string]loginAttempt{}}
	user := func(id, role string) map[string]any {
		return map[string]any{"id": id, "username": id, "fullName": "Cuenta " + id, "role": role, "active": true, "accountState": "active", "password": testPasswordRecord("Prueba123!"), "region": "1"}
	}
	admin := user("admin", "admin")
	a := user("a", "aspirant")
	b := user("b", "aspirant")
	c := user("c", "aspirant")
	tm := user("tm", "tutor")
	tg := user("tg", "tutor")
	other := user("other-tm", "tutor")
	tm["tutorProgramIds"] = []any{musicProgram553}
	tg["tutorProgramIds"] = []any{"guide_major"}
	other["tutorProgramIds"] = []any{musicProgram553}
	a["tutorByProgram"] = map[string]any{musicProgram553: "tm", "guide_major": "tg"}
	b["tutorByProgram"] = map[string]any{musicProgram553: "other-tm"}
	db := map[string]any{"version": "5.53", "users": []any{admin, a, b, c, tm, tg, other, user("secretaria", "secretaria"), user("departamental", "departamental"), user("administrativo", "administrativo"), user("coordinador", "coordinador_musica")}, "settings": map[string]any{"autoSync": true, "appLogo": "SGL-logo-original"}, "modules": []any{map[string]any{"id": musicProgram553, "active": true, "assignedUserIds": []any{"a", "b"}}, map[string]any{"id": "guide_major", "active": true, "assignedUserIds": []any{"a", "c"}}}, "classes": []any{map[string]any{"id": "class-private"}}, "clubs": []any{map[string]any{"id": "club-private", "name": "Club de SGL"}}, "progress": map[string]any{}, "programProgress": map[string]any{"a": map[string]any{musicProgram553: map[string]any{"status": map[string]any{}, "submissions": map[string]any{"LM-PRE-1": map[string]any{"attachmentIds": []any{}}}}, "guide_major": map[string]any{"status": map[string]any{"GM-PRE-1": "approved"}}}, "b": map[string]any{musicProgram553: map[string]any{"status": map[string]any{}, "submissions": map[string]any{}}}}, "materials": []any{map[string]any{"id": "music-material", "programId": musicProgram553, "title": "Material de Música", "fileId": "music-a", "url": "https://example.test/music"}, map[string]any{"id": "other-material", "programId": "guide_major", "title": "Material privado de Guía", "fileId": "other-file"}}, "tasks": []any{map[string]any{"id": "other-task", "programId": "guide_major", "assigneeId": "c"}}, "configs": map[string]any{}, "indicationRules": []any{}, "audit": []any{}, "music550": map[string]any{}}
	st := legacyStore{State: legacyState{ServerVersion: 7, RestoreGeneration: 1, Database: db, Files: []any{map[string]any{"id": "music-a", "ownerId": "a", "name": "evidencia-music.txt", "dataUrl": "data:text/plain;base64,QQ=="}, map[string]any{"id": "other-file", "ownerId": "a", "name": "evidencia-guide.txt", "dataUrl": "data:text/plain;base64,Qg=="}}}}
	s.mu.Lock()
	err := s.saveLegacy(st)
	s.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	s.routes(mux)
	s.webRoutes(mux)
	return s, s.cors(s.securityHeaders(mux))
}

func musicCall553(t *testing.T, h http.Handler, method, path, token string, body any) (int, map[string]any) {
	t.Helper()
	var data []byte
	if body != nil {
		data, _ = json.Marshal(body)
	}
	r := httptest.NewRequest(method, "http://example.test"+path, bytes.NewReader(data))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var payload map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &payload)
	return w.Code, payload
}
func musicLogin553(t *testing.T, h http.Handler, user string, music bool) (int, map[string]any) {
	path := "/api/remote/login"
	if music {
		path = musicAPI553 + path
	}
	return musicCall553(t, h, "POST", path, "", map[string]any{"identity": user, "password": "Prueba123!", "authOnly": true})
}

func TestMusicPortalAccountAndSessionIsolation553(t *testing.T) {
	_, h := musicFixture553(t)
	for _, id := range []string{"admin", "a", "tm", "secretaria", "departamental", "administrativo", "coordinador"} {
		code, login := musicLogin553(t, h, id, true)
		if code != 200 {
			t.Fatalf("%s: %d %v", id, code, login)
		}
		token := stringValue(login["sessionToken"])
		code, p := musicCall553(t, h, "GET", musicAPI553+"/api/remote/session", token, nil)
		if code != 200 || p["product"] != "music" {
			t.Fatalf("session: %d %v", code, p)
		}
		code, _ = musicCall553(t, h, "GET", "/api/bootstrap", token, nil)
		if code != 403 {
			t.Fatal("Music token escaped into SGL", code)
		}
	}
	for _, id := range []string{"c", "tg"} {
		code, _ := musicLogin553(t, h, id, true)
		if code != 403 {
			t.Fatal("foreign account entered Music", id, code)
		}
		code, login := musicLogin553(t, h, id, false)
		if code != 200 {
			t.Fatal("existing SGL login changed", id, code)
		}
		code, _ = musicCall553(t, h, "GET", musicAPI553+"/api/bootstrap", stringValue(login["sessionToken"]), nil)
		if code != 401 {
			t.Fatal("SGL session reused by Music", code)
		}
	}
}

func TestMusicPortalSnapshotsAndFiles553(t *testing.T) {
	_, h := musicFixture553(t)
	for _, id := range []string{"admin", "a", "tm"} {
		_, login := musicLogin553(t, h, id, true)
		token := stringValue(login["sessionToken"])
		code, p := musicCall553(t, h, "GET", musicAPI553+"/api/bootstrap", token, nil)
		if code != 200 {
			t.Fatal(code, p)
		}
		db := legacyMap(p["database"])
		encoded, _ := json.Marshal(p)
		text := string(encoded)
		for _, forbidden := range []string{"guide_major", "class-private", "club-private", "other-material", "other-task", "other-file", "password", "recoveryKeys"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%s exposed %s: %s", id, forbidden, text)
			}
		}
		if legacyUser(db, "c") != nil || legacyUser(db, "tg") != nil {
			t.Fatal("foreign accounts exposed")
		}
		if id != "admin" && (legacyMap(db["programProgress"])["b"] != nil || legacyUser(db, "b") != nil) {
			t.Fatal("unassigned aspirant exposed", id)
		}
		expectedFiles := 1
		if id == "tm" {
			expectedFiles = 0
		}
		if len(anySlice548(p["files"])) != expectedFiles {
			t.Fatalf("files for %s: %v full=%v", id, p["files"], p)
		}
		code, files := musicCall553(t, h, "GET", musicAPI553+"/api/remote/files?fileId=other-file", token, nil)
		if code == 200 && len(anySlice548(files["files"])) != 0 {
			t.Fatal("foreign file downloaded")
		}
	}
}

func TestMusicPortalWritesPreserveSGL553(t *testing.T) {
	s, h := musicFixture553(t)
	_, login := musicLogin553(t, h, "admin", true)
	token := stringValue(login["sessionToken"])
	_, p := musicCall553(t, h, "GET", musicAPI553+"/api/bootstrap", token, nil)
	view := legacyMap(p["database"])
	legacyUser(view, "a")["fullName"] = "Nombre compartido actualizado"
	code, p := musicCall553(t, h, "POST", musicAPI553+"/api/sync", token, map[string]any{"actorId": "admin", "actorRole": "admin", "baseVersion": 7, "restoreGeneration": 1, "database": view, "files": p["files"]})
	if code != 200 {
		t.Fatal(code, p)
	}
	s.mu.Lock()
	state, _ := s.loadLegacy()
	s.mu.Unlock()
	db := legacyMap(state.State.Database)
	if legacyUser(db, "c") == nil || len(arrayField(db, "clubs")) != 1 || len(arrayField(db, "classes")) != 1 || len(arrayField(db, "modules")) != 2 || len(state.State.Files) != 2 || legacyMap(legacyMap(db["programProgress"])["a"])["guide_major"] == nil {
		t.Fatalf("Music replaced other SGL data: users=%d clubs=%d classes=%d modules=%d files=%d guide=%v", len(legacyUsers(db)), len(arrayField(db, "clubs")), len(arrayField(db, "classes")), len(arrayField(db, "modules")), len(state.State.Files), legacyMap(legacyMap(db["programProgress"])["a"])["guide_major"])
	}
	if legacyName(legacyUser(db, "a")) != "Nombre compartido actualizado" {
		t.Fatal("shared profile did not synchronize")
	}
	if legacyMap(legacyUser(db, "a")["tutorByProgram"])["guide_major"] != "tg" {
		t.Fatal("foreign tutor assignment lost")
	}
	for _, body := range []any{map[string]any{"programId": "guide_major"}, map[string]any{"ownerId": "c", "programId": musicProgram553}, map[string]any{"patch": map[string]any{"users": []any{map[string]any{"id": "c", "fullName": "Changed"}}}}} {
		code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/admin/changes", token, body)
		if code != 403 {
			t.Fatal("foreign change accepted", code, body)
		}
	}
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/admin/restore", token, map[string]any{})
	if code != 403 {
		t.Fatal("global restore exposed")
	}
}

func TestMusicPortalBrowser553(t *testing.T) {
	if os.Getenv("MUSICA_BROWSER_RUN") != "1" {
		t.Skip("browser integration is opt-in")
	}
	s, h := musicFixture553(t)
	st, _ := s.loadLegacy()
	p := reviewProgressContainer(legacyMap(st.State.Database), "a", "music_ministry")
	ensureMapField(p, "status")["LM-REL-5"] = "approved"
	ensureMapField(p, "status")["LM-NUT-4"] = "approved"
	s.saveLegacy(st)
	server := httptest.NewServer(h)
	defer server.Close()
	cmd := exec.Command("node", "../../../tests/browser-music-portal.cjs")
	cmd.Env = append(os.Environ(), "SGL_BROWSER_BASE="+server.URL)
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	if err != nil {
		t.Fatal(err)
	}
}

func TestMusicPortalScopedEditsAndRevocation553(t *testing.T) {
	s, h := musicFixture553(t)
	_, login := musicLogin553(t, h, "admin", true)
	token := stringValue(login["sessionToken"])
	code, p := musicCall553(t, h, "POST", musicAPI553+"/api/admin/changes", token, map[string]any{"actorId": "admin", "operationId": "music-material-delete", "ackOnly": true, "patch": map[string]any{"materials": map[string]any{"__sglReplace": []any{}}}})
	if code != 200 {
		t.Fatal("Music collection edit failed", code, p)
	}
	s.mu.Lock()
	state, _ := s.loadLegacy()
	s.mu.Unlock()
	materials := arrayField(legacyMap(state.State.Database), "materials")
	if len(materials) != 1 || stringValue(legacyMap(materials[0])["id"]) != "other-material" {
		t.Fatal("Music edit removed foreign material", materials)
	}
	// Assigning a foreign file to a Music material must not permit overwriting its bytes.
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/admin/changes", token, map[string]any{"actorId": "admin", "operationId": "foreign-file-overwrite", "ackOnly": true, "patch": map[string]any{"materials": []any{map[string]any{"id": "music-new", "programId": musicProgram553, "fileId": "other-file"}}}, "files": []any{map[string]any{"id": "other-file", "ownerId": "a", "dataUrl": "data:text/plain;base64,Qw=="}}})
	if code != 403 {
		t.Fatal("foreign file overwrite accepted", code)
	}
	_, aLogin := musicLogin553(t, h, "a", true)
	s.mu.Lock()
	state, _ = s.loadLegacy()
	for _, raw := range arrayField(legacyMap(state.State.Database), "modules") {
		m := legacyMap(raw)
		if stringValue(m["id"]) == musicProgram553 {
			m["assignedUserIds"] = []any{"b"}
		}
	}
	err := s.saveLegacy(state)
	s.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	code, _ = musicCall553(t, h, "GET", musicAPI553+"/api/remote/session", stringValue(aLogin["sessionToken"]), nil)
	if code != 403 {
		t.Fatal("revoked Music membership kept access", code)
	}
	code, _ = musicLogin553(t, h, "a", false)
	if code != 200 {
		t.Fatal("removing Music access disabled shared SGL account", code)
	}
}

func TestMusicPortalPresenceUsesMusicSessions553(t *testing.T) {
	_, h := musicFixture553(t)
	for _, tc := range []struct {
		id    string
		music bool
	}{{"a", true}, {"c", false}} {
		_, login := musicLogin553(t, h, tc.id, tc.music)
		prefix := ""
		if tc.music {
			prefix = musicAPI553
		}
		code, p := musicCall553(t, h, "POST", prefix+"/api/remote/heartbeat", stringValue(login["sessionToken"]), map[string]any{"visible": true})
		if code != 200 {
			t.Fatal("heartbeat", code, p)
		}
	}
	_, login := musicLogin553(t, h, "admin", true)
	code, p := musicCall553(t, h, "GET", musicAPI553+"/api/remote/presence", stringValue(login["sessionToken"]), nil)
	users := anySlice548(p["users"])
	if code != 200 || len(users) != 1 || stringValue(legacyMap(users[0])["userId"]) != "a" || p["count"] != float64(1) {
		t.Fatal("Music presence counted a general SGL session", code, p)
	}
}
