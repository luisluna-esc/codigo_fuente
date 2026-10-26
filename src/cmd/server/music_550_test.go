package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func fixture550(t *testing.T) (*Server, map[string]any) {
	s, db := fixture548R3(t)
	for _, role := range []string{"admin", "departamental", "secretaria", "coordinador_musica", "administrativo", "regional", "tutor"} {
		db["users"] = append(arrayField(db, "users"), map[string]any{"id": role, "role": role, "fullName": role, "region": "1", "active": true})
		s.remoteSessions[hashToken(role)] = remoteSession{UserID: role, Role: role, ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339)}
	}
	p := legacyUser(db, "p")
	legacyUser(db, "regional")["regionalAssignedUserIds"] = []any{"p"}
	p["region"] = "1"
	p["tutorId"] = "tutor"
	p["church"] = "Prueba"
	p["mission"] = "Misión"
	s.remoteSessions[hashToken("p")] = remoteSession{UserID: "p", Role: "pastor_aspirant", ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339)}
	db["modules"] = []any{map[string]any{"id": "music_ministry", "active": true, "assignedUserIds": []any{"p"}}}
	testEnrollConquistadores593(db)
	pstate := reviewProgressContainer(db, "p", "music_ministry")
	for _, id := range musicRequirements550 {
		ensureMapField(pstate, "status")[id] = "approved"
		ensureMapField(pstate, "submissions")[id] = map[string]any{"text": "Informe", "updatedAt": "before"}
	}
	st, _ := s.loadLegacy()
	st.State.Database = db
	s.saveLegacy(st)
	return s, db
}
func call550(t *testing.T, s *Server, actor, owner, action string, value map[string]any, stale bool) *httptest.ResponseRecorder {
	t.Helper()
	st, _ := s.loadLegacy()
	base := stringValue(musicData550(legacyMap(st.State.Database))["updatedAt"])
	if stale {
		base = "stale"
	}
	body, _ := json.Marshal(map[string]any{"actorId": actor, "ownerId": owner, "action": action, "value": value, "restoreGeneration": 1, "baseUpdatedAt": base})
	r := httptest.NewRequest("POST", "/api/music/module", bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+actor)
	w := httptest.NewRecorder()
	s.securityHeaders(http.HandlerFunc(s.music550)).ServeHTTP(w, r)
	return w
}
func Test550RoleMatrixAndWorkflow(t *testing.T) {
	for _, role := range []string{"admin", "departamental", "secretaria", "coordinador_musica", "administrativo", "regional", "tutor", "p"} {
		t.Run(role, func(t *testing.T) {
			s, _ := fixtureRole560(t, role)
			w := call550(t, s, role, "p", "recommend", nil, false)
			allowed := musicRecommend550(role)
			if allowed && w.Code != 200 {
				t.Fatalf("recommend %d %s", w.Code, w.Body.String())
			}
			if !allowed && w.Code != 403 {
				t.Fatalf("recommend permission %d", w.Code)
			}
			s, _ = fixtureRole560(t, role)
			call550(t, s, "secretaria", "p", "recommend", nil, false)
			w = call550(t, s, role, "p", "approve", nil, false)
			allowed = musicApprove550(role)
			if allowed && w.Code != 200 {
				t.Fatalf("approve %d %s", w.Code, w.Body.String())
			}
			if !allowed && w.Code != 403 {
				t.Fatalf("approve permission %d", w.Code)
			}
		})
	}
	s, _ := fixture550(t)
	if w := call550(t, s, "departamental", "p", "approve", nil, false); w.Code != 400 {
		t.Fatal("approved without recommendation")
	}
	if w := call550(t, s, "secretaria", "p", "recommend", nil, true); w.Code != 409 {
		t.Fatal("stale overwrite")
	}
	st, _ := s.loadLegacy()
	ensureMapField(reviewProgressContainer(legacyMap(st.State.Database), "p", "music_ministry"), "status")["LM-PRE-1"] = "in_progress"
	s.saveLegacy(st)
	if w := call550(t, s, "secretaria", "p", "recommend", nil, false); w.Code != 400 {
		t.Fatal("incomplete recommendation")
	}
}
func Test550SingleRequirementCorrectionAndImmutability(t *testing.T) {
	s, _ := fixtureRole560(t, "regional")
	call550(t, s, "secretaria", "p", "recommend", nil, false)
	call550(t, s, "departamental", "p", "approve", nil, false)
	w := call550(t, s, "regional", "p", "correct", map[string]any{"requirementId": "LM-PRE-1", "feedback": "Añade evidencia"}, false)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, _ := s.loadLegacy()
	p := legacyProgressForRequirement(legacyMap(st.State.Database), "p", "music_ministry", false)
	for _, id := range musicRequirements550 {
		expected := "approved"
		if id == "LM-PRE-1" {
			expected = "needs_changes"
		}
		if stringValue(mapField(p, "status")[id]) != expected {
			t.Fatal("correction reopened unrelated requirement", id)
		}
	}
	if musicEligible550(legacyMap(st.State.Database), "p", "final", "") {
		t.Fatal("final cert survived correction")
	}
	frozen := map[string]any{"status": map[string]any{"R": "approved"}, "submissions": map[string]any{"R": map[string]any{"text": "Original"}}}
	changed := map[string]any{"status": map[string]any{"R": "approved"}, "submissions": map[string]any{"R": map[string]any{"text": "Alterado"}}}
	out := mergeAspirantProgressState(frozen, changed)
	if stringValue(legacyMap(mapField(out, "submissions")["R"])["text"]) != "Original" {
		t.Fatal("reviewed text mutated")
	}
}
func Test550EventsCertificatesAndPublicQR(t *testing.T) {
	s, _ := fixture550(t)
	v := map[string]any{"name": "Formación", "date": "2026-01-01", "time": "09:00", "place": "Auditorio", "category": "leadership", "hours": 10}
	if w := call550(t, s, "p", "", "event", v, false); w.Code != 403 {
		t.Fatal("aspirant created event")
	}
	w := call550(t, s, "admin", "", "event", v, false)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, _ := s.loadLegacy()
	var eid string
	for id := range mapField(musicData550(legacyMap(st.State.Database)), "events") {
		eid = id
	}
	call550(t, s, "secretaria", "p", "enroll", map[string]any{"eventId": eid}, false)
	if w := call550(t, s, "secretaria", "p", "enroll", map[string]any{"eventId": eid}, false); w.Code != 400 {
		t.Fatal("duplicate enrollment")
	}
	cv := map[string]any{"kind": "leadership", "eventId": eid}
	if w := call550(t, s, "p", "p", "certificate", cv, false); w.Code != 400 {
		t.Fatal("cert without attendance")
	}
	call550(t, s, "admin", "p", "attendance", map[string]any{"eventId": eid, "attended": true}, false)
	w = call550(t, s, "p", "p", "certificate", cv, false)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, _ = s.loadLegacy()
	c := legacyMap(mapField(musicData550(legacyMap(st.State.Database)), "certificates")["p:leadership:"+eid])
	r := httptest.NewRequest("GET", "/api/music/certificate?token="+stringValue(c["token"]), nil)
	qr := httptest.NewRecorder()
	s.musicCertificate550(qr, r)
	if qr.Code != 200 || !strings.Contains(qr.Body.String(), "Pastor") || strings.Contains(qr.Body.String(), "password") || strings.Contains(qr.Body.String(), "Informe") {
		t.Fatal("invalid QR data")
	}
	call550(t, s, "admin", "p", "attendance", map[string]any{"eventId": eid, "attended": false}, false)
	qr = httptest.NewRecorder()
	s.musicCertificate550(qr, r)
	if !strings.Contains(qr.Body.String(), "Pendiente de nueva validación") {
		t.Fatal("QR remained valid after attendance removed")
	}
}

