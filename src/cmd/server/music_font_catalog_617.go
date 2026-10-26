package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf16"
)

type musicFontSource617 struct {
	AssetID621 string `json:"assetId621,omitempty"`
	ID         string `json:"id"`
	Family     string `json:"family"`
	Bold       bool   `json:"bold"`
	Italic     bool   `json:"italic"`
	Source     string `json:"source"`
	path       string
}

type musicFontCatalog617 struct {
	fonts   []musicFontSource617
	expires time.Time
}

var musicBundledCatalogOnce624 sync.Once
var musicBundledCatalogFonts624 []musicFontSource617

func musicBundledCatalog624() []musicFontSource617 {
	musicBundledCatalogOnce624.Do(func() { musicBundledCatalogFonts624 = musicScanFonts617(nil, true) })
	return musicBundledCatalogFonts624
}

func musicFontName617(raw []byte) (family string, bold, italic bool) {
	if len(raw) < 12 || len(raw) > 4<<20 {
		return
	}
	tag := string(raw[:4])
	if tag != "\x00\x01\x00\x00" && tag != "true" && tag != "OTTO" {
		return
	}
	n := int(binary.BigEndian.Uint16(raw[4:]))
	if n < 1 || n > 128 || 12+n*16 > len(raw) {
		return
	}
	tables := map[string][]byte{}
	for at := 12; at < 12+n*16; at += 16 {
		offset, length := int(binary.BigEndian.Uint32(raw[at+8:])), int(binary.BigEndian.Uint32(raw[at+12:]))
		if offset > len(raw) || length > len(raw)-offset {
			return
		}
		tables[string(raw[at:at+4])] = raw[offset : offset+length]
	}
	if tables["fvar"] != nil || tables["head"] == nil || tables["hmtx"] == nil || tables["cmap"] == nil || tag == "OTTO" && tables["CFF "] == nil || tag != "OTTO" && tables["glyf"] == nil {
		return
	}
	if os2 := tables["OS/2"]; len(os2) >= 10 {
		if binary.BigEndian.Uint16(os2[8:])&0x202 != 0 {
			return
		}
		if len(os2) >= 64 {
			flags := binary.BigEndian.Uint16(os2[62:])
			bold, italic = flags&32 != 0, flags&1 != 0
		}
	}
	if head := tables["head"]; len(head) >= 46 {
		flags := binary.BigEndian.Uint16(head[44:])
		bold, italic = bold || flags&1 != 0, italic || flags&2 != 0
	}
	name := tables["name"]
	if len(name) < 6 {
		return
	}
	count, start := int(binary.BigEndian.Uint16(name[2:])), int(binary.BigEndian.Uint16(name[4:]))
	if count > 1024 || 6+count*12 > len(name) {
		return
	}
	best := -1
	legacyFamily := ""
	for at := 6; at < 6+count*12; at += 12 {
		id := binary.BigEndian.Uint16(name[at+6:])
		if id != 1 && id != 16 {
			continue
		}
		platform, language := binary.BigEndian.Uint16(name[at:]), binary.BigEndian.Uint16(name[at+4:])
		length, offset := int(binary.BigEndian.Uint16(name[at+8:])), start+int(binary.BigEndian.Uint16(name[at+10:]))
		if offset > len(name) || length > len(name)-offset {
			continue
		}
		score := 0
		if id == 16 {
			score += 4
		}
		if platform == 3 || platform == 0 {
			score++
		}
		if language == 0x409 || language == 0 {
			score += 2
		}
		value := name[offset : offset+length]
		decoded := ""
		if platform == 0 || platform == 3 {
			chars := []uint16{}
			for j := 0; j+1 < len(value); j += 2 {
				chars = append(chars, binary.BigEndian.Uint16(value[j:]))
			}
			decoded = string(utf16.Decode(chars))
		} else {
			decoded = string(value)
		}
		if id == 1 && (language == 0x409 || legacyFamily == "") {
			legacyFamily = decoded
		}
		if score <= best {
			continue
		}
		family = decoded
		best = score
	}
	// Stretch is not a supported face selector. Keep Arial Narrow distinct
	// from Arial even when its typographic family (name ID 16) says Arial.
	if strings.EqualFold(family, "Arial") && strings.EqualFold(legacyFamily, "Arial Narrow") {
		family = legacyFamily
	}
	family = strings.TrimSpace(strings.Map(func(r rune) rune {
		if r < 32 || strings.ContainsRune("\"\\<>;", r) {
			return -1
		}
		return r
	}, family))
	if len([]rune(family)) > 100 {
		family = string([]rune(family)[:100])
	}
	return
}

func musicFontRoots617() []string {
	if runtime.GOOS == "windows" {
		root := os.Getenv("WINDIR")
		if root == "" {
			root = os.Getenv("SystemRoot")
		}
		if root == "" {
			root = `C:\Windows`
		}
		roots := []string{filepath.Join(root, "Fonts")}
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			roots = append(roots, filepath.Join(local, "Microsoft", "Windows", "Fonts"))
		}
		return roots
	}
	if runtime.GOOS == "linux" {
		return []string{"/usr/share/fonts", "/usr/local/share/fonts"}
	}
	return nil
}

