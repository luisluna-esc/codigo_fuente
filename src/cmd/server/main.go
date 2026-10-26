package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sgl.local/sistema-gestion-liderazgo/internal/processutil"
	"sgl.local/sistema-gestion-liderazgo/internal/release"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const systemName = "Ministerio de la Música Online"

type Config struct {
	BindAddress    string   `json:"bindAddress"`
	Port           int      `json:"port"`
	AdvertisedURL  string   `json:"advertisedUrl"`
	AdminToken     string   `json:"adminToken"`
	ClientToken    string   `json:"clientToken"`
	DataFile       string   `json:"dataFile"`
	ServerName     string   `json:"serverName"`
	WebRoot        string   `json:"webRoot,omitempty"`
	PublicWeb      bool     `json:"publicWeb,omitempty"`
	AllowedOrigins []string `json:"allowedOrigins,omitempty"`
}
type Entity struct {
	EntityKey string `json:"entityKey"`
	Revision  int64  `json:"revision"`
	Payload   any    `json:"payload"`
	UpdatedAt string `json:"updatedAt"`
	UpdatedBy string `json:"updatedBy"`
	DeviceID  string `json:"deviceId"`
}
type Pending struct {
	ID              string `json:"id"`
	ChangeID        string `json:"changeId"`
	EntityKey       string `json:"entityKey"`
	BaseRevision    int64  `json:"baseRevision"`
	CurrentRevision int64  `json:"currentRevision"`
	Payload         any    `json:"payload"`
	UserID          string `json:"userId"`
	Role            string `json:"role"`
	DeviceID        string `json:"deviceId"`
	Kind            string `json:"kind"`
	Status          string `json:"status"`
	Message         string `json:"message"`
	CreatedAt       string `json:"createdAt"`
	DecidedAt       string `json:"decidedAt,omitempty"`
	Decision        string `json:"decision,omitempty"`
	DecisionMessage string `json:"decisionMessage,omitempty"`
}
type Audit struct {
	At        string `json:"at"`
	Action    string `json:"action"`
	EntityKey string `json:"entityKey"`
	UserID    string `json:"userId"`
	DeviceID  string `json:"deviceId"`
	Details   string `json:"details"`
}
type Store struct {
	SchemaVersion int                `json:"schemaVersion"`
	ServerID      string             `json:"serverId"`
	Revision      int64              `json:"revision"`
	Entities      map[string]Entity  `json:"entities"`
	Pending       map[string]Pending `json:"pending"`
	Audit         []Audit            `json:"audit"`
}
type Change struct {
	ID           string `json:"id"`
	EntityKey    string `json:"entityKey"`
	BaseRevision int64  `json:"baseRevision"`
	Payload      any    `json:"payload"`
	UserID       string `json:"userId"`
	Role         string `json:"role"`
	DeviceID     string `json:"deviceId"`
}
type SyncRequest struct {
	DeviceID     string   `json:"deviceId"`
	UserID       string   `json:"userId"`
	Role         string   `json:"role"`
	LastRevision int64    `json:"lastRevision"`
	Changes      []Change `json:"changes"`
}
type Applied struct {
	ChangeID  string `json:"changeId"`
	EntityKey string `json:"entityKey"`
	Revision  int64  `json:"revision"`
}
type SyncResponse struct {
	ServerID  string    `json:"serverId"`
	Revision  int64     `json:"revision"`
	Applied   []Applied `json:"applied"`
	Pending   []Pending `json:"pending"`
	Conflicts []Pending `json:"conflicts"`
	Updates   []Entity  `json:"updates"`
}
type Server struct {
	mu                         sync.Mutex
	remoteMu                   sync.Mutex
	loginMu                    sync.Mutex
	authMu                     sync.Mutex
	cfg                        Config
	store                      Store
	path                       string
	remoteSessions             map[string]remoteSession
	loginAttempts              map[string]loginAttempt
	authIndex                  map[string]remoteAuthUser
	authIndexStamp             string
	control                    controlState
	configPath                 string
	stop                       chan struct{}
	legacyCache                legacyStore
	legacyCacheLoaded          bool
	legacyStorageMode          string
	legacyPersistedFilesSig    string
	legacySQLiteFallbackLogged bool
	fontCatalog617             *musicFontCatalog617
}

