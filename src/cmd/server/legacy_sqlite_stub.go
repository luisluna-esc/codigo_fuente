//go:build !windows

package main

func legacySQLiteRuntimeAvailable() bool { return false }

func legacySQLiteLoad(path string) (legacyStore, bool, error) {
	return legacyStore{}, false, nil
}

func legacySQLiteSave(path string, st legacyStore, persistFiles bool) error { return nil }
