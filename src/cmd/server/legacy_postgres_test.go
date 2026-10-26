//go:build !windows

package main

import (
	"encoding/json"
	"os"
	"testing"
)

// Requiere una base vacía: MUSICA_TEST_DATABASE_URL=postgres://... go test -run Postgres ./cmd/server/
func TestPostgresStorageRoundTrip(t *testing.T) {
	url := os.Getenv("MUSICA_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("MUSICA_TEST_DATABASE_URL no definido")
	}
	t.Setenv("MUSICA_DATABASE_URL", url)
	if !legacySQLiteRuntimeAvailable() || !legacyStorageStrict() {
		t.Fatal("PostgreSQL no quedó activo")
	}
	db, err := legacyPostgres()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("TRUNCATE legacy.sgl_kv, legacy.sgl_files"); err != nil {
		t.Fatal(err)
	}
	if _, found, err := legacySQLiteLoad(""); err != nil || found {
		t.Fatal("base vacía", found, err)
	}

	st := legacyStore{
		State: legacyState{ServerVersion: 7, RestoreGeneration: 2, RestoreID: "r1", UpdatedAt: "2026-10-26T10:00:00Z", InstallationID: "inst",
			Database: map[string]any{"users": []any{map[string]any{"id": "usr_1", "fullName": "Ñandú Pérez", "role": "admin"}}},
			Files:    []any{map[string]any{"id": "file_a", "dataUrl": "data:image/png;base64,AAAA"}, map[string]any{"id": "file_b", "size": 3.0}}},
		Proposals:         []legacyProposal{},
		Notifications:     []legacyNotification{{ID: "n1"}},
		AppliedOperations: []legacyAppliedOperation{},
	}
	if err = legacySQLiteSave("", st, true); err != nil {
		t.Fatal(err)
	}
	got, found, err := legacySQLiteLoad("")
	if err != nil || !found {
		t.Fatal(found, err)
	}
	want, _ := json.Marshal(st)
	have, _ := json.Marshal(got)
	if string(want) != string(have) {
		t.Fatalf("no coincide\nwant %s\nhave %s", want, have)
	}

	// Sin persistFiles se actualizan los datos y los archivos guardados se conservan.
	st.State.ServerVersion = 8
	st.State.Files = nil
	if err = legacySQLiteSave("", st, false); err != nil {
		t.Fatal(err)
	}
	got, _, err = legacySQLiteLoad("")
	if err != nil || got.State.ServerVersion != 8 || len(got.State.Files) != 2 || legacyMap(got.State.Files[1])["id"] != "file_b" {
		t.Fatal("persistFiles=false", got.State.ServerVersion, len(got.State.Files), err)
	}
}
