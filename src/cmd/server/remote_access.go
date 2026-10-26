package main

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sgl.local/sistema-gestion-liderazgo/internal/release"
	"strings"
	"time"
)

type remoteSession struct {
	Product     string `json:"product,omitempty"`
	TokenHash   string `json:"tokenHash"`
	UserID      string `json:"userId"`
	Role        string `json:"role"`
	DeviceID    string `json:"deviceId,omitempty"`
	IP          string `json:"ip,omitempty"`
	CreatedAt   string `json:"createdAt"`
	LastSeen    string `json:"lastSeen"`
	HeartbeatAt string `json:"heartbeatAt,omitempty"`
	ExpiresAt   string `json:"expiresAt"`
}

type remoteSessionStore struct {
	Sessions []remoteSession `json:"sessions"`
}
type loginAttempt struct {
	Count       int
	First       time.Time
	LockedUntil time.Time
}

type remoteAuthUser struct {
	UserID      string
	Role        string
	FullName    string
	Active      bool
	Password    map[string]any
	AmbiguousID bool
}

// remoteAuthByIdentity maintains a compact authentication index in memory.
// It is rebuilt only when the legacy data file changes, so repeated logins do
// not deserialize the whole database on every password attempt.
func (s *Server) remoteAuthByIdentity(identity string) (remoteAuthUser, error) {
	identity = strings.ToLower(strings.TrimSpace(identity))
	if identity == "" {
		return remoteAuthUser{}, nil
	}
	// 5.18: authentication invalidation follows the canonical server version,
	// not the compatibility JSON mtime (which can be only a rollback mirror
	// while SQLite is active).
	s.mu.Lock()
	legacyStateNow, stateErr := s.loadLegacy()
	stamp := fmt.Sprintf("%d:%s", legacyStateNow.State.ServerVersion, legacyStateNow.State.UpdatedAt)
	s.mu.Unlock()
	if stateErr != nil {
		return remoteAuthUser{}, stateErr
	}
	s.authMu.Lock()
	defer s.authMu.Unlock()
	if s.authIndex == nil || s.authIndexStamp != stamp {
		s.mu.Lock()
		state, err := s.loadLegacyLoginState()
		s.mu.Unlock()
		if err != nil {
			return remoteAuthUser{}, err
		}
		index := map[string]remoteAuthUser{}
		for _, raw := range legacyUsers(state.Database) {
			u := legacyMap(raw)
			if u == nil {
				continue
			}
			// El indice de autenticacion guarda solo lo imprescindible. No conserva
			// recoveryKeys, telefono, fecha de nacimiento ni el resto del perfil.
			item := remoteAuthUser{UserID: fmt.Sprint(u["id"]), Role: legacyRole(u), FullName: legacyName(u), Active: legacyActive(u), Password: legacyMap(u["password"])}
			for _, candidate := range remoteIdentity(u) {
				if candidate == "" {
					continue
				}
				if previous, exists := index[candidate]; exists && previous.UserID != item.UserID {
					// Nunca elegir silenciosamente entre dos cuentas con la misma identidad.
					index[candidate] = remoteAuthUser{AmbiguousID: true}
					continue
				}
				index[candidate] = item
			}
		}
		s.authIndex = index
		s.authIndexStamp = stamp
	}
	return s.authIndex[identity], nil
}

func hashToken(v string) string { h := sha256.Sum256([]byte(v)); return hex.EncodeToString(h[:]) }
func publicServerCode(serverID string) string {
	clean := strings.ToUpper(strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, serverID))
	if len(clean) < 8 {
		sum := sha256.Sum256([]byte(serverID))
		clean = strings.ToUpper(hex.EncodeToString(sum[:]))
	}
	return "SGL-" + clean[:4] + "-" + clean[len(clean)-4:]
}
func (s *Server) remoteSessionsPath() string {
	return filepath.Join(filepath.Dir(s.path), "sgl-remote-sessions.json")
}

func (s *Server) loadRemoteSessions() {
	s.remoteSessions = map[string]remoteSession{}
	b, err := os.ReadFile(s.remoteSessionsPath())
	if err != nil {
		return
	}
	var st remoteSessionStore
	if json.Unmarshal(b, &st) != nil {
		return
	}
	now := time.Now()
	for _, sess := range st.Sessions {
		exp, _ := time.Parse(time.RFC3339, sess.ExpiresAt)
		if exp.After(now) {
			s.remoteSessions[sess.TokenHash] = sess
		}
	}
}
func (s *Server) saveRemoteSessionsLocked() {
	list := make([]remoteSession, 0, len(s.remoteSessions))
	now := time.Now()
	for k, v := range s.remoteSessions {
		exp, _ := time.Parse(time.RFC3339, v.ExpiresAt)
		if exp.After(now) {
			list = append(list, v)
		} else {
			delete(s.remoteSessions, k)
		}
	}
	b, _ := json.MarshalIndent(remoteSessionStore{Sessions: list}, "", "  ")
	_ = os.MkdirAll(filepath.Dir(s.remoteSessionsPath()), 0755)
	tmp := s.remoteSessionsPath() + ".tmp"
	if os.WriteFile(tmp, b, 0600) == nil {
		if runtimeIsWindows() {
			_ = os.Remove(s.remoteSessionsPath())
		}
		_ = os.Rename(tmp, s.remoteSessionsPath())
	}
}

func remoteIdentity(user map[string]any) []string {
	return []string{strings.ToLower(strings.TrimSpace(fmt.Sprint(user["username"]))), strings.ToLower(strings.TrimSpace(fmt.Sprint(user["email"])))}
}
func legacyUserByIdentity(db any, identity string) map[string]any {
	identity = strings.ToLower(strings.TrimSpace(identity))
	if identity == "" {
		return nil
	}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		for _, candidate := range remoteIdentity(u) {
			if candidate != "" && candidate == identity {
				return u
			}
		}
	}
	return nil
}

