package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"net/http/httptest"
	"strings"
	"testing"
)

func fontData613(t *testing.T) []byte {
	t.Helper()
	raw, err := musicFontFiles610.ReadFile("fonts610/" + musicFonts610[0].File)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func Test613SFNTValidation(t *testing.T) {
	raw := fontData613(t)
	font, err := parseMusicSFNT613(raw)
	if err != nil || len(font.Widths) != 224 || font.Widths[int('W')-32] <= font.Widths[int('i')-32] {
		t.Fatal("real font metrics", font.Name, err)
	}
	for _, value := range [][]byte{nil, []byte("not a font"), raw[:80], make([]byte, 4<<20+1)} {
		if _, err := parseMusicSFNT613(value); err == nil {
			t.Fatal("accepted corrupt font")
		}
	}
	bad := append([]byte{}, raw...)
	binary.BigEndian.PutUint32(bad[20:], 0xffffffff)
	if _, err := parseMusicSFNT613(bad); err == nil {
		t.Fatal("accepted table outside file")
	}
	license := append([]byte{}, raw...)
	for at := 12; at < 12+int(binary.BigEndian.Uint16(license[4:]))*16; at += 16 {
		if string(license[at:at+4]) == "OS/2" {
			offset := int(binary.BigEndian.Uint32(license[at+8:]))
			binary.BigEndian.PutUint16(license[offset+8:], 2)
		}
	}
	if _, err := parseMusicSFNT613(license); err == nil {
		t.Fatal("restricted font embedded")
	}
}
func Test613FontAssetsTemplateAndPDF(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	st, _ := s.loadLegacy()
	st.State.Database = db
	_ = s.saveLegacy(st)
	raw := fontData613(t)
	v := map[string]any{"dataUrl": "data:font/ttf;base64," + base64.StdEncoding.EncodeToString(raw)}
	if code, _ := call610(t, s, "p", "", "font.save", v); code != 403 {
		t.Fatal("aspirant font upload", code)
	}
	code, out := call610(t, s, "admin", "", "font.save", v)
	if code != 200 {
		t.Fatal(code, out)
	}
	id := stringValue(out["id"])
	if !strings.HasPrefix(id, "font613_") {
		t.Fatal(out)
	}
	_, _ = call610(t, s, "admin", "", "font.save", v)
	st, _ = s.loadLegacy()
	count := 0
	for _, raw := range st.State.Files {
		if legacyMap(raw)["kind"] == "music_font613" && legacyMap(raw)["id"] == id {
			count++
		}
	}
	if count != 1 {
		t.Fatal("duplicate font programs", count)
	}
	for _, actor := range []string{"admin", "p"} {
		r := httptest.NewRequest("GET", "/api/music/v610?actorId="+actor+"&resource=font&id="+id, nil)
		r.Header.Set("Authorization", "Bearer "+actor)
		w := httptest.NewRecorder()
		s.music610(w, r)
		if actor == "admin" {
			if w.Code != 200 || !bytes.Equal(w.Body.Bytes(), raw) {
				t.Fatal("font roundtrip", w.Code)
			}
		} else if w.Code != 403 {
			t.Fatal("aspirant font access", w.Code)
		}
	}
	f := map[string]any{"text": "Música impresa", "font": "Fuente del equipo", "fontAsset613": id, "x": 10., "y": 20., "width": 80., "size": 30., "color": "#173c76", "align": "center", "italic": true, "underline": true}
	template := map[string]any{"name": "Plantilla con fuentes", "paperSize617": "a4", "signatureDepartment": "Firma Departamental", "signatureCoordinator": "Firma Coordinación General", "fields": []any{f}, "backgroundRemoved613": true}
	signatures := musicSignatures613(nil)
	moved := map[string]any{}
	for key, signature := range signatures {
		signature["x"] = 10.
		signature["y"] = 70.
		signature["text"] = key
		moved[key] = signature
	}
	template["signatures613"] = moved
	code, out = call610(t, s, "admin", "", "template.save", template)
	if code != 200 {
		t.Fatal(code, out)
	}
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	var saved map[string]any
	for _, raw := range mapField(musicData550(db), "templates") {
		if legacyMap(raw)["name"] == template["name"] {
			saved = legacyMap(raw)
		}
	}
	if saved == nil {
		t.Fatal("template missing")
	}
	ids := map[string]bool{}
	musicFileIDs553(db, ids)
	if !ids[id] || !musicFileBelongs(db, musicFontFile613(st.State.Files, id), ids) {
		t.Fatal("backup lost font reference")
	}
	copy := musicHydrateFonts613(saved, st.State.Files)
	pdf := musicCertificateBytes610(map[string]any{"fullName": "Ana Música", "title": "Líder de Música", "issuedAt": "2026-10-21", "token": "prueba"}, copy)
	if !bytes.Contains(pdf, []byte("/BaseFont /")) || !bytes.Contains(pdf, []byte("1 0 0.220 1")) || bytes.Contains(pdf, []byte("138 115 m 338 115 l")) || !bytes.Contains(pdf, []byte("84.20 168.50 Tm (Firma Departamental)")) {
		t.Fatal("font/style/signature rendering")
	}
	if !bytes.Contains(pdf, []byte("/L1 ")) || bytes.Contains(pdf, []byte("25 25 792 545 re S")) {
		t.Fatal("custom font or blank background lost")
	}
	if mapField(saved, "fontPrograms613") != nil {
		t.Fatal("font program duplicated in template")
	}
	missing := cloneMap610(template)
	missing["fields"] = []any{map[string]any{"fontAsset613": "font613_" + strings.Repeat("a", 64)}}
	if code, _ = call610(t, s, "admin", "", "template.save", missing); code != 400 {
		t.Fatal("missing font accepted", code)
	}
	currentYear := int(numericValue(musicData550(legacyMap(st.State.Database))["managementYear610"]))
	code, out = call610(t, s, "admin", "", "system.year", map[string]any{"year": currentYear + 1, "confirmation": "INICIAR GESTION " + parseYearString610(currentYear+1), "keepTemplates": true})
	if code != 200 {
		t.Fatal("new year", code, out)
	}
	st, _ = s.loadLegacy()
	refs := map[string]bool{}
	musicFileIDs553(st.State.Database, refs)
	if !refs[id] || musicFontFile613(st.State.Files, id) == nil {
		t.Fatal("new year lost referenced font")
	}

}
func Test613QRConfigurationAndIssuedDocument(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	st, _ := s.loadLegacy()
	st.State.Database = db
	_ = s.saveLegacy(st)
	for _, value := range []string{"http://example.org", "https://user:password@example.org", "https://example.org/path", "https://example.org?q=x", "https://example.org#x", "javascript:alert(1)"} {
		if _, err := musicQRURL613(value); err == nil {
			t.Fatal("invalid URL", value)
		}
	}
	for _, actor := range []string{"departamental", "administrativo", "p"} {
		if code, _ := call610(t, s, actor, "", "qr.configure", map[string]any{"url": "https://musica.example.test"}); code != 403 {
			t.Fatal("unauthorized QR config", actor, code)
		}
	}
	code, out := call610(t, s, "secretaria", "", "qr.configure", map[string]any{"url": "https://musica.example.test/"})
	if code != 200 || out["effectiveURL"] != "https://musica.example.test" {
		t.Fatal(code, out)
	}
	_, out = call610(t, s, "admin", "", "ministry.save", ministryFixture610())
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	var old map[string]any
	for _, raw := range mapField(musicData550(db), "certificates") {
		old = legacyMap(raw)
		break
	}
	if old == nil || !strings.HasPrefix(stringValue(old["verificationURL610"]), "https://musica.example.test/api/music/certificate?") {
		t.Fatal("new QR URL", old, out)
	}
	file, hash, identity := old["fileId610"], old["sha256610"], old["token"]
	_, _ = call610(t, s, "admin", "", "qr.configure", map[string]any{"url": "https://nuevo.example.test"})
	st, _ = s.loadLegacy()
	current := legacyMap(mapField(musicData550(legacyMap(st.State.Database)), "certificates")[stringValue(old["id"])])
	if current["fileId610"] != file || current["sha256610"] == hash || current["token"] != identity || !strings.HasPrefix(stringValue(current["verificationURL610"]), "https://nuevo.example.test/") {
		t.Fatal("current QR did not follow the configured domain while keeping identity")
	}
	currentYear := int(numericValue(musicData550(legacyMap(st.State.Database))["managementYear610"]))
	code, out = call610(t, s, "admin", "", "system.year", map[string]any{"year": currentYear + 1, "confirmation": "INICIAR GESTION " + parseYearString610(currentYear+1)})
	if code != 200 {
		t.Fatal("new year", code, out)
	}
	st, _ = s.loadLegacy()
	if musicQRBase613(legacyMap(st.State.Database)) != "https://nuevo.example.test" {
		t.Fatal("new year lost public QR URL")
	}
}
