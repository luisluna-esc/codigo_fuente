package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sgl.local/sistema-gestion-liderazgo/internal/release"
	"sort"
	"strconv"
	"strings"
	"time"
)

// legacyState/legacyProposal implement the synchronization contract used by the
// 3.x client UI that is intentionally retained in SGL 4.2.13. The v4 entity API
// remains available in parallel under /api/v4/*.
type legacyState struct {
	ServerVersion     int64  `json:"serverVersion"`
	RestoreGeneration int64  `json:"restoreGeneration,omitempty"`
	RestoreID         string `json:"restoreId,omitempty"`
	RestoredAt        string `json:"restoredAt,omitempty"`
	Database          any    `json:"database"`
	Files             []any  `json:"files"`
	UpdatedAt         string `json:"updatedAt"`
	InstallationID    string `json:"installationId"`
}

type legacyProposal struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	ActorID          string `json:"actorId"`
	ActorName        string `json:"actorName"`
	ActorRole        string `json:"actorRole"`
	InstallationID   string `json:"installationId"`
	BaseVersion      int64  `json:"baseVersion"`
	Revision         int64  `json:"revision"`
	FileCount        int    `json:"fileCount"`
	SubmittedAt      string `json:"submittedAt"`
	UpdatedAt        string `json:"updatedAt"`
	DecidedAt        string `json:"decidedAt,omitempty"`
	DecidedBy        string `json:"decidedBy,omitempty"`
	DecisionNote     string `json:"decisionNote"`
	OperationID      string `json:"operationId,omitempty"`
	OperationType    string `json:"operationType,omitempty"`
	RequirementID    string `json:"requirementId,omitempty"`
	RequirementTitle string `json:"requirementTitle,omitempty"`
	OwnerID          string `json:"ownerId,omitempty"`
	ProgramID        string `json:"programId,omitempty"`
	ContentHash      string `json:"contentHash,omitempty"`
	CanonicalID      string `json:"canonicalId,omitempty"`
	Database         any    `json:"database,omitempty"`
	Files            []any  `json:"files,omitempty"`
}

type legacyMessage struct {
	ID              string   `json:"id"`
	ClientMessageID string   `json:"clientMessageId,omitempty"`
	SenderID        string   `json:"senderId"`
	SenderName      string   `json:"senderName"`
	SenderRole      string   `json:"senderRole"`
	RecipientID     string   `json:"recipientId"`
	RecipientName   string   `json:"recipientName"`
	RecipientIDs    []string `json:"recipientIds,omitempty"`
	RecipientNames  []string `json:"recipientNames,omitempty"`
	ConversationID  string   `json:"conversationId,omitempty"`
	Text            string   `json:"text"`
	SentAt          string   `json:"sentAt"`
	ReadAt          string   `json:"readAt,omitempty"`
}

type legacyAppliedOperation struct {
	OperationID   string `json:"operationId"`
	ActorID       string `json:"actorId"`
	OwnerID       string `json:"ownerId,omitempty"`
	OperationType string `json:"operationType"`
	RequirementID string `json:"requirementId,omitempty"`
	ProgramID     string `json:"programId,omitempty"`
	ContentHash   string `json:"contentHash,omitempty"`
	AppliedAt     string `json:"appliedAt"`
	ServerVersion int64  `json:"serverVersion"`
}

type legacyStore struct {
	Recoveries548     []recoveryReceipt548      `json:"recoveries548,omitempty"`
	State             legacyState               `json:"state"`
	Proposals         []legacyProposal          `json:"proposals"`
	Messages          []legacyMessage           `json:"messages,omitempty"`
	Notifications     []legacyNotification      `json:"notifications,omitempty"`
	UserChanges       []legacyUserChangeRequest `json:"userChanges,omitempty"`
	AppliedOperations []legacyAppliedOperation  `json:"appliedOperations,omitempty"`
}

type legacySyncRequest struct {
	InstallationID    string `json:"installationId"`
	RestoreGeneration int64  `json:"restoreGeneration,omitempty"`
	ProposalID        string `json:"proposalId"`
	ActorID           string `json:"actorId"`
	ActorRole         string `json:"actorRole"`
	ActorName         string `json:"actorName"`
	BaseVersion       int64  `json:"baseVersion"`
	OperationID       string `json:"operationId,omitempty"`
	OperationType     string `json:"operationType,omitempty"`
	RequirementID     string `json:"requirementId,omitempty"`
	RequirementTitle  string `json:"requirementTitle,omitempty"`
	OwnerID           string `json:"ownerId,omitempty"`
	ProgramID         string `json:"programId,omitempty"`
	ContentHash       string `json:"contentHash,omitempty"`
	AckOnly           bool   `json:"ackOnly,omitempty"`
	Database          any    `json:"database"`
	Patch             any    `json:"patch,omitempty"`
	Files             []any  `json:"files"`
}

type legacyDecisionRequest struct {
	ActorID     string `json:"actorId"`
	ProposalID  string `json:"proposalId"`
	Decision    string `json:"decision"`
	Note        string `json:"note"`
	BaseVersion int64  `json:"baseVersion"`
	Database    any    `json:"database"`
	Files       []any  `json:"files"`
}

func normalizeLegacySyncOperationType(req legacySyncRequest) string {
	operationType := strings.TrimSpace(req.OperationType)
	if operationType == "" {
		return "auto_sync"
	}
	switch operationType {
	case "auto_sync", "save_progress", "drive_finalize", "drive_reopen", "bible_reading_delete", "review", "cancel_review", "review_decision":
		return operationType
	case "check", "checked", "progress", "requirement", "requirement_save", "class_tracking", "profile_update", "photo_upload", "file_upload", "attachment", "normal", "normal_change", "data_change":
		return "auto_sync"
	default:
		// Any Aspirant/Tutor delta that reaches /api/sync is ordinary data unless
		// it explicitly uses the review channel. This keeps normal checkbox/profile/
		// class edits from being converted into Administrator approval requests.
		return "auto_sync"
	}
}

