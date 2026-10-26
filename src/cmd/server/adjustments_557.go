package main

func reavivadosResetRole557(role string) bool {
	return role == "departamental" || role == "administrativo" || role == "secretaria" || role == "regional"
}

// Institutional roles receive this single operation, never a general reset permission.
func canResetReavivados557(db map[string]any, actorID, role string, req legacyProgressResetRequest, applied bool) bool {
	actor, owner := legacyUser(db, actorID), legacyUser(db, req.OwnerID)
	if !reavivadosResetRole557(role) || legacyRole(actor) != role || actorID == req.OwnerID || !managementProgram555(db, actor, musicProgram553) || !canReadOwner555(db, actor, owner, musicProgram553) {
		return false
	}
	if req.ProgramID != musicProgram553 || req.Scope != "requirement" || req.RequirementID != "LM-NUT-1" || len(req.RequirementIDs) != 0 {
		return false
	}
	return applied || musicAnnualChoice551(db, req.OwnerID) == "reavivados"
}

func resetMusicDerived557(db map[string]any, req legacyProgressResetRequest, actorID, now string) {
	if req.ProgramID != musicProgram553 {
		return
	}
	selected := legacyRequirementSelection(req)
	data := musicData550(db)
	if selected["LM-NUT-1"] {
		checks := legacyMap(mapField(data, "completions")[req.OwnerID])
		delete(checks, "reavivados")
		delete(checks, "bible_plan")
		// Legacy annual records must not repopulate a deleted diary.
		delete(legacyMap(mapField(db, "bibleYear")[req.OwnerID]), musicProgram553)
	}
	if len(selected) > 0 {
		ensureMapField(data, "workflow")[req.OwnerID] = map[string]any{"state": "correction", "requirementId": req.RequirementID, "changedBy": actorID, "at": now}
		data["updatedAt"], data["updatedBy"] = now, actorID
	}
}

func eventCertificate557(kind string) bool {
	return kind == "leadership" || kind == "beliefs" || kind == "egw_reading"
}
