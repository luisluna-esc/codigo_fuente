package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestSGL5071SelfPlannedTaskSyncsAndTutorCanEdit(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Administrador", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "t1", "fullName": "Tutor Uno", "role": "tutor", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Lider Uno", "role": "aspirant", "tutorId": "t1", "active": true, "accountState": "active"},
		},
		"modules":         []any{map[string]any{"id": "class_grouped", "active": true, "assignedUserIds": []any{"asp1"}}},
		"progress":        map[string]any{"asp1": map[string]any{"status": map[string]any{"I-01": "pending"}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "classes": []any{}, "audit": []any{}, "revision": float64(1),
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 10, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}

	selfTask := map[string]any{
		"id": "task-self-1", "programId": "class_grouped", "assigneeId": "asp1", "requirementId": "I-01",
		"plannedDate": "2026-09-13", "dueDate": "2026-09-20", "planningDays": float64(8), "instructions": "Plan propio",
		"createdBy": "asp1", "createdByRole": "aspirant", "updatedBy": "asp1", "active": true,
		"selfPlanned": true, "lockedForAssignee": true, "createdAt": "2026-09-13T01:00:00Z", "updatedAt": "2026-09-13T01:00:00Z",
	}
	body, _ := json.Marshal(map[string]any{
		"installationId": "phone1", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": 10,
		"operationId": "SGL-5071-SELF-TASK-1", "operationType": "auto_sync", "programId": "class_grouped", "ackOnly": true,
		"patch": map[string]any{
			"tasks":    []any{selfTask},
			"progress": map[string]any{"asp1": map[string]any{"status": map[string]any{"I-01": "assigned"}}},
		}, "files": []any{},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("aspirant task sync status=%d body=%s", rr.Code, rr.Body.String())
	}

	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	tasks := arrayField(legacyMap(st.State.Database), "tasks")
	if len(tasks) != 1 {
		t.Fatalf("expected one synchronized self task, got %d: %#v", len(tasks), tasks)
	}
	stored := legacyMap(tasks[0])
	if stringValue(stored["id"]) != "task-self-1" || stringValue(stored["createdBy"]) != "asp1" || stored["selfPlanned"] != true {
		t.Fatalf("self-planned task not preserved: %#v", stored)
	}

	// The canonical server snapshot visible to the assigned Tutor must include
	// the leader's self-planned task, not only tasks created by the Tutor.
	tutorSnap := s.filteredLegacySnapshot(st.State, "t1", "tutor")
	tutorTasks := arrayField(legacyMap(tutorSnap["database"]), "tasks")
	if len(tutorTasks) != 1 || stringValue(legacyMap(tutorTasks[0])["id"]) != "task-self-1" {
		t.Fatalf("tutor snapshot does not include self-planned task: %#v", tutorTasks)
	}
	adminSnap := s.filteredLegacySnapshot(st.State, "admin1", "admin")
	adminTasks := arrayField(legacyMap(adminSnap["database"]), "tasks")
	if len(adminTasks) != 1 || stringValue(legacyMap(adminTasks[0])["id"]) != "task-self-1" {
		t.Fatalf("admin snapshot does not include self-planned task: %#v", adminTasks)
	}

	// Tutor may edit dates/instructions even when the task was created by the leader.
	tutorEdit := map[string]any{}
	for k, v := range stored {
		tutorEdit[k] = cloneAny(v)
	}
	tutorEdit["dueDate"] = "2026-09-25"
	tutorEdit["instructions"] = "Fecha ajustada por Tutor"
	tutorEdit["updatedBy"] = "t1"
	tutorEdit["updatedAt"] = "2026-09-13T02:00:00Z"
	tutorBody, _ := json.Marshal(map[string]any{
		"installationId": "tutor-laptop", "actorId": "t1", "actorRole": "tutor", "baseVersion": st.State.ServerVersion,
		"operationId": "SGL-5071-TUTOR-EDIT-1", "operationType": "auto_sync", "programId": "class_grouped", "ackOnly": true,
		"patch": map[string]any{"tasks": []any{tutorEdit}}, "files": []any{},
	})
	tutorReq := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(tutorBody))
	tutorReq.Header.Set("Authorization", "Bearer cli")
	tutorRR := httptest.NewRecorder()
	s.legacySync(tutorRR, tutorReq)
	if tutorRR.Code != http.StatusOK {
		t.Fatalf("tutor edit status=%d body=%s", tutorRR.Code, tutorRR.Body.String())
	}
	st, _ = s.loadLegacy()
	stored = legacyMap(arrayField(legacyMap(st.State.Database), "tasks")[0])
	if got := stringValue(stored["dueDate"]); got != "2026-09-25" {
		t.Fatalf("tutor edit did not reach canonical task: dueDate=%q task=%#v", got, stored)
	}

	// A later Aspirant sync cannot rewrite the locked plan dates after save.
	malicious := map[string]any{}
	for k, v := range stored {
		malicious[k] = cloneAny(v)
	}
	malicious["dueDate"] = "2027-01-01"
	malicious["instructions"] = "No debería aplicarse"
	malicious["updatedBy"] = "asp1"
	aspBody, _ := json.Marshal(map[string]any{
		"installationId": "phone1", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": st.State.ServerVersion,
		"operationId": "SGL-5071-LOCKED-EDIT-1", "operationType": "auto_sync", "programId": "class_grouped", "ackOnly": true,
		"patch": map[string]any{"tasks": []any{malicious}}, "files": []any{},
	})
	aspReq := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(aspBody))
	aspReq.Header.Set("Authorization", "Bearer cli")
	aspRR := httptest.NewRecorder()
	s.legacySync(aspRR, aspReq)
	if aspRR.Code != http.StatusOK {
		t.Fatalf("locked aspirant edit status=%d body=%s", aspRR.Code, aspRR.Body.String())
	}
	st, _ = s.loadLegacy()
	stored = legacyMap(arrayField(legacyMap(st.State.Database), "tasks")[0])
	if got := stringValue(stored["dueDate"]); got != "2026-09-25" {
		t.Fatalf("aspirant changed locked task after save: dueDate=%q task=%#v", got, stored)
	}
}