func Test550EventDeletionRoleMatrix(t *testing.T) {
	for _, tc := range []struct {
		role    string
		allowed bool
	}{{"admin", true}, {"departamental", true}, {"secretaria", true}, {"coordinador_musica", true}, {"administrativo", false}, {"regional", false}, {"tutor", false}, {"p", false}} {
		t.Run(tc.role, func(t *testing.T) {
			s, _ := fixture550(t)
			created := call550(t, s, "admin", "", "event", map[string]any{"name": "Ensayo", "date": "2026-10-06", "time": "10:00", "place": "Iglesia", "category": "music_ministry"}, false)
			if created.Code != 200 {
				t.Fatalf("create event: %d %s", created.Code, created.Body.String())
			}
			st, _ := s.loadLegacy()
			var id string
			for eventID := range mapField(musicData550(legacyMap(st.State.Database)), "events") {
				id = eventID
			}
			w := call550(t, s, tc.role, "", "delete_event", map[string]any{"eventId": id}, false)
			if tc.allowed && w.Code != 200 {
				t.Fatalf("authorized deletion: %d %s", w.Code, w.Body.String())
			}
			if !tc.allowed && w.Code != 403 {
				t.Fatalf("unauthorized deletion returned %d", w.Code)
			}
			st, _ = s.loadLegacy()
			exists := mapField(musicData550(legacyMap(st.State.Database)), "events")[id] != nil
			if exists == tc.allowed {
				t.Fatalf("event existence after deletion = %v, allowed = %v", exists, tc.allowed)
			}
		})
	}
}

