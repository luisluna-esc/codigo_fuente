package main

func canResetMusic571(db map[string]any, actorID, role, ownerID string) bool {
	actor, owner := legacyUser(db, actorID), legacyUser(db, ownerID)
	if !legacyActive(actor) || !legacyActive(owner) || !musicAssigned550(db, ownerID) || !managementProgram555(db, actor, musicProgram553) {
		return false
	}
	if musicManager550(role) {
		return true
	}
	return (role == "regional" || role == "tutor") && actorID != ownerID && legacyTutorAssigned(owner, actorID, musicProgram553)
}
