package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sgl.local/sistema-gestion-liderazgo/internal/release"
	"strings"
	"testing"
	"time"
)

func testPasswordRecord(password string) map[string]any {
	salt := []byte("0123456789abcdef")
	hash := pbkdf2SHA256([]byte(password), salt, 1000, 32)
	return map[string]any{"algorithm": "PBKDF2-SHA256", "iterations": 1000, "salt": base64.StdEncoding.EncodeToString(salt), "hash": base64.StdEncoding.EncodeToString(hash)}
}
func TestRemoteLoginFiltersDatabase(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{PublicWeb: true, ServerName: "Test"}, path: filepath.Join(dir, "data", "sgl-server-data.json"), remoteSessions: map[string]remoteSession{}, loginAttempts: map[string]loginAttempt{}}
	db := map[string]any{
		"settings": map[string]any{"serverToken": "ADMIN-SECRET", "serverUrl": "http://private"},
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "username": "admin", "role": "admin", "active": true, "accountState": "active", "password": testPasswordRecord("Admin123!")},
			map[string]any{"id": "tutor1", "fullName": "Tutor Uno", "username": "tutor", "email": "t@example.test", "role": "tutor", "active": true, "accountState": "active", "password": testPasswordRecord("Tutor123!")},
			map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "username": "asp1", "role": "aspirant", "tutorId": "tutor1", "active": true, "accountState": "active", "password": testPasswordRecord("Asp12345!")},
			map[string]any{"id": "asp2", "fullName": "Aspirante Otro", "username": "asp2", "role": "aspirant", "tutorId": "otro", "active": true, "accountState": "active", "password": testPasswordRecord("Asp22345!")},
		},
		"progress": map[string]any{"asp1": map[string]any{"x": 1}, "asp2": map[string]any{"x": 2}},
		"audit":    []any{}, "classes": []any{}, "tasks": []any{}, "materials": []any{}, "modules": []any{map[string]any{"id": "class_grouped", "active": true, "assignedUserIds": []any{"asp1", "asp2"}}},
	}
	st := legacyStore{State: legacyState{ServerVersion: 2, Database: db, Files: []any{
		map[string]any{"id": "file-asp1", "ownerId": "asp1", "name": "permitido.jpg", "type": "image/jpeg", "dataUrl": "data:image/jpeg;base64,QQ=="},
		map[string]any{"id": "file-asp2", "ownerId": "asp2", "name": "privado.jpg", "type": "image/jpeg", "dataUrl": "data:image/jpeg;base64,Qg=="},
	}}}
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"identity": "tutor", "password": "Tutor123!"})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/remote/login", bytes.NewReader(body))
	req.RemoteAddr = "203.0.113.5:1234"
	rr := httptest.NewRecorder()
	s.remoteLogin(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["sessionToken"] == "" {
		t.Fatal("missing session token")
	}
	if _, ok := payload["files"]; ok {
		t.Fatal("login should not block on evidence files")
	}
	if payload["filesDeferred"] != true {
		t.Fatalf("expected filesDeferred=true, got %#v", payload["filesDeferred"])
	}
	gotDB := legacyMap(payload["database"])
	settings := legacyMap(gotDB["settings"])
	if _, ok := settings["serverToken"]; ok {
		t.Fatal("serverToken leaked")
	}
	ids := map[string]bool{}
	for _, raw := range legacyUsers(gotDB) {
		u := legacyMap(raw)
		ids[u["id"].(string)] = true
		if u["id"] == "asp1" {
			if _, ok := u["password"]; ok {
				t.Fatal("aspirant password leaked to tutor")
			}
		}
	}
	if !ids["tutor1"] || !ids["asp1"] || !ids["admin1"] {
		t.Fatalf("expected authorized users, got %#v", ids)
	}
	if ids["asp2"] {
		t.Fatal("unassigned aspirant leaked")
	}

	filesReq := httptest.NewRequest(http.MethodGet, "http://example/api/remote/files", nil)
	filesReq.Header.Set("Authorization", "Bearer "+payload["sessionToken"].(string))
	filesReq.RemoteAddr = "203.0.113.5:1234"
	filesRR := httptest.NewRecorder()
	s.remoteFiles(filesRR, filesReq)
	if filesRR.Code != 200 {
		t.Fatalf("files status=%d body=%s", filesRR.Code, filesRR.Body.String())
	}
	var filesPayload map[string]any
	if err := json.Unmarshal(filesRR.Body.Bytes(), &filesPayload); err != nil {
		t.Fatal(err)
	}
	files, _ := filesPayload["files"].([]any)
	if len(files) != 1 || legacyMap(files[0])["id"] != "file-asp1" {
		t.Fatalf("unexpected filtered files: %#v", files)
	}
}

