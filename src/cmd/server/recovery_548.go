package main

// Recovery is separate from normal synchronization: a generation mismatch never
// grants permission to overwrite live data. Only an administrator can commit.
import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"time"
)

type recoveryCapsule548 struct {
	Format         string         `json:"format"`
	OwnerID        string         `json:"ownerId"`
	InstallationID string         `json:"installationId"`
	Secret         string         `json:"secret"`
	Fingerprint    string         `json:"fingerprint"`
	Progress       map[string]any `json:"progress"`
	Programs       map[string]any `json:"programs"`
	Bible          any            `json:"bible"`
	Files          []any          `json:"files"`
	SourceUser     map[string]any `json:"sourceUser,omitempty"`
	SourceVersion  string         `json:"sourceVersion,omitempty"`
}
type recoveryReceipt548 struct {
	ID             string `json:"id"`
	OwnerID        string `json:"ownerId"`
	InstallationID string `json:"installationId"`
	SecretHash     string `json:"secretHash"`
	Fingerprint    string `json:"fingerprint"`
	Generation     int64  `json:"generation"`
	RestoreID      string `json:"restoreId"`
	CreatedAt      string `json:"createdAt"`
}
type recoveryRequest548 struct {
	Capsule         recoveryCapsule548 `json:"capsule"`
	Commit          bool               `json:"commit"`
	BaseVersion     int64              `json:"baseVersion"`
	BaseFingerprint string             `json:"baseFingerprint,omitempty"`
	Choices         map[string]string  `json:"choices"`
}

