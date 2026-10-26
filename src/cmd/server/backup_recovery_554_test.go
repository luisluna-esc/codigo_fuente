package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func fixtureBackup554(t *testing.T) (*Server, recoveryCapsule548) {
	s, c := recoveryFixture548(t)
	s.cfg.WebRoot = "../../../app"
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	u := legacyUser(db, "a")
	u["email"] = "ana@example.test"
	u["username"] = "ana"
	db["modules"] = []any{map[string]any{"id": "class_grouped", "active": true, "assignedUserIds": []any{"a"}}}
	db["settings"] = map[string]any{"serverToken": "keep-secret", "theme": "blue"}
	db["users"] = append(legacyUsers(db), map[string]any{"id": "other", "role": "aspirant", "password": map[string]any{"hash": "keep"}, "active": true})
	ensureMapField(db, "progress")["other"] = map[string]any{"status": map[string]any{"I-01": "approved"}, "submissions": map[string]any{"I-01": map[string]any{"text": "other report"}}}
	s.saveLegacy(st)
	c.Progress = map[string]any{"status": map[string]any{"I-01": "in_progress", "I-02": "pending", "OLD-ID": "in_progress"}, "submissions": map[string]any{"I-01": map[string]any{"text": "recovered", "richText": "<p>recovered</p>", "attachmentIds": []any{"photo"}}, "I-02": map[string]any{"text": ""}, "OLD-ID": map[string]any{"text": "old unmatched"}}}
	c.SourceUser = map[string]any{"id": "a", "email": "ana@example.test", "username": "ana"}
	c.SourceVersion = "5.30"
	c.Programs = map[string]any{"guide_major": map[string]any{"status": map[string]any{"GM-CPE-1": "approved"}}}
	c.Files = []any{map[string]any{"id": "photo", "ownerId": "a", "programId": "class_grouped", "requirementId": "I-01", "kind": "photo", "type": "image/png", "size": 5, "dataUrl": "data:image/png;base64,cGhvdG8="}}
	return s, c
}
func Test554RecoveryIdentityCatalogAndUnchangedSystem(t *testing.T) {
 t.Skip("La recuperación individual antigua fue retirada; la importación exclusiva de Música tiene pruebas independientes")
	s, c := fixtureBackup554(t)
	before, _ := s.loadLegacy()
	original, _ := json.Marshal(before)
	preview := recoveryCall548(s.adminRecovery548, "adm", recoveryRequest548{Capsule: c})
	if preview.Code != 200 {
		t.Fatal(preview.Body.String())
	}
	var result map[string]any
	json.Unmarshal(preview.Body.Bytes(), &result)
	if numericValue(result["fileCount"]) != 1 {
		t.Fatalf("filtered files in preview: %s", preview.Body.String())
	}
	rows := arrayField(result, "rows")
	if len(rows) != 1 || stringValue(legacyMap(rows[0])["requirementId"]) != "I-01" {
		t.Fatalf("unexpected rows %s", preview.Body.String())
	}
	if len(arrayField(result, "skipped")) != 2 {
		t.Fatalf("unsupported requirement/program not reported: %s", preview.Body.String())
	}
	afterPreview, _ := s.loadLegacy()
	b, _ := json.Marshal(afterPreview)
	if string(b) != string(original) {
		t.Fatal("preview changed data")
	}
	commit := recoveryCall548(s.adminRecovery548, "adm", recoveryRequest548{Capsule: c, Commit: true, BaseVersion: 10, Choices: map[string]string{"class_grouped|I-01": "local"}})
	if commit.Code != 200 {
		t.Fatal(commit.Body.String())
	}
	after, _ := s.loadLegacy()
	a := legacyMap(after.State.Database)
	old := legacyMap(before.State.Database)
	for _, field := range []string{"users", "modules", "settings", "programProgress"} {
		if !reflect.DeepEqual(a[field], old[field]) {
			t.Fatalf("changed %s", field)
		}
	}
	if !reflect.DeepEqual(mapField(a, "progress")["other"], mapField(old, "progress")["other"]) {
		t.Fatal("changed other owner")
	}
	if after.State.RestoreGeneration != before.State.RestoreGeneration || after.State.RestoreID != before.State.RestoreID {
		t.Fatal("changed restore generation")
	}
	p := legacyMap(mapField(a, "progress")["a"])
	if stringValue(mapField(p, "status")["I-01"]) != "in_progress" || mapField(p, "status")["OLD-ID"] != nil {
		t.Fatal("bad restored progress")
	}
	if len(after.State.Files) != 1 {
		t.Fatalf("missing recovered photo: %#v", after.State.Files)
	}
}
func Test554WrongAccountAndForeignFilesCannotRecover(t *testing.T) {
	for _, kind := range []string{"identity", "owner", "files"} {
		t.Run(kind, func(t *testing.T) {
			s, c := fixtureBackup554(t)
			switch kind {
			case "identity":
				c.SourceUser["email"] = "different@example.test"
			case "owner":
				c.SourceUser["id"] = "other"
			case "files":
				legacyMap(c.Files[0])["ownerId"] = "other"
			}
			result := recoveryCall548(s.adminRecovery548, "adm", recoveryRequest548{Capsule: c, Commit: true, BaseVersion: 10, Choices: map[string]string{"class_grouped|I-01": "local"}})
			if result.Code != 400 {
				t.Fatalf("not rejected %d", result.Code)
			}
			st, _ := s.loadLegacy()
			if st.State.ServerVersion != 10 || len(st.Recoveries548) != 0 || len(st.State.Files) != 0 {
				t.Fatal("failed import mutated system")
			}
		})
	}
}
func Test554RemappedIDRequiresUniqueMatchingIdentity(t *testing.T) {
	s, c := fixtureBackup554(t)
	c.OwnerID = "old-a"
	c.SourceUser["id"] = "old-a"
	legacyMap(c.Files[0])["ownerId"] = "old-a"
	if w := recoveryCall548(s.adminRecovery548, "adm", recoveryRequest548{Capsule: c}); w.Code != 200 || !strings.Contains(w.Body.String(), `"ownerId":"a"`) {
		t.Fatal(w.Body.String())
	}
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	legacyUser(db, "other")["email"] = "ana@example.test"
	s.saveLegacy(st)
	if w := recoveryCall548(s.adminRecovery548, "adm", recoveryRequest548{Capsule: c}); w.Code != 400 {
		t.Fatal("ambiguous account accepted")
	}
}
func Test554ReviewedReadingsDoNotIssueWhiteEventCertificate(t *testing.T) {
	s, _ := fixture550(t)
	if w := call550(t, s, "p", "p", "certificate", map[string]any{"kind": "egw_reading"}, false); w.Code != 400 {
		t.Fatal("reviewed requirements issued an event certificate without an event", w.Code)
	}
}

