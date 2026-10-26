package main

import (
	"encoding/base64"
	"html"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

type legacyReviewSubmitRequest struct {
	RestoreGeneration int64          `json:"restoreGeneration"`
	InstallationID    string         `json:"installationId"`
	ActorID           string         `json:"actorId"`
	ActorRole         string         `json:"actorRole"`
	OwnerID           string         `json:"ownerId"`
	OperationID       string         `json:"operationId"`
	RequirementID     string         `json:"requirementId"`
	RequirementTitle  string         `json:"requirementTitle"`
	ProgramID         string         `json:"programId"`
	ContentHash       string         `json:"contentHash"`
	AckOnly           bool           `json:"ackOnly,omitempty"`
	Submission        map[string]any `json:"submission"`
	Files             []any          `json:"files"`
}

func safeReviewSubmission(in map[string]any, now string) map[string]any {
	out := map[string]any{}
	for _, key := range []string{"reportWatermark", "reportPhotoBoxes", "reportLayoutPages", "reportFont", "reportFontSize", "text", "richText", "completionDate", "selections", "attachmentIds", "orientation", "submissionMode", "deliveryMode", "documentMode", "documentFormat", "primaryDocumentId", "documentPageCount", "documentSource", "documentOriginalName", "documentHash", "readingReport", "readingReportVersion", "bibleReadings", "bibleReadingVersion", "bibleVersionId", "bibleVersionName"} {
		if value, ok := in[key]; ok {
			out[key] = cloneAny(value)
		}
	}
	out["feedback"] = ""
	out["approval"] = nil
	if updatedAt := strings.TrimSpace(stringValue(in["updatedAt"])); updatedAt != "" {
		out["updatedAt"] = updatedAt
	} else {
		out["updatedAt"] = now
	}
	return out
}

func reviewProgressContainer(db map[string]any, ownerID, programID string) map[string]any {
	if db == nil || ownerID == "" {
		return nil
	}
	if programID == "" || programID == "class_grouped" {
		all := ensureMapField(db, "progress")
		p := legacyMap(all[ownerID])
		if p == nil {
			p = map[string]any{}
			all[ownerID] = p
		}
		ensureMapField(p, "status")
		ensureMapField(p, "submissions")
		return p
	}
	programs := ensureMapField(db, "programProgress")
	ownerPrograms := legacyMap(programs[ownerID])
	if ownerPrograms == nil {
		ownerPrograms = map[string]any{}
		programs[ownerID] = ownerPrograms
	}
	p := legacyMap(ownerPrograms[programID])
	if p == nil {
		p = map[string]any{}
		ownerPrograms[programID] = p
	}
	ensureMapField(p, "status")
	ensureMapField(p, "submissions")
	return p
}

func reviewFileIsPhoto(file map[string]any) bool {
	if file == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(stringValue(file["kind"])), "photo") {
		return true
	}
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(stringValue(file["type"]))), "image/")
}

func reviewFileIsRequirementDocument(file map[string]any) bool {
	if file == nil {
		return false
	}
	kind := strings.ToLower(strings.TrimSpace(stringValue(file["kind"])))
	typeName := strings.ToLower(strings.TrimSpace(stringValue(file["type"])))
	name := strings.ToLower(strings.TrimSpace(stringValue(file["name"])))
	return kind == "requirement_document" && (typeName == "application/pdf" || strings.HasSuffix(name, ".pdf"))
}

func reviewSubmissionIsDocumentMode(submission map[string]any) bool {
	if submission == nil {
		return false
	}
	mode := strings.ToLower(strings.TrimSpace(stringValue(submission["submissionMode"])))
	if mode == "" {
		mode = strings.ToLower(strings.TrimSpace(stringValue(submission["deliveryMode"])))
	}
	return mode == "document"
}

func reviewRequirementMode(db map[string]any, programID, requirementID string) string {
	if db == nil || strings.TrimSpace(requirementID) == "" {
		return ""
	}
	if strings.TrimSpace(programID) == "" {
		programID = "class_grouped"
	}
	modes := mapField(db, "reportRequirementModes")
	if len(modes) == 0 {
		settings := mapField(db, "settings")
		modes = mapField(settings, "reportRequirementModes")
	}
	value, ok := modes[programID+":"+requirementID]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case bool:
		if typed {
			return "document"
		}
	case string:
		mode := strings.ToLower(strings.TrimSpace(typed))
		if mode == "document" || mode == "reading_report" {
			return mode
		}
	default:
		entry := legacyMap(value)
		mode := strings.ToLower(strings.TrimSpace(stringValue(entry["mode"])))
		if mode == "document" || mode == "reading_report" {
			return mode
		}
	}
	return ""
}

