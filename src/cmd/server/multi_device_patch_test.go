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

func TestCompactPatchPreservesNewerServerRequirementFromOtherDevice(t *testing.T) {
	serverDB := map[string]any{
		"users": []any{map[string]any{"id": "asp1", "fullName": "Lider Uno", "role": "aspirant", "active": true, "accountState": "active"}},
		"progress": map[string]any{"asp1": map[string]any{
			"status": map[string]any{"I-09": "approved", "I-10": "pending"},
			"submissions": map[string]any{
				"I-09": map[string]any{"text": "aprobado en telefono", "updatedAt": "2026-08-31T22:00:00Z", "approval": map[string]any{"reviewerId": "admin"}},
				"I-10": map[string]any{"text": "", "updatedAt": "2026-08-31T20:00:00Z"},
			},
		}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "classes": []any{}, "audit": []any{}, "revision": float64(12),
	}
	// A stale laptop sends ONLY I-10. The server must not lose I-09.
	patch := map[string]any{"progress": map[string]any{"asp1": map[string]any{
		"status":      map[string]any{"I-10": "in_progress"},
		"submissions": map[string]any{"I-10": map[string]any{"text": "avance laptop", "updatedAt": "2026-08-31T22:10:00Z"}},
	}}}
	expanded := expandLegacyPatch(serverDB, patch)
	mergedAny, _, err := mergeLegacyProposalOnServer(serverDB, &legacyProposal{ActorID: "asp1", ActorRole: "aspirant", Database: expanded})
	if err != nil {
		t.Fatal(err)
	}
	merged := legacyMap(mergedAny)
	progress := legacyMap(legacyMap(merged["progress"])["asp1"])
	status := legacyMap(progress["status"])
	if got := stringValue(status["I-09"]); got != "approved" {
		t.Fatalf("I-09 was overwritten: %q", got)
	}
	if got := stringValue(status["I-10"]); got != "in_progress" {
		t.Fatalf("I-10 not applied: %q", got)
	}
}

func TestCompactClassPatchPreservesOtherCheckboxes(t *testing.T) {
	serverDB := map[string]any{
		"users":           []any{map[string]any{"id": "asp1", "fullName": "Lider Uno", "role": "aspirant", "active": true, "accountState": "active"}},
		"progress":        map[string]any{"asp1": map[string]any{"status": map[string]any{}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "audit": []any{}, "revision": float64(4),
		"classes": []any{map[string]any{
			"id": "class1", "assignedUserIds": []any{"asp1"},
			"work": map[string]any{"tracking": map[string]any{"member1:req1": "completed", "member1:req2": "pending"}, "members": []any{}},
		}},
	}
	patch := map[string]any{"classes": []any{map[string]any{"id": "class1", "work": map[string]any{"tracking": map[string]any{"member1:req2": "completed"}}}}}
	expanded := expandLegacyPatch(serverDB, patch)
	mergedAny, _, err := mergeLegacyProposalOnServer(serverDB, &legacyProposal{ActorID: "asp1", ActorRole: "aspirant", Database: expanded})
	if err != nil {
		t.Fatal(err)
	}
	merged := legacyMap(mergedAny)
	class := legacyMap(arrayField(merged, "classes")[0])
	tracking := legacyMap(mapField(class, "work")["tracking"])
	if got := stringValue(tracking["member1:req1"]); got != "completed" {
		t.Fatalf("existing checkbox lost: %q", got)
	}
	if got := stringValue(tracking["member1:req2"]); got != "completed" {
		t.Fatalf("new checkbox not applied: %q", got)
	}
}

func TestLegacySyncAcceptsCompactPatchWithoutFullDatabase(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users":           []any{map[string]any{"id": "asp1", "fullName": "Lider Uno", "role": "aspirant", "active": true, "accountState": "active"}},
		"progress":        map[string]any{"asp1": map[string]any{"status": map[string]any{"I-10": "pending"}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "classes": []any{}, "audit": []any{}, "revision": float64(1),
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 3, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"installationId": "phone1", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": 1,
		"operationId": "SGL-AVA-ASP1-I10-1", "operationType": "save_progress", "requirementId": "I-10", "programId": "class_grouped",
		"patch": map[string]any{"progress": map[string]any{"asp1": map[string]any{
			"status": map[string]any{"I-10": "in_progress"}, "submissions": map[string]any{"I-10": map[string]any{"text": "avance compacto", "updatedAt": "2026-08-31T22:10:00Z"}},
		}}}, "files": []any{},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != 200 {
		t.Fatalf("compact patch status=%d body=%s", rr.Code, rr.Body.String())
	}
	st, _ := s.loadLegacy()
	progress := legacyMap(legacyMap(legacyMap(st.State.Database)["progress"])["asp1"])
	if got := stringValue(legacyMap(progress["status"])["I-10"]); got != "in_progress" {
		t.Fatalf("compact patch did not apply: %q", got)
	}
	if st.State.ServerVersion != 4 {
		t.Fatalf("server version not incremented: %d", st.State.ServerVersion)
	}
}

func TestStaleDeviceCannotRevertApprovedRequirement(t *testing.T) {
	serverDB := map[string]any{
		"users": []any{map[string]any{"id": "asp1", "fullName": "Lider Uno", "role": "aspirant", "active": true, "accountState": "active"}},
		"progress": map[string]any{"asp1": map[string]any{
			"status":      map[string]any{"I-10": "approved"},
			"submissions": map[string]any{"I-10": map[string]any{"text": "aprobado", "updatedAt": "2026-08-31T22:00:00Z", "approval": map[string]any{"reviewerId": "admin"}}},
		}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "classes": []any{}, "audit": []any{}, "revision": float64(2),
	}
	patch := map[string]any{"progress": map[string]any{"asp1": map[string]any{
		"status":      map[string]any{"I-10": "in_progress"},
		"submissions": map[string]any{"I-10": map[string]any{"text": "copia vieja laptop", "updatedAt": "2026-08-31T21:00:00Z"}},
	}}}
	expanded := expandLegacyPatch(serverDB, patch)
	mergedAny, _, err := mergeLegacyProposalOnServer(serverDB, &legacyProposal{ActorID: "asp1", ActorRole: "aspirant", Database: expanded})
	if err != nil {
		t.Fatal(err)
	}
	progress := legacyMap(legacyMap(legacyMap(mergedAny)["progress"])["asp1"])
	if got := stringValue(legacyMap(progress["status"])["I-10"]); got != "approved" {
		t.Fatalf("approved requirement was reverted: %q", got)
	}
	if got := stringValue(legacyMap(legacyMap(progress["submissions"])["I-10"])["text"]); got != "aprobado" {
		t.Fatalf("approved submission was replaced: %q", got)
	}
}

func TestThreeRapidClassPlanningChecksAllRemainApplied(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users":           []any{map[string]any{"id": "asp1", "fullName": "Lider Uno", "role": "aspirant", "active": true, "accountState": "active"}},
		"progress":        map[string]any{"asp1": map[string]any{"status": map[string]any{}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "audit": []any{}, "revision": float64(1),
		"classes": []any{map[string]any{
			"id": "class1", "assignedUserIds": []any{"asp1"},
			"work": map[string]any{"planning": map[string]any{
				"R1": map[string]any{"progressed": false},
				"R2": map[string]any{"progressed": false},
				"R3": map[string]any{"progressed": false},
			}, "tracking": map[string]any{}, "members": []any{}, "attendance": []any{}},
		}},
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 10, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	for i, reqID := range []string{"R1", "R2", "R3"} {
		body, _ := json.Marshal(map[string]any{
			"installationId": "phone1", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": 10 + i,
			"operationId": "SGL-AUT-PLAN-" + reqID, "operationType": "auto_sync",
			"patch": map[string]any{"classes": []any{map[string]any{"id": "class1", "work": map[string]any{"planning": map[string]any{reqID: map[string]any{"progressed": true}}}}}},
			"files": []any{},
		})
		req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer cli")
		rr := httptest.NewRecorder()
		s.legacySync(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("check %s status=%d body=%s", reqID, rr.Code, rr.Body.String())
		}
	}
	st, _ := s.loadLegacy()
	classes := arrayField(legacyMap(st.State.Database), "classes")
	if len(classes) != 1 {
		t.Fatalf("expected one class, got %d", len(classes))
	}
	planning := mapField(legacyMap(classes[0]), "work")
	planMap := legacyMap(planning["planning"])
	for _, reqID := range []string{"R1", "R2", "R3"} {
		row := legacyMap(planMap[reqID])
		if got, _ := row["progressed"].(bool); !got {
			t.Fatalf("rapid check %s was lost: %#v", reqID, row)
		}
	}
}

func TestAutomaticSyncAckOnlyDoesNotReturnWholeDatabase(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users":           []any{map[string]any{"id": "asp1", "fullName": "Lider Uno", "role": "aspirant", "active": true, "accountState": "active"}},
		"progress":        map[string]any{"asp1": map[string]any{"status": map[string]any{"I-10": "pending"}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "classes": []any{}, "audit": []any{}, "revision": float64(1),
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 20, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"installationId": "phone1", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": 20,
		"operationId": "SGL-AVA-ACK-I10", "operationType": "save_progress", "requirementId": "I-10", "programId": "class_grouped", "ackOnly": true,
		"patch": map[string]any{"progress": map[string]any{"asp1": map[string]any{
			"status": map[string]any{"I-10": "in_progress"}, "submissions": map[string]any{"I-10": map[string]any{"text": "avance", "updatedAt": "2026-08-31T22:30:00Z"}},
		}}}, "files": []any{},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("ack status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["ackOnly"] != true || payload["status"] != "applied" {
		t.Fatalf("unexpected ack: %#v", payload)
	}
	if _, exists := payload["database"]; exists {
		t.Fatalf("ack returned whole database: %#v", payload)
	}
	if _, exists := payload["files"]; exists {
		t.Fatalf("ack returned files: %#v", payload)
	}
	if got := int64(payload["serverVersion"].(float64)); got != 21 {
		t.Fatalf("serverVersion=%d", got)
	}
}

func TestNormalUserSyncWithoutOperationTypeAppliesDirectly(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Administrador", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress":        map[string]any{"asp1": map[string]any{"status": map[string]any{"C1": "pending"}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "revision": 1.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 1, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	patch := map[string]any{"progress": map[string]any{"asp1": map[string]any{"status": map[string]any{"C1": "in_progress"}, "submissions": map[string]any{"C1": map[string]any{"text": "dato normal", "attachmentIds": []any{}}}}}}
	send := func() map[string]any {
		body, _ := json.Marshal(map[string]any{
			"installationId": "phone1", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": 1,
			"ackOnly": true, "patch": patch, "files": []any{},
		})
		req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer cli")
		rr := httptest.NewRecorder()
		s.legacySync(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("normal sync status=%d body=%s", rr.Code, rr.Body.String())
		}
		var payload map[string]any
		_ = json.Unmarshal(rr.Body.Bytes(), &payload)
		return payload
	}
	first := send()
	if first["status"] != "applied" || first["channel"] != "sync" || first["clearLocalOperation"] != true {
		t.Fatalf("normal data must be acknowledged as direct sync: %#v", first)
	}
	if strings.TrimSpace(stringValue(first["operationId"])) == "" {
		t.Fatalf("server must return a generated operationId for legacy deltas: %#v", first)
	}
	second := send()
	if second["deduplicated"] != true {
		t.Fatalf("retry was not deduplicated: %#v", second)
	}
	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Proposals) != 0 {
		t.Fatalf("normal data must not create review/proposal rows: %#v", st.Proposals)
	}
	if st.State.ServerVersion != 2 {
		t.Fatalf("retry must not increment server version twice, got %d", st.State.ServerVersion)
	}
	progress := legacyMap(legacyMap(legacyMap(st.State.Database)["progress"])["asp1"])
	if got := stringValue(legacyMap(progress["status"])["C1"]); got != "in_progress" {
		t.Fatalf("normal status not applied: %q", got)
	}
	if got := stringValue(legacyMap(legacyMap(progress["submissions"])["C1"])["text"]); got != "dato normal" {
		t.Fatalf("normal submission not applied: %q", got)
	}
}
