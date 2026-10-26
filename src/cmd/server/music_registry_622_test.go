package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRegistryAge622(t *testing.T) {
	for _, v := range []struct {
		start, end string
		y, m       int
	}{{"2020-10-23", "2026-10-22", 5, 11}, {"2020-10-22", "2026-10-22", 6, 0}, {"2026-02-01", "2026-10-22", 0, 8}, {"2026-10-22", "2026-10-22", 0, 0}, {"2024-02-29", "2025-02-28", 0, 11}} {
		y, m := musicAge622(v.start, v.end)
		if y != v.y || m != v.m {
			t.Fatalf("%+v got %d %d", v, y, m)
		}
	}
}
func TestRegistry622MigrationAndOtherTemplates(t *testing.T) {
	_, db := fixture550(t)
	musicEnsure610(db)
	d := musicData550(db)
	before := cloneAny(mapField(d, "templates")["event"])
	records := ensureMapField(d, "ministries610")
	records["a"] = map[string]any{"createdAt": "2020", "name": "A"}
	records["b"] = map[string]any{"createdAt": "2021", "name": "B", "deletedAt": "2022"}
	musicRegistryDefaults622(db)
	if legacyMap(records["a"])["registrationCode622"] != "MBOS-MMUS-0001" || legacyMap(records["b"])["registrationCode622"] != "MBOS-MMUS-0002" {
		t.Fatal(records)
	}
	if musicNextCode622(d) != "MBOS-MMUS-0003" {
		t.Fatal("sequence reused")
	}
	if musicRegistryDefaults622(db) {
		t.Fatal("migration not idempotent")
	}
	if !reflect.DeepEqual(before, mapField(d, "templates")["event"]) {
		t.Fatal("changed another certificate")
	}
}
func TestRegistry622ModalitiesAndImmutableCode(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	st, _ := s.loadLegacy()
	st.State.Database = db
	s.saveLegacy(st)
	v := ministryFixture610()
	v["type"] = "Ensamble de Cuerdas"
	v["addModality622"] = true
	code, p := call610(t, s, "admin", "", "ministry.save", v)
	if code != 200 {
		t.Fatal(code, p)
	}
	m := legacyMap(arrayField(p, "ministries")[0])
	if m["registrationCode622"] != "MBOS-MMUS-0001" {
		t.Fatal(m)
	}
	v["id"] = m["id"]
	v["expectedUpdatedAt"] = m["updatedAt"]
	v["registrationCode622"] = "HACK"
	v["type"] = " ensamble de cuerdas "
	code, p = call610(t, s, "admin", "", "ministry.save", v)
	if code != 200 {
		t.Fatal(code, p)
	}
	m = legacyMap(arrayField(p, "ministries")[0])
	if m["registrationCode622"] != "MBOS-MMUS-0001" || m["type"] != "Ensamble de Cuerdas" {
		t.Fatal(m)
	}
}
func TestRegistry622VariablesAndAutomaticBinding(t *testing.T) {
	c := map[string]any{"kind": "ministry", "issuedAt": "2026-10-22T12:00:00Z", "registration": map[string]any{"name": "Marisol Sánchez", "type": "Solista", "createdDate": "2020-10-22", "registrationCode622": "MBOS-MMUS-0001"}}
	v := musicRegistryVars622(c)
	if v["nombre"] != "Marisol Sánchez" || v["trayectoria"] != "6 años" {
		t.Fatal(v)
	}
	if text := musicReplace618("{registro} · {modalidad} · {iglesia} · {fecha_inicio} · {trayectoria}", v); text != "MBOS-MMUS-0001 · Solista ·  · 22 de octubre de 2020 · 6 años" {
		t.Fatal(text)
	}
	fields := musicRegistryFields622()
	b, _ := json.Marshal(fields)
	var decoded []any
	json.Unmarshal(b, &decoded)
	if _, ok := musicTemplateFields592(decoded); !ok {
		t.Fatal("invalid registry fields")
	}
	legacyMap(decoded[2])["text"] = "Código cambiado"
	if _, ok := musicTemplateFields592(decoded); ok {
		t.Fatal("automatic binding writable")
	}
}

