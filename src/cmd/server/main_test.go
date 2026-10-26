package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMerge(t *testing.T) {
	a := map[string]any{"a": 1.0, "same": "old"}
	b := map[string]any{"b": 2.0, "same": "new"}
	m := merge(a, b).(map[string]any)
	if m["same"] != "new" || m["a"] == nil || m["b"] == nil {
		t.Fatalf("merge incorrecto: %#v", m)
	}
}
func TestTokenEquality(t *testing.T) {
	if !eq("abc", "abc") || eq("abc", "abd") || eq("", "") {
		t.Fatal("comparación de token incorrecta")
	}
}

func TestSecurityHeadersAllowWebSpeechMicrophone(t *testing.T) {
	srv := &Server{}
	handler := srv.securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "https://sgl.example/", nil))

	policy := recorder.Header().Get("Permissions-Policy")
	if !strings.Contains(policy, "microphone=(self)") {
		t.Fatalf("el dictado web necesita microphone=(self); Permissions-Policy=%q", policy)
	}
	if strings.Contains(policy, "microphone=()") {
		t.Fatalf("el servidor no debe bloquear el microfono; Permissions-Policy=%q", policy)
	}
	if !strings.Contains(policy, "on-device-speech-recognition=(self)") {
		t.Fatalf("la transcripcion moderna local debe estar habilitada cuando el navegador la soporte; Permissions-Policy=%q", policy)
	}
	if !strings.Contains(policy, "camera=(self)") {
		t.Fatalf("la captura movil del informe necesita camera=(self); Permissions-Policy=%q", policy)
	}
	if strings.Contains(policy, "camera=()") || !strings.Contains(policy, "geolocation=()") {
		t.Fatalf("la camara no debe quedar bloqueada y la geolocalizacion debe seguir deshabilitada; Permissions-Policy=%q", policy)
	}
}

func TestWebRoutesMissingStaticAssetReturns404(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>SGL INDEX</html>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "class-policy.js"), []byte("'use strict';"), 0644); err != nil {
		t.Fatal(err)
	}
	srv := &Server{cfg: Config{PublicWeb: true, WebRoot: dir}}
	mux := http.NewServeMux()
	srv.webRoutes(mux)

	missing := httptest.NewRecorder()
	mux.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "http://example/missing-policy.js", nil))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing JS must return 404, got %d body=%q", missing.Code, missing.Body.String())
	}
	if strings.Contains(missing.Body.String(), "SGL INDEX") {
		t.Fatal("missing JS must not fall back to index.html")
	}

	existing := httptest.NewRecorder()
	mux.ServeHTTP(existing, httptest.NewRequest(http.MethodGet, "http://example/class-policy.js?v=5.0.28", nil))
	if existing.Code != http.StatusOK || !strings.Contains(existing.Body.String(), "use strict") {
		t.Fatalf("existing JS must be served normally, status=%d body=%q", existing.Code, existing.Body.String())
	}

	navigationReq := httptest.NewRequest(http.MethodGet, "http://example/dashboard", nil)
	navigationReq.Header.Set("Accept", "text/html")
	navigation := httptest.NewRecorder()
	mux.ServeHTTP(navigation, navigationReq)
	if navigation.Code != http.StatusOK || !strings.Contains(navigation.Body.String(), "SGL INDEX") {
		t.Fatalf("extensionless navigation should keep SPA fallback, status=%d body=%q", navigation.Code, navigation.Body.String())
	}
}