func reviewRequirementDocumentMode(db map[string]any, programID, requirementID string) bool {
	return reviewRequirementMode(db, programID, requirementID) == "document"
}

func reviewRequirementReadingReportMode(db map[string]any, programID, requirementID string) bool {
	switch programID {
	case "class_grouped", "guide_major", "guide_master", "guide_master_advanced":
		return reviewRequirementMode(db, programID, requirementID) == "reading_report"
	default:
		return false
	}
}

func reviewSubmissionIsReadingReport(submission map[string]any) bool {
	if submission == nil {
		return false
	}
	mode := strings.ToLower(strings.TrimSpace(stringValue(submission["submissionMode"])))
	if mode == "" {
		mode = strings.ToLower(strings.TrimSpace(stringValue(submission["deliveryMode"])))
	}
	return mode == "reading_report"
}

func reviewSubmissionIsReadingReportComplete(submission map[string]any) (bool, string) {
	if submission == nil || !reviewSubmissionIsReadingReport(submission) {
		return false, "Este requisito requiere completar el Informe de lectura dentro de SGL."
	}
	report := legacyMap(submission["readingReport"])
	if report == nil {
		return false, "Completa el Informe de lectura antes de presentarlo."
	}
	elements := arrayField(report, "elements")
	count := 0
	for _, raw := range elements {
		if strings.TrimSpace(stringValue(raw)) != "" {
			count++
		}
	}
	if count < 3 {
		return false, "Escribe al menos 3 elementos clave o caracteristicas destacadas."
	}
	checks := []struct{ key, message string }{
		{"impactPhrase", "Completa la Frase de impacto."},
		{"impactPage", "Indica la pagina de la Frase de impacto."},
		{"personal", "Completa la aplicacion en Vida personal."},
		{"professional", "Completa la aplicacion en Vida profesional."},
		{"spiritual", "Completa la aplicacion en Vida espiritual."},
		{"finalReflection", "Completa la Reflexion final."},
	}
	for _, item := range checks {
		if strings.TrimSpace(stringValue(report[item.key])) == "" {
			return false, item.message
		}
	}
	if len(arrayField(submission, "attachmentIds")) > 0 {
		return false, "El Informe de lectura no admite archivos, PDF ni fotografias adjuntas."
	}
	return true, ""
}

func reviewSubmissionIsRequirementDocument(submission map[string]any, files []any) (bool, string) {
	if submission == nil || !reviewSubmissionIsDocumentMode(submission) {
		return false, "Este requisito requiere un documento PDF guardado desde SGL."
	}
	pageCount := int(numericValue(submission["documentPageCount"]))
	if pageCount < 1 || pageCount > 5 {
		return false, "El documento debe contener entre 1 y 5 paginas."
	}
	primary := strings.TrimSpace(stringValue(submission["primaryDocumentId"]))
	if primary == "" {
		return false, "Guarda el PDF o las fotografias antes de presentar el requisito."
	}
	attached := false
	for _, rawID := range arrayField(submission, "attachmentIds") {
		if strings.TrimSpace(stringValue(rawID)) == primary {
			attached = true
			break
		}
	}
	if !attached {
		return false, "El documento principal no esta vinculado a la entrega."
	}
	for _, raw := range files {
		file := legacyMap(raw)
		if strings.TrimSpace(stringValue(file["id"])) != primary {
			continue
		}
		if !reviewFileIsRequirementDocument(file) {
			return false, "La entrega documental debe ser un PDF generado por SGL."
		}
		return true, ""
	}
	return false, "El PDF guardado no esta disponible. Vuelve a guardar el documento antes de presentarlo."
}

const (
	reviewMinNormalReportTextRows = 3
	reviewNormalReportRowChars    = 72
)

