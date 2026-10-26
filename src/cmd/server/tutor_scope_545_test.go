package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fixture545() map[string]any {
	return map[string]any{"users": []any{
		map[string]any{"id": "admin", "role": "admin", "active": true},
		map[string]any{"id": "music", "role": "tutor", "active": true, "fullName": "Tutor Música"},
		map[string]any{"id": "grouped", "role": "tutor", "active": true},
		map[string]any{"id": "a", "role": "aspirant", "active": true, "tutorId": "old", "tutorByProgram": map[string]any{"music_ministry": "music", "class_grouped": "grouped"}},
		map[string]any{"id": "other", "role": "aspirant", "active": true, "tutorId": "another"}},
		"progress":        map[string]any{"a": map[string]any{"status": map[string]any{"R": "in_review"}, "submissions": map[string]any{"R": map[string]any{"text": "grouped"}}}},
		"programProgress": map[string]any{"a": map[string]any{"music_ministry": map[string]any{"status": map[string]any{"M": "in_review"}, "submissions": map[string]any{"M": map[string]any{"text": "music"}}}}}}
}
func Test545TutorAssignments(t *testing.T) {
	owner := legacyUser(fixture545(), "a")
	for _, tc := range []struct {
		id, p string
		want  bool
	}{{"music", "music_ministry", true}, {"music", "class_grouped", false}, {"grouped", "class_grouped", true}, {"grouped", "music_ministry", false}, {"old", "", false}, {"music", "", true}} {
		if got := legacyTutorAssigned(owner, tc.id, tc.p); got != tc.want {
			t.Fatalf("%+v got %v", tc, got)
		}
	}
	owner["tutorByProgram"] = map[string]any{}
	if legacyTutorAssigned(owner, "old", "") {
		t.Fatal("explicit unassignment revived legacy tutor")
	}
	delete(owner, "tutorByProgram")
	if !legacyTutorAssigned(owner, "old", "music_ministry") {
		t.Fatal("legacy tutor lost")
	}
}
func Test545SnapshotAndReviewScope(t *testing.T) {
	db := fixture545()
	s := &Server{cfg: Config{AdminToken: "admin-token", ClientToken: "client-token"}, path: filepath.Join(t.TempDir(), "data.json")}
	st := legacyStore{State: legacyState{ServerVersion: 1, Database: db}}
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	snapshot := s.filteredLegacySnapshotMode(st.State, "music", "tutor", false)
	filtered := legacyMap(snapshot["database"])
	if legacyMap(filtered["progress"])["a"] != nil {
		t.Fatal("music tutor received grouped progress")
	}
	if legacyMap(legacyMap(filtered["programProgress"])["a"])["music_ministry"] == nil {
		t.Fatal("music progress missing")
	}
	for _, tc := range []struct {
		p, r   string
		status int
	}{{"class_grouped", "R", 403}, {"music_ministry", "M", 200}} {
		q := url.Values{"actorId": {"music"}, "ownerId": {"a"}, "programId": {tc.p}, "requirementId": {tc.r}}
		req := httptest.NewRequest("GET", "/api/reviews/item?"+q.Encode(), nil)
		req.Header.Set("X-SGL-Token", "client-token")
		rr := httptest.NewRecorder()
		s.legacyReviewItem(rr, req)
		if rr.Code != tc.status {
			t.Fatalf("%s got %d %s", tc.p, rr.Code, rr.Body.String())
		}
	}
}
func Test545PresenceScopedOnServer(t *testing.T) {
	now := time.Now().UTC()
	s := &Server{cfg: Config{AdminToken: "admin-token"}, path: filepath.Join(t.TempDir(), "data.json"), remoteSessions: map[string]remoteSession{}}
	if err := s.saveLegacy(legacyStore{State: legacyState{Database: fixture545()}}); err != nil {
		t.Fatal(err)
	}
	for _, user := range []struct{ id, role string }{{"music", "tutor"}, {"grouped", "tutor"}, {"a", "aspirant"}, {"other", "aspirant"}} {
		token := hashToken(user.id)
		s.remoteSessions[token] = remoteSession{TokenHash: token, UserID: user.id, Role: user.role, HeartbeatAt: now.Format(time.RFC3339), ExpiresAt: now.Add(time.Hour).Format(time.RFC3339)}
	}
	for _, tc := range []struct {
		token string
		count int
	}{{"music", 1}, {"admin-token", 4}} {
		req := httptest.NewRequest(http.MethodGet, "/api/remote/presence", nil)
		req.Header.Set("Authorization", "Bearer "+tc.token)
		rr := httptest.NewRecorder()
		s.remotePresence(rr, req)
		var p struct {
			Count int
			Users []struct{ UserID string }
		}
		json.Unmarshal(rr.Body.Bytes(), &p)
		if rr.Code != 200 || p.Count != tc.count {
			t.Fatalf("%s: %d %s", tc.token, rr.Code, rr.Body.String())
		}
		if tc.token == "music" && p.Users[0].UserID != "a" {
			t.Fatal("unassigned presence leaked")
		}
	}
}
func Test545AssignmentNoticeOnlyOwnerAndIdempotent(t *testing.T) {
	db := fixture545()
	before := cloneAny(db)
	legacyUser(before, "a")["tutorByProgram"] = map[string]any{"class_grouped": "grouped"}
	st := legacyStore{State: legacyState{Database: db}}
	actor := legacyUser(db, "admin")
	addTutorAssignmentNotifications(&st, before, actor, "op", "2026-09-28T00:00:00Z")
	addTutorAssignmentNotifications(&st, before, actor, "op", "2026-09-28T00:00:00Z")
	if len(st.Notifications) != 1 || st.Notifications[0].UserID != "a" || st.Notifications[0].ProgramID != "music_ministry" {
		t.Fatalf("incorrect recipients: %#v", st.Notifications)
	}
	ids := activeLegacyReviewerIDs(db, "a", "music_ministry")
	if len(ids) != 2 || ids[0] != "admin" || ids[1] != "music" {
		t.Fatalf("review recipients: %v", ids)
	}
}