// Optional local fixture validates a user-provided backup; no personal copy is
// bundled with this source tree, application assets or executable.
func Test554ProvidedBackup(t *testing.T) {
	path := os.Getenv("SGL_BACKUP_TEST_PATH")
	if path == "" {
		t.Skip("optional local backup fixture")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if json.Unmarshal(data, &payload) != nil {
		t.Fatal("bad backup")
	}
	db := legacyMap(payload["database"])
	owner := stringValue(mapField(payload, "exportedBy")["id"])
	u := legacyUser(db, owner)
	s, _ := recoveryFixture548(t)
	s.cfg.WebRoot = "../../../app"
	db["users"] = append(legacyUsers(db), map[string]any{"id": "other", "active": true, "role": "aspirant"})
	db["settings"] = map[string]any{"sentinel": "keep"}
	progress := cloneAny(mapField(db, "progress")[owner])
	programs := cloneAny(mapField(db, "programProgress")[owner])
	ensureMapField(db, "progress")[owner] = map[string]any{}
	st, _ := s.loadLegacy()
	st.State.Database = db
	st.State.Files = []any{}
	st.State.RestoreGeneration = 4
	st.State.RestoreID = "current"
	s.saveLegacy(st)
	c := recoveryCapsule548{Format: "SGLRecovery548", OwnerID: owner, InstallationID: stringValue(db["installationId"]), Secret: strings.Repeat("a", 64), Fingerprint: strings.Repeat("b", 64), Progress: legacyMap(progress), Programs: legacyMap(programs), Bible: mapField(db, "bibleYear")[owner], Files: arrayField(payload, "files"), SourceUser: map[string]any{"id": owner, "email": u["email"], "username": u["username"]}, SourceVersion: stringValue(payload["version"])}
	w := recoveryCall548(s.adminRecovery548, "adm", recoveryRequest548{Capsule: c})
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	var preview map[string]any
	json.Unmarshal(w.Body.Bytes(), &preview)
	choices := map[string]string{}
	for _, raw := range arrayField(preview, "rows") {
		choices[stringValue(legacyMap(raw)["key"])] = "local"
	}
	t.Logf("individual fixture: %d requirement differences, %d files", len(choices), len(c.Files))
	w = recoveryCall548(s.adminRecovery548, "adm", recoveryRequest548{Capsule: c, Commit: true, BaseVersion: 10, Choices: choices})
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	after, _ := s.loadLegacy()
	if len(after.State.Files) != len(c.Files) || after.State.RestoreGeneration != 4 || legacyUser(after.State.Database, "other") == nil || stringValue(mapField(legacyMap(after.State.Database), "settings")["sentinel"]) != "keep" {
		t.Fatal("provided backup isolation failed")
	}
}

func Test554ProvidedBackupBrowser(t *testing.T) {
	if os.Getenv("SGL_RECOVERY_BROWSER_RUN") != "1" {
		t.Skip("optional browser fixture")
	}
	data, err := os.ReadFile(os.Getenv("SGL_BACKUP_TEST_PATH"))
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	json.Unmarshal(data, &payload)
	s, _ := recoveryFixture548(t)
	db := legacyMap(payload["database"])
	owner := stringValue(mapField(payload, "exportedBy")["id"])
	u := legacyUser(db, owner)
	u["password"] = testPasswordRecord("Prueba123!")
	db["users"] = append(legacyUsers(db), map[string]any{"id": "admin-test", "username": "admin-test", "fullName": "Administrador de prueba", "active": true, "role": "admin", "password": testPasswordRecord("Prueba123!")}, map[string]any{"id": "other", "active": true, "role": "aspirant"})
	ensureMapField(db, "progress")[owner] = map[string]any{}
	db["settings"] = map[string]any{"autoSync": false}
	db = legacyMap(cloneAny(db))
	st, _ := s.loadLegacy()
	st.State.Database = db
	st.State.Files = []any{}
	st.State.RestoreGeneration = 4
	st.State.RestoreID = "current"
	s.saveLegacy(st)
	s.cfg.WebRoot = "../../../app"
	s.cfg.PublicWeb = true
	s.remoteSessions = map[string]remoteSession{}
	s.loginAttempts = map[string]loginAttempt{}
	mux := http.NewServeMux()
	s.routes(mux)
	s.webRoutes(mux)
	server := httptest.NewServer(s.cors(s.securityHeaders(mux)))
	defer server.Close()
	cmd := exec.Command("node", "../../../tests/browser-recovery-554.cjs")
	cmd.Env = append(os.Environ(), "SGL_BROWSER_BASE="+server.URL)
	out, err := cmd.CombinedOutput()
	t.Log(string(out))
	if err != nil {
		t.Fatal(err)
	}
	after, _ := s.loadLegacy()
	if after.State.RestoreGeneration != 4 || legacyUser(after.State.Database, "other") == nil || len(after.State.Files) != 15 {
		t.Fatal("browser recovery changed another account or restore state")
	}
}

func Test554ComparisonIgnoresOtherAccountsButRejectsSameOwnerChanges(t *testing.T) {
	for _, sameOwner := range []bool{false, true} {
		s, c := fixtureBackup554(t)
		st, _ := s.loadLegacy()
		fingerprint := recoveryStateFingerprint554(st.State, "a")
		db := legacyMap(st.State.Database)
		owner := "other"
		if sameOwner {
			owner = "a"
		}
		ensureMapField(ensureMapField(ensureMapField(db, "progress"), owner), "submissions")["I-01"] = map[string]any{"text": "new server work"}
		st.State.ServerVersion++
		s.saveLegacy(st)
		w := recoveryCall548(s.adminRecovery548, "adm", recoveryRequest548{Capsule: c, Commit: true, BaseVersion: 10, BaseFingerprint: fingerprint, Choices: map[string]string{"class_grouped|I-01": "local"}})
		if sameOwner && w.Code != 409 {
			t.Fatal("concurrent same-owner edit accepted")
		}
		if !sameOwner && w.Code != 200 {
			t.Fatalf("unrelated work blocked recovery: %s", w.Body.String())
		}
		after, _ := s.loadLegacy()
		text := stringValue(legacyMap(mapField(legacyMap(mapField(legacyMap(after.State.Database), "progress")[owner]), "submissions")["I-01"])["text"])
		if text != "new server work" {
			t.Fatal("concurrent work overwritten")
		}
	}
}
