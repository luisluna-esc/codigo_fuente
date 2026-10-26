//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

const (
	sqliteOK       = 0
	sqliteROW      = 100
	sqliteDONE     = 101
	sqliteOpenRW   = 0x00000002
	sqliteOpenMake = 0x00000004
	sqliteOpenFull = 0x00010000
)

var (
	winsqliteDLL      = syscall.NewLazyDLL("winsqlite3.dll")
	pSQLiteOpenV2     = winsqliteDLL.NewProc("sqlite3_open_v2")
	pSQLiteCloseV2    = winsqliteDLL.NewProc("sqlite3_close_v2")
	pSQLitePrepareV2  = winsqliteDLL.NewProc("sqlite3_prepare_v2")
	pSQLiteStep       = winsqliteDLL.NewProc("sqlite3_step")
	pSQLiteFinalize   = winsqliteDLL.NewProc("sqlite3_finalize")
	pSQLiteBindText   = winsqliteDLL.NewProc("sqlite3_bind_text")
	pSQLiteBindBlob   = winsqliteDLL.NewProc("sqlite3_bind_blob")
	pSQLiteBindInt64  = winsqliteDLL.NewProc("sqlite3_bind_int64")
	pSQLiteColumnBlob = winsqliteDLL.NewProc("sqlite3_column_blob")
	pSQLiteColumnByte = winsqliteDLL.NewProc("sqlite3_column_bytes")
	pSQLiteErrmsg     = winsqliteDLL.NewProc("sqlite3_errmsg")
	pSQLiteBusy       = winsqliteDLL.NewProc("sqlite3_busy_timeout")
)

type winSQLite struct{ db uintptr }

func legacySQLiteRuntimeAvailable() bool {
	if err := winsqliteDLL.Load(); err != nil {
		return false
	}
	for _, proc := range []*syscall.LazyProc{pSQLiteOpenV2, pSQLiteCloseV2, pSQLitePrepareV2, pSQLiteStep, pSQLiteFinalize, pSQLiteBindText, pSQLiteBindBlob, pSQLiteBindInt64, pSQLiteColumnBlob, pSQLiteColumnByte, pSQLiteErrmsg, pSQLiteBusy} {
		if err := proc.Find(); err != nil {
			return false
		}
	}
	return true
}

func sqliteCString(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	const max = 4096
	buf := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), max)
	n := 0
	for n < len(buf) && buf[n] != 0 {
		n++
	}
	return string(buf[:n])
}

func (s *winSQLite) err(code uintptr, op string) error {
	msg := ""
	if s != nil && s.db != 0 {
		p, _, _ := pSQLiteErrmsg.Call(s.db)
		msg = sqliteCString(p)
	}
	if msg == "" {
		msg = fmt.Sprintf("codigo %d", code)
	}
	return fmt.Errorf("sqlite %s: %s", op, msg)
}

func sqliteOpen(path string) (*winSQLite, error) {
	if !legacySQLiteRuntimeAvailable() {
		return nil, errors.New("winsqlite3.dll no disponible")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	name, err := syscall.BytePtrFromString(path)
	if err != nil {
		return nil, err
	}
	var db uintptr
	rc, _, _ := pSQLiteOpenV2.Call(uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(&db)), sqliteOpenRW|sqliteOpenMake|sqliteOpenFull, 0)
	runtime.KeepAlive(name)
	s := &winSQLite{db: db}
	if rc != sqliteOK || db == 0 {
		if db != 0 {
			defer pSQLiteCloseV2.Call(db)
		}
		return nil, s.err(rc, "abrir")
	}
	_, _, _ = pSQLiteBusy.Call(db, 10000)
	return s, nil
}

func (s *winSQLite) close() {
	if s != nil && s.db != 0 {
		_, _, _ = pSQLiteCloseV2.Call(s.db)
		s.db = 0
	}
}

func (s *winSQLite) prepare(query string) (uintptr, error) {
	q, err := syscall.BytePtrFromString(query)
	if err != nil {
		return 0, err
	}
	var stmt uintptr
	rc, _, _ := pSQLitePrepareV2.Call(s.db, uintptr(unsafe.Pointer(q)), uintptr(len(query)), uintptr(unsafe.Pointer(&stmt)), 0)
	runtime.KeepAlive(q)
	if rc != sqliteOK {
		return 0, s.err(rc, "preparar")
	}
	return stmt, nil
}