func TestLegacySyncAcceptsLargeAuthenticatedPayload(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "admintest", ClientToken: "clienttest"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	large := strings.Repeat("x", 12<<20)
	db := map[string]any{
		"users":         []any{map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"}},
		"largeEvidence": large,
	}
	body, err := json.Marshal(map[string]any{
		"installationId": "pc-admin", "actorId": "admin1", "actorRole": "admin", "actorName": "Admin", "baseVersion": 0,
		"database": db, "files": []any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(body) <= 8<<20 {
		t.Fatalf("test payload unexpectedly small: %d", len(body))
	}
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer admintest")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRemotePresenceRequiresHeartbeat(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{PublicWeb: true, AdminToken: "admintest", ServerName: "Test"}, path: filepath.Join(dir, "data", "sgl-server-data.json"), remoteSessions: map[string]remoteSession{}, loginAttempts: map[string]loginAttempt{}}
	db := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "username": "asp1", "role": "aspirant", "active": true, "accountState": "active", "password": testPasswordRecord("Asp12345!")},
		},
	}
	testEnrollConquistadores593(db)
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 1, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"identity": "asp1", "password": "Asp12345!", "deviceId": "phone-1", "authOnly": true})
	loginReq := httptest.NewRequest(http.MethodPost, "http://example/api/remote/login", bytes.NewReader(body))
	loginReq.RemoteAddr = "203.0.113.7:1234"
	loginRR := httptest.NewRecorder()
	s.remoteLogin(loginRR, loginReq)
	if loginRR.Code != 200 {
		t.Fatalf("login status=%d body=%s", loginRR.Code, loginRR.Body.String())
	}
	var loginPayload map[string]any
	if err := json.Unmarshal(loginRR.Body.Bytes(), &loginPayload); err != nil {
		t.Fatal(err)
	}
	token := fmt.Sprint(loginPayload["sessionToken"])
	if token == "" {
		t.Fatal("login did not return session token")
	}

	// Session validation alone must not create a connected-presence false positive.
	sessionReq := httptest.NewRequest(http.MethodGet, "http://example/api/remote/session", nil)
	sessionReq.Header.Set("Authorization", "Bearer "+token)
	sessionRR := httptest.NewRecorder()
	s.remoteSessionInfo(sessionRR, sessionReq)
	if sessionRR.Code != http.StatusOK {
		t.Fatalf("session validation status=%d body=%s", sessionRR.Code, sessionRR.Body.String())
	}

	presenceCount := func() (int, []struct {
		UserID   string `json:"userId"`
		Name     string `json:"fullName"`
		DeviceID string `json:"deviceId"`
	}) {
		presenceReq := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/api/remote/presence", nil)
		presenceReq.Header.Set("Authorization", "Bearer admintest")
		presenceRR := httptest.NewRecorder()
		s.remotePresence(presenceRR, presenceReq)
		if presenceRR.Code != 200 {
			t.Fatalf("presence status=%d body=%s", presenceRR.Code, presenceRR.Body.String())
		}
		var payload struct {
			Count int `json:"count"`
			Users []struct {
				UserID   string `json:"userId"`
				Name     string `json:"fullName"`
				DeviceID string `json:"deviceId"`
			} `json:"users"`
		}
		if err := json.Unmarshal(presenceRR.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		return payload.Count, payload.Users
	}

	count, users := presenceCount()
	if count != 0 || len(users) != 0 {
		t.Fatalf("session validation must not mark user connected: count=%d users=%#v", count, users)
	}

	heartbeatReq := httptest.NewRequest(http.MethodPost, "http://example/api/remote/heartbeat", nil)
	heartbeatReq.Header.Set("Authorization", "Bearer "+token)
	heartbeatRR := httptest.NewRecorder()
	s.remoteHeartbeat(heartbeatRR, heartbeatReq)
	if heartbeatRR.Code != http.StatusOK {
		t.Fatalf("heartbeat status=%d body=%s", heartbeatRR.Code, heartbeatRR.Body.String())
	}

	count, users = presenceCount()
	if count != 1 || len(users) != 1 || users[0].UserID != "asp1" || users[0].Name != "Aspirante Uno" || users[0].DeviceID != "phone-1" {
		t.Fatalf("unexpected presence after heartbeat: count=%d users=%#v", count, users)
	}
}

func TestLoadLegacyLoginStateReturnsAuthDataWithoutEvidence(t *testing.T) {
	dir := t.TempDir()
	s := &Server{path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{
		"users":    []any{map[string]any{"id": "u1", "username": "user", "role": "aspirant", "active": true}},
		"settings": map[string]any{"dailyRecognitionHour": 18.0},
	}
	files := []any{map[string]any{"id": "f1", "ownerId": "u1", "dataUrl": "data:image/jpeg;base64," + strings.Repeat("A", 1<<20)}}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 17, Database: db, Files: files}}); err != nil {
		t.Fatal(err)
	}
	state, err := s.loadLegacyLoginState()
	if err != nil {
		t.Fatal(err)
	}
	if state.ServerVersion != 17 || state.Database == nil {
		t.Fatalf("unexpected login state: %#v", state)
	}
	if len(state.Files) != 0 {
		t.Fatalf("login reader must not materialize evidence files, got %d", len(state.Files))
	}
	users := legacyUsers(legacyMap(state.Database))
	if len(users) != 1 || legacyMap(users[0])["id"] != "u1" {
		t.Fatalf("unexpected login users: %#v", users)
	}
}

