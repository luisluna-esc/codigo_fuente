package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func annualFixture551(t *testing.T, choice string) *Server {
	t.Helper()
	s, _ := fixture550(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	p := reviewProgressContainer(db, "p", "music_ministry")
	ensureMapField(p, "status")["LM-NUT-1"] = "in_progress"
	ensureMapField(p, "submissions")["LM-NUT-1"] = map[string]any{
		"text":          "Primera reflexión\nSegunda reflexión\nTercera reflexión",
		"bibleReadings": map[string]any{"kind": "music_devotion", "choice": choice, "entries": map[string]any{}},
	}
	st.State.Database = db
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	return s
}

func submitAnnual551(t *testing.T, s *Server, choice string) *httptest.ResponseRecorder {
	t.Helper()
	sub := map[string]any{"text": "Primera reflexión\nSegunda reflexión\nTercera reflexión", "attachmentIds": []any{}, "bibleReadings": map[string]any{"kind": "music_devotion", "choice": choice}}
	body, _ := json.Marshal(legacyReviewSubmitRequest{ActorID: "p", OwnerID: "p", OperationID: "annual-551", ProgramID: "music_ministry", RequirementID: "LM-NUT-1", Submission: sub})
	r := httptest.NewRequest(http.MethodPost, "/api/reviews/submit", bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer p")
	w := httptest.NewRecorder()
	s.securityHeaders(http.HandlerFunc(s.submitReview)).ServeHTTP(w, r)
	return w
}

func Test551AnnualMustBeConcludedBeforeAspirantSends(t *testing.T) {
	for _, choice := range []string{"reavivados", "bible_plan"} {
		t.Run(choice, func(t *testing.T) {
			s := annualFixture551(t, choice)
			if w := submitAnnual551(t, s, choice); w.Code != 422 {
				t.Fatalf("submitted without conclusion: %d %s", w.Code, w.Body.String())
			}
			other := "reavivados"
			if choice == other {
				other = "bible_plan"
			}
			if w := call550(t, s, "secretaria", "p", "conclude", map[string]any{"kind": other}, false); w.Code != 400 {
				t.Fatal("inactive modality concluded", w.Code)
			}
			if w := call550(t, s, "p", "p", "conclude", map[string]any{"kind": choice}, false); w.Code != 403 {
				t.Fatal("aspirant self-validated")
			}
			if w := call550(t, s, "secretaria", "p", "conclude", map[string]any{"kind": choice}, false); w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			st, _ := s.loadLegacy()
			db := legacyMap(st.State.Database)
			if stringValue(mapField(legacyProgressForRequirement(db, "p", "music_ministry", false), "status")["LM-NUT-1"]) != "in_progress" {
				t.Fatal("conclusion automatically submitted")
			}
			if w := submitAnnual551(t, s, other); w.Code != 422 {
				t.Fatal("changed choice submitted", w.Code)
			}
			if w := submitAnnual551(t, s, choice); w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			st, _ = s.loadLegacy()
			db = legacyMap(st.State.Database)
			if stringValue(mapField(legacyProgressForRequirement(db, "p", "music_ministry", false), "status")["LM-NUT-1"]) != "in_review" {
				t.Fatal("validated review was not submitted")
			}
		})
	}
}

func Test591GenericSyncCanChangeAnnualChoiceButCannotBypassConclusion(t *testing.T) {
	s := annualFixture551(t, "reavivados")
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	for _, choice := range []string{"reavivados", "bible_plan"} {
		incoming := legacyMap(cloneAny(db))
		p := reviewProgressContainer(incoming, "p", "music_ministry")
		ensureMapField(p, "status")["LM-NUT-1"] = "in_review"
		legacyMap(mapField(p, "submissions")["LM-NUT-1"])["bibleReadings"] = map[string]any{"kind": "music_devotion", "choice": choice}
        if _,_,err:=mergeLegacyProposalOnServer(db,&legacyProposal{ActorID:"p",ActorRole:"pastor_aspirant",Database:incoming,OperationType:"auto_sync"});err==nil {t.Fatal("generic sync accepted presentation outside validated channel")}
        ensureMapField(p,"status")["LM-NUT-1"]="in_progress"
		merged, _, err := mergeLegacyProposalOnServer(db, &legacyProposal{ActorID: "p", ActorRole: "pastor_aspirant", Database: incoming, OperationType: "auto_sync"})
		if err != nil {
			t.Fatal(err)
		}
		out := legacyMap(merged)
		if musicAnnualChoice551(out, "p") != choice {
			t.Fatal("sync did not preserve the new editable choice")
		}
		if stringValue(mapField(legacyProgressForRequirement(out, "p", "music_ministry", false), "status")["LM-NUT-1"]) == "in_review" {
			t.Fatal("sync bypassed conclusion")
		}
	}
}

func Test551AdministrativeEventsUseOneCategoryAndKeepCertificates(t *testing.T) {
	for _, role := range []string{"admin", "departamental", "secretaria", "coordinador_musica"} {
		t.Run(role, func(t *testing.T) {
			s, _ := fixture550(t)
			v := map[string]any{"name": "10 horas de liderazgo", "category": "music_ministry", "certificateKind": "leadership", "description": "Formación musical y liderazgo", "date": "2026-01-01", "time": "09:00", "place": "Iglesia", "hours": 10}
			if w := call550(t, s, role, "", "event", v, false); w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			st, _ := s.loadLegacy()
			db := legacyMap(st.State.Database)
			var ev map[string]any
			for _, raw := range mapField(musicData550(db), "events") {
				ev = legacyMap(raw)
			}
			if ev["category"] != "music_ministry" || ev["description"] != v["description"] || musicEventKind551(ev) != "leadership" {
				t.Fatal("event data not retained", ev)
			}
			id := stringValue(ev["id"])
			if w := call550(t, s, role, "p", "enroll", map[string]any{"eventId": id}, false); w.Code != 200 {
				t.Fatal(w.Body.String())
			}
			if w := call550(t, s, role, "p", "attendance", map[string]any{"eventId": id, "attended": true}, false); w.Code != 200 {
				t.Fatal(w.Body.String())
			}
			st, _ = s.loadLegacy()
			db = legacyMap(st.State.Database)
			if !musicEligible550(db, "p", "leadership", id) {
				t.Fatal("one category broke leadership eligibility")
			}
		})
	}
}

func Test551OtherMinistryTypeIsRequiredAndRetained(t *testing.T) {
	s, _ := fixture550(t)
	v := map[string]any{"type": "Otro", "name": "Ministerio", "church": "Iglesia", "district": "Distrito", "region": "Región", "mission": "Misión", "union": "Unión", "createdDate": "2026-01-01", "boardVote": "12", "director": "Director", "directorPhone": "12345678"}
	call := func() *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]any{"actorId": "secretaria", "ownerId": "p", "restoreGeneration": 1, "value": v})
		r := httptest.NewRequest("POST", "/api/music-registration", bytes.NewReader(body))
		r.Header.Set("Authorization", "Bearer secretaria")
		w := httptest.NewRecorder()
		s.securityHeaders(http.HandlerFunc(s.ministryResource548)).ServeHTTP(w, r)
		return w
	}
	if w := call(); w.Code != 400 {
		t.Fatal("Otro accepted without description")
	}
	v["otherType"] = "Ensamble de cuerdas"
	if w := call(); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, _ := s.loadLegacy()
	saved := legacyMap(legacyUser(st.State.Database, "p")["musicMinistryRegistration"])
	if saved["otherType"] != "Ensamble de cuerdas" {
		t.Fatal("custom ministry type not saved")
	}
}

