package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func get624(t *testing.T, s *Server, actor, resource string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	r := httptest.NewRequest("GET", "/api/music/v610?actorId="+actor+"&resource="+resource, nil)
	r.Header.Set("Authorization", "Bearer "+actor)
	w := httptest.NewRecorder()
	s.music610(w, r)
	p := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &p)
	return w, p
}

func Test624IndexesAreSmallReadOnlyAndAuthorized(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	tpl := legacyMap(mapField(musicData550(db), "templates")["leadership"])
	tpl["dataUrl"] = "data:image/png;base64," + strings.Repeat("A", 1<<20)
	st, _ := s.loadLegacy()
	st.State.Database = db
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	before, _ := s.loadLegacy()
	snapshot, _ := json.Marshal(before)
	w, p := get624(t, s, "admin", "templates")
	if w.Code != 200 || w.Body.Len() > 16000 || strings.Contains(w.Body.String(), "dataUrl") || len(mapField(p, "templates")) < 10 {
		t.Fatal("heavy template index", w.Code, w.Body.Len())
	}
	w, p = get624(t, s, "admin", "template&id=leadership")
	if w.Code != 200 || mapField(p, "template")["dataUrl"] != tpl["dataUrl"] {
		t.Fatal("detail did not retain artwork")
	}
	w, _ = get624(t, s, "p", "templates")
	if w.Code != 403 {
		t.Fatal("aspirant accessed template editing")
	}
	for _, resource := range []string{"certificates", "registry", "templates", "followup"} {
		w, _ = get624(t, s, "admin", resource)
		if w.Code != 200 {
			t.Fatal(resource, w.Code)
		}
	}
	after, _ := s.loadLegacy()
	current, _ := json.Marshal(after)
	if !bytes.Equal(snapshot, current) {
		t.Fatal("read changed records, files or server version")
	}
}

func Test624TrashPurgePermissionsAndEvidence(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	st, _ := s.loadLegacy()
	st.State.Database = db
	_ = s.saveLegacy(st)
	code, p := call610(t, s, "admin", "", "ministry.save", ministryFixture610())
	if code != 200 {
		t.Fatal(code, p)
	}
	id := registryID610(t, p)
	if code, _ = call610(t, s, "admin", "", "ministry.purge624", map[string]any{"id": id, "confirmation": "ELIMINAR DEFINITIVAMENTE"}); code != 403 {
		t.Fatal("purged active record")
	}
	if code, p = call610(t, s, "admin", "p", "ministry.link", map[string]any{"id": id}); code != 200 {
		t.Fatal(p)
	}
	if code, p = call610(t, s, "admin", "", "ministry.delete", map[string]any{"id": id, "confirmation": "ELIMINAR MINISTERIO"}); code != 200 {
		t.Fatal(p)
	}
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	m := legacyMap(mapField(musicData550(db), "ministries610")[id])
	c := legacyMap(mapField(musicData550(db), "certificates")["ministry-registration:"+id])
	fileID := stringValue(c["fileId610"])
	// A closed year is independent evidence and must survive active-registry purge.
	ensureMapField(musicData550(db), "archiveData610")["2025"] = map[string]any{"database": map[string]any{"music550": map[string]any{"certificates": map[string]any{"historical": cloneAny(c)}}}}
	_ = s.saveLegacy(st)
	v := map[string]any{"id": id, "confirmation": "ELIMINAR DEFINITIVAMENTE", "expectedUpdatedAt": m["updatedAt"]}
	if code, _ = call610(t, s, "p", "", "ministry.purge624", v); code != 403 {
		t.Fatal("aspirant purged record")
	}
	v["confirmation"] = "wrong"
	if code, _ = call610(t, s, "admin", "", "ministry.purge624", v); code != 403 {
		t.Fatal("missing confirmation accepted")
	}
	v["confirmation"] = "ELIMINAR DEFINITIVAMENTE"
	if code, p = call610(t, s, "admin", "", "ministry.purge624", v); code != 200 {
		t.Fatal(code, p)
	}
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	d := musicData550(db)
	if mapField(d, "ministries610")[id] != nil || len(mapField(d, "ministryLinks610")) != 0 {
		t.Fatal("purged ministry remains linked")
	}
	for _, raw := range mapField(d, "certificates") {
		if legacyMap(raw)["ministryId"] == id {
			t.Fatal("active certificate retained")
		}
	}
	for _, raw := range arrayField(d, "certificateHistory610") {
		if legacyMap(raw)["ministryId"] == id {
			t.Fatal("active certificate history retained")
		}
	}
	found := false
	for _, raw := range st.State.Files {
		found = found || legacyMap(raw)["id"] == fileID
	}
	if !found {
		t.Fatal("closed-year evidence deleted")
	}
	if next := musicNextCode622(d); next != "MBOS-MMUS-0002" {
		t.Fatal("registration number reused", next)
	}
}

