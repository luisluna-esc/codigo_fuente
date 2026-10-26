package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func progressResetTestServer(t *testing.T) *Server {
	t.Helper()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(t.TempDir(), "data", "sgl-server-data.json")}
	db := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "t1", "fullName": "Tutor Uno", "role": "tutor", "active": true, "accountState": "active"},
			map[string]any{"id": "t2", "fullName": "Tutor Dos", "role": "tutor", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "role": "aspirant", "tutorId": "t1", "active": true, "accountState": "active"},
		},
		"progress": map[string]any{"asp1": map[string]any{
			"status": map[string]any{"R1": "in_progress", "R2": "approved"},
			"submissions": map[string]any{
				"R1": map[string]any{"text": "texto", "richText": "<p>texto</p>", "attachmentIds": []any{"f1"}},
				"R2": map[string]any{"text": "otro", "richText": "<p>otro</p>", "attachmentIds": []any{"f2"}},
			},
			"readings": map[string]any{"R1": map[string]any{"done": true}, "R2": map[string]any{"done": true}},
		}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"classes": []any{}, "tasks": []any{}, "materials": []any{}, "modules": []any{}, "audit": []any{}, "revision": 1.0,
	}
	files := []any{
		map[string]any{"id": "f1", "ownerId": "asp1", "programId": "class_grouped", "requirementId": "R1", "kind": "photo"},
		map[string]any{"id": "f2", "ownerId": "asp1", "programId": "class_grouped", "requirementId": "R2", "kind": "photo"},
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 2, Database: db, Files: files}}); err != nil {
		t.Fatal(err)
	}
	return s
}

func callProgressReset(t *testing.T, s *Server, token string, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "http://example/api/progress/reset", bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	s.legacyProgressReset(rr, req)
	return rr
}

func TestAspirantCanDeleteOwnRequirementProgress(t *testing.T) {
	s := progressResetTestServer(t)
	rr := callProgressReset(t, s, "cli", map[string]any{"actorId": "asp1", "ownerId": "asp1", "programId": "class_grouped", "requirementId": "R1", "scope": "requirement", "operationId": "reset-aspirant"})
	if rr.Code != http.StatusOK {
		t.Fatalf("aspirant reset=%d %s", rr.Code, rr.Body.String())
	}
	st, _ := s.loadLegacy()
	p := legacyProgressForRequirement(st.State.Database, "asp1", "class_grouped", false)
	if got := stringValue(mapField(p, "status")["R1"]); got != "pending" {
		t.Fatalf("R1 status=%q", got)
	}
	if _, ok := mapField(p, "submissions")["R1"]; ok {
		t.Fatal("R1 submission survived aspirant reset")
	}
}

func TestAspirantCannotDeleteSelection(t *testing.T) {
	s := progressResetTestServer(t)
	rr := callProgressReset(t, s, "cli", map[string]any{"actorId": "asp1", "ownerId": "asp1", "programId": "class_grouped", "requirementIds": []string{"R1", "R2"}, "scope": "selection", "operationId": "reset-aspirant-selection"})
	if rr.Code != http.StatusForbidden {
		t.Fatalf("aspirant selection reset=%d %s", rr.Code, rr.Body.String())
	}
}

func TestTutorCanDeleteSelectedAspirantProgressOnly(t *testing.T) {
	s := progressResetTestServer(t)
	rr := callProgressReset(t, s, "cli", map[string]any{"actorId": "t1", "ownerId": "asp1", "programId": "class_grouped", "requirementIds": []string{"R1"}, "scope": "selection", "operationId": "reset-selection"})
	if rr.Code != http.StatusOK {
		t.Fatalf("selection reset=%d %s", rr.Code, rr.Body.String())
	}
	st, _ := s.loadLegacy()
	p := legacyProgressForRequirement(st.State.Database, "asp1", "class_grouped", false)
	if got := stringValue(mapField(p, "status")["R1"]); got != "pending" {
		t.Fatalf("R1 status=%q", got)
	}
	if got := stringValue(mapField(p, "status")["R2"]); got != "approved" {
		t.Fatalf("R2 status changed=%q", got)
	}
	if _, ok := mapField(p, "submissions")["R1"]; ok {
		t.Fatal("R1 submission survived")
	}
	if _, ok := mapField(p, "submissions")["R2"]; !ok {
		t.Fatal("R2 submission was deleted")
	}
	if len(st.State.Files) != 1 || stringValue(legacyMap(st.State.Files[0])["id"]) != "f2" {
		t.Fatalf("files=%#v", st.State.Files)
	}
	if len(arrayField(legacyMap(st.State.Database), "audit")) == 0 {
		t.Fatal("missing audit")
	}
}

func TestTutorCannotDeleteAnotherTutorsAspirantProgress(t *testing.T) {
	s := progressResetTestServer(t)
	denied := callProgressReset(t, s, "cli", map[string]any{"actorId": "t2", "ownerId": "asp1", "programId": "class_grouped", "requirementIds": []string{"R1"}, "scope": "selection", "operationId": "reset-denied"})
	if denied.Code != http.StatusForbidden {
		t.Fatalf("other tutor=%d %s", denied.Code, denied.Body.String())
	}
}

