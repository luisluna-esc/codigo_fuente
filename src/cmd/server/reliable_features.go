package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// proposalDedupeKey is stable across retries. New 4.2.5+ clients send an
// operationId. Older clients are repaired by a normalized content fingerprint
// so repeated retries remain one request even if volatile revision/sync fields
// changed while the same saved change was being retransmitted.
func proposalDedupeKey(p legacyProposal) string {
	if strings.TrimSpace(p.OperationID) != "" {
		return "op:" + strings.TrimSpace(p.OperationID)
	}
	if fp := legacyProposalContentFingerprint(p); fp != "" {
		return "content:" + p.ActorID + ":" + p.InstallationID + ":" + fp
	}
	if p.ActorID != "" && p.InstallationID != "" && p.Revision != 0 {
		return fmt.Sprintf("legacy:%s:%s:%d", p.ActorID, p.InstallationID, p.Revision)
	}
	return "id:" + p.ID
}

// stripProposalVolatileFields removes values that change during local autosave
// bookkeeping but do not represent a new explicit user submission. This is only
// used to repair historical duplicate proposals; new clients use operationId.
func stripProposalVolatileFields(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, value := range x {
			switch k {
			case "revision", "serverVersion", "updatedAt", "sync", "audit", "errors", "conflicts", "syncQueue", "syncHistory", "messageCache", "messageOutbox":
				continue
			}
			out[k] = stripProposalVolatileFields(value)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, value := range x {
			out[i] = stripProposalVolatileFields(value)
		}
		return out
	default:
		return x
	}
}

func legacyProposalContentFingerprint(p legacyProposal) string {
	if p.Database == nil || strings.TrimSpace(p.ActorID) == "" {
		return ""
	}
	normalized := stripProposalVolatileFields(cloneAny(p.Database))
	b, err := json.Marshal(normalized)
	if err != nil || len(b) == 0 {
		return ""
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:12])
}

func generatedLegacyOperationID(actorID, installationID, operationType string, database any) string {
	fingerprint := legacyProposalContentFingerprint(legacyProposal{ActorID: actorID, InstallationID: installationID, OperationType: operationType, Database: database})
	if fingerprint == "" {
		b, _ := json.Marshal(stripProposalVolatileFields(cloneAny(database)))
		h := sha256.Sum256(b)
		fingerprint = hex.EncodeToString(h[:12])
	}
	kind := "AUT"
	switch strings.TrimSpace(operationType) {
	case "save_progress":
		kind = "AVA"
	case "review":
		kind = "REV"
	case "cancel_review":
		kind = "CAN"
	case "review_decision":
		kind = "RSP"
	}
	actor := strings.ToUpper(strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, actorID))
	if actor == "" {
		actor = "USR"
	}
	if len(actor) > 8 {
		actor = actor[len(actor)-8:]
	}
	install := strings.ToUpper(strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, installationID))
	if install == "" {
		install = "DEVICE"
	}
	if len(install) > 6 {
		install = install[len(install)-6:]
	}
	return fmt.Sprintf("SGL-%s-%s-%s-%s", kind, actor, install, strings.ToUpper(fingerprint[:12]))
}

