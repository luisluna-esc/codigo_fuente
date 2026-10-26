package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Synthetic inputs exercise integration with 6.0 without changing the embedded
// production manifest or requiring private production backups.
func Test60ReconcilePreservesMusicOptionsAndTutors(t *testing.T) {
	original := append([]byte(nil), reconciliationManifest598...)
	t.Cleanup(func() { reconciliationManifest598 = original })
	s, _ := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	beforePassword := map[string]any{"test": "before"}
	afterPassword := map[string]any{"test": "after"}
	n := map[string]any{"id": noemi598, "role": "aspirant", "active": true, "password": beforePassword, "tutorByProgram": map[string]any{musicProgram553: "regional", "guide_major": "tm"}}
	f := map[string]any{"id": franklin598, "role": "aspirant", "active": true, "tutorByProgram": map[string]any{musicProgram553: "tm"}}
	db["users"] = append(legacyUsers(db), n, f)
	opts := map[string]any{"driveVisible": true, "whatsappUsers": []any{"regional"}}
	musicData550(db)["interface598"] = cloneAny(opts)
	st.State.RestoreGeneration = 3
	st.State.RestoreID = "synthetic-60"
	db["restoreGeneration"] = 3
	db["restoreId"] = "synthetic-60"
	patches := []map[string]any{{"users": []any{map[string]any{"id": noemi598, "password": afterPassword}}}, {"test": "enroll-one"}, {"test": "enroll-two"}}
	manifest := reconcileManifest598{ActorID: "admin", SourceGeneration: 2, Generation: 3, RestoreID: "synthetic-60", PasswordBefore: hash598(beforePassword), Patches: map[string]string{}}
	ops := []any{}
	for i, p := range patches {
		id := []string{"s1", "s2", "s3"}[i]
		manifest.Patches[id] = hash598(p)
		ops = append(ops, map[string]any{"id": id, "actorId": "admin", "restoreGeneration": 2, "action": "admin_change", "programId": musicProgram553, "patch": p})
	}
	reconciliationManifest598, _ = json.Marshal(manifest)
	beforeFiles := cloneAny(st.State.Files)
	tutors := cloneAny(n["tutorByProgram"])
	receipt, reason := reconcileState598(&st, "admin", ops, "2026-10-12T23:00:00Z")
	if reason != "" || receipt == nil {
		t.Fatal(reason)
	}
	current := legacyMap(st.State.Database)
	if !reflect.DeepEqual(mapField(musicData550(current), "interface598"), opts) || !reflect.DeepEqual(legacyUser(current, noemi598)["tutorByProgram"], tutors) || !reflect.DeepEqual(st.State.Files, beforeFiles) {
		t.Fatal("configuration, tutors or files changed")
	}
	if !musicAssigned550(current, noemi598) || !musicAssigned550(current, franklin598) {
		t.Fatal("enrollment missing")
	}
	version := st.State.ServerVersion
	newPassword := map[string]any{"test": "subsequent-change"}
	legacyUser(current, noemi598)["password"] = newPassword
	again, reason := reconcileState598(&st, "admin", ops, "2026-10-12T23:01:00Z")
	if reason != "" || hash598(receipt) != hash598(again) || st.State.ServerVersion != version || !reflect.DeepEqual(legacyUser(st.State.Database, noemi598)["password"], newPassword) {
		t.Fatal("retry overwrote subsequent work", reason)
	}
}
func Test60ReconcileRejectsWrongGenerationWithoutChanges(t *testing.T) {
	s, _ := excelFixture560(t)
	st, _ := s.loadLegacy()
	legacyUser(st.State.Database, "admin")["id"] = manifest598().ActorID
	before, _ := json.Marshal(st)
	if _, reason := reconcileState598(&st, manifest598().ActorID, nil, "now"); reason == "" {
		t.Fatal("unrelated generation accepted")
	}
	after, _ := json.Marshal(st)
	if string(before) != string(after) {
		t.Fatal("rejected repair modified state")
	}
}