func automaticLegacyOperationType(operationType string) bool {
	switch strings.TrimSpace(operationType) {
	case "auto_sync", "save_progress", "drive_finalize", "drive_reopen", "bible_reading_delete", "review", "cancel_review", "review_decision":
		return true
	default:
		return false
	}
}

func legacyOperationChannel(operationType string) string {
	switch strings.TrimSpace(operationType) {
	case "review", "cancel_review", "review_decision":
		return "reviews"
	default:
		return "sync"
	}
}

func findAppliedLegacyOperation(st *legacyStore, actorID, operationID string) *legacyAppliedOperation {
	if st == nil || strings.TrimSpace(operationID) == "" {
		return nil
	}
	for i := range st.AppliedOperations {
		op := &st.AppliedOperations[i]
		if op.ActorID == actorID && op.OperationID == operationID {
			return op
		}
	}
	return nil
}

func rememberAppliedLegacyOperation(st *legacyStore, op legacyAppliedOperation) {
	if st == nil {
		return
	}
	st.AppliedOperations = append([]legacyAppliedOperation{op}, st.AppliedOperations...)
	if len(st.AppliedOperations) > 600 {
		st.AppliedOperations = st.AppliedOperations[:600]
	}
}

func (s *Server) legacySyncSnapshotForRequest(st legacyState, r *http.Request) map[string]any {
	if sess, remote := s.remoteSessionFromRequest(r); remote {
		return s.filteredLegacySnapshot(st, sess.UserID, sess.Role)
	}
	return legacySnapshot(st)
}

// legacyAutomaticAck returns a deliberately small confirmation for ordinary
// delta synchronization.  Previous versions returned the complete filtered
// database after every checkbox/change, which became increasingly slow once
// photos and documents accumulated.  Clients pull one canonical snapshot only
// after their local queue has drained.
func legacyAutomaticAck(st legacyState, op legacyAppliedOperation, deduplicated bool) map[string]any {
	return map[string]any{
		"ok": true, "status": "applied", "operationId": op.OperationID,
		"operationType": op.OperationType, "channel": legacyOperationChannel(op.OperationType),
		"requirementId": op.RequirementID, "programId": op.ProgramID, "appliedAt": op.AppliedAt,
		"serverVersion": st.ServerVersion, "restoreGeneration": legacyCanonicalRestoreGeneration(st),
		"restoreId": st.RestoreID, "restoredAt": st.RestoredAt, "deduplicated": deduplicated,
		"ackOnly": true, "clearLocalOperation": true,
	}
}

func (s *Server) legacyPath() string {
	return filepath.Join(filepath.Dir(s.path), "sgl-server-compat.json")
}

func (s *Server) legacySQLitePath() string {
	return filepath.Join(filepath.Dir(s.path), "sgl.db")
}

func normalizeLegacyStore(st *legacyStore) {
	if st == nil {
		return
	}
	if st.State.RestoreGeneration < 1 {
		st.State.RestoreGeneration = 1
	}
	// Repair metadata embedded by older releases without requiring a new restore.
	st.State.Database = legacyStampRestoreMetadata(st.State.Database, st.State)
	if st.State.Files == nil {
		st.State.Files = []any{}
	}
	// Loading or saving must preserve historical attachments and review decisions.
	// Validate new deliveries at submission time, never purge stored evidence here.
	if st.Proposals == nil {
		st.Proposals = []legacyProposal{}
	}
	if st.Messages == nil {
		st.Messages = []legacyMessage{}
	}
	if st.Notifications == nil {
		st.Notifications = []legacyNotification{}
	}
	if st.UserChanges == nil {
		st.UserChanges = []legacyUserChangeRequest{}
	}
	if st.AppliedOperations == nil {
		st.AppliedOperations = []legacyAppliedOperation{}
	}
}

func (s *Server) loadLegacyJSON() (legacyStore, bool, error) {
	var st legacyStore
	b, err := os.ReadFile(s.legacyPath())
	if os.IsNotExist(err) {
		normalizeLegacyStore(&st)
		return st, false, nil
	}
	if err != nil {
		return st, false, err
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return st, false, err
	}
	normalizeLegacyStore(&st)
	return st, true, nil
}

func (s *Server) writeLegacyJSON(st legacyStore) error {
	normalizeLegacyStore(&st)
	path := s.legacyPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	if runtimeIsWindows() {
		_ = os.Remove(path)
	}
	return os.Rename(tmp, path)
}

func (s *Server) preservePre517LegacyJSON() {
	src := s.legacyPath()
	backup := filepath.Join(filepath.Dir(src), "sgl-server-compat.pre-5.17.json")
	if _, err := os.Stat(backup); err == nil {
		return
	}
	b, err := os.ReadFile(src)
	if err != nil {
		return
	}
	if err := os.WriteFile(backup, b, 0600); err != nil {
		log.Printf("SGL %s: no se pudo crear respaldo previo de persistencia: %v", release.Version, err)
	}
}

func (s *Server) noteSQLiteFallback(err error) {
	if s.legacySQLiteFallbackLogged {
		return
	}
	s.legacySQLiteFallbackLogged = true
	log.Printf("SGL %s: SQLite no disponible; se conserva persistencia JSON compatible: %v", release.Version, err)
}