func TestRegistry622ReissueKeepsLinkedEvidence(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	st, _ := s.loadLegacy()
	st.State.Database = db
	s.saveLegacy(st)
	code, p := call610(t, s, "admin", "", "ministry.save", ministryFixture610())
	if code != 200 {
		t.Fatal(code, p)
	}
	id := registryID610(t, p)
	code, p = call610(t, s, "admin", "p", "ministry.link", map[string]any{"id": id})
	if code != 200 {
		t.Fatal(code, p)
	}
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	before := cloneAny(reviewProgressContainer(db, "p", musicProgram553))
	old := stringValue(legacyMap(mapField(musicData550(db), "certificates")["ministry-registration:"+id])["token"])
	code, p = call610(t, s, "admin", "", "ministry.reissue622", map[string]any{"id": id})
	if code != 200 {
		t.Fatal(code, p)
	}
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	if !reflect.DeepEqual(before, reviewProgressContainer(db, "p", musicProgram553)) {
		t.Fatal("renewal changed linked requirement")
	}
	c := legacyMap(mapField(musicData550(db), "certificates")["ministry-registration:"+id])
	if c["token"] == old {
		t.Fatal("did not renew")
	}
	if legacyMap(c["registration"])["registrationCode622"] != "MBOS-MMUS-0001" {
		t.Fatal("changed registration code")
	}
}

func TestRegistry623TrashAndRestoreNumber(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	st, _ := s.loadLegacy()
	st.State.Database = db
	s.saveLegacy(st)
	code, p := call610(t, s, "admin", "", "ministry.save", ministryFixture610())
	if code != 200 {
		t.Fatal(code, p)
	}
	id := registryID610(t, p)
	if c, p := call610(t, s, "admin", "p", "ministry.link", map[string]any{"id": id}); c != 200 {
		t.Fatal(c, p)
	}
	if c, p := call610(t, s, "admin", "", "ministry.delete", map[string]any{"id": id, "confirmation": "ELIMINAR MINISTERIO"}); c != 200 {
		t.Fatal(c, p)
	}
	if c, _ := call610(t, s, "p", "", "ministry.restore623", map[string]any{"id": id}); c != 403 {
		t.Fatal("unauthorized restore", c)
	}
	code, p = call610(t, s, "admin", "", "ministry.restore623", map[string]any{"id": id})
	if code != 200 {
		t.Fatal(code, p)
	}
	m := legacyMap(arrayField(p, "ministries")[0])
	if m["deletedAt"] != nil || m["registrationCode622"] != "MBOS-MMUS-0001" {
		t.Fatal(m)
	}
	if len(mapField(p, "links")) != 0 {
		t.Fatal("restored links unexpectedly")
	}
	v := ministryFixture610()
	v["name"] = "Segundo ministerio"
	code, p = call610(t, s, "admin", "", "ministry.save", v)
	if code != 200 {
		t.Fatal(code, p)
	}
	for _, raw := range arrayField(p, "ministries") {
		m := legacyMap(raw)
		if m["id"] != id && m["registrationCode622"] != "MBOS-MMUS-0002" {
			t.Fatal("reused number", m)
		}
	}
}
func TestRegistry623PrefixMigrationAndLayout(t *testing.T) {
	_, db := fixture550(t)
	musicEnsure610(db)
	d := musicData550(db)
	records := ensureMapField(d, "ministries610")
	records["old"] = map[string]any{"registrationCode622": "MBOS-MM-0017", "name": "Dúo Emanuel", "type": "Otro", "otherType": "Ensamble de Cuerdas"}
	musicRegistryDefaults622(db)
	if legacyMap(records["old"])["registrationCode622"] != "MBOS-MMUS-0017" {
		t.Fatal(records)
	}
	if musicNextCode622(d) != "MBOS-MMUS-0018" {
		t.Fatal("lost high water mark")
	}
	for _, raw := range musicRegistryFields622() {
		f := legacyMap(raw)
		if f["text"] == "{modalidad}" || f["text"] == "Modalidad" {
			t.Fatal("repeated modality")
		}
	}
	c := map[string]any{"kind": "ministry", "issuedAt": "2026-10-22", "registration": records["old"]}
	if musicReplace618("{registro_rotulo}", musicRegistryVars622(c)) != "REGISTRO N°: MBOS-MMUS-0017" {
		t.Fatal("missing registration label")
	}
}
