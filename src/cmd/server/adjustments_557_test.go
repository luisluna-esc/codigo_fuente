package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func seedReavivados557(db map[string]any, owner string) {
	p := reviewProgressContainer(db, owner, musicProgram553)
	ensureMapField(p, "status")["LM-NUT-1"] = "approved"
	ensureMapField(p, "submissions")["LM-NUT-1"] = map[string]any{"text": "Lecturas anuales", "attachmentIds": []any{"year-file"}, "bibleReadings": map[string]any{"kind": "music_devotion", "choice": "reavivados", "startDate": "2026-01-01", "entries": map[string]any{"2026-01-01": map[string]any{"completed": true, "date": "2026-01-01", "favoriteVerses": "Génesis 1:1", "impactStory": "Dios crea", "lifeLesson": "Confianza"}}}}
	p["updatedAt"] = "reading-version"
	ensureMapField(p, "status")["LM-TEM-1"] = "approved"
	ensureMapField(p, "submissions")["LM-TEM-1"] = map[string]any{"text": "Otros 40 días"}
	d := musicData550(db)
	ensureMapField(d, "completions")[owner] = map[string]any{"reavivados": map[string]any{"concluded": true}, "forty_days": map[string]any{"concluded": true}}
	ensureMapField(d, "workflow")[owner] = map[string]any{"state": "approved"}
	ensureMapField(ensureMapField(db, "bibleYear"), owner)[musicProgram553] = map[string]any{"text": "Lecturas antiguas"}
}

