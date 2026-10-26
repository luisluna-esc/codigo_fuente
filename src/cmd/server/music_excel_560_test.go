package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func excelFixture560(t *testing.T) (*Server, http.Handler) {
	s, h := musicFixture553(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	makeUser := func(id, role, region string) map[string]any {
		return map[string]any{"id": id, "username": id, "email": id + "@example.test", "fullName": "Prueba " + id, "role": role, "region": region, "active": true, "accountState": "active", "password": testPasswordRecord("Prueba123!")}
	}
	db["users"] = append(legacyUsers(db), makeUser("regional", "regional", "1"), makeUser("pastor", "pastor_aspirant", "MBOS"), makeUser("d", "aspirant", "2"), makeUser("e", "aspirant", "1"))
	legacyUser(db, "b")["region"] = "2"
	legacyUser(db, "b")["tutorByProgram"] = map[string]any{musicProgram553: "regional"}
	legacyUser(db, "c")["tutorByProgram"] = map[string]any{musicProgram553: "other-tm"}
	legacyUser(db, "d")["tutorByProgram"] = map[string]any{musicProgram553: "other-tm"}
	legacyUser(db, "e")["tutorByProgram"] = map[string]any{musicProgram553: "administrativo"}
	legacyUser(db, "other-tm")["region"] = "2"
	for _, id := range []string{"departamental", "secretaria", "coordinador", "administrativo", "tm", "regional", "c", "d", "e", "pastor"} {
		for _, raw := range arrayField(db, "modules") {
			m := legacyMap(raw)
			if stringValue(m["id"]) == musicProgram553 {
				m["assignedUserIds"] = append(arrayField(m, "assignedUserIds"), id)
			}
		}
		if !strings.Contains(id, "other") && !containsString([]any{"c", "d", "e", "pastor", "regional"}, id) {
			legacyUser(db, id)["tutorByProgram"] = map[string]any{musicProgram553: "coordinador"}
		}
	}
	legacyUser(db, "e")["createdBy"] = "administrativo"
	for _, id := range []string{"b", "departamental"} {
		p := reviewProgressContainer(db, id, musicProgram553)
		for _, r := range musicRequirements550 {
			ensureMapField(p, "status")[r] = "approved"
			ensureMapField(p, "submissions")[r] = map[string]any{"text": "Informe de prueba revisado"}
		}
		ensureMapField(musicData550(db), "workflow")[id] = map[string]any{"state": "recommended"}
	}
	for _, id := range []string{"a", "e"} {
		p := reviewProgressContainer(db, id, musicProgram553)
		ensureMapField(p, "status")["LM-NUT-2"] = "in_review"
		ensureMapField(p, "submissions")["LM-NUT-2"] = map[string]any{"text": "Primera línea de informe de prueba.\nSegunda línea de evidencia verificable.\nTercera línea de reflexión personal.", "updatedAt": "before", "submissionMode": "written"}
	}
	s.saveLegacy(st)
	return s, h
}

func Test560ExcelRegionalUnionAndTutorDirectory(t *testing.T) {
	s, h := excelFixture560(t)
	_, login := musicLogin553(t, h, "regional", true)
	token := stringValue(login["sessionToken"])
	code, snap := musicCall553(t, h, "GET", musicAPI553+"/api/accounts/view?actorId=regional", token, nil)
	if code != 200 {
		t.Fatal(code, snap)
	}
	db := legacyMap(snap["database"])
	for _, id := range []string{"b"} {
		if !musicAssigned550(db, id) {
			t.Fatal("visible regional member missing", id)
		}
	}
	if musicAssigned550(db, "d") || musicAssigned550(db, "a") || musicAssigned550(db, "c") || musicAssigned550(db, "e") {
		t.Fatal("foreign learner card leaked")
	}
	foreign := legacyUser(db, "other-tm")
	if foreign != nil {
		t.Fatal("regional must not see a tutor created by somebody else")
	}
	if legacyUser(db, "coordinador") != nil {
		t.Fatal("uncreated tutor exposed")
	}
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", token, map[string]any{"actorId": "regional", "ownerId": "a", "action": "assign_self", "restoreGeneration": 1, "expectedUpdatedAt": "", "value": map[string]any{}})
	if code != 403 {
		t.Fatal("self assignment must be denied", code)
	}
	st, _ := s.loadLegacy()
	if legacyTutorAssigned(legacyUser(st.State.Database, "a"), "regional", musicProgram553) || legacyRole(legacyUser(st.State.Database, "regional")) != "regional" {
		t.Fatal("self assignment changed role or was lost")
	}
}

func Test560ExcelPersonalNamesAndAdministrativeScope(t *testing.T) {
	_, h := excelFixture560(t)
	for _, id := range []string{"a", "pastor", "administrativo"} {
		t.Run(id, func(t *testing.T) {
			_, login := musicLogin553(t, h, id, true)
			code, snap := musicCall553(t, h, "GET", musicAPI553+"/api/accounts/view?actorId="+id, stringValue(login["sessionToken"]), nil)
			if code != 200 {
				t.Fatal(code, snap)
			}
			db := legacyMap(snap["database"])
			if id == "administrativo" {
				if !musicAssigned550(db, "e") || musicAssigned550(db, "a") || musicAssigned550(db, "b") {
					t.Fatal("Administrative sees unassigned cards")
				}
				return
			}
			for owner := range legacyMap(db["programProgress"]) {
				if owner != id {
					t.Fatal("another person's progress leaked", owner)
				}
			}
			for _, raw := range legacyUsers(db) {
				u := legacyMap(raw)
				if stringValue(u["id"]) != id {
					if stringValue(u["id"]) != "tm" || len(u) > 5 {
						t.Fatal("aspirant sees more than assigned tutor name", u)
					}
				}
			}
		})
	}
}

func Test560ExcelNineRolesNarrowActions(t *testing.T) {
	// Expected values transcribed from spreadsheet columns B, D, I, K, N, O, P, Q.
	rows := []struct {
		id                                                                          string
		review, recommend, tutor, instructions, registry, events, annual, templates bool
	}{
		{"admin", true, true, true, true, true, true, true, true},
		{"departamental", true, false, true, true, true, true, true, true},
		{"secretaria", false, true, true, true, true, true, true, true},
		{"coordinador", true, true, true, true, true, true, false, true},
		{"administrativo", true, true, true, false, false, true, false, false},
		{"regional", true, true, true, false, false, false, false, false},
		{"tm", true, false, false, false, false, false, false, false},
		{"a", false, false, false, false, false, false, false, false},
		{"pastor", false, false, false, false, false, false, false, false},
	}
	for _, row := range rows {
		t.Run(row.id, func(t *testing.T) {
			s, h := excelFixture560(t)
			_, login := musicLogin553(t, h, row.id, true)
			token := stringValue(login["sessionToken"])
			st, _ := s.loadLegacy()
			db := legacyMap(st.State.Database)
			actor := legacyUser(db, row.id)
			owner := legacyUser(db, "a")
			if row.id == "regional" {
				owner = legacyUser(db, "b")
			}
			if row.id == "administrativo" {
				owner = legacyUser(db, "e")
			}
			if got := canReview556(db, actor, owner, musicProgram553); got != row.review {
				t.Fatalf("review=%v want=%v", got, row.review)
			}
			if canReview556(db, actor, actor, musicProgram553) {
				t.Fatal("self review")
			}
			event := map[string]any{"name": "Creencias de prueba", "category": musicProgram553, "date": "2026-01-01", "time": "09:00", "place": "Prueba", "description": "Descripción de prueba", "certificateKind": "general"}
			st, _ = s.loadLegacy()
			base := stringValue(musicData550(legacyMap(st.State.Database))["updatedAt"])
			code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/music/module", token, map[string]any{"actorId": row.id, "action": "event", "value": event, "baseUpdatedAt": base, "restoreGeneration": 1})
			if (code == 200) != row.events {
				t.Fatal("events", code)
			}
			code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/tutor", token, map[string]any{"actorId": row.id, "id": "new-tutor", "programId": musicProgram553, "programIds": []any{musicProgram553}, "restoreGeneration": 1, "user": map[string]any{"fullName": "Tutor prueba", "email": "new@example.test", "username": "new-tutor", "region": "1", "password": tutorPassword556()}})
			if (code == 200) != row.tutor {
				t.Fatal("create tutor", code)
			}
			code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/indications", token, map[string]any{"actorId": row.id, "restoreGeneration": 1, "edits": []any{map[string]any{"id": "excel-indication", "programId": musicProgram553, "requirementId": "LM-PRE-1", "instructions": "Indicación de prueba"}}})
			if (code == 200) != row.instructions {
				t.Fatal("instructions", code)
			}
			policy := musicPolicies560[legacyRole(actor)]
			if policy.Recommend != row.recommend || policy.Registry != row.registry || policy.Annual != row.annual || policy.Templates != row.templates {
				t.Fatal("spreadsheet column mismatch")
			}
		})
	}
}

func Test560FinalApprovalExceptionAndNotifications(t *testing.T) {
	for _, tc := range []struct {
		actor, owner string
		want         int
	}{{"admin", "departamental", 200}, {"admin", "b", 200}, {"departamental", "b", 200}, {"departamental", "departamental", 403}, {"secretaria", "b", 403}} {
		s, h := excelFixture560(t)
		_, login := musicLogin553(t, h, tc.actor, true)
		st, _ := s.loadLegacy()
		base := stringValue(musicData550(legacyMap(st.State.Database))["updatedAt"])
		code, out := musicCall553(t, h, "POST", musicAPI553+"/api/music/module", stringValue(login["sessionToken"]), map[string]any{"actorId": tc.actor, "ownerId": tc.owner, "action": "approve", "baseUpdatedAt": base, "restoreGeneration": 1})
		if code != tc.want {
			t.Fatal(tc, code, out)
		}
	}
	s, _ := excelFixture560(t)
	st, _ := s.loadLegacy()
	ids := activeLegacyReviewerIDs(st.State.Database, "a", musicProgram553)
	if !containsString(stringSliceToAny550(ids), "secretaria") || containsString(stringSliceToAny550(ids), "regional") || containsString(stringSliceToAny550(ids), "administrativo") {
		t.Fatal("wrong notification scope", ids)
	}
}

func Test560RemovalIsScopedAndRequiresBothCreatorAndTutor(t *testing.T) {
	for _, tc := range []struct {
		actor, owner string
		want         int
	}{{"admin", "a", 200}, {"departamental", "admin", 403}, {"secretaria", "departamental", 403}, {"coordinador", "a", 200}, {"administrativo", "a", 403}, {"administrativo", "e", 403}, {"regional", "a", 403}, {"regional", "d", 403}, {"tm", "a", 403}, {"a", "b", 403}} {
		s, h := excelFixture560(t)
		st, _ := s.loadLegacy()
		other := cloneAny(legacyProgressForRequirement(st.State.Database, "a", "guide_major", false))
		_, login := musicLogin553(t, h, tc.actor, true)
		code, out := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", stringValue(login["sessionToken"]), map[string]any{"actorId": tc.actor, "ownerId": tc.owner, "action": "remove", "restoreGeneration": 1, "expectedUpdatedAt": ""})
		if code != tc.want {
			t.Fatal(tc, code, out)
		}
		st, _ = s.loadLegacy()
		if legacyUser(st.State.Database, tc.owner) == nil || !assignedProgram555(legacyMap(st.State.Database), "a", "guide_major") || !reflect.DeepEqual(other, legacyProgressForRequirement(st.State.Database, "a", "guide_major", false)) {
			t.Fatal("removal altered another SGL program/account")
		}
	}
}

func Test560RegionsProfileGuardAndMetadataCompression(t *testing.T) {
	s, h := excelFixture560(t)
	_, login := musicLogin553(t, h, "admin", true)
	token := stringValue(login["sessionToken"])
	for _, region := range []string{"", "13", "MBOS"} {
		code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/tutor", token, map[string]any{"actorId": "admin", "id": "region-tutor", "programId": musicProgram553, "programIds": []any{musicProgram553}, "restoreGeneration": 1, "user": map[string]any{"fullName": "Región prueba", "email": "region@example.test", "username": "region-tutor", "region": region, "password": tutorPassword556()}})
		if region == "MBOS" && code != 200 || region != "MBOS" && code != 400 {
			t.Fatal(region, code)
		}
	}
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	for _, id := range []string{"a", "administrativo", "regional", "tm"} {
		if musicAccountPatch560(db, legacyUser(db, id), map[string]any{"users": []any{map[string]any{"id": id, "fullName": "Cambio no autorizado"}}}) == "" {
			t.Fatal("profile guard", id)
		}
	}
	r := httptest.NewRequest("GET", musicAPI553+"/api/accounts/view?actorId=admin", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 200 || w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal(w.Code, w.Header())
	}
	gz, e := gzip.NewReader(bytes.NewReader(w.Body.Bytes()))
	if e != nil {
		t.Fatal(e)
	}
	data, _ := io.ReadAll(gz)
	gz.Close()
	var snap map[string]any
	json.Unmarshal(data, &snap)
	if snap["filesDeferred"] != true || strings.Contains(string(data), "data:text/plain;base64") || strings.Contains(string(data), "password") {
		t.Fatal("metadata view leaked files/credentials")
	}
	_, dept := musicLogin553(t, h, "departamental", true)
	code, _ := call555(t, h, "GET", musicAPI553+"/api/accounts/view?actorId=departamental", stringValue(dept["sessionToken"]), true, nil)
	if code != 200 {
		t.Fatal("Dept personal card", code)
	}
	code, _ = call555(t, h, "GET", musicAPI553+"/api/accounts/view?actorId=admin", token, true, nil)
	if code != 403 {
		t.Fatal("Admin personal Music card", code)
	}
}

func Test560CertificateEligibilityAttendanceAndName(t *testing.T) {
	s, h := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	events := ensureMapField(musicData550(db), "events")
	for _, id := range []string{"done", "enrolled", "untouched"} {
		e := map[string]any{"id": id, "name": id, "date": "2026-01-01", "time": "09:00", "place": "Prueba", "category": musicProgram553, "certificateKind": "beliefs", "participants": map[string]any{}}
		if id != "untouched" {
			ensureMapField(e, "participants")["a"] = map[string]any{"attended": id == "done"}
		}
		events[id] = e
	}
	s.saveLegacy(st)
	_, login := musicLogin553(t, h, "a", true)
	token := stringValue(login["sessionToken"])
	code, p := musicCall553(t, h, "GET", musicAPI553+"/api/music/module?actorId=a", token, nil)
	if code != 200 || len(mapField(legacyMap(p["data"]), "events")) != 3 {
		t.Fatal("aspirant must see every created event", code, p)
	}
	code, p = musicCall553(t, h, "POST", musicAPI553+"/api/music/module", token, map[string]any{"actorId": "a", "ownerId": "a", "action": "certificates", "baseUpdatedAt": stringValue(musicData550(db)["updatedAt"]), "restoreGeneration": 1})
	if code != 200 {
		t.Fatal(code, p)
	}
	certs := mapField(legacyMap(p["data"]), "certificates")
	if len(certs) != 1 {
		t.Fatal("unearned certificate generated", certs)
	}
	for _, raw := range certs {
		c := legacyMap(raw)
		if stringValue(c["ownerId"]) != "a" || stringValue(c["fullName"]) != "Cuenta a" || stringValue(c["eventId"]) != "done" {
			t.Fatal("wrong certificate identity", c)
		}
	}
}

func Test560WorkbookResourceActionsThroughPortal(t *testing.T) {
	var template bytes.Buffer
	if err := png.Encode(&template, image.NewRGBA(image.Rect(0, 0, 400, 200))); err != nil {
		t.Fatal(err)
	}
	for _, actorID := range []string{"admin", "departamental", "secretaria", "coordinador", "administrativo", "regional", "tm", "a", "pastor"} {
		for _, action := range []string{"registry", "template", "conclude", "investiture"} {
			t.Run(actorID+"/"+action, func(t *testing.T) {
				s, h := excelFixture560(t)
				st, _ := s.loadLegacy()
				db := legacyMap(st.State.Database)
				ownerID := "a"
				if actorID == "a" {
					ownerID = "b"
				} else if actorID == "administrativo" {
					ownerID = "e"
				}
				actor := legacyUser(db, actorID)
				policy := musicPolicies560[legacyRole(actor)]
				allowed := false
				path := musicAPI553 + "/api/music/module"
				value := map[string]any{}
				switch action {
				case "registry":
					allowed = policy.Registry
					path = musicAPI553 + "/api/music-registration"
					value = map[string]any{"type": "Otro", "otherType": "Ministerio de instrumentos", "name": "Ministerio de prueba", "church": "Iglesia", "district": "Distrito", "region": "MBOS", "mission": "Misión", "union": "Unión", "createdDate": "2026-01-01", "boardVote": "12", "director": "Director", "directorPhone": "12345678"}
				case "template":
					allowed = policy.Templates
					value = map[string]any{"kind": "final", "dataUrl": "data:image/png;base64," + base64.StdEncoding.EncodeToString(template.Bytes())}
				case "conclude":
					allowed = policy.Annual
					p := reviewProgressContainer(db, ownerID, musicProgram553)
					ensureMapField(p, "status")["LM-NUT-1"] = "in_progress"
					ensureMapField(p, "submissions")["LM-NUT-1"] = map[string]any{"bibleReadings": map[string]any{"choice": "reavivados", "kind": "music_devotion"}}
					value = map[string]any{"kind": "reavivados"}
				case "investiture":
					allowed = policy.InvestOthers
					p := reviewProgressContainer(db, ownerID, musicProgram553)
					for _, reqID := range musicRequirements550 {
						ensureMapField(p, "status")[reqID] = "approved"
					}
					ensureMapField(musicData550(db), "workflow")[ownerID] = map[string]any{"state": "approved"}
					value = map[string]any{"place": "Iglesia", "date": "2026-10-10", "time": "10:00", "investedBy": "Departamental"}
				}
				if err := s.saveLegacy(st); err != nil {
					t.Fatal(err)
				}
				_, login := musicLogin553(t, h, actorID, true)
				base := stringValue(musicData550(db)["updatedAt"])
				if action == "registry" {
					base = stringValue(legacyMap(legacyUser(db, ownerID)["musicMinistryRegistration"])["updatedAt"])
				}
				code, result := musicCall553(t, h, "POST", path, stringValue(login["sessionToken"]), map[string]any{"actorId": actorID, "ownerId": ownerID, "action": action, "value": value, "baseUpdatedAt": base, "restoreGeneration": 1})
				if allowed && code != 200 || !allowed && code != 403 {
					t.Fatalf("allowed=%v status=%d body=%v", allowed, code, result)
				}
			})
		}
	}
}

func TestBrowserFixture560(t *testing.T) {
	if os.Getenv("SGL560_BROWSER_FIXTURE") != "1" {
		t.Skip("interactive fixture")
	}
	s, h := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	ensureMapField(musicData550(db), "events")["unattended"] = map[string]any{"id": "unattended", "name": "Evento sin asistencia", "date": "2026-01-01", "time": "09:00", "category": musicProgram553, "participants": map[string]any{}}
	s.saveLegacy(st)
	ln, e := net.Listen("tcp", "127.0.0.1:8900")
	if e != nil {
		t.Fatal(e)
	}
	defer ln.Close()
	fmt.Println("SGL560_BROWSER_READY")
	http.Serve(ln, h)
}
