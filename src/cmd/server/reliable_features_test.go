package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestDedupeOldProposalsIgnoresVolatileFields(t *testing.T) {
	db1 := map[string]any{
		"revision":  10.0,
		"updatedAt": "2026-08-31T01:00:00Z",
		"sync":      map[string]any{"state": "syncing", "message": "uno"},
		"progress":  map[string]any{"asp1": map[string]any{"status": map[string]any{"C1": "in_review"}}},
	}
	db2 := cloneAny(db1).(map[string]any)
	db2["revision"] = 14.0
	db2["updatedAt"] = "2026-08-31T01:01:00Z"
	db2["sync"] = map[string]any{"state": "pending", "message": "reintento"}
	st := legacyStore{Proposals: []legacyProposal{
		{ID: "old1", Status: "pending", ActorID: "asp1", InstallationID: "phone1", Revision: 10, SubmittedAt: "2026-08-31T01:00:00Z", UpdatedAt: "2026-08-31T01:00:00Z", Database: db1},
		{ID: "old2", Status: "pending", ActorID: "asp1", InstallationID: "phone1", Revision: 14, SubmittedAt: "2026-08-31T01:01:00Z", UpdatedAt: "2026-08-31T01:01:00Z", Database: db2},
	}}
	changed, superseded := dedupeLegacyProposals(&st)
	if !changed || superseded != 1 {
		t.Fatalf("expected one repaired duplicate, changed=%v superseded=%d state=%#v", changed, superseded, st.Proposals)
	}
	statuses := map[string]string{}
	for _, p := range st.Proposals {
		statuses[p.ID] = p.Status
	}
	if statuses["old2"] != "pending" || statuses["old1"] != "superseded" {
		t.Fatalf("newest must remain canonical: %#v", statuses)
	}
}

func TestLegacySyncOperationIdIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress":        map[string]any{"asp1": map[string]any{"status": map[string]any{"C1": "in_progress"}, "submissions": map[string]any{}}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "revision": 1.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 1, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	send := func(revision float64) map[string]any {
		db := cloneAny(baseDB).(map[string]any)
		db["revision"] = revision
		legacyMap(legacyMap(legacyMap(db["progress"])["asp1"])["status"])["C1"] = "in_progress"
		body, _ := json.Marshal(map[string]any{
			"installationId": "phone1", "actorId": "asp1", "actorRole": "aspirant", "actorName": "Aspirante", "baseVersion": 1,
			"operationId": "SGL-REV-ASP1-C1-CLASS-ABC123", "operationType": "save_progress", "requirementId": "C1", "programId": "class_grouped",
			"database": db, "files": []any{},
		})
		req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer cli")
		rr := httptest.NewRecorder()
		s.legacySync(rr, req)
		if rr.Code != 200 {
			t.Fatalf("sync status=%d body=%s", rr.Code, rr.Body.String())
		}
		var payload map[string]any
		_ = json.Unmarshal(rr.Body.Bytes(), &payload)
		if payload["status"] != "applied" {
			t.Fatalf("expected applied status, payload=%#v", payload)
		}
		return payload
	}
	first := send(2)
	second := send(3)
	if first["operationId"] != second["operationId"] || second["deduplicated"] != true {
		t.Fatalf("same operation was not deduplicated: %#v %#v", first, second)
	}
	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Proposals) != 0 {
		t.Fatalf("automatic sync must not create administrative proposals: %#v", st.Proposals)
	}
	if len(st.AppliedOperations) != 1 {
		t.Fatalf("expected one applied operation receipt, got %d", len(st.AppliedOperations))
	}
	if st.State.ServerVersion != 2 {
		t.Fatalf("retry must not increment server version twice, got %d", st.State.ServerVersion)
	}
	status := legacyMap(legacyMap(legacyMap(legacyMap(st.State.Database)["progress"])["asp1"])["status"])
	if got := stringValue(status["C1"]); got != "in_progress" {
		t.Fatalf("review did not synchronize automatically: %q", got)
	}
}