func TestRemoteLoginSupportsGzipAndServerTiming(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{PublicWeb: true, ServerName: "Test"}, path: filepath.Join(dir, "data", "sgl-server-data.json"), remoteSessions: map[string]remoteSession{}, loginAttempts: map[string]loginAttempt{}}
	db := map[string]any{
		"settings": map[string]any{},
		"users":    []any{map[string]any{"id": "asp1", "fullName": "Aspirante", "username": "asp", "role": "aspirant", "active": true, "accountState": "active", "password": testPasswordRecord("Asp12345!")}},
		"audit":    []any{}, "classes": []any{}, "tasks": []any{}, "materials": []any{}, "modules": []any{map[string]any{"id": "class_grouped", "active": true, "assignedUserIds": []any{"asp1", "asp2"}}},
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 3, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"identity": "asp", "password": "Asp12345!"})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/remote/login", bytes.NewReader(body))
	req.Header.Set("Accept-Encoding", "gzip")
	req.RemoteAddr = "203.0.113.20:1234"
	rr := httptest.NewRecorder()
	s.remoteLogin(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip response, headers=%v", rr.Header())
	}
	if !strings.Contains(rr.Header().Get("Server-Timing"), "sgl_load") || !strings.Contains(rr.Header().Get("Server-Timing"), "sgl_auth") {
		t.Fatalf("missing login timing header: %q", rr.Header().Get("Server-Timing"))
	}
	gz, err := gzip.NewReader(bytes.NewReader(rr.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	var payload map[string]any
	if err := json.NewDecoder(gz).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload["sessionToken"] == "" || payload["filesDeferred"] != true {
		t.Fatalf("unexpected compressed login payload: %#v", payload)
	}
}

func TestPBKDF2SHA256KnownVector(t *testing.T) {
	got := hex.EncodeToString(pbkdf2SHA256([]byte("password"), []byte("salt"), 4096, 32))
	const want = "c5e478d59288c841aa530db6845c4c8d962893a001ce4e11a4963873aa98134a"
	if got != want {
		t.Fatalf("PBKDF2-SHA256 changed: got %s want %s", got, want)
	}
}

func TestRemoteAuthOnlySessionAndBootstrapFlow508(t *testing.T) {
	dir := t.TempDir()
	s := &Server{
		cfg:            Config{PublicWeb: true, ServerName: "Club Central"},
		path:           filepath.Join(dir, "data", "sgl-server-data.json"),
		store:          Store{ServerID: "srv-test-508"},
		remoteSessions: map[string]remoteSession{},
		loginAttempts:  map[string]loginAttempt{},
	}
	db := map[string]any{
		"settings": map[string]any{},
		"users": []any{
			map[string]any{"id": "leader1", "fullName": "Lider Uno", "username": "lider", "role": "aspirant", "active": true, "accountState": "active", "password": testPasswordRecord("Clave123!")},
		},
		"progress": map[string]any{"leader1": map[string]any{"req": true}},
		"audit":    []any{}, "classes": []any{}, "tasks": []any{}, "materials": []any{}, "modules": []any{map[string]any{"id": "class_grouped", "active": true, "assignedUserIds": []any{"asp1", "asp2"}}},
	}
	testEnrollConquistadores593(db)
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 8, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}

	remember := false
	body, _ := json.Marshal(map[string]any{
		"identity": "lider", "password": "Clave123!", "deviceId": "phone-508",
		"authOnly": true, "rememberDevice": remember,
	})
	loginReq := httptest.NewRequest(http.MethodPost, "http://example/api/remote/login", bytes.NewReader(body))
	loginReq.RemoteAddr = "203.0.113.80:4567"
	loginRR := httptest.NewRecorder()
	s.remoteLogin(loginRR, loginReq)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", loginRR.Code, loginRR.Body.String())
	}
	var loginPayload map[string]any
	if err := json.Unmarshal(loginRR.Body.Bytes(), &loginPayload); err != nil {
		t.Fatal(err)
	}
	if _, exists := loginPayload["database"]; exists {
		t.Fatal("5.0.9 auth-only login must not include the database")
	}
	token, _ := loginPayload["sessionToken"].(string)
	if token == "" || loginPayload["serverId"] != "srv-test-508" || loginPayload["serverName"] != "Club Central" {
		t.Fatalf("unexpected auth-only payload: %#v", loginPayload)
	}
	if loginPayload["rememberDevice"] != false {
		t.Fatalf("rememberDevice not preserved: %#v", loginPayload["rememberDevice"])
	}
	expRaw, _ := loginPayload["expiresAt"].(string)
	exp, err := time.Parse(time.RFC3339, expRaw)
	if err != nil {
		t.Fatalf("invalid expiresAt %q: %v", expRaw, err)
	}
	remaining := time.Until(exp)
	if remaining < 11*time.Hour || remaining > 13*time.Hour {
		t.Fatalf("non-remembered session should be about 12h, got %v", remaining)
	}

	sessionReq := httptest.NewRequest(http.MethodGet, "http://example/api/remote/session", nil)
	sessionReq.Header.Set("Authorization", "Bearer "+token)
	sessionRR := httptest.NewRecorder()
	s.remoteSessionInfo(sessionRR, sessionReq)
	if sessionRR.Code != http.StatusOK {
		t.Fatalf("session status=%d body=%s", sessionRR.Code, sessionRR.Body.String())
	}
	var sessionPayload map[string]any
	if err := json.Unmarshal(sessionRR.Body.Bytes(), &sessionPayload); err != nil {
		t.Fatal(err)
	}
	if sessionPayload["userId"] != "leader1" || sessionPayload["serverId"] != "srv-test-508" {
		t.Fatalf("unexpected session payload: %#v", sessionPayload)
	}

	bootstrapReq := httptest.NewRequest(http.MethodGet, "http://example/api/remote/bootstrap", nil)
	bootstrapReq.Header.Set("Authorization", "Bearer "+token)
	bootstrapRR := httptest.NewRecorder()
	s.remoteBootstrap(bootstrapRR, bootstrapReq)
	if bootstrapRR.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrapRR.Code, bootstrapRR.Body.String())
	}
	var bootstrap map[string]any
	if err := json.Unmarshal(bootstrapRR.Body.Bytes(), &bootstrap); err != nil {
		t.Fatal(err)
	}
	if bootstrap["userId"] != "leader1" || bootstrap["serverId"] != "srv-test-508" {
		t.Fatalf("unexpected bootstrap identity: %#v", bootstrap)
	}
	if legacyMap(bootstrap["database"]) == nil {
		t.Fatalf("bootstrap missing authorized database: %#v", bootstrap)
	}
}

