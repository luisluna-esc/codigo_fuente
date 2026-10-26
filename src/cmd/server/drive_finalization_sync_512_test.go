package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestSGL512DriveFinalizationPersistsAndKeepsApprovedSubmission(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "role": "aspirant", "active": true, "accountState": "active"}},
		"progress": map[string]any{"asp1": map[string]any{
			"status": map[string]any{"I-02": "approved"},
			"submissions": map[string]any{"I-02": map[string]any{
				"text": "texto aprobado canonico", "updatedAt": "2026-09-16T22:00:00Z",
				"approval": map[string]any{"reviewerId": "admin1", "approvedAt": "2026-09-16T21:00:00Z"},
				"drive":    map[string]any{"status": "pending"},
			}},
		}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "classes": []any{}, "audit": []any{}, "revision": float64(7),
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 20, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}

	patch := map[string]any{"progress": map[string]any{"asp1": map[string]any{
		"status": map[string]any{"I-02": "approved"},
		"submissions": map[string]any{"I-02": map[string]any{
			// A drive-finalize operation must not be able to alter approved report content.
			"text": "NO DEBE REEMPLAZAR EL TEXTO", "updatedAt": "2026-09-16T22:10:00Z",
			"drive": map[string]any{
				"status": "sent", "fileId": "drive-file-123", "filename": "I-02 - Informe.pdf",
				"sentAt": "2026-09-16T22:10:00Z", "sentBy": "asp1",
			},
		}},
	}}}
	body, _ := json.Marshal(map[string]any{
		"installationId": "phone-asp1", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": 20,
		"operationId": "SGL-DRV-ASP1-I02-512", "operationType": "drive_finalize", "ownerId": "asp1",
		"requirementId": "I-02", "programId": "class_grouped", "ackOnly": true, "patch": patch, "files": []any{},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("drive finalize sync=%d %s", rr.Code, rr.Body.String())
	}
	var ack map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &ack); err != nil {
		t.Fatal(err)
	}
	if got := stringValue(ack["operationType"]); got != "drive_finalize" {
		t.Fatalf("operationType=%q ack=%#v", got, ack)
	}

	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	progress := legacyMap(legacyMap(legacyMap(st.State.Database)["progress"])["asp1"])
	if got := stringValue(legacyMap(progress["status"])["I-02"]); got != "approved" {
		t.Fatalf("status=%q", got)
	}
	sub := legacyMap(legacyMap(progress["submissions"])["I-02"])
	if got := stringValue(sub["text"]); got != "texto aprobado canonico" {
		t.Fatalf("approved text changed during drive finalize: %q", got)
	}
	drive := legacyMap(sub["drive"])
	if got := stringValue(drive["status"]); got != "sent" {
		t.Fatalf("drive status=%q submission=%#v", got, sub)
	}
	if got := stringValue(drive["fileId"]); got != "drive-file-123" {
		t.Fatalf("drive fileId=%q", got)
	}
}

func TestSGL512TutorCanPersistDelegatedDriveFinalizationWithoutEditingApprovedText(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "tut1", "fullName": "Tutor Uno", "role": "tutor", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "role": "aspirant", "active": true, "accountState": "active", "tutorId": "tut1"},
		},
		"progress": map[string]any{"asp1": map[string]any{"status": map[string]any{}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{"asp1": map[string]any{"guide_major": map[string]any{
			"status": map[string]any{"GM-10": "approved"},
			"submissions": map[string]any{"GM-10": map[string]any{
				"text": "informe aprobado por tutor", "updatedAt": "2026-09-16T20:00:00Z",
				"approval": map[string]any{"reviewerId": "tut1", "approvedAt": "2026-09-16T20:00:00Z"},
				"drive":    map[string]any{"status": "pending"},
			}},
		}}},
		"bibleYear": map[string]any{}, "configs": map[string]any{}, "tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "classes": []any{}, "audit": []any{}, "revision": float64(5),
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 30, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	patch := map[string]any{"programProgress": map[string]any{"asp1": map[string]any{"guide_major": map[string]any{
		"status": map[string]any{"GM-10": "approved"},
		"submissions": map[string]any{"GM-10": map[string]any{
			"text": "NO CAMBIAR", "updatedAt": "2026-09-16T22:15:00Z",
			"drive": map[string]any{"status": "sent", "fileId": "gm-drive-10", "sentAt": "2026-09-16T22:15:00Z", "sentBy": "tut1"},
		}},
	}}}}
	body, _ := json.Marshal(map[string]any{
		"installationId": "phone-tut1", "actorId": "tut1", "actorRole": "tutor", "baseVersion": 30,
		"operationId": "SGL-DRV-TUT1-GM10-512", "operationType": "drive_finalize", "ownerId": "asp1",
		"requirementId": "GM-10", "programId": "guide_major", "ackOnly": true, "patch": patch, "files": []any{},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("delegated drive finalize sync=%d %s", rr.Code, rr.Body.String())
	}
	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	programs := legacyMap(legacyMap(legacyMap(st.State.Database)["programProgress"])["asp1"])
	progress := legacyMap(programs["guide_major"])
	sub := legacyMap(legacyMap(progress["submissions"])["GM-10"])
	if got := stringValue(sub["text"]); got != "informe aprobado por tutor" {
		t.Fatalf("approved text changed during delegated drive finalize: %q", got)
	}
	drive := legacyMap(sub["drive"])
	if got := stringValue(drive["status"]); got != "sent" {
		t.Fatalf("drive status=%q", got)
	}
	if got := stringValue(drive["fileId"]); got != "gm-drive-10" {
		t.Fatalf("drive fileId=%q", got)
	}
}
