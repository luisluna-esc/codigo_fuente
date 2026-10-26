package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func Test620SignaturesBecomeOrdinaryTextWithoutLoss(t *testing.T) {
	f := map[string]any{"x": 11., "y": 80., "width": 30., "size": 12., "font": "Georgia", "align": "center", "color": "#123456", "bold": true, "opacity618": .6, "runs618": []any{map[string]any{"text": "Ana", "bold": true}, map[string]any{"text": "\nDepartamental", "italic": true}}}
	tpl := map[string]any{"fields": []any{map[string]any{"text": "CUERPO {nombre}"}}, "signatureDepartment": "Ana\nDepartamental", "signatureCoordinator": "Luis\nCoordinador", "signatures613": map[string]any{"department": f}}
	if !musicTextOnly620(tpl) {
		t.Fatal("migration skipped")
	}
	fields := arrayField(tpl, "fields")
	if len(fields) != 3 || legacyMap(fields[1])["text"] != "Ana\nDepartamental" || legacyMap(fields[1])["font"] != "Georgia" || legacyMap(fields[1])["color"] != "#123456" || legacyMap(fields[1])["x"] != 11. {
		t.Fatal("appearance changed")
	}
	if legacyMap(fields[1])["fitLines620"] != true || len(arrayField(legacyMap(fields[1]), "runs618")) != 2 {
		t.Fatal("manual lines or rich style lost")
	}
	if tpl["signatures613"] != nil || tpl["signatureDepartment"] != nil || len(musicSignatures613(tpl)) != 0 {
		t.Fatal("special signatures remain")
	}
	before, _ := json.Marshal(tpl)
	if musicTextOnly620(tpl) {
		t.Fatal("migration repeated")
	}
	after, _ := json.Marshal(tpl)
	if !bytes.Equal(before, after) {
		t.Fatal("repeated migration modified layout")
	}
	tpl["fields"] = []any{fields[0]}
	musicTextOnly620(tpl)
	if len(arrayField(tpl, "fields")) != 1 {
		t.Fatal("deleted text returned")
	}
}

func Test620SplitSignatureVariablesKeepFormattingAndOtherVariables(t *testing.T) {
	f := map[string]any{"text": "Hola {firma_departamental} {nombre}", "runs618": []any{map[string]any{"text": "Hola {firma_", "bold": true}, map[string]any{"text": "departamental} {nombre}", "italic": true}}}
	musicSignatureMarkers620(f, map[string]string{"firma_departamental": "Ana Pérez"})
	if f["text"] != "Hola Ana Pérez {nombre}" {
		t.Fatal(f["text"])
	}
	var text strings.Builder
	found := false
	for _, raw := range arrayField(f, "runs618") {
		r := legacyMap(raw)
		text.WriteString(stringValue(r["text"]))
		if r["text"] == "Ana Pérez" {
			found = r["bold"] == true
		}
	}
	if !found || text.String() != f["text"] {
		t.Fatal("style lost or runs mismatched")
	}
}

func Test620TextChangesAndRemovalUpdateEveryLinkedCertificate(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	q := cloneMap610(legacyUser(db, "p"))
	q["id"] = "q"
	q["fullName"] = "Otra Persona"
	db["users"] = append(arrayField(db, "users"), q)
	m := legacyMap(arrayField(db, "modules")[0])
	m["assignedUserIds"] = append(arrayField(m, "assignedUserIds"), "q")
	d := musicData550(db)
	tpl := certificateTemplate617()
	tpl["id"] = "sample620"
	tpl["textOnly620"] = true
	delete(tpl, "signatureDepartment")
	delete(tpl, "signatureCoordinator")
	legacyMap(arrayField(tpl, "fields")[0])["text"] = "ORIGINAL 620"
	ensureMapField(d, "templates")["sample620"] = tpl
	ensureMapField(d, "events")["ev620"] = map[string]any{"date": "2026-10-21", "name": "Prueba", "templateId610": "sample620", "participants": map[string]any{"p": map[string]any{"attended": true}, "q": map[string]any{"attended": true}}}
	st, _ := s.loadLegacy()
	st.State.Database = db
	musicAutoCertificates610(&st)
	identities := map[string]any{}
	hashes := map[string]any{}
	for _, owner := range []string{"p", "q"} {
		c := legacyMap(mapField(d, "certificates")[owner+":event:ev620"])
		identities[owner] = c["token"]
		hashes[owner] = c["sha256610"]
	}
	legacyMap(arrayField(tpl, "fields")[0])["text"] = "DUPLICADO 620"
	if !musicRefreshLinked618(&st, db) {
		t.Fatal("linked PDFs not updated")
	}
	for _, owner := range []string{"p", "q"} {
		c := legacyMap(mapField(d, "certificates")[owner+":event:ev620"])
		if c["token"] != identities[owner] || c["sha256610"] == hashes[owner] {
			t.Fatal("identity changed or stale PDF", owner)
		}
	}
	tpl["fields"] = []any{}
	musicRefreshLinked618(&st, db)
	for _, owner := range []string{"p", "q"} {
		c := legacyMap(mapField(d, "certificates")[owner+":event:ev620"])
		pdf := musicCertificateBytes610(c, tpl)
		if bytes.Contains(pdf, []byte("DUPLICADO 620")) || bytes.Contains(pdf, []byte("Firma Departamental")) || bytes.Contains(pdf, []byte("SISTEMA DE")) {
			t.Fatal("deleted text or signatures returned")
		}
		if c["token"] != identities[owner] {
			t.Fatal("code changed")
		}
	}
}

func Test620SaveWithoutSignatureOptionsAndEmptyTextList(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	st, _ := s.loadLegacy()
	st.State.Database = db
	_ = s.saveLegacy(st)
	v := map[string]any{"name": "Solo textos", "textOnly620": true, "fields": []any{}, "paperSize617": "carta"}
	code, out := call610(t, s, "admin", "", "template.save", v)
	if code != 200 {
		t.Fatal(code, out)
	}
	var saved map[string]any
	for _, raw := range mapField(out, "templates") {
		if legacyMap(raw)["name"] == "Solo textos" {
			saved = legacyMap(raw)
		}
	}
	if saved == nil || saved["textOnly620"] != true || len(arrayField(saved, "fields")) != 0 || saved["signatures613"] != nil {
		t.Fatal("special signatures recreated")
	}
}
