package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func call610(t *testing.T, s *Server, actor, owner, action string, v map[string]any) (int, map[string]any) {
	t.Helper()
	st, e := s.loadLegacy()
	if e != nil {
		t.Fatal(e)
	}
	b, _ := json.Marshal(musicRequest610{ActorID: actor, OwnerID: owner, Action: action, Value: v, RestoreGeneration: legacyCanonicalRestoreGeneration(st.State)})
	r := httptest.NewRequest("POST", "/api/music/v610", bytes.NewReader(b))
	r.Header.Set("Authorization", "Bearer "+actor)
	w := httptest.NewRecorder()
	s.music610(w, r)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}
func ministryFixture610() map[string]any {
	return map[string]any{"name": "Armonía", "type": "Coro", "church": "Central", "district": "Centro", "region": "1", "mission": "Misión", "union": "Unión", "createdDate": "2020-01-01", "boardVote": "18", "director": "Ana Música", "directorPhone": "12345678", "biography": "Ministerio dedicado al servicio musical.", "years": 6, "months": 2}
}
func registryID610(t *testing.T, p map[string]any) string {
	t.Helper()
	rows := arrayField(p, "ministries")
	if len(rows) == 0 {
		t.Fatal(p)
	}
	return stringValue(legacyMap(rows[len(rows)-1])["id"])
}
func Test610MinistryPermissionsEvidenceAndPDF(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	st, _ := s.loadLegacy()
	st.State.Database = db
	p := reviewProgressContainer(db, "p", musicProgram553)
	ensureMapField(p, "status")[ministryRequirement610] = "in_progress"
	previous := map[string]any{"text": "evidencia anterior", "primaryDocumentId": "manual", "attachmentIds": []any{"manual"}}
	ensureMapField(p, "submissions")[ministryRequirement610] = previous
	st.State.Files = append(st.State.Files, map[string]any{"id": "manual", "ownerId": "p", "programId": musicProgram553, "dataUrl": "data:application/pdf;base64,JVBERi0xLjQ="})
	if e := s.saveLegacy(st); e != nil {
		t.Fatal(e)
	}
	code, payload := call610(t, s, "administrativo", "", "ministry.save", ministryFixture610())
	if code != 200 {
		t.Fatal(code, payload)
	}
	id := registryID610(t, payload)
	v := ministryFixture610()
	v["id"] = id
	v["expectedUpdatedAt"] = legacyMap(arrayField(payload, "ministries")[0])["updatedAt"]
	if c, _ := call610(t, s, "administrativo", "", "ministry.save", v); c != 403 {
		t.Fatal("administrativo edited", c)
	}
	if c, p := call610(t, s, "administrativo", "p", "ministry.link", map[string]any{"id": id}); c != 200 {
		t.Fatal(c, p)
	}
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	sub := legacyMap(mapField(reviewProgressContainer(db, "p", musicProgram553), "submissions")[ministryRequirement610])
	if !musicLinkedSubmissionGuard610(db, "p", ministryRequirement610, sub) || musicLinkedSubmissionGuard610(db, "p", ministryRequirement610, previous) {
		t.Fatal("linked file guard")
	}
	if mapField(reviewProgressContainer(db, "p", musicProgram553), "status")[ministryRequirement610] != "in_progress" {
		t.Fatal("link must not submit or approve")
	}
	c := legacyMap(mapField(musicData550(db), "certificates")["p:ministry:"])
	if c == nil || c["fileId610"] == nil {
		t.Fatal("missing certificate")
	}
	found := false
	for _, f := range st.State.Files {
		file := legacyMap(f)
		if file["id"] == c["fileId610"] {
			raw, _ := base64.StdEncoding.DecodeString(strings.SplitN(stringValue(file["dataUrl"]), ",", 2)[1])
			if !bytes.HasPrefix(raw, []byte("%PDF-1.4")) || !bytes.Contains(raw, []byte("/FontFile")) {
				t.Fatal("bad PDF")
			}
			found = true
		}
	}
	if !found {
		t.Fatal("missing persisted PDF")
	}
	if c, p := call610(t, s, "administrativo", "p", "ministry.unlink", map[string]any{"id": id}); c != 200 {
		t.Fatal(c, p)
	}
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	restored := legacyMap(mapField(reviewProgressContainer(db, "p", musicProgram553), "submissions")[ministryRequirement610])
	if restored["primaryDocumentId"] != "manual" || restored["text"] != "evidencia anterior" {
		t.Fatal("manual evidence lost", restored)
	}
	if legacyMap(mapField(musicData550(db), "certificates")["p:ministry:"])["revokedAt"] == nil {
		t.Fatal("certificate stayed valid")
	}
	ensureMapField(musicData550(db), "permissions602")["roles"] = map[string]any{"administrativo": map[string]any{"ministries.create": false}, "admin": map[string]any{"ministries.create": false}}
	st.State.Database = db
	s.saveLegacy(st)
	if c, _ := call610(t, s, "administrativo", "", "ministry.save", ministryFixture610()); c != 403 {
		t.Fatal("disabled creation", c)
	}
	if c, p := call610(t, s, "admin", "", "ministry.save", ministryFixture610()); c != 200 {
		t.Fatal("admin access", c, p)
	}
}
func Test610EventsAllViewRosterAndAttendanceCertificate(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	d := musicData550(db)
	ensureMapField(d, "templates")["new"] = map[string]any{"name": "Encuentro coral", "signatureDepartment": "Firma misión", "signatureCoordinator": "Firma coordinación"}
	ensureMapField(d, "events")["ev"] = map[string]any{"id": "ev", "name": "Ensayo", "date": "2026-01-01", "templateId610": "new", "participants": map[string]any{"p": map[string]any{"attended": false}, "admin": map[string]any{"attended": false}}}
	st, _ := s.loadLegacy()
	st.State.Database = db
	s.saveLegacy(st)
	if len(mapField(musicView550(db, legacyUser(db, "p")), "events")) != 1 {
		t.Fatal("created event hidden")
	}
	if len(mapField(legacyMap(mapField(musicView550(db, legacyUser(db, "p")), "events")["ev"]), "participants")) != 1 {
		t.Fatal("learner roster leaked")
	}
	if len(mapField(legacyMap(mapField(musicView550(db, legacyUser(db, "regional")), "events")["ev"]), "participants")) != 2 {
		t.Fatal("regional roster missing")
	}
	if c := mapField(d, "certificates")["p:event:ev"]; c != nil {
		t.Fatal("enrolled got certificate")
	}
	legacyMap(mapField(legacyMap(mapField(d, "events")["ev"]), "participants")["p"])["attended"] = true
	st.State.Database = db
	s.saveLegacy(st)
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	c := legacyMap(mapField(musicData550(db), "certificates")["p:event:ev"])
	if c["title"] != "Encuentro coral" || c["fileId610"] == nil {
		t.Fatal(c)
	}
}
func Test610ReviewClockAndDedup(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	p := reviewProgressContainer(db, "p", musicProgram553)
	for _, rid := range []string{"LM-PRE-1", "LM-PRE-2", "LM-PRE-3"} {
		ensureMapField(p, "status")[rid] = "in_review"
	}
	subs := ensureMapField(p, "submissions")
	now := time.Now()
	subs["LM-PRE-1"] = map[string]any{"submittedAt": now.Add(-71 * time.Hour).Format(time.RFC3339Nano)}
	subs["LM-PRE-2"] = map[string]any{"submittedAt": now.Add(-73 * time.Hour).Format(time.RFC3339Nano)}
	subs["LM-PRE-3"] = map[string]any{"submittedAt": now.Add(-121 * time.Hour).Format(time.RFC3339Nano)}
	st, _ := s.loadLegacy()
	st.Notifications = nil
	st.State.Database = db
	if !musicFollowupNotifications610(&st) {
		t.Fatal("no followup")
	}
	count := len(st.Notifications)
	if musicFollowupNotifications610(&st) || len(st.Notifications) != count {
		t.Fatal("duplicate followup")
	}
	reports := map[string]bool{}
	for _, n := range st.Notifications {
		if strings.Contains(n.SourceOperationID, ":reporte:") {
			reports[n.UserID] = true
		}
		if strings.Contains(n.SourceOperationID, "LM-PRE-1") {
			t.Fatal("early reminder")
		}
		if n.Type == "revision610" && strings.Contains(strings.ToLower(n.Message), "admin") {
			t.Fatal("harsh named visible recipient")
		}
	}
	for _, id := range []string{"admin", "secretaria", "departamental", "coordinador_musica", "administrativo"} {
		if !reports[id] {
			t.Fatal("missing recipient", id)
		}
	}
}
func Test610BackupBlocksStaleResetAndArchiveSharesFiles(t *testing.T) {
	s, db := fixture550(t)
	s.configPath = filepath.Join(t.TempDir(), "server.json")
	musicEnsure610(db)
	st, _ := s.loadLegacy()
	st.State.Database = db
	s.saveLegacy(st)
	if c, _ := call610(t, s, "secretaria", "", "system.reset", map[string]any{"confirmation": "REINICIAR DESDE CERO"}); c != 409 {
		t.Fatal("reset without backup", c)
	}
	if c, _ := call610(t, s, "departamental", "", "system.backup", nil); c != 403 {
		t.Fatal("wrong reset role", c)
	}
	code, p := call610(t, s, "secretaria", "", "system.backup", nil)
	if code != 200 || p["backupToken"] == nil {
		t.Fatal(code, p)
	}
	backup := legacyMap(p["backup"])
	if backup["database"] == nil || backup["files"] == nil || legacyMap(backup["store610"])["state"] != nil {
		t.Fatal("backup data")
	}
	v := ministryFixture610()
	if c, _ := call610(t, s, "admin", "", "ministry.save", v); c != 200 {
		t.Fatal(c)
	}
	if c, _ := call610(t, s, "secretaria", "", "system.reset", map[string]any{"confirmation": "REINICIAR DESDE CERO", "backupToken": p["backupToken"]}); c != 409 {
		t.Fatal("stale backup accepted", c)
	}
	code, p = call610(t, s, "secretaria", "", "system.backup", nil)
	if code != 200 {
		t.Fatal(code, p)
	}
	if c, p := call610(t, s, "secretaria", "", "system.reset", map[string]any{"confirmation": "REINICIAR DESDE CERO", "backupToken": p["backupToken"]}); c != 200 {
		t.Fatal(c, p)
	}
	st, _ = s.loadLegacy()
	if len(legacyUsers(st.State.Database)) != 0 {
		t.Fatal("active data remains")
	}
	if legacyCanonicalRestoreGeneration(st.State) != 2 {
		t.Fatal("old writes not invalidated")
	}
	if _, e := os.Stat(filepath.Join(filepath.Dir(s.path), "respaldos", stringValue(p["fileName"]))); e != nil {
		t.Fatal("backup deleted", e)
	}
	s, db = fixture550(t)
	musicEnsure610(db)
	st, _ = s.loadLegacy()
	st.State.Database = db
	s.saveLegacy(st)
	year := int(numericValue(musicData550(db)["managementYear610"])) + 1
	code, p = call610(t, s, "admin", "", "system.year", map[string]any{"year": year, "confirmation": "INICIAR GESTION " + parseYearString610(year), "keepUsers": true, "keepTemplates": true, "keepMinistries": true, "carryUserIds": []any{"p"}})
	if code != 200 {
		t.Fatal(code, p)
	}
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	if len(musicArchiveIndex610(db)) != 1 {
		t.Fatal("archive missing")
	}
	for _, value := range mapField(reviewProgressContainer(db, "p", musicProgram553), "status") {
		if value == "approved" {
			t.Fatal("approval copied")
		}
	}
	projection := musicView550(db, legacyUser(db, "admin"))
	if projection["archiveData610"] != nil || projection["maintenanceBackup610"] != nil {
		t.Fatal("archive data leaked into live sync")
	}
}
func parseYearString610(year int) string {
	return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).Format("2006")
}
func Test610AuthStateTamperingAndResetSingleUse(t *testing.T) {
	s, _ := fixture550(t)
	value := authState610{State: "valid", Nonce: "nonce", Verifier: "proof", Expires: time.Now().Add(time.Minute).Unix()}
	cookie := s.authCookie610(value)
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "musica_oidc610", Value: cookie})
	if _, ok := s.readAuthCookie610(r); !ok {
		t.Fatal("valid cookie rejected")
	}
	r = httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "musica_oidc610", Value: cookie + "tampered"})
	if _, ok := s.readAuthCookie610(r); ok {
		t.Fatal("tampered cookie accepted")
	}
	st, _ := s.loadLegacy()
	u := legacyUser(st.State.Database, "p")
	proof := "single-use-password-token"
	s.saveReset610([]resetRecord610{{UserID: "p", Hash: hashToken(proof), PasswordStamp: passwordStamp610(u), Generation: 1, Expires: time.Now().Add(time.Minute).Unix()}})
	run := func() int {
		b, _ := json.Marshal(map[string]any{"token": proof, "password": "NuevaClave610!"})
		r := httptest.NewRequest("POST", "/api/auth610/reset", bytes.NewReader(b))
		w := httptest.NewRecorder()
		s.resetPassword610(w, r)
		return w.Code
	}
	if code := run(); code != 200 {
		t.Fatal("valid reset", code)
	}
	if code := run(); code != 400 {
		t.Fatal("token replay", code)
	}
	st, _ = s.loadLegacy()
	if !verifyRemotePasswordRecord("NuevaClave610!", legacyMap(legacyUser(st.State.Database, "p")["password"])) {
		t.Fatal("password not saved")
	}
	if _, e := googleClaims610("bad.token.claim", "nonce"); e == nil {
		t.Fatal("unverified Google identity accepted")
	}
}