func (s *Server) loadLegacy() (legacyStore, error) {
	if s.legacyCacheLoaded {
		return s.legacyCache, nil
	}

	// Windows 10/11 includes winsqlite3.dll. SGL uses it directly, without CGO
	// or an installer dependency. If it is absent/corrupt, JSON remains a safe
	// fallback and the HTTP/API contract does not change.
	if legacySQLiteRuntimeAvailable() {
		if st, found, err := legacySQLiteLoad(s.legacySQLitePath()); err == nil && found {
			normalizeLegacyStore(&st)
			s.cleanStandaloneMusicStore(&st)
			s.legacyCache = st
			s.legacyCacheLoaded = true
			s.legacyStorageMode = "sqlite"
			s.legacyPersistedFilesSig = legacyFilesSignature(st.State.Files)
			return st, nil
		} else if err != nil {
			s.noteSQLiteFallback(err)
			s.legacyStorageMode = "json-fallback"
		}
	}

	st, jsonFound, err := s.loadLegacyJSON()
	if err != nil {
		return st, err
	}
	s.cleanStandaloneMusicStore(&st)

	// One-time migration is intentionally behind loadLegacy/saveLegacy. Every
	// endpoint above this layer keeps the exact same behavior and validation.
	if jsonFound && s.legacyStorageMode != "json-fallback" && legacySQLiteRuntimeAvailable() {
		if err := legacySQLiteSave(s.legacySQLitePath(), st, true); err == nil {
			s.preservePre517LegacyJSON()
			s.legacyStorageMode = "sqlite"
		} else {
			s.noteSQLiteFallback(err)
			s.legacyStorageMode = "json-fallback"
		}
	}
	if s.legacyStorageMode == "" {
		s.legacyStorageMode = "json"
	}
	s.legacyCache = st
	s.legacyCacheLoaded = true
	s.legacyPersistedFilesSig = legacyFilesSignature(st.State.Files)
	return st, nil
}

func (s *Server) saveLegacy(st legacyStore) error {
	normalizeLegacyStore(&st)
	musicAutoCertificates610(&st)
	musicFollowupNotifications610(&st)
	s.cleanStandaloneMusicStore(&st)
	fileSig := legacyFilesSignature(st.State.Files)
	persistFiles := s.legacyStorageMode != "sqlite" || fileSig != s.legacyPersistedFilesSig

	if s.legacyStorageMode != "json-fallback" && legacySQLiteRuntimeAvailable() {
		if err := legacySQLiteSave(s.legacySQLitePath(), st, persistFiles); err == nil {
			if s.legacyStorageMode != "sqlite" {
				s.preservePre517LegacyJSON()
			}
			s.legacyStorageMode = "sqlite"
			s.legacyCache = st
			s.legacyCacheLoaded = true
			s.legacyPersistedFilesSig = fileSig
			return nil
		} else {
			s.noteSQLiteFallback(err)
			s.legacyStorageMode = "json-fallback"
		}
	}

	if err := s.writeLegacyJSON(st); err != nil {
		return err
	}
	if s.legacyStorageMode == "" {
		s.legacyStorageMode = "json"
	}
	s.legacyCache = st
	s.legacyCacheLoaded = true
	s.legacyPersistedFilesSig = fileSig
	return nil
}

// flushLegacyCompatibilityJSON creates a current rollback/export mirror only
// when the SQLite backend is active. It is called after HTTP shutdown, so this
// full JSON serialization does not sit on the normal synchronization hot path.
func (s *Server) flushLegacyCompatibilityJSON() error {
	if !s.legacyCacheLoaded || s.legacyStorageMode != "sqlite" {
		return nil
	}
	return s.writeLegacyJSON(s.legacyCache)
}

func runtimeIsWindows() bool { return os.PathSeparator == '\\' }

func cloneAny(v any) any {
	if v == nil {
		return nil
	}
	b, _ := json.Marshal(v)
	var out any
	_ = json.Unmarshal(b, &out)
	return out
}

func legacyMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	b, _ := json.Marshal(v)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return m
}

func legacyUsers(db any) []any {
	m := legacyMap(db)
	if m == nil {
		return nil
	}
	if a, ok := m["users"].([]any); ok {
		return a
	}
	return nil
}

func legacyUser(db any, id string) map[string]any {
	for _, raw := range legacyUsers(db) {
		m := legacyMap(raw)
		if fmt.Sprint(m["id"]) == id {
			return m
		}
	}
	return nil
}

func legacyActive(user map[string]any) bool {
	if user == nil {
		return false
	}
	if state := strings.TrimSpace(fmt.Sprint(user["accountState"])); state != "" && state != "<nil>" && state != "active" {
		return false
	}
	if v, exists := user["active"]; exists {
		if b, ok := v.(bool); ok && !b {
			return false
		}
	}
	return true
}

func canonicalRole548(role string) string {
	if role == "coordinador_sgm" {
		return "admin"
	}
	if role == "administrativo_sgm" {
		return "secretaria"
	}
	if role == "lider_musica" {
		return "tutor"
	}
	if role == "pastor_aspirant" {
		return "aspirant"
	}
	return role
}
func legacyRole(user map[string]any) string { return canonicalRole548(fmt.Sprint(user["role"])) }
func viewAll548(role string) bool {
	return role == "admin" || role == "departamental" || role == "administrativo" || role == "secretaria" || role == "coordinador_musica"
}
func legacyName(user map[string]any) string { return fmt.Sprint(user["fullName"]) }

func legacyDBRevision(db any) int64 {
	m := legacyMap(db)
	if m == nil {
		return 0
	}
	switch v := m["revision"].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case json.Number:
		n, _ := v.Int64()
		return n
	default:
		n, _ := strconv.ParseInt(fmt.Sprint(v), 10, 64)
		return n
	}
}

func legacyEnforceSGL427Rules(db any) any { return db }

func preserveLegacyCredentials(existingDB, incomingDB any) any {
	out := legacyMap(cloneAny(incomingDB))
	if out == nil {
		return incomingDB
	}
	existingByID := map[string]map[string]any{}
	for _, raw := range legacyUsers(existingDB) {
		u := legacyMap(raw)
		if id := strings.TrimSpace(fmt.Sprint(u["id"])); id != "" && id != "<nil>" {
			existingByID[id] = u
		}
	}
	users := legacyUsers(out)
	for _, raw := range users {
		u := legacyMap(raw)
		prev := existingByID[strings.TrimSpace(fmt.Sprint(u["id"]))]
		if prev == nil {
			continue
		}
		if _, supplied := u["password"]; !supplied && prev["password"] != nil {
			u["password"] = cloneAny(prev["password"])
		}
		if _, supplied := u["recoveryKeys"]; !supplied && prev["recoveryKeys"] != nil {
			u["recoveryKeys"] = cloneAny(prev["recoveryKeys"])
		}
	}
	out["users"] = users
	return out
}