func TestAutomaticSyncAppliesAssignedClassWorkWithoutProposal(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Instructor", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"classes":  []any{map[string]any{"id": "class1", "assignedUserIds": []any{"asp1"}, "work": map[string]any{"planning": map[string]any{}, "attendance": []any{}, "tracking": map[string]any{}, "members": []any{}}}},
		"progress": map[string]any{}, "programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "revision": 1.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 4, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	incoming := cloneAny(baseDB).(map[string]any)
	class := legacyMap(arrayField(incoming, "classes")[0])
	class["work"] = map[string]any{
		"planning":   map[string]any{"R1": map[string]any{"date": "2026-09-01", "objective": "Plan sincronizado"}},
		"attendance": []any{map[string]any{"id": "att1", "date": "2026-09-01", "entries": map[string]any{"m:m1": "Presente"}}},
		"tracking":   map[string]any{"m1:R1": "completed"},
		"members":    []any{map[string]any{"id": "m1", "fullName": "Conquistador", "photo": "data:image/png;base64,AAA"}},
	}
	body, _ := json.Marshal(map[string]any{"installationId": "phone1", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": 4, "operationId": "SGL-AUT-CLASS-1", "operationType": "auto_sync", "database": incoming, "files": []any{}})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != 200 {
		t.Fatalf("automatic class sync status=%d body=%s", rr.Code, rr.Body.String())
	}
	st, _ := s.loadLegacy()
	if len(st.Proposals) != 0 {
		t.Fatalf("class work created a proposal: %#v", st.Proposals)
	}
	storedClass := legacyMap(arrayField(legacyMap(st.State.Database), "classes")[0])
	work := legacyMap(storedClass["work"])
	plan := legacyMap(legacyMap(work["planning"])["R1"])
	if stringValue(plan["objective"]) != "Plan sincronizado" || len(arrayField(work, "attendance")) != 1 || stringValue(legacyMap(work["tracking"])["m1:R1"]) != "completed" {
		t.Fatalf("class work did not synchronize: %#v", work)
	}
}

func TestTutorCreatedAspirantSynchronizesAutomatically(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "t1", "fullName": "Tutor", "role": "tutor", "active": true, "accountState": "active"},
		},
		"progress": map[string]any{}, "programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "revision": 1.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 7, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	incoming := cloneAny(baseDB).(map[string]any)
	incoming["users"] = append(arrayField(incoming, "users"), map[string]any{"id": "new1", "fullName": "Nuevo Aspirante", "role": "aspirant", "tutorId": "t1", "active": true, "accountState": "active", "username": "nuevo"})
	ensureMapField(incoming, "progress")["new1"] = map[string]any{"status": map[string]any{"R1": "pending"}, "submissions": map[string]any{}}
	body, _ := json.Marshal(map[string]any{"installationId": "laptop1", "actorId": "t1", "actorRole": "tutor", "baseVersion": 7, "operationId": "SGL-AUT-TUTOR-NEW1", "operationType": "auto_sync", "database": incoming, "files": []any{}})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != 200 {
		t.Fatalf("tutor automatic sync status=%d body=%s", rr.Code, rr.Body.String())
	}
	st, _ := s.loadLegacy()
	created := legacyUser(st.State.Database, "new1")
	if created == nil || stringValue(created["tutorId"]) != "t1" || legacyRole(created) != "aspirant" {
		t.Fatalf("tutor-created aspirant was not synchronized safely: %#v", created)
	}
	if legacyMap(legacyMap(st.State.Database)["progress"])["new1"] == nil {
		t.Fatalf("new aspirant progress was not initialized on server")
	}
	if len(st.Proposals) != 0 {
		t.Fatalf("tutor-created aspirant must not require admin proposal: %#v", st.Proposals)
	}
}

func TestCancelProposalIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{"users": []any{map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"}}}
	st := legacyStore{State: legacyState{ServerVersion: 1, Database: db, Files: []any{}}, Proposals: []legacyProposal{{ID: "p1", Status: "pending", ActorID: "asp1", InstallationID: "phone1", SubmittedAt: "2026-08-31T01:00:00Z"}}}
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	call := func() int {
		body, _ := json.Marshal(map[string]any{"actorId": "asp1", "proposalId": "p1", "installationId": "phone1"})
		req := httptest.NewRequest(http.MethodPost, "http://example/api/proposals/cancel", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer cli")
		rr := httptest.NewRecorder()
		s.legacyProposalCancel(rr, req)
		return rr.Code
	}
	if code := call(); code != 200 {
		t.Fatalf("first cancel status=%d", code)
	}
	if code := call(); code != 200 {
		t.Fatalf("second cancel must be idempotent, status=%d", code)
	}
}

func TestMessageClientIdDeduplicatesRetries(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{"users": []any{
		map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
		map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"},
	}}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 1, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	call := func() {
		body, _ := json.Marshal(map[string]any{"senderId": "admin1", "recipientId": "asp1", "text": "Mensaje de prueba", "clientMessageId": "msg-client-1"})
		req := httptest.NewRequest(http.MethodPost, "http://example/api/messages/send", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer adm")
		rr := httptest.NewRecorder()
		s.legacyMessageSend(rr, req)
		if rr.Code != 201 && rr.Code != 200 {
			t.Fatalf("message status=%d body=%s", rr.Code, rr.Body.String())
		}
	}
	call()
	call()
	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Messages) != 1 {
		t.Fatalf("expected one persisted message, got %d", len(st.Messages))
	}
}

func TestMessageGroupConversationSupportsUpToFiveParticipants(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{"users": []any{
		map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
		map[string]any{"id": "tutor1", "fullName": "Tutor", "role": "tutor", "active": true, "accountState": "active"},
		map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active", "tutorId": "tutor1"},
	}}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 1, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"senderId": "admin1", "recipientIds": []string{"tutor1", "asp1"}, "conversationId": "conv-admin-tutor-asp", "text": "Mensaje grupal", "clientMessageId": "group-msg-1"})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/messages/send", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.legacyMessageSend(rr, req)
	if rr.Code != 201 {
		t.Fatalf("group message status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	message := legacyMap(payload["message"])
	if message["conversationId"] != "conv-admin-tutor-asp" {
		t.Fatalf("missing conversation id: %#v", message)
	}
	recipients := arrayField(message, "recipientIds")
	if len(recipients) != 2 {
		t.Fatalf("expected two group recipients: %#v", recipients)
	}
	listReq := httptest.NewRequest(http.MethodGet, "http://example/api/messages?userId=asp1", nil)
	listReq.Header.Set("Authorization", "Bearer adm")
	listRR := httptest.NewRecorder()
	s.legacyMessages(listRR, listReq)
	if listRR.Code != 200 {
		t.Fatalf("list=%d body=%s", listRR.Code, listRR.Body.String())
	}
	var listed map[string]any
	_ = json.Unmarshal(listRR.Body.Bytes(), &listed)
	if messages := arrayField(listed, "messages"); len(messages) != 1 {
		t.Fatalf("aspirant should see group message: %#v", listed)
	}
}

func TestServerSideProposalMergeKeepsAspirantScoped(t *testing.T) {
	serverDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active", "phone": "admin-original"},
			map[string]any{"id": "asp1", "fullName": "Aspirant", "role": "aspirant", "active": true, "accountState": "active", "phone": "old"},
		},
		"progress": map[string]any{
			"admin1": map[string]any{"secret": "keep"},
			"asp1":   map[string]any{"status": map[string]any{"R1": "draft"}},
		},
		"programProgress": map[string]any{},
		"bibleYear":       map[string]any{},
		"configs":         map[string]any{},
		"revision":        3.0,
	}
	proposalDB := cloneAny(serverDB).(map[string]any)
	legacyUser(proposalDB, "admin1")["phone"] = "malicious-change"
	legacyUser(proposalDB, "asp1")["phone"] = "new-phone"
	proposalDB["progress"].(map[string]any)["asp1"] = map[string]any{"status": map[string]any{"R1": "in_progress"}}
	proposalDB["progress"].(map[string]any)["admin1"] = map[string]any{"secret": "changed"}
	p := &legacyProposal{ID: "p1", Status: "pending", ActorID: "asp1", ActorRole: "aspirant", Database: proposalDB}

	mergedAny, allowed, err := mergeLegacyProposalOnServer(serverDB, p)
	if err != nil {
		t.Fatal(err)
	}
	merged := legacyMap(mergedAny)
	if !allowed["asp1"] || len(allowed) != 1 {
		t.Fatalf("unexpected allowed users: %#v", allowed)
	}
	if got := stringValue(legacyUser(merged, "admin1")["phone"]); got != "admin-original" {
		t.Fatalf("admin field was changed by aspirant proposal: %q", got)
	}
	if got := stringValue(legacyUser(merged, "asp1")["phone"]); got != "new-phone" {
		t.Fatalf("aspirant own field was not applied: %q", got)
	}
	if got := stringValue(legacyMap(legacyMap(merged["progress"])["admin1"])["secret"]); got != "keep" {
		t.Fatalf("admin progress was changed: %q", got)
	}
	status := legacyMap(legacyMap(legacyMap(merged["progress"])["asp1"])["status"])
	if got := stringValue(status["R1"]); got != "in_progress" {
		t.Fatalf("aspirant progress not applied: %q", got)
	}
}

