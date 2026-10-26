package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadControlStateNormalizesLegacySystemRestoredForOldClients(t *testing.T) {
	dir := t.TempDir()
	s := &Server{path: filepath.Join(dir, "data", "sgl-server-data.json")}
	if err := os.MkdirAll(filepath.Dir(s.controlPath()), 0755); err != nil {
		t.Fatal(err)
	}
	legacy := controlState{Revision: 9, Kind: "system-restored", Message: "Restaurado", RestoreGeneration: 3, RestoreID: "restore-old", ServerVersion: 17}
	b, _ := json.Marshal(legacy)
	if err := os.WriteFile(s.controlPath(), b, 0600); err != nil {
		t.Fatal(err)
	}
	s.loadControlState()
	if s.control.Kind != "force-relogin" || s.control.Revision != 9 || s.control.RestoreGeneration != 3 || s.control.RestoreID != "restore-old" || s.control.ServerVersion != 17 {
		t.Fatalf("legacy restore control was not normalized safely: %#v", s.control)
	}
	persisted, err := os.ReadFile(s.controlPath())
	if err != nil {
		t.Fatal(err)
	}
	var saved controlState
	if err := json.Unmarshal(persisted, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Kind != "force-relogin" || saved.RestoreGeneration != 3 {
		t.Fatalf("normalized control was not persisted: %#v", saved)
	}
}

func TestRestoreGenerationDefaultsToOneAndSnapshotExposesIt(t *testing.T) {
	dir := t.TempDir()
	s := &Server{path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{"users": []any{}}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 4, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}

	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if st.State.RestoreGeneration != 1 {
		t.Fatalf("legacy state without generation must normalize to 1, got %d", st.State.RestoreGeneration)
	}

	snap := legacySnapshot(st.State)
	if got := int64(snap["restoreGeneration"].(int64)); got != 1 {
		t.Fatalf("snapshot restoreGeneration=%d want 1", got)
	}
}

func TestRestoreGenerationPersistsMetadata(t *testing.T) {
	dir := t.TempDir()
	s := &Server{path: filepath.Join(dir, "data", "sgl-server-data.json")}
	original := legacyStore{State: legacyState{
		ServerVersion:     9,
		Database:          map[string]any{"users": []any{}},
		Files:             []any{},
		RestoreGeneration: 3,
		RestoreID:         "restore-abc",
		RestoredAt:        "2026-09-08T12:00:00Z",
	}}
	if err := s.saveLegacy(original); err != nil {
		t.Fatal(err)
	}
	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if st.State.RestoreGeneration != 3 || st.State.RestoreID != "restore-abc" || st.State.RestoredAt == "" {
		b, _ := json.Marshal(st.State)
		t.Fatalf("restore metadata not persisted: %s", b)
	}
}

func TestOptimizedLoginStateCarriesRestoreMetadata(t *testing.T) {
	dir := t.TempDir()
	s := &Server{path: filepath.Join(dir, "data", "sgl-server-data.json")}
	st := legacyStore{State: legacyState{
		ServerVersion:     12,
		Database:          map[string]any{"users": []any{}},
		Files:             []any{map[string]any{"id": "large-file"}},
		RestoreGeneration: 4,
		RestoreID:         "restore-login",
		RestoredAt:        "2026-09-08T13:00:00Z",
	}}
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	got, err := s.loadLegacyLoginState()
	if err != nil {
		t.Fatal(err)
	}
	if got.RestoreGeneration != 4 || got.RestoreID != "restore-login" || got.RestoredAt == "" {
		t.Fatalf("optimized login lost restore metadata: %#v", got)
	}
}

func TestLegacySyncRejectsPreviousRestoreGenerationWithoutMutation(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress":        map[string]any{"asp1": map[string]any{"status": map[string]any{"R1": "pending"}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "revision": 1.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 10, RestoreGeneration: 2, RestoreID: "restore-2", Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	incoming := cloneAny(baseDB).(map[string]any)
	legacyMap(legacyMap(legacyMap(incoming["progress"])["asp1"])["status"])["R1"] = "completed"
	body, _ := json.Marshal(map[string]any{
		"installationId": "phone1", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": 10,
		"restoreGeneration": 1, "operationId": "OLD-OP", "operationType": "auto_sync", "database": incoming, "files": []any{},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("old generation must be rejected with 409, got %d body=%s", rr.Code, rr.Body.String())
	}
	var payload map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &payload)
	if payload["code"] != "RESTORE_REQUIRED" || int64(payload["restoreGeneration"].(float64)) != 2 {
		t.Fatalf("missing restore-required contract: %#v", payload)
	}
	st, _ := s.loadLegacy()
	if st.State.ServerVersion != 10 || len(st.AppliedOperations) != 0 {
		t.Fatalf("rejected old operation mutated state: version=%d applied=%d", st.State.ServerVersion, len(st.AppliedOperations))
	}
	status := legacyMap(legacyMap(legacyMap(legacyMap(st.State.Database)["progress"])["asp1"])["status"])
	if stringValue(status["R1"]) != "pending" {
		t.Fatalf("old operation changed database: %#v", status)
	}
}

func TestLegacySyncRejectsClientWithoutGenerationAfterFirstRestore(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{"users": []any{
		map[string]any{"id": "admin1", "role": "admin", "active": true, "accountState": "active"},
		map[string]any{"id": "asp1", "role": "aspirant", "active": true, "accountState": "active"},
	}, "progress": map[string]any{}, "programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 3, RestoreGeneration: 2, RestoreID: "restore-2", Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"installationId": "old-client", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": 3, "operationId": "NO-GEN", "operationType": "auto_sync", "database": db, "files": []any{}})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("client without generation must be rejected after restore, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestAdminRestoreCreatesNewGenerationAndClearsPreRestoreSyncDebris(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm"}, path: filepath.Join(dir, "data", "sgl-server-data.json"), remoteSessions: map[string]remoteSession{}}
	oldDB := map[string]any{"users": []any{
		map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
		map[string]any{"id": "asp-old", "role": "aspirant", "active": true, "accountState": "active"},
	}, "marker": "old"}
	initial := legacyStore{
		State:             legacyState{ServerVersion: 7, RestoreGeneration: 1, Database: oldDB, Files: []any{map[string]any{"id": "old-file"}}},
		Proposals:         []legacyProposal{{ID: "old-proposal", Status: "pending"}},
		Messages:          []legacyMessage{{ID: "m1", Text: "conservar"}},
		Notifications:     []legacyNotification{{ID: "n1", Message: "conservar"}},
		UserChanges:       []legacyUserChangeRequest{{ID: "u1"}},
		AppliedOperations: []legacyAppliedOperation{{OperationID: "op1", ActorID: "asp-old"}},
	}
	if err := s.saveLegacy(initial); err != nil {
		t.Fatal(err)
	}

	restoredDB := map[string]any{"users": []any{
		map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
		map[string]any{"id": "asp-new", "role": "aspirant", "active": true, "accountState": "active"},
	}, "marker": "restored"}
	body, _ := json.Marshal(map[string]any{"database": restoredDB, "files": []any{map[string]any{"id": "new-file"}}, "message": "Restauracion de prueba"})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/admin/restore", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.adminRestore(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("restore status=%d body=%s", rr.Code, rr.Body.String())
	}

	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if st.State.RestoreGeneration != 2 || st.State.RestoreID == "" || st.State.RestoredAt == "" {
		t.Fatalf("restore metadata missing: %#v", st.State)
	}
	if st.State.ServerVersion != 8 {
		t.Fatalf("server version=%d want 8", st.State.ServerVersion)
	}
	if stringValue(legacyMap(st.State.Database)["marker"]) != "restored" {
		t.Fatalf("database was not replaced: %#v", st.State.Database)
	}
	if len(st.State.Files) != 1 || stringValue(legacyMap(st.State.Files[0])["id"]) != "new-file" {
		t.Fatalf("files were not replaced: %#v", st.State.Files)
	}
	if len(st.Proposals) != 0 || len(st.UserChanges) != 0 || len(st.AppliedOperations) != 0 {
		t.Fatalf("pre-restore sync debris survived: proposals=%d changes=%d applied=%d", len(st.Proposals), len(st.UserChanges), len(st.AppliedOperations))
	}
	if len(st.Messages) != 1 || len(st.Notifications) != 1 {
		t.Fatalf("historical messages/notifications should remain in 5.0.28")
	}

	control := s.currentControlState()
	if control.Kind != "force-relogin" || control.RestoreGeneration != 2 || control.RestoreID != st.State.RestoreID || control.ServerVersion != 8 {
		t.Fatalf("bad restore control: %#v", control)
	}
}

