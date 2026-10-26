package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func call556(t *testing.T, h http.Handler, method, path, token, program string, body any) (int, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(body)
	r := httptest.NewRequest(method, path, bytes.NewReader(b))
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-SGL-Program", program)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	out := map[string]any{}
	json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func Test556TutorCreationRolesScopeAndIdentity(t *testing.T) {
	for _, role := range []string{"departamental", "administrativo", "secretaria"} {
		t.Run(role, func(t *testing.T) {
			s, h := setup555(t)
			st, _ := s.loadLegacy()
			legacyUser(st.State.Database, role)["managementProgramIds"] = []any{musicProgram553}
			s.saveLegacy(st)
			_, login := musicLogin553(t, h, role, true)
			token := stringValue(login["sessionToken"])
			body := map[string]any{"actorId": role, "id": "new-tutor", "programId": musicProgram553, "programIds": []any{musicProgram553}, "restoreGeneration": 1, "user": map[string]any{"fullName": "Tutor nuevo", "email": "tutor@example.test", "username": "tutor-nuevo", "region": "12", "password": tutorPassword556(), "role": "admin", "managementProgramIds": []any{"guide_major"}}}
			code, result := call556(t, h, "POST", musicAPI553+"/api/accounts/tutor", token, musicProgram553, body)
			if code != 200 {
				t.Fatal(code, result)
			}
			st, _ = s.loadLegacy()
			u := legacyUser(st.State.Database, "new-tutor")
			if legacyRole(u) != "tutor" || stringValue(u["region"]) != "12" || u["managementProgramIds"] != nil || !containsString(arrayField(u, "tutorProgramIds"), musicProgram553) {
				t.Fatal("forged privileges or missing tutor", u)
			}
			code, _ = call556(t, h, "POST", musicAPI553+"/api/accounts/tutor", token, musicProgram553, body)
			if code != 200 {
				t.Fatal("lost ACK retry failed", code)
			}
			body["id"] = "duplicate-email"
			if code, _ = call556(t, h, "POST", musicAPI553+"/api/accounts/tutor", token, musicProgram553, body); code != 409 {
				t.Fatal("duplicate identity", code)
			}
			body["programIds"] = []any{"guide_major"}
			if code, _ = call556(t, h, "POST", musicAPI553+"/api/accounts/tutor", token, musicProgram553, body); code != 403 {
				t.Fatal("foreign module", code)
			}
			body["programIds"] = []any{musicProgram553}
			body["user"].(map[string]any)["email"] = "different@example.test"
			body["user"].(map[string]any)["region"] = "13"
			if code, _ = call556(t, h, "POST", musicAPI553+"/api/accounts/tutor", token, musicProgram553, body); code != 400 {
				t.Fatal("invalid region", code)
			}
		})
	}
}

func Test556RosterReflectsNewAssignmentAndRevocationWithoutLogin(t *testing.T) {
	s, h := setup555(t)
	_, login := musicLogin553(t, h, "tm", true)
	token := stringValue(login["sessionToken"])
	path := musicAPI553 + "/api/accounts/roster?actorId=tm"
	code, snap := call556(t, h, "GET", path, token, musicProgram553, nil)
	if code != 200 || legacyUser(snap["database"], "b") != nil {
		t.Fatal("unassigned owner exposed", code)
	}
	hash := stringValue(snap["rosterHash"])
	if code, out := call556(t, h, "GET", path+"&knownHash="+hash, token, musicProgram553, nil); code != 200 || out["unchanged"] != true {
		t.Fatal("roster unnecessarily downloaded", out)
	}
	st, _ := s.loadLegacy()
	legacyUser(st.State.Database, "b")["tutorByProgram"] = map[string]any{musicProgram553: "tm"}
	s.saveLegacy(st)
	code, snap = call556(t, h, "GET", path+"&knownHash="+hash, token, musicProgram553, nil)
	if code != 200 || legacyUser(snap["database"], "b") == nil || stringValue(snap["rosterHash"]) == hash {
		t.Fatal("new aspirant did not appear", code, snap)
	}
	encoded, _ := json.Marshal(snap)
	if strings.Contains(string(encoded), "password") || strings.Contains(string(encoded), "guide_major") {
		t.Fatal("roster leaked credentials/module")
	}
	st, _ = s.loadLegacy()
	legacyUser(st.State.Database, "b")["tutorByProgram"] = map[string]any{musicProgram553: nil}
	s.saveLegacy(st)
	_, snap = call556(t, h, "GET", path+"&knownHash="+stringValue(snap["rosterHash"]), token, musicProgram553, nil)
	if legacyUser(snap["database"], "b") != nil {
		t.Fatal("revoked aspirant remains accessible")
	}
}

func Test556MusicAuthoritiesReviewRegionalWithoutSelfApproval(t *testing.T) {
	for _, role := range []string{"administrativo", "departamental"} {
		t.Run(role, func(t *testing.T) {
			s, h := setup555(t)
			st, _ := s.loadLegacy()
			db := legacyMap(st.State.Database)
			regional := legacyUser(db, "coordinador")
			regional["role"] = "regional"
			regional["tutorByProgram"] = map[string]any{musicProgram553: role}
			p := reviewProgressContainer(db, "coordinador", musicProgram553)
			ensureMapField(p, "status")["LM-PRE-1"] = "in_review"
			ensureMapField(p, "submissions")["LM-PRE-1"] = map[string]any{"text": "Un informe de prueba completo.\nOtra línea con evidencia real.\nUna tercera línea para su revisión.", "submissionMode": "written"}
			s.saveLegacy(st)
			_, login := musicLogin553(t, h, role, true)
			token := stringValue(login["sessionToken"])
			code, out := call556(t, h, "POST", musicAPI553+"/api/reviews/decision", token, musicProgram553, map[string]any{"actorId": role, "ownerId": "coordinador", "programId": musicProgram553, "requirementId": "LM-PRE-1", "decision": "approved", "ackOnly": true})
			if code != 200 {
				t.Fatal("authority could not review regional", code, out)
			}
			st, _ = s.loadLegacy()
			if !legacyTutorAssigned(legacyUser(st.State.Database, "coordinador"), role, musicProgram553) {
				t.Fatal("institutional reviewer assignment lost")
			}
			if canReview556(db, regional, regional, musicProgram553) || canReview556(db, legacyUser(db, role), regional, "guide_major") {
				t.Fatal("approval scope widened")
			}
		})
	}
}

func Test556MusicApprovalGatesInvestitureAndNotifiesDepartment(t *testing.T) {
	s, db := fixture550(t)
	v := map[string]any{"place": "Iglesia", "date": "2026-12-01", "time": "10:00", "investedBy": "Departamental"}
	if w := call550(t, s, "p", "p", "investiture", v, false); w.Code != 400 {
		t.Fatal("100% reviewed enabled investiture")
	}
	if w := call550(t, s, "secretaria", "p", "recommend", nil, false); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	st, _ := s.loadLegacy()
	found := false
	for _, n := range st.Notifications {
		if n.UserID == "departamental" && n.Type == "music_recommend" && n.ProgramID == musicProgram553 {
			found = true
		}
	}
	if !found {
		t.Fatal("Departamental not notified")
	}
	if w := call550(t, s, "p", "p", "investiture", v, false); w.Code != 400 {
		t.Fatal("recommendation enabled investiture")
	}
	if w := call550(t, s, "secretaria", "p", "approve", nil, false); w.Code != 403 {
		t.Fatal("wrong final approver")
	}
	if w := call550(t, s, "departamental", "p", "approve", nil, false); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	if w := call550(t, s, "p", "p", "investiture", v, false); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	progress := reviewProgressContainer(db, "p", musicProgram553)
	ensureMapField(progress, "submissions")["LM-TEM-1"] = map[string]any{"primaryDocumentId": "certificate", "submissionMode": "document", "documentMode": true}
	if !musicFortyComplete550(db, "p") {
		t.Fatal("reviewed PDF certificate rejected")
	}
	ensureMapField(progress, "status")["LM-TEM-1"] = "in_review"
	if musicFortyComplete550(db, "p") {
		t.Fatal("unreviewed certificate counted")
	}
}

func Test556AdministrativeNotificationsStayInTheirModule(t *testing.T) {
	s, h := setup555(t)
	_, login := musicLogin553(t, h, "admin", false)
	token := stringValue(login["sessionToken"])
	code, out := call556(t, h, "POST", "/api/admin/changes", token, musicProgram553, map[string]any{"actorId": "admin", "operationId": "music-profile", "programId": musicProgram553, "entityId": "a", "summary": "Cambio en Música", "patch": map[string]any{"users": []any{map[string]any{"id": "a", "phone": "123"}}}})
	if code != 200 {
		t.Fatal(code, out)
	}
	st, _ := s.loadLegacy()
	count := 0
	for _, n := range st.Notifications {
		if n.SourceOperationID == "music-profile" {
			count++
			if n.ProgramID != musicProgram553 {
				t.Fatal("unscoped change", n)
			}
		}
	}
	if count == 0 {
		t.Fatal("missing notice")
	}
	_, login = musicLogin553(t, h, "a", false)
	token = stringValue(login["sessionToken"])
	for _, p := range []string{musicProgram553, "guide_major"} {
		_, out = call556(t, h, "GET", "/api/notifications?userId=a", token, p, nil)
		found := false
		for _, raw := range anySlice548(out["notifications"]) {
			if stringValue(legacyMap(raw)["sourceOperationId"]) == "music-profile" {
				found = true
			}
		}
		if found != (p == musicProgram553) {
			t.Fatal("notification leaked or missing", p, out)
		}
	}
}

func tutorPassword556() map[string]any {
	salt := []byte("0123456789abcdef")
	return map[string]any{"algorithm": "PBKDF2-SHA256", "iterations": 150000, "salt": base64.StdEncoding.EncodeToString(salt), "hash": base64.StdEncoding.EncodeToString(pbkdf2SHA256([]byte("Prueba123!"), salt, 150000, 32))}
}

func TestBrowserFixture556(t *testing.T) {
	if os.Getenv("SGL556_BROWSER_FIXTURE") != "1" {
		t.Skip("interactive fixture")
	}
	s, h := setup555(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	for _, id := range []string{"admin", "a", "b"} {
		u := legacyUser(db, id)
		u["drive"] = map[string]any{"connected": true, "folderName": "Prueba"}
		p := reviewProgressContainer(db, id, musicProgram553)
		ensureMapField(p, "status")["LM-PRE-1"] = "approved"
		ensureMapField(p, "submissions")["LM-PRE-1"] = map[string]any{"text": "Un informe completo para verificar Drive.\nLa segunda línea de evidencia.\nLa tercera línea de reflexión.", "updatedAt": "before", "approval": map[string]any{"at": time.Now().Add(-7 * 24 * time.Hour).UTC().Format(time.RFC3339)}}
	}
	p := reviewProgressContainer(db, "b", musicProgram553)
	for _, id := range musicRequirements550 {
		ensureMapField(p, "status")[id] = "approved"
		ensureMapField(p, "submissions")[id] = map[string]any{"text": "Informe revisado", "updatedAt": "before"}
	}
	p = reviewProgressContainer(db, "a", musicProgram553)
	ensureMapField(p, "status")["LM-NUT-2"] = "in_review"
	ensureMapField(p, "submissions")["LM-NUT-2"] = map[string]any{"text": "Informe de lectura completo\nSegunda línea de aplicación\nTercera línea de reflexión", "submissionMode": "reading_report", "readingReport": map[string]any{"bookTitle": "Lectura de prueba", "authors": "Autor de prueba", "reportDate": "2026-10-04", "elements": []any{"Uno", "Dos", "Tres"}, "impactPhrase": "Una frase de prueba", "impactPage": "1", "personal": "Aplicación personal", "professional": "Aplicación profesional", "spiritual": "Aplicación espiritual", "finalReflection": "Reflexión de prueba"}, "updatedAt": "before"}
	ensureMapField(p, "status")["LM-TEM-1"] = "in_progress"
	ensureMapField(p, "submissions")["LM-TEM-1"] = map[string]any{"bibleReadings": map[string]any{"kind": "music_devotion", "choice": "forty", "entries": map[string]any{"1": map[string]any{"date": "2026-10-01", "theme": "Confianza", "reference": "Salmos 1", "favoriteVerses": "Salmos 1:1", "impactStory": "Mensaje", "lifeLesson": "Aplicación", "songTheme": "Himno de confianza", "completed": true}}}, "updatedAt": "before"}
	db["reportRequirementModes"] = map[string]any{musicProgram553 + ":LM-PRE-1": map[string]any{"mode": "document"}, musicProgram553 + ":LM-NUT-2": map[string]any{"mode": "reading_report"}}
	s.saveLegacy(st)
	ln, err := net.Listen("tcp", "127.0.0.1:8897")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	fmt.Println("SGL556_BROWSER_READY")
	http.Serve(ln, h)
}