func musicScanFonts617(roots []string, includeBundled bool) []musicFontSource617 {
	fonts := []musicFontSource617{}
	seen := map[string]bool{}
	add := func(raw []byte, source, path string) {
		family, bold, italic := musicFontName617(raw)
		if family == "" {
			return
		}
		key := fmt.Sprintf("%s:%t:%t", family, bold, italic)
		if seen[key] {
			return
		}
		seen[key] = true
		hash := sha256.Sum256(raw)
		assetID := ""
		if source == "bundled" {
			assetID = "font613_" + hex.EncodeToString(hash[:])
		}
		fonts = append(fonts, musicFontSource617{AssetID621: assetID, ID: "source617_" + hex.EncodeToString(hash[:]), Family: family, Bold: bold, Italic: italic, Source: source, path: path})
	}
	if includeBundled {
		entries, _ := musicFontFiles610.ReadDir("fonts610")
		for _, entry := range entries {
			path := "fonts610/" + entry.Name()
			if ext := strings.ToLower(filepath.Ext(path)); ext == ".ttf" || ext == ".otf" {
				raw, _ := musicFontFiles610.ReadFile(path)
				add(raw, "bundled", path)
			}
		}
	}
	visited := 0
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if visited >= 4096 {
				return fs.SkipAll
			}
			if err != nil || d.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".ttf" && ext != ".otf" {
				return nil
			}
			visited++
			info, err := d.Info()
			if err != nil || info.Size() < 12 || info.Size() > 4<<20 {
				return nil
			}
			raw, err := os.ReadFile(path)
			if err == nil {
				add(raw, "server", path)
			}
			return nil
		})
	}
	sort.SliceStable(fonts, func(i, j int) bool {
		if fonts[i].Source != fonts[j].Source {
			return fonts[i].Source == "bundled"
		}
		return fonts[i].Family < fonts[j].Family
	})
	return fonts
}

func (s *Server) musicFonts617() []musicFontSource617 {
	// The caller holds s.mu. Do not retain all font programs in memory or copy
	// every installed font into the database: only a chosen face is saved.
	if s.fontCatalog617 == nil || time.Now().After(s.fontCatalog617.expires) {
		s.fontCatalog617 = &musicFontCatalog617{fonts: musicScanFonts617(musicFontRoots617(), true), expires: time.Now().Add(2 * time.Minute)}
	}
	return s.fontCatalog617.fonts
}

func (s *Server) musicReadFonts617(w http.ResponseWriter, r *http.Request, db, actor map[string]any) bool {
	resource := r.URL.Query().Get("resource")
	if r.Method != http.MethodGet || resource != "font-catalog" && resource != "font-source" {
		return false
	}
	if !musicFontManager613(db, actor) || personalRequest555(r) {
		legacyError(w, 403, "No tienes permiso para administrar las fuentes.", nil)
		return true
	}
	fonts := musicBundledCatalog624()
	if r.URL.Query().Get("installed") == "1" {
		fonts = s.musicFonts617()
	}
	if resource == "font-catalog" {
		jsonOut(w, 200, map[string]any{"ok": true, "fonts": fonts, "hostOS": runtime.GOOS})
		return true
	}
	// A bundled source is resolved without scanning system font directories.
	if resource == "font-source" {
		found := false
		for _, source := range fonts {
			if source.ID == r.URL.Query().Get("id") {
				found = true
				break
			}
		}
		if !found {
			fonts = s.musicFonts617()
		}
	}
	for _, source := range fonts {
		if source.ID != r.URL.Query().Get("id") {
			continue
		}
		var raw []byte
		if source.Source == "bundled" {
			raw, _ = musicFontFiles610.ReadFile(source.path)
		} else {
			raw, _ = os.ReadFile(source.path)
		}
		hash := sha256.Sum256(raw)
		if source.ID != "source617_"+hex.EncodeToString(hash[:]) {
			s.fontCatalog617 = nil
			legacyError(w, 409, "La fuente cambió. Vuelve a abrir el editor.", nil)
			return true
		}
		if _, err := musicCachedFont613(raw); err != nil {
			legacyError(w, 400, err.Error(), nil)
			return true
		}
		mime := "font/ttf"
		if string(raw[:4]) == "OTTO" {
			mime = "font/otf"
		}
		w.Header().Set("Content-Type", mime)
		etag := `"` + source.ID + `"`
		w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
		w.Header().Set("ETag", etag)
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return true
		}
		w.WriteHeader(200)
		_, _ = w.Write(raw)
		return true
	}
	legacyError(w, 404, "Fuente no encontrada. Vuelve a abrir el editor.", nil)
	return true
}
