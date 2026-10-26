package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestSGL513DriveReopenReturnsFinalizedRequirementToDevelopment(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "role": "aspirant", "active": true, "accountState": "active"}},
		"progress": map[string]any{"asp1": map[string]any{
			"status": map[string]any{"I-02": "approved"},
			"submissions": map[string]any{"I-02": map[string]any{
				"text": "texto aprobado que debe quedar editable", "feedback": "", "updatedAt": "2026-09-16T22:00:00Z",
				"approval": map[string]any{"reviewerId": "admin1", "approvedAt": "2026-09-16T21:00:00Z"},
				"drive":    map[string]any{"status": "sent", "fileId": "drive-file-123", "sentAt": "2026-09-16T22:10:00Z"},
			}},
		}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"tasks":     []any{map[string]any{"id": "task-1", "assigneeId": "asp1", "programId": "class_grouped", "requirementId": "I-02", "presentedAt": "2026-09-16T20:00:00Z", "presentedBy": "asp1", "reviewSubmittedAt": "2026-09-16T20:00:00Z"}},
		"materials": []any{}, "indicationRules": []any{}, "classes": []any{}, "audit": []any{}, "revision": float64(8),
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 40, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}

	patch := map[string]any{
		"progress": map[string]any{"asp1": map[string]any{
			"status": map[string]any{"I-02": "in_progress"},
			"submissions": map[string]any{"I-02": map[string]any{
				"text": "NO DEBE REEMPLAZAR EL TEXTO", "feedback": "", "approval": nil, "updatedAt": "2026-09-16T22:30:00Z",
				"drive": map[string]any{"status": "deleted", "fileId": "", "deletedFileId": "drive-file-123", "deletedAt": "2026-09-16T22:30:00Z", "returnedToStatus": "in_progress"},
			}},
		}},
		"tasks": []any{map[string]any{"id": "task-1", "assigneeId": "asp1", "programId": "class_grouped", "requirementId": "I-02", "presentedAt": nil, "presentedBy": nil, "reviewSubmittedAt": nil, "updatedAt": "2026-09-16T22:30:00Z", "updatedBy": "asp1"}},
	}
	body, _ := json.Marshal(map[string]any{
		"installationId": "phone-asp1", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": 40,
		"operationId": "SGL-RDR-ASP1-I02-513", "operationType": "drive_reopen", "ownerId": "asp1",
		"requirementId": "I-02", "programId": "class_grouped", "ackOnly": true, "patch": patch, "files": []any{},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("drive reopen sync=%d %s", rr.Code, rr.Body.String())
	}
	var ack map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &ack); err != nil {
		t.Fatal(err)
	}
	if got := stringValue(ack["operationType"]); got != "drive_reopen" {
		t.Fatalf("operationType=%q ack=%#v", got, ack)
	}

	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	db := legacyMap(st.State.Database)
	progress := legacyMap(legacyMap(db["progress"])["asp1"])
	if got := stringValue(legacyMap(progress["status"])["I-02"]); got != "in_progress" {
		t.Fatalf("status=%q", got)
	}
	sub := legacyMap(legacyMap(progress["submissions"])["I-02"])
	if got := stringValue(sub["text"]); got != "texto aprobado que debe quedar editable" {
		t.Fatalf("approved body should be preserved as editable starting point, got %q", got)
	}
	if sub["approval"] != nil {
		t.Fatalf("approval must be cleared: %#v", sub["approval"])
	}
	drive := legacyMap(sub["drive"])
	if got := stringValue(drive["status"]); got != "deleted" {
		t.Fatalf("drive status=%q", got)
	}
	if got := stringValue(drive["fileId"]); got != "" {
		t.Fatalf("drive fileId must be empty, got %q", got)
	}
	tasks := arrayField(db, "tasks")
	task := legacyMap(tasks[0])
	if task["presentedAt"] != nil || task["reviewSubmittedAt"] != nil {
		t.Fatalf("task review markers were not reset: %#v", task)
	}
}