func TestRemoteInfo508ExposesStableServerIdentity(t *testing.T) {
	s := &Server{cfg: Config{PublicWeb: true, ServerName: "Servidor Juvenil"}, store: Store{ServerID: "server-identity-123"}}
	req := httptest.NewRequest(http.MethodGet, "http://example/api/remote/info", nil)
	rr := httptest.NewRecorder()
	s.remoteInfo(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["serverId"] != "server-identity-123" || payload["serverName"] != "Servidor Juvenil" || payload["version"] != release.Version {
		t.Fatalf("unexpected info payload: %#v", payload)
	}
}

func TestRemoteBootstrap508DoesNotExposeCredentialVerifiers(t *testing.T) {
	dir := t.TempDir()
	s := &Server{
		cfg:            Config{PublicWeb: true, ServerName: "Seguro"},
		path:           filepath.Join(dir, "data", "sgl-server-data.json"),
		store:          Store{ServerID: "server-sensitive-508"},
		remoteSessions: map[string]remoteSession{},
		loginAttempts:  map[string]loginAttempt{},
	}
	adminPassword := testPasswordRecord("Admin123!")
	db := map[string]any{
		"settings": map[string]any{},
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Administrador", "username": "admin", "role": "admin", "active": true, "accountState": "active", "password": adminPassword, "recoveryKeys": []any{testPasswordRecord("RECOVERY-ONE")}},
			map[string]any{"id": "asp1", "fullName": "Lider", "username": "lider", "role": "aspirant", "active": true, "accountState": "active", "password": testPasswordRecord("Lider123!")},
		},
		"audit": []any{}, "classes": []any{}, "tasks": []any{}, "materials": []any{}, "modules": []any{map[string]any{"id": "class_grouped", "active": true, "assignedUserIds": []any{"asp1", "asp2"}}},
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 2, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"identity": "admin", "password": "Admin123!", "authOnly": true, "rememberDevice": true})
	loginReq := httptest.NewRequest(http.MethodPost, "http://example/api/remote/login", bytes.NewReader(body))
	loginReq.RemoteAddr = "203.0.113.90:1000"
	loginRR := httptest.NewRecorder()
	s.remoteLogin(loginRR, loginReq)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", loginRR.Code, loginRR.Body.String())
	}
	var login map[string]any
	if err := json.Unmarshal(loginRR.Body.Bytes(), &login); err != nil {
		t.Fatal(err)
	}
	token, _ := login["sessionToken"].(string)
	bootstrapReq := httptest.NewRequest(http.MethodGet, "http://example/api/remote/bootstrap", nil)
	bootstrapReq.Header.Set("Authorization", "Bearer "+token)
	bootstrapRR := httptest.NewRecorder()
	s.remoteBootstrap(bootstrapRR, bootstrapReq)
	if bootstrapRR.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrapRR.Code, bootstrapRR.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(bootstrapRR.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	for _, raw := range legacyUsers(legacyMap(payload["database"])) {
		u := legacyMap(raw)
		if _, ok := u["password"]; ok {
			t.Fatalf("remote bootstrap leaked password verifier for %v", u["id"])
		}
		if _, ok := u["recoveryKeys"]; ok {
			t.Fatalf("remote bootstrap leaked recovery keys for %v", u["id"])
		}
	}
}

func TestRemoteAdminSync508PreservesHiddenCredentials(t *testing.T) {
	dir := t.TempDir()
	s := &Server{
		cfg:            Config{PublicWeb: true, ServerName: "Seguro"},
		path:           filepath.Join(dir, "data", "sgl-server-data.json"),
		store:          Store{ServerID: "server-preserve-508"},
		remoteSessions: map[string]remoteSession{},
		loginAttempts:  map[string]loginAttempt{},
	}
	adminPassword := testPasswordRecord("Admin123!")
	leaderPassword := testPasswordRecord("Lider123!")
	db := map[string]any{
		"settings": map[string]any{},
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Administrador", "username": "admin", "role": "admin", "active": true, "accountState": "active", "password": adminPassword, "recoveryKeys": []any{testPasswordRecord("RECOVERY-ONE")}},
			map[string]any{"id": "asp1", "fullName": "Lider", "username": "lider", "role": "aspirant", "active": true, "accountState": "active", "password": leaderPassword},
		},
		"audit": []any{}, "classes": []any{}, "tasks": []any{}, "materials": []any{}, "modules": []any{map[string]any{"id": "class_grouped", "active": true, "assignedUserIds": []any{"asp1", "asp2"}}},
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 4, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"identity": "admin", "password": "Admin123!", "authOnly": true, "rememberDevice": true})
	loginReq := httptest.NewRequest(http.MethodPost, "http://example/api/remote/login", bytes.NewReader(body))
	loginReq.RemoteAddr = "203.0.113.91:1000"
	loginRR := httptest.NewRecorder()
	s.remoteLogin(loginRR, loginReq)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", loginRR.Code, loginRR.Body.String())
	}
	var login map[string]any
	_ = json.Unmarshal(loginRR.Body.Bytes(), &login)
	token := login["sessionToken"].(string)

	bootstrapReq := httptest.NewRequest(http.MethodGet, "http://example/api/remote/bootstrap", nil)
	bootstrapReq.Header.Set("Authorization", "Bearer "+token)
	bootstrapRR := httptest.NewRecorder()
	s.remoteBootstrap(bootstrapRR, bootstrapReq)
	var bootstrap map[string]any
	_ = json.Unmarshal(bootstrapRR.Body.Bytes(), &bootstrap)
	clientDB := legacyMap(bootstrap["database"])
	leader := legacyUser(clientDB, "asp1")
	leader["fullName"] = "Lider Renombrado"

	syncBody, _ := json.Marshal(map[string]any{
		"installationId": "remote-admin-508", "actorId": "admin1", "actorRole": "admin", "actorName": "Administrador",
		"baseVersion": int64(4), "database": clientDB, "files": []any{},
	})
	syncReq := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(syncBody))
	syncReq.Header.Set("Authorization", "Bearer "+token)
	syncReq.RemoteAddr = "203.0.113.91:1000"
	syncRR := httptest.NewRecorder()
	s.legacySync(syncRR, syncReq)
	if syncRR.Code != http.StatusOK {
		t.Fatalf("sync status=%d body=%s", syncRR.Code, syncRR.Body.String())
	}
	stored, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	storedLeader := legacyUser(stored.State.Database, "asp1")
	if legacyMap(storedLeader["password"]) == nil {
		t.Fatal("hidden leader password verifier was lost after remote admin sync")
	}
	storedAdmin := legacyUser(stored.State.Database, "admin1")
	if legacyMap(storedAdmin["password"]) == nil || len(arrayField(storedAdmin, "recoveryKeys")) == 0 {
		t.Fatal("hidden administrator credentials were lost after remote admin sync")
	}
	if fmt.Sprint(storedLeader["fullName"]) != "Lider Renombrado" {
		t.Fatalf("ordinary admin edit was not applied: %#v", storedLeader)
	}
}

