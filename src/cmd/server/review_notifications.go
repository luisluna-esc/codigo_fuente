package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

type legacyNotification struct {
	ID                string `json:"id"`
	UserID            string `json:"userId"`
	Type              string `json:"type"`
	Title             string `json:"title"`
	Message           string `json:"message"`
	RequirementID     string `json:"requirementId,omitempty"`
	RequirementTitle  string `json:"requirementTitle,omitempty"`
	ProgramID         string `json:"programId,omitempty"`
	OwnerID           string `json:"ownerId,omitempty"`
	ActorID           string `json:"actorId,omitempty"`
	ActorName         string `json:"actorName,omitempty"`
	Feedback          string `json:"feedback,omitempty"`
	ReviewedAt        string `json:"reviewedAt,omitempty"`
	ServerVersion     int64  `json:"serverVersion,omitempty"`
	SourceOperationID string `json:"sourceOperationId,omitempty"`
	ChangeRequestID   string `json:"changeRequestId,omitempty"`
	CreatedAt         string `json:"createdAt"`
	ReadAt            string `json:"readAt,omitempty"`
	DeletedAt         string `json:"deletedAt,omitempty"`
}

type legacyNotificationReadRequest struct {
	UserID          string   `json:"userId"`
	NotificationIDs []string `json:"notificationIds"`
}

type legacyNotificationDeleteRequest struct {
	UserID          string   `json:"userId"`
	NotificationIDs []string `json:"notificationIds"`
	ReadOnly        bool     `json:"readOnly"`
}

type legacyReviewDecisionRequest struct {
	ExpectedSubmittedAt *string `json:"expectedSubmittedAt,omitempty"`
	ExpectedUpdatedAt   *string `json:"expectedUpdatedAt,omitempty"`
	ActorID             string  `json:"actorId"`
	OwnerID             string  `json:"ownerId"`
	ProgramID           string  `json:"programId"`
	RequirementID       string  `json:"requirementId"`
	RequirementTitle    string  `json:"requirementTitle"`
	Decision            string  `json:"decision"`
	Feedback            string  `json:"feedback"`
	AckOnly             bool    `json:"ackOnly,omitempty"`
}

func legacyRequirementLabel(requirementID, supplied string) string {
	label := strings.TrimSpace(supplied)
	if label == "" {
		label = strings.TrimSpace(requirementID)
	}
	if len([]rune(label)) > 260 {
		label = string([]rune(label)[:260])
	}
	return label
}

func legacyProgressForRequirement(db any, ownerID, programID string, create bool) map[string]any {
	root := legacyMap(db)
	if root == nil || strings.TrimSpace(ownerID) == "" {
		return nil
	}
	programID = strings.TrimSpace(programID)
	if programID == "" || programID == "class_grouped" {
		all := mapField(root, "progress")
		if all == nil && create {
			all = ensureMapField(root, "progress")
		}
		progress := legacyMap(all[ownerID])
		if progress == nil && create {
			progress = map[string]any{"status": map[string]any{}, "submissions": map[string]any{}}
			all[ownerID] = progress
		}
		return progress
	}
	programs := mapField(root, "programProgress")
	if programs == nil && create {
		programs = ensureMapField(root, "programProgress")
	}
	ownerPrograms := legacyMap(programs[ownerID])
	if ownerPrograms == nil && create {
		ownerPrograms = map[string]any{}
		programs[ownerID] = ownerPrograms
	}
	progress := legacyMap(ownerPrograms[programID])
	if progress == nil && create {
		progress = map[string]any{"status": map[string]any{}, "submissions": map[string]any{}}
		ownerPrograms[programID] = progress
	}
	return progress
}

func legacyRequirementStatus(db any, ownerID, programID, requirementID string) string {
	progress := legacyProgressForRequirement(db, ownerID, programID, false)
	return stringValue(mapField(progress, "status")[requirementID])
}