func Test550SystemAndExternalEventParticipants(t *testing.T) {
	s, db := fixture550(t)
	db["users"] = append(arrayField(db, "users"), map[string]any{"id": "guest_account", "role": "tutor", "fullName": "Tutor de Música", "active": true, "tutorProgramIds": []any{"music_ministry"}})
	db["users"] = append(arrayField(db, "users"), map[string]any{"id": "outside", "role": "aspirant", "fullName": "Solo Conquistadores", "active": true})
	st, _ := s.loadLegacy()
	st.State.Database = db
	s.saveLegacy(st)
	created := call550(t, s, "admin", "", "event", map[string]any{"name": "Encuentro", "date": "2026-10-06", "time": "10:00", "place": "Iglesia", "category": "music_ministry"}, false)
	if created.Code != http.StatusOK {
		t.Fatalf("create event: %d %s", created.Code, created.Body.String())
	}
	st, _ = s.loadLegacy()
	var eventID string
	for id := range mapField(musicData550(legacyMap(st.State.Database)), "events") {
		eventID = id
	}
	if res := call550(t, s, "admin", "outside", "enroll", map[string]any{"eventId": eventID}, false); res.Code != http.StatusForbidden {
		t.Fatalf("non-Music enrollment allowed: %d", res.Code)
	}
	if res := call550(t, s, "admin", "guest_account", "enroll", map[string]any{"eventId": eventID}, false); res.Code != http.StatusOK {
		t.Fatalf("enroll registered system account: %d %s", res.Code, res.Body.String())
	}
	if res := call550(t, s, "admin", "", "external_enroll", map[string]any{"eventId": eventID, "fullName": "  Ana   Fuera del Sistema "}, false); res.Code != http.StatusOK {
		t.Fatalf("enroll external guest: %d %s", res.Code, res.Body.String())
	}
	st, _ = s.loadLegacy()
	event := legacyMap(mapField(musicData550(legacyMap(st.State.Database)), "events")[eventID])
	if len(mapField(event, "participants")) != 1 || len(mapField(event, "externalParticipants")) != 1 {
		t.Fatalf("participant lists were not stored separately: %#v", event)
	}
	for id := range mapField(event, "externalParticipants") {
		if res := call550(t, s, "admin", "", "external_attendance", map[string]any{"eventId": eventID, "participantId": id, "attended": true}, false); res.Code != http.StatusOK {
			t.Fatalf("mark external attendance: %d %s", res.Code, res.Body.String())
		}
	}
	if res := call550(t, s, "tutor", "", "external_enroll", map[string]any{"eventId": eventID, "fullName": "No autorizado"}, false); res.Code != http.StatusForbidden {
		t.Fatalf("non-manager registered an external guest: %d", res.Code)
	}
	if res := call550(t, s, "tutor", "guest_account", "unenroll", map[string]any{"eventId": eventID}, false); res.Code != http.StatusForbidden {
		t.Fatalf("non-manager removed participant: %d", res.Code)
	}
	if res := call550(t, s, "admin", "guest_account", "unenroll", map[string]any{"eventId": eventID}, false); res.Code != http.StatusOK {
		t.Fatalf("remove registered participant: %d %s", res.Code, res.Body.String())
	}
	st, _ = s.loadLegacy()
	event = legacyMap(mapField(musicData550(legacyMap(st.State.Database)), "events")[eventID])
	if len(mapField(event, "participants")) != 0 || len(mapField(event, "externalParticipants")) != 1 {
		t.Fatal("removal changed the wrong list")
	}

}

