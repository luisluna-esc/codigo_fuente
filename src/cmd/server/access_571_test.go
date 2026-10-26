package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func accessDB571() (map[string]any, map[string]any, time.Time) {
	u := map[string]any{"id": "learner", "role": "aspirant", "active": true, "createdAt": "2026-01-01T00:00:00Z"}
	db := map[string]any{"users": []any{u}, "modules": []any{map[string]any{"id": "guide_major", "active": true, "assignedUserIds": []any{"learner"}}}, "programProgress": map[string]any{}, "tasks": []any{}}
	return db, u, time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
}
func Test571NinetyPercentPresentedExemptsWeeklyAndTasks(t *testing.T) {
	db, u, now := accessDB571()
	p := reviewProgressContainer(db, "learner", "guide_major")
	ids := requirements571["guide_major"]
	n := (len(ids)*9 + 9) / 10
	for _, id := range ids[:n] {
		ensureMapField(p, "status")[id] = "in_review"
	}
	db["tasks"] = []any{map[string]any{"assigneeId": "learner", "programId": "guide_major", "requirementId": ids[len(ids)-1], "dueDate": "2026-10-01"}}
	if reasons := autoReasons571(db, u, now, ""); len(reasons) != 0 {
		t.Fatal(reasons)
	}
	delete(ensureMapField(p, "status"), ids[n-1])
	if reasons := autoReasons571(db, u, now, ""); len(reasons) != 2 {
		t.Fatalf("below90 should fail each goal: %v", reasons)
	}
}
func Test571TwoWeeklyDeliveriesAndSeventyFivePercent(t *testing.T) {
	db, u, now := accessDB571()
	p := reviewProgressContainer(db, "learner", "guide_major")
	ids := requirements571["guide_major"]
	for i, id := range ids[:4] {
		at := now.AddDate(0, 0, -2-i*3).Format(time.RFC3339)
		ensureMapField(p, "status")[id] = "in_review"
		ensureMapField(p, "submissions")[id] = map[string]any{"submittedAt": at}
		db["tasks"] = append(arrayField(db, "tasks"), map[string]any{"id": id, "assigneeId": "learner", "programId": "guide_major", "requirementId": id, "dueDate": "2026-10-01", "presentedAt": at})
	}
	if r := autoReasons571(db, u, now, ""); len(r) != 0 {
		t.Fatal(r)
	}
	delete(ensureMapField(p, "status"), ids[0])
	delete(ensureMapField(p, "submissions"), ids[0])
	if r := autoReasons571(db, u, now, ""); len(r) != 0 {
		t.Fatalf("task delivery must count: %v", r)
	}
	for _, raw := range arrayField(db, "tasks") {
		delete(legacyMap(raw), "presentedAt")
	}
	delete(ensureMapField(p, "status"), ids[1])
	if r := autoReasons571(db, u, now, ""); len(r) != 2 {
		t.Fatalf("either goal blocks independently: %v", r)
	}
}
func Test571SameTaskAndReportNotCountedTwice(t *testing.T) {
	db, u, now := accessDB571()
	p := reviewProgressContainer(db, "learner", "guide_major")
	id := requirements571["guide_major"][0]
	ensureMapField(p, "status")[id] = "in_review"
	ensureMapField(p, "submissions")[id] = map[string]any{"submittedAt": now.AddDate(0, 0, -2).Format(time.RFC3339)}
	db["tasks"] = []any{map[string]any{"assigneeId": "learner", "programId": "guide_major", "requirementId": id, "presentedAt": now.AddDate(0, 0, -8).Format(time.RFC3339)}}
	if r := autoReasons571(db, u, now, ""); len(r) != 1 {
		t.Fatal(r)
	}
}
func Test571GraceExpiryAndScope(t *testing.T) {
	db, u, now := accessDB571()
	rec := &accessRecord571{Manual: true}
	if accessStatus571(db, u, "conquis", rec, now)["canExtend"] != true {
		t.Fatal("missing grace")
	}
	rec.ExtensionUsed = true
	rec.ExtensionUntil = now.Add(time.Hour).Format(time.RFC3339)
	if accessStatus571(db, u, "conquis", rec, now)["blocked"] == true {
		t.Fatal("active grace blocked")
	}
	later := now.Add(2 * time.Hour)
	status := accessStatus571(db, u, "conquis", rec, later)
	if status["blocked"] != true || status["canExtend"] != false {
		t.Fatal(status)
	}
	u["role"] = "regional"
	if accessStatus571(db, u, "music", &accessRecord571{}, now)["blocked"] == true {
		t.Fatal("management requires no tutor")
	}
	rec.Manual = true
	rec.ExtensionUntil = later.Add(time.Hour).Format(time.RFC3339)
	if accessStatus571(db, u, "conquis", rec, now)["blocked"] != false {
		t.Fatal("manual grace should enable access")
	}
}
func accessCall571(t *testing.T, s *Server, token, scope, action string, ids []string) *httptest.ResponseRecorder {
	t.Helper()
	method := "GET"
	var b bytes.Buffer
	if action != "" {
		method = "POST"
		json.NewEncoder(&b).Encode(map[string]any{"actorId": token, "action": action, "ids": ids})
	}
	p := "class_grouped"
	if scope == "music" {
		p = musicProgram553
	}
	r := httptest.NewRequest(method, "/api/access571?actorId="+token+"&programId="+p+"&manage=1", &b)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	s.access571(w, r)
	return w
}
func Test571OnlyManagersCanBlockAndInactivationIsScoped(t *testing.T) {
	s, _ := fixture550(t)
	for _, role := range []string{"admin", "departamental", "secretaria", "coordinador_musica"} {
		if w := accessCall571(t, s, role, "music", "block", []string{"p"}); w.Code != 200 {
			t.Fatal(role, w.Code, w.Body.String())
		}
	}
	for _, role := range []string{"regional", "tutor", "administrativo"} {
		if w := accessCall571(t, s, role, "music", "block", []string{"p"}); w.Code != 403 {
			t.Fatal(role, w.Code)
		}
	}
	if w := accessCall571(t, s, "admin", "music", "disable", []string{"p"}); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	data, e := s.loadAccess571()
	if e != nil || !data["music"]["p"].Disabled || data["conquis"]["p"] != nil {
		t.Fatal(data, e)
	}
	if w := accessCall571(t, s, "admin", "music", "enable", []string{"p"}); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	data, _ = s.loadAccess571()
	if data["music"]["p"].Disabled {
		t.Fatal("enable not persisted")
	}
}
func Test571ServerDeniesBlockedSessionOutsideSaveWindow(t *testing.T) {
	s, _ := fixture550(t)
	data := accessFile571{}
	rec := accessEntry571(data, "music", "p")
	rec.Manual = true
	rec.BlockedAt = time.Now().Add(-time.Hour).Format(time.RFC3339)
	s.saveAccess571(data)
	r := httptest.NewRequest("GET", "/api/music/module", nil)
	r.Header.Set("Authorization", "Bearer p")
	r.Header.Set("X-SGL-Program", musicProgram553)
	w := httptest.NewRecorder()
	called := false
	s.accessGuard571(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })).ServeHTTP(w, r)
	if called || w.Code != 423 {
		t.Fatal(called, w.Code, w.Body.String())
	}
}
func Test571MusicResetRoleAndAssignedTutorBoundaries(t *testing.T) {
	s, _ := fixture550(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	for _, role := range []string{"admin", "departamental", "secretaria", "coordinador_musica"} {
		if !canResetMusic571(db, role, role, "p") {
			t.Fatal(role)
		}
	}
	if canResetMusic571(db, "regional", "regional", "p") {
		t.Fatal("same-region scope is not a tutor assignment")
	}
	legacyUser(db, "p")["tutorByProgram"] = map[string]any{musicProgram553: "regional"}
	if !canResetMusic571(db, "regional", "regional", "p") {
		t.Fatal("assigned regional denied")
	}
	if canResetMusic571(db, "administrativo", "administrativo", "p") {
		t.Fatal("unassigned administrative role received global reset")
	}
}
func Test571NoTutorBlocksOnlyMusicLearnerEntry(t *testing.T) {
	db, u, now := accessDB571()
	status := accessStatus571(db, u, "music", &accessRecord571{}, now)
	if status["blocked"] != true || !strings.Contains(strings.Join(status["reasons"].([]string), " "), "tutor") {
		t.Fatal(status)
	}
	u["role"] = "secretaria"
	if accessStatus571(db, u, "music", &accessRecord571{}, now)["blocked"] == true {
		t.Fatal("management blocked")
	}
}
func Test571ExpandedMusicResetThroughEndpoint(t *testing.T) {
	for _, role := range []string{"admin", "departamental", "secretaria", "coordinador_musica", "regional", "tutor"} {
		t.Run(role, func(t *testing.T) {
			s := resetFixture557(t)
			st, _ := s.loadLegacy()
			if role == "regional" {
				legacyUser(st.State.Database, "p")["tutorByProgram"] = map[string]any{musicProgram553: role}
				s.saveLegacy(st)
			}
			body := resetBody557(role, "p")
			body["requirementId"] = "LM-TEM-1"
			w := reset557(t, s, role, body)
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			st, _ = s.loadLegacy()
			if legacyRequirementStatus(st.State.Database, "p", musicProgram553, "LM-TEM-1") != "pending" {
				t.Fatal("not deleted")
			}
		})
	}
}
