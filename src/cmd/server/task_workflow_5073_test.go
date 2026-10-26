package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func sync5073(t *testing.T, s *Server, actorID, actorRole, operationID string, baseVersion int64, patch map[string]any) {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"installationId": "test-" + actorID,
		"actorId":        actorID,
		"actorRole":      actorRole,
		"baseVersion":    baseVersion,
		"operationId":    operationID,
		"operationType":  "auto_sync",
		"programId":      "guide_major",
		"ackOnly":        true,
		"patch":          patch,
		"files":          []any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("sync %s status=%d body=%s", operationID, rr.Code, rr.Body.String())
	}
}

func TestSGL5073TaskStaysAssignedUntilTutorSendsToDevelopment(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Administrador", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "t1", "fullName": "Tutor Uno", "role": "tutor", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Lider Uno", "role": "aspirant", "tutorId": "t1", "active": true, "accountState": "active"},
		},
		"modules":  []any{map[string]any{"id": "guide_major", "active": true, "assignedUserIds": []any{"asp1"}}},
		"progress": map[string]any{"asp1": map[string]any{"status": map[string]any{}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{
			"asp1": map[string]any{
				"guide_major": map[string]any{
					"status": map[string]any{"GM-01": "pending", "GM-02": "assigned"},
					"submissions": map[string]any{
						"GM-01": map[string]any{"text": "", "richText": "", "completionDate": "", "attachmentIds": []any{}, "selections": []any{}},
					},
				},
			},
		},
		"bibleYear": map[string]any{}, "configs": map[string]any{},
		"tasks": []any{
			map[string]any{
				"id": "task-gm-2", "programId": "guide_major", "assigneeId": "asp1", "requirementId": "GM-02",
				"plannedDate": "2026-09-13", "dueDate": "2026-09-30", "createdBy": "t1", "createdByRole": "tutor",
				"active": true, "createdAt": "2026-09-13T01:00:00Z", "updatedAt": "2026-09-13T01:00:00Z",
			},
		},
		"materials": []any{}, "indicationRules": []any{}, "classes": []any{}, "audit": []any{}, "revision": float64(1),
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 20, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}

	// El Aspirante crea una tarea propia y guarda avance. El estado canónico se
	// mantiene "assigned": Guardar avance no puede convertirla en En desarrollo.
	selfTask := map[string]any{
		"id": "task-gm-1", "programId": "guide_major", "assigneeId": "asp1", "requirementId": "GM-01",
		"plannedDate": "2026-09-13", "dueDate": "2026-09-25", "planningDays": float64(13), "instructions": "Plan propio",
		"createdBy": "asp1", "createdByRole": "aspirant", "updatedBy": "asp1", "active": true,
		"selfPlanned": true, "lockedForAssignee": true, "taskWorkflowVersion": "5.0.75",
		"createdAt": "2026-09-13T02:00:00Z", "updatedAt": "2026-09-13T02:00:00Z",
	}
	progressAssigned := map[string]any{
		"asp1": map[string]any{
			"guide_major": map[string]any{
				"status": map[string]any{"GM-01": "assigned"},
				"submissions": map[string]any{
					"GM-01": map[string]any{"text": "Avance guardado", "richText": "<p>Avance guardado</p>", "completionDate": "2026-09-13", "attachmentIds": []any{}, "selections": []any{}, "updatedAt": "2026-09-13T02:00:00Z"},
				},
			},
		},
	}
	sync5073(t, s, "asp1", "aspirant", "SGL-5073-CREATE-TASK", 20, map[string]any{"tasks": []any{selfTask}, "programProgress": progressAssigned})

	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	db := legacyMap(st.State.Database)
	gm := legacyMap(legacyMap(legacyMap(db["programProgress"])["asp1"])["guide_major"])
	if got := stringValue(legacyMap(gm["status"])["GM-01"]); got != "assigned" {
		t.Fatalf("Guardar avance moved task to development: status=%q", got)
	}

	// Presentar solo marca presentedAt; la tarea sigue en assigned y por tanto
	// continúa en Tareas hasta la decisión del Tutor/Administrador.
	storedTasks := arrayField(db, "tasks")
	var presented map[string]any
	for _, raw := range storedTasks {
		item := legacyMap(raw)
		if stringValue(item["id"]) == "task-gm-1" {
			presented = legacyMap(cloneAny(item))
			break
		}
	}
	if presented == nil {
		t.Fatal("self task was not synchronized")
	}
	presented["presentedAt"] = "2026-09-13T03:00:00Z"
	presented["presentedBy"] = "asp1"
	presented["updatedAt"] = "2026-09-13T03:00:00Z"
	sync5073(t, s, "asp1", "aspirant", "SGL-5073-PRESENT-TASK", st.State.ServerVersion, map[string]any{"tasks": []any{presented}, "programProgress": progressAssigned})

	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	gm = legacyMap(legacyMap(legacyMap(db["programProgress"])["asp1"])["guide_major"])
	if got := stringValue(legacyMap(gm["status"])["GM-01"]); got != "assigned" {
		t.Fatalf("Presentar moved task to development before Tutor: status=%q", got)
	}
	var canonicalPresented map[string]any
	for _, raw := range arrayField(db, "tasks") {
		item := legacyMap(raw)
		if stringValue(item["id"]) == "task-gm-1" {
			canonicalPresented = item
			break
		}
	}
	if stringValue(canonicalPresented["presentedAt"]) == "" || stringValue(canonicalPresented["developmentAcceptedAt"]) != "" {
		t.Fatalf("unexpected presented task state: %#v", canonicalPresented)
	}

	// El Tutor acepta una tarea YA presentada. El servidor normaliza al Tutor que
	// tomó la decisión y recién entonces cambia el requisito a in_progress.
	tutorAccept := legacyMap(cloneAny(canonicalPresented))
	tutorAccept["developmentAcceptedAt"] = "2026-09-13T04:00:00Z"
	tutorAccept["developmentAcceptedBy"] = "fake-client-value"
	tutorAccept["updatedBy"] = "t1"
	tutorAccept["updatedAt"] = "2026-09-13T04:00:00Z"
	tutorProgress := map[string]any{
		"asp1": map[string]any{
			"guide_major": map[string]any{"status": map[string]any{"GM-01": "in_progress"}},
		},
	}
	sync5073(t, s, "t1", "tutor", "SGL-5073-TO-DEVELOPMENT", st.State.ServerVersion, map[string]any{"tasks": []any{tutorAccept}, "programProgress": tutorProgress})

	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	gm = legacyMap(legacyMap(legacyMap(db["programProgress"])["asp1"])["guide_major"])
	if got := stringValue(legacyMap(gm["status"])["GM-01"]); got != "in_progress" {
		t.Fatalf("Tutor acceptance did not move requirement to development: status=%q", got)
	}
	for _, raw := range arrayField(db, "tasks") {
		item := legacyMap(raw)
		if stringValue(item["id"]) != "task-gm-1" {
			continue
		}
		if stringValue(item["developmentAcceptedAt"]) == "" || stringValue(item["developmentAcceptedBy"]) != "t1" || stringValue(item["developmentAcceptedByRole"]) != "tutor" {
			t.Fatalf("development acceptance was not normalized: %#v", item)
		}
	}

	// Una tarea NO presentada no puede saltar directamente a En desarrollo.
	var unpresented map[string]any
	for _, raw := range arrayField(db, "tasks") {
		item := legacyMap(raw)
		if stringValue(item["id"]) == "task-gm-2" {
			unpresented = legacyMap(cloneAny(item))
			break
		}
	}
	unpresented["developmentAcceptedAt"] = "2026-09-13T05:00:00Z"
	unpresented["developmentAcceptedBy"] = "t1"
	invalidProgress := map[string]any{"asp1": map[string]any{"guide_major": map[string]any{"status": map[string]any{"GM-02": "in_progress"}}}}
	sync5073(t, s, "t1", "tutor", "SGL-5073-INVALID-DEVELOPMENT", st.State.ServerVersion, map[string]any{"tasks": []any{unpresented}, "programProgress": invalidProgress})

	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	gm = legacyMap(legacyMap(legacyMap(db["programProgress"])["asp1"])["guide_major"])
	if got := stringValue(legacyMap(gm["status"])["GM-02"]); got != "assigned" {
		t.Fatalf("unpresented task bypassed presentation gate: status=%q", got)
	}
	for _, raw := range arrayField(db, "tasks") {
		item := legacyMap(raw)
		if stringValue(item["id"]) == "task-gm-2" && stringValue(item["developmentAcceptedAt"]) != "" {
			t.Fatalf("unpresented task received development acceptance: %#v", item)
		}
	}
}