func Test624CoordinatorHasAdminAuthority(t *testing.T) {
	s, db := fixture550(t)
	u := map[string]any{"id": "coord624", "role": "coordinador_sgm", "fullName": "Coordinador SGM", "active": true}
	db["users"] = append(arrayField(db, "users"), u)
	s.remoteSessions[hashToken("coord624")] = remoteSession{UserID: "coord624", Role: canonicalRole548("coordinador_sgm"), ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339)}
	musicEnsure610(db)
	st, _ := s.loadLegacy()
	st.State.Database = db
	_ = s.saveLegacy(st)
	if legacyRole(u) != "admin" || !musicMaintenance610(u) || !managementProgram555(db, u, musicProgram553) {
		t.Fatal("coordinator is not an admin")
	}
	for _, key := range musicKeys602 {
		if musicPermit602(db, u, key) != musicPermit602(db, legacyUser(db, "admin"), key) {
			t.Fatal("permission differs from admin", key)
		}
	}
	if code, p := call610(t, s, "coord624", "", "ministry.save", ministryFixture610()); code != 200 {
		t.Fatal(code, p)
	}
	w, _ := get624(t, s, "coord624", "templates")
	if w.Code != 200 {
		t.Fatal("cannot manage templates", w.Code)
	}
}

func Test624PortableSeedLayoutAndFonts(t *testing.T) {
	db := map[string]any{}
	musicEnsure610(db)
	st := legacyStore{}
	st.State.Database = db
	musicInstallAssets618(&st)
	templates := mapField(musicData550(db), "templates")
	for id, raw := range templates {
		tpl := legacyMap(raw)
		if stringValue(tpl["dataUrl"]) != "" || len(arrayField(tpl, "images610")) != 0 {
			t.Fatal("artwork was packaged", id)
		}
		if err := musicValidateFonts613(tpl, st.State.Files); err != nil {
			t.Fatal(id, err)
		}
		refs := map[string]bool{}
		musicFileIDs553(tpl, refs)
		for fid := range refs {
			if strings.HasPrefix(fid, "font613_") {
				if _, err := parseMusicSFNT613(musicFontRaw613(musicFontFile613(st.State.Files, fid))); err != nil {
					t.Fatal(id, err)
				}
			}
		}
	}
	field := legacyMap(arrayField(legacyMap(templates["leadership"]), "fields")[0])
	if field["font"] != "Great Vibes" || numericValue(field["x"]) != 13.865545889126453 || numericValue(field["y"]) != 31.51207674408865 {
		t.Fatal("configured position or font changed")
	}
	regular, narrow := false, false
	for _, f := range musicBundledCatalog624() {
		regular = regular || f.Family == "Arial" && !f.Bold && !f.Italic
		narrow = narrow || f.Family == "Arial Narrow"
	}
	if !regular || !narrow {
		t.Fatal("Arial and Arial Narrow must both be available")
	}
	before := cloneAny(templates)
	next, fonts := musicResetDesign619(db, st.State.Files)
	musicEnsure610(next)
	if !reflect.DeepEqual(before, mapField(musicData550(next), "templates")) || len(fonts) == 0 {
		t.Fatal("reset discarded certificate configuration")
	}
}

func Test624LocalSetupDoesNotExposeKeyRemotely(t *testing.T) {
	s, _ := fixture550(t)
	s.configPath = filepath.Join(t.TempDir(), "server.json")
	_ = os.WriteFile(s.installationKeyPath(), []byte("one-use-key"), 0600)
	for _, tc := range []struct {
		host, peer, origin, forwarded string
		want                          bool
	}{
		{"127.0.0.1:8796", "127.0.0.1:1234", "", "", true},
		{"localhost:8796", "127.0.0.1:1234", "http://localhost:8796", "", true},
		{"example.org", "127.0.0.1:1234", "", "", false},
		{"127.0.0.1:8796", "192.0.2.2:1234", "", "", false},
		{"127.0.0.1:8796", "127.0.0.1:1234", "https://other.example", "", false},
		{"127.0.0.1:8796", "127.0.0.1:1234", "", "192.0.2.2", false},
	} {
		r := httptest.NewRequest("GET", "http://"+tc.host+"/api/installation/status", nil)
		r.RemoteAddr = tc.peer
		r.Header.Set("Origin", tc.origin)
		r.Header.Set("X-Forwarded-For", tc.forwarded)
		if got := s.localSetupKey624(r) != ""; got != tc.want {
			t.Fatal(tc, got)
		}
	}
}
