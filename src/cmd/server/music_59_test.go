package main

import "testing"

func Test59MusicAuthoritiesAvailableBeforeReceivingAspirants(t *testing.T) {
	s, _ := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		u["tutorByProgram"] = map[string]any{}
		delete(u, "tutorId")
	}
	for _, id := range []string{"admin", "departamental", "secretaria", "coordinador", "administrativo", "regional", "tm"} {
		u := legacyUser(db, id)
		if !musicTutor560(db, u) {
			t.Fatal("missing automatic tutor", id)
		}
		if !musicSeeTutor560(db, legacyUser(db, "admin"), u) {
			t.Fatal("missing directory entry", id)
		}
	}
	for _, id := range []string{"a", "tg"} {
		if musicTutor560(db, legacyUser(db, id)) {
			t.Fatal("unrelated tutor", id)
		}
	}
	u := legacyUser(db, "regional")
	u["managementProgramIds"] = []any{"guide_major"}
	if musicTutor560(db, u) {
		t.Fatal("Conquistadores-only authority leaked")
	}
	u["managementProgramIds"] = []any{musicProgram553}
	u["active"] = false
	if musicTutor560(db, u) {
		t.Fatal("inactive tutor")
	}
}

func Test59AssignSeveralAspirantsAndRejectSelfTutor(t *testing.T) {
	s, h := excelFixture560(t)
	_, login := musicLogin553(t, h, "admin", true)
	token := stringValue(login["sessionToken"])
	for _, owner := range []string{"a", "b", "regional"} {
		st, _ := s.loadLegacy()
		u := legacyUser(st.State.Database, owner)
		code, out := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", token, map[string]any{"actorId": "admin", "ownerId": owner, "action": "assign_tutor", "restoreGeneration": 1, "expectedUpdatedAt": u["updatedAt"], "value": map[string]any{"tutorId": "regional", "region": "1"}})
		if owner == "regional" {
			if code != 400 {
				t.Fatal("self assignment accepted", code, out)
			}
		} else if code != 200 {
			t.Fatal(code, out)
		}
	}
	st, _ := s.loadLegacy()
	for _, id := range []string{"a", "b"} {
		if !legacyTutorAssigned(legacyUser(st.State.Database, id), "regional", musicProgram553) {
			t.Fatal("lost assignment", id)
		}
	}
	if legacyRole(legacyUser(st.State.Database, "regional")) != "regional" {
		t.Fatal("role changed")
	}
	db := legacyMap(st.State.Database)
	patch := map[string]any{"users": []any{map[string]any{"id": "regional", "tutorByProgram": map[string]any{musicProgram553: "regional"}}}}
	if musicAssignmentPatch561(db, legacyUser(db, "admin"), patch) == "" {
		t.Fatal("admin sync bypassed self tutor protection")
	}
}