func TestAdminCanDeleteMultipleSelectedRequirements(t *testing.T) {
	s := progressResetTestServer(t)
	rr := callProgressReset(t, s, "adm", map[string]any{"actorId": "admin1", "ownerId": "asp1", "programId": "class_grouped", "requirementIds": []string{"R1", "R2", "R1"}, "scope": "selection", "operationId": "reset-multi"})
	if rr.Code != http.StatusOK {
		t.Fatalf("reset=%d %s", rr.Code, rr.Body.String())
	}
	st, _ := s.loadLegacy()
	p := legacyProgressForRequirement(st.State.Database, "asp1", "class_grouped", false)
	if len(mapField(p, "submissions")) != 0 || len(mapField(p, "readings")) != 0 {
		t.Fatalf("selected progress not cleared: %#v", p)
	}
	if stringValue(mapField(p, "status")["R1"]) != "pending" || stringValue(mapField(p, "status")["R2"]) != "pending" {
		t.Fatalf("statuses=%#v", mapField(p, "status"))
	}
	if len(st.State.Files) != 0 {
		t.Fatalf("files survived=%#v", st.State.Files)
	}
}

func TestWholeProgramResetIsDisabled(t *testing.T) {
	s := progressResetTestServer(t)
	rr := callProgressReset(t, s, "adm", map[string]any{"actorId": "admin1", "ownerId": "asp1", "programId": "class_grouped", "scope": "program", "operationId": "reset-program"})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("program reset=%d %s", rr.Code, rr.Body.String())
	}
	st, _ := s.loadLegacy()
	p := legacyProgressForRequirement(st.State.Database, "asp1", "class_grouped", false)
	if stringValue(mapField(p, "status")["R1"]) != "in_progress" || len(mapField(p, "submissions")) != 2 {
		t.Fatalf("program reset changed progress: %#v", p)
	}
}

func TestProgressResetSelectionIsIdempotent(t *testing.T) {
	s := progressResetTestServer(t)
	body := map[string]any{"actorId": "admin1", "ownerId": "asp1", "programId": "class_grouped", "requirementIds": []string{"R1"}, "scope": "selection", "operationId": "same-op"}
	first := callProgressReset(t, s, "adm", body)
	second := callProgressReset(t, s, "adm", body)
	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("codes %d/%d", first.Code, second.Code)
	}
	var payload map[string]any
	_ = json.Unmarshal(second.Body.Bytes(), &payload)
	if payload["deduplicated"] != true {
		t.Fatalf("second not dedup: %#v", payload)
	}
}

func TestProgressResetWorksForGuideMajorAndGuideMasterSpecialDeliveries(t *testing.T) {
	s := progressResetTestServer(t)
	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	db := legacyMap(st.State.Database)
	db["programProgress"] = map[string]any{
		"asp1": map[string]any{
			"guide_major": map[string]any{
				"status": map[string]any{"GM1": "in_review"},
				"submissions": map[string]any{
					"GM1": map[string]any{"submissionMode": "document", "primaryDocumentId": "fgm", "attachmentIds": []any{"fgm"}, "completionDate": "2026-09-14"},
				},
			},
			"guide_master": map[string]any{
				"status": map[string]any{"GMM1": "approved"},
				"submissions": map[string]any{
					"GMM1": map[string]any{"submissionMode": "reading_report", "readingReport": map[string]any{"bookTitle": "Libro", "finalReflection": "Reflexion"}, "completionDate": "2026-09-14"},
				},
			},
		},
	}
	st.State.Files = append(st.State.Files, map[string]any{"id": "fgm", "ownerId": "asp1", "programId": "guide_major", "requirementId": "GM1", "kind": "requirement_document"})
	st.State.Database = db
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}

	// Tutor asignado elimina una entrega PDF/fotos de Guia Mayor.
	rr := callProgressReset(t, s, "cli", map[string]any{"actorId": "t1", "ownerId": "asp1", "programId": "guide_major", "requirementId": "GM1", "scope": "requirement", "operationId": "reset-guide-major-doc"})
	if rr.Code != http.StatusOK {
		t.Fatalf("guide major reset=%d %s", rr.Code, rr.Body.String())
	}
	st, _ = s.loadLegacy()
	p := legacyProgressForRequirement(st.State.Database, "asp1", "guide_major", false)
	if got := stringValue(mapField(p, "status")["GM1"]); got != "pending" {
		t.Fatalf("GM1 status=%q", got)
	}
	if _, ok := mapField(p, "submissions")["GM1"]; ok {
		t.Fatal("GM1 document submission survived reset")
	}

	// 5.39: el autor no elimina un informe aprobado; el tutor asignado sí.
	rr = callProgressReset(t, s, "cli", map[string]any{"actorId": "asp1", "ownerId": "asp1", "programId": "guide_master", "requirementId": "GMM1", "scope": "requirement", "operationId": "reset-guide-master-reading"})
	if rr.Code != http.StatusForbidden {
		t.Fatalf("approved author reset=%d", rr.Code)
	}
	rr = callProgressReset(t, s, "cli", map[string]any{"actorId": "t1", "ownerId": "asp1", "programId": "guide_master", "requirementId": "GMM1", "scope": "requirement", "operationId": "reset-guide-master-reading-tutor"})
	if rr.Code != http.StatusOK {
		t.Fatalf("guide master reset=%d %s", rr.Code, rr.Body.String())
	}
	st, _ = s.loadLegacy()
	p = legacyProgressForRequirement(st.State.Database, "asp1", "guide_master", false)
	if got := stringValue(mapField(p, "status")["GMM1"]); got != "pending" {
		t.Fatalf("GMM1 status=%q", got)
	}
	if _, ok := mapField(p, "submissions")["GMM1"]; ok {
		t.Fatal("GMM1 reading report survived reset")
	}
}
