package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestLegacyFileKeysStableAndUnique(t *testing.T) {
	files := []any{map[string]any{"id": "a"}, map[string]any{"id": "b"}, map[string]any{"id": "a"}, map[string]any{"name": "sin id"}}
	got := legacyFileKeys(files)
	want := []string{"id:a", "id:b", "ord:00000002:a", "ord:00000003:"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("claves %v, se esperaba %v", got, want)
	}
	// Quitar el primero no cambia la clave de los demás.
	if legacyFileKeys(files[1:])[0] != "id:b" {
		t.Fatal("la clave depende del orden")
	}
}

func TestLegacyFileSplitDetectsChanges(t *testing.T) {
	a, _ := legacyFileSplit(map[string]any{"id": "x", "name": "foto", "dataUrl": "data:image/png;base64,AAAA"})
	b, _ := legacyFileSplit(map[string]any{"name": "foto", "id": "x", "dataUrl": "data:image/png;base64,AAAA"})
	if a.Data != b.Data || string(a.Meta) != string(b.Meta) || !a.HasData {
		t.Fatal("el mismo archivo debe describirse igual")
	}
	c, _ := legacyFileSplit(map[string]any{"id": "x", "name": "foto", "dataUrl": "data:image/png;base64,AAAB"})
	d, _ := legacyFileSplit(map[string]any{"id": "x", "name": "otra", "dataUrl": "data:image/png;base64,AAAA"})
	e, _ := legacyFileSplit(map[string]any{"id": "x", "name": "foto"})
	if c.Data == a.Data || string(d.Meta) == string(a.Meta) || e.HasData {
		t.Fatal("no detectó el cambio")
	}
}

// Contra una base vacía: MUSICA_TEST_DATABASE_URL=postgres://... go test -run LegacyFilesDelta ./cmd/server/
func TestLegacyFilesDeltaPostgres(t *testing.T) {
	url := os.Getenv("MUSICA_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("MUSICA_TEST_DATABASE_URL no definido")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	for _, q := range []string{
		"CREATE SCHEMA IF NOT EXISTS legacy",
		"CREATE TABLE IF NOT EXISTS legacy.sgl_files (file_key TEXT PRIMARY KEY, ord INTEGER NOT NULL, v BYTEA NOT NULL)",
		"TRUNCATE legacy.sgl_files",
		// Filas con las claves anteriores ("orden:id"): el primer guardado las reemplaza.
		`INSERT INTO legacy.sgl_files (file_key, ord, v) VALUES ('00000000:a', 0, '{"id":"a"}'::bytea)`,
	} {
		if _, err = db.ExecContext(ctx, q); err != nil {
			t.Fatal(q, err)
		}
	}
	read := func() ([]string, []any) {
		rows, err := db.QueryContext(ctx, "SELECT file_key, ord, v FROM legacy.sgl_files ORDER BY ord")
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		keys, files := []string{}, []any{}
		for rows.Next() {
			var key string
			var ord int
			var raw []byte
			var item any
			if err = rows.Scan(&key, &ord, &raw); err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(raw, &item); err != nil {
				t.Fatal(err)
			}
			keys, files = append(keys, key), append(files, item)
		}
		return keys, files
	}
	save := func(files []any, previous map[string]legacyFileRow) (map[string]legacyFileRow, legacyFileStats) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		next, stats, err := legacyWriteFiles(ctx, tx, files, previous)
		if err != nil {
			_ = tx.Rollback()
			t.Fatal(err)
		}
		if err = tx.Commit(); err != nil {
			t.Fatal(err)
		}
		return next, stats
	}
	same := func(want []any) {
		_, got := read()
		a, _ := json.Marshal(want)
		b, _ := json.Marshal(got)
		if string(a) != string(b) {
			t.Fatalf("en la base:\n%s\nse esperaba:\n%s", b, a)
		}
	}
	file := func(id, data string) map[string]any {
		return map[string]any{"id": id, "kind": "photo", "dataUrl": "data:image/png;base64," + data, "size": 3.0}
	}

	// Lo que había (claves anteriores) se conoce como tras una carga.
	keys, loaded := read()
	previous, err := legacyFileRowsOf(keys, []int{0}, loaded)
	if err != nil {
		t.Fatal(err)
	}
	files := []any{file("a", "AAAA"), file("b", "BBBB"), file("c", "CCCC")}
	previous, stats := save(files, previous)
	if stats.Written != 3 || stats.Deleted != 1 {
		t.Fatalf("primer guardado: %+v", stats)
	}
	same(files)

	// Guardar lo mismo no escribe nada.
	previous, stats = save(files, previous)
	if stats.changed() {
		t.Fatalf("sin cambios escribió: %+v", stats)
	}

	// Se cambia b, se quita a y se agrega d: solo b y d viajan; c solo cambia de orden.
	files = []any{file("b", "BBBX"), files[2], file("d", "DDDD")}
	previous, stats = save(files, previous)
	if stats.Written != 2 || stats.Deleted != 1 || stats.Reordered != 1 {
		t.Fatalf("cambios: %+v", stats)
	}
	same(files)

	// Un cambio de datos sin cambiar el contenido (nombre) también se guarda.
	renamed := map[string]any{"id": "c", "kind": "photo", "dataUrl": "data:image/png;base64,CCCC", "size": 3.0, "name": "Foto"}
	files = []any{files[0], renamed, files[2]}
	previous, stats = save(files, previous)
	if stats.Written != 1 || stats.Deleted != 0 {
		t.Fatalf("cambio de nombre: %+v", stats)
	}
	same(files)

	// Lo que describe la carga coincide con lo recordado tras guardar.
	keys, loaded = read()
	reloaded, err := legacyFileRowsOf(keys, []int{0, 1, 2}, loaded)
	if err != nil || !reflect.DeepEqual(reloaded, previous) {
		t.Fatalf("carga distinta de lo guardado: %v", err)
	}
	if _, stats = save(loaded, reloaded); stats.changed() {
		t.Fatalf("después de cargar escribió: %+v", stats)
	}

	// Sin datos conocidos se reescribe todo, como antes.
	if _, stats = save(files, nil); !stats.Full || stats.Written != 3 {
		t.Fatalf("reescritura completa: %+v", stats)
	}
	same(files)
}
