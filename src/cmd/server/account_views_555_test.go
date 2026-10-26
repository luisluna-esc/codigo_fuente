package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func call555(t *testing.T, h http.Handler, method, path, token string, personal bool, body any) (int, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(body)
	r := httptest.NewRequest(method, "http://example.test"+path, bytes.NewReader(b))
	r.Header.Set("Authorization", "Bearer "+token)
	r.Header.Set("Content-Type", "application/json")
	if personal {
		r.Header.Set("X-SGL-Personal", "1")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	out := map[string]any{}
	json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}
func setup555(t *testing.T) (*Server, http.Handler) {
	s, h := musicFixture553(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if legacyRole(u) != "aspirant" {
			for _, rm := range arrayField(db, "modules") {
				m := legacyMap(rm)
				m["assignedUserIds"] = append(arrayField(m, "assignedUserIds"), stringValue(u["id"]))
			}
			u["tutorByProgram"] = map[string]any{"guide_major": "tg", musicProgram553: "tm"}
		}
	}
	for _, raw := range st.State.Files {
		f := legacyMap(raw)
		if stringValue(f["id"]) == "music-a" {
			f["programId"] = musicProgram553
		} else {
			f["programId"] = "guide_major"
		}
	}
	s.saveLegacy(st)
	return s, h
}
func Test555AuthorityOwnCardSameIdentity(t *testing.T) {
	for _, id := range []string{"admin", "departamental", "secretaria", "administrativo", "tm", "coordinador"} {
		t.Run(id, func(t *testing.T) {
			s, h := setup555(t)
			_, login := musicLogin553(t, h, id, false)
			token := stringValue(login["sessionToken"])
			if token == "" {
				t.Fatal(login)
			}
			code, snap := call555(t, h, "GET", "/api/remote/bootstrap", token, true, nil)
			if code != 200 {
				t.Fatal(code, snap)
			}
			db := legacyMap(snap["database"])
			if legacyMap(db["programProgress"])["a"] != nil {
				t.Fatal("personal view leaked another card")
			}
			actual := legacyUser(db, id)
			if actual == nil || legacyRole(actual) == "aspirant" {
				t.Fatal("stored authority role was replaced")
			}
			source := legacyMap(cloneAny(db))
			p := reviewProgressContainer(source, id, "guide_major")
			ensureMapField(p, "status")["GM-PRE-1"] = "in_progress"
			ensureMapField(p, "submissions")["GM-PRE-1"] = map[string]any{"text": "Primera línea de mi informe personal.\nSegunda línea para explicar mi formación.\nTercera línea para mostrar mis aprendizajes.", "richText": "", "attachmentIds": []any{}, "updatedAt": time.Now().UTC().Format(time.RFC3339Nano)}
			code, result := call555(t, h, "POST", "/api/sync", token, true, map[string]any{"actorId": id, "actorRole": "aspirant", "installationId": "self-test", "baseVersion": 7, "restoreGeneration": 1, "operationId": "own-" + id, "operationType": "save_progress", "programId": "guide_major", "ownerId": id, "requirementId": "GM-PRE-1", "database": source, "files": []any{}, "ackOnly": true})
			if code != 200 {
				t.Fatal(code, result)
			}
			st, _ := s.loadLegacy()
			if legacyRole(legacyUser(st.State.Database, id)) != legacyRole(actual) {
				t.Fatal("authority role lost")
			}
			if stringValue(mapField(legacyProgressForRequirement(st.State.Database, id, "guide_major", false), "status")["GM-PRE-1"]) != "in_progress" {
				t.Fatal("own progress missing")
			}
			code, result = call555(t, h, "POST", "/api/reviews/submit", token, true, map[string]any{"actorId": id, "actorRole": "aspirant", "ownerId": id, "programId": "guide_major", "requirementId": "GM-PRE-1", "operationId": "review-own-" + id, "installationId": "self-test", "submission": mapField(p, "submissions")["GM-PRE-1"], "files": []any{}, "ackOnly": true})
			if code != 200 {
				t.Fatal("own review submit", code, result)
			}
			code, _ = call555(t, h, "POST", "/api/reviews/decision", token, false, map[string]any{"actorId": id, "ownerId": id, "programId": "guide_major", "requirementId": "GM-PRE-1", "decision": "approved"})
			if code != 403 {
				t.Fatal("self approval permitted", code)
			}
			st, _ = s.loadLegacy()
			if stringValue(legacyMap(legacyMap(legacyMap(st.State.Database)["programProgress"])["a"])["guide_major"]) == "" {
				t.Fatal("other programs changed")
			}
		})
	}
}
func Test555ScopedModulesRegionalAndRevocation(t *testing.T) {
	s, h := setup555(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	u := legacyUser(db, "departamental")
	u["managementProgramIds"] = []any{musicProgram553}
	legacyUser(db, "tm")["regionalAssignedUserIds"] = []any{}
	s.saveLegacy(st)
	_, login := musicLogin553(t, h, "departamental", false)
	token := stringValue(login["sessionToken"])
	_, snap := call555(t, h, "GET", "/api/remote/bootstrap", token, false, nil)
	out := legacyMap(snap["database"])
	if legacyMap(legacyMap(out["programProgress"])["a"])["guide_major"] != nil {
		t.Fatal("unassigned module exposed")
	}
	if legacyMap(legacyMap(out["programProgress"])["a"])[musicProgram553] == nil {
		t.Fatal("assigned module missing")
	}
	st, _ = s.loadLegacy()
	legacyUser(st.State.Database, "departamental")["managementProgramIds"] = []any{}
	s.saveLegacy(st)
	_, outLogin := musicLogin553(t, h, "departamental", true)
	if stringValue(outLogin["sessionToken"]) == "" {
		t.Fatal("personal Music enrollment must still allow login", outLogin)
	}
	// Music requires actual tutoring; regional lists and region do not grant access.
	actor := map[string]any{"id": "regional", "role": "regional", "active": true, "region": "1", "managementProgramIds": []any{musicProgram553}, "regionalAssignedUserIds": []any{"a"}}
	if canReadOwner555(db, actor, legacyUser(db, "a"), musicProgram553) || canReadOwner555(db, actor, legacyUser(db, "b"), musicProgram553) {
		t.Fatal("regional inherited access without tutoring")
	}
}
func Test555DriveErrorPersistsAndCannotEraseSuccess(t *testing.T) {
 t.Skip("Drive fue retirado del producto Música Online; la denegación de rutas se prueba en TestMusicOnlineRemovedEndpointsAndProgramBoundary")
	s, h := setup555(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	ensureMapField(musicData550(db), "interface598")["driveVisible"] = true // 6.0.2 requires explicit Drive enablement.
	p := reviewProgressContainer(db, "a", musicProgram553)
	ensureMapField(p, "status")["LM-PRE-1"] = "approved"
	sub := map[string]any{"text": "Contenido anterior", "updatedAt": "before", "approval": map[string]any{"at": time.Now().Add(-7 * 24 * time.Hour).UTC().Format(time.RFC3339)}}
	ensureMapField(p, "submissions")["LM-PRE-1"] = sub
	s.saveLegacy(st)
	_, login := musicLogin553(t, h, "admin", true)
	token := stringValue(login["sessionToken"])
	body := map[string]any{"actorId": "admin", "ownerId": "a", "programId": musicProgram553, "requirementId": "LM-PRE-1", "error": "No se pudo recuperar el archivo", "sourceUpdatedAt": "before"}
	code, result := call555(t, h, "POST", musicAPI553+"/api/reports/drive/attempt", token, false, body)
	if code != 200 {
		t.Fatal(code, result)
	}
	_, snap := call555(t, h, "GET", musicAPI553+"/api/remote/bootstrap", token, false, nil)
	out := legacyMap(snap["database"])
	p = legacyProgressForRequirement(out, "a", musicProgram553, false)
	d := legacyMap(legacyMap(mapField(p, "submissions")["LM-PRE-1"])["drive"])
	if stringValue(d["status"]) != "error" {
		t.Fatal("Drive failure disappeared")
	}
	st, _ = s.loadLegacy()
	sub = legacyMap(mapField(legacyProgressForRequirement(st.State.Database, "a", musicProgram553, false), "submissions")["LM-PRE-1"])
	sub["drive"] = map[string]any{"status": "sent", "fileId": "receipt"}
	s.saveLegacy(st)
	delete(body, "sourceUpdatedAt")
	code, _ = call555(t, h, "POST", musicAPI553+"/api/reports/drive/attempt", token, false, body)
	if code != 409 {
		t.Fatal("stale failure replaced successful upload", code)
	}
}
func TestBrowserFixture555(t *testing.T) {
	if os.Getenv("SGL555_BROWSER_FIXTURE") != "1" {
		t.Skip("interactive fixture")
	}
	s, h := setup555(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	for _, id := range []string{"a", "admin", "departamental", "secretaria", "administrativo", "tm", "tg", "coordinador"} {
		u := legacyUser(db, id)
		u["drive"] = map[string]any{"connected": true, "folderName": "Prueba"}
		p := reviewProgressContainer(db, id, musicProgram553)
		ensureMapField(p, "status")["LM-PRE-1"] = "approved"
		ensureMapField(p, "submissions")["LM-PRE-1"] = map[string]any{"text": "Contenido real del informe anterior.\nUna recomendación de la iglesia.\nEl documento tiene tres líneas de contenido.", "richText": "<p>Contenido real del informe anterior.</p><p>Una recomendación de la iglesia.</p><p>El documento tiene tres líneas de contenido.</p>", "updatedAt": "before", "approval": map[string]any{"at": time.Now().Add(-7 * 24 * time.Hour).UTC().Format(time.RFC3339)}}
	}
	db["reportRequirementModes"] = map[string]any{musicProgram553 + ":LM-PRE-1": map[string]any{"mode": "document"}}
	s.saveLegacy(st)
	ln, err := net.Listen("tcp", "127.0.0.1:8896")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	fmt.Println("SGL555_BROWSER_READY")
	http.Serve(ln, h)
}

func Test555PersonalCannotChangeAccountOrPublishPartialAdminState(t *testing.T) {
	s, h := setup555(t)
	_, login := musicLogin553(t, h, "admin", false)
	token := stringValue(login["sessionToken"])
	_, snap := call555(t, h, "GET", "/api/remote/bootstrap", token, true, nil)
	source := legacyMap(snap["database"])
	legacyUser(source, "admin")["managementProgramIds"] = []any{}
	code, _ := call555(t, h, "POST", "/api/sync", token, true, map[string]any{"actorId": "admin", "actorRole": "aspirant", "installationId": "account-test", "baseVersion": 7, "restoreGeneration": 1, "operationType": "save_progress", "database": source, "files": []any{}})
	if code != 200 {
		t.Fatal(code)
	}
	st, _ := s.loadLegacy()
	if legacyUser(st.State.Database, "admin")["managementProgramIds"] != nil {
		t.Fatal("personal request changed authority permissions")
	}
	source["accountScope555"] = map[string]any{"view": "personal"}
	code, _ = call555(t, h, "POST", "/api/sync", token, false, map[string]any{"actorId": "admin", "actorRole": "admin", "baseVersion": st.State.ServerVersion, "restoreGeneration": 1, "database": source})
	if code != 409 {
		t.Fatal("partial card replaced administrative database", code)
	}
	after, _ := s.loadLegacy()
	if len(legacyUsers(after.State.Database)) != len(legacyUsers(st.State.Database)) {
		t.Fatal("accounts removed")
	}
}

func Test555AuthorityCardReviewedByAssignedTutor(t *testing.T) {
	s, h := setup555(t)
	st, _ := s.loadLegacy()
	p := reviewProgressContainer(legacyMap(st.State.Database), "departamental", "guide_major")
	ensureMapField(p, "status")["GM-PRE-1"] = "in_review"
	ensureMapField(p, "submissions")["GM-PRE-1"] = map[string]any{"text": "Tarjeta propia de una autoridad.", "updatedAt": "before", "reviewSubmittedAt": "2026-10-01T00:00:00Z"}
	s.saveLegacy(st)
	_, login := musicLogin553(t, h, "tg", false)
	code, out := call555(t, h, "POST", "/api/reviews/decision", stringValue(login["sessionToken"]), false, map[string]any{"actorId": "tg", "ownerId": "departamental", "programId": "guide_major", "requirementId": "GM-PRE-1", "decision": "approved", "feedback": "Revisado por su tutor"})
	if code != 200 {
		t.Fatal(code, out)
	}
	after, _ := s.loadLegacy()
	if stringValue(mapField(legacyProgressForRequirement(after.State.Database, "departamental", "guide_major", false), "status")["GM-PRE-1"]) != "approved" {
		t.Fatal("authority card was not reviewed")
	}
	if legacyRole(legacyUser(after.State.Database, "departamental")) != "departamental" {
		t.Fatal("authority identity lost")
	}
}

func Test555ModulePermissionsProtectDriveAndMusicOperations(t *testing.T) {
	s, h := setup555(t)
	st, _ := s.loadLegacy()
	legacyUser(st.State.Database, "departamental")["managementProgramIds"] = []any{}
	s.saveLegacy(st)
	_, login := musicLogin553(t, h, "departamental", false)
	token := stringValue(login["sessionToken"])
	code, _ := call555(t, h, "POST", "/api/google/upload", token, false, map[string]any{"userId": "a", "programId": musicProgram553, "requirementId": "LM-PRE-1", "filename": "informe.pdf", "folderName": "Prueba", "dataBase64": "JVBERi0="})
	if code != 403 {
		t.Fatal("Drive upload outside enabled modules", code)
	}
	code, _ = call555(t, h, "POST", "/api/music/module", token, false, map[string]any{"actorId": "departamental", "action": "event", "value": map[string]any{"name": "Evento sin permiso"}})
	if code != 403 {
		t.Fatal("personal enrollment granted music administration", code)
	}
}