func (s *winSQLite) exec(query string) error {
	stmt, err := s.prepare(query)
	if err != nil {
		return err
	}
	defer pSQLiteFinalize.Call(stmt)
	for {
		rc, _, _ := pSQLiteStep.Call(stmt)
		switch rc {
		case sqliteROW:
			continue
		case sqliteDONE:
			return nil
		default:
			return s.err(rc, "ejecutar")
		}
	}
}

func sqliteBindText(stmt uintptr, index int, value string) error {
	p, err := syscall.BytePtrFromString(value)
	if err != nil {
		return err
	}
	transient := ^uintptr(0)
	rc, _, _ := pSQLiteBindText.Call(stmt, uintptr(index), uintptr(unsafe.Pointer(p)), uintptr(len(value)), transient)
	runtime.KeepAlive(p)
	if rc != sqliteOK {
		return fmt.Errorf("sqlite bind text: codigo %d", rc)
	}
	return nil
}

func sqliteBindBlob(stmt uintptr, index int, value []byte) error {
	var ptr uintptr
	if len(value) > 0 {
		ptr = uintptr(unsafe.Pointer(&value[0]))
	}
	transient := ^uintptr(0)
	rc, _, _ := pSQLiteBindBlob.Call(stmt, uintptr(index), ptr, uintptr(len(value)), transient)
	runtime.KeepAlive(value)
	if rc != sqliteOK {
		return fmt.Errorf("sqlite bind blob: codigo %d", rc)
	}
	return nil
}

func sqliteBindInt64(stmt uintptr, index int, value int64) error {
	rc, _, _ := pSQLiteBindInt64.Call(stmt, uintptr(index), uintptr(value))
	if rc != sqliteOK {
		return fmt.Errorf("sqlite bind int64: codigo %d", rc)
	}
	return nil
}

func sqliteColumnBlob(stmt uintptr, column int) []byte {
	ptr, _, _ := pSQLiteColumnBlob.Call(stmt, uintptr(column))
	n, _, _ := pSQLiteColumnByte.Call(stmt, uintptr(column))
	if ptr == 0 || n == 0 {
		return []byte{}
	}
	return append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(ptr)), int(n))...)
}

func (s *winSQLite) ensureSchema() error {
	for _, query := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA wal_autocheckpoint=1000",
		"CREATE TABLE IF NOT EXISTS sgl_kv (k TEXT PRIMARY KEY, v BLOB NOT NULL, updated_at TEXT NOT NULL DEFAULT '')",
		"CREATE TABLE IF NOT EXISTS sgl_files (file_key TEXT PRIMARY KEY, ord INTEGER NOT NULL, v BLOB NOT NULL)",
		fmt.Sprintf("PRAGMA user_version=%d", legacySQLiteSchemaVersion),
	} {
		if err := s.exec(query); err != nil {
			return err
		}
	}
	return nil
}

func (s *winSQLite) putKV(key string, value []byte, updatedAt string) error {
	stmt, err := s.prepare("INSERT OR REPLACE INTO sgl_kv(k,v,updated_at) VALUES(?1,?2,?3)")
	if err != nil {
		return err
	}
	defer pSQLiteFinalize.Call(stmt)
	if err = sqliteBindText(stmt, 1, key); err != nil {
		return err
	}
	if err = sqliteBindBlob(stmt, 2, value); err != nil {
		return err
	}
	if err = sqliteBindText(stmt, 3, updatedAt); err != nil {
		return err
	}
	rc, _, _ := pSQLiteStep.Call(stmt)
	if rc != sqliteDONE {
		return s.err(rc, "guardar "+key)
	}
	return nil
}

func (s *winSQLite) getKV(key string) ([]byte, bool, error) {
	stmt, err := s.prepare("SELECT v FROM sgl_kv WHERE k=?1")
	if err != nil {
		return nil, false, err
	}
	defer pSQLiteFinalize.Call(stmt)
	if err = sqliteBindText(stmt, 1, key); err != nil {
		return nil, false, err
	}
	rc, _, _ := pSQLiteStep.Call(stmt)
	switch rc {
	case sqliteDONE:
		return nil, false, nil
	case sqliteROW:
		return sqliteColumnBlob(stmt, 0), true, nil
	default:
		return nil, false, s.err(rc, "leer "+key)
	}
}