func pbkdf2SHA256(password, salt []byte, iterations, keyLen int) []byte {
	hLen := 32
	blocks := (keyLen + hLen - 1) / hLen
	out := make([]byte, 0, blocks*hLen)
	// Mantiene exactamente PBKDF2-SHA256 y las mismas iteraciones, pero reutiliza
	// los buffers de U/T para no asignar memoria en cada una de las 150.000 rondas.
	mac := hmac.New(sha256.New, password)
	u := make([]byte, hLen)
	t := make([]byte, hLen)
	for i := 1; i <= blocks; i++ {
		mac.Reset()
		_, _ = mac.Write(salt)
		var ib [4]byte
		binary.BigEndian.PutUint32(ib[:], uint32(i))
		_, _ = mac.Write(ib[:])
		u = mac.Sum(u[:0])
		copy(t, u)
		for j := 1; j < iterations; j++ {
			mac.Reset()
			_, _ = mac.Write(u)
			u = mac.Sum(u[:0])
			for k := range t {
				t[k] ^= u[k]
			}
		}
		out = append(out, t...)
	}
	return out[:keyLen]
}
func verifyRemotePasswordRecord(password string, rec map[string]any) bool {
	if rec == nil {
		return false
	}
	salt, err1 := base64.StdEncoding.DecodeString(fmt.Sprint(rec["salt"]))
	expected, err2 := base64.StdEncoding.DecodeString(fmt.Sprint(rec["hash"]))
	if err1 != nil || err2 != nil || len(expected) == 0 {
		return false
	}
	iterations := 150000
	switch v := rec["iterations"].(type) {
	case float64:
		iterations = int(v)
	case json.Number:
		n, _ := v.Int64()
		iterations = int(n)
	}
	// Evita registros corruptos que reduzcan la seguridad o provoquen un DoS de CPU.
	if iterations < 1000 || iterations > 2000000 || len(expected) > 128 {
		return false
	}
	got := pbkdf2SHA256([]byte(password), salt, iterations, len(expected))
	return subtle.ConstantTimeCompare(got, expected) == 1
}

func verifyRemotePassword(password string, user map[string]any) bool {
	return verifyRemotePasswordRecord(password, legacyMap(user["password"]))
}

func requestIP(r *http.Request) string {
	if x := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); x != "" {
		return x
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
func (s *Server) loginAllowed(key string) (bool, time.Duration) {
	s.loginMu.Lock()
	defer s.loginMu.Unlock()
	if s.loginAttempts == nil {
		s.loginAttempts = map[string]loginAttempt{}
	}
	a := s.loginAttempts[key]
	now := time.Now()
	if now.Before(a.LockedUntil) {
		return false, time.Until(a.LockedUntil)
	}
	if a.First.IsZero() || now.Sub(a.First) > 10*time.Minute {
		delete(s.loginAttempts, key)
		return true, 0
	}
	return true, 0
}
func (s *Server) loginFailed(key string) {
	s.loginMu.Lock()
	defer s.loginMu.Unlock()
	if s.loginAttempts == nil {
		s.loginAttempts = map[string]loginAttempt{}
	}
	now := time.Now()
	a := s.loginAttempts[key]
	if a.First.IsZero() || now.Sub(a.First) > 10*time.Minute {
		a = loginAttempt{First: now}
	}
	a.Count++
	if a.Count >= 5 {
		a.LockedUntil = now.Add(2 * time.Minute)
		a.Count = 0
		a.First = now
	}
	s.loginAttempts[key] = a
}
func (s *Server) loginSucceeded(key string) {
	s.loginMu.Lock()
	defer s.loginMu.Unlock()
	delete(s.loginAttempts, key)
}

// loadLegacyLoginState reads only the authentication snapshot needed by a login.
// The compatibility store can contain many megabytes of evidence in state.files;
// decoding those files before checking a password made remote login progressively
// slower as a club accumulated photos and PDFs. The store is written atomically
// by saveLegacy, so returning after database + serverVersion are decoded is safe.
// If the optimized parser cannot locate the expected state/database shape, callers
// fall back to loadLegacy for backwards compatibility.
func (s *Server) loadLegacyLoginState() (legacyState, error) {
	var out legacyState
	// 5.18 keeps the canonical compatibility state parsed in memory. Clone only
	// the database because remote snapshot filtering mutates its working copy;
	// evidence blobs stay deferred and are never copied during login.
	if s.legacyCacheLoaded {
		out = s.legacyCache.State
		out.Database = cloneAny(s.legacyCache.State.Database)
		out.Files = []any{}
		return out, nil
	}
	f, err := os.Open(s.legacyPath())
	if os.IsNotExist(err) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	tok, err := dec.Token()
	if err != nil {
		return out, err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return out, fmt.Errorf("almacen SGL invalido")
	}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return out, err
		}
		key, _ := keyTok.(string)
		if key != "state" {
			var skip json.RawMessage
			if err := dec.Decode(&skip); err != nil {
				return out, err
			}
			continue
		}
		tok, err = dec.Token()
		if err != nil {
			return out, err
		}
		if delim, ok := tok.(json.Delim); !ok || delim != '{' {
			return out, fmt.Errorf("estado SGL invalido")
		}
		haveDB := false
		haveVersion := false
		for dec.More() {
			fieldTok, err := dec.Token()
			if err != nil {
				return out, err
			}
			field, _ := fieldTok.(string)
			switch field {
			case "serverVersion":
				if err := dec.Decode(&out.ServerVersion); err != nil {
					return out, err
				}
				haveVersion = true
			case "restoreGeneration":
				if err := dec.Decode(&out.RestoreGeneration); err != nil {
					return out, err
				}
			case "restoreId":
				if err := dec.Decode(&out.RestoreID); err != nil {
					return out, err
				}
			case "restoredAt":
				if err := dec.Decode(&out.RestoredAt); err != nil {
					return out, err
				}
			case "database":
				if err := dec.Decode(&out.Database); err != nil {
					return out, err
				}
				haveDB = true
			default:
				// Marshal order writes serverVersion and database before files.
				// Do not decode a potentially huge files array once auth data is ready.
				if haveDB && haveVersion {
					if out.RestoreGeneration < 1 {
						out.RestoreGeneration = 1
					}
					out.Files = []any{}
					return out, nil
				}
				var skip json.RawMessage
				if err := dec.Decode(&skip); err != nil {
					return out, err
				}
			}
			if haveDB && haveVersion {
				if out.RestoreGeneration < 1 {
					out.RestoreGeneration = 1
				}
				out.Files = []any{}
				return out, nil
			}
		}
		return out, fmt.Errorf("almacen SGL sin base de datos de autenticacion")
	}
	return out, fmt.Errorf("almacen SGL sin estado")
}