func TestFilteredRemoteSnapshotCarriesCanonicalRestoreMetadataInEnvelopeAndDatabase(t *testing.T) {
	st := legacyState{
		ServerVersion:     21,
		RestoreGeneration: 4,
		RestoreID:         "restore-current",
		RestoredAt:        "2026-09-08T15:00:00Z",
		Database: map[string]any{
			"restoreGeneration": 1.0,
			"restoreId":         "restore-stale",
			"restorePending":    true,
			"users": []any{
				map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
				map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"},
			},
			"progress": map[string]any{"asp1": map[string]any{}},
			"settings": map[string]any{},
		},
		Files: []any{},
	}
	s := &Server{}
	snap := s.filteredLegacySnapshotOwnedMode(st, "asp1", "aspirant", false)
	if got := int64(snap["restoreGeneration"].(int64)); got != 4 {
		t.Fatalf("filtered envelope restoreGeneration=%d want 4", got)
	}
	if snap["restoreId"] != "restore-current" || snap["restoredAt"] != st.RestoredAt {
		t.Fatalf("filtered envelope lost restore identity: %#v", snap)
	}
	db := legacyMap(snap["database"])
	if got := int64(db["restoreGeneration"].(int64)); got != 4 {
		t.Fatalf("filtered database restoreGeneration=%d want 4; this would recreate the RESTORE_REQUIRED loop", got)
	}
	if db["restoreId"] != "restore-current" || db["restoredAt"] != st.RestoredAt {
		t.Fatalf("filtered database lost canonical restore identity: %#v", db)
	}
	if pending, ok := db["restorePending"].(bool); !ok || pending {
		t.Fatalf("server snapshot must clear device-local restorePending, got %#v", db["restorePending"])
	}
}