func recoveryHash548(value string) string {
	h := sha256.Sum256([]byte(value))
	return hex.EncodeToString(h[:])
}
func recoveryPrograms548(db any, owner string) map[string]any {
	out := map[string]any{"class_grouped": cloneAny(mapField(legacyMap(db), "progress")[owner])}
	for k, v := range legacyMap(mapField(legacyMap(db), "programProgress")[owner]) {
		if k != "class_grouped" {
			out[k] = cloneAny(v)
		}
	}
	return out
}
func recoveryRecord548(p map[string]any, id string) map[string]any {
	out := map[string]any{}
	for _, field := range []string{"submissions", "status", "readings"} {
		if v, ok := mapField(p, field)[id]; ok {
			out[field] = v
		}
	}
	return out
}
func (s *Server) adminRecovery548(w http.ResponseWriter, r *http.Request) {
	role, ok := s.auth(r)
	if !ok || role != "admin" {
		legacyError(w, 403, "Esta acción requiere Administrador.", nil)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	var req recoveryRequest548
	if err := readJSONLimited(r, &req, 256<<20); err != nil {
		legacyError(w, 400, "No se pudo leer la copia.", nil)
		return
	}
	c := req.Capsule
	if c.Format != "SGLRecovery548" || c.OwnerID == "" || c.InstallationID == "" || len(c.Secret) != 64 || len(c.Fingerprint) != 64 {
		legacyError(w, 400, "Exporta una copia de recuperación desde el equipo del aspirante actualizado.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	user := legacyUser(st.State.Database, c.OwnerID)
	skipped := []string{}
	if c.SourceUser != nil {
		var prepareErr error
		c, user, skipped, prepareErr = s.prepareBackupRecovery554(st.State.Database, c)
		if prepareErr != nil {
			legacyError(w, 400, prepareErr.Error(), nil)
			return
		}
	}
	if !leader555(legacyMap(st.State.Database), user) {
		legacyError(w, 400, "El aspirante de la copia no existe o no está activo en este servidor.", nil)
		return
	}
	// Work on a deep copy: even a validation or disk failure must not mutate cache.
	cloned, _ := json.Marshal(st)
	var next legacyStore
	if err = json.Unmarshal(cloned, &next); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	local := map[string]any{"class_grouped": c.Progress}
	for k, v := range c.Programs {
		if k != "class_grouped" {
			local[k] = v
		}
	}
	server := recoveryPrograms548(next.State.Database, c.OwnerID)
	rows := []any{}
	keys := []string{}
	for pid := range local {
		keys = append(keys, pid)
	}
	sort.Strings(keys)
	for _, pid := range keys {
		lp := legacyMap(local[pid])
		sp := legacyMap(server[pid])
		if sp == nil {
			sp = map[string]any{}
			server[pid] = sp
		}
		ids := map[string]bool{}
		for _, field := range []string{"submissions", "status", "readings"} {
			for id := range mapField(lp, field) {
				ids[id] = true
			}
		}
		ordered := []string{}
		for id := range ids {
			ordered = append(ordered, id)
		}
		sort.Strings(ordered)
		for _, id := range ordered {
			l, sr := recoveryRecord548(lp, id), recoveryRecord548(sp, id)
			if c.SourceUser != nil && !recoveryHasWork554(l) {
				continue
			}
			if reflect.DeepEqual(l, sr) {
				continue
			}
			key := pid + "|" + id
			missing := len(sr) == 0 || (c.SourceUser != nil && !recoveryHasWork554(sr))
			rows = append(rows, map[string]any{"key": key, "programId": pid, "requirementId": id, "local": l, "server": sr, "missing": missing})
			if req.Commit {
				choice := req.Choices[key]
				if choice != "local" && choice != "server" {
					legacyError(w, 400, "Selecciona qué versión conservar en cada diferencia.", nil)
					return
				}
				if choice == "local" {
					for _, field := range []string{"submissions", "status", "readings"} {
						target := ensureMapField(sp, field)
						if v, exists := l[field]; exists {
							target[id] = cloneAny(v)
						} else if c.SourceUser == nil {
							delete(target, id)
						}
					}
				}
			}
		}
	}
	db := legacyMap(next.State.Database)
	serverBible := mapField(db, "bibleYear")[c.OwnerID]
	if c.Bible != nil && !reflect.DeepEqual(c.Bible, serverBible) {
		rows = append(rows, map[string]any{"key": "bible", "programId": "Lectura bíblica", "requirementId": "Lecturas", "local": c.Bible, "server": serverBible, "missing": serverBible == nil})
		if req.Commit {
			if req.Choices["bible"] == "local" {
				ensureMapField(db, "bibleYear")[c.OwnerID] = cloneAny(c.Bible)
			} else if req.Choices["bible"] != "server" {
				legacyError(w, 400, "Selecciona la versión de lectura bíblica.", nil)
				return
			}
		}
	}
	existing := map[string]any{}
	for _, f := range next.State.Files {
		existing[stringValue(legacyMap(f)["id"])] = f
	}
	for _, f := range c.Files {
		file := legacyMap(f)
		id := stringValue(file["id"])
		if id == "" || stringValue(file["ownerId"]) != c.OwnerID {
			legacyError(w, 400, "La copia contiene archivos ajenos o sin identificador.", nil)
			return
		}
		if old, exists := existing[id]; exists {
			if stringValue(legacyMap(old)["ownerId"]) != c.OwnerID || stringValue(legacyMap(old)["dataUrl"]) != stringValue(file["dataUrl"]) {
				legacyError(w, 409, "Hay un archivo con el mismo identificador y diferente contenido. No se ha reemplazado ningún dato.", nil)
				return
			}
		} else {
			next.State.Files = append(next.State.Files, cloneAny(f))
			existing[id] = f
		}
	}
	if !req.Commit {
		jsonOut(w, 200, map[string]any{"ownerName": user["fullName"], "ownerId": c.OwnerID, "baseVersion": st.State.ServerVersion, "baseFingerprint": recoveryStateFingerprint554(st.State, c.OwnerID), "rows": rows, "skipped": skipped, "sourceVersion": c.SourceVersion, "fileCount": len(c.Files)})
		return
	}
	if (req.BaseFingerprint != "" && req.BaseFingerprint != recoveryStateFingerprint554(st.State, c.OwnerID)) || (req.BaseFingerprint == "" && req.BaseVersion != st.State.ServerVersion) {
		legacyError(w, 409, "El servidor cambió. Vuelve a comparar la copia antes de recuperar.", nil)
		return
	}
	now := time.Now().UTC()
	if err = s.savePreRestoreSafety(st, now); err != nil {
		legacyError(w, 500, "No se pudo guardar la copia de seguridad del servidor.", nil)
		return
	}
	// Retain the imported version too, even when the administrator chooses server.
	archive, err := json.Marshal(req)
	if err == nil {
		err = os.WriteFile(filepath.Join(filepath.Dir(s.legacyPath()), fmt.Sprintf("sgl-recovery-%d.json", now.UnixNano())), archive, 0600)
	}
	if err != nil {
		legacyError(w, 500, "No se pudo conservar la copia del aspirante.", nil)
		return
	}
	ensureMapField(db, "progress")[c.OwnerID] = server["class_grouped"]
	for pid, p := range server {
		if pid != "class_grouped" {
			ensureMapField(ensureMapField(db, "programProgress"), c.OwnerID)[pid] = p
		}
	}
	receipt := recoveryReceipt548{ID: "recovery_" + token(16), OwnerID: c.OwnerID, InstallationID: c.InstallationID, SecretHash: recoveryHash548(c.Secret), Fingerprint: c.Fingerprint, Generation: st.State.RestoreGeneration, RestoreID: st.State.RestoreID, CreatedAt: now.Format(time.RFC3339Nano)}
	next.Recoveries548 = append(next.Recoveries548, receipt)
	next.State.ServerVersion++
	next.State.UpdatedAt = receipt.CreatedAt
	// Recovery changes only this owner and canonical synchronization metadata.
	db["serverVersion"] = next.State.ServerVersion
	next.State.Database = legacyStampRestoreMetadata(db, next.State)
	if err = s.saveLegacy(next); err != nil {
		legacyError(w, 500, "No se pudo confirmar la recuperación.", nil)
		return
	}
	jsonOut(w, 200, map[string]any{"recovered": true, "receiptId": receipt.ID, "ownerId": c.OwnerID, "database": next.State.Database, "files": next.State.Files, "serverVersion": next.State.ServerVersion, "restoreGeneration": next.State.RestoreGeneration, "restoreId": next.State.RestoreID})
}
func (s *Server) recoveryStatus548(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	if _, ok := s.auth(r); !ok {
		legacyError(w, 401, "Inicia sesión para consultar la recuperación.", nil)
		return
	}
	var req struct {
		OwnerID        string `json:"ownerId"`
		InstallationID string `json:"installationId"`
		Secret         string `json:"secret"`
	}
	if readJSONLimited(r, &req, 4096) != nil || len(req.Secret) != 64 {
		legacyError(w, 400, "Solicitud inválida.", nil)
		return
	}
	if sess, remote := s.remoteSessionFromRequest(r); remote && sess.UserID != req.OwnerID {
		legacyError(w, 403, "La recuperación corresponde a otra cuenta.", nil)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	if !legacyActive(legacyUser(st.State.Database, req.OwnerID)) {
		legacyError(w, 403, "Cuenta inactiva.", nil)
		return
	}
	for i := len(st.Recoveries548) - 1; i >= 0; i-- {
		receipt := st.Recoveries548[i]
		if receipt.OwnerID != req.OwnerID || receipt.InstallationID != req.InstallationID || receipt.SecretHash != recoveryHash548(req.Secret) || receipt.Generation != st.State.RestoreGeneration || receipt.RestoreID != st.State.RestoreID {
			continue
		}
		receipt.SecretHash = ""
		files := []any{}
		for _, f := range st.State.Files {
			if stringValue(legacyMap(f)["ownerId"]) == req.OwnerID {
				files = append(files, f)
			}
		}
		jsonOut(w, 200, map[string]any{"ready": true, "recovery": map[string]any{"receipt": receipt, "programs": recoveryPrograms548(st.State.Database, req.OwnerID), "bible": mapField(legacyMap(st.State.Database), "bibleYear")[req.OwnerID], "files": files, "serverVersion": st.State.ServerVersion}})
		return
	}
	jsonOut(w, 200, map[string]any{"ready": false})
}
