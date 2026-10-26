package main

import (
	"net/http"
	"strings"
	"time"
)

// legacyProgressResetRequest removes only a user's saved advancement data.
// Requirement definitions/configuration are never deleted by this endpoint.
type legacyProgressResetRequest struct {
	ActorID           string   `json:"actorId"`
	OwnerID           string   `json:"ownerId"`
	ProgramID         string   `json:"programId"`
	RequirementID     string   `json:"requirementId,omitempty"`
	RequirementIDs    []string `json:"requirementIds,omitempty"`
	Scope             string   `json:"scope,omitempty"` // requirement | selection
	OperationID       string   `json:"operationId"`
	BaseUpdatedAt     string   `json:"baseUpdatedAt,omitempty"`
	RestoreGeneration int64    `json:"restoreGeneration,omitempty"`
}

func legacyCanResetProgress(db any, actorID, actorRole, ownerID string) bool {
	actorID = strings.TrimSpace(actorID)
	ownerID = strings.TrimSpace(ownerID)
	actorRole = strings.TrimSpace(actorRole)
	if actorID == "" || ownerID == "" {
		return false
	}
	owner := legacyUser(db, ownerID)
	if !leader555(legacyMap(db), owner) {
		return false
	}
	switch actorRole {
	case "aspirant":
		return actorID == ownerID
	case "admin":
		return true
	case "tutor":
		return legacyTutorAssigned(owner, actorID, "")
	default:
		return false
	}
}

func legacyRequirementSelection(req legacyProgressResetRequest) map[string]bool {
	selected := map[string]bool{}
	if req.Scope == "requirement" {
		if id := strings.TrimSpace(req.RequirementID); id != "" {
			selected[id] = true
		}
		return selected
	}
	if req.Scope == "selection" {
		for _, raw := range req.RequirementIDs {
			if id := strings.TrimSpace(raw); id != "" {
				selected[id] = true
			}
		}
	}
	return selected
}

func legacyProgressAttachmentIDs(progress map[string]any, req legacyProgressResetRequest) map[string]bool {
	ids := map[string]bool{}
	submissions := mapField(progress, "submissions")
	if submissions == nil {
		return ids
	}
	collect := func(raw any) {
		submission := legacyMap(raw)
		if submission == nil {
			return
		}
		for _, rawID := range arrayField(submission, "attachmentIds") {
			id := strings.TrimSpace(stringValue(rawID))
			if id != "" {
				ids[id] = true
			}
		}
	}
	for requirementID := range legacyRequirementSelection(req) {
		collect(submissions[requirementID])
	}
	return ids
}

func legacyResetProgressFiles(files []any, attachmentIDs map[string]bool, ownerID, programID string, req legacyProgressResetRequest) ([]any, int) {
	kept := make([]any, 0, len(files))
	removed := 0
	selected := legacyRequirementSelection(req)
	for _, raw := range files {
		file := legacyMap(raw)
		if file == nil {
			kept = append(kept, raw)
			continue
		}
		id := strings.TrimSpace(stringValue(file["id"]))
		remove := id != "" && attachmentIDs[id]
		if !remove && strings.TrimSpace(stringValue(file["ownerId"])) == ownerID {
			fileProgram := strings.TrimSpace(stringValue(file["programId"]))
			if fileProgram == "" {
				fileProgram = "class_grouped"
			}
			expectedProgram := strings.TrimSpace(programID)
			if expectedProgram == "" {
				expectedProgram = "class_grouped"
			}
			fileRequirement := strings.TrimSpace(stringValue(file["requirementId"]))
			if fileProgram == expectedProgram && fileRequirement != "" {
				remove = selected[fileRequirement]
			}
		}
		if remove {
			removed++
			continue
		}
		kept = append(kept, raw)
	}
	return kept, removed
}