func jsonOutRemote(w http.ResponseWriter, r *http.Request, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Add("Vary", "Accept-Encoding")
	if strings.Contains(strings.ToLower(r.Header.Get("Accept-Encoding")), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(status)
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			return
		}
		_ = json.NewEncoder(gz).Encode(v)
		_ = gz.Close()
		return
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) remoteSessionFromRequest(r *http.Request) (remoteSession, bool) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return remoteSession{}, false
	}
	raw := strings.TrimSpace(h[7:])
	if raw == "" {
		return remoteSession{}, false
	}
	key := hashToken(raw)
	s.remoteMu.Lock()
	defer s.remoteMu.Unlock()
	sess, ok := s.remoteSessions[key]
	if !ok {
		return remoteSession{}, false
	}
	exp, err := time.Parse(time.RFC3339, sess.ExpiresAt)
	if err != nil || time.Now().After(exp) {
		delete(s.remoteSessions, key)
		s.saveRemoteSessionsLocked()
		return remoteSession{}, false
	}
	sess.LastSeen = time.Now().UTC().Format(time.RFC3339)
	s.remoteSessions[key] = sess
	if personalRequest555(r) {
		sess.Role = "aspirant"
	}
	return sess, true
}

func (s *Server) remoteHeartbeatSessionFromRequest(r *http.Request) (remoteSession, bool) {
	if _, ok := s.remoteSessionFromRequest(r); !ok {
		return remoteSession{}, false
	}
	h := r.Header.Get("Authorization")
	raw := ""
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		raw = strings.TrimSpace(h[7:])
	}
	if raw == "" {
		return remoteSession{}, false
	}
	key := hashToken(raw)
	now := time.Now().UTC().Format(time.RFC3339)
	s.remoteMu.Lock()
	defer s.remoteMu.Unlock()
	current, exists := s.remoteSessions[key]
	if !exists {
		return remoteSession{}, false
	}
	current.LastSeen = now
	current.HeartbeatAt = now
	s.remoteSessions[key] = current
	return current, true
}