// dedupeLegacyProposals repairs old duplicate pending requests in-place.
// It keeps the newest request as canonical and marks the rest superseded.
func dedupeLegacyProposals(st *legacyStore) (changed bool, superseded int) {
	if st == nil || len(st.Proposals) < 2 {
		return false, 0
	}
	groups := map[string][]int{}
	for i := range st.Proposals {
		p := st.Proposals[i]
		if p.Status != "pending" && p.Status != "postponed" {
			continue
		}
		key := proposalDedupeKey(p)
		groups[key] = append(groups[key], i)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, idxs := range groups {
		if len(idxs) < 2 {
			continue
		}
		sort.SliceStable(idxs, func(i, j int) bool {
			a, b := st.Proposals[idxs[i]], st.Proposals[idxs[j]]
			at, bt := a.UpdatedAt, b.UpdatedAt
			if at == "" {
				at = a.SubmittedAt
			}
			if bt == "" {
				bt = b.SubmittedAt
			}
			return at > bt
		})
		canonical := &st.Proposals[idxs[0]]
		for _, idx := range idxs[1:] {
			p := &st.Proposals[idx]
			p.Status = "superseded"
			p.CanonicalID = canonical.ID
			p.DecisionNote = "Solicitud duplicada consolidada automáticamente en " + canonical.ID
			p.UpdatedAt = now
			p.DecidedAt = now
			p.Database = nil
			p.Files = []any{}
			changed = true
			superseded++
		}
	}
	return changed, superseded
}

func findCanonicalProposal(st *legacyStore, id string) (int, *legacyProposal) {
	idx, p := findLegacyProposal(st, id)
	if p == nil {
		return idx, nil
	}
	seen := map[string]bool{}
	for p != nil && p.Status == "superseded" && p.CanonicalID != "" && !seen[p.ID] {
		seen[p.ID] = true
		idx, p = findLegacyProposal(st, p.CanonicalID)
	}
	return idx, p
}

type legacyCancelRequest struct {
	ActorID        string `json:"actorId"`
	ProposalID     string `json:"proposalId"`
	InstallationID string `json:"installationId"`
	Note           string `json:"note"`
}

func (s *Server) legacyProposalCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := s.auth(r); !ok {
		legacyError(w, http.StatusUnauthorized, "Token incorrecto.", nil)
		return
	}
	var req legacyCancelRequest
	if err := readJSONLimited(r, &req, 1<<20); err != nil {
		legacyError(w, http.StatusBadRequest, "Solicitud de cancelación inválida.", nil)
		return
	}
	if sess, remote := s.remoteSessionFromRequest(r); remote {
		req.ActorID = sess.UserID
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	if changed, _ := dedupeLegacyProposals(&st); changed {
		_ = s.saveLegacy(st)
	}
	_, p := findCanonicalProposal(&st, req.ProposalID)
	if p == nil || p.ActorID != req.ActorID {
		legacyError(w, 404, "Solicitud no encontrada para este usuario.", nil)
		return
	}
	if req.InstallationID != "" && p.InstallationID != req.InstallationID {
		legacyError(w, 403, "La solicitud pertenece a otro dispositivo.", nil)
		return
	}
	switch p.Status {
	case "cancelled":
		jsonOut(w, 200, map[string]any{"ok": true, "status": p.Status, "proposalId": p.ID, "serverVersion": st.State.ServerVersion})
		return
	case "accepted", "rejected", "correction":
		legacyError(w, 409, "La solicitud ya fue resuelta y no puede cancelarse.", map[string]any{"status": p.Status, "proposalId": p.ID})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	p.Status = "cancelled"
	p.UpdatedAt = now
	p.DecidedAt = now
	p.DecidedBy = req.ActorID
	if strings.TrimSpace(req.Note) != "" {
		p.DecisionNote = strings.TrimSpace(req.Note)
	} else {
		p.DecisionNote = "Revisión cancelada por el usuario antes de la decisión administrativa."
	}
	p.Database = nil
	p.Files = []any{}
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "status": p.Status, "proposalId": p.ID, "note": p.DecisionNote, "serverVersion": st.State.ServerVersion})
}

type legacyMessageSendRequest struct {
	SenderID        string   `json:"senderId"`
	RecipientID     string   `json:"recipientId"`
	RecipientIDs    []string `json:"recipientIds"`
	ConversationID  string   `json:"conversationId"`
	Text            string   `json:"text"`
	ClientMessageID string   `json:"clientMessageId"`
}

type legacyMessageReadRequest struct {
	UserID     string   `json:"userId"`
	MessageIDs []string `json:"messageIds"`
}

func (s *Server) messageActor(r *http.Request, requested string) (id string, role string, ok bool) {
	if sess, remote := s.remoteSessionFromRequest(r); remote {
		return sess.UserID, sess.Role, true
	}
	authRole, authOK := s.auth(r)
	if !authOK || authRole != "admin" {
		return "", "", false
	}
	return requested, "admin", requested != ""
}

func messageRecipientList(req legacyMessageSendRequest, senderID string) []string {
	seen := map[string]bool{strings.TrimSpace(senderID): true}
	ids := []string{}
	add := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		ids = append(ids, id)
	}
	for _, id := range req.RecipientIDs {
		add(id)
	}
	add(req.RecipientID)
	return ids
}

func messageHasRecipient(m legacyMessage, userID string) bool {
	if m.RecipientID == userID {
		return true
	}
	for _, id := range m.RecipientIDs {
		if id == userID {
			return true
		}
	}
	return false
}

func messageInConversation(m legacyMessage, peer string) bool {
	if peer == "" {
		return true
	}
	if m.SenderID == peer || m.RecipientID == peer || m.ConversationID == peer {
		return true
	}
	for _, id := range m.RecipientIDs {
		if id == peer {
			return true
		}
	}
	return false
}

func roleCanSendGroupMessage(senderRole string, recipients []map[string]any) bool {
	if senderRole == "admin" {
		return true
	}
	hasAdmin := false
	for _, recipient := range recipients {
		if legacyRole(recipient) == "admin" {
			hasAdmin = true
			break
		}
	}
	return hasAdmin
}

func (s *Server) legacyMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	requested := r.URL.Query().Get("userId")
	userID, role, ok := s.messageActor(r, requested)
	if !ok {
		legacyError(w, 403, "No autorizado para consultar mensajes.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	if role == "admin" {
		u := legacyUser(st.State.Database, userID)
		if !legacyActive(u) {
			legacyError(w, 403, "El usuario no esta activo.", nil)
			return
		}
	}
	peer := r.URL.Query().Get("peerId")
	items := make([]legacyMessage, 0)
	unread := 0
	for _, m := range st.Messages {
		if m.SenderID != userID && !messageHasRecipient(m, userID) {
			continue
		}
		if !messageInConversation(m, peer) {
			continue
		}
		items = append(items, m)
		if messageHasRecipient(m, userID) && m.ReadAt == "" {
			unread++
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].SentAt < items[j].SentAt })
	if len(items) > 400 {
		items = items[len(items)-400:]
	}
	jsonOut(w, 200, map[string]any{"ok": true, "messages": items, "unread": unread, "serverTime": time.Now().UTC().Format(time.RFC3339Nano)})
}

