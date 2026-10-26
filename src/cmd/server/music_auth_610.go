package main

// Email/password reuses the existing account. OIDC never creates users and the
// password reset channel stores only a one-use hash, outside exported data.
import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type authState610 struct {
	State, Nonce, Verifier string
	Expires                int64
}
type resetRecord610 struct {
	UserID, Hash, PasswordStamp string
	Expires                     int64
	Generation                  int64
}

func authBase610() string {
	base := strings.TrimRight(os.Getenv("MUSICA_PUBLIC_URL"), "/")
	u, e := url.Parse(base)
	if e != nil || u.Host == "" || u.RawQuery != "" || u.Fragment != "" || u.Path != "" && u.Path != "/" {
		return ""
	}
	if u.Scheme != "https" && !(u.Scheme == "http" && (u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost")) {
		return ""
	}
	return base
}
func googleReady610() bool {
	return authBase610() != "" && os.Getenv("MUSICA_GOOGLE_CLIENT_ID") != "" && os.Getenv("MUSICA_GOOGLE_CLIENT_SECRET") != ""
}
func mailReady610() bool {
	return authBase610() != "" && os.Getenv("MUSICA_SMTP_HOST") != "" && os.Getenv("MUSICA_SMTP_FROM") != ""
}
func authSign610(value, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func (s *Server) authCookie610(value authState610) string {
	b, _ := json.Marshal(value)
	body := base64.RawURLEncoding.EncodeToString(b)
	return body + "." + authSign610(body, s.cfg.AdminToken)
}
func (s *Server) readAuthCookie610(r *http.Request) (authState610, bool) {
	var value authState610
	c, e := r.Cookie("musica_oidc610")
	if e != nil {
		return value, false
	}
	parts := strings.Split(c.Value, ".")
	if len(parts) != 2 || !hmac.Equal([]byte(parts[1]), []byte(authSign610(parts[0], s.cfg.AdminToken))) {
		return value, false
	}
	b, e := base64.RawURLEncoding.DecodeString(parts[0])
	if e != nil || json.Unmarshal(b, &value) != nil {
		return value, false
	}
	return value, value.Expires > time.Now().Unix()
}
func (s *Server) authError610(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(400)
	b, _ := json.Marshal(message)
	fmt.Fprintf(w, `<!doctype html><meta charset="utf-8"><title>Acceso a Música</title><p id="m"></p><a href="/">Volver al ingreso</a><script>document.getElementById('m').textContent=%s</script>`, b)
}

func (s *Server) musicAuth610(w http.ResponseWriter, r *http.Request, path string) {
	if path == "/api/auth610/status" && r.Method == "GET" {
		jsonOut(w, 200, map[string]any{"ok": true, "google": googleReady610(), "emailRecovery": mailReady610()})
		return
	}
	if path == "/api/auth610/google/start" && r.Method == "GET" {
		if !googleReady610() {
			s.authError610(w, "El acceso con Google necesita configuración del servidor. Puedes ingresar con tu usuario o correo y contraseña.")
			return
		}
		state := authState610{State: token(32), Nonce: token(32), Verifier: token(48), Expires: time.Now().Add(10 * time.Minute).Unix()}
		sum := sha256.Sum256([]byte(state.Verifier))
		http.SetCookie(w, &http.Cookie{Name: "musica_oidc610", Value: s.authCookie610(state), Path: "/", MaxAge: 600, HttpOnly: true, Secure: strings.HasPrefix(authBase610(), "https:"), SameSite: http.SameSiteLaxMode})
		query := url.Values{"client_id": {os.Getenv("MUSICA_GOOGLE_CLIENT_ID")}, "redirect_uri": {authBase610() + "/api/auth610/google/callback"}, "response_type": {"code"}, "scope": {"openid email profile"}, "state": {state.State}, "nonce": {state.Nonce}, "code_challenge": {base64.RawURLEncoding.EncodeToString(sum[:])}, "code_challenge_method": {"S256"}, "prompt": {"select_account"}}
		http.Redirect(w, r, "https://accounts.google.com/o/oauth2/v2/auth?"+query.Encode(), 302)
		return
	}
	if path == "/api/auth610/google/callback" && r.Method == "GET" {
		s.googleCallback610(w, r)
		return
	}
	if path == "/api/auth610/forgot" && r.Method == "POST" {
		s.forgot610(w, r)
		return
	}
	if path == "/api/auth610/reset" && r.Method == "POST" {
		s.resetPassword610(w, r)
		return
	}
	http.NotFound(w, r)
}

var identityHTTP610 = &http.Client{Timeout: 15 * time.Second}

func googleClaims610(raw, nonce string) (map[string]any, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("ID token inválido")
	}
	decode := func(x string, v any) error {
		b, e := base64.RawURLEncoding.DecodeString(x)
		if e != nil {
			return e
		}
		return json.Unmarshal(b, v)
	}
	var header struct{ Alg, Kid string }
	claims := map[string]any{}
	if decode(parts[0], &header) != nil || header.Alg != "RS256" || header.Kid == "" || decode(parts[1], &claims) != nil {
		return nil, fmt.Errorf("ID token inválido")
	}
	resp, e := identityHTTP610.Get("https://www.googleapis.com/oauth2/v3/certs")
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()
	var jwks struct {
		Keys []struct{ Kid, Kty, Alg, N, E string }
	}
	if resp.StatusCode != 200 || json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&jwks) != nil {
		return nil, fmt.Errorf("No se pudo verificar Google")
	}
	var public *rsa.PublicKey
	for _, key := range jwks.Keys {
		if key.Kid != header.Kid || key.Kty != "RSA" || key.Alg != "RS256" {
			continue
		}
		n, ne := base64.RawURLEncoding.DecodeString(key.N)
		ex, ee := base64.RawURLEncoding.DecodeString(key.E)
		if ne != nil || ee != nil || len(n) < 256 || len(ex) == 0 || len(ex) > 4 {
			continue
		}
		number := 0
		for _, b := range ex {
			number = number*256 + int(b)
		}
		if number < 3 {
			continue
		}
		public = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: number}
		break
	}
	sig, e := base64.RawURLEncoding.DecodeString(parts[2])
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if e != nil || public == nil || rsa.VerifyPKCS1v15(public, crypto.SHA256, sum[:], sig) != nil {
		return nil, fmt.Errorf("Firma de Google inválida")
	}
	issuer := stringValue(claims["iss"])
	audience := stringValue(claims["aud"])
	now := float64(time.Now().Unix())
	if issuer != "https://accounts.google.com" && issuer != "accounts.google.com" || audience != os.Getenv("MUSICA_GOOGLE_CLIENT_ID") || claims["nonce"] != nonce || numericValue(claims["exp"]) <= now || numericValue(claims["iat"]) > now+60 || claims["email_verified"] != true || stringValue(claims["sub"]) == "" {
		return nil, fmt.Errorf("Identidad de Google inválida o vencida")
	}
	return claims, nil
}
func (s *Server) googleCallback610(w http.ResponseWriter, r *http.Request) {
	state, ok := s.readAuthCookie610(r)
	http.SetCookie(w, &http.Cookie{Name: "musica_oidc610", Path: "/", MaxAge: -1, HttpOnly: true, Secure: strings.HasPrefix(authBase610(), "https:"), SameSite: http.SameSiteLaxMode})
	if !googleReady610() || !ok || r.URL.Query().Get("state") != state.State || r.URL.Query().Get("code") == "" {
		s.authError610(w, "El ingreso con Google venció o fue cancelado. Vuelve a intentarlo.")
		return
	}
	form := url.Values{"code": {r.URL.Query().Get("code")}, "client_id": {os.Getenv("MUSICA_GOOGLE_CLIENT_ID")}, "client_secret": {os.Getenv("MUSICA_GOOGLE_CLIENT_SECRET")}, "redirect_uri": {authBase610() + "/api/auth610/google/callback"}, "grant_type": {"authorization_code"}, "code_verifier": {state.Verifier}}
	resp, e := identityHTTP610.PostForm("https://oauth2.googleapis.com/token", form)
	if e != nil {
		s.authError610(w, "No se pudo contactar con Google.")
		return
	}
	defer resp.Body.Close()
	var result struct {
		IDToken string `json:"id_token"`
	}
	if resp.StatusCode != 200 || json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result) != nil {
		s.authError610(w, "Google no confirmó el ingreso.")
		return
	}
	claims, e := googleClaims610(result.IDToken, state.Nonce)
	if e != nil {
		s.authError610(w, "Google no confirmó una identidad válida.")
		return
	}
	email := strings.ToLower(strings.TrimSpace(stringValue(claims["email"])))
	subject := stringValue(claims["sub"])
	s.mu.Lock()
	st, e := s.loadLegacy()
	if e != nil {
		s.mu.Unlock()
		s.authError610(w, "No se pudo consultar la cuenta.")
		return
	}
	db := legacyMap(cloneAny(st.State.Database))
	var user map[string]any
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if strings.EqualFold(stringValue(u["email"]), email) {
			if user != nil {
				s.mu.Unlock()
				s.authError610(w, "El correo coincide con más de una cuenta. Contacta al equipo de gestión.")
				return
			}
			user = u
		}
	}
	if !legacyActive(user) || !musicAllowed553(db, user) {
		s.mu.Unlock()
		s.authError610(w, "El correo debe estar previamente registrado y habilitado en Música.")
		return
	}
	known := stringValue(user["identitySubject610"])
	if known != "" && known != subject || known == "" && !strings.HasSuffix(email, "@gmail.com") && stringValue(claims["hd"]) == "" {
		s.mu.Unlock()
		s.authError610(w, "Para este correo ingresa con la contraseña del sistema. El acceso directo admite Gmail o una organización de Google verificada.")
		return
	}
	user["identitySubject610"] = subject
	st.State.Database = db
	st.State.ServerVersion++
	st.State.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if s.saveLegacy(st) != nil {
		s.mu.Unlock()
		s.authError610(w, "No se pudo confirmar la cuenta.")
		return
	}
	s.mu.Unlock()
	now := time.Now()
	raw := "rmt_" + token(32)
	sess := remoteSession{Product: "music", TokenHash: hashToken(raw), UserID: stringValue(user["id"]), Role: legacyRole(user), DeviceID: "google_" + token(12), IP: requestIP(r), CreatedAt: now.UTC().Format(time.RFC3339), LastSeen: now.UTC().Format(time.RFC3339), ExpiresAt: now.Add(12 * time.Hour).UTC().Format(time.RFC3339)}
	s.remoteMu.Lock()
	s.remoteSessions[sess.TokenHash] = sess
	s.saveRemoteSessionsLocked()
	s.remoteMu.Unlock()
	session := map[string]any{"token": raw, "userId": sess.UserID, "role": sess.Role, "fullName": user["fullName"], "expiresAt": sess.ExpiresAt, "deviceId": sess.DeviceID, "offline": false, "validatedOnline": true}
	b, _ := json.Marshal(session)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html><meta charset="utf-8"><title>Ingresando a Música</title><p>Ingreso confirmado…</p><script>sessionStorage.setItem('musica_online_session_v1',JSON.stringify(%s));location.replace('/');</script>`, b)
}
func (s *Server) resetPath610() string {
	return filepath.Join(filepath.Dir(s.path), "auth-reset610.json")
}
func (s *Server) resetRecords610() []resetRecord610 {
	b, _ := os.ReadFile(s.resetPath610())
	var records []resetRecord610
	_ = json.Unmarshal(b, &records)
	return records
}
func (s *Server) saveReset610(records []resetRecord610) error {
	b, e := json.Marshal(records)
	if e != nil {
		return e
	}
	return os.WriteFile(s.resetPath610(), b, 0600)
}
func passwordStamp610(u map[string]any) string {
	b, _ := json.Marshal(map[string]any{"password": u["password"], "email": u["email"]})
	return hashToken(string(b))
}
func (s *Server) forgot610(w http.ResponseWriter, r *http.Request) {
	if !mailReady610() {
		legacyError(w, 503, "La recuperación por correo necesita configuración. Contacta al equipo de gestión para una contraseña temporal.", nil)
		return
	}
	var req struct{ Email string }
	if readJSONLimited(r, &req, 4096) != nil {
		legacyError(w, 400, "Correo inválido.", nil)
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if _, e := mail.ParseAddress(email); e != nil {
		legacyError(w, 400, "Indica un correo válido.", nil)
		return
	}
	rateKey := "reset610:" + requestIP(r)
	if ok, _ := s.loginAllowed(rateKey); !ok {
		legacyError(w, 429, "Espera unos minutos antes de solicitar otro enlace.", nil)
		return
	}
	s.loginFailed(rateKey)
	s.mu.Lock()
	st, e := s.loadLegacy()
	if e != nil {
		s.mu.Unlock()
		legacyError(w, 503, "No se pudo consultar el servicio.", nil)
		return
	}
	db := legacyMap(st.State.Database)
	var user map[string]any
	matches := 0
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		if strings.EqualFold(stringValue(u["email"]), email) {
			matches++
			user = u
		}
	}
	rawToken := ""
	if matches == 1 && legacyActive(user) && musicAllowed553(db, user) {
		rawToken = token(32)
		records := []resetRecord610{}
		for _, p := range s.resetRecords610() {
			if p.Expires > time.Now().Unix() && p.UserID != stringValue(user["id"]) {
				records = append(records, p)
			}
		}
		records = append(records, resetRecord610{UserID: stringValue(user["id"]), Hash: hashToken(rawToken), PasswordStamp: passwordStamp610(user), Expires: time.Now().Add(30 * time.Minute).Unix(), Generation: legacyCanonicalRestoreGeneration(st.State)})
		if s.saveReset610(records) != nil {
			s.mu.Unlock()
			legacyError(w, 503, "No se pudo preparar la recuperación.", nil)
			return
		}
	}
	s.mu.Unlock()
	if rawToken != "" {
		if e = sendResetEmail610(email, authBase610()+"/#restablecer="+url.QueryEscape(rawToken)); e != nil {
			log.Printf("Recuperación por correo: fallo del servicio SMTP (%T)", e)
		}
	}
	jsonOut(w, 200, map[string]any{"ok": true, "message": "Si el correo está registrado y habilitado, recibirás un enlace de recuperación válido durante 30 minutos. Si no llega, contacta al equipo de gestión."})
}
func sendResetEmail610(to, link string) error {
	from, e := mail.ParseAddress(os.Getenv("MUSICA_SMTP_FROM"))
	if e != nil {
		return e
	}
	if strings.ContainsAny(to, "\r\n") {
		return fmt.Errorf("correo inválido")
	}
	host := os.Getenv("MUSICA_SMTP_HOST")
	port := os.Getenv("MUSICA_SMTP_PORT")
	if port == "" {
		port = "587"
	}
	addr := net.JoinHostPort(host, port)
	conn, e := net.DialTimeout("tcp", addr, 10*time.Second)
	if e != nil {
		return e
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(20 * time.Second))
	client, e := smtp.NewClient(conn, host)
	if e != nil {
		return e
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); !ok {
		return fmt.Errorf("SMTP necesita STARTTLS")
	}
	if e = client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); e != nil {
		return e
	}
	if user := os.Getenv("MUSICA_SMTP_USER"); user != "" {
		if e = client.Auth(smtp.PlainAuth("", user, os.Getenv("MUSICA_SMTP_PASSWORD"), host)); e != nil {
			return e
		}
	}
	if e = client.Mail(from.Address); e != nil {
		return e
	}
	if e = client.Rcpt(to); e != nil {
		return e
	}
	out, e := client.Data()
	if e != nil {
		return e
	}
	message := "From: " + from.Address + "\r\nTo: " + to + "\r\nSubject: Recuperar acceso a Ministerio de Musica\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nSolicitaste restablecer tu contraseña de Música.\r\nAbre este enlace dentro de 30 minutos:\r\n" + link + "\r\n\r\nSi no lo solicitaste, ignora este mensaje.\r\n"
	_, e = io.Copy(out, bytes.NewBufferString(message))
	if e != nil {
		return e
	}
	if e = out.Close(); e != nil {
		return e
	}
	return client.Quit()
}
func (s *Server) resetPassword610(w http.ResponseWriter, r *http.Request) {
	var req struct{ Token, Password string }
	if readJSONLimited(r, &req, 4096) != nil || len(req.Password) < 10 || len(req.Password) > 256 {
		legacyError(w, 400, "Usa una contraseña de 10 a 256 caracteres.", nil)
		return
	}
	key := "resetcheck610:" + requestIP(r)
	if ok, _ := s.loginAllowed(key); !ok {
		legacyError(w, 429, "Espera unos minutos antes de reintentar.", nil)
		return
	}
	s.loginFailed(key)
	s.mu.Lock()
	defer s.mu.Unlock()
	st, e := s.loadLegacy()
	if e != nil {
		legacyError(w, 503, "No se pudo abrir la cuenta.", nil)
		return
	}
	db := legacyMap(cloneAny(st.State.Database))
	records := s.resetRecords610()
	var valid *resetRecord610
	for i := range records {
		p := &records[i]
		if p.Hash == hashToken(req.Token) && req.Token != "" && p.Expires > time.Now().Unix() && p.Generation == legacyCanonicalRestoreGeneration(st.State) {
			valid = p
			break
		}
	}
	var user map[string]any
	if valid != nil {
		user = legacyUser(db, valid.UserID)
	}
	if valid == nil || !legacyActive(user) || !musicAllowed553(db, user) || valid.PasswordStamp != passwordStamp610(user) {
		legacyError(w, 400, "El enlace venció o ya fue utilizado. Solicita uno nuevo.", nil)
		return
	}
	saltBytes := make([]byte, 16)
	if _, e := rand.Read(saltBytes); e != nil {
		legacyError(w, 500, "No se pudo preparar la contraseña.", nil)
		return
	}
	user["password"] = map[string]any{"algorithm": "PBKDF2-SHA256", "iterations": 150000, "salt": base64.StdEncoding.EncodeToString(saltBytes), "hash": base64.StdEncoding.EncodeToString(pbkdf2SHA256([]byte(req.Password), saltBytes, 150000, 32))}
	user["mustChangePassword"] = false
	user["updatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
	st.State.Database = db
	st.State.ServerVersion++
	st.State.UpdatedAt = stringValue(user["updatedAt"])
	if s.saveLegacy(st) != nil {
		legacyError(w, 500, "No se pudo guardar la contraseña.", nil)
		return
	}
	kept := []resetRecord610{}
	for _, p := range records {
		if p.UserID != valid.UserID && p.Expires > time.Now().Unix() {
			kept = append(kept, p)
		}
	}
	_ = s.saveReset610(kept)
	s.remoteMu.Lock()
	for hash, sess := range s.remoteSessions {
		if sess.UserID == valid.UserID {
			delete(s.remoteSessions, hash)
		}
	}
	s.saveRemoteSessionsLocked()
	s.remoteMu.Unlock()
	s.loginSucceeded(key)
	jsonOut(w, 200, map[string]any{"ok": true, "message": "Contraseña actualizada. Ingresa nuevamente."})
}

func loadIdentityConfig610(directory string) error {
	b, e := os.ReadFile(filepath.Join(directory, "acceso-servidor.json"))
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		return e
	}
	var config map[string]string
	if json.Unmarshal(b, &config) != nil {
		return fmt.Errorf("acceso-servidor.json debe contener un objeto JSON de textos")
	}
	for _, key := range []string{"MUSICA_PUBLIC_URL", "MUSICA_GOOGLE_CLIENT_ID", "MUSICA_GOOGLE_CLIENT_SECRET", "MUSICA_SMTP_HOST", "MUSICA_SMTP_PORT", "MUSICA_SMTP_FROM", "MUSICA_SMTP_USER", "MUSICA_SMTP_PASSWORD"} {
		if os.Getenv(key) == "" && config[key] != "" {
			if e = os.Setenv(key, config[key]); e != nil {
				return e
			}
		}
	}
	return nil
}
