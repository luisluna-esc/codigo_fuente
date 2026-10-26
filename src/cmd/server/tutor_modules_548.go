package main

// A missing field retains compatibility with existing tutor accounts; an explicit
// empty field is authoritative and cannot be widened by a client proposal.
func tutorPrograms548(db map[string]any, actor map[string]any) []any {
	if stringValue(actor["role"]) == "lider_musica" {
		if actor["musicManagementRemoved597"] == true {
			return []any{}
		}
		return []any{musicProgram553}
	}
	if ids, exists := actor["tutorProgramIds"]; exists {
		out := []any{}
		for _, id := range anySlice548(ids) {
			for _, raw := range arrayField(db, "modules") {
				m := legacyMap(raw)
				if stringValue(m["id"]) == stringValue(id) && m["active"] != false {
					out = append(out, id)
				}
			}
		}
		return out
	}
	out := []any{}
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if m["active"] == false {
			continue
		}
		id := stringValue(m["id"])
		for _, ownerID := range arrayField(m, "assignedUserIds") {
			if legacyTutorAssigned(legacyUser(db, stringValue(ownerID)), stringValue(actor["id"]), id) {
				out = append(out, id)
				break
			}
		}
	}
	if len(out) == 0 {
		out = append(out, "class_grouped")
	}
	return out
}
func anySlice548(value any) []any {
	if a, ok := value.([]any); ok {
		return a
	}
	if a, ok := value.([]string); ok {
		out := []any{}
		for _, id := range a {
			out = append(out, id)
		}
		return out
	}
	return nil
}
func tutorCreationPrograms548(db, source map[string]any, actor map[string]any, id string) []any {
	allowed := tutorPrograms548(db, actor)
	chosen := []any{}
	for _, raw := range arrayField(source, "modules") {
		m := legacyMap(raw)
		p := stringValue(m["id"])
		if containsString(arrayField(m, "assignedUserIds"), id) {
			if !containsString(allowed, p) {
				return nil
			}
			for _, existing := range arrayField(db, "modules") {
				server := legacyMap(existing)
				if stringValue(server["id"]) == p && server["active"] != false {
					chosen = append(chosen, p)
					break
				}
			}
		}
	}
	// Old clients without any module roster may create only their original class.
	if len(arrayField(source, "modules")) == 0 && containsString(allowed, "class_grouped") {
		chosen = append(chosen, "class_grouped")
	}
	return chosen
}
