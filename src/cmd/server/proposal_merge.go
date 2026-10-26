package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

var errLegacyProposalNotApplicable = errors.New("legacy proposal cannot be applied")

func mapField(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	return legacyMap(m[key])
}

func ensureMapField(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	if current := legacyMap(m[key]); current != nil {
		m[key] = current
		return current
	}
	current := map[string]any{}
	m[key] = current
	return current
}

func arrayField(m map[string]any, key string) []any {
	if m == nil {
		return nil
	}
	if values, ok := m[key].([]any); ok {
		return values
	}
	return nil
}

func stringValue(v any) string {
	if v == nil {
		return ""
	}
	s := fmt.Sprint(v)
	if s == "<nil>" {
		return ""
	}
	return s
}

func numericValue(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case int32:
		return float64(n)
	default:
		var out float64
		_, _ = fmt.Sscan(stringValue(v), &out)
		return out
	}
}

func userExists(users []any, id string) bool {
	for _, raw := range users {
		if stringValue(legacyMap(raw)["id"]) == id {
			return true
		}
	}
	return false
}

func containsString(items []any, wanted string) bool {
	for _, item := range items {
		if stringValue(item) == wanted {
			return true
		}
	}
	return false
}

func mergeReviewState(targetProgress, sourceProgress map[string]any, now string) {
	if targetProgress == nil || sourceProgress == nil {
		return
	}
	targetStatus := ensureMapField(targetProgress, "status")
	sourceStatus := mapField(sourceProgress, "status")
	if sourceStatus == nil {
		return
	}
	targetSubmissions := ensureMapField(targetProgress, "submissions")
	sourceSubmissions := mapField(sourceProgress, "submissions")
	for reqID, rawStatus := range sourceStatus {
		nextStatus := stringValue(rawStatus)
		priorStatus := stringValue(targetStatus[reqID])
		if priorStatus != "in_review" || (nextStatus != "approved" && nextStatus != "needs_changes") {
			continue
		}
		targetStatus[reqID] = nextStatus
		targetSubmission := legacyMap(targetSubmissions[reqID])
		if targetSubmission == nil {
			continue
		}
		incoming := legacyMap(sourceSubmissions[reqID])
		targetSubmission["feedback"] = stringValue(incoming["feedback"])
		if nextStatus == "approved" {
			targetSubmission["approval"] = cloneAny(incoming["approval"])
		} else {
			targetSubmission["approval"] = nil
		}
		if updatedAt := stringValue(incoming["updatedAt"]); updatedAt != "" {
			targetSubmission["updatedAt"] = updatedAt
		} else {
			targetSubmission["updatedAt"] = now
		}
		targetSubmissions[reqID] = targetSubmission
	}
}

func mergeItemsByID(base, incoming []any, allow func(map[string]any) bool) []any {
	out := make([]any, 0, len(base)+len(incoming))
	index := map[string]int{}
	for _, raw := range base {
		item := legacyMap(cloneAny(raw))
		id := stringValue(item["id"])
		if id == "" {
			out = append(out, item)
			continue
		}
		index[id] = len(out)
		out = append(out, item)
	}
	for _, raw := range incoming {
		item := legacyMap(raw)
		if item == nil || !allow(item) {
			continue
		}
		id := stringValue(item["id"])
		if id == "" {
			continue
		}
		cloned := legacyMap(cloneAny(item))
		if at, ok := index[id]; ok {
			out[at] = cloned
		} else {
			index[id] = len(out)
			out = append(out, cloned)
		}
	}
	return out
}

// expandLegacyPatch starts from the CURRENT server database and overlays only the
// fields explicitly sent by a client operation. This is the key to safe multi-device
// use: a stale laptop cannot erase a newer phone change simply because its local
// snapshot is older. Arrays of entities are upserted by id; maps are merged deeply.
func mergeLegacyPatchValue(base, patch any) any {
	if patch == nil {
		return cloneAny(base)
	}
	if pm := legacyMap(patch); pm != nil {
		// Assignment edits carry only the memberships explicitly added/removed.
		// Applying them to the current list preserves changes from another device.
		if raw, ok := pm["__sglSetMembers"]; ok && len(pm) == 1 {
			change := legacyMap(raw)
			remove := map[string]bool{}
			for _, id := range arrayField(change, "remove") {
				remove[stringValue(id)] = true
			}
			result := []any{}
			seen := map[string]bool{}
			for _, id := range anySlice548(base) {
				key := stringValue(id)
				if key != "" && !remove[key] && !seen[key] {
					result = append(result, cloneAny(id))
					seen[key] = true
				}
			}
			for _, id := range arrayField(change, "add") {
				key := stringValue(id)
				if key != "" && !seen[key] {
					result = append(result, cloneAny(id))
					seen[key] = true
				}
			}
			return result
		}
		// Explicit replacement marker for the few operations that truly replace a
		// collection (for example deleting a class member). Ordinary patches remain
		// deep merges so a stale device cannot erase unrelated server changes.
		if replacement, ok := pm["__sglReplace"]; ok && len(pm) == 1 {
			return cloneAny(replacement)
		}
		bm := legacyMap(cloneAny(base))
		if bm == nil {
			bm = map[string]any{}
		}
		for key, value := range pm {
			if key == "__sglDeleteKeys" {
				if keys, ok := value.([]any); ok {
					for _, raw := range keys {
						delete(bm, stringValue(raw))
					}
				}
				continue
			}
			bm[key] = mergeLegacyPatchValue(bm[key], value)
		}
		return bm
	}
	if pa, ok := patch.([]any); ok {
		if len(pa) == 0 {
			return []any{}
		}
		allHaveID := true
		for _, raw := range pa {
			item := legacyMap(raw)
			if item == nil || stringValue(item["id"]) == "" {
				allHaveID = false
				break
			}
		}
		if allHaveID {
			baseArray, _ := base.([]any)
			out := make([]any, 0, len(baseArray)+len(pa))
			index := map[string]int{}
			for _, raw := range baseArray {
				item := cloneAny(raw)
				id := stringValue(legacyMap(item)["id"])
				if id != "" {
					index[id] = len(out)
				}
				out = append(out, item)
			}
			for _, raw := range pa {
				id := stringValue(legacyMap(raw)["id"])
				if at, ok := index[id]; ok {
					out[at] = mergeLegacyPatchValue(out[at], raw)
				} else {
					index[id] = len(out)
					out = append(out, mergeLegacyPatchValue(nil, raw))
				}
			}
			return out
		}
		return cloneAny(pa)
	}
	return cloneAny(patch)
}

