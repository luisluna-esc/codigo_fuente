package main

import "testing"

func Test604DepartmentAndSecretaryTutorScopeAndRemoval(t *testing.T) {
	for _, actorID := range []string{"departamental", "secretaria"} {
		t.Run(actorID, func(t *testing.T) {
			s, _ := excelFixture560(t)
			st, _ := s.loadLegacy()
			db := legacyMap(st.State.Database)
			actor := legacyUser(db, actorID)
			for _, tutorID := range []string{"tm", "other-tm"} {
				if !musicSeeTutor560(db, actor, legacyUser(db, tutorID)) {
					t.Fatal("tutor not visible", tutorID)
				}
			}
			for _, ownerID := range []string{"a", "b", "c", "d"} {
				if !musicScope550(db, actor, legacyUser(db, ownerID)) {
					t.Fatal("aspirant not visible", ownerID)
				}
			}
			for _, ownerID := range []string{"tm", "other-tm", "a"} {
				if !musicDelete560(db, actor, legacyUser(db, ownerID)) {
					t.Fatal("removal denied", ownerID)
				}
			}
			musicData550(db)["permissions602"] = map[string]any{"users": map[string]any{actorID: map[string]any{"people.remove": false}}}
			if musicDelete560(db, actor, legacyUser(db, "a")) {
				t.Fatal("disabled removal allowed")
			}
		})
	}
}