func (s *Server) remoteInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", 405)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "remoteAccess": s.cfg.PublicWeb, "app": systemName, "version": release.Version, "assetVersion": release.AssetVersion, "serverName": s.cfg.ServerName, "serverId": s.store.ServerID, "serverCode": publicServerCode(s.store.ServerID), "remoteAdmin": true})
}
func (s *Server) remoteLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	if !s.cfg.PublicWeb {
		legacyError(w, 404, "Acceso remoto desactivado.", nil)
		return
	}
	var req struct {
		Identity       string `json:"identity"`
		Password       string `json:"password"`
		DeviceID       string `json:"deviceId"`
		AuthOnly       bool   `json:"authOnly"`
		RememberDevice *bool  `json:"rememberDevice"`
	}
	if err := readJSON(r, &req); err != nil {
		legacyError(w, 400, "Solicitud inválida.", nil)
		return
	}
	identity := strings.ToLower(strings.TrimSpace(req.Identity))
	key := requestIP(r) + "|" + identity
	if ok, wait := s.loginAllowed(key); !ok {
		w.Header().Set("Retry-After", fmt.Sprint(int(wait.Seconds())))
		legacyError(w, 429, "Demasiados intentos. Espera unos minutos.", nil)
		return
	}
	loginStarted := time.Now()
	authUser, err := s.remoteAuthByIdentity(identity)
	loadElapsed := time.Since(loginStarted)
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	authStarted := time.Now()
	if authUser.AmbiguousID || !authUser.Active || !verifyRemotePasswordRecord(req.Password, authUser.Password) {
		s.loginFailed(key)
		time.Sleep(250 * time.Millisecond)
		legacyError(w, 401, "Usuario o contraseña incorrectos.", nil)
		return
	}

	s.mu.Lock()
	state, loadErr := s.loadLegacyLoginState()
	s.mu.Unlock()
	if loadErr != nil {
		legacyError(w, 503, "No se pudieron comprobar los permisos. Intenta nuevamente.", nil)
		return
	}
	if !portalAllowed593(legacyMap(state.Database), legacyUser(state.Database, authUser.UserID), musicRequest553(r)) {
		legacyError(w, 403, "Tu cuenta no tiene acceso a este portal. Usa el enlace del ministerio asignado.", nil)
		return
	}
	s.loginSucceeded(key)
	authElapsed := time.Since(authStarted)
	remember := true // backward compatible with clients prior to 5.0.9
	if req.RememberDevice != nil {
		remember = *req.RememberDevice
	}
	ttl := 12 * time.Hour
	if remember {
		ttl = 30 * 24 * time.Hour
	}
	raw := "rmt_" + token(32)
	sess := remoteSession{TokenHash: hashToken(raw), UserID: authUser.UserID, Role: authUser.Role, DeviceID: strings.TrimSpace(req.DeviceID), IP: requestIP(r), CreatedAt: time.Now().UTC().Format(time.RFC3339), LastSeen: time.Now().UTC().Format(time.RFC3339), ExpiresAt: time.Now().Add(ttl).UTC().Format(time.RFC3339)}
	if musicRequest553(r) {
		sess.Product = "music"
	}
	s.remoteMu.Lock()
	s.remoteSessions[sess.TokenHash] = sess
	s.saveRemoteSessionsLocked()
	s.remoteMu.Unlock()
	control := s.currentControlState()

	// 5.0.9 clients ask only for authentication here. Heavy authorized data is
	// retrieved afterwards from /api/remote/bootstrap, outside the critical login path.
	if req.AuthOnly {
		w.Header().Set("Server-Timing", fmt.Sprintf("sgl_load;dur=%.1f, sgl_auth;dur=%.1f, sgl_auth_index;dur=%.1f, sgl_password;dur=%.1f", float64(loadElapsed.Microseconds())/1000, float64(authElapsed.Microseconds())/1000, float64(loadElapsed.Microseconds())/1000, float64(authElapsed.Microseconds())/1000))
		jsonOut(w, 200, map[string]any{
			"ok": true, "sessionToken": raw, "userId": sess.UserID, "role": sess.Role,
			"fullName": authUser.FullName, "expiresAt": sess.ExpiresAt, "rememberDevice": remember,
			"serverId": s.store.ServerID, "serverCode": publicServerCode(s.store.ServerID), "serverName": s.cfg.ServerName, "version": release.Version,
			"controlRevision": control.Revision,
		})
		return
	}

	// Compatibility path for 5.0.7 and older clients that still expect the
	// authorized snapshot in the login response.
	snapshotStarted := time.Now()
	s.mu.Lock()
	loginState, stateErr := s.loadLegacyLoginState()
	s.mu.Unlock()
	if stateErr != nil {
		legacyError(w, 500, stateErr.Error(), nil)
		return
	}
	snap := s.filteredLegacySnapshotOwnedMode(loginState, sess.UserID, sess.Role, false)
	snapshotElapsed := time.Since(snapshotStarted)
	snap["sessionToken"] = raw
	snap["userId"] = sess.UserID
	snap["role"] = sess.Role
	snap["expiresAt"] = sess.ExpiresAt
	snap["serverId"] = s.store.ServerID
	snap["serverCode"] = publicServerCode(s.store.ServerID)
	snap["serverName"] = s.cfg.ServerName
	snap["controlRevision"] = control.Revision
	if control.Revision > 0 {
		snap["control"] = control
	}
	w.Header().Set("Server-Timing", fmt.Sprintf("sgl_load;dur=%.1f, sgl_auth;dur=%.1f, sgl_auth_index;dur=%.1f, sgl_password;dur=%.1f, sgl_snapshot;dur=%.1f", float64(loadElapsed.Microseconds())/1000, float64(authElapsed.Microseconds())/1000, float64(loadElapsed.Microseconds())/1000, float64(authElapsed.Microseconds())/1000, float64(snapshotElapsed.Microseconds())/1000))
	jsonOutRemote(w, r, 200, snap)
}

func (s *Server) remoteSessionInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := s.remoteSessionFromRequest(r)
	if !ok {
		legacyError(w, http.StatusUnauthorized, "Tu sesión venció. Inicia sesión nuevamente.", nil)
		return
	}
	control := s.currentControlState()
	jsonOut(w, http.StatusOK, map[string]any{
		"ok": true, "userId": sess.UserID, "role": sess.Role, "expiresAt": sess.ExpiresAt, "product": sess.Product,
		"serverId": s.store.ServerID, "serverCode": publicServerCode(s.store.ServerID), "serverName": s.cfg.ServerName, "version": release.Version,
		"controlRevision": control.Revision,
	})
}

func (s *Server) remoteBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := s.remoteSessionFromRequest(r)
	if !ok {
		legacyError(w, http.StatusUnauthorized, "Tu sesión venció. Inicia sesión nuevamente.", nil)
		return
	}
	started := time.Now()
	s.mu.Lock()
	state, err := s.loadLegacyLoginState()
	s.mu.Unlock()
	if err != nil {
		legacyError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	snap := s.filteredLegacySnapshotOwnedMode(state, sess.UserID, sess.Role, false)
	// 5.0.9: un navegador remoto nunca necesita hashes de contraseña ni llaves
	// de recuperación para funcionar. Se conservan exclusivamente en el servidor.
	stripRemoteCredentialFields508(snap)
	control := s.currentControlState()
	snap["userId"] = sess.UserID
	snap["role"] = sess.Role
	snap["expiresAt"] = sess.ExpiresAt
	snap["serverId"] = s.store.ServerID
	snap["serverCode"] = publicServerCode(s.store.ServerID)
	snap["serverName"] = s.cfg.ServerName
	snap["controlRevision"] = control.Revision
	if control.Revision > 0 {
		snap["control"] = control
	}
	w.Header().Set("Server-Timing", fmt.Sprintf("sgl_bootstrap;dur=%.1f", float64(time.Since(started).Microseconds())/1000))
	jsonOutRemote(w, r, http.StatusOK, snap)
}

func (s *Server) remoteHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := s.remoteHeartbeatSessionFromRequest(r)
	if !ok {
		legacyError(w, http.StatusUnauthorized, "Sesión remota no válida.", nil)
		return
	}
	control := s.currentControlState()
	jsonOut(w, http.StatusOK, map[string]any{"ok": true, "userId": sess.UserID, "serverTime": time.Now().UTC().Format(time.RFC3339Nano), "control": control})
}