func legacyPrepareDB(db any, serverVersion int64, message string) any {
	out := legacyMap(cloneAny(db))
	if out == nil {
		return db
	}
	legacyEnforceSGL427Rules(out)
	if settings := legacyMap(out["settings"]); settings != nil {
		delete(settings, "serverToken")
		delete(settings, "serverUrl")
		out["settings"] = settings
	}
	// restorePending is device-local transient state. It must never become part
	// of the canonical server database or a later bootstrap can replay a restore.
	delete(out, "restorePending")
	out["serverVersion"] = serverVersion
	out["sync"] = map[string]any{
		"state":           "ok",
		"lastSyncedAt":    time.Now().UTC().Format(time.RFC3339Nano),
		"message":         message,
		"pendingProposal": nil,
	}
	return out
}

func legacyCanonicalRestoreGeneration(st legacyState) int64 {
	if st.RestoreGeneration < 1 {
		return 1
	}
	return st.RestoreGeneration
}

// legacyStampRestoreMetadata mirrors the authoritative state metadata into the
// database object sent to clients. Older/client-filtered bootstrap paths often
// replace their whole local database with payload.database, so keeping these
// fields only at the response top level can silently downgrade a device to
// generation 1 and create an endless RESTORE_REQUIRED -> relogin loop.
func legacyStampRestoreMetadata(db any, st legacyState) any {
	out := legacyMap(db)
	if out == nil {
		return db
	}
	out["serverVersion"] = st.ServerVersion
	out["restoreGeneration"] = legacyCanonicalRestoreGeneration(st)
	if strings.TrimSpace(st.RestoreID) != "" {
		out["restoreId"] = st.RestoreID
	} else {
		delete(out, "restoreId")
	}
	if strings.TrimSpace(st.RestoredAt) != "" {
		out["restoredAt"] = st.RestoredAt
	} else {
		delete(out, "restoredAt")
	}
	// This marker belongs to one browser/device, never to server truth.
	out["restorePending"] = false
	return out
}

func legacyPrepareCanonicalDB(db any, st legacyState, message string) any {
	return legacyStampRestoreMetadata(legacyPrepareDB(db, st.ServerVersion, message), st)
}

func legacySnapshot(st legacyState) map[string]any {
	db := cloneAny(st.Database)
	legacyEnforceSGL427Rules(db)
	db = legacyStampRestoreMetadata(db, st)
	if data := legacyMap(db); data != nil {
		data["syncCapabilities"] = map[string]any{"assignmentDeltaVersion": 1}
	}
	return map[string]any{
		"ok":                true,
		"initialized":       st.Database != nil,
		"serverVersion":     st.ServerVersion,
		"restoreGeneration": legacyCanonicalRestoreGeneration(st),
		"restoreId":         st.RestoreID,
		"restoredAt":        st.RestoredAt,
		"database":          db,
		"files":             st.Files,
		"updatedAt":         st.UpdatedAt,
	}
}

func legacySummary(p legacyProposal) map[string]any {
	return map[string]any{
		"id": p.ID, "status": p.Status, "actorId": p.ActorID, "actorName": p.ActorName,
		"actorRole": p.ActorRole, "installationId": p.InstallationID, "baseVersion": p.BaseVersion,
		"revision": p.Revision, "fileCount": p.FileCount, "submittedAt": p.SubmittedAt,
		"updatedAt": p.UpdatedAt, "decidedAt": p.DecidedAt, "decidedBy": p.DecidedBy,
		"decisionNote": p.DecisionNote, "operationId": p.OperationID, "operationType": p.OperationType,
		"requirementId": p.RequirementID, "programId": p.ProgramID, "contentHash": p.ContentHash, "canonicalId": p.CanonicalID,
	}
}

func findLegacyProposal(st *legacyStore, id string) (int, *legacyProposal) {
	for i := range st.Proposals {
		if st.Proposals[i].ID == id {
			return i, &st.Proposals[i]
		}
	}
	return -1, nil
}

func legacyError(w http.ResponseWriter, status int, msg string, extra map[string]any) {
	m := map[string]any{"error": msg}
	for k, v := range extra {
		m[k] = v
	}
	jsonOut(w, status, m)
}

func (s *Server) legacyHealth(w http.ResponseWriter, r *http.Request) {
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
	jsonOut(w, 200, map[string]any{"ok": true, "app": systemName, "apiVersion": 3, "compatVersion": "5.6", "serverRelease": release.Version, "storage": s.legacyStorageMode, "initialized": st.State.Database != nil, "serverVersion": st.State.ServerVersion, "time": time.Now().UTC().Format(time.RFC3339Nano)})
}