func reset557(t *testing.T, s *Server, role string, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	r := httptest.NewRequest("POST", "/api/progress/reset", bytes.NewReader(b))
	r.Header.Set("Authorization", "Bearer "+role)
	w := httptest.NewRecorder()
	s.securityHeaders(http.HandlerFunc(s.legacyProgressReset)).ServeHTTP(w, r)
	return w
}
func resetBody557(role, owner string) map[string]any {
	return map[string]any{"actorId": role, "ownerId": owner, "programId": musicProgram553, "requirementId": "LM-NUT-1", "scope": "requirement", "operationId": "reset-557", "baseUpdatedAt": "reading-version", "restoreGeneration": 1}
}
func resetFixture557(t *testing.T) *Server {
	s, _ := fixture550(t)
	st, _ := s.loadLegacy()
	seedReavivados557(legacyMap(st.State.Database), "p")
	st.State.Files = append(st.State.Files, map[string]any{"id": "year-file", "ownerId": "p", "programId": musicProgram553, "requirementId": "LM-NUT-1"})
	s.saveLegacy(st)
	return s
}
func Test557AllFourRolesResetOnlyReavivadosAndInvalidateDerivedApproval(t *testing.T) {
	for _, role := range []string{"administrativo", "departamental", "secretaria", "regional"} {
		t.Run(role, func(t *testing.T) {
			s := resetFixture557(t)
			if role == "administrativo" || role == "regional" {
				st, _ := s.loadLegacy()
				legacyUser(st.State.Database, "p")["tutorByProgram"] = map[string]any{musicProgram553: role}
				s.saveLegacy(st)
			}
			body := resetBody557(role, "p")
			w := reset557(t, s, role, body)
			if w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			st, _ := s.loadLegacy()
			db := legacyMap(st.State.Database)
			p := legacyProgressForRequirement(db, "p", musicProgram553, false)
			if stringValue(mapField(p, "status")["LM-NUT-1"]) != "pending" || mapField(p, "submissions")["LM-NUT-1"] != nil {
				t.Fatal("annual progress survived")
			}
			if mapField(p, "submissions")["LM-TEM-1"] == nil || stringValue(mapField(p, "status")["LM-TEM-1"]) != "approved" {
				t.Fatal("unrelated requirement changed")
			}
			d := musicData550(db)
			if legacyMap(mapField(d, "completions")["p"])["reavivados"] != nil || musicEligible550(db, "p", "reavivados", "") || legacyMap(mapField(db, "bibleYear")["p"])[musicProgram553] != nil {
				t.Fatal("deleted diary remained certified or recoverable")
			}
			if stringValue(legacyMap(mapField(d, "workflow")["p"])["state"]) == "approved" {
				t.Fatal("final approval survived deletion")
			}
			for _, f := range st.State.Files {
				if stringValue(legacyMap(f)["id"]) == "year-file" {
					t.Fatal("attachment survived")
				}
			}
			if stringValue(legacyMap(arrayField(db, "audit")[0])["programId"]) != musicProgram553 {
				t.Fatal("audit has no module")
			}
			// A delayed retry must not delete newly saved progress.
			ensureMapField(p, "submissions")["LM-NUT-1"] = map[string]any{"text": "Nueva lectura"}
			s.saveLegacy(st)
			if w = reset557(t, s, role, body); w.Code != 200 {
				t.Fatal("retry failed", w.Code, w.Body.String())
			}
			st, _ = s.loadLegacy()
			if stringValue(legacyMap(mapField(legacyProgressForRequirement(st.State.Database, "p", musicProgram553, false), "submissions")["LM-NUT-1"])["text"]) != "Nueva lectura" {
				t.Fatal("retry deleted new progress")
			}
		})
	}
}
func Test557ResetDeniesScopeForgingStaleDataAndOtherRequirements(t *testing.T) {
	cases := []string{"module", "regional_unassigned", "other_program", "stale", "restore", "forged_actor", "self", "inactive"}
	for _, test := range cases {
		t.Run(test, func(t *testing.T) {
			s := resetFixture557(t)
			st, _ := s.loadLegacy()
			db := legacyMap(st.State.Database)
			role := "secretaria"
			body := resetBody557(role, "p")
			expected := 403
			switch test {
			case "module":
				legacyUser(db, role)["managementProgramIds"] = []any{"guide_major"}
			case "regional_unassigned":
				role = "regional"
				body["actorId"] = role
				legacyUser(db, role)["regionalAssignedUserIds"] = []any{}
				legacyUser(db, role)["region"] = "2"
			case "other_requirement":
				body["requirementId"] = "LM-TEM-1"
			case "other_program":
				body["programId"] = "guide_major"
			case "selection":
				body["scope"] = "selection"
				body["requirementIds"] = []any{"LM-NUT-1"}
			case "stale":
				body["baseUpdatedAt"] = "old"
				expected = 409
			case "restore":
				body["restoreGeneration"] = 2
				expected = 409
			case "bible_plan":
				legacyMap(mapField(reviewProgressContainer(db, "p", musicProgram553), "submissions")["LM-NUT-1"])["bibleReadings"] = map[string]any{"choice": "bible_plan"}
			case "forged_actor":
				body["actorId"] = "admin"
			case "self":
				body["ownerId"] = role
				seedReavivados557(db, role)
			case "inactive":
				legacyUser(db, role)["active"] = false
			}
			s.saveLegacy(st)
			w := reset557(t, s, role, body)
			if w.Code != expected {
				t.Fatal(test, w.Code, w.Body.String())
			}
			st, _ = s.loadLegacy()
			if stringValue(mapField(legacyProgressForRequirement(st.State.Database, "p", musicProgram553, false), "status")["LM-NUT-1"]) != "approved" {
				t.Fatal("denied request changed progress")
			}
		})
	}
}
func Test557WhiteCertificateRequiresAssociatedEventAndAttendance(t *testing.T) {
	s, _ := fixture550(t)
	v := map[string]any{"name": "Lectura organizada", "date": "2026-01-01", "time": "09:00", "place": "Iglesia", "category": musicProgram553, "certificateKind": "egw_reading", "hours": 0}
	if w := call550(t, s, "coordinador_musica", "", "event", v, false); w.Code != 400 {
		t.Fatal("missing book accepted")
	}
	v["bookTitle"] = "Liderazgo Cristiano"
	if w := call550(t, s, "coordinador_musica", "", "event", v, false); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	st, _ := s.loadLegacy()
	var eid string
	for id := range mapField(musicData550(legacyMap(st.State.Database)), "events") {
		eid = id
	}
	if w := call550(t, s, "coordinador_musica", "p", "enroll", map[string]any{"eventId": eid}, false); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	cv := map[string]any{"kind": "egw_reading", "eventId": eid}
	if w := call550(t, s, "p", "p", "certificate", cv, false); w.Code != 400 {
		t.Fatal("attendance missing but certificate issued")
	}
	call550(t, s, "coordinador_musica", "p", "attendance", map[string]any{"eventId": eid, "attended": true}, false)
	for _, role := range []string{"coordinador_musica", "p"} {
		if w := call550(t, s, role, "p", "certificate", cv, false); w.Code != 200 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	v["id"] = eid
	v["bookTitle"] = "Otro libro"
	if w := call550(t, s, "coordinador_musica", "", "event", v, false); w.Code != 400 {
		t.Fatal("enrolled book replaced")
	}
	st, _ = s.loadLegacy()
	db := legacyMap(st.State.Database)
	c := legacyMap(mapField(musicData550(db), "certificates")["p:egw_reading:"+eid])
	if stringValue(legacyMap(c["event"])["bookTitle"]) != "Liderazgo Cristiano" {
		t.Fatal("certificate lost associated book")
	}
	// Reading reports are independent of this event certificate.
	ensureMapField(reviewProgressContainer(db, "p", musicProgram553), "status")["LM-NUT-4"] = "pending"
	s.saveLegacy(st)
	if !musicEligible550(db, "p", "egw_reading", eid) || musicEligible550(db, "p", "egw_reading", "") {
		t.Fatal("certificate still depends on requirement reviews")
	}
	call550(t, s, "coordinador_musica", "p", "attendance", map[string]any{"eventId": eid, "attended": false}, false)
	w := httptest.NewRecorder()
	s.musicCertificate550(w, httptest.NewRequest("GET", "/api/music/certificate?token="+stringValue(c["token"]), nil))
	if !strings.Contains(w.Body.String(), "Pendiente de nueva validación") {
		t.Fatal("removed attendance still validates")
	}
}

func TestBrowserFixture557(t *testing.T) {
	if os.Getenv("SGL557_BROWSER_FIXTURE") != "1" {
		t.Skip("interactive fixture")
	}
	s, h := setup555(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	for _, id := range []string{"a", "b", "departamental", "secretaria", "administrativo", "coordinador", "tg"} {
		seedReavivados557(db, id)
	}
	regional := legacyUser(db, "coordinador")
	regional["role"] = "regional"
	regional["managementProgramIds"] = []any{musicProgram553}
	regional["regionalAssignedUserIds"] = []any{"a", "b", "tg"}
	p := reviewProgressContainer(db, "coordinador", musicProgram553)
	ensureMapField(p, "status")["LM-PRE-1"] = "in_review"
	ensureMapField(p, "submissions")["LM-PRE-1"] = map[string]any{"text": "Informe personal del Regional.\nSegunda línea de evidencia.\nTercera línea de aplicación.", "updatedAt": "before"}
	// Institutional owner whose personal module differs from management modules.
	legacyUser(db, "tg")["managementProgramIds"] = []any{"guide_major"}
	for _, id := range []string{"a", "b", "departamental", "administrativo"} {
		p := reviewProgressContainer(db, id, musicProgram553)
		ensureMapField(p, "status")["LM-NUT-4"] = "approved"
		ensureMapField(p, "status")["LM-REL-5"] = "approved"
	}
	s.saveLegacy(st)
	ln, err := net.Listen("tcp", "127.0.0.1:8898")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	fmt.Println("SGL557_BROWSER_READY")
	http.Serve(ln, h)
}