func TestRemoteAuthIndexRejectsAmbiguousIdentity508(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{PublicWeb: true}, path: filepath.Join(dir, "data", "sgl-server-data.json"), remoteSessions: map[string]remoteSession{}, loginAttempts: map[string]loginAttempt{}}
	db := map[string]any{"users": []any{
		map[string]any{"id": "u1", "fullName": "Uno", "username": "duplicado", "role": "aspirant", "active": true, "accountState": "active", "password": testPasswordRecord("Primero123!")},
		map[string]any{"id": "u2", "fullName": "Dos", "username": "duplicado", "role": "aspirant", "active": true, "accountState": "active", "password": testPasswordRecord("Segundo123!")},
	}}
	if err := s.saveLegacy(legacyStore{State: legacyState{Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"identity": "duplicado", "password": "Segundo123!", "authOnly": true})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/remote/login", bytes.NewReader(body))
	req.RemoteAddr = "203.0.113.92:1000"
	rr := httptest.NewRecorder()
	s.remoteLogin(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("ambiguous identity must be rejected, status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func Test509AdminMaterialIsCatalogLightAndAvailableOnDemandToLeader(t *testing.T) {
	dir := t.TempDir()
	token := "leader-material-token"
	now := time.Now().UTC()
	s := &Server{
		cfg:            Config{PublicWeb: true, ServerName: "Test 5.0.9"},
		path:           filepath.Join(dir, "data", "sgl-server-data.json"),
		remoteSessions: map[string]remoteSession{},
		loginAttempts:  map[string]loginAttempt{},
	}
	s.remoteSessions[hashToken(token)] = remoteSession{
		TokenHash: hashToken(token), UserID: "director1", Role: "director", DeviceID: "phone-leader",
		CreatedAt: now.Format(time.RFC3339), LastSeen: now.Format(time.RFC3339), ExpiresAt: now.Add(2 * time.Hour).Format(time.RFC3339),
	}
	db := map[string]any{
		"settings": map[string]any{},
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "director1", "fullName": "Director", "role": "director", "active": true, "accountState": "active"},
		},
		"audit": []any{}, "classes": []any{}, "tasks": []any{}, "modules": []any{},
		"materials": []any{
			map[string]any{"id": "mat1", "title": "Manual", "createdBy": "admin-historico", "publisherRole": "admin", "visibleToUsers": true, "active": true, "fileId": "mat-file", "fileName": "manual.pdf", "fileType": "application/pdf", "dataUrl": "data:application/pdf;base64,SHOULD_NOT_TRAVEL_IN_CATALOG"},
		},
	}
	files := []any{
		map[string]any{"id": "mat-file", "ownerId": "admin1", "kind": "support_material", "name": "manual.pdf", "type": "application/pdf", "dataUrl": "data:application/pdf;base64,JVBERi0xLjQ="},
		map[string]any{"id": "private-file", "ownerId": "someone-else", "kind": "evidence", "name": "private.pdf", "type": "application/pdf", "dataUrl": "data:application/pdf;base64,UFJJVkFURQ=="},
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 9, Database: db, Files: files}}); err != nil {
		t.Fatal(err)
	}

	bootstrapReq := httptest.NewRequest(http.MethodGet, "https://example.test/api/remote/bootstrap", nil)
	bootstrapReq.Header.Set("Authorization", "Bearer "+token)
	bootstrapRR := httptest.NewRecorder()
	s.remoteBootstrap(bootstrapRR, bootstrapReq)
	if bootstrapRR.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", bootstrapRR.Code, bootstrapRR.Body.String())
	}
	var bootstrap map[string]any
	if err := json.Unmarshal(bootstrapRR.Body.Bytes(), &bootstrap); err != nil {
		t.Fatal(err)
	}
	catalog := legacyMap(bootstrap["database"])
	materials, _ := catalog["materials"].([]any)
	if len(materials) != 1 {
		t.Fatalf("leader should see admin material metadata, got %#v", materials)
	}
	material := legacyMap(materials[0])
	if material["fileId"] != "mat-file" {
		t.Fatalf("unexpected material metadata: %#v", material)
	}
	if _, leaked := material["dataUrl"]; leaked {
		t.Fatal("bootstrap material catalog must not contain dataUrl")
	}

	fileReq := httptest.NewRequest(http.MethodGet, "https://example.test/api/remote/files?fileId=mat-file", nil)
	fileReq.Header.Set("Authorization", "Bearer "+token)
	fileRR := httptest.NewRecorder()
	s.remoteFiles(fileRR, fileReq)
	if fileRR.Code != http.StatusOK {
		t.Fatalf("material file status=%d body=%s", fileRR.Code, fileRR.Body.String())
	}
	var filePayload map[string]any
	if err := json.Unmarshal(fileRR.Body.Bytes(), &filePayload); err != nil {
		t.Fatal(err)
	}
	visible, _ := filePayload["files"].([]any)
	if len(visible) != 1 || legacyMap(visible[0])["id"] != "mat-file" || legacyMap(visible[0])["dataUrl"] == "" {
		t.Fatalf("unexpected on-demand material payload: %#v", visible)
	}

	privateReq := httptest.NewRequest(http.MethodGet, "https://example.test/api/remote/files?fileId=private-file", nil)
	privateReq.Header.Set("Authorization", "Bearer "+token)
	privateRR := httptest.NewRecorder()
	s.remoteFiles(privateRR, privateReq)
	if privateRR.Code != http.StatusNotFound {
		t.Fatalf("unauthorized file must stay hidden, status=%d body=%s", privateRR.Code, privateRR.Body.String())
	}
}

