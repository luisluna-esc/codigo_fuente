package main

// Personal reminders belong to the person responsible for this applicant,
// independently of the wider scope of management reports.
func musicAssignedReview611(db, actor, owner map[string]any) bool {
	if !musicReview560(db, actor, owner) {
		return false
	}
	id := stringValue(actor["id"])
	tutor := stringValue(owner["tutorId"])
	if _, explicit := owner["tutorByProgram"]; explicit {
		tutor = stringValue(mapField(owner, "tutorByProgram")[musicProgram553])
	}
	return tutor == id || tutor == "" && owner["createdBy"] == id
}

func musicHistoryAllowed611(actor map[string]any) bool {
	role := legacyRole(actor)
	return role != "aspirant" && role != "pastor_aspirant"
}
