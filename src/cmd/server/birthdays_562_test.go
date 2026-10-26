package main

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test562BirthdayScopeAndDayOnlyException(t *testing.T) {
	users := []any{}
	for _, row := range []struct{ id, region, date string }{{"one", "1", "1980-10-10"}, {"two", "2", "1981-10-11"}, {"today", "2", "1990-10-05"}, {"assigned", "2", "1992-10-12"}, {"revoked", "2", "1985-10-15"}} {
		users = append(users, map[string]any{"id": row.id, "role": "aspirant", "fullName": row.id, "region": row.region, "birthDate": row.date})
	}
	legacyMap(users[3])["tutorByProgram"] = map[string]any{"music_ministry": "t"}
	legacyMap(users[4])["tutorId"] = "t"
	legacyMap(users[4])["tutorByProgram"] = map[string]any{"music_ministry": ""}
	db := map[string]any{"music550": map[string]any{"permissions602": map[string]any{"roles": map[string]any{"aspirant": map[string]any{"birthdays.all": false}, "tutor": map[string]any{"birthdays.all": false}, "regional": map[string]any{"birthdays.all": false}}}}, "users": users, "modules": []any{map[string]any{"id": musicProgram553, "assignedUserIds": []any{"one", "two", "today", "assigned", "revoked"}}}}
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		actor map[string]any
		count int
	}{
		{map[string]any{"id": "a", "role": "admin"}, 5},
		{map[string]any{"id": "d", "role": "departamental"}, 5},
		{map[string]any{"id": "s", "role": "secretaria"}, 5},
		{map[string]any{"id": "g", "role": "coordinador_musica"}, 5},
		{map[string]any{"id": "o", "role": "administrativo"}, 5},
		{map[string]any{"id": "r", "role": "regional", "region": "1", "regionalAssignedUserIds": []any{"two"}}, 3},
		{map[string]any{"id": "t", "role": "tutor"}, 2},
		{map[string]any{"id": "learner", "role": "aspirant"}, 1},
	} {
		rows := birthdayDirectory562(db, tc.actor, "music", now)
		if len(rows) != tc.count {
			t.Fatalf("%v: %v", tc.actor, rows)
		}
		for _, raw := range rows {
			r := legacyMap(raw)
			if len(r) != 3 || !strings.HasPrefix(stringValue(r["birthDate"]), "2000-") {
				t.Fatal("private fields leaked", r)
			}
		}
	}
	if len(birthdayDirectory562(db, map[string]any{"id": "learner", "role": "aspirant"}, "music", now.AddDate(0, 0, 1))) != 0 {
		t.Fatal("public birthday persisted beyond today")
	}
	if len(birthdayDirectory562(db, map[string]any{"id": "a", "role": "admin"}, "conquistadores", now)) != 0 {
		t.Fatal("music leaked into Conquistadores")
	}
}
func Test562BirthdayEndpointRequiresActorAndRejectsForgedStaff(t *testing.T) {
	s, _ := recoveryFixture548(t)
	for _, q := range []string{"", "?actorId=admin"} {
		r := httptest.NewRequest("GET", "/api/birthdays"+q, nil)
		r.Header.Set("X-SGL-Token", "cli")
		w := httptest.NewRecorder()
		s.birthdays548(w, r)
		if w.Code != 403 {
			t.Fatalf("%s: %d", q, w.Code)
		}
	}
}