func (s *Server) loadLegacyFilesSelective(manifestOnly bool, requestedID string) ([]any, int64, error) {
	if s.legacyCacheLoaded {
		out := make([]any, 0)
		for _, raw := range s.legacyCache.State.Files {
			m := legacyMap(raw)
			id := strings.TrimSpace(fmt.Sprint(m["id"]))
			if requestedID != "" && id != requestedID {
				continue
			}
			if manifestOnly {
				item := make(map[string]any, len(m))
				for k, value := range m {
					if k != "dataUrl" {
						item[k] = value
					}
				}
				out = append(out, item)
			} else {
				out = append(out, raw)
			}
			if requestedID != "" {
				break
			}
		}
		return out, s.legacyCache.State.ServerVersion, nil
	}
	f, err := os.Open(s.legacyPath())
	if os.IsNotExist(err) {
		return []any{}, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	tok, err := dec.Token()
	if err != nil {
		return nil, 0, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, 0, fmt.Errorf("almacen SGL invalido")
	}
	var version int64
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, version, err
		}
		key, _ := keyTok.(string)
		if key != "state" {
			var skip json.RawMessage
			if err := dec.Decode(&skip); err != nil {
				return nil, version, err
			}
			continue
		}
		tok, err = dec.Token()
		if err != nil {
			return nil, version, err
		}
		if d, ok := tok.(json.Delim); !ok || d != '{' {
			return nil, version, fmt.Errorf("estado SGL invalido")
		}
		for dec.More() {
			fieldTok, err := dec.Token()
			if err != nil {
				return nil, version, err
			}
			field, _ := fieldTok.(string)
			switch field {
			case "serverVersion":
				if err := dec.Decode(&version); err != nil {
					return nil, version, err
				}
			case "files":
				tok, err = dec.Token()
				if err != nil {
					return nil, version, err
				}
				if d, ok := tok.(json.Delim); !ok || d != '[' {
					return nil, version, fmt.Errorf("archivos SGL invalidos")
				}
				out := make([]any, 0)
				for dec.More() {
					var raw map[string]json.RawMessage
					if err := dec.Decode(&raw); err != nil {
						return nil, version, err
					}
					var id string
					if v, ok := raw["id"]; ok {
						_ = json.Unmarshal(v, &id)
					}
					if requestedID != "" && id != requestedID {
						continue
					}
					item := map[string]any{}
					for k, v := range raw {
						if manifestOnly && k == "dataUrl" {
							continue
						}
						var value any
						if err := json.Unmarshal(v, &value); err != nil {
							continue
						}
						item[k] = value
					}
					out = append(out, item)
					if requestedID != "" {
						for dec.More() {
							var skip json.RawMessage
							if err := dec.Decode(&skip); err != nil {
								return nil, version, err
							}
						}
					}
				}
				if _, err := dec.Token(); err != nil {
					return nil, version, err
				}
				return out, version, nil
			default:
				var skip json.RawMessage
				if err := dec.Decode(&skip); err != nil {
					return nil, version, err
				}
			}
		}
	}
	return []any{}, version, nil
}

func remoteFilePolicy(database any, userID, role string) (map[string]bool, map[string]bool) {
	owners := map[string]bool{userID: true}
	materials := map[string]bool{}
	if viewAll548(role) {
		owners["*"] = true
		return owners, materials
	}
	db := legacyMap(database)
	if db == nil {
		return owners, materials
	}
	if role == "tutor" {
		for _, raw := range legacyUsers(db) {
			u := legacyMap(raw)
			if leader555(db, u) && legacyTutorAssigned(u, userID, "") {
				owners[fmt.Sprint(u["id"])] = true
			}
		}
	}
	if mats, ok := db["materials"].([]any); ok {
		for _, raw := range mats {
			m := legacyMap(raw)
			if id := strings.TrimSpace(fmt.Sprint(m["fileId"])); id != "" && id != "<nil>" {
				materials[id] = true
			}
		}
	}
	// SGL 5.22: Bible versions are library resources configured by the
	// Administrator. They are readable by authenticated users just like support
	// materials, while upload/delete ownership remains with the Administrator.
	if versions, ok := db["bibleVersions"].([]any); ok {
		for _, raw := range versions {
			version := legacyMap(raw)
			if id := strings.TrimSpace(fmt.Sprint(version["fileId"])); id != "" && id != "<nil>" {
				materials[id] = true
			}
		}
	}
	return owners, materials
}

func remoteFileAllowed(item map[string]any, owners, materials map[string]bool) bool {
	if owners["*"] {
		return true
	}
	id := fmt.Sprint(item["id"])
	owner := fmt.Sprint(item["ownerId"])
	return owners[owner] || materials[id]
}

