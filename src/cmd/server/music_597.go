package main

import "strings"

func musicAssignOwner597(db, actor, owner map[string]any) bool {
	if actor == nil || owner == nil {
		return false
	}
	if musicManager550(legacyRole(actor)) {
		return true
	}
	if legacyRole(actor) == "regional" {
		return musicScope550(db, actor, owner)
	}
	return stringValue(owner["createdBy"]) == stringValue(actor["id"]) || musicScope550(db, actor, owner)
}
func musicAssignTarget597(actor, tutor map[string]any) bool {
	if tutor == nil {
		return false
	}
	if musicManager550(legacyRole(actor)) {
		return true
	}
	switch legacyRole(tutor) {
	case "administrativo", "regional", "tutor":
		return true
	}
	return false
}
func musicTutorNotifications597(st *legacyStore, actor, owner map[string]any, oldID, newID, now string) {
	if oldID == newID {
		return
	}
	seen := map[string]bool{}
	for _, id := range []string{oldID, newID, stringValue(owner["id"])} {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		addLegacyNotification(st, legacyNotification{UserID: id, Type: "tutor_changed", Title: "Cambio de tutor de Música", Message: "Se actualizó la tutoría de " + stringValue(owner["fullName"]) + ". Consulta tus aspirantes y tu perfil.", ActorID: stringValue(actor["id"]), ActorName: stringValue(actor["fullName"]), ProgramID: musicProgram553, SourceOperationID: "music_tutor_" + stringValue(owner["id"]) + "_" + now, CreatedAt: now})
	}
}

func musicEmptyReason597(db, u map[string]any, files []any) string {
	id := stringValue(u["id"])
	for _, k := range []string{"fullName", "username", "email", "role", "phone", "documentNumber"} {
		if strings.TrimSpace(stringValue(u[k])) != "" {
			return "El registro tiene identidad. No se elimina como cuenta vacía."
		}
	}
	if len(mapField(u, "password")) > 0 {
		return "La cuenta tiene credenciales."
	}
	for _, raw := range files {
		if stringValue(legacyMap(raw)["ownerId"]) == id {
			return "El registro tiene archivos: requiere revisión manual."
		}
	}
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if stringValue(m["id"]) != musicProgram553 && containsString(arrayField(m, "assignedUserIds"), id) {
			return "El registro está vinculado a otro módulo."
		}
	}
	if len(legacyMap(legacyMap(db["bibleYear"])[id])) > 0 {
		return "El registro tiene lecturas bíblicas."
	}
	containers := []any{legacyMap(db["progress"])[id]}
	for _, p := range legacyMap(legacyMap(db["programProgress"])[id]) {
		containers = append(containers, p)
	}
	for _, raw := range containers {
		p := legacyMap(raw)
		if len(mapField(p, "readings")) > 0 {
			return "El registro tiene lecturas: requiere revisión manual."
		}
		for _, status := range mapField(p, "status") {
			if x := stringValue(status); x != "" && x != "pending" {
				return "El registro tiene avances."
			}
		}
		for _, sub := range mapField(p, "submissions") {
			if musicHasProgress592(legacyMap(sub)) {
				return "El registro tiene evidencias."
			}
		}
	}
	for _, raw := range legacyUsers(db) {
		o := legacyMap(raw)
		if legacyTutorAssigned(o, id, musicProgram553) {
			return "Hay aspirantes asignados al registro."
		}
	}
	for _, key := range []string{"tasks", "materials"} {
		for _, raw := range arrayField(db, key) {
			item := legacyMap(raw)
			for _, k := range []string{"ownerId", "assigneeId", "createdBy"} {
				if stringValue(item[k]) == id {
					return "El registro tiene tareas o materiales asociados."
				}
			}
		}
	}
	return ""
}