func Test610MandatoryBackupRestoresAfterReset(t *testing.T) {
	s, db := fixture550(t)
	s.configPath = filepath.Join(t.TempDir(), "server.json")
	legacyUser(db, "admin")["password"] = tutorPassword556()
	musicEnsure610(db)
	st, _ := s.loadLegacy()
	st.State.Database = db
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	code, p := call610(t, s, "secretaria", "", "system.backup", nil)
	if code != 200 {
		t.Fatal(code, p)
	}
	saved := p["backup"]
	code, p = call610(t, s, "secretaria", "", "system.reset", map[string]any{"confirmation": "REINICIAR DESDE CERO", "backupToken": p["backupToken"]})
	if code != 200 {
		t.Fatal(code, p)
	}
	body, _ := json.Marshal(map[string]any{"key": p["setupKey"], "payload": saved, "confirmation": "IMPORTAR MUSICA"})
	r := httptest.NewRequest("POST", "/api/music-data/import", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.musicDataEndpoint(w, r, "/api/music-data/import")
	if w.Code != 200 {
		t.Fatal("mandatory backup was not restorable", w.Code, w.Body.String())
	}
	st, _ = s.loadLegacy()
	if !verifyRemotePasswordRecord("Prueba123!", legacyMap(legacyUser(st.State.Database, "admin")["password"])) {
		t.Fatal("administrator password lost")
	}
	if len(st.State.Files) == 0 || len(musicMissingFiles(legacyMap(st.State.Database), st.State.Files)) != 0 {
		t.Fatal("certificate attachment lost on restore")
	}
	if legacyCanonicalRestoreGeneration(st.State) != 3 {
		t.Fatal("restore generation not advanced")
	}
}
