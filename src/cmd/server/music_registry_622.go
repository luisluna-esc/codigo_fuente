package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

func musicModalities622(d map[string]any) []any {
	out := []any{"Coro", "Solista", "Conjunto", "Dúo", "Trío", "Cuarteto", "Quinteto", "Sexteto", "Conjunto Instrumental", "Conjunto Vocal", "Ensamble Vocal", "Ensamble Instrumental", "Orquesta"}
	for _, raw := range arrayField(d, "modalities622") {
		v := strings.TrimSpace(stringValue(raw))
		found := false
		for _, x := range out {
			if strings.EqualFold(stringValue(x), v) {
				found = true
				break
			}
		}
		if v != "" && v != "Otro" && !found {
			out = append(out, v)
		}
	}
	return out
}

// A durable high-water mark includes deleted records and archived years.
func musicNextCode622(d map[string]any) string {
	n := int(numericValue(d["registrationSequence622"]))
	scan := func(records map[string]any) {
		for _, raw := range records {
			m := legacyMap(raw)
			s := strings.TrimPrefix(strings.TrimPrefix(stringValue(m["registrationCode622"]), "MBOS-MMUS-"), "MBOS-MM-")
			if v, err := strconv.Atoi(s); err == nil && v > n {
				n = v
			}
		}
	}
	scan(mapField(d, "ministries610"))
	for _, raw := range mapField(d, "archiveData610") {
		scan(mapField(musicData550(mapField(legacyMap(raw), "database")), "ministries610"))
	}
	n++
	d["registrationSequence622"] = n
	return fmt.Sprintf("MBOS-MMUS-%04d", n)
}

func musicRegistryDefaults622(db map[string]any) bool {
	d := musicData550(db)
	changed := false
	records := mapField(d, "ministries610")
	ids := []string{}
	for id := range records {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := legacyMap(records[ids[i]]), legacyMap(records[ids[j]])
		x, y := stringValue(a["createdAt"]), stringValue(b["createdAt"])
		if x == y {
			return ids[i] < ids[j]
		}
		return x < y
	})
	for _, id := range ids {
		m := legacyMap(records[id])
		if m["type"] == "Otro" && strings.TrimSpace(stringValue(m["otherType"])) != "" {
			value := strings.Join(strings.Fields(stringValue(m["otherType"])), " ")
			if !containsString(musicModalities622(d), value) {
				d["modalities622"] = append(musicModalities622(d), value)
			}
			m["type"] = value
			changed = true
		}
		if code := stringValue(m["registrationCode622"]); strings.HasPrefix(code, "MBOS-MM-") {
			m["registrationCode622"] = strings.Replace(code, "MBOS-MM-", "MBOS-MMUS-", 1)
			changed = true
		}
		if m["registrationCode622"] == nil {
			m["registrationCode622"] = musicNextCode622(d)
			changed = true
		}
	}
	t := legacyMap(mapField(d, "templates")["ministry"])
	// Install the new body only for the untouched legacy body. Customized layouts survive.
	if t != nil && t["registryLayout622"] != true && (t["legacyBody620"] == true || len(arrayField(t, "fields")) == 0) {
		signatures := arrayField(t, "fields")
		t["fields"] = append(musicRegistryFields622(), signatures...)
		t["textOnly620"] = true
		delete(t, "legacyBody620")
		t["registryLayout622"] = true
		changed = true
	}
	for _, raw := range mapField(d, "templates") {
		t := legacyMap(raw)
		if t["certificateType"] != "ministry" || t["registryLayout623"] == true {
			continue
		}
		fields := []any{}
		for _, raw := range arrayField(t, "fields") {
			f := legacyMap(raw)
			text := strings.TrimSpace(stringValue(f["text"]))
			if text == "{modalidad}" || text == "Modalidad" {
				continue
			}
			if text == "{registro}" {
				f["text"] = "{registro_rotulo}"
				f["automatic622"] = "registro_rotulo"
				delete(f, "runs618")
			}
			fields = append(fields, f)
		}
		t["fields"] = fields
		t["registryLayout623"] = true
		changed = true
	}
	return changed
}

