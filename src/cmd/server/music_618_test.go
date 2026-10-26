package main

import (
	"bytes"
	"encoding/base64"
	"math"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test618ParticipationIsVisibleWithAssociatedTemplate(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	d := musicData550(db)
	ensureMapField(d, "events")["zero-hours"] = map[string]any{"id": "zero-hours", "name": "Curso", "date": "2030-01-01", "certificateKind": "leadership", "templateId610": "leadership", "hours": 0., "participants": map[string]any{"p": map[string]any{"attended": true}}}
	st, _ := s.loadLegacy()
	st.State.Database = db
	if !musicAutoCertificates610(&st) {
		t.Fatal("attendance was not reconciled")
	}
	c := legacyMap(mapField(d, "certificates")["p:leadership:zero-hours"])
	if c == nil || c["fileId610"] == nil || !musicEligible550(db, "p", "leadership", "zero-hours") {
		t.Fatal("explicit participation hidden by optional hours or date")
	}
	if mapField(musicView550(db, legacyUser(db, "p")), "certificates")[stringValue(c["id"])] == nil {
		t.Fatal("aspirant cannot see issued event certificate")
	}
	_ = s.saveLegacy(st)
	r := httptest.NewRequest("GET", "/api/music/v610?actorId=p&resource=certificate&token="+stringValue(c["token"]), nil)
	r.Header.Set("Authorization", "Bearer p")
	w := httptest.NewRecorder()
	s.music610(w, r)
	if w.Code != 200 || !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) {
		t.Fatal("certificate download", w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Disposition"), "filename") {
		t.Fatal("download filename missing")
	}
	ev := legacyMap(mapField(d, "events")["zero-hours"])
	legacyMap(mapField(ev, "participants")["p"])["attended"] = false
	if musicEligible550(db, "p", "leadership", "zero-hours") {
		t.Fatal("nonparticipant retained eligibility")
	}
}

func Test618FinalDateComesFromApprovalAndSurvivesReprint(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	d := musicData550(db)
	ensureMapField(d, "templates")["final"] = map[string]any{"name": "Líder de Música", "fields": []any{map[string]any{"text": "{Nombre} {fecha}", "x": 10., "y": 30., "width": 80., "size": 15., "font": "DejaVu Sans", "align": "left", "color": "#000000"}}}
	st, _ := s.loadLegacy()
	st.State.Database = db
	musicAutoCertificates610(&st)
	if mapField(d, "certificates")["p:final:"] != nil {
		t.Fatal("issued before approval")
	}
	ensureMapField(d, "workflow")["p"] = map[string]any{"state": "approved", "approvedAt": "2026-02-03T21:44:00Z", "approvedBy": "departamental"}
	musicAutoCertificates610(&st)
	c := legacyMap(mapField(d, "certificates")["p:final:"])
	if c["approvalDate618"] != "2026-02-03T21:44:00Z" {
		t.Fatal("wrong approval date", c["approvalDate618"])
	}
	pdf := musicCertificateBytes610(c, musicHydrateFonts613(legacyMap(mapField(d, "templates")["final"]), st.State.Files))
	if !bytes.Contains(pdf, []byte("3 de febrero de 2026")) || bytes.Contains(pdf, []byte("{Nombre}")) {
		t.Fatal("date or case-insensitive marker failed")
	}
	token := c["token"]
	delete(c, "approvalDate618")
	if !musicRefreshFinal618(&st, db, c) || c["token"] != token {
		t.Fatal("legacy PDF repair changed identity")
	}
	if musicRefreshFinal618(&st, db, c) {
		t.Fatal("same certificate regenerated on every read")
	}
}

func Test618RichTextOpacityAndProtectedVerification(t *testing.T) {
	f := map[string]any{"text": "Hola {Nombre}", "font": "DejaVu Sans", "x": 10., "y": 20., "width": 80., "size": 18., "color": "#000000", "align": "left", "opacity618": .4, "lineSpacing618": 1.8, "runs618": []any{map[string]any{"text": "Hola ", "bold": false}, map[string]any{"text": "{Nom", "size": 26., "bold": true, "underline": true}, map[string]any{"text": "bre}", "italic": true}}}
	clean, valid := musicTemplateFields592([]any{f})
	if !valid {
		t.Fatal("valid styled selection rejected")
	}
	tplt := map[string]any{"name": "Prueba", "fields": clean, "verification618": map[string]any{"x": 12., "y": 88., "text": "borrar", "opacity618": 0.}}
	p := newMusicPDF610()
	commands := p.richText618(legacyMap(clean[0]), tplt, map[string]string{"nombre": "Ana"})
	if !strings.Contains(commands, "(Ana)") || strings.Contains(commands, "{Nom") || !strings.Contains(commands, "26.00 Tf") {
		t.Fatal("styled marker expansion failed", commands)
	}
	pdf := musicCertificateBytes610(map[string]any{"fullName": "Ana", "token": "PROTEGIDO", "issuedAt": "2026-01-01"}, tplt)
	if !bytes.Contains(pdf, []byte("/ca 0.400")) || !bytes.Contains(pdf, []byte("PROTEGIDO")) {
		t.Fatal("opacity or protected code missing")
	}
	if !bytes.Contains(pdf, []byte("/ExtGState <<")) {
		t.Fatal("transparency resource missing")
	}
	out := map[string]any{}
	if err := musicTemplateSettings618(tplt, out); err != nil {
		t.Fatal(err)
	}
	if len(mapField(out, "verification618")) != 2 {
		t.Fatal("protected verification accepted editable content")
	}
	for _, bad := range []map[string]any{{"opacity618": math.NaN()}, {"lineSpacing618": 0.}, {"runs618": []any{map[string]any{"text": "different"}}}} {
		copy := cloneMap610(f)
		for k, v := range bad {
			copy[k] = v
		}
		if _, valid := musicTemplateFields592([]any{copy}); valid {
			t.Fatal("invalid styled input accepted")
		}
	}
}

func Test618InstallerDefaultsAndEmbeddedFontsArePortable(t *testing.T) {
	db := map[string]any{}
	if !musicDefaults618(db) {
		t.Fatal("missing fresh defaults")
	}
	templates := mapField(musicData550(db), "templates")
	for _, id := range []string{"final", "leadership", "beliefs"} {
		item := legacyMap(templates[id])
		if stringValue(item["dataUrl"]) != "" || len(arrayField(item, "images610")) != 0 || len(arrayField(item, "fields")) == 0 {
			t.Fatal("configured text missing or unwanted artwork shipped", id)
		}
		if item["updatedBy"] != nil || item["ownerId"] != nil {
			t.Fatal("private record metadata included")
		}
	}
	st := legacyStore{}
	st.State.Database = db
	if !musicInstallAssets618(&st) {
		t.Fatal("bundled fonts were not added to full-backup assets")
	}
	item := legacyMap(templates["beliefs"])
	fontID := stringValue(legacyMap(arrayField(item, "fields")[0])["fontAsset613"])
	asset := musicFontFile613(st.State.Files, fontID)
	raw := musicFontRaw613(asset)
	family, _, _ := musicFontName617(raw)
	if family != "Georgia" {
		t.Fatal("exact configured Georgia font missing", family)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.SplitN(stringValue(asset["dataUrl"]), ",", 2)[1])
	if err != nil || len(decoded) == 0 {
		t.Fatal("invalid portable font asset")
	}
	before := item["dataUrl"]
	musicDefaults618(db)
	if legacyMap(templates["beliefs"])["dataUrl"] != before {
		t.Fatal("existing customization overwritten")
	}
}

func Test618ManualDeletionRemovesApplicantViewHistoryAndVerification(t *testing.T) {
	for _, role := range []string{"admin", "departamental", "secretaria", "coordinador_musica", "administrativo", "regional", "tutor", "p"} {
		t.Run(role, func(t *testing.T) {
			s, db := fixture550(t)
			musicEnsure610(db)
			d := musicData550(db)
			legacyUser(db, "p")["tutorId"] = role
			ensureMapField(d, "templates")["sample"] = map[string]any{"name": "Certificado prueba"}
			ensureMapField(d, "events")["ev"] = map[string]any{"id": "ev", "name": "Evento borrado", "templateId610": "sample", "participants": map[string]any{"p": map[string]any{"attended": true}}}
			st, _ := s.loadLegacy()
			st.State.Database = db
			c := musicIssue610(&st, db, "p", "event", "ev", "", "sample", "2026-10-21T12:00:00Z")
			req := musicRequest610{Value: map[string]any{"id": c["id"], "token": c["token"]}}
			if musicDeleteCertificate618(&st, db, legacyUser(db, role), req) == nil {
				t.Fatal("certificate removed while event still exists")
			}
			musicRevokeEvent610(d, "ev", "", "Evento eliminado", "2026-10-21T12:10:00Z")
			delete(mapField(d, "events"), "ev")
			d["certificateHistory610"] = []any{cloneAny(c)}
			if len(mapField(d, "certificates")) != 1 {
				t.Fatal("event deletion automatically deleted certificate")
			}
			err := musicDeleteCertificate618(&st, db, legacyUser(db, role), req)
			allowed := role == "admin" || role == "departamental" || role == "secretaria" || role == "coordinador_musica" || role == "administrativo"
			if (err == nil) != allowed {
				t.Fatal("wrong role permission", role, err)
			}
			if !allowed {
				return
			}
			if len(mapField(d, "certificates")) != 0 || len(arrayField(d, "certificateHistory610")) != 0 {
				t.Fatal("deleted certificate remains in applicant history")
			}
			if len(arrayField(musicRegistryState610(db, legacyUser(db, "p"), st), "certificates")) != 0 {
				t.Fatal("certificate visible to applicant")
			}
			if musicView550(db, legacyUser(db, "p"))["certificateDeletions618"] != nil {
				t.Fatal("internal audit exposed in applicant view")
			}
			musicAutoCertificates610(&st)
			if mapField(d, "certificates")[stringValue(c["id"])] != nil {
				t.Fatal("deleted event certificate regenerated")
			}
			_ = s.saveLegacy(st)
			r := httptest.NewRequest("GET", "/api/music/v610?actorId=p&resource=certificate&token="+stringValue(c["token"]), nil)
			r.Header.Set("Authorization", "Bearer p")
			w := httptest.NewRecorder()
			s.music610(w, r)
			if w.Code != 404 {
				t.Fatal("deleted certificate still available", w.Code)
			}
		})
	}
}

func Test618TemplateChangesUpdateAlreadyViewedCertificateWithSameToken(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	d := musicData550(db)
	tpl := map[string]any{"name": "Curso original", "id": "sample", "fields": []any{map[string]any{"text": "ORIGINAL", "x": 10., "y": 20., "width": 80., "size": 15., "font": "DejaVu Sans", "color": "#000000", "align": "left"}}}
	ensureMapField(d, "templates")["sample"] = tpl
	ensureMapField(d, "events")["ev"] = map[string]any{"date": "2026-01-01", "templateId610": "sample", "participants": map[string]any{"p": map[string]any{"attended": true}}}
	st, _ := s.loadLegacy()
	st.State.Database = db
	c := musicIssue610(&st, db, "p", "event", "ev", "", "sample", "2026-02-03T12:00:00Z")
	musicAutoCertificates610(&st)
	token, issued, oldID, oldHash := c["token"], c["issuedAt"], c["fileId610"], c["sha256610"]
	// A closed management year shares the previous PDF and must keep it.
	archiveDB := map[string]any{"music550": map[string]any{"certificates": map[string]any{"old": cloneAny(c)}}}
	ensureMapField(d, "archiveData610")["previous"] = map[string]any{"database": archiveDB}
	legacyMap(arrayField(tpl, "fields")[0])["text"] = "NUEVO DISENO"
	tpl["name"] = "Curso actualizado"
	if !musicRefreshLinked618(&st, db) {
		t.Fatal("linked certificate did not refresh")
	}
	if c["token"] != token || c["issuedAt"] != issued || c["sha256610"] == oldHash || c["fileId610"] == oldID {
		t.Fatal("token/date changed or archive file overwritten")
	}
	current := []byte{}
	old := []byte{}
	for _, raw := range st.State.Files {
		f := legacyMap(raw)
		if f["id"] == c["fileId610"] {
			current = musicFontRaw613(f)
		}
		if f["id"] == oldID {
			old = musicFontRaw613(f)
		}
	}
	if !bytes.Contains(current, []byte("NUEVO DISENO")) || !bytes.Contains(old, []byte("ORIGINAL")) {
		t.Fatal("current or archived PDF incorrect")
	}
	if musicRefreshLinked618(&st, db) {
		t.Fatal("unchanged certificate regenerated on every read")
	}
	_ = s.saveLegacy(st)
	r := httptest.NewRequest("GET", "/api/music/v610?actorId=p&resource=certificate&token="+stringValue(token), nil)
	r.Header.Set("Authorization", "Bearer p")
	w := httptest.NewRecorder()
	s.music610(w, r)
	if w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte("NUEVO DISENO")) {
		t.Fatal("aspirant received stale PDF after prior viewing", w.Code)
	}
}