func legacyAppendProgressResetAudit(db map[string]any, actor map[string]any, owner map[string]any, req legacyProgressResetRequest, now string) {
	if db == nil {
		return
	}
	target := req.RequirementID
	detail := "Avance eliminado definitivamente; el requisito quedó Pendiente."
	if req.Scope == "selection" {
		target = req.ProgramID
		detail = "Los avances seleccionados fueron eliminados definitivamente y quedaron Pendientes."
	}
	entry := map[string]any{
		"id":        "audit_" + token(10),
		"at":        now,
		"actorId":   stringValue(actor["id"]),
		"actorName": legacyName(actor),
		"action":    "Avance eliminado",
		"entity":    target,
		"detail":    detail + " Aspirante: " + legacyName(owner),
		"programId": req.ProgramID, "ownerId": req.OwnerID, "requirementId": req.RequirementID,
	}
	audit := arrayField(db, "audit")
	audit = append([]any{entry}, audit...)
	if len(audit) > 2500 {
		audit = audit[:2500]
	}
	db["audit"] = audit
}

func (s *Server) legacyProgressReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	var req legacyProgressResetRequest
	if err := readJSONLimited(r, &req, 256<<10); err != nil {
		legacyError(w, 400, "Solicitud inválida.", nil)
		return
	}
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.OwnerID = strings.TrimSpace(req.OwnerID)
	req.ProgramID = strings.TrimSpace(req.ProgramID)
	req.RequirementID = strings.TrimSpace(req.RequirementID)
	seenRequirementIDs := map[string]bool{}
	normalizedRequirementIDs := make([]string, 0, len(req.RequirementIDs))
	for _, raw := range req.RequirementIDs {
		id := strings.TrimSpace(raw)
		if id == "" || seenRequirementIDs[id] {
			continue
		}
		seenRequirementIDs[id] = true
		normalizedRequirementIDs = append(normalizedRequirementIDs, id)
	}
	req.RequirementIDs = normalizedRequirementIDs
	req.Scope = strings.TrimSpace(strings.ToLower(req.Scope))
	req.OperationID = strings.TrimSpace(req.OperationID)
	if req.ProgramID == "" {
		req.ProgramID = "class_grouped"
	}
	if req.Scope == "" {
		req.Scope = "requirement"
	}
	if req.Scope != "requirement" && req.Scope != "selection" {
		legacyError(w, 400, "Alcance de eliminación inválido.", nil)
		return
	}
	if req.ActorID == "" || req.OwnerID == "" || req.OperationID == "" || (req.Scope == "requirement" && req.RequirementID == "") || (req.Scope == "selection" && len(req.RequirementIDs) == 0) {
		legacyError(w, 400, "Faltan datos para eliminar el avance.", nil)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	actorID, actorRole, ok := s.notificationActor(r, req.ActorID, &st)
	db := legacyMap(st.State.Database)
	if !ok && (reavivadosResetRole557(legacyRole(legacyUser(db, req.ActorID))) || req.ProgramID == musicProgram553) {
		actorID, actorRole, ok = s.accountActor556(r, req.ActorID, &st)
	}
	if personalRequest555(r) {
		actorRole = "aspirant"
	}
	previous := findAppliedLegacyOperation(&st, actorID, req.OperationID)
	extended := reavivadosResetRole557(actorRole)
	allowed := legacyCanResetProgress(db, actorID, actorRole, req.OwnerID)
	if extended {
		allowed = canResetReavivados557(db, actorID, actorRole, req, previous != nil)
	}
	if req.ProgramID == musicProgram553 && actorRole != "aspirant" {
		allowed = canResetMusic571(db, actorID, actorRole, req.OwnerID) || actorRole == "administrativo" && allowed
	}
	if !ok || actorID != req.ActorID || !allowed || (actorRole == "tutor" && !legacyTutorAssigned(legacyUser(db, req.OwnerID), actorID, req.ProgramID)) {
		legacyError(w, 403, "No autorizado para eliminar este avance.", nil)
		return
	}
	// El Aspirante puede eliminar únicamente su propio requisito individual.
	// Tutor y Administrador conservan la eliminación individual o por selección.
	if actorRole == "aspirant" && (actorID != req.OwnerID || req.Scope != "requirement") {
		legacyError(w, 403, "El Aspirante solo puede eliminar su propio avance de un requisito.", nil)
		return
	}
	if !extended && !(req.ProgramID == musicProgram553 && canResetMusic571(db, actorID, actorRole, req.OwnerID)) && actorRole != "aspirant" && actorRole != "admin" && actorRole != "tutor" {
		legacyError(w, 403, "No autorizado para eliminar este avance.", nil)
		return
	}
	if actorRole == "aspirant" && legacyRequirementStatus(st.State.Database, req.OwnerID, req.ProgramID, req.RequirementID) == "approved" {
		legacyError(w, 403, "El informe revisado solo puede ser eliminado por el Tutor asignado o el Administrador.", nil)
		return
	}
	// The destructive action is idempotent. Retrying the exact operation can
	// safely return the current canonical snapshot without deleting anything else.
	if previous != nil {
		if extended && (previous.OwnerID != req.OwnerID || previous.OperationType != "progress_reset" || previous.ProgramID != req.ProgramID || previous.RequirementID != req.RequirementID) {
			legacyError(w, 409, "El identificador ya corresponde a otra operación.", nil)
			return
		}
		payload := s.legacySyncSnapshotForRequest(st.State, r)
		payload["status"] = "applied"
		payload["operationId"] = req.OperationID
		payload["operationType"] = "progress_reset"
		payload["deduplicated"] = true
		payload["scope"] = req.Scope
		payload["ownerId"] = req.OwnerID
		payload["programId"] = req.ProgramID
		payload["requirementId"] = req.RequirementID
		payload["requirementIds"] = req.RequirementIDs
		jsonOut(w, 200, payload)
		return
	}

	progress := legacyProgressForRequirement(st.State.Database, req.OwnerID, req.ProgramID, false)
	if extended && (req.RestoreGeneration != legacyCanonicalRestoreGeneration(st.State) || req.BaseUpdatedAt != stringValue(progress["updatedAt"])) {
		legacyError(w, 409, "El avance cambió. Actualiza y vuelve a confirmar la eliminación.", nil)
		return
	}
	attachmentIDs := legacyProgressAttachmentIDs(progress, req)
	if progress != nil {
		status := ensureMapField(progress, "status")
		submissions := ensureMapField(progress, "submissions")
		for requirementID := range legacyRequirementSelection(req) {
			status[requirementID] = "pending"
			delete(submissions, requirementID)
			if readings := mapField(progress, "readings"); readings != nil {
				delete(readings, requirementID)
			}
		}
		progress["updatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
	}
	var removedFiles int
	st.State.Files, removedFiles = legacyResetProgressFiles(st.State.Files, attachmentIDs, req.OwnerID, req.ProgramID, req)

	now := time.Now().UTC().Format(time.RFC3339Nano)
	actor := legacyUser(st.State.Database, actorID)
	owner := legacyUser(st.State.Database, req.OwnerID)
	resetMusicDerived557(db, req, actorID, now)
	legacyAppendProgressResetAudit(db, actor, owner, req, now)
	if db != nil {
		db["updatedAt"] = now
		db["revision"] = numericValue(db["revision"]) + 1
	}
	st.State.ServerVersion++
	st.State.UpdatedAt = now
	st.State.Database = legacyPrepareCanonicalDB(db, st.State, "Avance eliminado")
	rememberAppliedLegacyOperation(&st, legacyAppliedOperation{
		OperationID: req.OperationID, ActorID: actorID, OwnerID: req.OwnerID, OperationType: "progress_reset",
		RequirementID: func() string {
			if req.Scope == "selection" {
				return strings.Join(req.RequirementIDs, ",")
			}
			return req.RequirementID
		}(), ProgramID: req.ProgramID, AppliedAt: now, ServerVersion: st.State.ServerVersion,
	})
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}

	payload := s.legacySyncSnapshotForRequest(st.State, r)
	payload["status"] = "applied"
	payload["operationId"] = req.OperationID
	payload["operationType"] = "progress_reset"
	payload["scope"] = req.Scope
	payload["ownerId"] = req.OwnerID
	payload["programId"] = req.ProgramID
	payload["requirementId"] = req.RequirementID
	payload["requirementIds"] = req.RequirementIDs
	payload["removedFiles"] = removedFiles
	payload["clearLocalOperation"] = true
	jsonOut(w, 200, payload)
}