func (s *Server) legacyMessageSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	var req legacyMessageSendRequest
	if err := readJSONLimited(r, &req, 1<<20); err != nil {
		legacyError(w, 400, "Mensaje inválido.", nil)
		return
	}
	senderID, senderRole, ok := s.messageActor(r, req.SenderID)
	if !ok {
		legacyError(w, 403, "No autorizado para enviar mensajes.", nil)
		return
	}
	req.SenderID = senderID
	text := strings.TrimSpace(req.Text)
	if text == "" || len([]rune(text)) > 2000 {
		legacyError(w, 400, "El mensaje debe tener entre 1 y 2000 caracteres.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	sender := legacyUser(st.State.Database, req.SenderID)
	recipientIDs := messageRecipientList(req, req.SenderID)
	if len(recipientIDs) == 0 || len(recipientIDs)+1 > 50 {
		legacyError(w, 400, "La conversación debe tener entre 2 y 50 participantes.", nil)
		return
	}
	recipients := make([]map[string]any, 0, len(recipientIDs))
	recipientNames := make([]string, 0, len(recipientIDs))
	for _, recipientID := range recipientIDs {
		recipient := legacyUser(st.State.Database, recipientID)
		if !legacyActive(recipient) {
			legacyError(w, 404, "Remitente o destinatario no disponible.", nil)
			return
		}
		recipients = append(recipients, recipient)
		recipientNames = append(recipientNames, legacyName(recipient))
	}
	if !legacyActive(sender) {
		legacyError(w, 404, "Remitente o destinatario no disponible.", nil)
		return
	}
	if senderRole == "admin" && legacyRole(sender) != "admin" {
		legacyError(w, 403, "El remitente no es Administrador.", nil)
		return
	}
	if senderRole != "admin" && legacyRole(sender) != senderRole {
		legacyError(w, 403, "El rol de la sesión no coincide con el remitente.", nil)
		return
	}
	if !roleCanSendGroupMessage(senderRole, recipients) {
		legacyError(w, 403, "Tutores, Regionales y Aspirantes deben incluir un Administrador en la conversación.", nil)
		return
	}
	if strings.TrimSpace(req.ClientMessageID) != "" {
		for _, existing := range st.Messages {
			if existing.SenderID == req.SenderID && existing.ClientMessageID == req.ClientMessageID {
				jsonOut(w, 200, map[string]any{"ok": true, "message": existing, "deduplicated": true})
				return
			}
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	conversationID := strings.TrimSpace(req.ConversationID)
	if conversationID == "" && len(recipientIDs) > 1 {
		parts := append([]string{req.SenderID}, recipientIDs...)
		sort.Strings(parts)
		conversationID = "conv_" + strings.Join(parts, "_")
	}
	msg := legacyMessage{
		ID: "msg_" + token(12), ClientMessageID: strings.TrimSpace(req.ClientMessageID),
		SenderID: req.SenderID, SenderName: legacyName(sender), SenderRole: legacyRole(sender),
		RecipientID: recipientIDs[0], RecipientName: recipientNames[0], RecipientIDs: recipientIDs, RecipientNames: recipientNames,
		ConversationID: conversationID, Text: text, SentAt: now,
	}
	st.Messages = append(st.Messages, msg)
	if len(st.Messages) > 5000 {
		st.Messages = st.Messages[len(st.Messages)-5000:]
	}
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 201, map[string]any{"ok": true, "message": msg})
}

func (s *Server) legacyMessageRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	var req legacyMessageReadRequest
	if err := readJSONLimited(r, &req, 1<<20); err != nil {
		legacyError(w, 400, "Lectura inválida.", nil)
		return
	}
	userID, role, ok := s.messageActor(r, req.UserID)
	if !ok {
		legacyError(w, 403, "No autorizado.", nil)
		return
	}
	req.UserID = userID
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	if role == "admin" {
		u := legacyUser(st.State.Database, userID)
		if !legacyActive(u) || legacyRole(u) != "admin" {
			legacyError(w, 403, "El usuario no es Administrador.", nil)
			return
		}
	}
	ids := map[string]bool{}
	for _, id := range req.MessageIDs {
		ids[id] = true
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	count := 0
	for i := range st.Messages {
		m := &st.Messages[i]
		if messageHasRecipient(*m, req.UserID) && m.ReadAt == "" && (len(ids) == 0 || ids[m.ID]) {
			m.ReadAt = now
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