func TestCompactAcceptDecisionUsesServerMerge(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirant", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress":        map[string]any{"asp1": map[string]any{"status": map[string]any{"R1": "draft"}}},
		"programProgress": map[string]any{},
		"bibleYear":       map[string]any{},
		"configs":         map[string]any{},
		"revision":        1.0,
	}
	proposalDB := cloneAny(baseDB).(map[string]any)
	proposalDB["progress"].(map[string]any)["asp1"] = map[string]any{"status": map[string]any{"R1": "in_progress"}}
	st := legacyStore{
		State:     legacyState{ServerVersion: 5, Database: baseDB, Files: []any{}},
		Proposals: []legacyProposal{{ID: "p1", Status: "pending", ActorID: "asp1", ActorName: "Aspirant", ActorRole: "aspirant", InstallationID: "phone1", Database: proposalDB, Files: []any{}}},
	}
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"actorId": "admin1", "proposalId": "p1", "decision": "accept", "note": "ok"})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/proposals/decision", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.legacyProposalDecision(rr, req)
	if rr.Code != 200 {
		t.Fatalf("compact accept status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &payload)
	if payload["acceptMode"] != "server-merge" {
		t.Fatalf("expected server merge mode, payload=%#v", payload)
	}
	stored, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if stored.State.ServerVersion != 6 {
		t.Fatalf("expected server version 6, got %d", stored.State.ServerVersion)
	}
	status := legacyMap(legacyMap(legacyMap(legacyMap(stored.State.Database)["progress"])["asp1"])["status"])
	if got := stringValue(status["R1"]); got != "in_progress" {
		t.Fatalf("accepted progress not stored: %q", got)
	}
	_, accepted := findLegacyProposal(&stored, "p1")
	if accepted == nil || accepted.Status != "accepted" {
		t.Fatalf("proposal not accepted: %#v", accepted)
	}
}

func TestCompactAcceptResolvesSupersededProposal(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirant", "role": "aspirant", "active": true, "accountState": "active", "phone": "old"},
		},
		"progress": map[string]any{}, "programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "revision": 1.0,
	}
	proposalDB := cloneAny(baseDB).(map[string]any)
	legacyUser(proposalDB, "asp1")["phone"] = "new"
	st := legacyStore{
		State: legacyState{ServerVersion: 2, Database: baseDB, Files: []any{}},
		Proposals: []legacyProposal{
			{ID: "old-duplicate", Status: "superseded", CanonicalID: "canonical", ActorID: "asp1", ActorRole: "aspirant"},
			{ID: "canonical", Status: "pending", ActorID: "asp1", ActorRole: "aspirant", Database: proposalDB, Files: []any{}},
		},
	}
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"actorId": "admin1", "proposalId": "old-duplicate", "decision": "accept"})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/proposals/decision", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.legacyProposalDecision(rr, req)
	if rr.Code != 200 {
		t.Fatalf("accept through superseded id status=%d body=%s", rr.Code, rr.Body.String())
	}
	stored, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	_, canonical := findLegacyProposal(&stored, "canonical")
	if canonical == nil || canonical.Status != "accepted" {
		t.Fatalf("canonical proposal was not accepted: %#v", canonical)
	}
	if got := stringValue(legacyUser(stored.State.Database, "asp1")["phone"]); got != "new" {
		t.Fatalf("canonical proposal data was not applied: %q", got)
	}
}

