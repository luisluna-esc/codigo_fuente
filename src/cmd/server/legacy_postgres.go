//go:build !windows

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Fuera de Windows no hay winsqlite3.dll. Con MUSICA_DATABASE_URL el estado se guarda en
// PostgreSQL con el mismo formato que SQLite (sgl_kv y sgl_files, en el esquema legacy);
// sin esa variable se mantiene la persistencia JSON de siempre.
func legacyPostgresURL() string { return strings.TrimSpace(os.Getenv("MUSICA_DATABASE_URL")) }

func legacySQLiteRuntimeAvailable() bool { return legacyPostgresURL() != "" }

// Con PostgreSQL configurado, un error de la base se informa en lugar de pasar a JSON:
// guardar en dos lugares distintos dejaría datos fuera de la base al reiniciar.
func legacyStorageStrict() bool { return legacyPostgresURL() != "" }

var legacyPG struct {
	once sync.Once
	db   *sql.DB
	err  error
}

func legacyPostgres() (*sql.DB, error) {
	legacyPG.once.Do(func() {
		db, err := sql.Open("pgx", legacyPostgresURL())
		if err != nil {
			legacyPG.err = err
			return
		}
		db.SetMaxOpenConns(4)
		db.SetConnMaxIdleTime(5 * time.Minute)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		for _, query := range []string{
			"CREATE SCHEMA IF NOT EXISTS legacy",
			"CREATE TABLE IF NOT EXISTS legacy.sgl_kv (k TEXT PRIMARY KEY, v BYTEA NOT NULL, updated_at TEXT NOT NULL DEFAULT '')",
			"CREATE TABLE IF NOT EXISTS legacy.sgl_files (file_key TEXT PRIMARY KEY, ord INTEGER NOT NULL, v BYTEA NOT NULL)",
		} {
			if _, err = db.ExecContext(ctx, query); err != nil {
				_ = db.Close()
				legacyPG.err = fmt.Errorf("PostgreSQL: %w", err)
				return
			}
		}
		log.Printf("Almacenamiento: PostgreSQL (esquema legacy)")
		legacyPG.db = db
	})
	return legacyPG.db, legacyPG.err
}

var legacyPostgresKeys = []string{"database", "proposals", "messages", "notifications", "user_changes", "applied_operations", "recoveries548"}

func legacySQLiteLoad(_ string) (legacyStore, bool, error) {
	var out legacyStore
	// Hasta leer los archivos no se sabe qué tiene la base: el próximo guardado los reescribe.
	legacyPGFiles.forget()
	db, err := legacyPostgres()
	if err != nil {
		return out, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return out, false, err
	}
	defer tx.Rollback()
	read := func(key string, target any) (bool, error) {
		var raw []byte
		e := tx.QueryRowContext(ctx, "SELECT v FROM legacy.sgl_kv WHERE k = $1", key).Scan(&raw)
		if errors.Is(e, sql.ErrNoRows) {
			return false, nil
		}
		if e != nil {
			return false, e
		}
		return true, legacySQLiteDecode(raw, target)
	}
	var meta legacySQLiteStateMeta
	found, err := read("state_meta", &meta)
	if err != nil || !found {
		return out, false, err
	}
	legacySQLiteApplyMeta(&out.State, meta)
	targets := map[string]any{
		"database": &out.State.Database, "proposals": &out.Proposals, "messages": &out.Messages,
		"notifications": &out.Notifications, "user_changes": &out.UserChanges,
		"applied_operations": &out.AppliedOperations, "recoveries548": &out.Recoveries548,
	}
	for _, key := range legacyPostgresKeys {
		if _, err = read(key, targets[key]); err != nil {
			return out, false, fmt.Errorf("leer %s: %w", key, err)
		}
	}
	rows, err := tx.QueryContext(ctx, "SELECT file_key, ord, v FROM legacy.sgl_files ORDER BY ord")
	if err != nil {
		return out, false, err
	}
	defer rows.Close()
	out.State.Files = []any{}
	keys, ords := []string{}, []int{}
	for rows.Next() {
		var key string
		var ord int
		var raw []byte
		var item any
		if err = rows.Scan(&key, &ord, &raw); err != nil {
			return out, false, err
		}
		if err = legacySQLiteDecode(raw, &item); err != nil {
			return out, false, err
		}
		out.State.Files = append(out.State.Files, item)
		keys, ords = append(keys, key), append(ords, ord)
	}
	if err = rows.Err(); err != nil {
		return out, false, err
	}
	// Se recuerda lo que hay en la base para que el próximo guardado escriba solo lo que cambie.
	if saved, e := legacyFileRowsOf(keys, ords, out.State.Files); e == nil {
		legacyPGFiles.set(saved)
	} else {
		legacyPGFiles.forget()
	}
	return out, true, nil
}

// Archivos guardados en legacy.sgl_files (ver legacy_files_delta.go).
var legacyPGFiles legacyFileState

func legacySQLiteSave(_ string, st legacyStore, persistFiles bool) error {
	values := map[string]any{
		"state_meta": legacySQLiteMetaFromState(st.State), "database": st.State.Database, "proposals": st.Proposals,
		"messages": st.Messages, "notifications": st.Notifications, "user_changes": st.UserChanges,
		"applied_operations": st.AppliedOperations, "recoveries548": st.Recoveries548,
	}
	encoded := make(map[string][]byte, len(values))
	for key, value := range values {
		b, err := legacySQLiteEncode(value)
		if err != nil {
			return err
		}
		encoded[key] = b
	}
	db, err := legacyPostgres()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for key, value := range encoded {
		if _, err = tx.ExecContext(ctx, "INSERT INTO legacy.sgl_kv (k, v, updated_at) VALUES ($1, $2, $3) ON CONFLICT (k) DO UPDATE SET v = EXCLUDED.v, updated_at = EXCLUDED.updated_at", key, value, st.State.UpdatedAt); err != nil {
			return fmt.Errorf("guardar %s: %w", key, err)
		}
	}
	// Cuando cambiaron los archivos se escriben solo los nuevos o modificados (con lo guardado
	// conocido); sin ese dato, se reescriben todos como antes. Sin persistFiles no se tocan.
	if persistFiles {
		previous, known := legacyPGFiles.snapshot()
		if !known {
			previous = nil
		}
		next, stats, e := legacyWriteFiles(ctx, tx, st.State.Files, previous)
		if e != nil {
			legacyPGFiles.forget()
			return e
		}
		if err = tx.Commit(); err != nil {
			legacyPGFiles.forget()
			return err
		}
		legacyPGFiles.set(next)
		if stats.changed() {
			log.Printf("Archivos guardados: %d escritos, %d quitados, %d reordenados (completo: %t)", stats.Written, stats.Deleted, stats.Reordered, stats.Full)
		}
		return nil
	}
	return tx.Commit()
}
