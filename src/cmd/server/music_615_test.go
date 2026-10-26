package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Display suppression must not erase an assigned review or its five-day report.
func Test615SeniorRolesKeepFollowupWithoutLoginReminder(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	owner := legacyUser(db, "p")
	p := reviewProgressContainer(db, "p", musicProgram553)
	ensureMapField(p, "status")["LM-PRE-1"] = "in_review"
	ensureMapField(p, "submissions")["LM-PRE-1"] = map[string]any{"submittedAt": time.Now().Add(-121 * time.Hour).UTC().Format(time.RFC3339Nano)}
	for _, role := range []string{"admin", "departamental", "secretaria", "coordinador_musica", "administrativo", "regional", "tutor"} {
		t.Run(role, func(t *testing.T) {
			owner["tutorByProgram"] = map[string]any{musicProgram553: role}
			actor := legacyUser(db, role)
			view := musicFollowupView610(db, actor)
			expected := !musicLeaders610(role)
			if view["showReminder"] != expected {
				t.Fatalf("reminder display policy for %s: %v", role, view)
			}
			if musicReview560(db, actor, owner) && numericValue(view["pendingCount"]) != 1 {
				t.Fatal("assigned review was lost")
			}
			if musicLeaders610(role) && len(view["rows"].([]any)) != 1 {
				t.Fatal("management report was lost")
			}
		})
	}
	st, _ := s.loadLegacy()
	st.State.Database = db
	st.Notifications = nil
	musicFollowupNotifications610(&st)
	for _, role := range []string{"admin", "departamental", "secretaria", "coordinador_musica"} {
		found := false
		for _, item := range st.Notifications {
			if item.UserID == role && item.Type == "revision610" {
				found = true
			}
		}
		if !found {
			t.Fatalf("five-day report missing for %s", role)
		}
	}
}

// Personal training needs the same read-only delivery modes as any aspirant.
func Test615TutorPersonalViewReadsReportModesWithoutWriteAccess(t *testing.T) {
	s, db := fixture550(t)
	module := legacyMap(arrayField(db, "modules")[0])
	module["assignedUserIds"] = append(arrayField(module, "assignedUserIds"), "tutor")
	db["reportRequirementModes"] = map[string]any{"music_ministry:LM-PRE-1": map[string]any{"mode": "document"}}
	st, _ := s.loadLegacy()
	st.State.Database = db
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	handler := s.personalAccess555(http.HandlerFunc(s.reportConfig5079))
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		r := httptest.NewRequest(method, "/api/report-config", nil)
		r.Header.Set("Authorization", "Bearer tutor")
		r.Header.Set("X-SGL-Personal", "1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if method == http.MethodPost {
			if w.Code != 403 {
				t.Fatalf("personal write was not blocked: %d", w.Code)
			}
			continue
		}
		if w.Code != 200 {
			t.Fatalf("read-only modes: %d %s", w.Code, w.Body.String())
		}
		var payload map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		if len(legacyMap(payload["reportRequirementModes"])) != 1 || payload["database"] != nil || payload["files"] != nil {
			t.Fatal("incorrect or excessive report configuration", payload)
		}
	}
}