func expandLegacyPatch(serverDB any, patch any) any {
	base := legacyMap(cloneAny(serverDB))
	if base == nil {
		base = map[string]any{}
	}
	pm := legacyMap(patch)
	if pm == nil {
		return base
	}
	for key, value := range pm {
		base[key] = mergeLegacyPatchValue(base[key], value)
	}
	return base
}

func driveReopenDevelopmentBarrier(progress map[string]any, requirementID string) bool {
	if progress == nil || strings.TrimSpace(requirementID) == "" {
		return false
	}
	if strings.TrimSpace(stringValue(mapField(progress, "status")[requirementID])) != "in_progress" {
		return false
	}
	submission := legacyMap(mapField(progress, "submissions")[requirementID])
	drive := legacyMap(submission["drive"])
	if drive == nil {
		return false
	}
	return strings.TrimSpace(stringValue(drive["status"])) == "deleted" &&
		strings.TrimSpace(stringValue(drive["returnedToStatus"])) == "in_progress" &&
		strings.TrimSpace(stringValue(drive["fileId"])) == "" &&
		strings.TrimSpace(stringValue(drive["deletedFileId"])) != ""
}

func mergeAspirantProgressState(target, source map[string]any) map[string]any {
	if target == nil && source == nil {
		return nil
	}
	result := legacyMap(mergeLegacyPatchValue(target, source))
	if result == nil {
		result = map[string]any{}
	}
	targetStatus := mapField(target, "status")
	resultStatus := ensureMapField(result, "status")
	targetSubs := mapField(target, "submissions")
	resultSubs := ensureMapField(result, "submissions")
	for reqID, raw := range targetStatus {
		priorStatus := strings.TrimSpace(stringValue(raw))
		nextStatus := strings.TrimSpace(stringValue(resultStatus[reqID]))

		// Generic/offline patches are never allowed to downgrade an already approved
		// server record. Approval transitions are handled by the dedicated review paths.
		if priorStatus == "approved" || priorStatus == "in_review" {
			resultStatus[reqID] = priorStatus
			if sub := targetSubs[reqID]; sub != nil {
				resultSubs[reqID] = cloneAny(sub)
			}
			continue
		}

		// SGL 5.27: once Drive has been explicitly removed from Informes, the
		// canonical in_progress state is a workflow barrier. A stale auto_sync/save
		// patch captured while the item was still approved must not resurrect the
		// approval (or its old Drive receipt). A later valid review uses the dedicated
		// review endpoint and is therefore unaffected by this guard.
		if priorStatus == "in_progress" && nextStatus != "in_progress" && driveReopenDevelopmentBarrier(target, reqID) {
			resultStatus[reqID] = "in_progress"
			if sub := targetSubs[reqID]; sub != nil {
				resultSubs[reqID] = cloneAny(sub)
			}
		}
	}
	// Review metadata is authoritative on the server even in editable drafts.
	for rid, raw := range resultSubs {
		sub := legacyMap(raw)
		if sub == nil {
			continue
		}
		prior := legacyMap(targetSubs[rid])
		for _, key := range []string{"approval", "feedback", "submittedAt"} {
			if v, ok := prior[key]; ok {
				sub[key] = cloneAny(v)
			} else {
				delete(sub, key)
			}
		}
	}
	return result
}

