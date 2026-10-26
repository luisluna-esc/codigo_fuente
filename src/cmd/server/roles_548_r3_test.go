package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func fixture548R3(t *testing.T) (*Server, map[string]any) {
	t.Helper()
	db := map[string]any{"users": []any{
		map[string]any{"id": "d", "role": "departamental", "active": true, "password": "secret"},
		map[string]any{"id": "r", "role": "administrativo", "active": true, "password": "secret"},
		map[string]any{"id": "p", "role": "pastor_aspirant", "fullName": "Pastor", "username": "pastor", "email": "fijo@test", "phone": "nuevo desde otro equipo", "active": true},
	}, "settings": map[string]any{"googleClientSecret": "secret"}, "progress": map[string]any{"p": map[string]any{"status": map[string]any{"X": "approved"}}}, "reportRequirementModes": map[string]any{"music_ministry:X": "document"}, "indicationRules": []any{map[string]any{"id": "rule1", "programId": "music_ministry", "requirementId": "X", "scope": "all", "active": true, "updatedAt": "before", "config": map[string]any{"instructions": "Inicial", "mode": "document"}}}}
	testEnrollConquistadores593(db)
	s := &Server{path: filepath.Join(t.TempDir(), "data", "sgl-server-data.json"), cfg: Config{ClientToken: "client"}, remoteSessions: map[string]remoteSession{}, loginAttempts: map[string]loginAttempt{}}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 4, RestoreGeneration: 1, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	return s, db
}
func Test548R3ReaderSeesEveryProgressWithoutMutatingCredentials(t *testing.T) {
	s, db := fixture548R3(t)
	for _, role := range []string{"departamental", "administrativo"} {
		p := s.filteredLegacySnapshot(legacyState{Database: db}, "d", role)
		out := legacyMap(p["database"])
		if len(legacyUsers(out)) != 3 || legacyMap(out["progress"])["p"] == nil {
			t.Fatal("full scope missing")
		}
		if legacyUser(out, "d")["password"] != nil || legacyMap(out["settings"])["googleClientSecret"] != nil {
			t.Fatal("secret exposed")
		}
	}
	if legacyUser(db, "d")["password"] != "secret" {
		t.Fatal("read request changed original credentials")
	}
}
func Test548R3ReadOnlyBlocksRemoteMutation(t *testing.T) {
	s, _ := fixture548R3(t)
	for _, role := range []string{"departamental", "administrativo"} {
		raw := "session" + role
		s.remoteSessions[hashToken(raw)] = remoteSession{UserID: "r", Role: role, ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339)}
		for _, path := range []string{"/api/sync", "/api/google/upload", "/api/reviews/decision", "/api/admin/changes"} {
			r := httptest.NewRequest("POST", path, bytes.NewBufferString("{}"))
			r.Header.Set("Authorization", "Bearer "+raw)
			w := httptest.NewRecorder()
			s.securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })).ServeHTTP(w, r)
			if w.Code != 403 {
				t.Fatalf("%s allowed %s: %d", role, path, w.Code)
			}
		}
	}
}
func Test548R3PastorProfileDeltaPreservesOtherDeviceAndFixedEmail(t *testing.T) {
	_, db := fixture548R3(t)
	source := legacyMap(cloneAny(db))
	p := legacyUser(source, "p")
	p["phone"] = "stale"
	p["email"] = "changed"
	p["role"] = "admin"
	p["fullName"] = "Pastor actualizado"
	p["profileChangedFields548"] = []any{"fullName"}
	merged, _, err := mergeLegacyProposalOnServer(db, &legacyProposal{ActorID: "p", ActorRole: "pastor_aspirant", Database: source})
	if err != nil {
		t.Fatal(err)
	}
	u := legacyUser(merged, "p")
	if u["fullName"] != "Pastor actualizado" || u["phone"] != "nuevo desde otro equipo" || u["email"] != "fijo@test" || u["role"] != "pastor_aspirant" {
		t.Fatalf("unsafe merge: %v", u)
	}
}
func Test548R3IndicationsRejectStaleAndPreserveConfiguration(t *testing.T) {
	s, _ := fixture548R3(t)
	call := func(actor, base, text string) int {
		body, _ := json.Marshal(map[string]any{"actorId": actor, "restoreGeneration": 1, "edits": []any{map[string]any{"id": "rule1", "programId": "music_ministry", "requirementId": "X", "baseUpdatedAt": base, "instructions": text, "mode": "written"}}, "settings": map[string]any{"arbitrary": true}})
		r := httptest.NewRequest("POST", "/api/indications", bytes.NewReader(body))
		r.Header.Set("X-SGL-Token", "client")
		w := httptest.NewRecorder()
		s.indications548(w, r)
		return w.Code
	}
	if call("r", "before", "No") != 403 {
		t.Fatal("administrativo can edit")
	}
	if call("d", "before", "Revisada") != 200 {
		t.Fatal("department save failed")
	}
	if call("d", "before", "Otra") != 409 {
		t.Fatal("stale change accepted")
	}
	if call("d", "before", "Revisada") != 200 {
		t.Fatal("retry not idempotent")
	}
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	rule := legacyMap(arrayField(db, "indicationRules")[0])
	if legacyMap(rule["config"])["mode"] != "document" || legacyMap(db["reportRequirementModes"])["music_ministry:X"] != "document" || legacyMap(legacyMap(legacyMap(db["progress"])["p"])["status"])["X"] != "approved" {
		t.Fatal("instruction edit changed configuration/progress")
	}
}
func Test548R3LoginLockLastsTwoMinutes(t *testing.T) {
	s, _ := fixture548R3(t)
	for i := 0; i < 5; i++ {
		s.loginFailed("key")
	}
	ok, wait := s.loginAllowed("key")
	if ok || wait < 119*time.Second || wait > 120*time.Second {
		t.Fatalf("wrong lock duration %v %v", ok, wait)
	}
}
