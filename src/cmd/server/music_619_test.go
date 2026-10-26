package main

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func Test619ResetKeepsDesignsWithoutIssuedData(t *testing.T) {
	s, db := fixture550(t)
	s.configPath = filepath.Join(t.TempDir(), "server.json")
	musicEnsure610(db)
	d := musicData550(db)
	template := certificateTemplate617()
	template["dataUrl"] = photo592()
	template["signatureDepartment"] = "Nombre largo\nDepartamental"
	legacyMap(arrayField(template, "fields")[0])["fontAsset613"] = "custom-font"
	template["images610"] = []any{map[string]any{"dataUrl": photo592(), "x": 30., "y": 30., "width": 10., "height": 10., "rotation619": 32.}}
	mapField(d, "templates")["custom619"] = template
	st, _ := s.loadLegacy()
	st.State.Database = db
	st.State.Files = append(st.State.Files, map[string]any{"id": "personal-pdf", "kind": "requirement_document", "ownerId": "p", "dataUrl": "data:application/pdf;base64,JVBERi0="}, map[string]any{"id": "custom-font", "kind": "music_font613", "ownerId": "p", "dataUrl": "data:font/ttf;base64,YWJj"})
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	code, backup := call610(t, s, "secretaria", "", "system.backup", nil)
	if code != 200 {
		t.Fatal(code)
	}
	code, resetResult := call610(t, s, "secretaria", "", "system.reset", map[string]any{"confirmation": "REINICIAR DESDE CERO", "backupToken": backup["backupToken"]})
	if code != 200 {
		t.Fatal(code)
	}
	st, _ = s.loadLegacy()
	next := legacyMap(st.State.Database)
	nd := musicData550(next)
	tpl := legacyMap(mapField(nd, "templates")["custom619"])
	if tpl["dataUrl"] != template["dataUrl"] || tpl["signatureDepartment"] != template["signatureDepartment"] || numericValue(legacyMap(arrayField(tpl, "images610")[0])["rotation619"]) != 32 {
		t.Fatal("design settings lost")
	}
	if len(legacyUsers(next)) != 0 || len(mapField(nd, "certificates")) != 0 || len(arrayField(nd, "certificateHistory610")) != 0 || len(mapField(nd, "templateVersions610")) != 0 {
		t.Fatal("issued data survived reset")
	}
	found := false
	for _, raw := range st.State.Files {
		f := legacyMap(raw)
		if f["kind"] != "music_font613" || f["ownerId"] != nil {
			t.Fatal("personal data retained", f["id"])
		}
		found = found || f["id"] == "custom-font"
	}
	if !found {
		t.Fatal("custom font lost")
	}
	body, _ := json.Marshal(map[string]any{"key": resetResult["setupKey"], "fullName": "Nuevo responsable", "username": "nuevo", "password": "Nueva619Segura!"})
	w := httptest.NewRecorder()
	s.installationEndpoint(w, httptest.NewRequest("POST", "/api/installation/create", bytes.NewReader(body)), "/api/installation/create")
	if w.Code != 200 {
		t.Fatal("new admin setup", w.Code, w.Body.String())
	}
	st, _ = s.loadLegacy()
	nd = musicData550(legacyMap(st.State.Database))
	if mapField(nd, "templates")["custom619"] == nil || len(legacyUsers(st.State.Database)) != 1 || legacyCanonicalRestoreGeneration(st.State) != 2 || musicFontFile613(st.State.Files, "custom-font") == nil {
		t.Fatal("creating administrator discarded retained designs or reset generation")
	}
}

func Test619QRBackfillAndDomainChangesKeepIdentity(t *testing.T) {
	t.Setenv("MUSICA_PUBLIC_URL", "")
	s, db := fixture550(t)
	musicEnsure610(db)
	d := musicData550(db)
	ev := map[string]any{"id": "qr619", "name": "Evento", "date": "2026-10-26", "templateId610": "leadership", "certificateKind": "leadership", "participants": map[string]any{"p": map[string]any{"attended": true}}}
	ensureMapField(d, "events")["qr619"] = ev
	st, _ := s.loadLegacy()
	st.State.Database = db
	musicAutoCertificates610(&st)
	c := legacyMap(mapField(d, "certificates")["p:leadership:qr619"])
	identity := c["token"]
	if !bytes.Contains(musicCertificateBytes610(c, map[string]any{}), []byte("64.000 64.000 re f")) {
		t.Fatal("code QR missing without domain")
	}
	for _, base := range []string{"https://uno.example.org", "https://dos.example.org"} {
		d["publicQRURL613"] = base
		if !musicRefreshLinked618(&st, db) || c["token"] != identity || c["verificationURL610"] != base+"/api/music/certificate?token="+stringValue(identity) {
			t.Fatal("QR was not refreshed without changing token")
		}
		if musicRefreshLinked618(&st, db) {
			t.Fatal("unchanged certificate regenerated")
		}
	}
	_ = s.saveLegacy(st)
	before, _ := s.loadLegacy()
	r := httptest.NewRequest("GET", "/api/music/v610?actorId=p&resource=certificate&token="+stringValue(identity), nil)
	r.Header.Set("Authorization", "Bearer p")
	w := httptest.NewRecorder()
	s.music610(w, r)
	after, _ := s.loadLegacy()
	if w.Code != 200 || before.State.ServerVersion != after.State.ServerVersion || !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) {
		t.Fatal("view must be read-only", w.Code)
	}
}

func Test619DateExplicitSignatureAndImageRotation(t *testing.T) {
	if musicDate619("2026-10-26T22:00:00Z") != "26 de octubre de 2026" {
		t.Fatal("Spanish date")
	}
	p := newMusicPDF610()
	f := map[string]any{"x": 20., "y": 60., "width": 15., "size": 16., "align": "center", "font": "Arial"}
	label := "Alejandro Nombre Apellido Muy Largo\nCoordinador General"
	cmd := p.templateText613(musicSignature619(f), label, map[string]any{})
	if strings.Count(cmd, " Tj ET") != 2 || !strings.Contains(cmd, "Alejandro Nombre Apellido Muy Largo") {
		t.Fatal("signature wrapped despite explicit line break", cmd)
	}
	f["text"] = label
	f["runs618"] = []any{map[string]any{"text": "Alejandro Nombre ", "bold": true}, map[string]any{"text": "Apellido Muy Largo\nCoordinador General"}}
	cmd = p.richText618(musicSignature619(f), map[string]any{}, nil)
	if !strings.Contains(cmd, "(Apellido Muy Largo)") {
		t.Fatal("rich signature wrapped", cmd)
	}
	im := map[string]any{"x": 30., "y": 30., "width": 20., "height": 15., "rotation619": 45.}
	if !musicImageBounds619(im, nil) || !strings.Contains(musicImageCommands619(im, nil, "Logo0", ""), "0.707107 -0.707107") {
		t.Fatal("clockwise centered rotation")
	}
	im["x"] = 0.
	im["y"] = 0.
	if musicImageBounds619(im, nil) {
		t.Fatal("rotated image allowed outside page")
	}
	for _, angle := range []float64{math.NaN(), 181, -181} {
		v := certificateTemplate617()
		v["images610"] = []any{map[string]any{"dataUrl": photo592(), "x": 30., "y": 30., "width": 20., "height": 15., "rotation619": angle}}
		_, db := fixture550(t)
		musicEnsure610(db)
		if musicTemplateMutation610(db, legacyUser(db, "admin"), musicRequest610{Action: "template.save", Value: v}) == nil {
			t.Fatal("invalid rotation accepted")
		}
	}
}