// mergeBibleReadingDelete removes one or more completed Bible reading records
// from the canonical Class Grouped progress. The operation is intentionally
// narrow: only the Aspirant can delete their own reading while the requirement
// is still editable. The client sends the exact reading IDs in a transient
// bibleReadingDeletedIds field; that marker is never persisted.
func mergeBibleReadingDelete(merged, source map[string]any, proposal *legacyProposal, actorRole string, allowed map[string]bool, now string) bool {
	if merged == nil || source == nil || proposal == nil || actorRole != "aspirant" {
		return false
	}
	ownerID := strings.TrimSpace(proposal.OwnerID)
	if ownerID == "" {
		ownerID = strings.TrimSpace(proposal.ActorID)
	}
	if ownerID == "" || ownerID != strings.TrimSpace(proposal.ActorID) || !allowed[ownerID] {
		return false
	}
	requirementID := strings.TrimSpace(proposal.RequirementID)
	if requirementID == "" {
		return false
	}
	programID := strings.TrimSpace(proposal.ProgramID)
	if programID == "" {
		programID = "class_grouped"
	}
	if programID != "class_grouped" {
		return false
	}

	targetProgress := legacyMap(ensureMapField(merged, "progress")[ownerID])
	sourceProgress := legacyMap(mapField(source, "progress")[ownerID])
	if targetProgress == nil || sourceProgress == nil {
		return false
	}
	targetStatus := ensureMapField(targetProgress, "status")
	priorStatus := strings.TrimSpace(stringValue(targetStatus[requirementID]))
	if priorStatus != "pending" && priorStatus != "in_progress" && priorStatus != "needs_changes" {
		return false
	}

	targetSubmissions := ensureMapField(targetProgress, "submissions")
	sourceSubmissions := mapField(sourceProgress, "submissions")
	targetSubmission := legacyMap(targetSubmissions[requirementID])
	incomingSubmission := legacyMap(sourceSubmissions[requirementID])
	if targetSubmission == nil || incomingSubmission == nil {
		return false
	}
	deletedRaw, _ := incomingSubmission["bibleReadingDeletedIds"].([]any)
	if len(deletedRaw) == 0 {
		return false
	}

	validIDs := make([]string, 0, len(deletedRaw))
	seen := map[string]bool{}
	for _, raw := range deletedRaw {
		id := strings.TrimSpace(stringValue(raw))
		if id == "" || seen[id] || !strings.HasPrefix(id, requirementID+"-L") {
			continue
		}
		seen[id] = true
		validIDs = append(validIDs, id)
	}
	if len(validIDs) == 0 {
		return false
	}

	targetBibleReadings := ensureMapField(targetSubmission, "bibleReadings")
	for _, id := range validIDs {
		delete(targetBibleReadings, id)
	}
	targetSubmission["bibleReadings"] = targetBibleReadings
	delete(targetSubmission, "bibleReadingDeletedIds")
	if updatedAt := strings.TrimSpace(stringValue(incomingSubmission["updatedAt"])); updatedAt != "" {
		targetSubmission["updatedAt"] = updatedAt
	} else {
		targetSubmission["updatedAt"] = now
	}
	targetSubmissions[requirementID] = targetSubmission

	// The local journal also keeps a mirror in progress.readings. Remove the
	// same IDs there so a later pull cannot resurrect a deleted chapter.
	targetReadings := ensureMapField(targetProgress, "readings")
	requirementReadings := legacyMap(targetReadings[requirementID])
	if requirementReadings == nil {
		requirementReadings = map[string]any{}
	}
	for _, id := range validIDs {
		delete(requirementReadings, id)
	}
	targetReadings[requirementID] = requirementReadings
	targetProgress["readings"] = targetReadings

	// Keep needs_changes if a reviewer already requested correction. Otherwise
	// derive pending/in_progress from the canonical records that remain after the
	// deletion, which prevents a stale device from hiding a newer chapter.
	if priorStatus == "needs_changes" {
		targetStatus[requirementID] = "needs_changes"
	} else if len(targetBibleReadings) > 0 || len(requirementReadings) > 0 {
		targetStatus[requirementID] = "in_progress"
	} else {
		targetStatus[requirementID] = "pending"
	}
	if updatedAt := strings.TrimSpace(stringValue(sourceProgress["updatedAt"])); updatedAt != "" {
		targetProgress["updatedAt"] = updatedAt
	} else {
		targetProgress["updatedAt"] = now
	}
	return true
}

// mergeDriveFinalization applies only the Drive receipt of an already approved
// requirement. The finalization operation is intentionally narrow: it never lets
// an Aspirant/Tutor rewrite the approved text, approval or requirement status.
func mergeDriveFinalization(merged, source map[string]any, proposal *legacyProposal, actorRole string, allowed map[string]bool, now string) bool {
	if merged == nil || source == nil || proposal == nil {
		return false
	}
	ownerID := strings.TrimSpace(proposal.OwnerID)
	if ownerID == "" {
		ownerID = strings.TrimSpace(proposal.ActorID)
	}
	if ownerID == "" || !allowed[ownerID] {
		return false
	}
	if actorRole == "aspirant" && ownerID != strings.TrimSpace(proposal.ActorID) {
		return false
	}
	requirementID := strings.TrimSpace(proposal.RequirementID)
	if requirementID == "" {
		return false
	}
	programID := strings.TrimSpace(proposal.ProgramID)
	if programID == "" {
		programID = "class_grouped"
	}

	var targetProgress, sourceProgress map[string]any
	if programID == "class_grouped" {
		targetProgress = legacyMap(ensureMapField(merged, "progress")[ownerID])
		sourceProgress = legacyMap(mapField(source, "progress")[ownerID])
	} else {
		targetPrograms := legacyMap(ensureMapField(merged, "programProgress")[ownerID])
		sourcePrograms := legacyMap(mapField(source, "programProgress")[ownerID])
		if targetPrograms == nil || sourcePrograms == nil {
			return false
		}
		targetProgress = legacyMap(targetPrograms[programID])
		sourceProgress = legacyMap(sourcePrograms[programID])
	}
	if targetProgress == nil || sourceProgress == nil {
		return false
	}
	if stringValue(mapField(targetProgress, "status")[requirementID]) != "approved" || stringValue(mapField(sourceProgress, "status")[requirementID]) != "approved" {
		return false
	}
	targetSubmissions := ensureMapField(targetProgress, "submissions")
	sourceSubmissions := mapField(sourceProgress, "submissions")
	targetSubmission := legacyMap(targetSubmissions[requirementID])
	incomingSubmission := legacyMap(sourceSubmissions[requirementID])
	if targetSubmission == nil || incomingSubmission == nil {
		return false
	}
	incomingDrive := legacyMap(incomingSubmission["drive"])
	if incomingDrive == nil {
		return false
	}
	driveStatus := strings.TrimSpace(stringValue(incomingDrive["status"]))
	if (driveStatus != "sent" && driveStatus != "verified") || strings.TrimSpace(stringValue(incomingDrive["fileId"])) == "" {
		return false
	}
	targetSubmission["drive"] = cloneAny(incomingDrive)
	if updatedAt := strings.TrimSpace(stringValue(incomingSubmission["updatedAt"])); updatedAt != "" {
		targetSubmission["updatedAt"] = updatedAt
	} else {
		targetSubmission["updatedAt"] = now
	}
	targetSubmissions[requirementID] = targetSubmission
	if updatedAt := strings.TrimSpace(stringValue(sourceProgress["updatedAt"])); updatedAt != "" {
		targetProgress["updatedAt"] = updatedAt
	} else {
		targetProgress["updatedAt"] = now
	}
	return true
}

