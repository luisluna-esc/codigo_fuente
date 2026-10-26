package main

import (
	"net/http"
	"strings"
)

// legacyReviewItem returns the canonical state for one submitted requirement.
// SGL 5.18 uses this small delta after a review_submitted notification so the
// reviewer inbox does not have to wait for (or download) a full bootstrap.
func (s *Server) legacyReviewItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	actorID := strings.TrimSpace(q.Get("actorId"))
	ownerID := strings.TrimSpace(q.Get("ownerId"))
	programID := strings.TrimSpace(q.Get("programId"))
	requirementID := strings.TrimSpace(q.Get("requirementId"))
	if programID == "" {
		programID = "class_grouped"
	}
	if actorID == "" || ownerID == "" || requirementID == "" {
		legacyError(w, 400, "Faltan el revisor, Aspirante o requisito.", nil)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	if st.State.Database == nil {
		legacyError(w, 409, "El servidor todavía no tiene datos inicializados.", nil)
		return
	}

	reviewer, _, ok := s.reviewDecisionActor(r, actorID, &st)
	if !ok {
		legacyError(w, 403, "No autorizado para consultar este requisito.", nil)
		return
	}
	owner := legacyUser(st.State.Database, ownerID)
	if !leader555(legacyMap(st.State.Database), owner) {
		legacyError(w, 404, "El Aspirante no está disponible.", nil)
		return
	}
	if !canReview556(legacyMap(st.State.Database), reviewer, owner, programID) {
		legacyError(w, 403, "El Aspirante no está asignado a este Tutor.", nil)
		return
	}

	progress := legacyProgressForRequirement(st.State.Database, ownerID, programID, false)
	if progress == nil {
		legacyError(w, 404, "No se encontró el avance del requisito.", nil)
		return
	}
	status := strings.TrimSpace(stringValue(mapField(progress, "status")[requirementID]))
	submission := legacyMap(mapField(progress, "submissions")[requirementID])
	if status == "" || submission == nil {
		legacyError(w, 404, "No se encontró la entrega del requisito.", nil)
		return
	}

	fileIDs := map[string]bool{}
	for _, rawID := range arrayField(submission, "attachmentIds") {
		if id := strings.TrimSpace(stringValue(rawID)); id != "" {
			fileIDs[id] = true
		}
	}
	if id := strings.TrimSpace(stringValue(submission["primaryDocumentId"])); id != "" {
		fileIDs[id] = true
	}

	files := make([]any, 0, len(fileIDs))
	for _, raw := range st.State.Files {
		file := legacyMap(raw)
		id := strings.TrimSpace(stringValue(file["id"]))
		if id == "" || !fileIDs[id] {
			continue
		}
		if fid := strings.TrimSpace(stringValue(file["ownerId"])); fid != "" && fid != ownerID {
			continue
		}
		if rid := strings.TrimSpace(stringValue(file["requirementId"])); rid != "" && rid != requirementID {
			continue
		}
		if pid := strings.TrimSpace(stringValue(file["programId"])); pid != "" && pid != programID {
			continue
		}
		files = append(files, cloneAny(file))
	}

	updatedAt := strings.TrimSpace(stringValue(submission["updatedAt"]))
	if updatedAt == "" {
		updatedAt = strings.TrimSpace(stringValue(progress["updatedAt"]))
	}
	jsonOut(w, 200, map[string]any{
		"ok":            true,
		"ownerId":       ownerID,
		"programId":     programID,
		"requirementId": requirementID,
		"status":        status,
		"submission":    cloneAny(submission),
		"files":         files,
		"updatedAt":     updatedAt,
		"serverVersion": st.State.ServerVersion,
	})
}