func Test618CertificateDatesStartIn2026(t *testing.T) {
	for _, date := range []string{"2025-12-31", "2020-02-30", "", "1999-01-01"} {
		if musicFrom2026_618(date) {
			t.Fatal("old or invalid date accepted", date)
		}
	}
	if !musicFrom2026_618("2026-01-01") || !musicFrom2026_618("2027-10-21T12:00:00Z") {
		t.Fatal("supported management year rejected")
	}
	s, db := fixture550(t)
	musicEnsure610(db)
	ev := map[string]any{"id": "ev", "name": "Anterior", "date": "2025-12-31", "certificateKind": "leadership", "templateId610": "leadership", "hours": 10., "participants": map[string]any{"p": map[string]any{"attended": true}}}
	ensureMapField(musicData550(db), "events")["ev"] = ev
	if musicEligible550(db, "p", "leadership", "ev") {
		t.Fatal("pre-2026 event counted for a certificate")
	}
	st, _ := s.loadLegacy()
	st.State.Database = db
	_ = s.saveLegacy(st)
	response := call550(t, s, "admin", "", "event", map[string]any{"name": "Anterior", "date": "2025-12-31", "time": "10:00", "place": "Sala"}, false)
	if response.Code == 200 {
		t.Fatal("server allowed recording a pre-2026 certificate event")
	}
}
