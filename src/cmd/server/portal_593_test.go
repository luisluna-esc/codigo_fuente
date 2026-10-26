package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func Test593PortalIsolationAndDualMembership(t *testing.T) {
	s, h := musicFixture553(t)
	st, e := s.loadLegacy()
	if e != nil {
		t.Fatal(e)
	}
	db := legacyMap(st.State.Database)
	legacyUser(db, "administrativo")["managementProgramIds"] = []any{musicProgram553}
	if e = s.saveLegacy(st); e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		id     string
		music  bool
		status int
	}{
		{"administrativo", false, 403}, {"administrativo", true, 200},
		{"b", false, 403}, {"b", true, 200}, {"a", false, 200}, {"a", true, 200},
		{"c", false, 200}, {"c", true, 403}, {"admin", false, 200}, {"admin", true, 200},
	} {
		code, _ := musicLogin553(t, h, tc.id, tc.music)
		if code != tc.status {
			t.Fatalf("%s music=%v: got %d want %d", tc.id, tc.music, code, tc.status)
		}
	}
	raw := "legacy-music-manager-session"
	s.remoteSessions[hashToken(raw)] = remoteSession{TokenHash: hashToken(raw), UserID: "administrativo", Role: "administrativo", ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339)}
	for _, path := range []string{"/api/bootstrap", "/api/remote/session", "/api/remote/files"} {
		code, _ := musicCall553(t, h, "GET", path, raw, nil)
		if code != 403 {
			t.Fatalf("old token on %s: %d", path, code)
		}
	}
}
func Test593NewEntityResolvesAssignmentDelta(t *testing.T) {
	got := mergeLegacyPatchValue([]any{}, []any{map[string]any{"id": "new", "managementProgramIds": map[string]any{"__sglSetMembers": map[string]any{"add": []any{musicProgram553}, "remove": []any{}}}}})
	users := got.([]any)
	if !containsString(arrayField(legacyMap(users[0]), "managementProgramIds"), musicProgram553) {
		t.Fatal("unresolved nested delta", got)
	}
}

func Test593RecoveryPreservesUnrelatedData(t *testing.T) {
	var plan struct{ RemoveClassIds, RestoreMusicIds []string }
	if err := json.Unmarshal([]byte(incident593), &plan); err != nil {
		t.Fatal(err)
	}
	users := []any{}
	for _, id := range plan.RestoreMusicIds {
		users = append(users, map[string]any{"id": id, "active": true, "tutorByProgram": map[string]any{musicProgram553: "original-tutor"}})
	}
	cls := []any{"legitimate-dual"}
	for _, id := range plan.RemoveClassIds {
		cls = append(cls, id)
	}
	db := map[string]any{"users": users, "modules": []any{map[string]any{"id": "class_grouped", "assignedUserIds": cls}, map[string]any{"id": musicProgram553, "assignedUserIds": []any{"legitimate-dual"}}}, "programProgress": map[string]any{"evidence": "unchanged"}}
	before := cloneAny(db)
	if !repairDatabase593(db) {
		t.Fatal("repair did not run")
	}
	if repairDatabase593(db) {
		t.Fatal("repair ran twice")
	}
	modules := arrayField(db, "modules")
	if len(arrayField(legacyMap(modules[0]), "assignedUserIds")) != 1 {
		t.Fatal("wrong class list")
	}
	if !reflect.DeepEqual(legacyMap(before)["users"], db["users"]) || !reflect.DeepEqual(legacyMap(before)["programProgress"], db["programProgress"]) {
		t.Fatal("unrelated data changed")
	}
	s := &Server{path: filepath.Join(t.TempDir(), "state.json")}
	if err := s.saveLegacy(legacyStore{State: legacyState{Database: before, Files: []any{map[string]any{"id": "file", "dataUrl": "preserved"}}, RestoreGeneration: 2, ServerVersion: 9}}); err != nil {
		t.Fatal(err)
	}
	if err := s.repairIncident593(); err != nil {
		t.Fatal(err)
	}
	after, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if after.State.RestoreGeneration != 3 || after.State.ServerVersion != 10 {
		t.Fatal("sync generation not advanced")
	}
	backups, _ := filepath.Glob(filepath.Join(filepath.Dir(s.path), "Backups", "*.json"))
	if len(backups) != 1 {
		t.Fatal("backup missing")
	}
	data, err := os.ReadFile(backups[0])
	if err != nil {
		t.Fatal(err)
	}
	var original legacyStore
	if json.Unmarshal(data, &original) != nil {
		t.Fatal("invalid backup")
	}
	if legacyMap(original.State.Database)["incidentMusic593"] == true {
		t.Fatal("backup already repaired")
	}
	if !reflect.DeepEqual(original.State.Files, after.State.Files) {
		t.Fatal("files changed")
	}
}

// Old feature fixtures call general endpoints, so they explicitly enroll their
// test accounts there. Portal-isolation fixtures do not call this helper.
func testEnrollConquistadores593(db map[string]any) {
	ids := []any{}
	for _, raw := range legacyUsers(db) {
		ids = append(ids, legacyMap(raw)["id"])
	}
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if m["id"] == "class_grouped" {
			m["assignedUserIds"] = ids
			return
		}
	}
	db["modules"] = append(arrayField(db, "modules"), map[string]any{"id": "class_grouped", "active": true, "assignedUserIds": ids})
}

func Test593ProjectedRosterCannotEraseHiddenMemberships(t *testing.T) {
	central := map[string]any{"modules": []any{map[string]any{"id": musicProgram553, "assignedUserIds": []any{"hidden", "visible"}}}}
	incoming := map[string]any{"modules": []any{map[string]any{"id": musicProgram553, "assignedUserIds": []any{"visible", "new"}}}}
	merged := legacyMap(musicMerge553(central, incoming))
	ids := arrayField(legacyMap(arrayField(merged, "modules")[0]), "assignedUserIds")
	for _, id := range []string{"hidden", "visible", "new"} {
		if !containsString(ids, id) {
			t.Fatal("lost member", id)
		}
	}
	// Explicit removal remains supported by the delta protocol.
	result := mergeLegacyPatchValue(ids, map[string]any{"__sglSetMembers": map[string]any{"remove": []any{"visible"}, "add": []any{}}}).([]any)
	if containsString(result, "visible") || !containsString(result, "hidden") {
		t.Fatal("explicit delta failed")
	}
}
