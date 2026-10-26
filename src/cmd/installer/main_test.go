package main

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"sgl.local/sistema-gestion-liderazgo/internal/installstate"
	"testing"
)

func testPayload(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var data bytes.Buffer
	z := zip.NewWriter(&data)
	for name, content := range files {
		w, err := z.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(content))
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}

func Test625InstallUninstallReinstallPreservesUserRecords(t *testing.T) {
	old := payload
	defer func() { payload = old }()
	payload = testPayload(t, map[string]string{"app/index.html": "new app", installstate.UninstallerName: "uninstaller", "Musica.exe": "program"})
	dir, _ := installstate.CanonicalRoot(t.TempDir())
	_ = os.MkdirAll(filepath.Join(dir, "data"), 0755)
	_ = os.WriteFile(filepath.Join(dir, "data", "records.json"), []byte("original records"), 0600)
	_ = os.WriteFile(filepath.Join(dir, "server-config.json"), []byte("original config"), 0600)
	if err := install(dir, true); err != nil {
		t.Fatal(err)
	}
	m, err := installstate.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.RemoveProgram(); err != nil {
		t.Fatal(err)
	}
	if err = m.Finish(); err != nil {
		t.Fatal(err)
	}
	if err = install(dir, true); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{"data/records.json": "original records", "server-config.json": "original config", "app/index.html": "new app"} {
		b, e := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
		if e != nil || string(b) != want {
			t.Fatal("reinstall lost", name, e)
		}
	}
}
func TestInstallerPreservesUserDataAndConfiguration(t *testing.T) {
	old := payload
	defer func() { payload = old }()
	payload = testPayload(t, map[string]string{"app/index.html": "new app", "data/old.json": "wrong", "server-config.json": "wrong"})
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "data"), 0755)
	_ = os.WriteFile(filepath.Join(dir, "data", "old.json"), []byte("original data"), 0600)
	_ = os.WriteFile(filepath.Join(dir, "server-config.json"), []byte("original config"), 0600)
	if err := install(dir, true); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{"data/old.json": "original data", "server-config.json": "original config", "app/index.html": "new app"} {
		b, e := os.ReadFile(filepath.Join(dir, path))
		if e != nil || string(b) != want {
			t.Fatalf("%s not preserved", path)
		}
	}
}
func TestInstallerRejectsTraversalAndRunningServer(t *testing.T) {
	old := payload
	defer func() { payload = old }()
	payload = testPayload(t, map[string]string{"../outside.txt": "unsafe"})
	if install(t.TempDir(), true) == nil {
		t.Fatal("accepted traversal")
	}
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "musica-server.lock"), []byte("running"), 0600)
	payload = testPayload(t, map[string]string{"app/index.html": "new"})
	if install(dir, true) == nil {
		t.Fatal("updated a running server")
	}
}

func Test624InstallerRemovesOnlyKnownObsoleteFiles(t *testing.T) {
	old := payload
	defer func() { payload = old }()
	payload = testPayload(t, map[string]string{"app/music-runtime.js": "bundle", "app/index.html": "new"})
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "app"), 0755)
	for _, name := range []string{"CAMBIOS_6.2.3.md", "app/music-610.js", "mis_notas.md"} {
		_ = os.WriteFile(filepath.Join(dir, filepath.FromSlash(name)), []byte("anterior"), 0644)
	}
	if err := install(dir, true); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"CAMBIOS_6.2.3.md", "app/music-610.js"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(name))); !os.IsNotExist(err) {
			t.Fatal("obsolete file remains", name)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "mis_notas.md")); err != nil {
		t.Fatal("user notes removed")
	}
}
