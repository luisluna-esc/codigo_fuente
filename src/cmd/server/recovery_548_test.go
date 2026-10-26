package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func recoveryFixture548(t *testing.T) (*Server, recoveryCapsule548) {
	t.Helper()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(t.TempDir(), "data", "sgl-server-data.json")}
	db := map[string]any{"users": []any{map[string]any{"id": "admin", "role": "admin", "active": true}, map[string]any{"id": "a", "fullName": "Ana", "role": "aspirant", "active": true}}, "progress": map[string]any{"a": map[string]any{"submissions": map[string]any{"r": map[string]any{"text": "servidor"}}, "status": map[string]any{"r": "in_progress"}}}, "programProgress": map[string]any{}, "bibleYear": map[string]any{}}
	if err := s.saveLegacy(legacyStore{State: legacyState{Database: db, Files: []any{}, ServerVersion: 10, RestoreGeneration: 2, RestoreID: "restored"}}); err != nil {
		t.Fatal(err)
	}
	return s, recoveryCapsule548{Format: "SGLRecovery548", OwnerID: "a", InstallationID: "device", Secret: strings.Repeat("a", 64), Fingerprint: strings.Repeat("b", 64), Progress: map[string]any{"submissions": map[string]any{"r": map[string]any{"text": "aspirante"}}, "status": map[string]any{"r": "in_progress"}}, Programs: map[string]any{}, Files: []any{}}
}
func recoveryCall548(handler http.HandlerFunc, token string, body any) *httptest.ResponseRecorder {
	data, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "http://example/api/admin/recovery", bytes.NewReader(data))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler(rr, req)
	return rr
}
func TestRecovery548AdminOnlyAndPreviewDoesNotMutate(t *testing.T) {
	s, c := recoveryFixture548(t)
	if rr := recoveryCall548(s.adminRecovery548, "cli", recoveryRequest548{Capsule: c}); rr.Code != 403 {
		t.Fatalf("non-admin: %d", rr.Code)
	}
	if rr := recoveryCall548(s.adminRecovery548, "adm", recoveryRequest548{Capsule: c}); rr.Code != 200 {
		t.Fatalf("preview: %d %s", rr.Code, rr.Body.String())
	}
	st, _ := s.loadLegacy()
	if st.State.ServerVersion != 10 || len(st.Recoveries548) != 0 {
		t.Fatal("preview mutated state")
	}
}
func TestRecovery548CommitAndSecretBoundReceipt(t *testing.T) {
	s, c := recoveryFixture548(t)
	req := recoveryRequest548{Capsule: c, Commit: true, BaseVersion: 10, Choices: map[string]string{"class_grouped|r": "local"}}
	rr := recoveryCall548(s.adminRecovery548, "adm", req)
	if rr.Code != 200 {
		t.Fatalf("commit: %d %s", rr.Code, rr.Body.String())
	}
	st, _ := s.loadLegacy()
	if st.State.ServerVersion != 11 || st.State.RestoreGeneration != 2 || len(st.Recoveries548) != 1 {
		t.Fatal("invalid committed state")
	}
	sub := legacyMap(mapField(legacyProgressForRequirement(legacyMap(st.State.Database), "a", "class_grouped", false), "submissions")["r"])
	if stringValue(sub["text"]) != "aspirante" {
		t.Fatal("report not recovered")
	}
	status := map[string]any{"ownerId": "a", "installationId": "device", "secret": c.Secret}
	rr = recoveryCall548(s.recoveryStatus548, "cli", status)
	if !strings.Contains(rr.Body.String(), `"ready":true`) {
		t.Fatal(rr.Body.String())
	}
	status["secret"] = strings.Repeat("c", 64)
	rr = recoveryCall548(s.recoveryStatus548, "cli", status)
	if strings.Contains(rr.Body.String(), `"ready":true`) {
		t.Fatal("receipt leaked to wrong secret")
	}
	// A later server restore invalidates the old permission.
	st.State.RestoreGeneration++
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	status["secret"] = c.Secret
	rr = recoveryCall548(s.recoveryStatus548, "cli", status)
	if strings.Contains(rr.Body.String(), `"ready":true`) {
		t.Fatal("stale receipt survived restore")
	}
}
func TestRecovery548ConcurrentServerEditRejectsWithoutMutation(t *testing.T) {
	s, c := recoveryFixture548(t)
	rr := recoveryCall548(s.adminRecovery548, "adm", recoveryRequest548{Capsule: c, Commit: true, BaseVersion: 9, Choices: map[string]string{"class_grouped|r": "local"}})
	if rr.Code != 409 {
		t.Fatalf("expected conflict, got %d", rr.Code)
	}
	st, _ := s.loadLegacy()
	sub := legacyMap(mapField(legacyProgressForRequirement(legacyMap(st.State.Database), "a", "class_grouped", false), "submissions")["r"])
	if stringValue(sub["text"]) != "servidor" || len(st.Recoveries548) != 0 {
		t.Fatal("failed commit modified cache")
	}
}