func Test510NotificationReadOnPhoneIsReadOnLaptop(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC()
	phoneToken, laptopToken := "phone-token", "laptop-token"
	s := &Server{cfg: Config{PublicWeb: true}, path: filepath.Join(dir, "data", "sgl-server-data.json"), remoteSessions: map[string]remoteSession{}, loginAttempts: map[string]loginAttempt{}}
	for token, device := range map[string]string{phoneToken: "phone", laptopToken: "laptop"} {
		s.remoteSessions[hashToken(token)] = remoteSession{TokenHash: hashToken(token), UserID: "director1", Role: "director", DeviceID: device, CreatedAt: now.Format(time.RFC3339), LastSeen: now.Format(time.RFC3339), ExpiresAt: now.Add(time.Hour).Format(time.RFC3339)}
	}
	db := map[string]any{"settings": map[string]any{}, "users": []any{map[string]any{"id": "director1", "fullName": "Director", "role": "director", "active": true, "accountState": "active"}}, "audit": []any{}}
	st := legacyStore{State: legacyState{ServerVersion: 1, Database: db, Files: []any{}}, Notifications: []legacyNotification{{ID: "n1", UserID: "director1", Type: "info", Title: "Aviso", Message: "Una sola vez", CreatedAt: now.Format(time.RFC3339)}}}
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]any{"userId": "director1", "notificationIds": []string{"n1"}})
	readReq := httptest.NewRequest(http.MethodPost, "https://example.test/api/notifications/read", bytes.NewReader(body))
	readReq.Header.Set("Authorization", "Bearer "+phoneToken)
	readRR := httptest.NewRecorder()
	s.legacyNotificationRead(readRR, readReq)
	if readRR.Code != http.StatusOK {
		t.Fatalf("phone read status=%d body=%s", readRR.Code, readRR.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "https://example.test/api/notifications?userId=director1", nil)
	getReq.Header.Set("Authorization", "Bearer "+laptopToken)
	getRR := httptest.NewRecorder()
	s.legacyNotifications(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("laptop notifications status=%d body=%s", getRR.Code, getRR.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(getRR.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if int(payload["unread"].(float64)) != 0 {
		t.Fatalf("laptop must not repeat a phone-read notice: %#v", payload)
	}
	notices, _ := payload["notifications"].([]any)
	if len(notices) != 1 || stringValue(legacyMap(notices[0])["readAt"]) == "" {
		t.Fatalf("readAt must be shared by devices: %#v", notices)
	}
}

func Test510AdminCanRepairHistoricalMaterialFileWithoutRepublishingNotice(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC()
	leaderToken, adminSessionToken := "leader-repair-token", "admin-repair-token"
	s := &Server{cfg: Config{AdminToken: "adm", PublicWeb: true}, path: filepath.Join(dir, "data", "sgl-server-data.json"), remoteSessions: map[string]remoteSession{}, loginAttempts: map[string]loginAttempt{}}
	s.remoteSessions[hashToken(leaderToken)] = remoteSession{TokenHash: hashToken(leaderToken), UserID: "director1", Role: "director", DeviceID: "laptop", CreatedAt: now.Format(time.RFC3339), LastSeen: now.Format(time.RFC3339), ExpiresAt: now.Add(time.Hour).Format(time.RFC3339)}
	s.remoteSessions[hashToken(adminSessionToken)] = remoteSession{TokenHash: hashToken(adminSessionToken), UserID: "admin1", Role: "admin", DeviceID: "admin-laptop", CreatedAt: now.Format(time.RFC3339), LastSeen: now.Format(time.RFC3339), ExpiresAt: now.Add(time.Hour).Format(time.RFC3339)}
	db := map[string]any{
		"settings": map[string]any{},
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "director1", "fullName": "Director", "role": "director", "active": true, "accountState": "active"},
		},
		"materials": []any{map[string]any{"id": "mat-old", "title": "Manual histórico", "createdBy": "admin1", "active": true, "fileId": "mat-old-file", "fileName": "manual.pdf"}},
		"audit":     []any{}, "modules": []any{}, "tasks": []any{}, "classes": []any{},
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 5, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	file := map[string]any{"id": "mat-old-file", "ownerId": "admin1", "kind": "support_material", "name": "manual.pdf", "type": "application/pdf", "dataUrl": "data:application/pdf;base64,JVBERi0xLjQ="}
	body, _ := json.Marshal(map[string]any{"actorId": "admin1", "operationId": "SGL-MATFILE-REPAIR-1", "summary": "Reparación silenciosa", "patch": map[string]any{}, "files": []any{file}})
	repairReq := httptest.NewRequest(http.MethodPost, "https://example.test/api/admin/changes", bytes.NewReader(body))
	repairReq.Header.Set("Authorization", "Bearer "+adminSessionToken)
	repairRR := httptest.NewRecorder()
	s.legacyAdminChangeSubmit(repairRR, repairReq)
	if repairRR.Code != http.StatusOK {
		t.Fatalf("repair status=%d body=%s", repairRR.Code, repairRR.Body.String())
	}
	spoofBody, _ := json.Marshal(map[string]any{"actorId": "other-admin", "operationId": "SGL-MATFILE-SPOOF", "patch": map[string]any{}, "files": []any{file}})
	spoofReq := httptest.NewRequest(http.MethodPost, "https://example.test/api/admin/changes", bytes.NewReader(spoofBody))
	spoofReq.Header.Set("Authorization", "Bearer "+adminSessionToken)
	spoofRR := httptest.NewRecorder()
	s.legacyAdminChangeSubmit(spoofRR, spoofReq)
	if spoofRR.Code != http.StatusForbidden {
		t.Fatalf("remote admin session must be bound to its actor, status=%d body=%s", spoofRR.Code, spoofRR.Body.String())
	}
	stored, _ := s.loadLegacy()
	if len(stored.State.Files) != 1 || len(stored.Notifications) != 0 {
		t.Fatalf("file repair must not republish notices: files=%#v notices=%#v", stored.State.Files, stored.Notifications)
	}

	manifestReq := httptest.NewRequest(http.MethodGet, "https://example.test/api/remote/files?manifest=1", nil)
	manifestReq.Header.Set("Authorization", "Bearer "+leaderToken)
	manifestRR := httptest.NewRecorder()
	s.remoteFiles(manifestRR, manifestReq)
	if manifestRR.Code != http.StatusOK {
		t.Fatalf("manifest status=%d body=%s", manifestRR.Code, manifestRR.Body.String())
	}
	var manifest map[string]any
	_ = json.Unmarshal(manifestRR.Body.Bytes(), &manifest)
	manifestFiles, _ := manifest["files"].([]any)
	if len(manifestFiles) != 1 {
		t.Fatalf("manifest missing repaired material: %#v", manifest)
	}
	if _, leaked := legacyMap(manifestFiles[0])["dataUrl"]; leaked {
		t.Fatal("manifest must not include file content")
	}

	fileReq := httptest.NewRequest(http.MethodGet, "https://example.test/api/remote/files?fileId=mat-old-file", nil)
	fileReq.Header.Set("Authorization", "Bearer "+leaderToken)
	fileRR := httptest.NewRecorder()
	s.remoteFiles(fileRR, fileReq)
	if fileRR.Code != http.StatusOK {
		t.Fatalf("leader repaired material status=%d body=%s", fileRR.Code, fileRR.Body.String())
	}
}

func Test514TutorMaterialSyncReachesOnlyAssignedAspirant(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().UTC()
	tutorToken := "tutor-material-514"
	aspToken := "asp-material-514"
	otherToken := "other-material-514"
	s := &Server{
		cfg:            Config{PublicWeb: true, ServerName: "Test 5.0.30"},
		path:           filepath.Join(dir, "data", "sgl-server-data.json"),
		remoteSessions: map[string]remoteSession{},
		loginAttempts:  map[string]loginAttempt{},
	}
	for tokenValue, session := range map[string]remoteSession{
		tutorToken: {TokenHash: hashToken(tutorToken), UserID: "t1", Role: "tutor", DeviceID: "tutor-device", CreatedAt: now.Format(time.RFC3339), LastSeen: now.Format(time.RFC3339), ExpiresAt: now.Add(2 * time.Hour).Format(time.RFC3339)},
		aspToken:   {TokenHash: hashToken(aspToken), UserID: "asp1", Role: "aspirant", DeviceID: "asp-device", CreatedAt: now.Format(time.RFC3339), LastSeen: now.Format(time.RFC3339), ExpiresAt: now.Add(2 * time.Hour).Format(time.RFC3339)},
		otherToken: {TokenHash: hashToken(otherToken), UserID: "asp2", Role: "aspirant", DeviceID: "other-device", CreatedAt: now.Format(time.RFC3339), LastSeen: now.Format(time.RFC3339), ExpiresAt: now.Add(2 * time.Hour).Format(time.RFC3339)},
	} {
		s.remoteSessions[hashToken(tokenValue)] = session
	}
	db := map[string]any{
		"settings": map[string]any{},
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "t1", "fullName": "Tutor Uno", "role": "tutor", "active": true, "accountState": "active"},
			map[string]any{"id": "t2", "fullName": "Tutor Dos", "role": "tutor", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "role": "aspirant", "tutorId": "t1", "active": true, "accountState": "active"},
			map[string]any{"id": "asp2", "fullName": "Aspirante Dos", "role": "aspirant", "tutorId": "t2", "active": true, "accountState": "active"},
		},
		"audit": []any{}, "classes": []any{}, "tasks": []any{}, "modules": []any{map[string]any{"id": "class_grouped", "active": true, "assignedUserIds": []any{"asp1", "asp2"}}}, "materials": []any{},
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 4, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}

	material := map[string]any{"id": "mat-tutor-514", "title": "Material del tutor", "createdBy": "t1", "publisherRole": "tutor", "audienceTutorId": "t1", "visibleToUsers": false, "active": true, "programId": "class_grouped", "fileId": "mat-file-514", "fileName": "guia.pdf", "fileType": "application/pdf"}
	file := map[string]any{"id": "mat-file-514", "ownerId": "t1", "kind": "support_material", "programId": "class_grouped", "name": "guia.pdf", "type": "application/pdf", "dataUrl": "data:application/pdf;base64,JVBERi0xLjQ="}
	syncBody, _ := json.Marshal(map[string]any{
		"installationId": "tutor-device", "actorId": "t1", "actorRole": "tutor", "baseVersion": 4,
		"operationId": "material-publish-514", "operationType": "material_publish", "ackOnly": true,
		"patch": map[string]any{"materials": []any{material}}, "files": []any{file},
	})
	syncReq := httptest.NewRequest(http.MethodPost, "https://example.test/api/sync", bytes.NewReader(syncBody))
	syncReq.Header.Set("Authorization", "Bearer "+tutorToken)
	syncRR := httptest.NewRecorder()
	s.legacySync(syncRR, syncReq)
	if syncRR.Code != http.StatusOK {
		t.Fatalf("tutor material sync status=%d body=%s", syncRR.Code, syncRR.Body.String())
	}

	bootstrapReq := httptest.NewRequest(http.MethodGet, "https://example.test/api/remote/bootstrap", nil)
	bootstrapReq.Header.Set("Authorization", "Bearer "+aspToken)
	bootstrapRR := httptest.NewRecorder()
	s.remoteBootstrap(bootstrapRR, bootstrapReq)
	if bootstrapRR.Code != http.StatusOK {
		t.Fatalf("aspirant bootstrap status=%d body=%s", bootstrapRR.Code, bootstrapRR.Body.String())
	}
	var bootstrap map[string]any
	if err := json.Unmarshal(bootstrapRR.Body.Bytes(), &bootstrap); err != nil {
		t.Fatal(err)
	}
	catalog := legacyMap(bootstrap["database"])
	materials, _ := catalog["materials"].([]any)
	if len(materials) != 1 || stringValue(legacyMap(materials[0])["id"]) != "mat-tutor-514" {
		t.Fatalf("assigned aspirant should receive tutor material metadata, got %#v", materials)
	}

	fileReq := httptest.NewRequest(http.MethodGet, "https://example.test/api/remote/files?fileId=mat-file-514", nil)
	fileReq.Header.Set("Authorization", "Bearer "+aspToken)
	fileRR := httptest.NewRecorder()
	s.remoteFiles(fileRR, fileReq)
	if fileRR.Code != http.StatusOK {
		t.Fatalf("assigned aspirant material file status=%d body=%s", fileRR.Code, fileRR.Body.String())
	}
	var filePayload map[string]any
	if err := json.Unmarshal(fileRR.Body.Bytes(), &filePayload); err != nil {
		t.Fatal(err)
	}
	visibleFiles, _ := filePayload["files"].([]any)
	if len(visibleFiles) != 1 || stringValue(legacyMap(visibleFiles[0])["id"]) != "mat-file-514" {
		t.Fatalf("assigned aspirant should receive tutor material file, got %#v", visibleFiles)
	}

	otherReq := httptest.NewRequest(http.MethodGet, "https://example.test/api/remote/bootstrap", nil)
	otherReq.Header.Set("Authorization", "Bearer "+otherToken)
	otherRR := httptest.NewRecorder()
	s.remoteBootstrap(otherRR, otherReq)
	if otherRR.Code != http.StatusOK {
		t.Fatalf("other aspirant bootstrap status=%d body=%s", otherRR.Code, otherRR.Body.String())
	}
	var otherPayload map[string]any
	if err := json.Unmarshal(otherRR.Body.Bytes(), &otherPayload); err != nil {
		t.Fatal(err)
	}
	otherMaterials, _ := legacyMap(otherPayload["database"])["materials"].([]any)
	if len(otherMaterials) != 0 {
		t.Fatalf("unassigned aspirant must not receive another tutor material, got %#v", otherMaterials)
	}
}

