package main

import "testing"

func Test591AnnualChoiceCanBeClearedWithoutLosingReadingsAndReviewLocks(t *testing.T) {
	for _, status := range []string{"in_progress", "needs_changes", "in_review", "approved"} {
		t.Run(status, func(t *testing.T) {
			s := annualFixture551(t, "reavivados")
			st, _ := s.loadLegacy()
			db := legacyMap(st.State.Database)
			p := reviewProgressContainer(db, "p", musicProgram553)
			ensureMapField(p, "status")["LM-NUT-1"] = status
			old := legacyMap(mapField(p, "submissions")["LM-NUT-1"])
			ensureMapField(old, "bibleReadings")["entries"] = map[string]any{"2026-10-01": map[string]any{"impactStory": "Guardado"}}
			incoming := legacyMap(cloneAny(db))
			sp := reviewProgressContainer(incoming, "p", musicProgram553)
			ensureMapField(legacyMap(mapField(sp, "submissions")["LM-NUT-1"]), "bibleReadings")["choice"] = ""
			merged, _, err := mergeLegacyProposalOnServer(db, &legacyProposal{ActorID: "p", ActorRole: "pastor_aspirant", Database: incoming, OperationType: "auto_sync"})
			if err != nil {
				t.Fatal(err)
			}
			result := legacyMap(merged)
			expected := ""
			if status == "in_review" || status == "approved" {
				expected = "reavivados"
			}
			if musicAnnualChoice551(result, "p") != expected {
				t.Fatalf("choice after %s: %s", status, musicAnnualChoice551(result, "p"))
			}
			if len(mapField(musicDevotion550(result, "p", "LM-NUT-1"), "entries")) != 1 {
				t.Fatal("reading lost")
			}
		})
	}
}
func Test591AnnualCertificateDoesNotDependOnChoiceAndKeepsTask(t *testing.T) {
	s := annualFixture551(t, "reavivados")
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	db["tasks"] = []any{map[string]any{"id": "linked", "assigneeId": "p", "programId": musicProgram553, "requirementId": "LM-NUT-1", "active": true, "dueDate": "2026-12-01"}}
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	sub, file := annualDocument558("bible_plan")
	if w := submitPDF558(t, s, sub, []any{file}); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	tasks := arrayField(db, "tasks")
	if len(tasks) != 1 {
		t.Fatal("task removed or duplicated")
	}
	task := legacyMap(tasks[0])
	if stringValue(task["id"]) != "linked" || stringValue(task["dueDate"]) != "2026-12-01" || stringValue(task["reviewSubmittedAt"]) == "" || stringValue(task["presentedBy"]) != "p" {
		t.Fatal("task synchronization missing")
	}
	if musicComplete550(db, "p") || musicEligible550(db, "p", "final", "") {
		t.Fatal("sent PDF incorrectly completed investiture")
	}
}
func Test591CertificateRequiresAllReviewedRequirements(t *testing.T) {
	s, _ := fixture550(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	p := reviewProgressContainer(db, "p", musicProgram553)
	for _, id := range musicRequirements550 {
		ensureMapField(p, "status")[id] = "approved"
	}
	ensureMapField(musicData550(db), "workflow")["p"] = map[string]any{"state": "recommended"}
	if musicEligible550(db, "p", "final", "") {
		t.Fatal("leader certificate must wait for the departmental approval date")
	}
	ensureMapField(musicData550(db), "workflow")["p"] = map[string]any{"state": "approved", "approvedAt": "2026-10-20T12:30:00Z"}
	if !musicEligible550(db, "p", "final", "") {
		t.Fatal("reviewed plus final approval rejected")
	}
	for _, status := range []string{"pending", "assigned", "in_progress", "in_review", "needs_changes"} {
		ensureMapField(p, "status")[musicRequirements550[0]] = status
		if musicEligible550(db, "p", "final", "") {
			t.Fatalf("%s counted as reviewed", status)
		}
	}
}