// legacySyncDiagnostics is intentionally small: it validates the configured token/session
// and reports only the synchronization metadata needed to diagnose one device. It never
// returns the database contents.
func (s *Server) legacySyncDiagnostics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	scope, ok := s.auth(r)
	if !ok {
		legacyError(w, 401, "Token o sesión no válidos.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	actorID := strings.TrimSpace(r.URL.Query().Get("actorId"))
	if sess, remote := s.remoteSessionFromRequest(r); remote {
		actorID = sess.UserID
	}
	actor := legacyUser(st.State.Database, actorID)
	actorRole := ""
	actorName := ""
	actorActive := false
	if actor != nil {
		actorRole = legacyRole(actor)
		actorName = legacyName(actor)
		actorActive = legacyActive(actor)
	}
	jsonOut(w, 200, map[string]any{
		"ok":              true,
		"compatVersion":   "5.6",
		"apiVersion":      3,
		"serverRelease":   release.Version,
		"storage":         s.legacyStorageMode,
		"initialized":     st.State.Database != nil,
		"serverVersion":   st.State.ServerVersion,
		"serverUpdatedAt": st.State.UpdatedAt,
		"authScope":       scope,
		"musicPortal":     1,
		"actorId":         actorID,
		"actorFound":      actor != nil,
		"actorActive":     actorActive,
		"actorRole":       actorRole,
		"actorName":       actorName,
		"time":            time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func (s *Server) legacyBootstrap(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.auth(r); !ok {
		legacyError(w, 401, "Token incorrecto.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	if sess, ok := s.remoteSessionFromRequest(r); ok {
		jsonOut(w, 200, s.filteredLegacySnapshot(st.State, sess.UserID, sess.Role))
		return
	}
	jsonOut(w, 200, legacySnapshot(st.State))
}

func (s *Server) legacySync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	scope, ok := s.auth(r)
	if !ok {
		legacyError(w, 401, "Token incorrecto.", nil)
		return
	}
	var req legacySyncRequest
	if err := readJSONLimited(r, &req, 256<<20); err != nil || (req.Database == nil && req.Patch == nil) {
		legacyError(w, 400, "No se pudieron leer los datos de sincronización.", nil)
		return
	}
	remoteSess, remoteRequest := s.remoteSessionFromRequest(r)
	if remoteRequest {
		sess := remoteSess
		if req.ActorID != sess.UserID || (req.ActorRole != "" && canonicalRole548(req.ActorRole) != canonicalRole548(sess.Role)) {
			legacyError(w, 403, "La sesión remota no corresponde al usuario enviado.", nil)
			return
		}
		req.ActorID = sess.UserID
		req.ActorRole = sess.Role
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	candidate561 := req.Database
	if req.Patch != nil {
		candidate561 = expandLegacyPatch(st.State.Database, req.Patch)
	}
	if reason := musicSyncGuard602(legacyMap(st.State.Database), legacyUser(st.State.Database, req.ActorID), legacyMap(candidate561)); reason != "" {
		legacyError(w, 403, reason, nil)
		return
	}
	if message := musicAssignmentPatch561(legacyMap(st.State.Database), legacyUser(st.State.Database, req.ActorID), candidate561); message != "" {
		legacyError(w, 403, message, nil)
		return
	}
	currentGeneration := st.State.RestoreGeneration
	if st.State.Database != nil && (req.ProgramID == musicProgram553 || musicRequest553(r)) {
		candidate := req.Database
		if req.Patch != nil {
			candidate = expandLegacyPatch(st.State.Database, req.Patch)
		}
		if message := musicAccountPatch560(legacyMap(st.State.Database), legacyUser(st.State.Database, req.ActorID), candidate); message != "" {
			legacyError(w, 403, message, nil)
			return
		}
	}
	if currentGeneration < 1 {
		currentGeneration = 1
	}
	generationMatches := req.RestoreGeneration == currentGeneration || (req.RestoreGeneration == 0 && currentGeneration == 1)
	if !generationMatches {
		legacyError(w, http.StatusConflict, "El sistema fue restaurado. Debes descargar la copia vigente del servidor antes de sincronizar.", map[string]any{
			"code": "RESTORE_REQUIRED", "restoreGeneration": currentGeneration, "restoreId": st.State.RestoreID, "restoredAt": st.State.RestoredAt, "serverVersion": st.State.ServerVersion,
		})
		return
	}

	if req.ActorRole == "admin" {
		if legacyMap(req.Database)["accountScope555"] != nil {
			legacyError(w, 409, "Actualiza la vista de Gestión antes de publicar. La tarjeta personal no reemplaza la base administrativa.", nil)
			return
		}
		if scope != "admin" {
			legacyError(w, 403, "Esta operación requiere el token de Administrador.", nil)
			return
		}
		if st.State.Database == nil {
			actor := legacyUser(req.Database, req.ActorID)
			if !legacyActive(actor) || legacyRole(actor) != "admin" {
				legacyError(w, 403, "La instalación inicial debe contener al Administrador que está sincronizando.", nil)
				return
			}
			// 4.2.26: when Tailscale Funnel starts a clean local server, the
			// Administrator may already have a local data revision. An empty
			// server must accept that first publication as version 0 instead of
			// rejecting it and leaving the UI at "falta publicar la base".
			req.BaseVersion = 0
		} else {
			actor := legacyUser(st.State.Database, req.ActorID)
			if !legacyActive(actor) || legacyRole(actor) != "admin" {
				legacyError(w, 403, "El usuario no es un Administrador activo del servidor.", nil)
				return
			}
			if req.BaseVersion != st.State.ServerVersion {
				legacyError(w, 409, "Conflicto de versión.", map[string]any{"serverVersion": st.State.ServerVersion, "database": st.State.Database, "files": st.State.Files})
				return
			}
		}
		st.State.ServerVersion++
		incomingDB := req.Database
		if remoteRequest && st.State.Database != nil {
			// El cliente remoto 5.0.9 no recibe hashes. Si no envía uno nuevo,
			// conservar el verificador existente antes de reemplazar la base.
			incomingDB = preserveLegacyCredentials(st.State.Database, incomingDB)
		}
		st.State.Database = legacyPrepareCanonicalDB(preserveMusicResource550(st.State.Database, incomingDB), st.State, "Sincronizado por el Administrador")
		st.State.Files = req.Files
		if st.State.Files == nil {
			st.State.Files = []any{}
		}
		st.State.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		st.State.InstallationID = req.InstallationID
		if err := s.saveLegacy(st); err != nil {
			legacyError(w, 500, err.Error(), nil)
			return
		}
		jsonOut(w, 200, legacySnapshot(st.State))
		return
	}

	if st.State.Database == nil {
		legacyError(w, 409, "El servidor todavía no fue inicializado por un Administrador.", map[string]any{"serverVersion": int64(0), "database": nil, "files": []any{}})
		return
	}
	actor := legacyUser(st.State.Database, req.ActorID)
	if !legacyActive(actor) {
		legacyError(w, 403, "El usuario no existe o no está activo en el servidor.", nil)
		return
	}
	role := requestRole555(r, actor)
	if role != "aspirant" && role != "tutor" {
		legacyError(w, 403, "Solo Aspirantes y Tutores pueden sincronizar cambios.", nil)
		return
	}
	if role == "aspirant" && req.ProgramID != "" && !assignedProgram555(legacyMap(st.State.Database), req.ActorID, req.ProgramID) && !(len(arrayField(legacyMap(st.State.Database), "modules")) == 0 && req.ProgramID == "class_grouped") {
		legacyError(w, 403, "No puedes guardar avances en un módulo que no está asignado a tu cuenta.", nil)
		return
	}
	if req.ActorRole != "" && canonicalRole548(req.ActorRole) != role {
		legacyError(w, 403, "El rol enviado no coincide con el rol autorizado.", nil)
		return
	}
	if req.BaseVersion > st.State.ServerVersion {
		payload := s.legacySyncSnapshotForRequest(st.State, r)
		payload["error"] = "La computadora indica una versión futura que el servidor no reconoce."
		jsonOut(w, 409, payload)
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	opType := normalizeLegacySyncOperationType(req)
	req.OperationType = opType
	if automaticLegacyOperationType(opType) {
		candidateDB := req.Database
		if req.Patch != nil {
			candidateDB = expandLegacyPatch(st.State.Database, req.Patch)
		}
		opID := strings.TrimSpace(req.OperationID)
		if opID == "" {
			opID = generatedLegacyOperationID(req.ActorID, req.InstallationID, opType, candidateDB)
			req.OperationID = opID
		}
		if previous := findAppliedLegacyOperation(&st, req.ActorID, opID); previous != nil {
			if previous.OperationType != opType || previous.ProgramID != strings.TrimSpace(req.ProgramID) || previous.RequirementID != strings.TrimSpace(req.RequirementID) || previous.ContentHash != strings.TrimSpace(req.ContentHash) {
				legacyError(w, 409, "Identificador de envío reutilizado para otro contenido. El pendiente se conserva.", map[string]any{"code": "OPERATION_CONFLICT", "clearLocalOperation": false})
				return
			}

			if req.AckOnly {
				jsonOut(w, 200, legacyAutomaticAck(st.State, *previous, true))
				return
			}
			payload := s.legacySyncSnapshotForRequest(st.State, r)
			payload["status"] = "applied"
			payload["operationId"] = previous.OperationID
			payload["operationType"] = previous.OperationType
			payload["channel"] = legacyOperationChannel(previous.OperationType)
			payload["clearLocalOperation"] = true
			payload["requirementId"] = previous.RequirementID
			payload["programId"] = previous.ProgramID
			payload["appliedAt"] = previous.AppliedAt
			payload["deduplicated"] = true
			jsonOut(w, 200, payload)
			return
		}
		candidate := &legacyProposal{
			ActorID: req.ActorID, ActorName: legacyName(actor), ActorRole: role, InstallationID: req.InstallationID,
			BaseVersion: req.BaseVersion, Revision: legacyDBRevision(candidateDB), OperationID: opID, OperationType: opType,
			RequirementID: strings.TrimSpace(req.RequirementID), RequirementTitle: strings.TrimSpace(req.RequirementTitle), OwnerID: strings.TrimSpace(req.OwnerID), ProgramID: strings.TrimSpace(req.ProgramID), ContentHash: strings.TrimSpace(req.ContentHash),
			Database: candidateDB, Files: req.Files,
		}
		merged, allowed, mergeErr := mergeLegacyProposalOnServer(st.State.Database, candidate)
		if personalRequest555(r) {
			merged, allowed, mergeErr = mergePersonal555(st.State.Database, candidate)
		}
		if mergeErr != nil {
			legacyError(w, 422, "El cambio automático ya no es aplicable a los datos actuales del servidor.", map[string]any{"operationId": opID, "operationType": opType})
			return
		}
		st.State.ServerVersion++
		st.State.Database = legacyPrepareCanonicalDB(merged, st.State, "Sincronización automática")
		st.State.Files = mergeLegacyProposalFiles(st.State.Files, req.Files, req.ActorID, allowed)
		if st.State.Files == nil {
			st.State.Files = []any{}
		}
		st.State.UpdatedAt = now
		rememberAppliedLegacyOperation(&st, legacyAppliedOperation{OperationID: opID, ActorID: req.ActorID, OperationType: opType, RequirementID: candidate.RequirementID, ProgramID: candidate.ProgramID, ContentHash: candidate.ContentHash, AppliedAt: now, ServerVersion: st.State.ServerVersion})
		s.addAutomaticSyncNotifications(&st, req, actor, role, now)
		if err := s.saveLegacy(st); err != nil {
			legacyError(w, 500, err.Error(), nil)
			return
		}
		applied := legacyAppliedOperation{OperationID: opID, ActorID: req.ActorID, OperationType: opType, RequirementID: candidate.RequirementID, ProgramID: candidate.ProgramID, ContentHash: candidate.ContentHash, AppliedAt: now, ServerVersion: st.State.ServerVersion}
		if req.AckOnly {
			jsonOut(w, 200, legacyAutomaticAck(st.State, applied, false))
			return
		}
		payload := s.legacySyncSnapshotForRequest(st.State, r)
		payload["status"] = "applied"
		payload["operationId"] = opID
		payload["operationType"] = opType
		payload["channel"] = legacyOperationChannel(opType)
		payload["clearLocalOperation"] = true
		payload["requirementId"] = candidate.RequirementID
		payload["programId"] = candidate.ProgramID
		payload["appliedAt"] = now
		payload["automatic"] = true
		jsonOut(w, 200, payload)
		return
	}
	if changed, _ := dedupeLegacyProposals(&st); changed {
		_ = s.saveLegacy(st)
	}
	var proposal *legacyProposal
	// 4.2.4: operationId is the idempotency key. A retry of the same saved
	// action must always resolve to the same proposal, even after a timeout.
	if strings.TrimSpace(req.OperationID) != "" {
		for i := range st.Proposals {
			p := &st.Proposals[i]
			if p.ActorID == req.ActorID && p.OperationID == req.OperationID {
				proposal = p
				break
			}
		}
	}
	if proposal == nil && req.ProposalID != "" {
		_, p := findCanonicalProposal(&st, req.ProposalID)
		if p != nil && p.ActorID == req.ActorID && p.InstallationID == req.InstallationID {
			proposal = p
		}
	}
	// Repair retries made by 4.2.3 and earlier: the same local revision from
	// the same actor/device represents the same explicit saved operation.
	if proposal == nil {
		revision := legacyDBRevision(req.Database)
		for i := range st.Proposals {
			p := &st.Proposals[i]
			if p.ActorID == req.ActorID && p.InstallationID == req.InstallationID && p.Revision == revision && (p.Status == "pending" || p.Status == "postponed") {
				proposal = p
				break
			}
		}
	}
	if proposal != nil && proposal.Status != "pending" && proposal.Status != "postponed" {
		jsonOut(w, 202, map[string]any{"ok": true, "proposalId": proposal.ID, "status": proposal.Status, "submittedAt": proposal.SubmittedAt, "updatedAt": proposal.UpdatedAt, "serverVersion": st.State.ServerVersion, "deduplicated": true})
		return
	}
	if proposal == nil {
		p := legacyProposal{ID: "sync_" + token(12), Status: "pending", ActorID: req.ActorID, ActorName: legacyName(actor), ActorRole: role, InstallationID: req.InstallationID, SubmittedAt: now}
		st.Proposals = append([]legacyProposal{p}, st.Proposals...)
		proposal = &st.Proposals[0]
	}
	proposal.ActorName = legacyName(actor)
	proposal.ActorRole = role
	proposal.BaseVersion = req.BaseVersion
	proposal.Revision = legacyDBRevision(req.Database)
	proposal.FileCount = len(req.Files)
	proposal.UpdatedAt = now
	proposal.OperationID = strings.TrimSpace(req.OperationID)
	proposal.OperationType = strings.TrimSpace(req.OperationType)
	proposal.RequirementID = strings.TrimSpace(req.RequirementID)
	proposal.ProgramID = strings.TrimSpace(req.ProgramID)
	proposal.ContentHash = strings.TrimSpace(req.ContentHash)
	proposal.DecisionNote = ""
	proposal.DecidedAt = ""
	proposal.DecidedBy = ""
	proposal.Database = req.Database
	proposal.Files = req.Files
	if proposal.Files == nil {
		proposal.Files = []any{}
	}
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 202, map[string]any{"ok": true, "proposalId": proposal.ID, "status": proposal.Status, "operationId": proposal.OperationID, "operationType": proposal.OperationType, "requirementId": proposal.RequirementID, "programId": proposal.ProgramID, "submittedAt": proposal.SubmittedAt, "updatedAt": proposal.UpdatedAt, "serverVersion": st.State.ServerVersion})
}

func (s *Server) legacyProposals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	role, ok := s.auth(r)
	if !ok || role != "admin" {
		legacyError(w, 403, "Esta ruta requiere el token de Administrador.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	if changed, _ := dedupeLegacyProposals(&st); changed {
		if err := s.saveLegacy(st); err != nil {
			legacyError(w, 500, err.Error(), nil)
			return
		}
	}
	actorID := r.URL.Query().Get("actorId")
	actor := legacyUser(st.State.Database, actorID)
	if !legacyActive(actor) || legacyRole(actor) != "admin" {
		legacyError(w, 403, "El usuario no es un Administrador activo.", nil)
		return
	}
	if id := r.URL.Query().Get("id"); id != "" {
		_, p := findCanonicalProposal(&st, id)
		if p == nil {
			legacyError(w, 404, "Solicitud no encontrada.", nil)
			return
		}
		jsonOut(w, 200, p)
		return
	}
	sort.SliceStable(st.Proposals, func(i, j int) bool { return st.Proposals[i].SubmittedAt > st.Proposals[j].SubmittedAt })
	limit := len(st.Proposals)
	if limit > 150 {
		limit = 150
	}
	items := make([]map[string]any, 0, limit)
	for _, p := range st.Proposals[:limit] {
		items = append(items, legacySummary(p))
	}
	jsonOut(w, 200, map[string]any{"ok": true, "proposals": items, "serverVersion": st.State.ServerVersion})
}

func (s *Server) legacyProposalStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := s.auth(r); !ok {
		legacyError(w, 401, "Token incorrecto.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	actorID := r.URL.Query().Get("actorId")
	if sess, remote := s.remoteSessionFromRequest(r); remote {
		if actorID != sess.UserID {
			legacyError(w, 403, "La sesión no corresponde a este usuario.", nil)
			return
		}
		actorID = sess.UserID
	}
	_, p := findCanonicalProposal(&st, r.URL.Query().Get("proposalId"))
	if p == nil || p.ActorID != actorID || p.InstallationID != r.URL.Query().Get("installationId") {
		legacyError(w, 404, "Solicitud no encontrada para este equipo y usuario.", nil)
		return
	}
	payload := map[string]any{"ok": true, "proposalId": p.ID, "status": p.Status, "note": p.DecisionNote, "operationId": p.OperationID, "operationType": p.OperationType, "requirementId": p.RequirementID, "programId": p.ProgramID, "submittedAt": p.SubmittedAt, "updatedAt": p.UpdatedAt, "serverVersion": st.State.ServerVersion}
	if p.Status == "accepted" {
		if sess, remote := s.remoteSessionFromRequest(r); remote {
			filtered := s.filteredLegacySnapshot(st.State, sess.UserID, sess.Role)
			payload["database"] = filtered["database"]
			payload["files"] = filtered["files"]
		} else {
			payload["database"] = st.State.Database
			payload["files"] = st.State.Files
		}
	}
	jsonOut(w, 200, payload)
}

func (s *Server) legacyProposalDecision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	role, ok := s.auth(r)
	if !ok || role != "admin" {
		legacyError(w, 403, "Esta operación requiere el token de Administrador.", nil)
		return
	}
	var req legacyDecisionRequest
	if err := readJSONLimited(r, &req, 256<<20); err != nil {
		legacyError(w, 400, "Falta la decisión.", nil)
		return
	}
	if req.Decision != "accept" && req.Decision != "reject" && req.Decision != "correction" && req.Decision != "postpone" {
		legacyError(w, 400, "Decisión no válida.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	admin := legacyUser(st.State.Database, req.ActorID)
	if !legacyActive(admin) || legacyRole(admin) != "admin" {
		legacyError(w, 403, "El usuario no es un Administrador activo.", nil)
		return
	}
	if changed, _ := dedupeLegacyProposals(&st); changed {
		_ = s.saveLegacy(st)
	}
	_, p := findCanonicalProposal(&st, req.ProposalID)
	if p == nil {
		legacyError(w, 404, "Solicitud no encontrada.", nil)
		return
	}
	// Idempotencia: repetir la misma decisión por un timeout no genera error
	// ni una segunda aprobación. Devuelve el estado ya consolidado.
	if p.Status != "pending" && p.Status != "postponed" {
		if p.Status == "accepted" && req.Decision == "accept" {
			jsonOut(w, 200, legacySnapshot(st.State))
			return
		}
		if (p.Status == "rejected" && req.Decision == "reject") || (p.Status == "correction" && req.Decision == "correction") {
			jsonOut(w, 200, map[string]any{"ok": true, "status": p.Status, "proposalId": p.ID, "note": p.DecisionNote, "serverVersion": st.State.ServerVersion, "idempotent": true})
			return
		}
		legacyError(w, 409, "La solicitud ya fue resuelta.", map[string]any{"status": p.Status, "proposalId": p.ID})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if req.Decision == "postpone" {
		p.Status, p.UpdatedAt, p.DecisionNote, p.DecidedBy = "postponed", now, req.Note, req.ActorID
		if err := s.saveLegacy(st); err != nil {
			legacyError(w, 500, err.Error(), nil)
			return
		}
		jsonOut(w, 200, map[string]any{"ok": true, "status": p.Status, "proposalId": p.ID, "note": req.Note, "serverVersion": st.State.ServerVersion})
		return
	}
	if req.Decision == "accept" {
		acceptedDB := req.Database
		acceptedFiles := req.Files
		acceptMode := "client-merge"
		if acceptedDB == nil {
			// 4.2.5+ clients send a compact decision. The server applies the same
			// restricted merge that older clients performed in the browser. This
			// keeps the request tiny without changing the normal sync mechanism.
			merged, allowed, mergeErr := mergeLegacyProposalOnServer(st.State.Database, p)
			if mergeErr != nil {
				// A request without an applicable source snapshot is historical debris.
				// Close it explicitly so it cannot keep blocking the review inbox.
				p.Status = "obsolete"
				p.UpdatedAt, p.DecidedAt, p.DecidedBy = now, now, req.ActorID
				p.DecisionNote = "Solicitud antigua/no aplicable. No se modificaron los datos actuales."
				p.Database, p.Files = nil, []any{}
				if err := s.saveLegacy(st); err != nil {
					legacyError(w, 500, err.Error(), nil)
					return
				}
				legacyError(w, 422, p.DecisionNote, map[string]any{"proposalId": p.ID, "status": p.Status})
				return
			}
			acceptedDB = merged
			acceptedFiles = mergeLegacyProposalFiles(st.State.Files, p.Files, p.ActorID, allowed)
			acceptMode = "server-merge"
		} else if req.BaseVersion != st.State.ServerVersion {
			// Backward compatibility for 4.2.4 and earlier clients that still post
			// the full merged database. Their optimistic concurrency check remains.
			legacyError(w, 409, "La base del servidor cambi\u00f3 mientras se revisaba la solicitud.", map[string]any{"serverVersion": st.State.ServerVersion, "database": st.State.Database, "files": st.State.Files})
			return
		}
		mergedAdmin := legacyUser(acceptedDB, req.ActorID)
		if !legacyActive(mergedAdmin) || legacyRole(mergedAdmin) != "admin" {
			legacyError(w, 400, "La versi\u00f3n aceptada debe conservar al Administrador que autoriza el cambio.", nil)
			return
		}
		st.State.ServerVersion++
		st.State.Database = legacyPrepareCanonicalDB(preserveMusicResource550(st.State.Database, acceptedDB), st.State, "Cambios aceptados por el Administrador")
		st.State.Files = acceptedFiles
		if st.State.Files == nil {
			st.State.Files = []any{}
		}
		st.State.UpdatedAt = now
		p.Status, p.UpdatedAt, p.DecidedAt, p.DecidedBy, p.DecisionNote = "accepted", now, now, req.ActorID, req.Note
		p.Database, p.Files = nil, []any{}
		if err := s.saveLegacy(st); err != nil {
			legacyError(w, 500, err.Error(), nil)
			return
		}
		payload := legacySnapshot(st.State)
		payload["acceptMode"] = acceptMode
		payload["proposalId"] = p.ID
		jsonOut(w, 200, payload)
		return
	}

	decisionStatus := req.Decision
	if decisionStatus == "reject" {
		decisionStatus = "rejected"
	}
	p.Status, p.UpdatedAt, p.DecidedAt, p.DecidedBy, p.DecisionNote = decisionStatus, now, now, req.ActorID, req.Note
	p.Database, p.Files = nil, []any{}
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "status": p.Status, "proposalId": p.ID, "note": req.Note, "serverVersion": st.State.ServerVersion})
}

func (s *Server) legacyBackup(w http.ResponseWriter, r *http.Request) {
	role, ok := s.auth(r)
	if !ok || role != "admin" {
		legacyError(w, 403, "Esta ruta requiere el token de Administrador.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "exportedAt": time.Now().UTC().Format(time.RFC3339Nano), "state": st.State, "proposals": st.Proposals})
}
