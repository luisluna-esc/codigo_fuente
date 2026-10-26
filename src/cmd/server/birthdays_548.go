package main

import (
	"net/http"
	"time"
)

// Birthday lists contain month/day only. Regional/tutor scope is enforced here,
// including the exception that today's Music birthdays are visible to everyone.
func birthdayMember562(db, u map[string]any, music bool) bool {
	if music && u["musicAccessRemoved"] == true {
		return false
	}
	for _, p := range []string{"class_grouped", "guide_major", "guide_master", "guide_master_advanced", "music_ministry"} {
		if (p == musicProgram553) != music {
			continue
		}
		if assignedProgram555(db, stringValue(u["id"]), p) || managementProgram555(db, u, p) {
			return true
		}
	}
	return false
}
func birthdayScope562(db, actor, owner map[string]any) bool {
	switch legacyRole(actor) {
	case "admin", "departamental", "secretaria", "coordinador_musica", "administrativo":
		return true
	case "regional":
		if musicSameRegion560(actor, owner) || containsString(arrayField(actor, "regionalAssignedUserIds"), stringValue(owner["id"])) {
			return true
		}
		for _, p := range []string{"class_grouped", "guide_major", "guide_master", "guide_master_advanced", "music_ministry"} {
			if stringValue(legacyMap(owner["regionalByProgram"])[p]) == stringValue(actor["id"]) || legacyTutorAssigned(owner, stringValue(actor["id"]), p) {
				return true
			}
		}
	case "tutor":
		for _, p := range []string{"class_grouped", "guide_major", "guide_master", "guide_master_advanced", "music_ministry"} {
			if legacyTutorAssigned(owner, stringValue(actor["id"]), p) {
				return true
			}
		}
	}
	return stringValue(actor["id"]) == stringValue(owner["id"])
}
func birthdayDirectory562(db, actor map[string]any, view string, now time.Time) []any {
	people := []any{}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if !legacyActive(u) {
			continue
		}
		date, err := time.Parse("2006-01-02", stringValue(u["birthDate"]))
		if err != nil {
			continue
		}
		music := birthdayMember562(db, u, true)
		if view == "music" && !music {
			continue
		}
		if view == "conquistadores" && !birthdayMember562(db, u, false) {
			continue
		}
		today := date.Month() == now.Month() && date.Day() == now.Day()
		if !birthdayScope562(db, actor, u) && !(music && today) && !(view == "music" && music && musicPermit602(db, actor, "birthdays.all")) {
			continue
		}
		people = append(people, map[string]any{"id": u["id"], "fullName": u["fullName"], "birthDate": "2000-" + date.Format("01-02")})
	}
	return people
}
func (s *Server) birthdays548(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", 405)
		return
	}
	if _, ok := s.auth(r); !ok {
		legacyError(w, 401, "Inicia sesión para ver los cumpleaños.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	id, role, ok := s.accountActor556(r, r.URL.Query().Get("actorId"), &st)
	if !ok {
		legacyError(w, 403, "Sesión no autorizada.", nil)
		return
	}
	db := legacyMap(st.State.Database)
	actor := legacyMap(cloneAny(legacyUser(db, id)))
	actor["role"] = role
	if r.URL.Query().Get("view") == "music" {
        if !birthdayMember562(db, actor, true) { legacyError(w, 403, "No tienes acceso a Música.", nil); return }
        calendar := r.URL.Query().Get("mode") != "today"
        if calendar && !musicPermit602(db, actor, "birthdays.calendar") || !calendar && !musicPermit602(db, actor, "birthdays.auto") && !musicPermit602(db, actor, "birthdays.reminders") { legacyError(w, 403, "Cumpleaños está desactivado.", nil); return }
    }
    // Bolivia's civil day, independent of the server machine's timezone.
	now := time.Now().In(time.FixedZone("Bolivia", -4*60*60))
	people := birthdayDirectory562(db, actor, r.URL.Query().Get("view"), now)
    if r.URL.Query().Get("view") == "music" && r.URL.Query().Get("mode") == "today" {
        dates := map[string]bool{now.Format("01-02"): true}
        if musicPermit602(db, actor, "birthdays.reminders") { dates[now.AddDate(0,0,4).Format("01-02")] = true; dates[now.AddDate(0,0,7).Format("01-02")] = true }
        filtered := []any{}
        for _, raw := range people { row := legacyMap(raw); date := stringValue(row["birthDate"]); if len(date) == 10 && dates[date[5:]] { filtered = append(filtered, row) } }
        people = filtered
    }
    jsonOut(w, 200, map[string]any{"people": people})
}