func TestRemoteFiles5048ManifestAndSingleFileStaySelective(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{PublicWeb: true, ServerName: "Test"}, path: filepath.Join(dir, "data", "sgl-server-data.json"), remoteSessions: map[string]remoteSession{}, loginAttempts: map[string]loginAttempt{}}
	db := map[string]any{
		"settings": map[string]any{},
		"users":    []any{map[string]any{"id": "asp1", "fullName": "Aspirante", "username": "asp1", "role": "aspirant", "active": true, "accountState": "active", "password": testPasswordRecord("Asp12345!")}},
		"progress": map[string]any{"asp1": map[string]any{}}, "classes": []any{}, "tasks": []any{}, "materials": []any{}, "modules": []any{map[string]any{"id": "class_grouped", "active": true, "assignedUserIds": []any{"asp1", "asp2"}}}, "audit": []any{},
	}
	large := "data:image/jpeg;base64," + strings.Repeat("A", 4<<20)
	st := legacyStore{State: legacyState{ServerVersion: 48, Database: db, Files: []any{
		map[string]any{"id": "photo-1", "ownerId": "asp1", "kind": "photo", "name": "foto.jpg", "type": "image/jpeg", "size": 3 << 20, "dataUrl": large},
		map[string]any{"id": "other", "ownerId": "other", "kind": "photo", "name": "privada.jpg", "type": "image/jpeg", "dataUrl": large},
	}}}
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"identity": "asp1", "password": "Asp12345!", "authOnly": true})
	loginReq := httptest.NewRequest(http.MethodPost, "http://example/api/remote/login", bytes.NewReader(body))
	loginReq.RemoteAddr = "203.0.113.20:9000"
	loginRR := httptest.NewRecorder()
	s.remoteLogin(loginRR, loginReq)
	if loginRR.Code != 200 {
		t.Fatalf("login=%d %s", loginRR.Code, loginRR.Body.String())
	}
	var login map[string]any
	if err := json.Unmarshal(loginRR.Body.Bytes(), &login); err != nil {
		t.Fatal(err)
	}
	token := fmt.Sprint(login["sessionToken"])

	manifestReq := httptest.NewRequest(http.MethodGet, "http://example/api/remote/files?manifest=1", nil)
	manifestReq.Header.Set("Authorization", "Bearer "+token)
	manifestRR := httptest.NewRecorder()
	s.remoteFiles(manifestRR, manifestReq)
	if manifestRR.Code != 200 {
		t.Fatalf("manifest=%d %s", manifestRR.Code, manifestRR.Body.String())
	}
	if manifestRR.Body.Len() > 32<<10 {
		t.Fatalf("manifest unexpectedly heavy: %d bytes", manifestRR.Body.Len())
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestRR.Body.Bytes(), &manifest); err != nil {
		t.Fatal(err)
	}
	items, _ := manifest["files"].([]any)
	if len(items) != 1 || fmt.Sprint(legacyMap(items[0])["id"]) != "photo-1" {
		t.Fatalf("unexpected manifest: %#v", items)
	}
	if _, ok := legacyMap(items[0])["dataUrl"]; ok {
		t.Fatal("manifest leaked dataUrl")
	}

	fileReq := httptest.NewRequest(http.MethodGet, "http://example/api/remote/files?fileId=photo-1", nil)
	fileReq.Header.Set("Authorization", "Bearer "+token)
	fileRR := httptest.NewRecorder()
	s.remoteFiles(fileRR, fileReq)
	if fileRR.Code != 200 {
		t.Fatalf("single=%d %s", fileRR.Code, fileRR.Body.String())
	}
	var one map[string]any
	if err := json.Unmarshal(fileRR.Body.Bytes(), &one); err != nil {
		t.Fatal(err)
	}
	files, _ := one["files"].([]any)
	if len(files) != 1 || fmt.Sprint(legacyMap(files[0])["dataUrl"]) != large {
		t.Fatal("single file did not return payload")
	}
}
