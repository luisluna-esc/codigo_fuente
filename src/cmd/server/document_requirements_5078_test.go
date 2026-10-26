package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestReviewDocumentSubmission5078(t *testing.T) {
	db := map[string]any{
		"reportRequirementModes": map[string]any{
			"class_grouped:req-1": map[string]any{"mode": "document"},
		},
	}
	if !reviewRequirementDocumentMode(db, "class_grouped", "req-1") {
		t.Fatal("explicit document mode must be recognized")
	}
	if reviewRequirementDocumentMode(db, "class_grouped", "req-normal") {
		t.Fatal("normal requirements must not become document mode implicitly")
	}

	submission := map[string]any{
		"submissionMode":    "document",
		"documentPageCount": 5,
		"primaryDocumentId": "doc-1",
		"attachmentIds":     []any{"doc-1"},
		"documentFormat":    "pdf",
	}
	files := []any{map[string]any{
		"id":   "doc-1",
		"kind": "requirement_document",
		"type": "application/pdf",
		"name": "requisito.pdf",
	}}
	if ok, message := reviewSubmissionIsRequirementDocument(submission, files); !ok {
		t.Fatalf("valid 5-page document must be accepted: %s", message)
	}

	submission["documentPageCount"] = 6
	if ok, _ := reviewSubmissionIsRequirementDocument(submission, files); ok {
		t.Fatal("documents above five pages must be rejected")
	}
}

func TestNormalReportStillRejectsDocumentAttachment5078(t *testing.T) {
	submission := map[string]any{
		"text":          "Informe normal",
		"attachmentIds": []any{"doc-1"},
	}
	files := []any{map[string]any{
		"id":   "doc-1",
		"kind": "requirement_document",
		"type": "application/pdf",
	}}
	if ok, _ := reviewSubmissionIsInternalReport(submission, files); ok {
		t.Fatal("a normal report must still reject a PDF document attachment")
	}
}

func TestAdminReportRequirementModesPersistAndReachAspirant5078(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"reportRequirementModes": map[string]any{},
		"settings":               map[string]any{"reportRequirementModes": map[string]any{}},
		"progress":               map[string]any{"asp1": map[string]any{"status": map[string]any{}, "submissions": map[string]any{}}},
		"programProgress":        map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"classes": []any{}, "tasks": []any{}, "materials": []any{}, "modules": []any{}, "audit": []any{}, "revision": 1.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 3, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}

	modes := map[string]any{
		"guide_major:REQ-01":    map[string]any{"mode": "document", "updatedBy": "admin1"},
		"guide_master:REQ-READ": map[string]any{"mode": "reading_report", "updatedBy": "admin1"},
	}
	body, _ := json.Marshal(map[string]any{
		"installationId": "admin-pc", "actorId": "admin1", "operationId": "SGL-ADM-REPORTMODES-5078",
		"summary": "Configuracion de informes",
		"patch": map[string]any{
			"reportRequirementModes": modes,
			"settings":               map[string]any{"reportRequirementModes": modes},
		},
		"files": []any{},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/admin/changes", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.legacyAdminChangeSubmit(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("submit=%d %s", rr.Code, rr.Body.String())
	}

	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	canonical := legacyMap(st.State.Database)
	if !reviewRequirementDocumentMode(canonical, "guide_major", "REQ-01") {
		t.Fatal("server canonical database lost the saved PDF/photo requirement mode")
	}
	if !reviewRequirementReadingReportMode(canonical, "guide_master", "REQ-READ") {
		t.Fatal("server canonical database lost the saved reading-report requirement mode")
	}
	if reviewRequirementDocumentMode(canonical, "guide_master", "REQ-READ") {
		t.Fatal("reading-report mode must not persist as PDF/photo mode")
	}
	settings := legacyMap(canonical["settings"])
	if !reviewRequirementDocumentMode(map[string]any{"settings": settings}, "guide_major", "REQ-01") {
		t.Fatal("server compatibility backup lost the saved report requirement mode")
	}
	if !reviewRequirementReadingReportMode(map[string]any{"settings": settings}, "guide_master", "REQ-READ") {
		t.Fatal("server compatibility backup lost the saved reading-report mode")
	}

	bootstrap := s.filteredLegacySnapshot(st.State, "asp1", "aspirant")
	clientDB := legacyMap(bootstrap["database"])
	if !reviewRequirementDocumentMode(clientDB, "guide_major", "REQ-01") {
		t.Fatal("aspirant bootstrap did not receive the Admin report mode configuration")
	}
	if !reviewRequirementReadingReportMode(clientDB, "guide_master", "REQ-READ") {
		t.Fatal("aspirant bootstrap did not receive the Admin reading-report configuration")
	}
}