func TestSGL513DriveReopenRejectsApprovedRequirementThatWasNeverFinalized(t *testing.T) {
	merged := map[string]any{
		"users": []any{map[string]any{"id": "asp1", "role": "aspirant", "active": true, "accountState": "active"}},
		"progress": map[string]any{"asp1": map[string]any{
			"status":      map[string]any{"I-02": "approved"},
			"submissions": map[string]any{"I-02": map[string]any{"text": "x", "approval": map[string]any{"reviewerId": "a"}, "drive": map[string]any{"status": "pending", "fileId": ""}}},
		}},
		"programProgress": map[string]any{}, "tasks": []any{},
	}
	source := map[string]any{"progress": map[string]any{"asp1": map[string]any{
		"status":      map[string]any{"I-02": "in_progress"},
		"submissions": map[string]any{"I-02": map[string]any{"drive": map[string]any{"status": "deleted", "fileId": ""}}},
	}}}
	proposal := &legacyProposal{ActorID: "asp1", OwnerID: "asp1", RequirementID: "I-02", ProgramID: "class_grouped", OperationType: "drive_reopen"}
	if mergeDriveReopen(merged, source, proposal, "aspirant", map[string]bool{"asp1": true}, "2026-09-16T22:30:00Z") {
		t.Fatal("reopen must be rejected unless the canonical requirement was actually finalized in Drive")
	}
}

func TestSGL513TutorCanReopenDelegatedFinalizedRequirement(t *testing.T) {
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
				"text": "informe aprobado por tutor", "feedback": "", "updatedAt": "2026-09-16T20:00:00Z",
				"approval": map[string]any{"reviewerId": "tut1", "approvedAt": "2026-09-16T20:00:00Z"},
				"drive":    map[string]any{"status": "sent", "fileId": "gm-drive-10", "sentAt": "2026-09-16T22:15:00Z"},
			}},
		}}},
		"bibleYear": map[string]any{}, "configs": map[string]any{}, "tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "classes": []any{}, "audit": []any{}, "revision": float64(5),
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 50, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	patch := map[string]any{"programProgress": map[string]any{"asp1": map[string]any{"guide_major": map[string]any{
		"status": map[string]any{"GM-10": "in_progress"},
		"submissions": map[string]any{"GM-10": map[string]any{
			"text": "NO CAMBIAR", "feedback": "", "approval": nil, "updatedAt": "2026-09-16T22:45:00Z",
			"drive": map[string]any{"status": "deleted", "fileId": "", "deletedFileId": "gm-drive-10", "deletedAt": "2026-09-16T22:45:00Z", "returnedToStatus": "in_progress"},
		}},
	}}}}
	body, _ := json.Marshal(map[string]any{
		"installationId": "phone-tut1", "actorId": "tut1", "actorRole": "tutor", "baseVersion": 50,
		"operationId": "SGL-RDR-TUT1-GM10-513", "operationType": "drive_reopen", "ownerId": "asp1",
		"requirementId": "GM-10", "programId": "guide_major", "ackOnly": true, "patch": patch, "files": []any{},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("delegated drive reopen sync=%d %s", rr.Code, rr.Body.String())
	}
	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	programs := legacyMap(legacyMap(legacyMap(st.State.Database)["programProgress"])["asp1"])
	progress := legacyMap(programs["guide_major"])
	if got := stringValue(legacyMap(progress["status"])["GM-10"]); got != "in_progress" {
		t.Fatalf("status=%q", got)
	}
	sub := legacyMap(legacyMap(progress["submissions"])["GM-10"])
	if got := stringValue(sub["text"]); got != "informe aprobado por tutor" {
		t.Fatalf("approved text changed during delegated reopen: %q", got)
	}
	if sub["approval"] != nil {
		t.Fatalf("approval must be cleared: %#v", sub["approval"])
	}
	if got := stringValue(legacyMap(sub["drive"])["status"]); got != "deleted" {
		t.Fatalf("drive status=%q", got)
	}
}