func (s *Server) remoteFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := s.remoteSessionFromRequest(r)
	if !ok {
		legacyError(w, http.StatusUnauthorized, "Sesion remota no valida.", nil)
		return
	}
	requestedID := strings.TrimSpace(r.URL.Query().Get("fileId"))
	manifestOnly := strings.TrimSpace(r.URL.Query().Get("manifest")) == "1"

	// 5.0.48: manifest and single-file reads never deserialize the complete
	// evidence collection. This keeps login and mobile photo reads independent
	// from the total number of photos already stored on the server.
	if manifestOnly || requestedID != "" {
		s.mu.Lock()
		state, stateErr := s.loadLegacyLoginState()
		files, version, filesErr := s.loadLegacyFilesSelective(manifestOnly, requestedID)
		s.mu.Unlock()
		if stateErr != nil {
			legacyError(w, http.StatusInternalServerError, stateErr.Error(), nil)
			return
		}
		if filesErr != nil {
			legacyError(w, http.StatusInternalServerError, filesErr.Error(), nil)
			return
		}
		owners, materials := remoteFilePolicy(state.Database, sess.UserID, sess.Role)
		visible := make([]any, 0, len(files))
		for _, raw := range files {
			item := legacyMap(raw)
			if item != nil && remoteFileAllowed(item, owners, materials) {
				if sess.Role != "aspirant" && sess.Role != "admin" && !canReadFile555(legacyMap(state.Database), legacyUser(state.Database, sess.UserID), item) {
					continue
				}
				if sess.Role == "tutor" && stringValue(item["ownerId"]) != sess.UserID && !materials[stringValue(item["id"])] && !legacyTutorAssigned(legacyUser(state.Database, stringValue(item["ownerId"])), sess.UserID, remoteFileProgram(item)) {
					continue
				}
				visible = append(visible, raw)
			}
		}
		if requestedID != "" && len(visible) == 0 {
			legacyError(w, http.StatusNotFound, "Archivo no encontrado o no autorizado.", nil)
			return
		}
		jsonOutRemote(w, r, http.StatusOK, map[string]any{"ok": true, "serverVersion": version, "files": visible, "fileId": requestedID, "manifest": manifestOnly})
		return
	}

	// Backward-compatible full download for older clients only.
	s.mu.Lock()
	st, err := s.loadLegacy()
	s.mu.Unlock()
	if err != nil {
		legacyError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	full := s.filteredLegacySnapshotMode(st.State, sess.UserID, sess.Role, true)
	files, _ := full["files"].([]any)
	jsonOutRemote(w, r, http.StatusOK, map[string]any{"ok": true, "serverVersion": st.State.ServerVersion, "files": files, "fileId": "", "manifest": false})
}

func (s *Server) remotePresence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	allowed := false
	viewerID, viewerRole := "", "admin"
	if sess, remote := s.remoteSessionFromRequest(r); remote && (sess.Role == "admin" || sess.Role == "tutor") {
		allowed = true
		viewerID, viewerRole = sess.UserID, sess.Role
	}
	if !allowed {
		role, ok := s.auth(r)
		allowed = ok && role == "admin"
	}
	if !allowed {
		legacyError(w, http.StatusForbidden, "Esta ruta requiere Administrador o Tutor conectado.", nil)
		return
	}
	now := time.Now()
	// Snapshot de sesiones sin mantener el mutex mientras se lee la base principal.
	// 95 segundos evita falsos desconectados en celulares o pestañas suspendidas.
	s.remoteMu.Lock()
	sessions := make([]remoteSession, 0, len(s.remoteSessions))
	for _, sess := range s.remoteSessions {
		if musicRequest553(r) && sess.Product != "music" {
			continue
		}
		if sess.HeartbeatAt == "" {
			continue
		}
		last, err := time.Parse(time.RFC3339, sess.HeartbeatAt)
		if err == nil && now.Sub(last) <= 95*time.Second {
			sessions = append(sessions, sess)
		}
	}
	s.remoteMu.Unlock()

	s.mu.Lock()
	st, err := s.loadLegacy()
	s.mu.Unlock()
	if err != nil {
		legacyError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	// Agrupa por usuario: si está conectado desde dos dispositivos mostramos una sola persona,
	// conservando la sesión más reciente y el número de dispositivos activos.
	type presence struct {
		UserID      string `json:"userId"`
		FullName    string `json:"fullName"`
		Role        string `json:"role"`
		LastSeen    string `json:"lastSeen"`
		DeviceID    string `json:"deviceId,omitempty"`
		DeviceCount int    `json:"deviceCount"`
	}
	grouped := map[string]presence{}
	for _, sess := range sessions {
		program := ""
		if musicRequest553(r) {
			program = musicProgram553
		}
		if viewerRole == "tutor" && !legacyTutorAssigned(legacyUser(st.State.Database, sess.UserID), viewerID, program) {
			continue
		}
		item := grouped[sess.UserID]
		item.UserID = sess.UserID
		item.Role = sess.Role
		item.DeviceCount++
		if item.LastSeen == "" || sess.HeartbeatAt > item.LastSeen {
			item.LastSeen = sess.HeartbeatAt
			item.DeviceID = sess.DeviceID
		}
		if u := legacyUser(st.State.Database, sess.UserID); u != nil {
			item.FullName = legacyName(u)
		}
		if item.FullName == "" {
			item.FullName = sess.UserID
		}
		grouped[sess.UserID] = item
	}
	users := make([]presence, 0, len(grouped))
	for _, item := range grouped {
		users = append(users, item)
	}
	jsonOut(w, http.StatusOK, map[string]any{"ok": true, "count": len(users), "users": users, "serverTime": now.UTC().Format(time.RFC3339Nano)})
}

func (s *Server) remoteLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		key := hashToken(strings.TrimSpace(h[7:]))
		s.remoteMu.Lock()
		delete(s.remoteSessions, key)
		s.saveRemoteSessionsLocked()
		s.remoteMu.Unlock()
	}
	w.WriteHeader(204)
}

func sanitizedUser(u map[string]any, full bool) map[string]any {
	out := legacyMap(cloneAny(u))
	if !full {
		for _, k := range []string{"password", "recoveryKeys", "drive", "email", "phone", "birthDate"} {
			delete(out, k)
		}
	}
	return out
}
func stripRemoteCredentialFields508(payload map[string]any) {
	db := legacyMap(payload["database"])
	if db == nil {
		return
	}
	users := legacyUsers(db)
	clean := make([]any, 0, len(users))
	for _, raw := range users {
		u := legacyMap(cloneAny(raw))
		if u == nil {
			continue
		}
		delete(u, "password")
		delete(u, "recoveryKeys")
		clean = append(clean, u)
	}
	db["users"] = clean
	// 5.0.9: el catálogo remoto conserva solo metadata. El archivo real vive en
	// state.files y se entrega bajo demanda por /api/remote/files?fileId=...
	if materials, ok := db["materials"].([]any); ok {
		light := make([]any, 0, len(materials))
		for _, raw := range materials {
			item := legacyMap(cloneAny(raw))
			if item == nil {
				continue
			}
			delete(item, "dataUrl")
			light = append(light, item)
		}
		db["materials"] = light
	}
	payload["database"] = db
}
func anyIDSet(values []any) map[string]bool {
	out := map[string]bool{}
	for _, v := range values {
		m := legacyMap(v)
		if id := fmt.Sprint(m["id"]); id != "" {
			out[id] = true
		}
	}
	return out
}

