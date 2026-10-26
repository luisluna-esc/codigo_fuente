package main

import "strings"

// Expand only the retired signature markers. Other automatic variables remain
// intact, and a marker split across rich runs inherits its first character's style.
func musicSignatureMarkers620(f map[string]any, labels map[string]string) {
	text := stringValue(f["text"])
	matches := musicVariables618.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return
	}
	runs := arrayField(f, "runs618")
	result := []any{}
	var complete strings.Builder
	appendPart := func(start, end int, replacement *string) {
		if start == end {
			return
		}
		if len(runs) == 0 {
			if replacement != nil {
				complete.WriteString(*replacement)
			} else {
				complete.WriteString(text[start:end])
			}
			return
		}
		at := 0
		for _, raw := range runs {
			r := legacyMap(raw)
			next := at + len(stringValue(r["text"]))
			a, b := max(start, at), min(end, next)
			if a < b {
				copy := cloneMap610(r)
				part := stringValue(r["text"])[a-at : b-at]
				if replacement != nil {
					part = *replacement
				}
				copy["text"] = part
				result = append(result, copy)
				complete.WriteString(part)
				if replacement != nil {
					return
				}
			}
			at = next
		}
	}
	start, changed := 0, false
	for _, loc := range matches {
		value, ok := labels[strings.ToLower(text[loc[0]+1:loc[1]-1])]
		if !ok {
			continue
		}
		appendPart(start, loc[0], nil)
		appendPart(loc[0], loc[1], &value)
		start = loc[1]
		changed = true
	}
	if !changed {
		return
	}
	appendPart(start, len(text), nil)
	f["text"] = complete.String()
	if len(runs) > 0 {
		f["runs618"] = result
	}
}

// Special signatures become ordinary removable/duplicable text exactly once.
// A cleared text list stays cleared; deleting a converted name never recreates it.
func musicTextOnly620(t map[string]any) bool {
	if t == nil || t["textOnly620"] == true {
		return false
	}
	fields := append([]any{}, arrayField(t, "fields")...)
	if len(fields) == 0 {
		t["legacyBody620"] = true
	}
	labels := map[string]string{"firma_departamental": stringValue(t["signatureDepartment"]), "firma_coordinador": stringValue(t["signatureCoordinator"])}
	if labels["firma_departamental"] == "" {
		labels["firma_departamental"] = "Firma Departamental"
	}
	if labels["firma_coordinador"] == "" {
		labels["firma_coordinador"] = "Firma Coord. General Música"
	}
	for _, raw := range fields {
		musicSignatureMarkers620(legacyMap(raw), labels)
	}
	signatures := musicSignatures613(t)
	for _, key := range []string{"department", "coordinator"} {
		f := cloneMap610(signatures[key])
		label := labels["firma_departamental"]
		if key == "coordinator" {
			label = labels["firma_coordinador"]
		}
		f["text"] = label
		f["fitLines620"] = true
		fields = append(fields, f)
	}
	t["fields"] = fields
	t["textOnly620"] = true
	delete(t, "signatureDepartment")
	delete(t, "signatureCoordinator")
	delete(t, "signatures613")
	return true
}
