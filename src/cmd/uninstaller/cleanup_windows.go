//go:build windows

package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"sgl.local/sistema-gestion-liderazgo/internal/installstate"
	"sgl.local/sistema-gestion-liderazgo/internal/processutil"
)

// A short-lived process waits for this worker before deleting its private
// temporary directory. Encoded literals protect paths containing punctuation.
func deleteWorker(exe string) {
	dir := filepath.Dir(exe)
	temp, err := installstate.CanonicalRoot(os.TempDir())
	parent, e := installstate.CanonicalRoot(filepath.Dir(dir))
	if err != nil || e != nil || !strings.EqualFold(temp, parent) || !strings.HasPrefix(filepath.Base(dir), "Musica-desinstalar-") || filepath.Base(exe) != installstate.UninstallerName {
		return
	}
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	script := fmt.Sprintf("Wait-Process -Id %d -ErrorAction SilentlyContinue;Remove-Item -LiteralPath %s -Force -ErrorAction SilentlyContinue;Remove-Item -LiteralPath %s -ErrorAction SilentlyContinue", os.Getpid(), quote(exe), quote(dir))
	units := utf16.Encode([]rune(script))
	b := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(b[i*2:], u)
	}
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", base64.StdEncoding.EncodeToString(b))
	cmd.Dir = temp
	processutil.Hidden(cmd)
	if cmd.Start() == nil {
		_ = cmd.Process.Release()
	}
}
