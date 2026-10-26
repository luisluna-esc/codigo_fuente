package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestReportConfig5079ReachesClientWithoutFullBootstrap(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	modes := map[string]any{
		"class_grouped:REQ-DOC":  map[string]any{"mode": "document"},
		"guide_major:REQ-READ":   map[string]any{"mode": "reading_report"},
		"guide_master:REQ-READ2": map[string]any{"mode": "reading_report"},
	}
	db := map[string]any{
		"reportRequirementModes": modes,
		"settings":               map[string]any{"reportRequirementModes": modes},
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 17, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://example/api/report-config", nil)
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.reportConfig5079(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["compatVersion"] != "5.6" {
		t.Fatalf("compatVersion=%v", payload["compatVersion"])
	}
	if int64(payload["serverVersion"].(float64)) != 17 {
		t.Fatalf("serverVersion=%v", payload["serverVersion"])
	}
	got := legacyMap(payload["reportRequirementModes"])
	if !reviewRequirementDocumentMode(map[string]any{"reportRequirementModes": got}, "class_grouped", "REQ-DOC") {
		t.Fatal("client endpoint lost document assignment")
	}
	if !reviewRequirementReadingReportMode(map[string]any{"reportRequirementModes": got}, "guide_major", "REQ-READ") {
		t.Fatal("client endpoint lost reading report assignment")
	}
}

func TestReportConfig5079RequiresAuthentication(t *testing.T) {
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(t.TempDir(), "data.json")}
	req := httptest.NewRequest(http.MethodGet, "http://example/api/report-config", nil)
	rr := httptest.NewRecorder()
	s.reportConfig5079(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
}
