package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestAdminCorrectionAppliesDirectlyAndNotifiesOwner(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Nombre original", "phone": "111", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress":        map[string]any{"asp1": map[string]any{"status": map[string]any{"R1": "in_progress"}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"classes": []any{}, "tasks": []any{}, "materials": []any{}, "modules": []any{}, "audit": []any{}, "revision": 1.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 3, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]any{
		"installationId": "admin-pc", "actorId": "admin1", "operationId": "SGL-ADM-ONE",
		"summary": "Correccion puntual",
		"patch": map[string]any{
			"users":    []any{map[string]any{"id": "asp1", "fullName": "Nombre corregido", "phone": "222"}},
			"progress": map[string]any{"asp1": map[string]any{"status": map[string]any{"R1": "approved"}}},
		},
		"files": []any{},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/admin/changes", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.legacyAdminChangeSubmit(rr, req)
	if rr.Code != 200 {
		t.Fatalf("submit=%d %s", rr.Code, rr.Body.String())
	}
	var submit map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &submit)
	if submit["status"] != "applied" {
		t.Fatalf("status=%v body=%#v", submit["status"], submit)
	}
	if requestIDs, _ := submit["requestIds"].([]any); len(requestIDs) != 0 {
		t.Fatalf("admin direct apply must not create approval requests: %#v", submit["requestIds"])
	}

	st, _ := s.loadLegacy()
	if got := stringValue(legacyUser(st.State.Database, "asp1")["fullName"]); got != "Nombre corregido" {
		t.Fatalf("admin correction was not applied immediately: %q", got)
	}
	status := legacyMap(legacyMap(legacyMap(legacyMap(st.State.Database)["progress"])["asp1"])["status"])
	if got := stringValue(status["R1"]); got != "approved" {
		t.Fatalf("admin progress correction was not applied immediately: %q", got)
	}
	if len(st.UserChanges) != 0 {
		t.Fatalf("direct admin changes must not create pending user approvals: %#v", st.UserChanges)
	}
	notified := false
	for _, n := range st.Notifications {
		if n.UserID == "asp1" && n.Type == "admin_change_applied" && n.SourceOperationID == "SGL-ADM-ONE" {
			notified = true
		}
	}
	if !notified {
		t.Fatal("owner did not receive direct-apply notification")
	}
}

func TestAdminCorrectionRetryDeduplicatesAppliedOperation(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante", "phone": "111", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress": map[string]any{}, "programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "classes": []any{}, "audit": []any{}, "revision": 1.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 1, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(map[string]any{"actorId": "admin1", "operationId": "SGL-ADM-RETRY", "summary": "Telefono", "patch": map[string]any{"users": []any{map[string]any{"id": "asp1", "phone": "999"}}}})
	send := func() map[string]any {
		req := httptest.NewRequest(http.MethodPost, "http://example/api/admin/changes", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer adm")
		rr := httptest.NewRecorder()
		s.legacyAdminChangeSubmit(rr, req)
		if rr.Code != 200 {
			t.Fatalf("submit=%d %s", rr.Code, rr.Body.String())
		}
		var payload map[string]any
		_ = json.Unmarshal(rr.Body.Bytes(), &payload)
		return payload
	}
	first := send()
	second := send()
	if first["status"] != "applied" || second["status"] != "applied" || second["deduplicated"] != true {
		t.Fatalf("retry changed semantics: first=%#v second=%#v", first, second)
	}
	st, _ := s.loadLegacy()
	if len(st.UserChanges) != 0 {
		t.Fatalf("duplicate user changes: %d", len(st.UserChanges))
	}
	if got := stringValue(legacyUser(st.State.Database, "asp1")["phone"]); got != "999" {
		t.Fatalf("phone not applied: %q", got)
	}
}

func TestAdminTaskAssignmentAppliesDirectlyAndNotifies(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress":        map[string]any{"asp1": map[string]any{"status": map[string]any{"R1": "pending"}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"classes": []any{}, "tasks": []any{}, "materials": []any{}, "modules": []any{}, "audit": []any{}, "revision": 1.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 1, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"actorId": "admin1", "operationId": "SGL-ADM-TASK", "summary": "Tarea",
		"patch": map[string]any{
			"tasks":    []any{map[string]any{"id": "task1", "assigneeId": "asp1", "requirementId": "R1", "programId": "class_grouped", "dueDate": "2026-09-10", "active": true}},
			"progress": map[string]any{"asp1": map[string]any{"status": map[string]any{"R1": "assigned"}}},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/admin/changes", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.legacyAdminChangeSubmit(rr, req)
	if rr.Code != 200 {
		t.Fatalf("submit=%d %s", rr.Code, rr.Body.String())
	}
	var payload map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &payload)
	if payload["status"] != "applied" {
		t.Fatalf("task assignment should not require user approval: %#v", payload)
	}
	st, _ := s.loadLegacy()
	if len(st.UserChanges) != 0 {
		t.Fatalf("created user change request for task: %#v", st.UserChanges)
	}
	status := legacyMap(legacyMap(legacyMap(legacyMap(st.State.Database)["progress"])["asp1"])["status"])
	if got := stringValue(status["R1"]); got != "assigned" {
		t.Fatalf("assigned status=%q", got)
	}
	tasks := arrayField(legacyMap(st.State.Database), "tasks")
	if len(tasks) != 1 || stringValue(legacyMap(tasks[0])["id"]) != "task1" {
		t.Fatalf("task not applied: %#v", tasks)
	}
	foundNotice := false
	for _, n := range st.Notifications {
		if n.UserID == "asp1" && n.Type == "task_assigned" {
			foundNotice = true
		}
	}
	if !foundNotice {
		t.Fatal("assignee did not receive task notification")
	}
}

func TestNotificationDeleteReadOnly(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{"users": []any{map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"}}}
	st := legacyStore{State: legacyState{ServerVersion: 1, Database: db, Files: []any{}}, Notifications: []legacyNotification{
		{ID: "n1", UserID: "asp1", Type: "info", Title: "Leido", ReadAt: "2026-09-01T00:00:00Z", CreatedAt: "2026-09-01T00:00:00Z"},
		{ID: "n2", UserID: "asp1", Type: "info", Title: "Nuevo", CreatedAt: "2026-09-01T00:01:00Z"},
	}}
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"userId": "asp1", "readOnly": true})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/notifications/delete", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacyNotificationDelete(rr, req)
	if rr.Code != 200 {
		t.Fatalf("delete=%d %s", rr.Code, rr.Body.String())
	}
	st2, _ := s.loadLegacy()
	if len(st2.Notifications) != 2 || st2.Notifications[0].ID != "n1" || st2.Notifications[0].DeletedAt == "" || st2.Notifications[1].ID != "n2" || st2.Notifications[1].DeletedAt != "" {
		t.Fatalf("read-only delete did not preserve the expected tombstone: %#v", st2.Notifications)
	}
	getReq := httptest.NewRequest(http.MethodGet, "http://example/api/notifications?userId=asp1", nil)
	getReq.Header.Set("Authorization", "Bearer cli")
	getRR := httptest.NewRecorder()
	s.legacyNotifications(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("notifications=%d %s", getRR.Code, getRR.Body.String())
	}
	var visible struct {
		Notifications []legacyNotification `json:"notifications"`
	}
	if err := json.Unmarshal(getRR.Body.Bytes(), &visible); err != nil {
		t.Fatal(err)
	}
	if len(visible.Notifications) != 1 || visible.Notifications[0].ID != "n2" {
		t.Fatalf("deleted notice leaked through API: %#v", visible.Notifications)
	}
	// A deleted notification stays as a tombstone so replaying the same operation
	// cannot recreate it on another device.
	addLegacyNotification(&st2, legacyNotification{UserID: "asp1", Type: "info", Title: "Leido otra vez", SourceOperationID: "op-delete-test", CreatedAt: "2026-09-01T00:02:00Z"})
	addLegacyNotification(&st2, legacyNotification{UserID: "asp1", Type: "approved", Title: "Aprobado", SourceOperationID: "op-approved", CreatedAt: "2026-09-01T00:03:00Z"})
	st2.Notifications[len(st2.Notifications)-1].DeletedAt = "2026-09-06T00:00:00Z"
	before := len(st2.Notifications)
	addLegacyNotification(&st2, legacyNotification{UserID: "asp1", Type: "approved", Title: "Aprobado repetido", SourceOperationID: "op-approved", CreatedAt: "2026-09-01T00:04:00Z"})
	if len(st2.Notifications) != before {
		t.Fatalf("deleted notification was recreated: %#v", st2.Notifications)
	}
}

func TestAdminClassChangeNotifiesOnlyAssignedInstructor(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "juan", "fullName": "Cuenta de prueba", "role": "aspirant", "active": true, "accountState": "active"},
			map[string]any{"id": "lizeth", "fullName": "Lizeth Tatiana Sarzuri Isidro", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress": map[string]any{}, "programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"classes": []any{map[string]any{"id": "A", "name": "Amigo + Amigo de la Naturaleza", "regular": "Amigo", "advanced": "Amigo de la Naturaleza", "assignedUserIds": []any{"lizeth"}, "active": true}},
		"tasks":   []any{}, "materials": []any{}, "modules": []any{}, "audit": []any{}, "revision": 1.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 4, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"actorId": "admin1", "operationId": "SGL-ADM-CLASS", "entityId": "A",
		"summary": "Clase actualizada - Amigo + Amigo de la Naturaleza. 1 instructor(es)",
		"patch":   map[string]any{"classes": []any{map[string]any{"id": "A", "assignedUserIds": []any{"lizeth"}, "updatedAt": "2026-09-19T22:00:00Z"}}},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/admin/changes", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.legacyAdminChangeSubmit(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("submit=%d %s", rr.Code, rr.Body.String())
	}
	st, _ := s.loadLegacy()
	var lizethNotice bool
	for _, n := range st.Notifications {
		if n.Type != "admin_change_applied" || n.SourceOperationID != "SGL-ADM-CLASS" {
			continue
		}
		if n.UserID == "juan" {
			t.Fatalf("unrelated user received class notification: %#v", n)
		}
		if n.UserID == "lizeth" && n.OwnerID == "lizeth" {
			lizethNotice = true
		}
	}
	if !lizethNotice {
		t.Fatal("assigned Instructor did not receive the class notification")
	}
}

func TestHistoricalAdminNoticeRelevanceRejectsOtherUserAndOtherClass(t *testing.T) {
	db := map[string]any{
		"users": []any{
			map[string]any{"id": "juan", "fullName": "Cuenta de prueba", "role": "aspirant", "active": true, "accountState": "active"},
			map[string]any{"id": "lizeth", "fullName": "Lizeth Tatiana Sarzuri Isidro", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"classes": []any{map[string]any{"id": "A", "name": "Amigo + Amigo de la Naturaleza", "assignedUserIds": []any{"lizeth"}}},
	}
	otherUser := legacyNotification{UserID: "juan", Type: "admin_change_applied", Message: "Usuario actualizado - Lizeth Tatiana Sarzuri Isidro - Lider / Aspirante"}
	if adminAppliedNotificationRelevant(db, "juan", otherUser) {
		t.Fatal("notification for another user must be hidden")
	}
	otherClass := legacyNotification{UserID: "juan", Type: "admin_change_applied", Message: "Clase actualizada - Amigo + Amigo de la Naturaleza. 1 instructor(es)"}
	if adminAppliedNotificationRelevant(db, "juan", otherClass) {
		t.Fatal("notification for an unrelated class must be hidden")
	}
	if !adminAppliedNotificationRelevant(db, "lizeth", otherClass) {
		t.Fatal("assigned Instructor should keep the historical class notification")
	}
	owned := legacyNotification{UserID: "juan", OwnerID: "juan", Type: "admin_change_applied", Message: "Cambio administrativo"}
	if !adminAppliedNotificationRelevant(db, "juan", owned) {
		t.Fatal("owner-tagged notification should remain visible to its owner")
	}
}
