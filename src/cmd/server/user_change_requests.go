package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// legacyUserChangeRequest represents a correction proposed by an Administrator
// over information owned/filled by a normal user. The canonical database is not
// changed until the owner explicitly accepts it.
type legacyUserChangeRequest struct {
	ID           string `json:"id"`
	ProgramID    string `json:"programId,omitempty"`
	Status       string `json:"status"`
	OwnerID      string `json:"ownerId"`
	ActorID      string `json:"actorId"`
	ActorName    string `json:"actorName"`
	OperationID  string `json:"operationId,omitempty"`
	Summary      string `json:"summary"`
	Patch        any    `json:"patch"`
	Files        []any  `json:"files,omitempty"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
	DecidedAt    string `json:"decidedAt,omitempty"`
	Decision     string `json:"decision,omitempty"`
	DecisionNote string `json:"decisionNote,omitempty"`
}

type legacyAdminChangeSubmitRequest struct {
	InstallationID    string `json:"installationId"`
	ActorID           string `json:"actorId"`
	OperationID       string `json:"operationId"`
	EntityID          string `json:"entityId,omitempty"`
	ProgramID         string `json:"programId,omitempty"`
	Summary           string `json:"summary"`
	Patch             any    `json:"patch"`
	Files             []any  `json:"files"`
	AckOnly           bool   `json:"ackOnly,omitempty"`
	RestoreGeneration int64  `json:"restoreGeneration,omitempty"`
}

type legacyUserChangeDecisionRequest struct {
	ActorID    string   `json:"actorId"`
	RequestIDs []string `json:"requestIds"`
	Decision   string   `json:"decision"`
	Note       string   `json:"note"`
	AllPending bool     `json:"allPending"`
}

var adminOwnedUserFields = map[string]bool{
	"id": true, "role": true, "active": true, "accountState": true,
	"clubId": true, "tutorId": true, "tutorByProgram": true, "password": true, "mustChangePassword": true,
	"recoveryKeys": true, "createdAt": true, "requirementViewPrograms": true,
}

func appendPatchArray(patch map[string]any, key string, item any) {
	current, _ := patch[key].([]any)
	patch[key] = append(current, cloneAny(item))
}

func addOwnerMapPatch(ownerPatches map[string]map[string]any, ownerID, key string, value any) {
	if strings.TrimSpace(ownerID) == "" {
		return
	}
	patch := ownerPatches[ownerID]
	if patch == nil {
		patch = map[string]any{}
		ownerPatches[ownerID] = patch
	}
	current := legacyMap(patch[key])
	if current == nil {
		current = map[string]any{}
	}
	incoming := legacyMap(value)
	if incoming != nil {
		for k, v := range incoming {
			current[k] = cloneAny(v)
		}
	}
	patch[key] = current
}

func singleAssignedOwner(db any, classPatch map[string]any) string {
	classID := stringValue(classPatch["id"])
	if classID == "" {
		return ""
	}
	var current map[string]any
	for _, raw := range arrayField(legacyMap(db), "classes") {
		item := legacyMap(raw)
		if stringValue(item["id"]) == classID {
			current = item
			break
		}
	}
	if current == nil {
		return ""
	}
	ids := arrayField(current, "assignedUserIds")
	owner := ""
	for _, raw := range ids {
		id := strings.TrimSpace(stringValue(raw))
		user := legacyUser(db, id)
		if id == "" || user == nil || !legacyActive(user) || legacyRole(user) == "admin" {
			continue
		}
		if owner != "" && owner != id {
			return "" // shared class: do not assign ownership to one person arbitrarily.
		}
		owner = id
	}
	return owner
}

// splitAdminPatchByOwnership separates fields owned by the server/Administrator
// from fields filled by a normal user. User-owned pieces become consent requests.
func splitAdminPatchByOwnership(serverDB, patch any) (map[string]any, map[string]map[string]any) {
	direct := map[string]any{}
	owners := map[string]map[string]any{}
	pm := legacyMap(patch)
	if pm == nil {
		return direct, owners
	}

	for key, raw := range pm {
		switch key {
		case "users":
			items, _ := raw.([]any)
			for _, rawUser := range items {
				incoming := legacyMap(rawUser)
				if incoming == nil {
					continue
				}
				id := strings.TrimSpace(stringValue(incoming["id"]))
				existing := legacyUser(serverDB, id)
				if existing == nil || legacyRole(existing) == "admin" {
					appendPatchArray(direct, "users", incoming)
					continue
				}
				directUser := map[string]any{"id": id}
				ownedUser := map[string]any{"id": id}
				for field, value := range incoming {
					if adminOwnedUserFields[field] {
						directUser[field] = cloneAny(value)
					} else {
						ownedUser[field] = cloneAny(value)
					}
				}
				if len(directUser) > 1 {
					appendPatchArray(direct, "users", directUser)
				}
				if len(ownedUser) > 1 {
					patchForOwner := owners[id]
					if patchForOwner == nil {
						patchForOwner = map[string]any{}
						owners[id] = patchForOwner
					}
					appendPatchArray(patchForOwner, "users", ownedUser)
				}
			}
		case "progress", "programProgress", "bibleYear":
			values := legacyMap(raw)
			if values == nil {
				direct[key] = cloneAny(raw)
				continue
			}
			directValues := map[string]any{}
			for ownerID, value := range values {
				user := legacyUser(serverDB, ownerID)
				if user != nil && legacyActive(user) && legacyRole(user) != "admin" {
					addOwnerMapPatch(owners, ownerID, key, map[string]any{ownerID: value})
				} else {
					directValues[ownerID] = cloneAny(value)
				}
			}
			if len(directValues) > 0 {
				direct[key] = directValues
			}
		case "configs":
			configs := legacyMap(raw)
			if configs == nil {
				direct[key] = cloneAny(raw)
				continue
			}
			directConfigs := map[string]any{}
			for configKey, value := range configs {
				ownerID := ""
				for _, rawUser := range legacyUsers(serverDB) {
					user := legacyMap(rawUser)
					id := stringValue(user["id"])
					if id == "" || legacyRole(user) == "admin" {
						continue
					}
					if configKey == id || strings.HasPrefix(configKey, id+":") || strings.Contains(configKey, ":"+id+":") {
						ownerID = id
						break
					}
				}
				if ownerID != "" {
					patchForOwner := owners[ownerID]
					if patchForOwner == nil {
						patchForOwner = map[string]any{}
						owners[ownerID] = patchForOwner
					}
					cfg := legacyMap(patchForOwner["configs"])
					if cfg == nil {
						cfg = map[string]any{}
					}
					cfg[configKey] = cloneAny(value)
					patchForOwner["configs"] = cfg
				} else {
					directConfigs[configKey] = cloneAny(value)
				}
			}
			if len(directConfigs) > 0 {
				direct[key] = directConfigs
			}
		case "classes":
			items, _ := raw.([]any)
			for _, rawClass := range items {
				classPatch := legacyMap(rawClass)
				ownerID := singleAssignedOwner(serverDB, classPatch)
				// Structural class configuration remains administrative. Only work
				// captured by the assigned user requires consent.
				if ownerID != "" && classPatch["work"] != nil {
					owned := map[string]any{"id": classPatch["id"], "work": cloneAny(classPatch["work"])}
					patchForOwner := owners[ownerID]
					if patchForOwner == nil {
						patchForOwner = map[string]any{}
						owners[ownerID] = patchForOwner
					}
					appendPatchArray(patchForOwner, "classes", owned)
					adminClass := legacyMap(cloneAny(classPatch))
					delete(adminClass, "work")
					if len(adminClass) > 1 {
						appendPatchArray(direct, "classes", adminClass)
					}
				} else {
					appendPatchArray(direct, "classes", classPatch)
				}
			}
		case "audit":
			// Audit records describe the Administrator action and may be saved now;
			// they do not alter user-owned content.
			direct[key] = cloneAny(raw)
		default:
			direct[key] = cloneAny(raw)
		}
	}
	return direct, owners
}

type legacyTaskAssignment struct {
	OwnerID       string
	ProgramID     string
	RequirementID string
}

func adminTaskAssignmentsFromPatch(patch any) []legacyTaskAssignment {
	pm := legacyMap(patch)
	if pm == nil {
		return nil
	}
	out := []legacyTaskAssignment{}
	for _, raw := range arrayField(pm, "tasks") {
		task := legacyMap(raw)
		ownerID := strings.TrimSpace(stringValue(task["assigneeId"]))
		reqID := strings.TrimSpace(stringValue(task["requirementId"]))
		programID := strings.TrimSpace(stringValue(task["programId"]))
		if programID == "" {
			programID = "class_grouped"
		}
		if ownerID != "" && reqID != "" {
			out = append(out, legacyTaskAssignment{OwnerID: ownerID, ProgramID: programID, RequirementID: reqID})
		}
	}
	return out
}

func compactLegacyMap(m map[string]any) bool {
	if len(m) == 0 {
		return false
	}
	for key, value := range m {
		if child := legacyMap(value); child != nil {
			if !compactLegacyMap(child) {
				delete(m, key)
			}
			continue
		}
		if arr, ok := value.([]any); ok && len(arr) == 0 {
			delete(m, key)
		}
	}
	return len(m) > 0
}

func removeAssignedStatusFromProgress(progress any, requirementID string) bool {
	pm := legacyMap(progress)
	if pm == nil {
		return false
	}
	status := legacyMap(pm["status"])
	if status == nil || stringValue(status[requirementID]) != "assigned" {
		return false
	}
	delete(status, requirementID)
	if len(status) == 0 {
		delete(pm, "status")
	}
	compactLegacyMap(pm)
	return true
}

func addDirectAssignedStatus(direct map[string]any, ownerID, programID, requirementID string) {
	if strings.TrimSpace(ownerID) == "" || strings.TrimSpace(requirementID) == "" {
		return
	}
	if programID == "" || programID == "class_grouped" {
		progressRoot := legacyMap(direct["progress"])
		if progressRoot == nil {
			progressRoot = map[string]any{}
			direct["progress"] = progressRoot
		}
		progress := legacyMap(progressRoot[ownerID])
		if progress == nil {
			progress = map[string]any{}
			progressRoot[ownerID] = progress
		}
		status := legacyMap(progress["status"])
		if status == nil {
			status = map[string]any{}
			progress["status"] = status
		}
		status[requirementID] = "assigned"
		return
	}
	programRoot := legacyMap(direct["programProgress"])
	if programRoot == nil {
		programRoot = map[string]any{}
		direct["programProgress"] = programRoot
	}
	ownerPrograms := legacyMap(programRoot[ownerID])
	if ownerPrograms == nil {
		ownerPrograms = map[string]any{}
		programRoot[ownerID] = ownerPrograms
	}
	progress := legacyMap(ownerPrograms[programID])
	if progress == nil {
		progress = map[string]any{}
		ownerPrograms[programID] = progress
	}
	status := legacyMap(progress["status"])
	if status == nil {
		status = map[string]any{}
		progress["status"] = status
	}
	status[requirementID] = "assigned"
}

func moveTaskAssignmentProgressToDirect(direct map[string]any, owners map[string]map[string]any, patch any) {
	for _, assignment := range adminTaskAssignmentsFromPatch(patch) {
		ownerPatch := owners[assignment.OwnerID]
		if ownerPatch == nil {
			continue
		}
		moved := false
		if assignment.ProgramID == "" || assignment.ProgramID == "class_grouped" {
			progressRoot := legacyMap(ownerPatch["progress"])
			ownerProgress := legacyMap(progressRoot[assignment.OwnerID])
			if removeAssignedStatusFromProgress(ownerProgress, assignment.RequirementID) {
				moved = true
				compactLegacyMap(progressRoot)
				if len(progressRoot) == 0 {
					delete(ownerPatch, "progress")
				}
			}
		} else {
			programRoot := legacyMap(ownerPatch["programProgress"])
			ownerPrograms := legacyMap(programRoot[assignment.OwnerID])
			progress := legacyMap(ownerPrograms[assignment.ProgramID])
			if removeAssignedStatusFromProgress(progress, assignment.RequirementID) {
				moved = true
				compactLegacyMap(ownerPrograms)
				compactLegacyMap(programRoot)
				if len(programRoot) == 0 {
					delete(ownerPatch, "programProgress")
				}
			}
		}
		if moved {
			addDirectAssignedStatus(direct, assignment.OwnerID, assignment.ProgramID, assignment.RequirementID)
			if !patchHasContent(ownerPatch) {
				delete(owners, assignment.OwnerID)
			}
		}
	}
}

func activeModuleAssignees(db any, programID string) []string {
	seen := map[string]bool{}
	out := []string{}
	programID = strings.TrimSpace(programID)
	if programID == "" {
		programID = "class_grouped"
	}
	for _, rawModule := range arrayField(legacyMap(db), "modules") {
		module := legacyMap(rawModule)
		if stringValue(module["id"]) != programID {
			continue
		}
		for _, rawID := range arrayField(module, "assignedUserIds") {
			id := strings.TrimSpace(stringValue(rawID))
			user := legacyUser(db, id)
			if id == "" || seen[id] || !legacyActive(user) || legacyRole(user) == "admin" {
				continue
			}
			seen[id] = true
			out = append(out, id)
		}
	}
	if len(out) > 0 {
		return out
	}
	for _, rawUser := range legacyUsers(db) {
		user := legacyMap(rawUser)
		id := strings.TrimSpace(stringValue(user["id"]))
		if id == "" || seen[id] || !legacyActive(user) || legacyRole(user) != "aspirant" {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func addDirectAdminChangeNotifications(st *legacyStore, patch map[string]any, actor map[string]any, operationID, now string) {
	if st == nil || st.State.Database == nil || patch == nil {
		return
	}
	actorID := stringValue(actor["id"])
	actorName := legacyName(actor)
	for _, rawTask := range arrayField(patch, "tasks") {
		task := legacyMap(rawTask)
		assigneeID := strings.TrimSpace(stringValue(task["assigneeId"]))
		if assigneeID == "" || !legacyActive(legacyUser(st.State.Database, assigneeID)) {
			continue
		}
		reqID := strings.TrimSpace(stringValue(task["requirementId"]))
		programID := strings.TrimSpace(stringValue(task["programId"]))
		if programID == "" {
			programID = "class_grouped"
		}
		label := legacyRequirementLabel(reqID, reqID)
		due := strings.TrimSpace(stringValue(task["dueDate"]))
		msg := fmt.Sprintf("%s te asignó una tarea: %s", actorName, label)
		if due != "" {
			msg += " · Presentación: " + due
		}
		addLegacyNotification(st, legacyNotification{UserID: assigneeID, Type: "task_assigned", Title: "Tarea asignada", Message: msg, RequirementID: reqID, RequirementTitle: label, ProgramID: programID, ActorID: actorID, ActorName: actorName, SourceOperationID: operationID + ":task:" + stringValue(task["id"]), CreatedAt: now})
	}
	for _, rawMaterial := range arrayField(patch, "materials") {
		material := legacyMap(rawMaterial)
		if material == nil {
			continue
		}
		if active, ok := material["active"].(bool); ok && !active {
			continue
		}
		programID := strings.TrimSpace(stringValue(material["programId"]))
		if programID == "" {
			programID = "class_grouped"
		}
		reqID := strings.TrimSpace(stringValue(material["requirementId"]))
		title := strings.TrimSpace(stringValue(material["title"]))
		if title == "" {
			title = "Material de apoyo"
		}
		for _, userID := range activeModuleAssignees(st.State.Database, programID) {
			if userID == actorID {
				continue
			}
			addLegacyNotification(st, legacyNotification{UserID: userID, Type: "material_published", Title: "Nuevo material disponible", Message: fmt.Sprintf("%s publicó material: %s", actorName, title), RequirementID: reqID, RequirementTitle: legacyRequirementLabel(reqID, reqID), ProgramID: programID, ActorID: actorID, ActorName: actorName, SourceOperationID: operationID + ":material:" + stringValue(material["id"]), CreatedAt: now})
		}
	}
}

func patchHasContent(p map[string]any) bool {
	if len(p) == 0 {
		return false
	}
	for key, value := range p {
		if key == "audit" {
			continue
		}
		if m := legacyMap(value); m != nil && len(m) == 0 {
			continue
		}
		if a, ok := value.([]any); ok && len(a) == 0 {
			continue
		}
		return true
	}
	return false
}

func userChangesByOperation(st *legacyStore, actorID, operationID string) []legacyUserChangeRequest {
	if st == nil || strings.TrimSpace(operationID) == "" {
		return nil
	}
	out := []legacyUserChangeRequest{}
	for i := range st.UserChanges {
		item := st.UserChanges[i]
		if item.ActorID == actorID && item.OperationID == operationID {
			out = append(out, item)
		}
	}
	return out
}

func findUserChangeByOperation(st *legacyStore, actorID, operationID string) *legacyUserChangeRequest {
	if st == nil || strings.TrimSpace(operationID) == "" {
		return nil
	}
	for i := range st.UserChanges {
		item := &st.UserChanges[i]
		if item.ActorID == actorID && item.OperationID == operationID {
			return item
		}
	}
	return nil
}

func addAdminAppliedOwner(out map[string]bool, db any, userID string) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return
	}
	user := legacyUser(db, userID)
	if !legacyActive(user) || legacyRole(user) == "admin" {
		return
	}
	out[userID] = true
}

func addAssignedClassOwners(out map[string]bool, db any, rawClass map[string]any) {
	if rawClass == nil {
		return
	}
	classID := strings.TrimSpace(stringValue(rawClass["id"]))
	if classID == "" {
		return
	}
	// Include the previous assignment so a removed Instructor can still receive
	// the notice, plus the incoming assignment so a newly assigned Instructor does too.
	for _, raw := range arrayField(legacyMap(db), "classes") {
		item := legacyMap(raw)
		if strings.TrimSpace(stringValue(item["id"])) != classID {
			continue
		}
		for _, userID := range arrayField(item, "assignedUserIds") {
			addAdminAppliedOwner(out, db, stringValue(userID))
		}
		break
	}
	for _, userID := range arrayField(rawClass, "assignedUserIds") {
		addAdminAppliedOwner(out, db, stringValue(userID))
	}
}

func sortedAdminOwnerIDs(owners map[string]bool) []string {
	ids := make([]string, 0, len(owners))
	for id := range owners {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func collectAdminAppliedOwners(db any, patch map[string]any, entityID string) []string {
	owners := map[string]bool{}
	entityID = strings.TrimSpace(entityID)
	if entityID != "" {
		// A user edit belongs only to that user, even if the local diff carried
		// unrelated rows from an older baseline.
		if user := legacyUser(db, entityID); user != nil {
			addAdminAppliedOwner(owners, db, entityID)
			return sortedAdminOwnerIDs(owners)
		}
		for _, rawClass := range arrayField(patch, "classes") {
			classPatch := legacyMap(rawClass)
			if strings.TrimSpace(stringValue(classPatch["id"])) == entityID {
				addAssignedClassOwners(owners, db, classPatch)
				return sortedAdminOwnerIDs(owners)
			}
		}
		// Tasks and materials already have dedicated notifications. If entityId
		// identifies one of them, do not add a second generic admin-change notice.
		for _, key := range []string{"tasks", "materials"} {
			for _, raw := range arrayField(patch, key) {
				if strings.TrimSpace(stringValue(legacyMap(raw)["id"])) == entityID {
					return []string{}
				}
			}
		}
	}
	for _, rawUser := range arrayField(patch, "users") {
		addAdminAppliedOwner(owners, db, stringValue(legacyMap(rawUser)["id"]))
	}
	for _, key := range []string{"progress", "programProgress", "bibleYear"} {
		for userID := range legacyMap(patch[key]) {
			addAdminAppliedOwner(owners, db, userID)
		}
	}
	for key := range legacyMap(patch["configs"]) {
		parts := strings.Split(key, ":")
		if len(parts) > 0 {
			addAdminAppliedOwner(owners, db, parts[0])
		}
	}
	for _, rawTask := range arrayField(patch, "tasks") {
		addAdminAppliedOwner(owners, db, stringValue(legacyMap(rawTask)["assigneeId"]))
	}
	for _, rawClass := range arrayField(patch, "classes") {
		addAssignedClassOwners(owners, db, legacyMap(rawClass))
	}
	return sortedAdminOwnerIDs(owners)
}

func addAppliedAdminChangeNotifications(st *legacyStore, ownerIDs []string, actor map[string]any, operationID, summary, now string) {
	if st == nil || st.State.Database == nil || len(ownerIDs) == 0 {
		return
	}
	actorID := stringValue(actor["id"])
	actorName := legacyName(actor)
	if strings.TrimSpace(summary) == "" {
		summary = "El Administrador realizó un cambio en tus datos. No requiere aprobación."
	}
	for _, ownerID := range ownerIDs {
		owner := legacyUser(st.State.Database, ownerID)
		if !legacyActive(owner) || legacyRole(owner) == "admin" {
			continue
		}
		addLegacyNotification(st, legacyNotification{
			UserID: ownerID, OwnerID: ownerID, Type: "admin_change_applied", Title: "Cambio realizado por el Administrador",
			Message: summary, ActorID: actorID, ActorName: actorName, SourceOperationID: operationID,
			CreatedAt: now,
		})
	}
}

func closeLegacyUserChangesForOperation(st *legacyStore, actorID, operationID, now string) {
	if st == nil || strings.TrimSpace(operationID) == "" {
		return
	}
	for i := range st.UserChanges {
		if st.UserChanges[i].ActorID == actorID && st.UserChanges[i].OperationID == operationID && st.UserChanges[i].Status == "pending" {
			st.UserChanges[i].Status = "applied"
			st.UserChanges[i].UpdatedAt = now
			st.UserChanges[i].Decision = "applied_directly"
			st.UserChanges[i].DecisionNote = "Aplicado automaticamente por regla 4.2.21: el Administrador no requiere aprobacion del usuario."
		}
	}
}

func (s *Server) legacyAdminChangeSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	remoteActorID := ""
	if sess, remote := s.remoteSessionFromRequest(r); remote {
		if sess.Role != "admin" {
			legacyError(w, 403, "Esta operación requiere una sesión de Administrador.", nil)
			return
		}
		remoteActorID = strings.TrimSpace(sess.UserID)
	} else {
		scope, ok := s.auth(r)
		if !ok || scope != "admin" {
			legacyError(w, 403, "Esta operación requiere el token de Administrador.", nil)
			return
		}
	}
	var req legacyAdminChangeSubmitRequest
	if err := readJSONLimited(r, &req, 96<<20); err != nil || (req.Patch == nil && len(req.Files) == 0) {
		legacyError(w, 400, "No se pudo leer el cambio del Administrador.", nil)
		return
	}
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.OperationID = strings.TrimSpace(req.OperationID)
	if req.ActorID == "" || req.OperationID == "" {
		legacyError(w, 400, "Faltan actorId u operationId.", nil)
		return
	}
	if remoteActorID != "" && req.ActorID != remoteActorID {
		legacyError(w, 403, "La sesión no corresponde al Administrador que envía el cambio.", nil)
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
		legacyError(w, 409, "El servidor todavía no fue inicializado.", map[string]any{"serverVersion": int64(0)})
		return
	}
	actor := legacyUser(st.State.Database, req.ActorID)
	if !legacyActive(actor) || legacyRole(actor) != "admin" {
		legacyError(w, 403, "El usuario no es un Administrador activo.", nil)
		return
	}
	if req.RestoreGeneration > 0 && req.RestoreGeneration != legacyCanonicalRestoreGeneration(st.State) {
		legacyError(w, 409, "El servidor fue restaurado. Se conserva el cambio local hasta recuperar la versión vigente.", map[string]any{"code": "RESTORE_REQUIRED", "restoreGeneration": legacyCanonicalRestoreGeneration(st.State), "restoreId": st.State.RestoreID, "restoredAt": st.State.RestoredAt, "serverVersion": st.State.ServerVersion})
		return
	}
	if previous := findAppliedLegacyOperation(&st, req.ActorID, req.OperationID); previous != nil {
		jsonOut(w, 200, map[string]any{"ok": true, "status": "applied", "operationId": req.OperationID, "requestIds": []string{}, "serverVersion": st.State.ServerVersion, "deduplicated": true, "ackOnly": true})
		return
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	patch := cloneAny(req.Patch)
	patchMap := legacyMap(patch)
	if patchMap == nil {
		patchMap = map[string]any{}
	}
	if musicRequest553(r) {
		prepareMusicAdminAssignments559(patchMap, legacyMap(st.State.Database))
	}
	hasPatch := patchHasContent(patchMap)
	hasFiles := len(req.Files) > 0
	ownerIDs := []string{}
	beforeTutorAssignments := cloneAny(st.State.Database)
	if hasPatch {
		ownerIDs = collectAdminAppliedOwners(st.State.Database, patchMap, req.EntityID)
		st.State.Database = expandLegacyPatch(st.State.Database, patchMap)
	}
	if hasFiles {
		// 5.0.10: permite reparar un material histórico cuyo catálogo llegó al
		// servidor pero cuyo archivo quedó únicamente en el equipo Administrador.
		// Una reparación solo de archivos no vuelve a publicar avisos.
		st.State.Files = mergeLegacyFilesUnrestricted(st.State.Files, req.Files)
	}
	if hasPatch || hasFiles {
		st.State.ServerVersion++
		db := legacyMap(st.State.Database)
		db["updatedAt"] = now
		db["revision"] = numericValue(db["revision"]) + 1
		st.State.Database = legacyPrepareCanonicalDB(db, st.State, "Cambio administrativo aplicado")
		st.State.UpdatedAt = now
		if hasPatch {
			addTutorAssignmentNotifications(&st, beforeTutorAssignments, actor, req.OperationID, now)
			addDirectAdminChangeNotifications(&st, patchMap, actor, req.OperationID, now)
			addAppliedAdminChangeNotifications(&st, ownerIDs, actor, req.OperationID, strings.TrimSpace(req.Summary), now)
			program := req.ProgramID
			if program == "" {
				program = programRequest556(r)
			}
			stampAdminNotifications556(&st, req.OperationID, program)
		}
	}
	closeLegacyUserChangesForOperation(&st, req.ActorID, req.OperationID, now)
	rememberAppliedLegacyOperation(&st, legacyAppliedOperation{OperationID: req.OperationID, ActorID: req.ActorID, OperationType: "admin_change", ProgramID: req.ProgramID, AppliedAt: now, ServerVersion: st.State.ServerVersion})
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "status": "applied", "operationId": req.OperationID, "requestIds": []string{}, "serverVersion": st.State.ServerVersion, "ackOnly": true})
}

func (s *Server) legacyUserChanges(w http.ResponseWriter, r *http.Request) {
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
	actorID, _, ok := s.notificationActor(r, r.URL.Query().Get("actorId"), &st)
	if !ok {
		legacyError(w, 403, "No autorizado para consultar cambios.", nil)
		return
	}
	statusFilter := strings.TrimSpace(r.URL.Query().Get("status"))
	items := []legacyUserChangeRequest{}
	for _, item := range st.UserChanges {
		if p := programRequest556(r); p != "" {
			ip := item.ProgramID
			if ip == "" {
				ip = "class_grouped"
			}
			if ip != p {
				continue
			}
		}
		if item.OwnerID != actorID {
			continue
		}
		if statusFilter != "" && item.Status != statusFilter {
			continue
		}
		items = append(items, item)
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt > items[j].CreatedAt })
	if len(items) > 150 {
		items = items[:150]
	}
	jsonOut(w, 200, map[string]any{"ok": true, "changes": items, "serverVersion": st.State.ServerVersion})
}

func (s *Server) legacyUserChangeDecision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	var req legacyUserChangeDecisionRequest
	if err := readJSONLimited(r, &req, 2<<20); err != nil {
		legacyError(w, 400, "Decisión inválida.", nil)
		return
	}
	req.Decision = strings.TrimSpace(req.Decision)
	if req.Decision != "accept" && req.Decision != "reject" {
		legacyError(w, 400, "La decisión debe ser accept o reject.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	actorID, _, ok := s.notificationActor(r, req.ActorID, &st)
	if !ok {
		legacyError(w, 403, "No autorizado para decidir estos cambios.", nil)
		return
	}
	wanted := map[string]bool{}
	for _, id := range req.RequestIDs {
		wanted[strings.TrimSpace(id)] = true
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	decided := []string{}
	accepted := 0
	for i := range st.UserChanges {
		item := &st.UserChanges[i]
		if item.OwnerID != actorID || item.Status != "pending" {
			continue
		}
		if !req.AllPending && !wanted[item.ID] {
			continue
		}
		if req.Decision == "accept" {
			st.State.Database = expandLegacyPatch(st.State.Database, item.Patch)
			st.State.Files = mergeLegacyFilesUnrestricted(st.State.Files, item.Files)
			item.Status = "accepted"
			item.Decision = "accept"
			accepted++
		} else {
			item.Status = "rejected"
			item.Decision = "reject"
		}
		item.DecisionNote = strings.TrimSpace(req.Note)
		item.DecidedAt = now
		item.UpdatedAt = now
		decided = append(decided, item.ID)
		addLegacyNotification(&st, legacyNotification{
			UserID: item.ActorID, Type: map[bool]string{true: "admin_change_accepted", false: "admin_change_rejected"}[req.Decision == "accept"],
			Title:   map[bool]string{true: "Cambio aceptado por el usuario", false: "Cambio rechazado por el usuario"}[req.Decision == "accept"],
			Message: fmt.Sprintf("%s: %s", legacyName(legacyUser(st.State.Database, actorID)), item.Summary),
			ActorID: actorID, ActorName: legacyName(legacyUser(st.State.Database, actorID)), ChangeRequestID: item.ID, ProgramID: item.ProgramID, CreatedAt: now,
		})
	}
	if len(decided) == 0 {
		legacyError(w, 404, "No hay cambios pendientes seleccionados.", nil)
		return
	}
	if accepted > 0 {
		st.State.ServerVersion++
		db := legacyMap(st.State.Database)
		db["updatedAt"] = now
		db["revision"] = numericValue(db["revision"]) + 1
		st.State.Database = legacyPrepareCanonicalDB(db, st.State, "Cambio administrativo aceptado por usuario")
		st.State.UpdatedAt = now
	}
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	payload := s.legacySyncSnapshotForRequest(st.State, r)
	payload["ok"] = true
	payload["decision"] = req.Decision
	payload["decidedIds"] = decided
	payload["acceptedCount"] = accepted
	jsonOut(w, 200, payload)
}
