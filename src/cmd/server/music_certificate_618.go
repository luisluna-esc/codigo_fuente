package main

import (
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"mime"
	"regexp"
	"strings"
	"sync"
)

// Only template artwork and layout are shipped. No database, account or
// original backup is embedded in an installation.
//
//go:embed defaults618.json
var musicDefaultJSON618 []byte

var musicDefaultOnce618 sync.Once
var musicDefaultsMap618 map[string]any
var musicAssetsOnce618 sync.Once
var musicAssets618 map[string]map[string]any

func musicBundledAsset618(id string) map[string]any {
	musicAssetsOnce618.Do(func() {
		musicAssets618 = map[string]map[string]any{}
		entries, _ := musicFontFiles610.ReadDir("fonts610")
		for _, e := range entries {
			raw, _ := musicFontFiles610.ReadFile("fonts610/" + e.Name())
			family, bold, italic := musicFontName617(raw)
			if family == "" {
				continue
			}
			hash := sha256.Sum256(raw)
			key := "font613_" + hex.EncodeToString(hash[:])
			kind := "font/ttf"
			if string(raw[:4]) == "OTTO" {
				kind = "font/otf"
			}
			musicAssets618[key] = map[string]any{"id": key, "kind": "music_font613", "name": family, "bold618": bold, "italic618": italic, "type": kind, "size": len(raw), "programId": musicProgram553, "dataUrl": "data:" + kind + ";base64," + base64.StdEncoding.EncodeToString(raw)}
		}
	})
	return musicAssets618[id]
}

func musicDefaults618(db map[string]any) bool {
	templates := ensureMapField(musicData550(db), "templates")
	musicDefaultOnce618.Do(func() { _ = json.Unmarshal(musicDefaultJSON618, &musicDefaultsMap618) })
	changed := false
	for id, raw := range musicDefaultsMap618 {
		if templates[id] == nil {
			templates[id] = cloneAny(raw)
			changed = true
		}
	}
	return changed
}

func musicInstallAssets618(st *legacyStore) bool {
	changed := false
	for _, raw := range mapField(musicData550(legacyMap(st.State.Database)), "templates") {
		if musicBindFonts618(legacyMap(raw)) {
			changed = true
		}
	}
	ids := map[string]bool{}
	musicFileIDs553(mapField(musicData550(legacyMap(st.State.Database)), "templates"), ids)
	present := map[string]bool{}
	for _, raw := range st.State.Files {
		present[stringValue(legacyMap(raw)["id"])] = true
	}
	for id := range ids {
		if !present[id] {
			if asset := musicBundledAsset618(id); asset != nil {
				st.State.Files = append(st.State.Files, cloneAny(asset))
				changed = true
			}
		}
	}
	return changed
}

func musicBindFonts618(t map[string]any) bool {
	musicBundledAsset618("")
	fields := append([]any{}, arrayField(t, "fields")...)
	for _, raw := range mapField(t, "signatures613") {
		fields = append(fields, raw)
	}
	changed := false
	for at := 0; at < len(fields); at++ {
		f := legacyMap(fields[at])
		fields = append(fields, arrayField(f, "runs618")...)
		if stringValue(f["fontAsset613"]) != "" {
			continue
		}
		bestID := ""
		bestScore := 100
		for id, asset := range musicAssets618 {
			if !strings.EqualFold(stringValue(asset["name"]), stringValue(f["font"])) {
				continue
			}
			score := 0
			if (asset["bold618"] == true) != (f["bold"] == true) {
				score += 2
			}
			if (asset["italic618"] == true) != (f["italic"] == true) {
				score++
			}
			if score < bestScore || score == bestScore && id < bestID {
				bestScore = score
				bestID = id
			}
		}
		if bestID != "" {
			f["fontAsset613"] = bestID
			changed = true
		}
	}
	return changed
}

func musicNumber618(v map[string]any, key string, low, high, fallback float64) (float64, bool) {
	if v[key] == nil {
		return fallback, true
	}
	n, ok := v[key].(float64)
	return n, ok && !math.IsNaN(n) && !math.IsInf(n, 0) && n >= low && n <= high
}