func addLegacyNotification(st *legacyStore, n legacyNotification) {
	if st == nil || strings.TrimSpace(n.UserID) == "" {
		return
	}
	if n.SourceOperationID != "" {
		for _, existing := range st.Notifications {
			if existing.UserID == n.UserID && existing.Type == n.Type && existing.SourceOperationID == n.SourceOperationID {
				return
			}
		}
	}
	if n.ID == "" {
		n.ID = "notice_" + token(12)
	}
	if n.CreatedAt == "" {
		n.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	st.Notifications = append(st.Notifications, n)
	if len(st.Notifications) > 2500 {
		st.Notifications = st.Notifications[len(st.Notifications)-2500:]
	}
}

func activeLegacyReviewerIDs(db any, ownerID string, programIDs ...string) []string {
	ids := []string{}
	program := "class_grouped"
	if len(programIDs) > 0 {
		program = programIDs[0]
	}
	if program == musicProgram553 && len(arrayField(legacyMap(db), "modules")) > 0 {
		owner := legacyUser(db, ownerID)
		for _, raw := range legacyUsers(db) {
			u := legacyMap(raw)
			if stringValue(u["id"]) != ownerID && musicPolicies560[legacyRole(u)].Notices && musicScope550(legacyMap(db), u, owner) {
				ids = append(ids, stringValue(u["id"]))
			}
		}
		return ids
	}
	seen := map[string]bool{}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if !legacyActive(u) || legacyRole(u) != "admin" {
			continue
		}
		id := stringValue(u["id"])
		if id != "" && id != ownerID && !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	owner := legacyUser(db, ownerID)
	programID := "class_grouped"
	if len(programIDs) > 0 {
		programID = programIDs[0]
	}
	if tutorID := legacyTutorForProgram(owner, programID); tutorID != "" && tutorID != ownerID && !seen[tutorID] {
		tutor := legacyUser(db, tutorID)
		if legacyActive(tutor) && legacyRole(tutor) == "tutor" {
			ids = append(ids, tutorID)
		}
	}
	return ids
}

func (s *Server) addAutomaticSyncNotifications(st *legacyStore, req legacySyncRequest, actor map[string]any, role, now string) {
	if st == nil || st.State.Database == nil {
		return
	}
	opType := strings.TrimSpace(req.OperationType)
	ownerID := strings.TrimSpace(req.OwnerID)
	if ownerID == "" && role == "aspirant" {
		ownerID = strings.TrimSpace(req.ActorID)
	}
	reqID := strings.TrimSpace(req.RequirementID)
	programID := strings.TrimSpace(req.ProgramID)
	label := legacyRequirementLabel(reqID, req.RequirementTitle)
	actorName := legacyName(actor)
	if opType == "review" && ownerID != "" && reqID != "" {
		owner := legacyUser(st.State.Database, ownerID)
		ownerName := legacyName(owner)
		if ownerName == "" || ownerName == "<nil>" {
			ownerName = actorName
		}
		for _, reviewerID := range activeLegacyReviewerIDs(st.State.Database, ownerID, programID) {
			if reviewerID == req.ActorID {
				continue
			}
			addLegacyNotification(st, legacyNotification{
				UserID: reviewerID, Type: "review_submitted", Title: "Nuevo requisito para revisión",
				Message: fmt.Sprintf("%s envió a revisión: %s", ownerName, label), RequirementID: reqID,
				RequirementTitle: label, ProgramID: programID, OwnerID: ownerID, ActorID: req.ActorID, ActorName: actorName,
				ServerVersion: st.State.ServerVersion, SourceOperationID: req.OperationID, CreatedAt: now,
			})
		}
		return
	}
	if opType == "review_decision" && ownerID != "" && reqID != "" {
		status := legacyRequirementStatus(st.State.Database, ownerID, programID, reqID)
		feedback := ""
		if progress := legacyProgressForRequirement(st.State.Database, ownerID, programID, false); progress != nil {
			if submission := legacyMap(mapField(progress, "submissions")[reqID]); submission != nil {
				feedback = strings.TrimSpace(stringValue(submission["feedback"]))
			}
		}
		switch status {
		case "approved":
			addLegacyNotification(st, legacyNotification{
				UserID: ownerID, Type: "approved", Title: "Requisito revisado",
				Message: fmt.Sprintf("Tu requisito \"%s\" fue revisado por %s.", label, actorName), RequirementID: reqID,
				RequirementTitle: label, ProgramID: programID, OwnerID: ownerID, ActorID: req.ActorID, ActorName: actorName,
				ReviewedAt: now, ServerVersion: st.State.ServerVersion, SourceOperationID: req.OperationID, CreatedAt: now,
			})
		case "needs_changes":
			addLegacyNotification(st, legacyNotification{
				UserID: ownerID, Type: "needs_changes", Title: "Corrección solicitada",
				Message: fmt.Sprintf("Debes corregir el requisito \"%s\". Revisa las observaciones de %s.", label, actorName), RequirementID: reqID,
				RequirementTitle: label, ProgramID: programID, OwnerID: ownerID, ActorID: req.ActorID, ActorName: actorName,
				Feedback: feedback, ReviewedAt: now, ServerVersion: st.State.ServerVersion, SourceOperationID: req.OperationID, CreatedAt: now,
			})
		}
	}
}

func adminAppliedNotificationRelevant(db any, userID string, n legacyNotification) bool {
	if n.Type != "admin_change_applied" {
		return true
	}
	if strings.TrimSpace(n.OwnerID) != "" {
		return strings.TrimSpace(n.OwnerID) == strings.TrimSpace(userID)
	}
	user := legacyUser(db, userID)
	if user == nil {
		return false
	}
	message := strings.TrimSpace(n.Message)
	name := strings.TrimSpace(legacyName(user))
	if strings.HasPrefix(message, "Usuario actualizado - ") || strings.HasPrefix(message, "Usuario creado - ") {
		return name != "" && strings.Contains(message, " - "+name+" - ")
	}
	if strings.HasPrefix(message, "Clase actualizada - ") {
		for _, raw := range arrayField(legacyMap(db), "classes") {
			class := legacyMap(raw)
			assigned := false
			for _, rawID := range arrayField(class, "assignedUserIds") {
				if strings.TrimSpace(stringValue(rawID)) == strings.TrimSpace(userID) {
					assigned = true
					break
				}
			}
			if !assigned {
				continue
			}
			className := strings.TrimSpace(stringValue(class["name"]))
			if className == "" {
				regular := strings.TrimSpace(stringValue(class["regular"]))
				advanced := strings.TrimSpace(stringValue(class["advanced"]))
				className = strings.TrimSpace(strings.TrimSpace(regular + " + " + advanced))
			}
			if className != "" && strings.Contains(message, className) {
				return true
			}
		}
		return false
	}
	if strings.HasPrefix(message, "Instructor quitado de clase - ") {
		return name != "" && strings.Contains(message, name)
	}
	return true
}

func (s *Server) notificationActor(r *http.Request, requested string, st *legacyStore) (string, string, bool) {
	if sess, remote := s.remoteSessionFromRequest(r); remote {
		if requested != "" && requested != sess.UserID {
			return "", "", false
		}
		return sess.UserID, sess.Role, true
	}
	scope, ok := s.auth(r)
	if !ok || strings.TrimSpace(requested) == "" || st == nil {
		return "", "", false
	}
	u := legacyUser(st.State.Database, requested)
	if !legacyActive(u) {
		return "", "", false
	}
	role := requestRole555(r, u)
	if scope == "admin" && role != "admin" {
		return "", "", false
	}
	if scope == "client" && role != "aspirant" && role != "tutor" {
		return "", "", false
	}
	return requested, role, true
}

func (s *Server) legacyNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	userID, _, ok := s.notificationActor(r, r.URL.Query().Get("userId"), &st)
	if !ok {
		legacyError(w, 403, "No autorizado para consultar avisos.", nil)
		return
	}
	now := time.Now()
	changed := s.ensureLegacyDailyActivityNotifications(&st, now)
	if expireLegacyDailyActivityNotifications(&st, now) {
		changed = true
	}
	if changed {
		if err := s.saveLegacy(st); err != nil {
			legacyError(w, 500, err.Error(), nil)
			return
		}
	}
	items := make([]legacyNotification, 0)
	unread := 0
	for _, n := range st.Notifications {
		if n.UserID != userID || n.DeletedAt != "" || !adminAppliedNotificationRelevant(st.State.Database, userID, n) {
			continue
		}
		if p := programRequest556(r); p != "" && notificationProgram556(n) != "" && notificationProgram556(n) != p {
			continue
		}
		items = append(items, n)
		if n.ReadAt == "" {
			unread++
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt > items[j].CreatedAt })
	if len(items) > 120 {
		items = items[:120]
	}
	jsonOut(w, 200, map[string]any{"ok": true, "notifications": items, "unread": unread, "serverTime": time.Now().UTC().Format(time.RFC3339Nano)})
}