func TestLoadLegacyRepairsStaleEmbeddedRestoreMetadata(t *testing.T) {
	dir := t.TempDir()
	s := &Server{path: filepath.Join(dir, "data", "sgl-server-data.json")}
	staleDB := map[string]any{
		"restoreGeneration": 1,
		"restoreId":         "old-id",
		"restorePending":    true,
		"users":             []any{},
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{
		ServerVersion: 33, RestoreGeneration: 5, RestoreID: "restore-5", RestoredAt: "2026-09-08T16:00:00Z", Database: staleDB, Files: []any{},
	}}); err != nil {
		t.Fatal(err)
	}
	loaded, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	db := legacyMap(loaded.State.Database)
	if got := int64(db["restoreGeneration"].(int64)); got != 5 {
		t.Fatalf("load did not repair embedded restore generation: %d", got)
	}
	if db["restoreId"] != "restore-5" || db["restoredAt"] != "2026-09-08T16:00:00Z" {
		t.Fatalf("load did not repair embedded restore identity: %#v", db)
	}
	if pending, ok := db["restorePending"].(bool); !ok || pending {
		t.Fatalf("load must clear stale restorePending, got %#v", db["restorePending"])
	}
}

func TestAutomaticAckReassertsCanonicalRestoreGeneration(t *testing.T) {
	st := legacyState{ServerVersion: 14, RestoreGeneration: 3, RestoreID: "restore-3", RestoredAt: "2026-09-08T14:30:00Z"}
	ack := legacyAutomaticAck(st, legacyAppliedOperation{OperationID: "SGL-AVA-1", OperationType: "save_progress"}, false)
	if got := int64(ack["restoreGeneration"].(int64)); got != 3 {
		t.Fatalf("ack restoreGeneration=%d want 3", got)
	}
	if ack["restoreId"] != "restore-3" || ack["restoredAt"] != st.RestoredAt {
		t.Fatalf("ack lost canonical restore metadata: %#v", ack)
	}
}

func TestFilteredBootstrapGenerationCanSyncWithoutFalseRestoreRequired(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"restoreGeneration": 1,
		"restorePending":    true,
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress":        map[string]any{"asp1": map[string]any{"status": map[string]any{"I-10": "pending"}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "classes": []any{}, "audit": []any{}, "revision": float64(1),
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{
		ServerVersion: 40, RestoreGeneration: 4, RestoreID: "restore-4", RestoredAt: "2026-09-08T17:00:00Z", Database: baseDB, Files: []any{},
	}}); err != nil {
		t.Fatal(err)
	}
	loaded, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	bootstrap := s.filteredLegacySnapshotOwnedMode(loaded.State, "asp1", "aspirant", false)
	clientDB := legacyMap(bootstrap["database"])
	generation, ok := clientDB["restoreGeneration"].(int64)
	if !ok || generation != 4 {
		t.Fatalf("bootstrap did not give the client generation 4: %#v", clientDB["restoreGeneration"])
	}
	body, _ := json.Marshal(map[string]any{
		"installationId": "phone-after-restore", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": 40,
		"restoreGeneration": generation, "operationId": "SGL-5030-AFTER-RESTORE", "operationType": "save_progress", "requirementId": "I-10", "programId": "class_grouped", "ackOnly": true,
		"patch": map[string]any{"progress": map[string]any{"asp1": map[string]any{
			"status": map[string]any{"I-10": "in_progress"}, "submissions": map[string]any{"I-10": map[string]any{"text": "guardado normal despues de restaurar", "updatedAt": "2026-09-08T17:01:00Z"}},
		}}}, "files": []any{},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("valid post-restore sync was rejected: status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["code"] == "RESTORE_REQUIRED" {
		t.Fatalf("false restore loop reproduced: %#v", payload)
	}
	if got := int64(payload["restoreGeneration"].(float64)); got != 4 {
		t.Fatalf("sync ack did not preserve generation 4: %d", got)
	}
	final, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if final.State.RestoreGeneration != 4 || int64(legacyMap(final.State.Database)["restoreGeneration"].(int64)) != 4 {
		t.Fatalf("server restore generation drifted after normal sync: state=%d db=%#v", final.State.RestoreGeneration, legacyMap(final.State.Database)["restoreGeneration"])
	}
}
