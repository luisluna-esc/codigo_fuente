package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
)

// Guardado incremental de los archivos (fotos, documentos y PDF) en legacy.sgl_files.
//
// Antes, cada guardado que tocaba un archivo borraba e insertaba de nuevo TODOS los
// archivos (decenas de MB) mientras el servidor tenía tomado su candado global; las demás
// pantallas esperaban y mostraban "Conexión interrumpida". Ahora se recuerda lo que ya está
// en la base y solo se escriben los archivos nuevos o modificados, se borran los quitados
// y se corrige el orden de los demás sin volver a enviar su contenido.

// legacyFileRow es lo que la base tiene guardado para un archivo.
type legacyFileRow struct {
	Ord     int
	Meta    []byte // el archivo sin dataUrl, en JSON
	Data    string // dataUrl (las cadenas de Go no cambian: compararlas es barato si son la misma)
	HasData bool
}

// legacyFileState recuerda las filas de legacy.sgl_files tras la última carga o guardado.
type legacyFileState struct {
	mu     sync.Mutex
	loaded bool
	rows   map[string]legacyFileRow
}

func (c *legacyFileState) snapshot() (map[string]legacyFileRow, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.rows, c.loaded
}

func (c *legacyFileState) set(rows map[string]legacyFileRow) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rows, c.loaded = rows, true
}

func (c *legacyFileState) forget() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rows, c.loaded = nil, false
}

// legacyFileKeys da una clave estable por archivo: su id (el orden ya no forma parte de la
// clave, así agregar o quitar un archivo no cambia la clave de los demás).
func legacyFileKeys(files []any) []string {
	keys := make([]string, len(files))
	seen := map[string]bool{}
	for i, raw := range files {
		id := ""
		if m := legacyMap(raw); m != nil {
			if v, ok := m["id"].(string); ok {
				id = v
			}
		}
		if id != "" && !seen["id:"+id] {
			keys[i] = "id:" + id
		} else {
			keys[i] = fmt.Sprintf("ord:%08d:%s", i, id)
		}
		seen[keys[i]] = true
	}
	return keys
}

// legacyFileSplit separa el contenido (dataUrl) del resto de los datos del archivo.
func legacyFileSplit(raw any) (legacyFileRow, error) {
	m := legacyMap(raw)
	if m == nil {
		meta, err := legacySQLiteEncode(raw)
		return legacyFileRow{Meta: meta}, err
	}
	data, isText := m["dataUrl"].(string)
	if !isText {
		meta, err := legacySQLiteEncode(m)
		return legacyFileRow{Meta: meta}, err
	}
	rest := make(map[string]any, len(m))
	for k, v := range m {
		if k != "dataUrl" {
			rest[k] = v
		}
	}
	meta, err := legacySQLiteEncode(rest)
	return legacyFileRow{Meta: meta, Data: data, HasData: true}, err
}

// legacyFileRowsOf describe los archivos tal como quedan guardados (después de una carga).
func legacyFileRowsOf(keys []string, ords []int, files []any) (map[string]legacyFileRow, error) {
	rows := make(map[string]legacyFileRow, len(files))
	for i, raw := range files {
		row, err := legacyFileSplit(raw)
		if err != nil {
			return nil, err
		}
		row.Ord = ords[i]
		rows[keys[i]] = row
	}
	return rows, nil
}

type legacyFileExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// legacyFileStats resume lo que hizo un guardado de archivos.
type legacyFileStats struct {
	Full      bool
	Written   int
	Deleted   int
	Reordered int
}

func (s legacyFileStats) changed() bool { return s.Full || s.Written+s.Deleted+s.Reordered > 0 }

// legacyWriteFiles deja legacy.sgl_files igual a files dentro de la transacción tx.
// previous es lo que hay hoy en la base; si es nil, se reescribe todo (como antes).
// Devuelve lo que quedará guardado si la transacción se confirma.
func legacyWriteFiles(ctx context.Context, tx legacyFileExec, files []any, previous map[string]legacyFileRow) (map[string]legacyFileRow, legacyFileStats, error) {
	keys := legacyFileKeys(files)
	next := make(map[string]legacyFileRow, len(files))
	stats := legacyFileStats{Full: previous == nil}
	if previous == nil {
		if _, err := tx.ExecContext(ctx, "DELETE FROM legacy.sgl_files"); err != nil {
			return nil, stats, err
		}
	} else {
		wanted := make(map[string]bool, len(keys))
		for _, key := range keys {
			wanted[key] = true
		}
		for key := range previous {
			if wanted[key] {
				continue
			}
			if _, err := tx.ExecContext(ctx, "DELETE FROM legacy.sgl_files WHERE file_key = $1", key); err != nil {
				return nil, stats, fmt.Errorf("quitar archivo: %w", err)
			}
			stats.Deleted++
		}
	}
	for i, raw := range files {
		row, err := legacyFileSplit(raw)
		if err != nil {
			return nil, stats, err
		}
		row.Ord = i
		next[keys[i]] = row
		if old, ok := previous[keys[i]]; ok && old.HasData == row.HasData && old.Data == row.Data && bytes.Equal(old.Meta, row.Meta) {
			if old.Ord != i {
				if _, err = tx.ExecContext(ctx, "UPDATE legacy.sgl_files SET ord = $2 WHERE file_key = $1", keys[i], i); err != nil {
					return nil, stats, fmt.Errorf("ordenar archivo: %w", err)
				}
				stats.Reordered++
			}
			continue
		}
		payload, err := json.Marshal(raw)
		if err != nil {
			return nil, stats, err
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO legacy.sgl_files (file_key, ord, v) VALUES ($1, $2, $3) ON CONFLICT (file_key) DO UPDATE SET ord = EXCLUDED.ord, v = EXCLUDED.v", keys[i], i, payload); err != nil {
			return nil, stats, fmt.Errorf("guardar archivo: %w", err)
		}
		stats.Written++
	}
	return next, stats, nil
}