func musicTextSettings618(f, out map[string]any) bool {
	if value, exists := f["fitLines620"]; exists {
		fit, ok := value.(bool)
		if !ok {
			return false
		}
		out["fitLines620"] = fit
	}
	for key, limits := range map[string][3]float64{"opacity618": {0, 1, 1}, "lineSpacing618": {0.7, 3, 1.25}} {
		n, valid := musicNumber618(f, key, limits[0], limits[1], limits[2])
		if !valid {
			return false
		}
		out[key] = n
	}
	if f["runs618"] == nil {
		return true
	}
	runs, ok := f["runs618"].([]any)
	if !ok || len(runs) > 256 {
		return false
	}
	clean := []any{}
	var text strings.Builder
	for _, raw := range runs {
		r := cloneMap610(f)
		delete(r, "runs618")
		for k, v := range legacyMap(raw) {
			if k != "runs618" {
				r[k] = v
			}
		}
		if legacyMap(raw) == nil {
			return false
		}
		valid, ok := musicTemplateFields592([]any{r})
		if !ok {
			return false
		}
		v := legacyMap(valid[0])
		part := map[string]any{}
		for _, k := range []string{"text", "size", "font", "fontAsset613", "bold", "italic", "underline", "color"} {
			if v[k] != nil {
				part[k] = v[k]
			}
		}
		text.WriteString(stringValue(v["text"]))
		clean = append(clean, part)
	}
	if text.String() != stringValue(f["text"]) {
		return false
	}
	out["runs618"] = clean
	return true
}

func musicTemplateSettings618(v, out map[string]any) error {
	opacity, ok := musicNumber618(v, "backgroundOpacity618", 0, 1, 1)
	if !ok {
		return fmt.Errorf("Revisa la opacidad del fondo.")
	}
	out["backgroundOpacity618"] = opacity
	if v["verification618"] != nil {
		f := cloneMap610(legacyMap(v["verification618"]))
		// The token is generated by the server. Neither its text nor visibility
		// can be supplied by the editor.
		x, okX := musicNumber618(f, "x", 0, 95, 35)
		y, okY := musicNumber618(f, "y", 0, 97, 96.5)
		if !okX || !okY {
			return fmt.Errorf("Coloca el código de verificación dentro de la hoja.")
		}
		out["verification618"] = map[string]any{"x": x, "y": y}
	}
	return nil
}

func (p *musicPDF610) opacity618(value any) string {
	n := 1.0
	if value != nil {
		n = numericValue(value)
	}
	if n >= 1 {
		return ""
	}
	if n < 0 {
		n = 0
	}
	key := fmt.Sprintf("O%03d", int(math.Round(n*1000)))
	if p.opacityResources618 == nil {
		p.opacityResources618 = map[string]int{}
	}
	if p.opacityResources618[key] == 0 {
		p.opacityResources618[key] = p.add(fmt.Sprintf("<< /Type /ExtGState /ca %.3f /CA %.3f >>", n, n))
	}
	return "/" + key + " gs\n"
}

var musicVariables618 = regexp.MustCompile(`(?i)\{(nombre|titulo|datos|fecha|lugar|ministerio|firma_departamental|firma_coordinador|investidura|voto_junta|registro_rotulo|registro|modalidad|iglesia|fecha_inicio|trayectoria)\}`)

func musicReplace618(text string, vars map[string]string) string {
	return musicVariables618.ReplaceAllStringFunc(text, func(marker string) string {
		key := strings.ToLower(marker[1 : len(marker)-1])
		if v, ok := vars[key]; ok {
			return v
		}
		switch key {
		case "registro_rotulo", "registro", "modalidad", "iglesia", "fecha_inicio", "trayectoria":
			return marker
		}
		return ""
	})
}

func musicApproval618(db map[string]any, owner string) string {
	w := legacyMap(mapField(musicData550(db), "workflow")[owner])
	if w["state"] != "approved" {
		return ""
	}
	date := stringValue(w["approvedAt"])
	if !musicFrom2026_618(date) {
		return ""
	}
	return date
}

func musicFilename618(c map[string]any, sample bool) string {
	name := stringValue(c["fullName"]) + " - " + stringValue(c["title"])
	if sample {
		name = "Muestra - " + stringValue(c["name"])
	}
	name = strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune(`/\:*?"<>|`, r) {
			return '-'
		}
		return r
	}, name)
	name = strings.Trim(strings.Join(strings.Fields(name), " "), " .-")
	if name == "" {
		name = "Certificado de Música"
	}
	return name + ".pdf"
}
func musicDisposition618(name string) string {
	return mime.FormatMediaType("inline", map[string]string{"filename": name})
}
