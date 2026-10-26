package installstate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"testing"
)

func write(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestUninstallPreservesDataBackupsProfilesAndCustomFiles(t *testing.T) {
	root, err := CanonicalRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m := New(root, "test")
	owned := []string{"Musica.exe", "ServidorMusica.exe", UninstallerName, "app/index.html", "app/sub/owned.js", "despliegue/script.ps1"}
	for _, name := range owned {
		write(t, root, name, "program")
		if err = m.Record(name); err != nil {
			t.Fatal(err)
		}
	}
	preserved := []string{"data/certificado.pdf", "data/fonts/custom.ttf", "respaldos/backup.json", "perfil-musica/Preferences", "server-config.json", "CLAVE_INSTALACION.txt", "notas.txt", "app/mi_archivo.txt"}
	for _, name := range preserved {
		write(t, root, name, "user data")
	}
	if err = m.Save(); err != nil {
		t.Fatal(err)
	}
	m, err = Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if changed, e := m.RemoveProgram(); e != nil || len(changed) != 0 {
		t.Fatal(changed, e)
	}
	if _, err = os.Stat(filepath.Join(root, UninstallerName)); err != nil {
		t.Fatal("retry uninstaller disappeared", err)
	}
	if err = m.Finish(); err != nil {
		t.Fatal(err)
	}
	for _, name := range owned {
		if _, e := os.Stat(filepath.Join(root, filepath.FromSlash(name))); !os.IsNotExist(e) {
			t.Fatal("program remains", name, e)
		}
	}
	for _, name := range preserved {
		b, e := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if e != nil || string(b) != "user data" {
			t.Fatal("user file changed", name, e)
		}
	}
}

func TestUninstallKeepsModifiedProgramFiles(t *testing.T) {
	root, _ := CanonicalRoot(t.TempDir())
	m := New(root, "test")
	write(t, root, "app/index.html", "original")
	_ = m.Record("app/index.html")
	write(t, root, "app/index.html", "personalized")
	changed, err := m.RemoveProgram()
	if err != nil || !reflect.DeepEqual(changed, []string{"app/index.html"}) {
		t.Fatal(changed, err)
	}
	b, _ := os.ReadFile(filepath.Join(root, "app", "index.html"))
	if string(b) != "personalized" {
		t.Fatal("modified file removed")
	}
}

func TestProgramPathRejectsTraversalAndProtectedFiles(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"../outside", "/outside", "C:/Windows/a", "app/../../data/a", "app/../index.html", "app//a", "app/x:stream", "app/dir./a", "app/dir /a", "app\\outside", "data/a", "DATA/a", "server-config.json", "CLAVE_INSTALACION.txt", "perfil-musica/a", "respaldos/a", "notas.txt", "app/./a"} {
		t.Run(name, func(t *testing.T) {
			if _, err := ProgramPath(root, name); err == nil {
				t.Fatal("accepted unsafe path", name)
			}
		})
	}
}

func TestUninstallRejectsSymlinkBeforeRemovingAnything(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symbolic links need optional Windows privilege")
	}
	root, _ := CanonicalRoot(t.TempDir())
	outside := t.TempDir()
	m := New(root, "test")
	write(t, root, "Musica.exe", "program")
	_ = m.Record("Musica.exe")
	write(t, root, "app/index.html", "program")
	_ = m.Record("app/index.html")
	write(t, outside, "index.html", "precious")
	if err := os.RemoveAll(filepath.Join(root, "app")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "app")); err != nil {
		t.Fatal(err)
	}
	if _, err := m.RemoveProgram(); err == nil {
		t.Fatal("followed a symlink")
	}
	if _, err := os.Stat(filepath.Join(root, "Musica.exe")); err != nil {
		t.Fatal("partial deletion before preflight", err)
	}
	b, _ := os.ReadFile(filepath.Join(outside, "index.html"))
	if string(b) != "precious" {
		t.Fatal("external file changed")
	}
}

func TestManifestRejectsWrongProductRootOrHash(t *testing.T) {
	root, _ := CanonicalRoot(t.TempDir())
	for _, m := range []*Manifest{
		{Product: "other", Root: root, Files: map[string]string{"Musica.exe": "abc"}},
		{Product: product, Root: filepath.Dir(root), Files: map[string]string{"Musica.exe": "abc"}},
		{Product: product, Root: root, Files: map[string]string{"Musica.exe": "abc"}},
	} {
		b, _ := json.Marshal(m)
		write(t, root, ManifestName, string(b))
		if _, err := Load(root); err == nil {
			t.Fatal("accepted invalid manifest", string(b))
		}
	}
}

func stopFixture(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	root := t.TempDir()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	u, _ := url.Parse(server.URL)
	port, _ := strconv.Atoi(u.Port())
	b, _ := json.Marshal(map[string]any{"port": port, "adminToken": "test-private-token"})
	write(t, root, "server-config.json", string(b))
	write(t, root, "musica-server.lock", fmt.Sprintf("pid=%d\n", os.Getpid()))
	return root
}

func TestStopServerUsesLocalAuthenticatedEndpointAndWaitsForUnlock(t *testing.T) {
	var root string
	root = stopFixture(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/server/stop" || r.Method != "POST" || r.Header.Get("Authorization") != "Bearer test-private-token" || r.Header.Get("Origin") != "" {
			t.Error("invalid stop request")
			w.WriteHeader(403)
			return
		}
		_ = os.Remove(filepath.Join(root, "musica-server.lock"))
		w.WriteHeader(200)
	})
	if err := StopServer(root); err != nil {
		t.Fatal(err)
	}
}

func TestStopServerDoesNotFollowRedirectOrAcceptDeniedRequest(t *testing.T) {
	for _, status := range []int{http.StatusForbidden, http.StatusTemporaryRedirect} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			calls := 0
			root := stopFixture(t, func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("Location", "http://example.invalid/private")
				w.WriteHeader(status)
			})
			if err := StopServer(root); err == nil {
				t.Fatal("accepted failed stop")
			}
			if calls != 1 {
				t.Fatal("unexpected redirect or retry", calls)
			}
			if _, err := os.Stat(filepath.Join(root, "musica-server.lock")); err != nil {
				t.Fatal("removed live lock")
			}
		})
	}
}
