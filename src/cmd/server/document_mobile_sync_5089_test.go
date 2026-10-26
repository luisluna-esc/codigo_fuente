package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestSGL5089MobileDocumentSaveProgressPersistsFileAndProgramSubmission(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
 "modules": []any{map[string]any{"id":"guide_major","active":true,"assignedUserIds":[]any{"asp1"}}},
		"users":    []any{map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "role": "aspirant", "active": true, "accountState": "active"}},
		"progress": map[string]any{"asp1": map[string]any{"status": map[string]any{}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{"asp1": map[string]any{"guide_major": map[string]any{
			"status":      map[string]any{"GM-DOC": "pending"},
			"submissions": map[string]any{"GM-DOC": map[string]any{"attachmentIds": []any{}}},
		}}},
		"bibleYear": map[string]any{}, "configs": map[string]any{}, "tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "classes": []any{}, "audit": []any{}, "revision": float64(1),
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 10, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}

	patch := map[string]any{"programProgress": map[string]any{"asp1": map[string]any{"guide_major": map[string]any{
		"status": map[string]any{"GM-DOC": "in_progress"},
		"submissions": map[string]any{"GM-DOC": map[string]any{
			"submissionMode": "document", "deliveryMode": "document", "documentMode": true, "documentFormat": "pdf",
			"primaryDocumentId": "doc-mobile-1", "documentPageCount": float64(1), "documentSource": "images",
			"attachmentIds": []any{"doc-mobile-1"}, "completionDate": "2026-09-14", "updatedAt": "2026-09-14T23:30:00Z",
		}},
	}}}}
	file := map[string]any{
		"id": "doc-mobile-1", "ownerId": "asp1", "programId": "guide_major", "requirementId": "GM-DOC",
		"name": "GM-DOC_documento.pdf", "kind": "requirement_document", "sourceType": "images", "pageCount": float64(1),
		"type": "application/pdf", "size": float64(12), "dataUrl": "data:application/pdf;base64,JVBERi0xLjQKJSVFT0Y=",
	}
	body, _ := json.Marshal(map[string]any{
		"installationId": "phone-1", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": 10,
		"operationId": "SGL-5089-MOBILE-DOC-SAVE", "operationType": "save_progress", "requirementId": "GM-DOC", "programId": "guide_major",
		"patch": patch, "files": []any{file},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("sync=%d %s", rr.Code, rr.Body.String())
	}

	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	gm := legacyMap(legacyMap(legacyMap(st.State.Database)["programProgress"])["asp1"])
	gm = legacyMap(gm["guide_major"])
	if got := stringValue(legacyMap(gm["status"])["GM-DOC"]); got != "in_progress" {
		t.Fatalf("status=%q", got)
	}
	sub := legacyMap(legacyMap(gm["submissions"])["GM-DOC"])
	if got := stringValue(sub["primaryDocumentId"]); got != "doc-mobile-1" {
		t.Fatalf("primaryDocumentId=%q submission=%#v", got, sub)
	}
	if len(st.State.Files) != 1 || stringValue(legacyMap(st.State.Files[0])["id"]) != "doc-mobile-1" {
		t.Fatalf("document file not persisted: %#v", st.State.Files)
	}
}