func Test550SystemParticipantsCanEarnEventCertificates(t *testing.T) {
	s, db := fixture550(t)
	db["users"] = append(arrayField(db, "users"), map[string]any{"id": "registered", "role": "tutor", "fullName": "Tutor registrado", "active": true, "tutorProgramIds": []any{"music_ministry"}})
	st, _ := s.loadLegacy()
	st.State.Database = db
	s.saveLegacy(st)
	if res := call550(t, s, "admin", "", "event", map[string]any{"name": "Liderazgo", "date": "2026-01-01", "time": "10:00", "place": "Iglesia", "category": "leadership", "hours": 10}, false); res.Code != http.StatusOK {
		t.Fatalf("create event: %d %s", res.Code, res.Body.String())
	}
	st, _ = s.loadLegacy()
	var eventID string
	for id := range mapField(musicData550(legacyMap(st.State.Database)), "events") {
		eventID = id
	}
	if res := call550(t, s, "admin", "registered", "enroll", map[string]any{"eventId": eventID}, false); res.Code != http.StatusOK {
		t.Fatalf("enroll system user: %d %s", res.Code, res.Body.String())
	}
	if res := call550(t, s, "admin", "registered", "attendance", map[string]any{"eventId": eventID, "attended": true}, false); res.Code != http.StatusOK {
		t.Fatalf("mark system user attendance: %d %s", res.Code, res.Body.String())
	}
	if res := call550(t, s, "admin", "registered", "certificates", nil, false); res.Code != http.StatusOK {
		t.Fatalf("issue event certificate to registered system user: %d %s", res.Code, res.Body.String())
	}
	st, _ = s.loadLegacy()
	certs := mapField(musicData550(legacyMap(st.State.Database)), "certificates")
	if legacyMap(certs["registered:leadership:"+eventID]) == nil {
		t.Fatal("certificate for the attended system participant was not issued")
	}
	if res := call550(t, s, "admin", "registered", "certificate", map[string]any{"kind": "leadership", "eventId": eventID}, false); res.Code != http.StatusOK {
		t.Fatalf("individual certificate for Music tutor: %d %s", res.Code, res.Body.String())
	}
	if res := call550(t, s, "admin", "registered", "unenroll", map[string]any{"eventId": eventID}, false); res.Code != http.StatusOK {
		t.Fatalf("remove participant: %d", res.Code)
	}
	st, _ = s.loadLegacy()
	if musicEligible550(legacyMap(st.State.Database), "registered", "leadership", eventID) {
		t.Fatal("removed participant still eligible for event certificate")
	}
}