func token(n int) string { b := make([]byte, n); _, _ = rand.Read(b); return hex.EncodeToString(b) }
func eq(a, b string) bool {
	if len(a) != len(b) || a == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
func main() {
	cfgPath := flag.String("config", "server-config.json", "archivo de configuración")
	extractPath := flag.String("extraer-musica", "", "extraer solo Música desde un respaldo JSON o cabackup")
	outputPath := flag.String("salida", "Musica_importar.json", "archivo de salida para la extracción")
	flag.Parse()
	if *extractPath != "" {
		if err := extractMusicCLI(*extractPath, *outputPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := loadIdentityConfig610(filepath.Dir(*cfgPath)); err != nil {
		log.Fatal(err)
	}
	cfg, created, err := loadConfig(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Dir(*cfgPath), 0755)
	if lf, e := os.OpenFile(filepath.Join(filepath.Dir(*cfgPath), "ServidorMusica.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); e == nil {
		defer lf.Close()
		log.SetOutput(lf)
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	}
	if created {
		log.Printf("Configuración creada en %s", *cfgPath)
	}
	lock := filepath.Join(filepath.Dir(*cfgPath), "musica-server.lock")
	if err = acquireLock(lock); err != nil {
		log.Fatalf("Ya existe un servidor activo o un bloqueo pendiente: %v. Si el proceso ya está detenido, consulta el manual para retirar musica-server.lock.", err)
	}
	defer os.Remove(lock)
	s, err := newServer(cfg, *cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := s.prepareInstallation(); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	s.routes(mux)
	s.webRoutes(mux)
	go s.musicFollowupLoop610()
	addr := net.JoinHostPort(cfg.BindAddress, strconv.Itoa(cfg.Port))
	httpSrv := &http.Server{Addr: addr, Handler: s.onlineProduct(s.securityHeaders(s.cors(mux))), ReadHeaderTimeout: 15 * time.Second, ReadTimeout: 2 * time.Minute, WriteTimeout: 2 * time.Minute, IdleTimeout: 2 * time.Minute}
	go func() {
		log.Printf("%s escuchando en http://%s (serverId=%s)", systemName, addr, s.store.ServerID)
		if e := httpSrv.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			log.Fatal(e)
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	select {
	case <-ch:
	case <-s.stop:
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
	s.mu.Lock()
	if e := s.flushLegacyCompatibilityJSON(); e != nil {
		log.Printf("SGL %s: no se pudo actualizar el espejo JSON de compatibilidad al cerrar: %v", release.Version, e)
	}
	s.mu.Unlock()
}
func loadConfig(path string) (Config, bool, error) {
	var c Config
	b, e := os.ReadFile(path)
	created := false
	if e != nil {
		if !os.IsNotExist(e) {
			return c, false, e
		}
		c = Config{BindAddress: "127.0.0.1", Port: 8796, AdminToken: token(24), ClientToken: token(24), DataFile: "data/musica-server-data.json", ServerName: "Ministerio de la Música", PublicWeb: true, WebRoot: "app"}
		created = true
		bb, _ := json.MarshalIndent(c, "", "  ")
		_ = os.MkdirAll(filepath.Dir(path), 0755)
		if e = os.WriteFile(path, bb, 0600); e != nil {
			return c, false, e
		}
	} else if e = json.Unmarshal(b, &c); e != nil {
		return c, false, e
	}
	if c.BindAddress == "" {
		c.BindAddress = "127.0.0.1"
	}
	if c.Port == 0 {
		c.Port = 8796
	}
	if c.DataFile == "" {
		c.DataFile = "data/musica-server-data.json"
	}
	if !filepath.IsAbs(c.DataFile) {
		c.DataFile = filepath.Join(filepath.Dir(path), c.DataFile)
	}
	if c.WebRoot != "" && !filepath.IsAbs(c.WebRoot) {
		c.WebRoot = filepath.Join(filepath.Dir(path), c.WebRoot)
	}
	return c, created, nil
}
func acquireLock(path string) error {
	if err := processutil.ClearStaleLock(path); err != nil {
		return err
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	fmt.Fprintf(f, "pid=%d\nstarted=%s\n", os.Getpid(), time.Now().Format(time.RFC3339))
	return f.Close()
}
func newServer(c Config, cfgPath string) (*Server, error) {
	s := &Server{cfg: c, path: c.DataFile, configPath: cfgPath, stop: make(chan struct{}, 1)}
	b, e := os.ReadFile(s.path)
	if e == nil {
		if e = json.Unmarshal(b, &s.store); e != nil {
			return nil, e
		}
	} else if !os.IsNotExist(e) {
		return nil, e
	}
	if s.store.ServerID == "" {
		s.store = Store{SchemaVersion: 4, ServerID: "srv_" + token(12), Entities: map[string]Entity{}, Pending: map[string]Pending{}, Audit: []Audit{}}
		if e = s.save(); e != nil {
			return nil, e
		}
	}
	if s.store.Entities == nil {
		s.store.Entities = map[string]Entity{}
	}
	if s.store.Pending == nil {
		s.store.Pending = map[string]Pending{}
	}
	s.loadRemoteSessions()
	s.loadControlState()
	// 5.18 loads the legacy synchronization state once at startup. On Windows
	// this also performs the one-time JSON -> SQLite migration when possible.
	if _, e = s.loadLegacy(); e != nil {
		return nil, fmt.Errorf("cargar persistencia de sincronización: %w", e)
	}

	return s, nil
}
func (s *Server) save() error {
	_ = os.MkdirAll(filepath.Dir(s.path), 0755)
	b, e := json.MarshalIndent(s.store, "", "  ")
	if e != nil {
		return e
	}
	tmp := s.path + ".tmp"
	if e = os.WriteFile(tmp, b, 0600); e != nil {
		return e
	}
	if runtime.GOOS == "windows" {
		_ = os.Remove(s.path)
	}
	return os.Rename(tmp, s.path)
}
func (s *Server) auth(r *http.Request) (role string, ok bool) {
	t := r.Header.Get("X-SGL-Token")
	if t == "" {
		if h := r.Header.Get("Authorization"); strings.HasPrefix(strings.ToLower(h), "bearer ") {
			t = strings.TrimSpace(h[7:])
		}
	}
	if eq(t, s.cfg.AdminToken) {
		if personalRequest555(r) {
			return "client", true
		}
		return "admin", true
	}
	if eq(t, s.cfg.ClientToken) {
		return "client", true
	}
	if sess, ok := s.remoteSessionFromRequest(r); ok {
		if sess.Role == "admin" {
			return "admin", true
		}
		return "client", true
	}
	return "", false
}
func (s *Server) originAllowed(r *http.Request, origin string) bool {
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if strings.EqualFold(u.Host, r.Host) {
		return true
	}
	host := strings.ToLower(u.Hostname())
	if host == "127.0.0.1" || host == "localhost" {
		return true
	}
	for _, allowed := range s.cfg.AllowedOrigins {
		if strings.EqualFold(strings.TrimRight(allowed, "/"), strings.TrimRight(origin, "/")) {
			return true
		}
	}
	return false
}
func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o := r.Header.Get("Origin")
		if o != "" && s.originAllowed(r, o) {
			w.Header().Set("Access-Control-Allow-Origin", o)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-SGL-Token, X-SGL-Personal, X-SGL-Actor")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == "OPTIONS" {
			if o != "" && !s.originAllowed(r, o) {
				http.Error(w, "origin no permitido", http.StatusForbidden)
				return
			}
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return s.personalAccess555(s.musicPortal553(s.accessGuard571(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(self), microphone=(self), on-device-speech-recognition=(self), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: blob: https:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; worker-src 'self'; connect-src 'self' http://127.0.0.1:* https:; font-src 'self' data:; frame-src 'self' blob: data: https://www.youtube.com https://youtube.com https://player.vimeo.com; object-src 'self' blob: data:; base-uri 'self'; frame-ancestors 'none'")
		if sess, ok := s.remoteSessionFromRequest(r); ok && (sess.Role == "departamental" || sess.Role == "administrativo" || sess.Role == "secretaria") {
			path := r.URL.Path
			allowedWrite := path == "/api/music/v610" || path == "/api/access571" || path == "/api/progress/reset" || path == "/api/accounts/tutor" || (path == "/api/reviews/decision" && programRequest556(r) == musicProgram553) || path == "/api/notifications/read" || path == "/api/notifications/delete" || path == "/api/activity/track" || path == "/api/music/module" || (path == "/api/music-registration" && (sess.Role == "departamental" || sess.Role == "secretaria")) || path == "/api/remote/logout" || path == "/api/remote/heartbeat" || path == "/api/remote/login" || (path == "/api/indications" && sess.Role == "departamental")
			if programRequest556(r) == musicProgram553 && (path == "/api/accounts/music" || path == "/api/accounts/unassigned" || path == "/api/indications" && musicPolicies560[sess.Role].Instructions) {
				allowedWrite = true
			}
			if (r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions && !allowedWrite) || path == "/api/google/start" || path == "/api/google/callback" {
				legacyError(w, http.StatusForbidden, "Tu rol solo puede consultar. Usa las opciones de edición habilitadas para tu función.", nil)
				return
			}
		}
		next.ServeHTTP(w, r)
	}))))
}
func (s *Server) webRoutes(m *http.ServeMux) {
	s.musicWebRoutes553(m)
	m.HandleFunc("/api/music/v610", s.music610)
	m.HandleFunc("/api/music/module", s.musicGate602(s.music550))
	m.HandleFunc("/api/music/contact", s.musicGate602(s.musicContact598))
	m.HandleFunc("/api/accounts/roster", s.accountRoster556)
	m.HandleFunc("/api/access571", s.access571)
	m.HandleFunc("/api/accounts/tutor", s.musicGate602(s.createTutor556))
	m.HandleFunc("/api/accounts/view", s.accountView560)
	m.HandleFunc("/api/accounts/music", s.musicGate602(s.musicAccounts560))
	m.HandleFunc("/api/accounts/unassigned", s.musicUnassigned561)
	m.HandleFunc("/api/accounts/import", s.musicGate602(s.accountsImport561))
	m.HandleFunc("/api/accounts/directory", s.accountDirectory561)
	m.HandleFunc("/api/music/certificate", s.musicCertificate550)
	m.HandleFunc("/api/music-registration", s.ministryResource548)
	m.HandleFunc("/api/report-watermark", s.ministryResource548)
	m.HandleFunc("/api/remote/info", s.remoteInfo)
	m.HandleFunc("/api/remote/login", s.remoteLogin)
	m.HandleFunc("/api/remote/session", s.remoteSessionInfo)
	m.HandleFunc("/api/remote/bootstrap", s.remoteBootstrap)
	m.HandleFunc("/api/remote/logout", s.remoteLogout)
	m.HandleFunc("/api/remote/heartbeat", s.remoteHeartbeat)
	m.HandleFunc("/api/remote/files", s.remoteFiles)
	m.HandleFunc("/api/remote/presence", s.remotePresence)
	m.HandleFunc("/api/admin/force-relogin", s.adminForceRelogin)
	m.HandleFunc("/api/admin/control", s.adminControlState)
	if !s.cfg.PublicWeb || s.cfg.WebRoot == "" {
		return
	}
	fs := compressedStatic(s.cfg.WebRoot, http.FileServer(http.Dir(s.cfg.WebRoot)))
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-store, max-age=0")
		if r.URL.Path == "/sw.js" {
			w.Header().Set("Service-Worker-Allowed", "/")
		}
		if r.URL.Path == "/" {
			fs.ServeHTTP(w, r)
			return
		}
		clean := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
		target := filepath.Join(s.cfg.WebRoot, clean)
		if st, err := os.Stat(target); err == nil && !st.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		// Missing static assets must be real 404 responses. Returning index.html
		// for a .js/.css/.json request makes browsers reject HTML under nosniff.
		if filepath.Ext(clean) != "" {
			http.NotFound(w, r)
			return
		}
		accept := strings.ToLower(r.Header.Get("Accept"))
		if accept != "" && !strings.Contains(accept, "text/html") && !strings.Contains(accept, "*/*") {
			http.NotFound(w, r)
			return
		}
		clone := r.Clone(r.Context())
		clone.URL.Path = "/"
		fs.ServeHTTP(w, clone)
	})
}
func (s *Server) routes(m *http.ServeMux) {
	m.HandleFunc("/api/birthdays", s.birthdays548)
	m.HandleFunc("/api/import/xlsx", s.importXlsx)
	m.HandleFunc("/api/export/xlsx", s.exportXlsx)
	m.HandleFunc("/api/health", s.legacyHealth)
	m.HandleFunc("/api/bootstrap", s.legacyBootstrap)
	m.HandleFunc("/api/report-config", s.reportConfig5079)
	m.HandleFunc("/api/indications", s.musicGate602(s.indications548))
	m.HandleFunc("/api/sync", s.legacySync)
	m.HandleFunc("/api/proposals", s.legacyProposals)
	m.HandleFunc("/api/proposals/status", s.legacyProposalStatus)
	m.HandleFunc("/api/proposals/decision", s.legacyProposalDecision)
	m.HandleFunc("/api/proposals/cancel", s.legacyProposalCancel)
	m.HandleFunc("/api/messages", s.legacyMessages)
	m.HandleFunc("/api/messages/send", s.legacyMessageSend)
	m.HandleFunc("/api/messages/read", s.legacyMessageRead)
	m.HandleFunc("/api/notifications", s.legacyNotifications)
	m.HandleFunc("/api/activity/daily", s.legacyDailyActivity)
	m.HandleFunc("/api/activity/track", s.legacyActivityTrack)
	m.HandleFunc("/api/notifications/read", s.legacyNotificationRead)
	m.HandleFunc("/api/notifications/delete", s.legacyNotificationDelete)
	m.HandleFunc("/api/reviews/submit", s.musicGate602(s.submitReview))
	m.HandleFunc("/api/reviews/item", s.musicGate602(s.legacyReviewItem))
	m.HandleFunc("/api/reviews/own-changes", s.legacyOwnReviewChanges)
	m.HandleFunc("/api/reviews/decision", s.musicGate602(s.legacyReviewDecision))
	m.HandleFunc("/api/progress/reset", s.musicGate602(s.legacyProgressReset))
	m.HandleFunc("/api/admin/changes", s.legacyAdminChangeSubmit)
	m.HandleFunc("/api/admin/restore", s.adminRestore)
	m.HandleFunc("/api/user-changes", s.legacyUserChanges)
	m.HandleFunc("/api/user-changes/decision", s.legacyUserChangeDecision)
	m.HandleFunc("/api/backup", s.legacyBackup)
}
func jsonOut(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func readJSONLimited(r *http.Request, v any, maxBytes int64) error {
	defer r.Body.Close()
	return json.NewDecoder(io.LimitReader(r.Body, maxBytes)).Decode(v)
}
func readJSON(r *http.Request, v any) error {
	return readJSONLimited(r, v, 8<<20)
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.auth(r); !ok {
		http.Error(w, "token inválido", 401)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	jsonOut(w, 200, map[string]any{"ok": true, "systemName": systemName, "version": release.Version, "serverId": s.store.ServerID, "revision": s.store.Revision, "serverName": s.cfg.ServerName, "advertisedUrl": s.cfg.AdvertisedURL})
}
func (s *Server) sync(w http.ResponseWriter, r *http.Request) {
	authRole, ok := s.auth(r)
	if !ok {
		http.Error(w, "token inválido", 401)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "method", 405)
		return
	}
	var req SyncRequest
	if e := readJSON(r, &req); e != nil {
		http.Error(w, e.Error(), 400)
		return
	}
	if sess, remote := s.remoteSessionFromRequest(r); remote {
		req.UserID = sess.UserID
		req.Role = sess.Role
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	resp := SyncResponse{ServerID: s.store.ServerID, Applied: []Applied{}, Pending: []Pending{}, Conflicts: []Pending{}, Updates: []Entity{}}
	for _, ch := range req.Changes {
		if ch.EntityKey == "" || len(ch.EntityKey) > 300 {
			continue
		}
		cur := s.store.Entities[ch.EntityKey]
		currentRev := cur.Revision
		kind := "change"
		if ch.BaseRevision != currentRev {
			kind = "conflict"
		}
		isAdmin := authRole == "admin" && req.Role == "admin"
		if isAdmin && kind != "conflict" {
			e := s.apply(ch.EntityKey, ch.Payload, req.UserID, req.DeviceID)
			resp.Applied = append(resp.Applied, Applied{ChangeID: ch.ID, EntityKey: ch.EntityKey, Revision: e.Revision})
		} else {
			p := Pending{ID: "pend_" + token(10), ChangeID: ch.ID, EntityKey: ch.EntityKey, BaseRevision: ch.BaseRevision, CurrentRevision: currentRev, Payload: ch.Payload, UserID: req.UserID, Role: req.Role, DeviceID: req.DeviceID, Kind: kind, Status: "pending", CreatedAt: time.Now().Format(time.RFC3339)}
			if kind == "conflict" {
				p.Message = "La entidad cambió en otro equipo desde la revisión base."
				resp.Conflicts = append(resp.Conflicts, p)
			} else {
				p.Message = "Cambio enviado para aprobación administrativa."
				resp.Pending = append(resp.Pending, p)
			}
			s.store.Pending[p.ID] = p
			s.audit("pending", p.EntityKey, p.UserID, p.DeviceID, p.Kind)
		}
	}
	for _, e := range s.store.Entities {
		if e.Revision > req.LastRevision {
			resp.Updates = append(resp.Updates, e)
		}
	}
	sort.Slice(resp.Updates, func(i, j int) bool { return resp.Updates[i].Revision < resp.Updates[j].Revision })
	resp.Revision = s.store.Revision
	_ = s.save()
	jsonOut(w, 200, resp)
}
func (s *Server) apply(key string, payload any, user, device string) Entity {
	s.store.Revision++
	e := Entity{EntityKey: key, Revision: s.store.Revision, Payload: payload, UpdatedAt: time.Now().Format(time.RFC3339), UpdatedBy: user, DeviceID: device}
	s.store.Entities[key] = e
	s.audit("apply", key, user, device, "")
	return e
}
func (s *Server) audit(action, key, user, device, details string) {
	s.store.Audit = append(s.store.Audit, Audit{At: time.Now().Format(time.RFC3339), Action: action, EntityKey: key, UserID: user, DeviceID: device, Details: details})
	if len(s.store.Audit) > 5000 {
		s.store.Audit = s.store.Audit[len(s.store.Audit)-5000:]
	}
}
func (s *Server) pendingList(w http.ResponseWriter, r *http.Request) {
	role, ok := s.auth(r)
	if !ok || role != "admin" {
		http.Error(w, "se requiere token de administrador", 403)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	list := []Pending{}
	for _, p := range s.store.Pending {
		if p.Status == "pending" || p.Status == "postponed" {
			list = append(list, p)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt < list[j].CreatedAt })
	jsonOut(w, 200, map[string]any{"serverId": s.store.ServerID, "pending": list})
}
func asMap(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	b, _ := json.Marshal(v)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return m
}
func merge(a, b any) any {
	ma, mb := asMap(a), asMap(b)
	if ma == nil || mb == nil {
		return b
	}
	for k, v := range mb {
		ma[k] = v
	}
	return ma
}
func (s *Server) pendingDecision(w http.ResponseWriter, r *http.Request) {
	role, ok := s.auth(r)
	if !ok || role != "admin" {
		http.Error(w, "se requiere token de administrador", 403)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "method", 405)
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v4/pending/"), "/")
	if len(parts) < 2 || parts[1] != "decision" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	var d struct {
		Decision    string `json:"decision"`
		Message     string `json:"message"`
		AdminUserID string `json:"adminUserId"`
		DeviceID    string `json:"deviceId"`
	}
	if e := readJSON(r, &d); e != nil {
		http.Error(w, e.Error(), 400)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p, exists := s.store.Pending[id]
	if !exists {
		http.Error(w, "solicitud no encontrada", 404)
		return
	}
	switch d.Decision {
	case "accept":
		s.apply(p.EntityKey, p.Payload, d.AdminUserID, d.DeviceID)
		p.Status = "accepted"
	case "merge":
		cur := s.store.Entities[p.EntityKey]
		s.apply(p.EntityKey, merge(cur.Payload, p.Payload), d.AdminUserID, d.DeviceID)
		p.Status = "merged"
	case "reject":
		p.Status = "rejected"
	case "correction":
		p.Status = "correction_requested"
	case "postpone":
		p.Status = "postponed"
	default:
		http.Error(w, "decisión inválida", 400)
		return
	}
	p.Decision = d.Decision
	p.DecisionMessage = d.Message
	p.DecidedAt = time.Now().Format(time.RFC3339)
	s.store.Pending[id] = p
	s.audit("decision:"+d.Decision, p.EntityKey, d.AdminUserID, d.DeviceID, d.Message)
	_ = s.save()
	jsonOut(w, 200, map[string]any{"ok": true, "serverId": s.store.ServerID, "revision": s.store.Revision, "pending": p})
}
func (s *Server) export(w http.ResponseWriter, r *http.Request) {
	role, ok := s.auth(r)
	if !ok || role != "admin" {
		http.Error(w, "se requiere token de administrador", 403)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	jsonOut(w, 200, s.store)
}
