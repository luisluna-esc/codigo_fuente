package main

import (
	"encoding/json"
	"net/http"
)

// A read-only, owner-scoped poll. It is independent of the upload queue and
// returns canonical workflow changes without downloading unrelated drafts.
func (s *Server) legacyOwnReviewChanges(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	var req struct {
		ActorID      string   `json:"actorId"`
		OperationIDs []string `json:"operationIds"`
		Items        []struct {
			ProgramID     string `json:"programId"`
			RequirementID string `json:"requirementId"`
			Status        string `json:"status"`
		} `json:"items"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 256<<10)).Decode(&req) != nil || len(req.Items) > 1000 || len(req.OperationIDs) > 1000 {
		legacyError(w, 400, "Consulta inválida", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	id, role, ok := s.notificationActor(r, req.ActorID, &st)
	if !ok || role != "aspirant" || !legacyActive(legacyUser(st.State.Database, id)) {
		legacyError(w, 403, "No autorizado", nil)
		return
	}
	acknowledged := []string{}
	for _, operationID := range req.OperationIDs {
		if findAppliedLegacyOperation(&st, id, operationID) != nil {
			acknowledged = append(acknowledged, operationID)
		}
	}
	changes := []any{}
	for _, item := range req.Items {
		if item.Status != "in_review" && item.Status != "approved" {
			continue
		}
		p := legacyProgressForRequirement(st.State.Database, id, item.ProgramID, false)
		if p == nil {
			continue
		}
		status := stringValue(mapField(p, "status")[item.RequirementID])
		if status == "" || status == item.Status {
			continue
		}
		changes = append(changes, map[string]any{"ownerId": id, "programId": item.ProgramID, "requirementId": item.RequirementID, "previousStatus": item.Status, "status": status, "submission": cloneAny(mapField(p, "submissions")[item.RequirementID])})
	}
	jsonOut(w, 200, map[string]any{"ok": true, "ownerId": id, "changes": changes, "acknowledgedOperationIds": acknowledged, "serverVersion": st.State.ServerVersion})
}