func (s *winSQLite) replaceFiles(files []any) error {
	if err := s.exec("DELETE FROM sgl_files"); err != nil {
		return err
	}
	for i, raw := range files {
		payload, err := legacySQLiteEncode(raw)
		if err != nil {
			return err
		}
		id := fmt.Sprint(legacyMap(raw)["id"])
		key := fmt.Sprintf("%08d:%s", i, id)
		stmt, err := s.prepare("INSERT INTO sgl_files(file_key,ord,v) VALUES(?1,?2,?3)")
		if err != nil {
			return err
		}
		bindErr := sqliteBindText(stmt, 1, key)
		if bindErr == nil {
			bindErr = sqliteBindInt64(stmt, 2, int64(i))
		}
		if bindErr == nil {
			bindErr = sqliteBindBlob(stmt, 3, payload)
		}
		if bindErr != nil {
			pSQLiteFinalize.Call(stmt)
			return bindErr
		}
		rc, _, _ := pSQLiteStep.Call(stmt)
		pSQLiteFinalize.Call(stmt)
		if rc != sqliteDONE {
			return s.err(rc, "guardar archivo")
		}
	}
	return nil
}

func (s *winSQLite) loadFiles() ([]any, error) {
	stmt, err := s.prepare("SELECT v FROM sgl_files ORDER BY ord ASC")
	if err != nil {
		return nil, err
	}
	defer pSQLiteFinalize.Call(stmt)
	files := []any{}
	for {
		rc, _, _ := pSQLiteStep.Call(stmt)
		switch rc {
		case sqliteDONE:
			return files, nil
		case sqliteROW:
			var item any
			if err := legacySQLiteDecode(sqliteColumnBlob(stmt, 0), &item); err != nil {
				return nil, err
			}
			files = append(files, item)
		default:
			return nil, s.err(rc, "leer archivos")
		}
	}
}

func legacySQLiteLoad(path string) (legacyStore, bool, error) {
	var out legacyStore
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return out, false, nil
	} else if err != nil {
		return out, false, err
	}
	db, err := sqliteOpen(path)
	if err != nil {
		return out, false, err
	}
	defer db.close()
	if err = db.ensureSchema(); err != nil {
		return out, false, err
	}
	metaRaw, found, err := db.getKV("state_meta")
	if err != nil || !found {
		return out, found, err
	}
	var meta legacySQLiteStateMeta
	if err = legacySQLiteDecode(metaRaw, &meta); err != nil {
		return out, false, err
	}
	legacySQLiteApplyMeta(&out.State, meta)
	read := func(key string, target any) error {
		raw, ok, e := db.getKV(key)
		if e != nil || !ok {
			return e
		}
		return legacySQLiteDecode(raw, target)
	}
	if err = read("database", &out.State.Database); err != nil {
		return out, false, err
	}
	if err = read("proposals", &out.Proposals); err != nil {
		return out, false, err
	}
	if err = read("messages", &out.Messages); err != nil {
		return out, false, err
	}
	if err = read("notifications", &out.Notifications); err != nil {
		return out, false, err
	}
	if err = read("user_changes", &out.UserChanges); err != nil {
		return out, false, err
	}
	if err = read("applied_operations", &out.AppliedOperations); err != nil {
		return out, false, err
	}
	if err = read("recoveries548", &out.Recoveries548); err != nil { return out, false, err }
 out.State.Files, err = db.loadFiles()
	if err != nil {
		return out, false, err
	}
	return out, true, nil
}

func legacySQLiteSave(path string, st legacyStore, persistFiles bool) error {
	meta, err := legacySQLiteEncode(legacySQLiteMetaFromState(st.State))
	if err != nil {
		return err
	}
	values := map[string]any{
		"recoveries548": st.Recoveries548,
 "database": st.State.Database, "proposals": st.Proposals, "messages": st.Messages,
		"notifications": st.Notifications, "user_changes": st.UserChanges, "applied_operations": st.AppliedOperations,
	}
	encoded := make(map[string][]byte, len(values))
	for key, value := range values {
		b, e := legacySQLiteEncode(value)
		if e != nil {
			return e
		}
		encoded[key] = b
	}
	db, err := sqliteOpen(path)
	if err != nil {
		return err
	}
	defer db.close()
	if err = db.ensureSchema(); err != nil {
		return err
	}
	if err = db.exec("BEGIN IMMEDIATE"); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = db.exec("ROLLBACK")
		}
	}()
	updated := st.State.UpdatedAt
	if err = db.putKV("state_meta", meta, updated); err != nil {
		return err
	}
	for _, key := range []string{"database", "proposals", "messages", "notifications", "user_changes", "applied_operations"} {
		if err = db.putKV(key, encoded[key], updated); err != nil {
			return err
		}
	}
	if persistFiles {
		if err = db.replaceFiles(st.State.Files); err != nil {
			return err
		}
	}
	if err = db.exec("COMMIT"); err != nil {
		return err
	}
	committed = true
	return nil
}
