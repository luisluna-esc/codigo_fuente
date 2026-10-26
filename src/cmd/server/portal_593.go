package main

// Every login and existing general-portal session must have an active
// enrollment or management permission in that portal.
func portalAllowed593(db map[string]any, user map[string]any, music bool) bool {
	if user == nil || !legacyActive(user) {
		return false
	}
	if music {
		return musicAllowed553(db, user)
	}
	return accessInScope571(db, user, "conquistadores")
}