func Test545OwnReviewPollIsCanonicalAndPrivate(t *testing.T) {
	db := fixture545()
	mapField(legacyMap(mapField(db, "progress")["a"]), "status")["R"] = "needs_changes"
	s := &Server{cfg: Config{ClientToken: "client-token"}, path: filepath.Join(t.TempDir(), "state.json")}
	s.saveLegacy(legacyStore{State: legacyState{Database: db}})
	for _, actor := range []string{"a", "music"} {
		body := `{"actorId":"` + actor + `","items":[{"programId":"class_grouped","requirementId":"R","status":"in_review"}]}`
		req := httptest.NewRequest("POST", "/api/reviews/own-changes", strings.NewReader(body))
		req.Header.Set("X-SGL-Token", "client-token")
		rr := httptest.NewRecorder()
		s.legacyOwnReviewChanges(rr, req)
		if actor == "music" {
			if rr.Code != 403 {
				t.Fatal("tutor read self-only poll")
			}
			continue
		}
		var payload map[string]any
		json.Unmarshal(rr.Body.Bytes(), &payload)
		changes := arrayField(payload, "changes")
		if rr.Code != 200 || len(changes) != 1 || stringValue(legacyMap(changes[0])["status"]) != "needs_changes" || stringValue(payload["ownerId"]) != "a" {
			t.Fatal(rr.Body.String())
		}
	}
}

func Test546OwnPollAcknowledgesOnlyCurrentActorsOperations(t *testing.T) {
	s := &Server{cfg: Config{ClientToken: "client-token"}, path: filepath.Join(t.TempDir(), "state.json")}
	st := legacyStore{State: legacyState{Database: fixture545()}}
	rememberAppliedLegacyOperation(&st, legacyAppliedOperation{OperationID: "mine", ActorID: "a", OperationType: "review", ServerVersion: 8})
	rememberAppliedLegacyOperation(&st, legacyAppliedOperation{OperationID: "foreign", ActorID: "other", OperationType: "review", ServerVersion: 8})
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/api/reviews/own-changes", strings.NewReader(`{"actorId":"a","operationIds":["mine","foreign","missing"],"items":[]}`))
	req.Header.Set("X-SGL-Token", "client-token")
	rr := httptest.NewRecorder()
	s.legacyOwnReviewChanges(rr, req)
	var payload map[string]any
	json.Unmarshal(rr.Body.Bytes(), &payload)
	ids := arrayField(payload, "acknowledgedOperationIds")
	if rr.Code != 200 || len(ids) != 1 || stringValue(ids[0]) != "mine" {
		t.Fatal(rr.Body.String())
	}
}
