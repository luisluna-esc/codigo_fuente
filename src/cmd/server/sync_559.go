package main

// Resolve legacy implicit permissions against the current canonical database.
// Music edits then change only the Music member of an explicit program list.
func implicitProgramList559(db, user map[string]any, key string) []any {
	if key == "tutorProgramIds" {
		return tutorPrograms548(db, user)
	}
	if key != "managementProgramIds" {
		return nil
	}
	result := []any{}
	for _, raw := range arrayField(db, "modules") {
		module := legacyMap(raw)
		id := stringValue(module["id"])
		if module["active"] != false && managementProgram555(db, user, id) {
			result = append(result, id)
		}
	}
	return result
}

func prepareMusicAdminAssignments559(patch, db map[string]any) {
	for _, raw := range arrayField(patch, "users") {
		u := legacyMap(raw)
		old := legacyUser(db, stringValue(u["id"]))
		if old == nil {
			continue
		}
		for _, key := range []string{"managementProgramIds", "tutorProgramIds"} {
			value, present := u[key]
			if !present || old[key] != nil {
				continue
			}
			base := implicitProgramList559(db, old, key)
			u[key] = map[string]any{"__sglReplace": mergeLegacyPatchValue(base, value)}
		}
	}
}