// mergeDriveReopen is the exact inverse of finalization after the Drive file is
// removed, trashed or confirmed unavailable. It only reopens an already-finalized requirement:
// approved + Drive sent/verified -> in_progress + Drive deleted. It preserves the
// approved report body as the new editable starting point while clearing approval.
func mergeDriveReopen(merged, source map[string]any, proposal *legacyProposal, actorRole string, allowed map[string]bool, now string) bool {
	if merged == nil || source == nil || proposal == nil {
		return false
	}
	ownerID := strings.TrimSpace(proposal.OwnerID)
	if ownerID == "" {
		ownerID = strings.TrimSpace(proposal.ActorID)
	}
	if ownerID == "" || !allowed[ownerID] {
		return false
	}
	if actorRole == "aspirant" && ownerID != strings.TrimSpace(proposal.ActorID) {
		return false
	}
	requirementID := strings.TrimSpace(proposal.RequirementID)
	if requirementID == "" {
		return false
	}
	programID := strings.TrimSpace(proposal.ProgramID)
	if programID == "" {
		programID = "class_grouped"
	}

	if programID == "music_ministry" {
		return false
	}

	var targetProgress, sourceProgress map[string]any
	if programID == "class_grouped" {
		targetProgress = legacyMap(ensureMapField(merged, "progress")[ownerID])
		sourceProgress = legacyMap(mapField(source, "progress")[ownerID])
	} else {
		targetPrograms := legacyMap(ensureMapField(merged, "programProgress")[ownerID])
		sourcePrograms := legacyMap(mapField(source, "programProgress")[ownerID])
		if targetPrograms == nil || sourcePrograms == nil {
			return false
		}
		targetProgress = legacyMap(targetPrograms[programID])
		sourceProgress = legacyMap(sourcePrograms[programID])
	}
	if targetProgress == nil || sourceProgress == nil {
		return false
	}
	targetStatus := ensureMapField(targetProgress, "status")
	sourceStatus := mapField(sourceProgress, "status")
	if stringValue(targetStatus[requirementID]) != "approved" || stringValue(sourceStatus[requirementID]) != "in_progress" {
		return false
	}
	targetSubmissions := ensureMapField(targetProgress, "submissions")
	sourceSubmissions := mapField(sourceProgress, "submissions")
	targetSubmission := legacyMap(targetSubmissions[requirementID])
	incomingSubmission := legacyMap(sourceSubmissions[requirementID])
	if targetSubmission == nil || incomingSubmission == nil {
		return false
	}
	currentDrive := legacyMap(targetSubmission["drive"])
	incomingDrive := legacyMap(incomingSubmission["drive"])
	if currentDrive == nil || incomingDrive == nil {
		return false
	}
	currentDriveStatus := strings.TrimSpace(stringValue(currentDrive["status"]))
	if currentDriveStatus != "sent" && currentDriveStatus != "verified" {
		return false
	}
	if strings.TrimSpace(stringValue(currentDrive["fileId"])) == "" {
		return false
	}
	if strings.TrimSpace(stringValue(incomingDrive["status"])) != "deleted" || strings.TrimSpace(stringValue(incomingDrive["fileId"])) != "" {
		return false
	}

	targetStatus[requirementID] = "in_progress"
	targetSubmission["approval"] = nil
	targetSubmission["feedback"] = stringValue(incomingSubmission["feedback"])
	targetSubmission["drive"] = cloneAny(incomingDrive)
	if updatedAt := strings.TrimSpace(stringValue(incomingSubmission["updatedAt"])); updatedAt != "" {
		targetSubmission["updatedAt"] = updatedAt
	} else {
		targetSubmission["updatedAt"] = now
	}
	targetSubmissions[requirementID] = targetSubmission
	if updatedAt := strings.TrimSpace(stringValue(sourceProgress["updatedAt"])); updatedAt != "" {
		targetProgress["updatedAt"] = updatedAt
	} else {
		targetProgress["updatedAt"] = now
	}

	// Reopen also resets the task presentation/review markers for the same
	// requirement. Do this explicitly instead of relying on JSON null patching:
	// expandLegacyPatch intentionally treats null as "no replacement" for safety.
	targetTasks := arrayField(merged, "tasks")
	for _, raw := range targetTasks {
		item := legacyMap(raw)
		if item == nil || stringValue(item["assigneeId"]) != ownerID || stringValue(item["requirementId"]) != requirementID {
			continue
		}
		itemProgram := strings.TrimSpace(stringValue(item["programId"]))
		if itemProgram == "" {
			itemProgram = "class_grouped"
		}
		if itemProgram != programID {
			continue
		}
		item["presentedAt"] = nil
		item["presentedBy"] = nil
		item["reviewSubmittedAt"] = nil
		item["updatedAt"] = now
		item["updatedBy"] = strings.TrimSpace(proposal.ActorID)
	}
	if len(targetTasks) > 0 {
		merged["tasks"] = targetTasks
	}
	// The generic task merge runs later in this function. Normalize the expanded
	// source too so it cannot reintroduce the pre-reopen presentation markers.
	for _, raw := range arrayField(source, "tasks") {
		item := legacyMap(raw)
		if item == nil || stringValue(item["assigneeId"]) != ownerID || stringValue(item["requirementId"]) != requirementID {
			continue
		}
		itemProgram := strings.TrimSpace(stringValue(item["programId"]))
		if itemProgram == "" {
			itemProgram = "class_grouped"
		}
		if itemProgram != programID {
			continue
		}
		item["presentedAt"] = nil
		item["presentedBy"] = nil
		item["reviewSubmittedAt"] = nil
	}
	return true
}