func Test550FortyRecordsAndManualConclusion(t *testing.T) {
	s, db := fixture550(t)
	p := reviewProgressContainer(db, "p", "music_ministry")
	entries := map[string]any{}
	for i := 1; i <= 40; i++ {
		entries[strconv.Itoa(i)] = map[string]any{"completed": true, "date": time.Date(2026, 1, i, 0, 0, 0, 0, time.UTC).Format("2006-01-02"), "reference": "Salmos 1", "theme": "Adoración", "favoriteVerses": "Salmos 1:1", "impactStory": "Mensaje", "lifeLesson": "Aplicación", "songTheme": "Himno y relación"}
	}
	ensureMapField(p, "submissions")["LM-TEM-1"] = map[string]any{"bibleReadings": map[string]any{"entries": entries}}
	if !musicFortyComplete550(db, "p") {
		t.Fatal("40 complete blocked")
	}
	delete(entries, "40")
	if musicFortyComplete550(db, "p") {
		t.Fatal("39 readings accepted")
	}
	st, _ := s.loadLegacy()
	db = legacyMap(st.State.Database)
	ensureMapField(reviewProgressContainer(db, "p", "music_ministry"), "status")["LM-NUT-1"] = "in_progress"
	ensureMapField(reviewProgressContainer(db, "p", "music_ministry"), "submissions")["LM-NUT-1"] = map[string]any{"bibleReadings": map[string]any{"kind": "music_devotion", "choice": "bible_plan"}}
	st.State.Database = db
	s.saveLegacy(st)
	if w := call550(t, s, "regional", "p", "conclude", map[string]any{"kind": "bible_plan"}, false); w.Code != 403 {
		t.Fatal("regional concluded")
	}
	if w := call550(t, s, "secretaria", "p", "conclude", map[string]any{"kind": "bible_plan"}, false); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	if stringValue(mapField(legacyProgressForRequirement(db, "p", "music_ministry", false), "status")["LM-NUT-1"]) != "in_progress" {
		t.Fatal("conclusion must wait for the aspirant to submit")
	}
	if !musicEligible550(db, "p", "bible_plan", "") {
		t.Fatal("Bible Plan certificate blocked")
	}
}
func Test550ForgedActorAndAssignmentScope(t *testing.T) {
	s, _ := fixture550(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	legacyUser(db, "regional")["regionalAssignedUserIds"] = []any{}
	legacyUser(db, "regional")["region"] = "2"
	st.State.Database = db
	s.saveLegacy(st)
	if w := call550(t, s, "regional", "p", "recommend", nil, false); w.Code != 403 {
		t.Fatal("regional crossed scope")
	}
	r := httptest.NewRequest("POST", "/api/music/module", bytes.NewBufferString(`{"actorId":"admin","action":"event"}`))
	r.Header.Set("Authorization", "Bearer p")
	w := httptest.NewRecorder()
	s.music550(w, r)
	if w.Code != 403 {
		t.Fatal("forged actor")
	}
}

func Test550SnapshotDoesNotExposeOtherCertificates(t *testing.T) {
	s, db := fixture550(t)
	data := musicData550(db)
	data["certificates"] = map[string]any{
		"p:bible_plan:":     map[string]any{"ownerId": "p", "kind": "bible_plan", "token": "own"},
		"other:bible_plan:": map[string]any{"ownerId": "other", "kind": "bible_plan", "token": "private"},
	}
	data["workflow"] = map[string]any{"p": map[string]any{"state": "recommended"}, "other": map[string]any{"state": "approved"}}
	legacyMap(mapField(reviewProgressContainer(db, "p", musicProgram553), "submissions")["LM-NUT-1"])["bibleReadings"] = map[string]any{"choice": "bible_plan"}
	data["completions"] = map[string]any{"p": map[string]any{"bible_plan": map[string]any{"concluded": true}}}
	st, _ := s.loadLegacy()
	st.State.Database = db
	result := s.filteredLegacySnapshotMode(st.State, "p", "pastor_aspirant", false)
	visible := mapField(mapField(result, "database"), "music550")
	if len(mapField(visible, "certificates")) != 1 || mapField(visible, "certificates")["p:bible_plan:"] == nil {
		t.Fatal("snapshot leaked another certificate or hid own certificate")
	}
	if len(mapField(visible, "workflow")) != 1 {
		t.Fatal("snapshot leaked another workflow")
	}
	if len(mapField(musicData550(db), "certificates")) != 2 {
		t.Fatal("filter mutated canonical data")
	}
}
func Test550RejectsFakeTemplateImage(t *testing.T) {
	s, _ := fixture550(t)
	w := call550(t, s, "admin", "", "template", map[string]any{"kind": "final", "dataUrl": "data:image/png;base64,aW52YWxpZA==", "width": 2000, "height": 1000}, false)
	if w.Code != 400 {
		t.Fatal("accepted fake image dimensions", w.Code)
	}
}

func Test550AdminSnapshotCannotEraseMusicResource(t *testing.T) {
	s, _ := fixture550(t)
	s.cfg.AdminToken = "adm-token"
	if w := call550(t, s, "secretaria", "p", "recommend", nil, false); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	st, _ := s.loadLegacy()
	oldClient := legacyMap(cloneAny(st.State.Database))
	delete(oldClient, "music550")
	body, _ := json.Marshal(map[string]any{"installationId": "adminpc", "actorId": "admin", "actorRole": "admin", "baseVersion": st.State.ServerVersion, "restoreGeneration": 1, "database": oldClient, "files": []any{}})
	r := httptest.NewRequest("POST", "/api/sync", bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+s.cfg.AdminToken)
	w := httptest.NewRecorder()
	s.legacySync(w, r)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	after, _ := s.loadLegacy()
	if stringValue(legacyMap(mapField(musicData550(legacyMap(after.State.Database)), "workflow")["p"])["state"]) != "recommended" {
		t.Fatal("ordinary admin snapshot erased independent music data")
	}
}

func fixtureRole560(t *testing.T, role string) (*Server, map[string]any) {
	s, db := fixture550(t)
	if role == "administrativo" || role == "regional" {
		legacyUser(db, "p")["tutorByProgram"] = map[string]any{musicProgram553: role}
		st, _ := s.loadLegacy()
		st.State.Database = db
		s.saveLegacy(st)
	}
	return s, db
}