// reviewRichTextPlainWithLines keeps meaningful line boundaries while removing
// markup. Normal reports are composed in a contenteditable surface, so the
// stored plain text is preferred; this fallback keeps older richText-only
// submissions compatible without allowing images or empty blocks to count.
func reviewRichTextPlainWithLines(rich string) string {
	if strings.TrimSpace(rich) == "" {
		return ""
	}
	var plain strings.Builder
	var tag strings.Builder
	inTag := false
	for _, r := range rich {
		switch r {
		case '<':
			if !inTag {
				inTag = true
				tag.Reset()
			}
		case '>':
			if inTag {
				name := strings.ToLower(strings.TrimSpace(tag.String()))
				if strings.HasPrefix(name, "br") || strings.HasPrefix(name, "/p") || strings.HasPrefix(name, "/div") || strings.HasPrefix(name, "/li") || strings.HasPrefix(name, "/h") {
					plain.WriteByte('\n')
				}
				inTag = false
			}
		default:
			if inTag {
				tag.WriteRune(r)
			} else {
				plain.WriteRune(r)
			}
		}
	}
	return html.UnescapeString(plain.String())
}

func reviewSubmissionReportText(submission map[string]any) string {
	if submission == nil {
		return ""
	}
	text := strings.ReplaceAll(stringValue(submission["text"]), "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "\u00a0", " ")
	rich := reviewRichTextPlainWithLines(stringValue(submission["richText"]))
	if strings.TrimSpace(rich) != "" && strings.Join(strings.Fields(text), "") == strings.Join(strings.Fields(rich), "") {
		return rich
	}
	if strings.TrimSpace(text) != "" {
		return text
	}
	return rich
}

// reviewNormalReportTextRows provides one deterministic definition for the
// minimum on mobile, desktop and server. Explicit non-empty lines count as one
// row; long lines also count the rows they occupy at the canonical report width.
func reviewNormalReportTextRows(submission map[string]any) int {
	rows := 0
	for _, raw := range strings.Split(reviewSubmissionReportText(submission), "\n") {
		line := strings.Join(strings.Fields(raw), " ")
		if line == "" {
			continue
		}
		runes := utf8.RuneCountInString(line)
		rows += (runes + reviewNormalReportRowChars - 1) / reviewNormalReportRowChars
	}
	return rows
}

func reviewSubmissionHasReportText(submission map[string]any) bool {
	return strings.TrimSpace(reviewSubmissionReportText(submission)) != ""
}

func purgeRequirementDocumentUploads(st *legacyStore) {
	if st == nil {
		return
	}
	db := legacyMap(st.State.Database)
	if db == nil {
		return
	}
	blocked := map[string]bool{}
	protected := map[string]bool{}
	filesByID := map[string]map[string]any{}
	for _, raw := range st.State.Files {
		file := legacyMap(raw)
		if file == nil {
			continue
		}
		id := strings.TrimSpace(stringValue(file["id"]))
		if id != "" {
			filesByID[id] = file
		}
	}

	cleanProgress := func(progress map[string]any, programID string) {
		if progress == nil {
			return
		}
		statuses := mapField(progress, "status")
		submissions := mapField(progress, "submissions")
		for requirementID, raw := range submissions {
			submission := legacyMap(raw)
			if submission == nil {
				continue
			}
			preserveDocument := reviewSubmissionIsDocumentMode(submission) || reviewRequirementDocumentMode(db, programID, requirementID)
			preserveReading := reviewSubmissionIsReadingReport(submission) || reviewRequirementReadingReportMode(db, programID, requirementID)
			if preserveReading {
				if primary := strings.TrimSpace(stringValue(submission["primaryDocumentId"])); primary != "" {
					blocked[primary] = true
				}
				for _, key := range []string{"documentMode", "documentFormat", "primaryDocumentId", "documentPageCount", "documentSource", "documentOriginalName", "documentHash"} {
					delete(submission, key)
				}
				submission["submissionMode"] = "reading_report"
				submission["deliveryMode"] = "reading_report"
				submission["attachmentIds"] = []any{}
				submissions[requirementID] = submission
				continue
			}
			if preserveDocument {
				if primary := strings.TrimSpace(stringValue(submission["primaryDocumentId"])); primary != "" {
					protected[primary] = true
				}
				for _, rawID := range arrayField(submission, "attachmentIds") {
					id := strings.TrimSpace(stringValue(rawID))
					if id != "" && reviewFileIsRequirementDocument(filesByID[id]) {
						protected[id] = true
					}
				}
				continue
			}
			removed := false
			if primary := strings.TrimSpace(stringValue(submission["primaryDocumentId"])); primary != "" {
				blocked[primary] = true
				removed = true
			}
			for _, key := range []string{"submissionMode", "deliveryMode", "documentMode", "documentFormat", "primaryDocumentId", "documentPageCount", "documentSource", "documentOriginalName", "documentHash", "readingReport", "readingReportVersion"} {
				if _, ok := submission[key]; ok {
					delete(submission, key)
					removed = true
				}
			}
			ids := arrayField(submission, "attachmentIds")
			filtered := make([]any, 0, len(ids))
			for _, rawID := range ids {
				id := strings.TrimSpace(stringValue(rawID))
				file := filesByID[id]
				if id != "" && file != nil && !reviewFileIsPhoto(file) {
					blocked[id] = true
					removed = true
					continue
				}
				filtered = append(filtered, rawID)
			}
			submission["attachmentIds"] = filtered
			if removed && !reviewSubmissionHasReportText(submission) && len(filtered) == 0 {
				status := strings.TrimSpace(stringValue(statuses[requirementID]))
				if status == "in_review" || status == "approved" {
					statuses[requirementID] = "in_progress"
				}
				submission["approval"] = nil
				submission["feedback"] = ""
			}
			submissions[requirementID] = submission
		}
	}

	for _, raw := range mapField(db, "progress") {
		cleanProgress(legacyMap(raw), "class_grouped")
	}
	for _, rawOwner := range mapField(db, "programProgress") {
		ownerPrograms := legacyMap(rawOwner)
		for programID, rawProgress := range ownerPrograms {
			cleanProgress(legacyMap(rawProgress), programID)
		}
	}
	for id, file := range filesByID {
		if id == "" || protected[id] || reviewFileIsPhoto(file) || strings.TrimSpace(stringValue(file["requirementId"])) == "" {
			continue
		}
		programID := strings.TrimSpace(stringValue(file["programId"]))
		if programID == "" {
			programID = "class_grouped"
		}
		if reviewFileIsRequirementDocument(file) && reviewRequirementDocumentMode(db, programID, strings.TrimSpace(stringValue(file["requirementId"]))) {
			continue
		}
		blocked[id] = true
	}
	if len(blocked) > 0 {
		kept := make([]any, 0, len(st.State.Files))
		for _, raw := range st.State.Files {
			file := legacyMap(raw)
			id := strings.TrimSpace(stringValue(file["id"]))
			if id != "" && blocked[id] && !protected[id] {
				continue
			}
			kept = append(kept, raw)
		}
		st.State.Files = kept
	}
	st.State.Database = db
}

// reviewStoredSubmissionIsInternalReport validates an already-submitted review.
// Releases before 5.6 allowed shorter normal reports (and even image-only
// reports), so reviewers must still be able to approve or return those legacy
// submissions after upgrading. The 3-row minimum is enforced only when a new
// review is submitted.
func reviewStoredSubmissionIsInternalReport(submission map[string]any, files []any) (bool, string) {
	if reviewSubmissionIsReadingReport(submission) {
		return reviewSubmissionIsReadingReportComplete(submission)
	}
	if reviewSubmissionIsDocumentMode(submission) {
		return reviewSubmissionIsRequirementDocument(submission, files)
	}
	if submission == nil {
		return false, "La entrega no contiene un informe interno."
	}
	hasText := reviewSubmissionHasReportText(submission)
	attachmentIDs := arrayField(submission, "attachmentIds")
	if len(attachmentIDs) == 0 {
		if !hasText {
			return false, "Completa el informe dentro de SGL antes de enviarlo a revisión."
		}
		return true, ""
	}
	byID := map[string]map[string]any{}
	for _, raw := range files {
		file := legacyMap(raw)
		id := strings.TrimSpace(stringValue(file["id"]))
		if id != "" {
			byID[id] = file
		}
	}
	photos := 0
	for _, rawID := range attachmentIDs {
		id := strings.TrimSpace(stringValue(rawID))
		file := byID[id]
		if id == "" || file == nil {
			return false, "Una evidencia del informe no está disponible. Vuelve a guardar el informe antes de enviarlo."
		}
		if !reviewFileIsPhoto(file) {
			return false, "Los documentos adjuntos no son una forma válida de entrega. Usa el informe interno y, si hace falta, agrega fotos o imágenes."
		}
		photos++
	}
	if !hasText && photos == 0 {
		return false, "Completa el informe dentro de SGL antes de enviarlo a revisión."
	}
	return true, ""
}

func reviewSubmissionIsInternalReport(submission map[string]any, files []any) (bool, string) {
	if reviewSubmissionIsReadingReport(submission) {
		return reviewSubmissionIsReadingReportComplete(submission)
	}
	if reviewSubmissionIsDocumentMode(submission) {
		return reviewSubmissionIsRequirementDocument(submission, files)
	}
	if submission == nil {
		return false, "La entrega no contiene un informe interno."
	}
	if rows := reviewNormalReportTextRows(submission); rows < reviewMinNormalReportTextRows {
		return false, "Escribe al menos 3 lineas de texto en el informe antes de enviarlo a revision. Las imagenes no reemplazan el texto."
	}
	attachmentIDs := arrayField(submission, "attachmentIds")
	if len(attachmentIDs) == 0 {
		return true, ""
	}
	byID := map[string]map[string]any{}
	for _, raw := range files {
		file := legacyMap(raw)
		id := strings.TrimSpace(stringValue(file["id"]))
		if id != "" {
			byID[id] = file
		}
	}
	photos := 0
	for _, rawID := range attachmentIDs {
		id := strings.TrimSpace(stringValue(rawID))
		file := byID[id]
		if id == "" || file == nil {
			return false, "Una evidencia del informe no está disponible. Vuelve a guardar el informe antes de enviarlo."
		}
		if !reviewFileIsPhoto(file) {
			return false, "Los documentos adjuntos no son una forma válida de entrega. Usa el informe interno y, si hace falta, agrega fotos o imágenes."
		}
		photos++
	}
	if photos == 0 {
		return false, "Las imagenes adjuntas no estan disponibles para esta entrega."
	}
	return true, ""
}

func hasReviewDocumentUpload(files []any, actorID, ownerID, programID, requirementID string) bool {
	for _, raw := range files {
		file := legacyMap(raw)
		if file == nil || stringValue(file["ownerId"]) != ownerID || ownerID != actorID {
			continue
		}
		if rid := strings.TrimSpace(stringValue(file["requirementId"])); rid != "" && rid != requirementID {
			continue
		}
		if pid := strings.TrimSpace(stringValue(file["programId"])); pid != "" && programID != "" && pid != programID {
			continue
		}
		if !reviewFileIsPhoto(file) {
			return true
		}
	}
	return false
}

func filterReviewFiles(files []any, actorID, ownerID, programID, requirementID string, allowDocument bool) []any {
	out := make([]any, 0, len(files))
	for _, raw := range files {
		file := legacyMap(raw)
		if file == nil || stringValue(file["ownerId"]) != ownerID || ownerID != actorID {
			continue
		}
		if rid := strings.TrimSpace(stringValue(file["requirementId"])); rid != "" && rid != requirementID {
			continue
		}
		if pid := strings.TrimSpace(stringValue(file["programId"])); pid != "" && programID != "" && pid != programID {
			continue
		}
		if !reviewFileIsPhoto(file) && !(allowDocument && reviewFileIsRequirementDocument(file)) {
			continue
		}
		out = append(out, cloneAny(file))
	}
	return out
}

// submitReview applies only one requirement instead of uploading the entire local
// database. This keeps review delivery fast and reliable over Funnel/mobile links.
func (s *Server) submitReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := s.auth(r); !ok {
		legacyError(w, 401, "Token incorrecto.", nil)
		return
	}
	var req legacyReviewSubmitRequest
	if err := readJSONLimited(r, &req, 96<<20); err != nil {
		legacyError(w, 400, "No se pudo leer el requisito enviado.", nil)
		return
	}
	if sess, remote := s.remoteSessionFromRequest(r); remote {
		if req.ActorID != "" && req.ActorID != sess.UserID {
			legacyError(w, 403, "La sesión remota no corresponde al usuario enviado.", nil)
			return
		}
		req.ActorID = sess.UserID
		req.ActorRole = sess.Role
	}
	req.ActorID = strings.TrimSpace(req.ActorID)
	req.OwnerID = strings.TrimSpace(req.OwnerID)
	if req.OwnerID == "" {
		req.OwnerID = req.ActorID
	}
	req.OperationID = strings.TrimSpace(req.OperationID)
	req.RequirementID = strings.TrimSpace(req.RequirementID)
	req.ProgramID = strings.TrimSpace(req.ProgramID)
	if req.ProgramID == "" {
		req.ProgramID = "class_grouped"
	}
	if req.ActorID == "" || req.OwnerID == "" || req.RequirementID == "" || req.OperationID == "" || req.Submission == nil {
		legacyError(w, 400, "Faltan datos para enviar el requisito a revisión.", nil)
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
		legacyError(w, 409, "El servidor todavía no fue inicializado por un Administrador.", nil)
		return
	}
	currentGeneration := legacyCanonicalRestoreGeneration(st.State)
	if req.RestoreGeneration != currentGeneration && !(req.RestoreGeneration == 0 && currentGeneration == 1) {
		legacyError(w, 409, "El servidor cambió de copia. Conserva tus avances y completa la recuperación antes de presentarlos.", map[string]any{"code": "RESTORE_REQUIRED", "restoreGeneration": currentGeneration, "restoreId": st.State.RestoreID, "restoredAt": st.State.RestoredAt, "serverVersion": st.State.ServerVersion})
		return
	}
	actor := legacyUser(st.State.Database, req.ActorID)
	if req.ProgramID == musicProgram553 && !musicPermit602(legacyMap(st.State.Database), actor, "reports.submit") {
		legacyError(w, 403, "Envío de informes desactivado.", nil)
		return
	}
	if !legacyActive(actor) || requestRole555(r, actor) != "aspirant" {
		legacyError(w, 403, "Solo el Aspirante activo puede enviar su requisito a revisión.", nil)
		return
	}
	if !assignedProgram555(legacyMap(st.State.Database), req.ActorID, req.ProgramID) && !(len(arrayField(legacyMap(st.State.Database), "modules")) == 0 && req.ProgramID == "class_grouped") {
		legacyError(w, 403, "Esta tarjeta no está asignada a tu cuenta.", nil)
		return
	}
	if req.OwnerID != req.ActorID {
		legacyError(w, 403, "El Aspirante solo puede enviar sus propios requisitos.", nil)
		return
	}
	if previous := findAppliedLegacyOperation(&st, req.ActorID, req.OperationID); previous != nil {
		if previous.OperationType != "review" || previous.ProgramID != req.ProgramID || previous.RequirementID != req.RequirementID || previous.ContentHash != req.ContentHash {
			legacyError(w, 409, "El identificador de envío pertenece a otro contenido. Se conserva el pendiente; no se confirma su recepción.", map[string]any{"code": "OPERATION_CONFLICT", "clearLocalOperation": false})
			return
		}
		if req.AckOnly {
			payload := legacyAutomaticAck(st.State, *previous, true)
			currentStatus := legacyRequirementStatus(st.State.Database, req.OwnerID, req.ProgramID, req.RequirementID)
			if currentStatus == "" && previous.OperationType == "review" {
				currentStatus = "in_review"
			}
			payload["currentStatus"] = currentStatus
			progress := legacyProgressForRequirement(st.State.Database, req.OwnerID, req.ProgramID, false)
			sub := legacyMap(mapField(progress, "submissions")[req.RequirementID])
			payload["filesReceived"] = previous.ContentHash == req.ContentHash && reviewFilesComplete(sub, st.State.Files, req.OwnerID)
			jsonOut(w, 200, payload)
			return
		}
		payload := s.legacySyncSnapshotForRequest(st.State, r)
		payload["status"] = "applied"
		payload["operationId"] = previous.OperationID
		payload["operationType"] = "review"
		payload["channel"] = "reviews"
		payload["clearLocalOperation"] = true
		payload["requirementId"] = req.RequirementID
		payload["programId"] = req.ProgramID
		payload["deduplicated"] = true
		jsonOut(w, 200, payload)
		return
	}

	st.State.Database = cloneAny(st.State.Database)
	st.Notifications = append([]legacyNotification(nil), st.Notifications...)
	st.AppliedOperations = append([]legacyAppliedOperation(nil), st.AppliedOperations...)
	db := legacyMap(st.State.Database)
	progress := reviewProgressContainer(db, req.OwnerID, req.ProgramID)
	if progress == nil {
		legacyError(w, 422, "No se encontró el expediente del Aspirante.", nil)
		return
	}
	status := ensureMapField(progress, "status")
	current := strings.TrimSpace(stringValue(status[req.RequirementID]))
	if current == "in_review" || current == "approved" {
		// El envio de revision es unico por Aspirante + modulo + requisito.
		// Si el celular o la laptop intenta enviar nuevamente, se confirma el
		// estado canonico del servidor y se ordena limpiar la operacion local
		// para no generar otra solicitud ni otra notificacion al Tutor/Admin.
		payload := s.legacySyncSnapshotForRequest(st.State, r)
		payload["ok"] = true
		payload["status"] = "applied"
		payload["currentStatus"] = current
		payload["operationId"] = req.OperationID
		payload["operationType"] = "review"
		payload["channel"] = "reviews"
		payload["clearLocalOperation"] = true
		payload["requirementId"] = req.RequirementID
		payload["programId"] = req.ProgramID
		payload["deduplicated"] = true
		payload["alreadyInReview"] = current == "in_review"
		payload["alreadyApproved"] = current == "approved"
		jsonOut(w, 200, payload)
		return
	}
	submissions := ensureMapField(progress, "submissions")
	existingSubmission := legacyMap(submissions[req.RequirementID])
	if req.ProgramID == musicProgram553 && !musicLinkedSubmissionGuard610(db, req.OwnerID, req.RequirementID, req.Submission) {
		legacyError(w, 422, "Utiliza el certificado oficial del ministerio vinculado para enviar este requisito.", nil)
		return
	}
	annualMusic := req.ProgramID == "music_ministry" && req.RequirementID == "LM-NUT-1"
	if annualMusic && !musicAnnualReviewReady551(db, req.OwnerID, req.Submission) {
		legacyError(w, 422, "Valida Concluido en Reavivados/Bible Plan o guarda un certificado PDF antes de enviar a revisión.", nil)
		return
	}
	configuredMode := reviewRequirementMode(db, req.ProgramID, req.RequirementID)
	documentMode := configuredMode == "document" || (configuredMode == "" && reviewSubmissionIsDocumentMode(existingSubmission))
	readingMode := configuredMode == "reading_report" || (configuredMode == "" && reviewSubmissionIsReadingReport(existingSubmission))
	if annualMusic {
		documentMode = reviewSubmissionIsDocumentMode(req.Submission)
		readingMode = false
	}
	if readingMode && !reviewSubmissionIsReadingReport(req.Submission) {
		legacyError(w, 422, "Este requisito se entrega unicamente mediante Informe de lectura. Completa el formulario dentro de SGL antes de presentarlo.", nil)
		return
	}
	if documentMode && !reviewSubmissionIsDocumentMode(req.Submission) {
		legacyError(w, 422, "Este requisito se entrega unicamente por PDF o fotografias. Guarda el documento desde SGL antes de presentarlo.", nil)
		return
	}
	if !documentMode && !readingMode {
		if reviewSubmissionIsReadingReport(req.Submission) {
			legacyError(w, 422, "Este requisito no esta configurado para Informe de lectura.", nil)
			return
		}
		if reviewSubmissionIsDocumentMode(req.Submission) || hasReviewDocumentUpload(req.Files, req.ActorID, req.OwnerID, req.ProgramID, req.RequirementID) {
			legacyError(w, 422, "Los documentos Word, PDF u otros archivos no se aceptan en un requisito de informe. Usa el informe interno y fotos o imagenes.", nil)
			return
		}
	}
	if readingMode && (len(req.Files) > 0 || len(arrayField(req.Submission, "attachmentIds")) > 0) {
		legacyError(w, 422, "El Informe de lectura no admite PDF, Word ni fotografias adjuntas.", nil)
		return
	}
	safeFiles := filterReviewFiles(req.Files, req.ActorID, req.OwnerID, req.ProgramID, req.RequirementID, documentMode)
	combinedFiles := make([]any, 0, len(st.State.Files)+len(safeFiles))
	combinedFiles = append(combinedFiles, st.State.Files...)
	combinedFiles = append(combinedFiles, safeFiles...)
	if annualMusic && documentMode && !musicAnnualPDF558(req.Submission, combinedFiles, req.OwnerID) {
		legacyError(w, 422, "Guarda un certificado PDF completo en este requisito antes de enviarlo a revisión.", nil)
		return
	}
	if !reviewFilesComplete(req.Submission, combinedFiles, req.OwnerID) {
		legacyError(w, 422, "Faltan archivos completos del informe. El envío sigue pendiente; vuelve a intentar la sincronización.", nil)
		return
	}
	if ok, reason := reviewSubmissionIsInternalReport(req.Submission, combinedFiles); !ok {
		legacyError(w, 422, reason, nil)
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	status[req.RequirementID] = "in_review"
	safeSubmission := safeReviewSubmission(req.Submission, now)
	safeSubmission["submittedAt"] = now
	if annualMusic && documentMode && len(legacyMap(existingSubmission["bibleReadings"])) > 0 {
		// Presenting a certificate does not edit the saved reading diary. Keep
		// the canonical records even when an older device omits them.
		safeSubmission["bibleReadings"] = cloneAny(existingSubmission["bibleReadings"])
	}
	if !documentMode && !readingMode {
		safeSubmission["normalReportTextRows"] = reviewNormalReportTextRows(req.Submission)
		safeSubmission["normalReportMinimumVersion"] = "5.6"
	}
	submissions[req.RequirementID] = safeSubmission
	progress["updatedAt"] = now

	for _, raw := range arrayField(db, "tasks") {
		task := legacyMap(raw)
		if stringValue(task["assigneeId"]) != req.OwnerID || stringValue(task["requirementId"]) != req.RequirementID {
			continue
		}
		pid := strings.TrimSpace(stringValue(task["programId"]))
		if pid == "" {
			pid = "class_grouped"
		}
		if pid != req.ProgramID || task["active"] == false || strings.TrimSpace(stringValue(task["deletedAt"])) != "" {
			continue
		}
		if strings.TrimSpace(stringValue(task["presentedAt"])) == "" {
			task["presentedAt"] = now
		}
		task["presentedBy"] = req.ActorID
		task["reviewSubmittedAt"] = now
		task["updatedAt"] = now
	}

	st.State.Files = mergeLegacyProposalFiles(st.State.Files, safeFiles, req.ActorID, map[string]bool{req.ActorID: true})
	if st.State.Files == nil {
		st.State.Files = []any{}
	}
	st.State.ServerVersion++
	db["revision"] = numericValue(db["revision"]) + 1
	db["updatedAt"] = now
	st.State.Database = legacyPrepareCanonicalDB(db, st.State, "Requisito enviado a revisión")
	st.State.UpdatedAt = now

	syncReq := legacySyncRequest{
		InstallationID: req.InstallationID, ActorID: req.ActorID, ActorRole: "aspirant", OwnerID: req.OwnerID,
		OperationID: req.OperationID, OperationType: "review", RequirementID: req.RequirementID,
		RequirementTitle: req.RequirementTitle, ProgramID: req.ProgramID, ContentHash: req.ContentHash,
	}
	s.addAutomaticSyncNotifications(&st, syncReq, actor, "aspirant", now)
	rememberAppliedLegacyOperation(&st, legacyAppliedOperation{
		OperationID: req.OperationID, ActorID: req.ActorID, OperationType: "review", RequirementID: req.RequirementID,
		ProgramID: req.ProgramID, ContentHash: req.ContentHash, AppliedAt: now, ServerVersion: st.State.ServerVersion,
	})
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	if req.AckOnly {
		applied := legacyAppliedOperation{OperationID: req.OperationID, ActorID: req.ActorID, OperationType: "review", RequirementID: req.RequirementID, ProgramID: req.ProgramID, ContentHash: req.ContentHash, AppliedAt: now, ServerVersion: st.State.ServerVersion}
		payload := legacyAutomaticAck(st.State, applied, false)
		payload["currentStatus"] = "in_review"
		payload["filesReceived"] = true
		jsonOut(w, 200, payload)
		return
	}
	payload := s.legacySyncSnapshotForRequest(st.State, r)
	payload["status"] = "applied"
	payload["operationId"] = req.OperationID
	payload["operationType"] = "review"
	payload["channel"] = "reviews"
	payload["clearLocalOperation"] = true
	payload["requirementId"] = req.RequirementID
	payload["programId"] = req.ProgramID
	payload["appliedAt"] = now
	payload["automatic"] = true
	payload["delta"] = true
	jsonOut(w, 200, payload)
}

// Metadata alone never proves that an attachment arrived. Validate the decoded
// bytes and ownership before changing status or creating any notification.
func reviewFilesComplete(sub map[string]any, files []any, owner string) bool {
	if sub == nil {
		return false
	}
	byID := map[string]map[string]any{}
	for _, raw := range files {
		f := legacyMap(raw)
		byID[stringValue(f["id"])] = f
	}
	ids := append([]any(nil), arrayField(sub, "attachmentIds")...)
	if id := strings.TrimSpace(stringValue(sub["primaryDocumentId"])); id != "" {
		ids = append(ids, id)
	}
	for _, raw := range ids {
		f := byID[stringValue(raw)]
		if f == nil || stringValue(f["ownerId"]) != owner {
			return false
		}
		value := stringValue(f["dataUrl"])
		comma := strings.Index(value, ",")
		if comma < 0 || !strings.HasPrefix(value, "data:") || !strings.HasSuffix(value[:comma], ";base64") {
			return false
		}
		data, err := base64.StdEncoding.DecodeString(value[comma+1:])
		if err != nil || len(data) == 0 {
			return false
		}
		if size := numericValue(f["size"]); size > 0 && size != float64(len(data)) {
			return false
		}
	}
	return true
}