func (s *Server) filteredLegacySnapshotMode(st legacyState, userID, role string, includeFiles bool) map[string]any {
	return s.filteredLegacySnapshotModeInternal(st, userID, role, includeFiles, true)
}

func (s *Server) filteredLegacySnapshotOwnedMode(st legacyState, userID, role string, includeFiles bool) map[string]any {
	return s.filteredLegacySnapshotModeInternal(st, userID, role, includeFiles, false)
}

func (s *Server) filteredLegacySnapshotModeInternal(st legacyState, userID, role string, includeFiles bool, cloneDatabase bool) map[string]any {
	role = canonicalRole548(role)
	var source any = st.Database
	if cloneDatabase || ((viewAll548(role) || role == "regional") && role != "admin") {
		source = cloneAny(st.Database)
	}
	db := legacyMap(source)
	if db == nil {
		return legacySnapshot(st)
	}
	legacyEnforceSGL427Rules(db)
	db = legacyMap(legacyStampRestoreMetadata(db, st))
	db["syncCapabilities"] = map[string]any{"assignmentDeltaVersion": 1}
	db = musicScopeSnapshot560(db, userID, role)
	if db["music550"] != nil {
		musicActor := legacyUser(db, userID)
		if role == "aspirant" && legacyRole(musicActor) != "aspirant" {
			musicActor = legacyMap(cloneAny(musicActor))
			musicActor["role"] = "aspirant"
		}
		db["music550"] = musicView550(db, musicActor)
	}
	if viewAll548(role) || role == "regional" {
		if role != "admin" {
			db = scopeProject555(db, userID, role)
		}
		if settings := legacyMap(db["settings"]); settings != nil {
			delete(settings, "serverToken")
			delete(settings, "serverUrl")
			db["settings"] = settings
		}
		if role != "admin" {
			sanitizeReadAll548(db, userID)
		}
		payload := map[string]any{
			"ok": true, "initialized": st.Database != nil, "serverVersion": st.ServerVersion,
			"restoreGeneration": legacyCanonicalRestoreGeneration(st), "restoreId": st.RestoreID, "restoredAt": st.RestoredAt,
			"database": db, "updatedAt": st.UpdatedAt,
		}
		if includeFiles {
			files := st.Files
			if cloned, ok := cloneAny(st.Files).([]any); ok {
				files = cloned
			}
			if role != "admin" {
				visible := []any{}
				actor := legacyUser(db, userID)
				for _, raw := range files {
					if canReadFile555(db, actor, legacyMap(raw)) {
						visible = append(visible, raw)
					}
				}
				files = visible
			}
			payload["files"] = files
		} else {
			payload["filesDeferred"] = true
		}
		return payload
	}
	users := legacyUsers(db)
	allowed := map[string]bool{userID: true}
	actor := legacyUser(db, userID)
	tutorID := legacyTutorForProgram(actor, "class_grouped")
	if role == "aspirant" {
		for _, p := range []string{"class_grouped", "guide_major", "guide_master", "guide_master_advanced", "music_ministry"} {
			if id := legacyTutorForProgram(actor, p); id != "" {
				allowed[id] = true
			}
		}
	}
	if role == "tutor" {
		for _, raw := range users {
			u := legacyMap(raw)
			if leader555(db, u) && legacyTutorAssigned(u, userID, "") {
				allowed[fmt.Sprint(u["id"])] = true
			}
		}
	}
	if role == "aspirant" && tutorID != "" && tutorID != "<nil>" {
		allowed[tutorID] = true
	}
	filteredUsers := []any{}
	for _, raw := range users {
		u := legacyMap(raw)
		id := fmt.Sprint(u["id"])
		r := legacyRole(u)
		if allowed[id] {
			filteredUsers = append(filteredUsers, sanitizedUser(u, id == userID))
		} else if r == "admin" {
			filteredUsers = append(filteredUsers, map[string]any{"id": id, "fullName": u["fullName"], "role": "admin", "active": legacyActive(u), "accountState": "active"})
		}
	}
	db["users"] = filteredUsers
	if settings := legacyMap(db["settings"]); settings != nil {
		delete(settings, "serverToken")
		delete(settings, "serverUrl")
		delete(settings, "nameAutocompleteHistory")
		db["settings"] = settings
	}
	clubIDs := map[string]bool{}
	for _, raw := range filteredUsers {
		u := legacyMap(raw)
		cid := fmt.Sprint(u["clubId"])
		if cid != "" && cid != "<nil>" {
			clubIDs[cid] = true
		}
	}
	if clubs, ok := db["clubs"].([]any); ok {
		out := []any{}
		for _, raw := range clubs {
			m := legacyMap(raw)
			if clubIDs[fmt.Sprint(m["id"])] {
				out = append(out, raw)
			}
		}
		db["clubs"] = out
	}
	db["errors"] = []any{}
	dataIDs := map[string]bool{userID: true}
	if role == "tutor" {
		for id := range allowed {
			if id != userID {
				dataIDs[id] = true
			}
		}
	}
	filterMapByIDs := func(key string) {
		m := legacyMap(db[key])
		if m == nil {
			return
		}
		out := map[string]any{}
		for id, v := range m {
			if dataIDs[id] {
				out[id] = v
			}
		}
		db[key] = out
	}
	for _, k := range []string{"progress", "programProgress", "bibleYear"} {
		filterMapByIDs(k)
	}
	if role == "tutor" {
		for id := range dataIDs {
			if id == userID {
				continue
			}
			owner := legacyUser(db, id)
			if !legacyTutorAssigned(owner, userID, "class_grouped") {
				delete(legacyMap(db["progress"]), id)
			}
			programs := legacyMap(legacyMap(db["programProgress"])[id])
			for programID := range programs {
				if !legacyTutorAssigned(owner, userID, programID) {
					delete(programs, programID)
				}
			}
		}
	}

	if cfg := legacyMap(db["configs"]); cfg != nil {
		out := map[string]any{}
		for k, v := range cfg {
			for id := range dataIDs {
				if k == id || strings.Contains(k, ":"+id+":") || strings.HasPrefix(k, id+":") {
					out[k] = v
					break
				}
			}
		}
		db["configs"] = out
	}
	if classes, ok := db["classes"].([]any); ok {
		out := []any{}
		for _, raw := range classes {
			c := legacyMap(raw)
			ids, _ := c["assignedUserIds"].([]any)
			keep := false
			for _, v := range ids {
				if allowed[fmt.Sprint(v)] {
					keep = true
				}
			}
			if keep {
				c2 := legacyMap(cloneAny(c))
				arr := []any{}
				for _, v := range ids {
					if allowed[fmt.Sprint(v)] {
						arr = append(arr, v)
					}
				}
				c2["assignedUserIds"] = arr
				out = append(out, c2)
			}
		}
		db["classes"] = out
	}
	if mods, ok := db["modules"].([]any); ok {
		out := []any{}
		for _, raw := range mods {
			m := legacyMap(cloneAny(raw))
			if ids, ok := m["assignedUserIds"].([]any); ok {
				arr := []any{}
				for _, v := range ids {
					if allowed[fmt.Sprint(v)] {
						arr = append(arr, v)
					}
				}
				m["assignedUserIds"] = arr
			}
			out = append(out, m)
		}
		db["modules"] = out
	}
	if tasks, ok := db["tasks"].([]any); ok {
		out := []any{}
		for _, raw := range tasks {
			m := legacyMap(raw)
			if role == "tutor" && !legacyTutorAssigned(legacyUser(db, stringValue(m["assigneeId"])), userID, stringValue(m["programId"])) {
				continue
			}
			if dataIDs[fmt.Sprint(m["assigneeId"])] || fmt.Sprint(m["createdBy"]) == userID {
				out = append(out, raw)
			}
		}
		db["tasks"] = out
	}
	if mats, ok := db["materials"].([]any); ok {
		out := []any{}
		for _, raw := range mats {
			m := legacyMap(raw)
			creatorID := fmt.Sprint(m["createdBy"])
			creator := legacyUser(db, creatorID)
			cr := legacyRole(creator)
			publisherRole := strings.ToLower(strings.TrimSpace(fmt.Sprint(m["publisherRole"])))
			visibleToUsers, _ := m["visibleToUsers"].(bool)
			adminPublished := cr == "admin" || publisherRole == "admin" || visibleToUsers
			audienceTutorID := strings.TrimSpace(fmt.Sprint(m["audienceTutorId"]))
			if audienceTutorID == "<nil>" {
				audienceTutorID = ""
			}
			tutorPublishedToActor := (role == "tutor" && (creatorID == userID || audienceTutorID == userID)) ||
				(role == "aspirant" && (legacyTutorAssigned(actor, creatorID, stringValue(m["programId"])) || legacyTutorAssigned(actor, audienceTutorID, stringValue(m["programId"]))))
			if adminPublished || creatorID == userID || tutorPublishedToActor {
				out = append(out, raw)
			}
		}
		db["materials"] = out
	}
	if audit, ok := db["audit"].([]any); ok {
		out := []any{}
		for _, raw := range audit {
			m := legacyMap(raw)
			if fmt.Sprint(m["actorId"]) == userID {
				out = append(out, raw)
			}
		}
		db["audit"] = out
	}
	db["conflicts"] = []any{}
	db["syncRequestsCache"] = []any{}
	db["syncNotifications"] = []any{}
	payload := map[string]any{
		"ok": true, "initialized": true, "serverVersion": st.ServerVersion,
		"restoreGeneration": legacyCanonicalRestoreGeneration(st), "restoreId": st.RestoreID, "restoredAt": st.RestoredAt,
		"database": db, "updatedAt": st.UpdatedAt,
	}
	if includeFiles {
		filteredFiles := []any{}
		materialIDs := map[string]bool{}
		if mats, ok := db["materials"].([]any); ok {
			for _, raw := range mats {
				m := legacyMap(raw)
				if id := fmt.Sprint(m["fileId"]); id != "" && id != "<nil>" {
					materialIDs[id] = true
				}
			}
		}
		if versions, ok := db["bibleVersions"].([]any); ok {
			for _, raw := range versions {
				version := legacyMap(raw)
				if id := fmt.Sprint(version["fileId"]); id != "" && id != "<nil>" {
					materialIDs[id] = true
				}
			}
		}
		for _, raw := range st.Files {
			m := legacyMap(raw)
			if dataIDs[fmt.Sprint(m["ownerId"])] || materialIDs[fmt.Sprint(m["id"])] {
				if role == "tutor" && stringValue(m["ownerId"]) != userID && !materialIDs[stringValue(m["id"])] && !legacyTutorAssigned(legacyUser(db, stringValue(m["ownerId"])), userID, remoteFileProgram(m)) {
					continue
				}
				filteredFiles = append(filteredFiles, raw)
			}
		}
		payload["files"] = filteredFiles
	} else {
		payload["filesDeferred"] = true
	}
	return payload
}

func (s *Server) filteredLegacySnapshot(st legacyState, userID, role string) map[string]any {
	return s.filteredLegacySnapshotMode(st, userID, role, true)
}