func TestCompactAcceptMarksHistoricalProposalObsolete(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirant", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress": map[string]any{}, "programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "revision": 1.0,
	}
	st := legacyStore{
		State: legacyState{ServerVersion: 3, Database: baseDB, Files: []any{}},
		Proposals: []legacyProposal{
			{ID: "historical", Status: "pending", ActorID: "asp1", ActorRole: "aspirant", Database: nil, Files: []any{}},
		},
	}
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"actorId": "admin1", "proposalId": "historical", "decision": "accept"})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/proposals/decision", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.legacyProposalDecision(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("historical compact accept status=%d body=%s", rr.Code, rr.Body.String())
	}
	stored, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	_, proposal := findLegacyProposal(&stored, "historical")
	if proposal == nil || proposal.Status != "obsolete" {
		t.Fatalf("historical proposal was not closed as obsolete: %#v", proposal)
	}
	if stored.State.ServerVersion != 3 {
		t.Fatalf("historical proposal must not change server data version; got %d", stored.State.ServerVersion)
	}
}

func TestReviewSubmissionNotifiesAdminAndAssignedTutor(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Administrador", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "t1", "fullName": "Tutor Uno", "role": "tutor", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "role": "aspirant", "tutorId": "t1", "active": true, "accountState": "active"},
		},
		"progress":        map[string]any{"asp1": map[string]any{"status": map[string]any{"R1": "in_progress"}, "submissions": map[string]any{"R1": map[string]any{"text": "evidencia"}}}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "revision": 1.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 3, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	incoming := cloneAny(baseDB).(map[string]any)
	legacyMap(legacyMap(legacyMap(incoming["progress"])["asp1"])["status"])["R1"] = "in_review"
	body, _ := json.Marshal(map[string]any{
		"installationId": "phone1", "actorId": "asp1", "actorRole": "aspirant", "baseVersion": 3,
		"operationId": "SGL-REV-ASP1-R1", "operationType": "review", "ownerId": "asp1",
		"requirementId": "R1", "requirementTitle": "01. Requisito de prueba", "programId": "class_grouped",
		"submission": map[string]any{"text":"Primera línea\nSegunda línea\nTercera línea"}, "files": []any{},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.submitReview(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("review sync status=%d body=%s", rr.Code, rr.Body.String())
	}
	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	recipients := map[string]legacyNotification{}
	for _, n := range st.Notifications {
		if n.SourceOperationID == "SGL-REV-ASP1-R1" {
			recipients[n.UserID] = n
		}
	}
	if recipients["admin1"].Type != "review_submitted" || recipients["t1"].Type != "review_submitted" {
		t.Fatalf("admin/tutor did not receive review notification: %#v", recipients)
	}
	if recipients["admin1"].RequirementTitle != "01. Requisito de prueba" {
		t.Fatalf("notification did not preserve requirement title: %#v", recipients["admin1"])
	}
	if recipients["admin1"].OwnerID != "asp1" || recipients["t1"].OwnerID != "asp1" {
		t.Fatalf("review notification must identify the aspirant owner for realtime UI updates: %#v", recipients)
	}
}

func TestAtomicAdminReviewDecisionUpdatesServerAndNotifiesAspirant(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Administrador", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress": map[string]any{"asp1": map[string]any{
			"status":      map[string]any{"R1": "in_review"},
			"submissions": map[string]any{"R1": map[string]any{"text": "listo", "feedback": "", "approval": nil}},
		}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "revision": 7.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 11, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"actorId": "admin1", "ownerId": "asp1", "programId": "class_grouped", "requirementId": "R1",
		"requirementTitle": "Requisito específico R1", "decision": "approved", "ackOnly": true,
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/reviews/decision", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.legacyReviewDecision(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("review decision status=%d body=%s", rr.Code, rr.Body.String())
	}
	var ack map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &ack); err != nil {
		t.Fatalf("compact review ack is invalid json: %v", err)
	}
	if ack["ackOnly"] != true || ack["status"] != "approved" || ack["database"] != nil || stringValue(ack["reviewedAt"]) == "" {
		t.Fatalf("review ack was not compact/current: %#v", ack)
	}
	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if st.State.ServerVersion != 12 {
		t.Fatalf("server version must increment atomically, got %d", st.State.ServerVersion)
	}
	progress := legacyProgressForRequirement(st.State.Database, "asp1", "class_grouped", false)
	if got := stringValue(mapField(progress, "status")["R1"]); got != "approved" {
		t.Fatalf("requirement was not approved on server: %q", got)
	}
	sub := legacyMap(mapField(progress, "submissions")["R1"])
	approval := legacyMap(sub["approval"])
	if stringValue(approval["reviewerId"]) != "admin1" || stringValue(approval["reviewerName"]) != "Administrador" {
		t.Fatalf("approval metadata missing: %#v", approval)
	}
	if len(st.Notifications) != 1 || st.Notifications[0].UserID != "asp1" || st.Notifications[0].Type != "approved" || st.Notifications[0].RequirementTitle != "Requisito específico R1" {
		t.Fatalf("aspirant approval notification missing: %#v", st.Notifications)
	}
	// A retry is idempotent: it must not bump version or create another notification.
	req2 := httptest.NewRequest(http.MethodPost, "http://example/api/reviews/decision", bytes.NewReader(body))
	req2.Header.Set("Authorization", "Bearer adm")
	rr2 := httptest.NewRecorder()
	s.legacyReviewDecision(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("idempotent review status=%d body=%s", rr2.Code, rr2.Body.String())
	}
	var retryAck map[string]any
	if err := json.Unmarshal(rr2.Body.Bytes(), &retryAck); err != nil || retryAck["ackOnly"] != true || retryAck["idempotent"] != true || retryAck["database"] != nil {
		t.Fatalf("idempotent compact ack invalid: %#v err=%v", retryAck, err)
	}
	st2, _ := s.loadLegacy()
	if st2.State.ServerVersion != 12 || len(st2.Notifications) != 1 {
		t.Fatalf("idempotent retry changed state: version=%d notices=%d", st2.State.ServerVersion, len(st2.Notifications))
	}
}

func TestNotificationReadPersistsForClientUser(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{"users": []any{map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"}}}
	st := legacyStore{State: legacyState{ServerVersion: 1, Database: db, Files: []any{}}, Notifications: []legacyNotification{{ID: "n1", UserID: "asp1", Type: "approved", Title: "Aprobado", Message: "R1", CreatedAt: "2026-08-31T12:00:00Z"}}}
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"userId": "asp1", "notificationIds": []string{"n1"}})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/notifications/read", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacyNotificationRead(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("notification read status=%d body=%s", rr.Code, rr.Body.String())
	}
	stored, _ := s.loadLegacy()
	if stored.Notifications[0].ReadAt == "" {
		t.Fatal("notification readAt was not persisted")
	}
}