func (s *Server) legacyNotificationRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	var req legacyNotificationReadRequest
	if err := readJSONLimited(r, &req, 1<<20); err != nil {
		legacyError(w, 400, "Lectura de avisos inválida.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	userID, _, ok := s.notificationActor(r, req.UserID, &st)
	if !ok {
		legacyError(w, 403, "No autorizado para marcar avisos.", nil)
		return
	}
	ids := map[string]bool{}
	for _, id := range req.NotificationIDs {
		ids[id] = true
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	count := 0
	for i := range st.Notifications {
		n := &st.Notifications[i]
		if n.UserID == userID && n.DeletedAt == "" && n.ReadAt == "" && (len(ids) == 0 || ids[n.ID]) {
			n.ReadAt = now
			count++
		}
	}
	if count > 0 {
		if err := s.saveLegacy(st); err != nil {
			legacyError(w, 500, err.Error(), nil)
			return
		}
	}
	jsonOut(w, 200, map[string]any{"ok": true, "marked": count})
}

func (s *Server) legacyNotificationDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	var req legacyNotificationDeleteRequest
	if err := readJSONLimited(r, &req, 1<<20); err != nil {
		legacyError(w, 400, "Eliminación de avisos inválida.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	userID, _, ok := s.notificationActor(r, req.UserID, &st)
	if !ok {
		legacyError(w, 403, "No autorizado para eliminar avisos.", nil)
		return
	}
	ids := map[string]bool{}
	for _, id := range req.NotificationIDs {
		ids[strings.TrimSpace(id)] = true
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	deleted := 0
	for i := range st.Notifications {
		n := &st.Notifications[i]
		remove := n.UserID == userID && n.DeletedAt == ""
		if remove && len(ids) > 0 {
			remove = ids[n.ID]
		}
		if remove && req.ReadOnly {
			remove = n.ReadAt != ""
		}
		if remove {
			n.DeletedAt = now
			deleted++
		}
	}
	if deleted > 0 {
		if err := s.saveLegacy(st); err != nil {
			legacyError(w, 500, err.Error(), nil)
			return
		}
	}
	jsonOut(w, 200, map[string]any{"ok": true, "deleted": deleted})
}

func (s *Server) reviewDecisionActor(r *http.Request, requested string, st *legacyStore) (map[string]any, string, bool) {
	if sess, remote := s.remoteSessionFromRequest(r); remote {
		if requested != "" && requested != sess.UserID {
			return nil, "", false
		}
		u := legacyUser(st.State.Database, sess.UserID)
		if !legacyActive(u) || legacyRole(u) != sess.Role || !reviewRole556(sess.Role) {
			return nil, "", false
		}
		return u, sess.Role, true
	}
	scope, ok := s.auth(r)
	if !ok || strings.TrimSpace(requested) == "" {
		return nil, "", false
	}
	u := legacyUser(st.State.Database, requested)
	if !legacyActive(u) {
		return nil, "", false
	}
	role := legacyRole(u)
	if scope == "admin" && reviewRole556(role) {
		return u, role, true
	}
	if scope == "client" && role == "tutor" {
		return u, role, true
	}
	return nil, "", false
}

func (s *Server) legacyReviewDecision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	var req legacyReviewDecisionRequest
	if err := readJSONLimited(r, &req, 1<<20); err != nil {
		legacyError(w, 400, "Decisión de requisito inválida.", nil)
		return
	}
	req.Decision = strings.TrimSpace(req.Decision)
	req.OwnerID = strings.TrimSpace(req.OwnerID)
	req.RequirementID = strings.TrimSpace(req.RequirementID)
	req.ProgramID = strings.TrimSpace(req.ProgramID)
	if req.ProgramID == "" {
		req.ProgramID = "class_grouped"
	}
	if (req.Decision != "approved" && req.Decision != "needs_changes") || req.OwnerID == "" || req.RequirementID == "" {
		legacyError(w, 400, "Faltan el Aspirante, requisito o decisión.", nil)
		return
	}
	if req.Decision == "needs_changes" && strings.TrimSpace(req.Feedback) == "" {
		legacyError(w, 400, "Escribe la corrección solicitada.", nil)
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
	reviewer, _, ok := s.reviewDecisionActor(r, req.ActorID, &st)
	if !ok {
		legacyError(w, 403, "No autorizado para revisar este requisito.", nil)
		return
	}
	if req.ProgramID == musicProgram553 {
		key := "reports.correct"
		if req.Decision == "approved" {
			key = "reports.approve"
		}
		if !musicPermit602(legacyMap(st.State.Database), reviewer, key) {
			legacyError(w, 403, "Función desactivada.", nil)
			return
		}
	}
	reviewerID := stringValue(reviewer["id"])
	owner := legacyUser(st.State.Database, req.OwnerID)
	if !leader555(legacyMap(st.State.Database), owner) {
		legacyError(w, 404, "El Aspirante no está disponible.", nil)
		return
	}
	if reviewerID == req.OwnerID {
		legacyError(w, 403, "Otra autoridad debe revisar tu tarjeta personal.", nil)
		return
	}
	if !canReview556(legacyMap(st.State.Database), reviewer, owner, req.ProgramID) {
		legacyError(w, 403, "El Aspirante no está asignado a este Tutor.", nil)
		return
	}
	// Mutate a detached copy so a disk failure cannot publish an uncommitted review.
	st.State.Database = cloneAny(st.State.Database)
	st.Notifications = append([]legacyNotification(nil), st.Notifications...)
	progress := legacyProgressForRequirement(st.State.Database, req.OwnerID, req.ProgramID, false)
	if progress == nil {
		legacyError(w, 404, "No se encontró el avance del requisito.", nil)
		return
	}
	statuses := mapField(progress, "status")
	current := stringValue(statuses[req.RequirementID])
	if current == req.Decision {
		if req.AckOnly {
			jsonOut(w, 200, map[string]any{
				"ok": true, "status": current, "idempotent": true, "ackOnly": true,
				"ownerId": req.OwnerID, "requirementId": req.RequirementID, "programId": req.ProgramID,
				"serverVersion": st.State.ServerVersion,
			})
			return
		}
		payload := s.legacySyncSnapshotForRequest(st.State, r)
		payload["status"] = current
		payload["idempotent"] = true
		jsonOut(w, 200, payload)
		return
	}
	submissions := ensureMapField(progress, "submissions")
	submission := legacyMap(submissions[req.RequirementID])
	if submission == nil {
		submission = map[string]any{}
	}
	if current != "in_review" {
		legacyError(w, 409, "El requisito todavía no fue enviado a revisión. Actualiza la pantalla.", map[string]any{"status": current, "serverVersion": st.State.ServerVersion})
		return
	}
	if (req.ExpectedUpdatedAt != nil && *req.ExpectedUpdatedAt != stringValue(submission["updatedAt"])) || (req.ExpectedSubmittedAt != nil && *req.ExpectedSubmittedAt != stringValue(submission["submittedAt"])) {
		legacyError(w, 409, "El informe cambió desde que lo abriste. Vuelve a abrirlo y revisa la versión actual antes de decidir.", map[string]any{"code": "STALE_REVIEW", "status": current, "serverVersion": st.State.ServerVersion})
		return
	}
	if ok, reason := reviewStoredSubmissionIsInternalReport(submission, st.State.Files); !ok {
		legacyError(w, 409, reason, map[string]any{"status": current, "serverVersion": st.State.ServerVersion})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	submission["feedback"] = strings.TrimSpace(req.Feedback)
	submission["updatedAt"] = now
	if req.Decision == "approved" {
		submission["approval"] = map[string]any{"reviewerId": reviewerID, "reviewerName": legacyName(reviewer), "at": now}
	} else {
		submission["approval"] = nil
	}
	submissions[req.RequirementID] = submission
	statuses[req.RequirementID] = req.Decision
	progress["updatedAt"] = now

	db := legacyMap(st.State.Database)
	db["updatedAt"] = now
	db["revision"] = numericValue(db["revision"]) + 1
	st.State.ServerVersion++
	message := "Requisito revisado"
	if req.Decision == "approved" {
		message = "Requisito revisado"
	}
	st.State.Database = legacyPrepareCanonicalDB(db, st.State, message)
	st.State.UpdatedAt = now
	label := legacyRequirementLabel(req.RequirementID, req.RequirementTitle)
	if req.Decision == "approved" {
		addLegacyNotification(&st, legacyNotification{
			UserID: req.OwnerID, Type: "approved", Title: "Requisito revisado",
			Message:       fmt.Sprintf("Tu requisito \"%s\" fue revisado por %s.", label, legacyName(reviewer)),
			RequirementID: req.RequirementID, RequirementTitle: label, ProgramID: req.ProgramID,
			OwnerID: req.OwnerID, ActorID: reviewerID, ActorName: legacyName(reviewer), ReviewedAt: now, ServerVersion: st.State.ServerVersion, CreatedAt: now,
		})
	} else {
		feedback := strings.TrimSpace(req.Feedback)
		message := fmt.Sprintf("Debes corregir el requisito \"%s\".", label)
		if feedback != "" {
			message += " Observación: " + feedback
		}
		addLegacyNotification(&st, legacyNotification{
			UserID: req.OwnerID, Type: "needs_changes", Title: "Corrección solicitada", Message: message,
			RequirementID: req.RequirementID, RequirementTitle: label, ProgramID: req.ProgramID, OwnerID: req.OwnerID,
			ActorID: reviewerID, ActorName: legacyName(reviewer), Feedback: feedback, ReviewedAt: now, ServerVersion: st.State.ServerVersion, CreatedAt: now,
		})
	}
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	if req.AckOnly {
		jsonOut(w, 200, map[string]any{
			"ok": true, "status": req.Decision, "ackOnly": true,
			"ownerId": req.OwnerID, "requirementId": req.RequirementID, "programId": req.ProgramID,
			"feedback": strings.TrimSpace(req.Feedback), "reviewerId": reviewerID, "reviewerName": legacyName(reviewer),
			"reviewedAt": now, "serverVersion": st.State.ServerVersion,
		})
		return
	}
	payload := s.legacySyncSnapshotForRequest(st.State, r)
	payload["status"] = req.Decision
	payload["ownerId"] = req.OwnerID
	payload["requirementId"] = req.RequirementID
	payload["programId"] = req.ProgramID
	payload["feedback"] = strings.TrimSpace(req.Feedback)
	payload["reviewerId"] = reviewerID
	payload["reviewerName"] = legacyName(reviewer)
	payload["reviewedAt"] = now
	jsonOut(w, 200, payload)
}
