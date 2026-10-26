package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDailyActivityExcludesAdminAndNotifiesNonAdminsOnce(t *testing.T) {
	dir := t.TempDir()
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(dir, "data", "sgl-server-data.json")}
	db := map[string]any{
		"settings": map[string]any{"dailyRecognitionHour": 18.0},
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Admin Principal", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "role": "aspirant", "active": true, "accountState": "active"},
			map[string]any{"id": "t1", "fullName": "Tutor Uno", "role": "tutor", "active": true, "accountState": "active"},
		},
		"audit": []any{
			map[string]any{"actorId": "admin1", "actorName": "Admin Principal", "action": "Modificar planificación", "entity": "plan", "at": dailyActivityLocalFixture("2026-09-04 12:00:00")},
			map[string]any{"actorId": "asp1", "actorName": "Aspirante Uno", "action": "Modificar informe", "entity": "requirement", "at": dailyActivityLocalFixture("2026-09-04 13:00:00")},
			map[string]any{"actorId": "asp1", "actorName": "Aspirante Uno", "action": "Planificar actividad", "entity": "plan", "at": dailyActivityLocalFixture("2026-09-04 14:00:00")},
			map[string]any{"actorId": "t1", "actorName": "Tutor Uno", "action": "Inicio de sesión", "entity": "auth", "at": dailyActivityLocalFixture("2026-09-04 15:00:00")},
			map[string]any{"actorId": "t1", "actorName": "Tutor Uno", "action": "Revisar requisito", "entity": "review", "at": dailyActivityLocalFixture("2026-09-04 16:00:00")},
		},
		"progress": map[string]any{}, "programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{}, "classes": []any{}, "tasks": []any{}, "materials": []any{}, "modules": []any{}, "revision": 1.0,
	}
	st := legacyStore{State: legacyState{ServerVersion: 1, Database: db, Files: []any{}}}

	top := legacyTopDailyInteractions(db, "2026-09-04", 5)
	if len(top) != 2 {
		t.Fatalf("top count=%d: %#v", len(top), top)
	}
	if top[0].UserID != "asp1" || top[0].Count != 2 {
		t.Fatalf("top[0]=%#v", top[0])
	}
	if top[1].UserID != "t1" || top[1].Count != 1 {
		t.Fatalf("top[1]=%#v", top[1])
	}

	late := legacyStore{State: legacyState{ServerVersion: 1, Database: db, Files: []any{}}}
	if s.ensureLegacyDailyActivityNotifications(&late, time.Date(2026, 9, 4, 20, 0, 0, 0, time.Local)) || len(late.Notifications) != 0 {
		t.Fatal("recognition must not be generated outside the configured hour")
	}
	if s.ensureLegacyDailyActivityNotifications(&st, time.Date(2026, 9, 4, 17, 59, 0, 0, time.Local)) {
		t.Fatal("recognition must not be generated before the configured hour")
	}
	changed := s.ensureLegacyDailyActivityNotifications(&st, time.Date(2026, 9, 4, 18, 15, 0, 0, time.Local))
	if !changed {
		t.Fatal("expected recognition notifications to be generated during the configured hour")
	}
	changedAgain := s.ensureLegacyDailyActivityNotifications(&st, time.Date(2026, 9, 4, 18, 45, 0, 0, time.Local))
	if changedAgain {
		t.Fatal("daily recognition must be idempotent within the configured hour")
	}
	adminNotified, aspNotified, tutorNotified := false, false, false
	for _, n := range st.Notifications {
		if n.Type != "daily_activity" {
			continue
		}
		if n.SourceOperationID != "daily_activity_2026-09-04" {
			t.Fatalf("unexpected source id: %s", n.SourceOperationID)
		}
		if n.UserID == "admin1" {
			adminNotified = true
		}
		if n.UserID == "asp1" && n.Message != "" {
			aspNotified = true
		}
		if n.UserID == "t1" && n.Message != "" {
			tutorNotified = true
		}
	}
	if adminNotified || !aspNotified || !tutorNotified {
		t.Fatalf("notifications admin=%v aspirant=%v tutor=%v all=%#v", adminNotified, aspNotified, tutorNotified, st.Notifications)
	}
}

func TestDailyRecognitionFreezesPointsAfterCutoffForNextPeriod(t *testing.T) {
	db := map[string]any{
		"settings": map[string]any{"dailyRecognitionHour": 18.0},
		"users": []any{
			map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "role": "aspirant", "active": true, "accountState": "active"},
		},
		"audit": []any{
			map[string]any{"actorId": "asp1", "action": "Avance guardado", "entity": "R1", "at": dailyActivityLocalFixture("2026-09-04 17:30:00")},
			map[string]any{"actorId": "asp1", "action": "Avance guardado", "entity": "R2", "at": dailyActivityLocalFixture("2026-09-04 18:30:00")},
			map[string]any{"actorId": "asp1", "action": "Requisito enviado", "entity": "R3", "at": dailyActivityLocalFixture("2026-09-05 10:00:00")},
		},
	}
	day4 := legacyDailyActivityItems(db, "2026-09-04")
	if len(day4) != 1 || day4[0].Count != 1 || day4[0].Points != 5 {
		t.Fatalf("period 2026-09-04 must stop at cutoff: %#v", day4)
	}
	day5 := legacyDailyActivityItems(db, "2026-09-05")
	if len(day5) != 1 || day5[0].Count != 2 || day5[0].Points != 11 {
		t.Fatalf("points after cutoff must accumulate for next period: %#v", day5)
	}
}

func TestDailyRecognitionUnreadNoticeExpiresOutsideConfiguredHour(t *testing.T) {
	db := map[string]any{"settings": map[string]any{"dailyRecognitionHour": 18.0}}
	st := legacyStore{State: legacyState{Database: db}, Notifications: []legacyNotification{
		{ID: "daily-current", UserID: "asp1", Type: "daily_activity", SourceOperationID: "daily_activity_2026-09-04", CreatedAt: "2026-09-04T18:05:00Z"},
		{ID: "normal", UserID: "asp1", Type: "approved", CreatedAt: "2026-09-04T18:05:00Z"},
	}}
	if expireLegacyDailyActivityNotifications(&st, time.Date(2026, 9, 4, 18, 30, 0, 0, time.Local)) {
		t.Fatal("current daily recognition must remain available during its configured hour")
	}
	if st.Notifications[0].ReadAt != "" {
		t.Fatal("daily recognition was marked read inside the configured hour")
	}
	if !expireLegacyDailyActivityNotifications(&st, time.Date(2026, 9, 4, 19, 0, 0, 0, time.Local)) {
		t.Fatal("daily recognition should expire once the configured hour ends")
	}
	if st.Notifications[0].ReadAt == "" {
		t.Fatal("expired daily recognition must be marked read so it cannot pop up again")
	}
	if st.Notifications[1].ReadAt != "" {
		t.Fatal("non-daily notifications must not be changed by recognition expiry")
	}
}

// Activity periods are evaluated in the server's local time. Keep these fixtures
// in that same zone instead of assuming the test machine uses UTC.
func dailyActivityLocalFixture(value string) string {
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
	if err != nil {
		panic(err)
	}
	return parsed.Format(time.RFC3339)
}
