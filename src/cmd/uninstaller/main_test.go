package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"sgl.local/sistema-gestion-liderazgo/internal/installstate"
)

func TestUninstallerPreservesEverythingWhenServerCannotStop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("native integration checks run on Linux without changing Windows registration")
	}
	root, _ := installstate.CanonicalRoot(t.TempDir())
	m := installstate.New(root, "test")
	_ = os.WriteFile(filepath.Join(root, "Musica.exe"), []byte("program"), 0600)
	_ = m.Record("Musica.exe")
	_ = m.Save()
	_ = os.WriteFile(filepath.Join(root, "musica-server.lock"), []byte(fmt.Sprintf("pid=%d\n", os.Getpid())), 0600)
	if err := uninstall(root); err == nil {
		t.Fatal("removed a running installation")
	}
	if _, err := os.Stat(filepath.Join(root, "Musica.exe")); err != nil {
		t.Fatal("removed program before stopping server")
	}
	if _, err := installstate.Load(root); err != nil {
		t.Fatal("removed retry information")
	}
}

func TestUninstallerRejectsUnmarkedDirectory(t *testing.T) {
	if err := uninstall(t.TempDir()); err == nil {
		t.Fatal("accepted arbitrary directory")
	}
}