// mergeLegacyProposalOnServer is the restricted role-aware merge shared by
// automatic 4.2.7 sync and by compact acceptance of historical proposals.
// Centralizing the merge keeps the browser request small and prevents an
// Aspirant/Tutor from overwriting fields outside the permissions already used
// by the legacy review flow.
func mergeLegacyProposalOnServer(serverDB any, proposal *legacyProposal) (any, map[string]bool, error) {
	if proposal == nil || proposal.Database == nil {
		return nil, nil, errLegacyProposalNotApplicable
	}
	merged := legacyMap(cloneAny(serverDB))
	source := legacyMap(proposal.Database)
	if merged == nil || source == nil {
		return nil, nil, errLegacyProposalNotApplicable
	}
	actor := legacyUser(merged, proposal.ActorID)
	if actor == nil || !legacyActive(actor) {
		return nil, nil, errLegacyProposalNotApplicable
	}
	actorID := stringValue(actor["id"])
	actorRole := legacyRole(actor)
	if actorID == "" || (actorRole != "aspirant" && actorRole != "tutor") {
		return nil, nil, errLegacyProposalNotApplicable
	}

	// Personal snapshots may save drafts, never grant approval or bypass delivery validation.
	if actorRole == "aspirant" {
		programs := map[string]any{"class_grouped": mapField(source, "progress")[actorID]}
		for pid, p := range legacyMap(mapField(source, "programProgress")[actorID]) {
			programs[pid] = p
		}
		for pid, p := range programs {
			prior := legacyProgressForRequirement(serverDB, actorID, pid, false)
			for rid, raw := range mapField(legacyMap(p), "status") {
				old := stringValue(mapField(prior, "status")[rid])
				next := stringValue(raw)
				if old == "approved" || old == "in_review" {
					continue
				}
				if next == "approved" || next == "in_review" {
					return nil, nil, errLegacyProposalNotApplicable
				}
			}
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	allowed := map[string]bool{actorID: true}
	newAspirants := map[string]bool{}
	users := legacyUsers(merged)
	sourceUsers := legacyUsers(source)

	if actorRole == "tutor" {
		for _, raw := range users {
			person := legacyMap(raw)
			if leader555(legacyMap(serverDB), person) && legacyTutorAssigned(person, actorID, "") {
				if id := stringValue(person["id"]); id != "" {
					allowed[id] = true
				}
			}
		}
		for _, raw := range sourceUsers {
			person := legacyMap(raw)
			id := stringValue(person["id"])
			if id == "" || legacyRole(person) != "aspirant" || stringValue(person["tutorId"]) != actorID || userExists(users, id) {
				continue
			}
			safe := legacyMap(cloneAny(person))
			safe["role"] = person["role"]
			safe["tutorId"] = actorID
			chosen := tutorCreationPrograms548(merged, source, actor, id)
			if len(chosen) == 0 {
				return nil, nil, fmt.Errorf("no authorized module for new aspirant")
			}
			assignments := map[string]any{}
			for _, programID := range chosen {
				assignments[stringValue(programID)] = actorID
			}
			safe["tutorByProgram"] = assignments
			delete(safe, "tutorProgramIds")
			for _, rawModule := range arrayField(merged, "modules") {
				module := legacyMap(rawModule)
				programID := stringValue(module["id"])
				if containsString(chosen, programID) {
					ids := arrayField(module, "assignedUserIds")
					if !containsString(ids, id) {
						module["assignedUserIds"] = append(ids, id)
					}
				}
			}
			delete(safe, "requirementViewPrograms")
			state := stringValue(safe["accountState"])
			safe["active"] = state != "inactive" && state != "archived"
			users = append(users, safe)
			allowed[id] = true
			newAspirants[id] = true
		}
		merged["users"] = users
	}

	sourceActor := legacyUser(source, actorID)
	if sourceActor != nil {
		targetActor := legacyUser(merged, actorID)
		if name := strings.ToLower(strings.TrimSpace(stringValue(sourceActor["username"]))); name != "" {
			for _, raw := range legacyUsers(merged) {
				u := legacyMap(raw)
				if stringValue(u["id"]) != actorID && (strings.EqualFold(strings.TrimSpace(stringValue(u["username"])), name) || strings.EqualFold(strings.TrimSpace(stringValue(u["email"])), name)) {
					return nil, nil, fmt.Errorf("username already in use")
				}
			}
		}
		for _, key := range []string{"photo", "password", "theme", "reportFont", "bibleVersionId", "bibleVersionPreference", "drive", "fullName", "username", "birthDate", "phone", "church", "clubPosition", "churchPosition", "missionaryDistrict", "region", "mission", "union"} {
			if value, exists := sourceActor[key]; exists {

				if fields, ok := sourceActor["profileChangedFields548"].([]any); !ok || containsString(fields, key) {
					targetActor[key] = cloneAny(value)
				}
			}
		}
	}

	mergedProgress := ensureMapField(merged, "progress")
	mergedPrograms := ensureMapField(merged, "programProgress")
	mergedBible := ensureMapField(merged, "bibleYear")
	mergedConfigs := ensureMapField(merged, "configs")
	sourceProgress := mapField(source, "progress")
	sourcePrograms := mapField(source, "programProgress")
	sourceBible := mapField(source, "bibleYear")

	if strings.TrimSpace(proposal.OperationType) == "drive_finalize" {
		if !mergeDriveFinalization(merged, source, proposal, actorRole, allowed, now) {
			return nil, nil, errLegacyProposalNotApplicable
		}
	} else if strings.TrimSpace(proposal.OperationType) == "drive_reopen" {
		if !mergeDriveReopen(merged, source, proposal, actorRole, allowed, now) {
			return nil, nil, errLegacyProposalNotApplicable
		}
	} else if strings.TrimSpace(proposal.OperationType) == "bible_reading_delete" {
		if !mergeBibleReadingDelete(merged, source, proposal, actorRole, allowed, now) {
			return nil, nil, errLegacyProposalNotApplicable
		}
	} else if actorRole == "aspirant" {
		if value := sourceProgress[actorID]; value != nil && (assignedProgram555(legacyMap(serverDB), actorID, "class_grouped") || len(arrayField(legacyMap(serverDB), "modules")) == 0) {
			mergedProgress[actorID] = mergeAspirantProgressState(legacyMap(mergedProgress[actorID]), legacyMap(value))
		}
		if value := sourcePrograms[actorID]; value != nil {
			targetPrograms := legacyMap(mergedPrograms[actorID])
			if targetPrograms == nil {
				targetPrograms = map[string]any{}
			}
			sourceProgramMap := legacyMap(value)
			for programID, rawProgress := range sourceProgramMap {
				if !assignedProgram555(legacyMap(serverDB), actorID, programID) {
					continue
				}
				prior := legacyMap(targetPrograms[programID])
				next := mergeAspirantProgressState(prior, legacyMap(rawProgress))
				if programID == "music_ministry" {
					nextSub := legacyMap(mapField(next, "submissions")["LM-NUT-1"])
					// 5.9.1: allow changing/clearing the reading choice in editable drafts.
					// mergeAspirantProgressState retains review locks and ownership.
					if stringValue(mapField(next, "status")["LM-NUT-1"]) == "in_review" && stringValue(mapField(prior, "status")["LM-NUT-1"]) != "in_review" && (reviewSubmissionIsDocumentMode(nextSub) || !musicAnnualReviewReady551(legacyMap(serverDB), actorID, nextSub)) {
						ensureMapField(next, "status")["LM-NUT-1"] = mapField(prior, "status")["LM-NUT-1"]
					}
				}
				targetPrograms[programID] = next
			}
			mergedPrograms[actorID] = targetPrograms
		}
		if value := sourceBible[actorID]; value != nil {
			mergedBible[actorID] = mergeLegacyPatchValue(mergedBible[actorID], value)
		}
	} else {
		for userID := range allowed {
			if userID == actorID {
				continue
			}
			if newAspirants[userID] {
				// A Tutor is allowed to create a new Aspirant assigned to that Tutor.
				// Bring the freshly initialized progress containers with the new account
				// so the first login does not depend on an Administrator round-trip.
				if value := sourceProgress[userID]; value != nil && legacyTutorAssigned(legacyUser(merged, userID), actorID, "class_grouped") {
					mergedProgress[userID] = cloneAny(value)
				}
				if value := sourcePrograms[userID]; value != nil {
					fresh := map[string]any{}
					for programID, progress := range legacyMap(value) {
						if legacyTutorAssigned(legacyUser(merged, userID), actorID, programID) {
							fresh[programID] = cloneAny(progress)
						}
					}
					mergedPrograms[userID] = fresh
				}
				if value := sourceBible[userID]; value != nil {
					mergedBible[userID] = cloneAny(value)
				}
				continue
			}
			if legacyTutorAssigned(legacyUser(merged, userID), actorID, "class_grouped") {
				mergeReviewState(legacyMap(mergedProgress[userID]), legacyMap(sourceProgress[userID]), now)
			}
			targetProgramMap := legacyMap(mergedPrograms[userID])
			incomingProgramMap := legacyMap(sourcePrograms[userID])
			for programID, rawProgress := range incomingProgramMap {
				if !legacyTutorAssigned(legacyUser(merged, userID), actorID, programID) {
					continue
				}
				mergeReviewState(legacyMap(targetProgramMap[programID]), legacyMap(rawProgress), now)
			}
		}
	}

	sourceConfigs := mapField(source, "configs")
	if actorRole == "tutor" {
		for key, config := range sourceConfigs {
			allowedConfig := false
			for userID := range allowed {
				if userID == actorID {
					continue
				}
				if key == userID || strings.Contains(key, ":"+userID+":") || strings.HasPrefix(key, userID+":") {
					allowedConfig = true
					break
				}
			}
			if allowedConfig {
				mergedConfigs[key] = cloneAny(config)
			}
		}
	}

	mergedTasks := arrayField(merged, "tasks")
	sourceTasks := arrayField(source, "tasks")
	if actorRole == "aspirant" {
		taskCollectionChanged := false
		// SGL 5.0.74: el Aspirante puede crear una tarea únicamente para sí mismo
		// cuando proviene del planificador personal. Una vez creada, posteriores
		// sincronizaciones del Aspirante no pueden cambiar sus fechas ni borrarla;
		// solo se conserva la actualización histórica de presentedAt.
		for _, rawIncoming := range sourceTasks {
			incoming := legacyMap(rawIncoming)
			if stringValue(incoming["assigneeId"]) != actorID {
				continue
			}
			incomingID := stringValue(incoming["id"])
			if incomingID == "" {
				continue
			}
			var target map[string]any
			for _, rawTarget := range mergedTasks {
				candidate := legacyMap(rawTarget)
				if stringValue(candidate["id"]) == incomingID && stringValue(candidate["assigneeId"]) == actorID {
					target = candidate
					break
				}
			}
			if target != nil {
				if stringValue(incoming["presentedAt"]) != "" {
					target["presentedAt"] = incoming["presentedAt"]
					if updatedAt := stringValue(incoming["updatedAt"]); updatedAt != "" {
						target["updatedAt"] = updatedAt
					} else {
						target["updatedAt"] = now
					}
				}
				continue
			}
			selfPlanned, _ := incoming["selfPlanned"].(bool)
			lockedForAssignee, _ := incoming["lockedForAssignee"].(bool)
			active, hasActive := incoming["active"].(bool)
			if stringValue(incoming["createdBy"]) != actorID || !selfPlanned || !lockedForAssignee || (hasActive && !active) || stringValue(incoming["deletedAt"]) != "" || stringValue(incoming["requirementId"]) == "" {
				continue
			}
			cloned := legacyMap(cloneAny(incoming))
			cloned["createdBy"] = actorID
			cloned["assigneeId"] = actorID
			cloned["createdByRole"] = "aspirant"
			cloned["selfPlanned"] = true
			cloned["lockedForAssignee"] = true
			cloned["active"] = true
			mergedTasks = append(mergedTasks, cloned)
			taskCollectionChanged = true
		}
		if taskCollectionChanged {
			merged["tasks"] = mergedTasks
		}
	} else {
		// SGL 5.0.74: una tarea presentada puede ser enviada a "En desarrollo"
		// por el Tutor asignado. La transición se valida contra la copia canónica
		// ANTERIOR del servidor para impedir que el Tutor fabrique presentedAt y
		// developmentAcceptedAt en una sola operación.
		validatedDevelopment := map[string]map[string]string{}
		merged["tasks"] = mergeItemsByID(mergedTasks, sourceTasks, func(item map[string]any) bool {
			assigneeID := stringValue(item["assigneeId"])
			baseAllowed := legacyTutorAssigned(legacyUser(merged, assigneeID), actorID, stringValue(item["programId"]))
			if !baseAllowed {
				return false
			}
			if stringValue(item["developmentAcceptedAt"]) == "" {
				return true
			}
			itemID := stringValue(item["id"])
			if actorRole != "tutor" || itemID == "" {
				return false
			}
			var prior map[string]any
			for _, rawPrior := range mergedTasks {
				candidate := legacyMap(rawPrior)
				if stringValue(candidate["id"]) == itemID {
					prior = candidate
					break
				}
			}
			if prior == nil || stringValue(prior["presentedAt"]) == "" {
				return false
			}
			programID := stringValue(prior["programId"])
			if programID == "" {
				programID = "class_grouped"
			}
			requirementID := stringValue(prior["requirementId"])
			if stringValue(prior["assigneeId"]) != assigneeID || stringValue(item["requirementId"]) != requirementID {
				return false
			}
			incomingProgramID := stringValue(item["programId"])
			if incomingProgramID == "" {
				incomingProgramID = "class_grouped"
			}
			if incomingProgramID != programID {
				return false
			}
			validatedDevelopment[itemID] = map[string]string{"assigneeId": assigneeID, "programId": programID, "requirementId": requirementID}
			return true
		})

		// Normaliza quién aceptó la tarea y aplica el cambio de estado. Para Clase
		// Agrupada el progreso vive en progress; para Guía Mayor/Máster vive en
		// programProgress. No se permite degradar una entrega ya en revisión/aprobada.
		if len(validatedDevelopment) > 0 {
			mergedTaskItems := arrayField(merged, "tasks")
			for taskID, meta := range validatedDevelopment {
				for _, rawTask := range mergedTaskItems {
					task := legacyMap(rawTask)
					if stringValue(task["id"]) != taskID {
						continue
					}
					task["developmentAcceptedBy"] = actorID
					task["developmentAcceptedByRole"] = "tutor"
					if stringValue(task["developmentAcceptedAt"]) == "" {
						task["developmentAcceptedAt"] = now
					}
					break
				}
				assigneeID, programID, requirementID := meta["assigneeId"], meta["programId"], meta["requirementId"]
				var progress map[string]any
				if programID == "class_grouped" {
					progress = legacyMap(mergedProgress[assigneeID])
					if progress == nil {
						progress = map[string]any{}
						mergedProgress[assigneeID] = progress
					}
				} else {
					programs := legacyMap(mergedPrograms[assigneeID])
					if programs == nil {
						programs = map[string]any{}
						mergedPrograms[assigneeID] = programs
					}
					progress = legacyMap(programs[programID])
					if progress == nil {
						progress = map[string]any{}
						programs[programID] = progress
					}
				}
				status := ensureMapField(progress, "status")
				priorStatus := stringValue(status[requirementID])
				if priorStatus != "approved" && priorStatus != "in_review" {
					status[requirementID] = "in_progress"
					progress["updatedAt"] = now
				}
			}
		}
		merged["materials"] = mergeItemsByID(arrayField(merged, "materials"), arrayField(source, "materials"), func(item map[string]any) bool { return stringValue(item["createdBy"]) == actorID })
		merged["indicationRules"] = mergeItemsByID(arrayField(merged, "indicationRules"), arrayField(source, "indicationRules"), func(item map[string]any) bool { return stringValue(item["createdBy"]) == actorID })
	}

	if actorRole == "aspirant" {
		assignedClasses := map[string]bool{}
		mergedClasses := arrayField(merged, "classes")
		for _, raw := range mergedClasses {
			class := legacyMap(raw)
			if containsString(arrayField(class, "assignedUserIds"), actorID) {
				assignedClasses[stringValue(class["id"])] = true
			}
		}
		for _, rawSourceClass := range arrayField(source, "classes") {
			sourceClass := legacyMap(rawSourceClass)
			classID := stringValue(sourceClass["id"])
			if !assignedClasses[classID] {
				continue
			}
			for _, rawTargetClass := range mergedClasses {
				targetClass := legacyMap(rawTargetClass)
				if stringValue(targetClass["id"]) == classID && sourceClass["work"] != nil {
					targetClass["work"] = cloneAny(sourceClass["work"])
					break
				}
			}
		}
	}

	audit := arrayField(merged, "audit")
	knownAudit := map[string]bool{}
	for _, raw := range audit {
		if id := stringValue(legacyMap(raw)["id"]); id != "" {
			knownAudit[id] = true
		}
	}
	for _, raw := range arrayField(source, "audit") {
		item := legacyMap(raw)
		id := stringValue(item["id"])
		if id == "" || knownAudit[id] || stringValue(item["actorId"]) != actorID {
			continue
		}
		audit = append(audit, cloneAny(item))
		knownAudit[id] = true
	}
	sort.SliceStable(audit, func(i, j int) bool {
		return stringValue(legacyMap(audit[i])["at"]) > stringValue(legacyMap(audit[j])["at"])
	})
	if len(audit) > 1000 {
		audit = audit[:1000]
	}
	merged["audit"] = audit
	merged["updatedAt"] = now
	revision := numericValue(merged["revision"])
	if sourceRevision := numericValue(source["revision"]); sourceRevision > revision {
		revision = sourceRevision
	}
	merged["revision"] = revision + 1
	return merged, allowed, nil
}

func mergeLegacyProposalFiles(serverFiles, proposalFiles []any, actorID string, allowed map[string]bool) []any {
	out := make([]any, 0, len(serverFiles)+len(proposalFiles))
	index := map[string]int{}
	for _, raw := range serverFiles {
		file := legacyMap(cloneAny(raw))
		id := stringValue(file["id"])
		if id != "" {
			index[id] = len(out)
		}
		out = append(out, file)
	}
	for _, raw := range proposalFiles {
		file := legacyMap(raw)
		id := stringValue(file["id"])
		ownerID := stringValue(file["ownerId"])
		if id == "" || (ownerID != actorID && !allowed[ownerID]) {
			continue
		}
		cloned := legacyMap(cloneAny(file))
		if at, ok := index[id]; ok {
			out[at] = cloned
		} else {
			index[id] = len(out)
			out = append(out, cloned)
		}
	}
	return out
}

// mergeLegacyFilesUnrestricted is used only for authenticated Administrator
// delta operations. Files are upserted by id without replacing the complete
// server file collection, so an Administrator upload cannot erase files added
// by a phone/laptop between synchronizations.
func mergeLegacyFilesUnrestricted(serverFiles, incomingFiles []any) []any {
	out := make([]any, 0, len(serverFiles)+len(incomingFiles))
	index := map[string]int{}
	for _, raw := range serverFiles {
		file := legacyMap(cloneAny(raw))
		id := stringValue(file["id"])
		if id != "" {
			index[id] = len(out)
		}
		out = append(out, file)
	}
	for _, raw := range incomingFiles {
		file := legacyMap(raw)
		id := stringValue(file["id"])
		if id == "" {
			continue
		}
		cloned := legacyMap(cloneAny(file))
		if at, ok := index[id]; ok {
			out[at] = cloned
		} else {
			index[id] = len(out)
			out = append(out, cloned)
		}
	}
	return out
}
