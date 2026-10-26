package main

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// This manifest contains only IDs and hashes, never credentials or file bytes.
//
//go:embed reconcile_598_manifest.json
var reconciliationManifest598 []byte

type reconcileManifest598 struct {
	ActorID          string            `json:"actorId"`
	SourceGeneration int64             `json:"sourceGeneration"`
	Generation       int64             `json:"generation"`
	RestoreID        string            `json:"restoreId"`
	PasswordBefore   string            `json:"passwordBefore"`
	Patches          map[string]string `json:"patches"`
}

func hash598(v any) string {
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
func manifest598() reconcileManifest598 {
	var m reconcileManifest598
	_ = json.Unmarshal(reconciliationManifest598, &m)
	return m
}

const ghost598 = "usr_4fb6c55b-d6a3-4986-8259-85114f432fb1"
const franklin598 = "usr_71f99e2e-044f-4292-b7d2-8a60ded73808"
const noemi598 = "usr_2b4e672a-7399-43d8-b20c-3f580f22b1f9"

// Reconcile only the exact three audited operations. No restore, generation
// increment, broad patch replay, credential export, or replacement of files.
func reconcileState598(st *legacyStore, actorID string, operations []any, now string) (map[string]any, string) {
	m := manifest598()
	db := legacyMap(st.State.Database)
	if actorID != m.ActorID || legacyRole(legacyUser(db, actorID)) != "admin" {
		return nil, "Solo el administrador propietario de estos pendientes puede reconciliarlos."
	}
	if legacyCanonicalRestoreGeneration(st.State) != m.Generation || st.State.RestoreID != m.RestoreID {
		return nil, "Esta base no coincide con la generación auditada. No se ha modificado nada."
	}
	if len(operations) != len(m.Patches) {
		return nil, "Se requieren exactamente los tres pendientes auditados. Conserva los cambios adicionales para revisión."
	}
	seen := map[string]bool{}
	var password any
	for _, raw := range operations {
		op := legacyMap(raw)
		id := stringValue(op["id"])
		if seen[id] || m.Patches[id] == "" || hash598(op["patch"]) != m.Patches[id] || stringValue(op["actorId"]) != actorID || int64(numericValue(op["restoreGeneration"])) != m.SourceGeneration || stringValue(op["action"]) != "admin_change" || stringValue(op["programId"]) != musicProgram553 {
			return nil, "Los pendientes cambiaron desde la copia auditada. No se han aplicado ni eliminado."
		}
		seen[id] = true
		for _, u := range arrayField(legacyMap(op["patch"]), "users") {
			if p := legacyMap(u)["password"]; p != nil {
				password = p
			}
		}
	}
	// An existing receipt is authoritative even after subsequent legitimate edits.
	if receipt := legacyMap(db["reconciledS3598"]); receipt != nil && receipt["actorId"] == actorID && receipt["restoreId"] == m.RestoreID {
		return receipt, ""
	}
	noemi := legacyUser(db, noemi598)
	franklin := legacyUser(db, franklin598)
	if noemi == nil || franklin == nil || password == nil {
		return nil, "Faltan las cuentas o el cambio original. No se ha modificado nada."
	}
	if hash598(noemi["password"]) != m.PasswordBefore && hash598(noemi["password"]) != hash598(password) {
		return nil, "La contraseña de Noemí cambió después de la auditoría. Se conserva la más reciente; hace falta revisar ese conflicto."
	}
	module := map[string]any(nil)
	for _, raw := range arrayField(db, "modules") {
		v := legacyMap(raw)
		if stringValue(v["id"]) == musicProgram553 && v["active"] != false {
			module = v
		}
	}
	if module == nil {
		return nil, "El módulo de Música no está disponible."
	}
	ghost := legacyUser(db, ghost598)
	if ghost != nil {
		if reason := musicEmptyReason597(db, ghost, st.State.Files); reason != "" {
			return nil, "El registro vacío cambió: " + reason
		}
	}
	// All validations precede all mutations. Enrollments are additive; existing
	// progress and tutors (including every Conquistadores module) are retained.
	for _, id := range []string{franklin598, noemi598} {
		if !containsString(arrayField(module, "assignedUserIds"), id) {
			module["assignedUserIds"] = append(arrayField(module, "assignedUserIds"), id)
		}
	}
	noemi["password"] = cloneAny(password)
	noemi["updatedAt"] = now
	if ghost != nil {
		db["musicEmptyArchive597"] = append(arrayField(db, "musicEmptyArchive597"), map[string]any{"user": cloneAny(ghost), "programProgress": cloneAny(legacyMap(db["programProgress"])[ghost598]), "removedAt": now, "removedBy": actorID})
		users := []any{}
		for _, u := range legacyUsers(db) {
			if stringValue(legacyMap(u)["id"]) != ghost598 {
				users = append(users, u)
			}
		}
		db["users"] = users
		ids := []any{}
		for _, id := range arrayField(module, "assignedUserIds") {
			if stringValue(id) != ghost598 {
				ids = append(ids, id)
			}
		}
		module["assignedUserIds"] = ids
		delete(legacyMap(db["programProgress"]), ghost598)
		delete(legacyMap(db["progress"]), ghost598)
	}
	st.State.ServerVersion++
	st.State.UpdatedAt = now
	ids := []any{}
	for _, raw := range operations {
		op := legacyMap(raw)
		id := stringValue(op["id"])
		ids = append(ids, id)
		rememberAppliedLegacyOperation(st, legacyAppliedOperation{OperationID: id, ActorID: actorID, OperationType: "admin_reconciled598", ProgramID: musicProgram553, ContentHash: stringValue(op["contentHash"]), AppliedAt: now, ServerVersion: st.State.ServerVersion})
	}
	receipt := map[string]any{"id": "SGM-S3-598", "actorId": actorID, "operationIds": ids, "generation": m.Generation, "restoreId": m.RestoreID, "appliedAt": now, "serverVersion": st.State.ServerVersion}
	db["reconciledS3598"] = receipt
	st.State.Database = legacyPrepareCanonicalDB(db, st.State, "S3 reconciliado: inscripciones recuperadas, credencial original confirmada y registro vacío depurado")
	return receipt, ""
}
func (s *Server) reconcile598(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	var req struct {
		ActorID    string `json:"actorId"`
		Operations []any  `json:"operations"`
	}
	if err := readJSONLimited(r, &req, 4<<20); err != nil {
		legacyError(w, 400, "Solicitud inválida", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, "No se pudo abrir la base", nil)
		return
	}
	id, _, ok := s.accountActor556(r, req.ActorID, &st)
	if !ok || personalRequest555(r) || legacyRole(legacyUser(st.State.Database, id)) != "admin" {
		legacyError(w, 403, "Acceso exclusivo del administrador", nil)
		return
	}
	before := st.State.ServerVersion
	// Clone the complete store: failed validations/saves cannot mutate live cache.
	b, err := json.Marshal(st)
	if err != nil {
		legacyError(w, 500, "No se pudo preparar el respaldo", nil)
		return
	}
	var next legacyStore
	if json.Unmarshal(b, &next) != nil {
		legacyError(w, 500, "No se pudo preparar la copia de trabajo", nil)
		return
	}
	receipt, message := reconcileState598(&next, id, req.Operations, time.Now().UTC().Format(time.RFC3339Nano))
	if message != "" {
		legacyError(w, 409, message, nil)
		return
	}
	if next.State.ServerVersion != before {
		folder := filepath.Join(filepath.Dir(s.path), "Backups")
		if err = os.MkdirAll(folder, 0700); err == nil {
			err = os.WriteFile(filepath.Join(folder, "antes-reconciliar-S3-"+time.Now().UTC().Format("20060102T150405.000000000")+".json"), b, 0600)
		}
		if err != nil {
			legacyError(w, 500, "No se pudo guardar el respaldo previo; no se modificó la base", nil)
			return
		}
		if err = s.saveLegacy(next); err != nil {
			legacyError(w, 500, "No se confirmó el guardado; conserva los pendientes y reintenta", nil)
			return
		}
	}
	jsonOut(w, 200, map[string]any{"ok": true, "receipt": receipt, "serverVersion": next.State.ServerVersion})
}
