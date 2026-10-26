package main

// A present map is authoritative, including an explicitly empty assignment.
// Accounts created before 5.45 keep their former tutor until an admin edits them.
func legacyTutorForProgram(owner map[string]any, programID string) string {
	if assignments, ok := owner["tutorByProgram"]; ok {
		if programID == "" {
			programID = "class_grouped"
		}
		return stringValue(legacyMap(assignments)[programID])
	}
	return stringValue(owner["tutorId"])
}
func legacyTutorAssigned(owner map[string]any, tutorID, programID string) bool {
	if tutorID == "" || owner == nil {
		return false
	}
	if programID != "" {
		return legacyTutorForProgram(owner, programID) == tutorID
	}
	if assignments, ok := owner["tutorByProgram"]; ok {
		for _, id := range legacyMap(assignments) {
			if stringValue(id) == tutorID {
				return true
			}
		}
		return false
	}
	return stringValue(owner["tutorId"]) == tutorID
}

// Notify the owner of the changed assignment, never an entire module or tutor roster.
func addTutorAssignmentNotifications(st *legacyStore, before any, actor map[string]any, operationID, now string) {
	for _, raw := range legacyUsers(st.State.Database) {
		owner := legacyMap(raw)
		if legacyRole(owner) != "aspirant" || !legacyActive(owner) {
			continue
		}
		ownerID := stringValue(owner["id"])
		old := legacyUser(before, ownerID)
		for _, programID := range []string{"class_grouped", "guide_major", "guide_master", "guide_master_advanced", "music_ministry"} {
			tutorID := legacyTutorForProgram(owner, programID)
			if tutorID == "" || tutorID == legacyTutorForProgram(old, programID) {
				continue
			}
			tutor := legacyUser(st.State.Database, tutorID)
			if !legacyActive(tutor) || legacyRole(tutor) != "tutor" && programID != musicProgram553 {
				continue
			}
			addLegacyNotification(st, legacyNotification{UserID: ownerID, OwnerID: ownerID, Type: "tutor_assigned", Title: "Tutor asignado", Message: legacyName(tutor) + " es tu tutor en " + programID + ".", ProgramID: programID, ActorID: stringValue(actor["id"]), ActorName: legacyName(actor), SourceOperationID: operationID + ":tutor:" + ownerID + ":" + programID, CreatedAt: now})
		}
	}
}

func remoteFileProgram(item map[string]any) string {
	if p := stringValue(item["programId"]); p != "" {
		return p
	}
	return "class_grouped"
}