func musicAge622(start, end string) (int, int) {
	if len(end) > 10 {
		end = end[:10]
	}
	a, e := time.Parse("2006-01-02", start)
	b, f := time.Parse("2006-01-02", end)
	if e != nil || f != nil || b.Before(a) {
		return 0, 0
	}
	months := (b.Year()-a.Year())*12 + int(b.Month()-a.Month())
	if b.Day() < a.Day() {
		months--
	}
	if months < 0 {
		months = 0
	}
	return months / 12, months % 12
}
func musicRegistryVars622(c map[string]any) map[string]string {
	if c["kind"] != "ministry" {
		return nil
	}
	r := legacyMap(c["registration"])
	typ := stringValue(r["type"])
	if typ == "Otro" {
		typ = stringValue(r["otherType"])
	}
	y, m := musicAge622(stringValue(r["createdDate"]), stringValue(c["issuedAt"]))
	parts := []string{}
	if y > 0 {
		s := "años"
		if y == 1 {
			s = "año"
		}
		parts = append(parts, fmt.Sprintf("%d %s", y, s))
	}
	if m > 0 {
		s := "meses"
		if m == 1 {
			s = "mes"
		}
		parts = append(parts, fmt.Sprintf("%d %s", m, s))
	}
	if len(parts) == 0 {
		parts = append(parts, "Menos de un mes")
	}
	return map[string]string{"nombre": stringValue(r["name"]), "ministerio": stringValue(r["name"]), "registro": stringValue(r["registrationCode622"]), "registro_rotulo": "REGISTRO N°: " + stringValue(r["registrationCode622"]), "modalidad": typ, "iglesia": stringValue(r["church"]), "fecha_inicio": musicDate619(stringValue(r["createdDate"])), "trayectoria": strings.Join(parts, " y "), "voto_junta": stringValue(r["boardVote"])}
}
func musicAutomatic622(key string) bool {
	switch key {
	case "ministerio", "registro_rotulo", "registro", "modalidad", "iglesia", "fecha_inicio", "trayectoria", "voto_junta", "fecha":
		return true
	}
	return false
}
func musicRegistryFields622() []any {
	out := []any{}
	add := func(text string, y, size float64, automatic bool) {
		f := map[string]any{"text": text, "x": float64(10), "y": y, "width": float64(80), "size": size, "font": "DejaVu Serif", "color": "#173c76", "align": "center"}
		if automatic {
			f["automatic622"] = strings.Trim(text, "{}")
		}
		out = append(out, f)
	}
	add("CERTIFICADO DE REGISTRO", 12, 25, false)
	add("MINISTERIO DE MÚSICA", 19, 16, false)
	add("{registro_rotulo}", 25, 13, true)
	add("Se deja constancia del registro de", 32, 13, false)
	add("{ministerio}", 38, 29, true)
	add("en el Registro de Ministerios de Música.", 57, 13, false)
	// Independent labels and values keep the source data read-only.
	labels := []string{"Iglesia:", "Fecha de inicio del ministerio:", "N.º de voto de Junta:", "Trayectoria a la fecha de emisión:"}
	keys := []string{"iglesia", "fecha_inicio", "voto_junta", "trayectoria"}
	for i, k := range keys {
		y := float64(63 + i*4)
		add(labels[i], y, 11, false)
		label := legacyMap(out[len(out)-1])
		label["x"], label["width"], label["align"] = float64(12), float64(39), "left"
		add("{"+k+"}", y, 11, true)
		f := legacyMap(out[len(out)-1])
		f["x"], f["width"], f["align"] = float64(52), float64(36), "left"
	}
	add("Se expide el presente certificado como constancia de su registro.", 80, 11, false)
	add("{fecha}", 89, 11, true)
	return out
}
