package processutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLockProtectsActiveOrUnknownProcess(t *testing.T) {
	p := filepath.Join(t.TempDir(), "server.lock")
	for _, content := range []string{"unknown", fmt.Sprintf("pid=%d\n", os.Getpid())} {
		if err := os.WriteFile(p, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if ClearStaleLock(p) == nil {
			t.Fatal("live or unknown lock removed")
		}
		if _, err := os.Stat(p); err != nil {
			t.Fatal("protected lock changed")
		}
	}
}
func TestLockClearsEndedProcess(t *testing.T) {
	p := filepath.Join(t.TempDir(), "server.lock")
	if err := ClearStaleLock(p); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("pid=2147483647\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ClearStaleLock(p); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("stale lock retained")
	}
}