func TestSyncDiagnosticsValidatesTokenAndActor(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{"users": []any{
		map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"},
		map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"},
	}}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 9, UpdatedAt: "2026-09-01T03:00:00Z", Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}

	unauthorized := httptest.NewRecorder()
	s.legacySyncDiagnostics(unauthorized, httptest.NewRequest(http.MethodGet, "http://example/api/sync/diagnostics?actorId=asp1", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d body=%s", unauthorized.Code, unauthorized.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "http://example/api/sync/diagnostics?actorId=asp1", nil)
	req.Header.Set("Authorization", "Bearer cli")
	rr := httptest.NewRecorder()
	s.legacySyncDiagnostics(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("diagnostics status=%d body=%s", rr.Code, rr.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["actorFound"] != true || payload["actorActive"] != true || payload["actorRole"] != "aspirant" {
		t.Fatalf("unexpected actor diagnostics: %#v", payload)
	}
	if payload["compatVersion"] != "5.6" || int64(payload["serverVersion"].(float64)) != 9 {
		t.Fatalf("unexpected version diagnostics: %#v", payload)
	}
}

func TestAdminInitialPublishAcceptsExistingLocalRevision(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	admin := map[string]any{"id": "admin1", "fullName": "Admin", "role": "admin", "active": true, "accountState": "active"}
	db := map[string]any{
		"version": "5.0.7", "revision": 23.0,
		"users": []any{admin}, "clubs": []any{}, "classes": []any{}, "modules": []any{},
		"progress": map[string]any{}, "programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"settings": map[string]any{}, "audit": []any{},
	}
	body, _ := json.Marshal(map[string]any{
		"installationId": "adminpc", "actorId": "admin1", "actorRole": "admin", "baseVersion": 23,
		"database": db, "files": []any{},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/sync", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.legacySync(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("initial publish status=%d body=%s", rr.Code, rr.Body.String())
	}
	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if st.State.Database == nil || st.State.ServerVersion != 1 {
		t.Fatalf("empty server was not initialized: version=%d database=%#v", st.State.ServerVersion, st.State.Database)
	}
	if len(st.Proposals) != 0 {
		t.Fatalf("initial admin publish must not create proposals: %#v", st.Proposals)
	}
}

func TestAdminRejectsUploadedDocumentAsRequirementDelivery(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Administrador", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress": map[string]any{"asp1": map[string]any{
			"status": map[string]any{"R-DOC": "in_review"},
			"submissions": map[string]any{"R-DOC": map[string]any{
				"text": "", "richText": "", "attachmentIds": []any{"doc1"},
			}},
		}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "revision": 3.0,
	}
	files := []any{map[string]any{
		"id": "doc1", "ownerId": "asp1", "programId": "class_grouped", "requirementId": "R-DOC",
		"name": "entrega.pdf", "kind": "document", "type": "application/pdf", "dataUrl": "data:application/pdf;base64,JVBERi0xLjQ=",
	}}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 8, Database: baseDB, Files: files}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"actorId": "admin1", "ownerId": "asp1", "programId": "class_grouped", "requirementId": "R-DOC",
		"requirementTitle": "Requisito", "decision": "approved",
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/reviews/decision", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.legacyReviewDecision(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("document-only review status=%d body=%s", rr.Code, rr.Body.String())
	}
	stored, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	progress := legacyProgressForRequirement(stored.State.Database, "asp1", "class_grouped", false)
	if got := stringValue(mapField(progress, "status")["R-DOC"]); got != "in_review" {
		t.Fatalf("rejected approval must preserve historical status: %q", got)
	}
 if len(stored.State.Files)!=1 {t.Fatal("historical evidence was deleted")}
}

func TestDirectReviewRejectsCompletelyEmptyInProgressDraft(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Administrador", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress": map[string]any{"asp1": map[string]any{
			"status":      map[string]any{"R-EMPTY": "in_progress"},
			"submissions": map[string]any{"R-EMPTY": map[string]any{"text": "", "richText": "", "attachmentIds": []any{}, "selections": []any{}}},
		}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "revision": 1.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 2, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"actorId": "admin1", "ownerId": "asp1", "programId": "class_grouped", "requirementId": "R-EMPTY", "decision": "approved",
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/reviews/decision", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.legacyReviewDecision(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("empty in-progress requirement should remain non-reviewable, status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestNeedsChangesNotificationCarriesCanonicalFeedbackForFastClientUpdate(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	baseDB := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Administrador", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"progress": map[string]any{"asp1": map[string]any{
			"status": map[string]any{"R1": "in_review"},
			"submissions": map[string]any{"R1": map[string]any{
				"text":     "Primera linea con contenido suficiente para revision.\nSegunda linea con desarrollo.\nTercera linea final.",
				"feedback": "", "approval": nil,
			}},
		}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "revision": 4.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 20, Database: baseDB, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"actorId": "admin1", "ownerId": "asp1", "programId": "class_grouped", "requirementId": "R1",
		"requirementTitle": "Requisito R1", "decision": "needs_changes", "feedback": "Agrega la fecha y explica mejor la conclusión.", "ackOnly": true,
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/reviews/decision", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer adm")
	rr := httptest.NewRecorder()
	s.legacyReviewDecision(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("needs_changes status=%d body=%s", rr.Code, rr.Body.String())
	}
	var ack map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &ack); err != nil {
		t.Fatal(err)
	}
	if ack["status"] != "needs_changes" || ack["feedback"] != "Agrega la fecha y explica mejor la conclusión." || stringValue(ack["reviewedAt"]) == "" {
		t.Fatalf("compact correction ack missing canonical metadata: %#v", ack)
	}
	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Notifications) != 1 {
		t.Fatalf("expected one correction notification, got %#v", st.Notifications)
	}
	n := st.Notifications[0]
	if n.Type != "needs_changes" || n.Feedback != "Agrega la fecha y explica mejor la conclusión." || n.ReviewedAt == "" || n.ServerVersion != 21 || n.OwnerID != "asp1" {
		t.Fatalf("correction notification missing fast-update fields: %#v", n)
	}
	progress := legacyProgressForRequirement(st.State.Database, "asp1", "class_grouped", false)
	if got := stringValue(mapField(progress, "status")["R1"]); got != "needs_changes" {
		t.Fatalf("server did not keep canonical correction state: %q", got)
	}
}