func Test551PersonalSnapshotHidesEveryOtherProgress(t *testing.T) {
	s, db := fixture550(t)
	db["users"] = append(arrayField(db, "users"), map[string]any{"id": "other", "role": "aspirant", "active": true, "fullName": "Private other"})
	for _, program := range []string{"class_grouped", "guide_major", "guide_master", "guide_master_advanced", "music_ministry"} {
		ensureMapField(reviewProgressContainer(db, "other", program), "status")["R-private"] = "approved"
	}
	ensureMapField(db, "bibleYear")["other"] = map[string]any{"private": "hidden"}
	for _, role := range []string{"aspirant", "pastor_aspirant"} {
		result := s.filteredLegacySnapshotMode(legacyState{Database: db}, "p", role, false)
		out := mapField(result, "database")
		for _, key := range []string{"progress", "programProgress", "bibleYear"} {
			if mapField(out, key)["other"] != nil {
				t.Fatal("private progress leaked", role, key)
			}
		}
		if legacyUser(out, "other") != nil {
			t.Fatal("other aspirant exposed")
		}
	}
}

func Test551DailyActivityDoesNotExposePeersToAspirants(t *testing.T) {
	s, db := fixture550(t)
	db["audit"] = []any{
		map[string]any{"actorId": "p", "action": "Modificar informe", "entity": "requirement", "at": dailyActivityLocalFixture("2026-09-04 13:00:00")},
		map[string]any{"actorId": "tutor", "action": "Revisar requisito", "entity": "review", "at": dailyActivityLocalFixture("2026-09-04 14:00:00")},
	}
	st, _ := s.loadLegacy()
	st.State.Database = db
	s.saveLegacy(st)
	r := httptest.NewRequest("GET", "/api/daily-activity?userId=p&date=2026-09-04", nil)
	r.Header.Set("Authorization", "Bearer p")
	w := httptest.NewRecorder()
	s.securityHeaders(http.HandlerFunc(s.legacyDailyActivity)).ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), `"userId":"tutor"`) {
		t.Fatal("daily activity exposed peer's progress")
	}
	if !strings.Contains(w.Body.String(), `"userId":"p"`) {
		t.Fatal("own activity missing", w.Body.String())
	}
}
