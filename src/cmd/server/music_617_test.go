package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func certificateTemplate617() map[string]any {
	return map[string]any{"name": "Plantilla de prueba", "signatureDepartment": "Firma Departamental", "signatureCoordinator": "Coordinación General", "paperSize617": "oficio", "showQR617": false,
		"fields": []any{map[string]any{"text": "{nombre}", "font": "DejaVu Serif", "x": 10., "y": 20., "width": 80., "size": 30., "align": "center", "color": "#173c76"}}}
}

func Test617PaperAndQRForEveryCertificate(t *testing.T) {
	mediaBox := regexp.MustCompile(`/MediaBox \[0 0 ([0-9.]+) ([0-9.]+)\]`)
	kinds := map[string]string{"event": "Evento"}
	for kind, title := range certificateTypes550 {
		kinds[kind] = title
	}
	for _, size := range []string{"a4", "carta", "oficio"} {
		for kind, title := range kinds {
			t.Run(size+"/"+kind, func(t *testing.T) {
				value := certificateTemplate617()
				value["paperSize617"] = size
				c := map[string]any{"kind": kind, "title": title, "fullName": "Ana Música", "verificationURL610": "https://example.org/verificar/prueba"}
				pdf := musicCertificateBytes610(c, value)
				match := mediaBox.FindSubmatch(pdf)
				if match == nil {
					t.Fatal("PDF without page dimensions")
				}
				w, _ := strconv.ParseFloat(string(match[1]), 64)
				h, _ := strconv.ParseFloat(string(match[2]), 64)
				wantW, wantH := musicPaper617(value)
				if math.Abs(w-wantW) > .01 || math.Abs(h-wantH) > .01 {
					t.Fatal("wrong physical dimensions", w, h, wantW, wantH)
				}
				qr := []byte("64.000 64.000 re f\n0 0 0 rg")
				if bytes.Contains(pdf, qr) {
					t.Fatal("hidden QR was printed")
				}
				value["showQR617"] = true
				if !bytes.Contains(musicCertificateBytes610(c, value), qr) {
					t.Fatal("visible QR missing")
				}
			})
		}
	}
	w, h := musicPaper617(nil)
	if math.Abs(w-280*72/25.4) > .01 || math.Abs(h-216*72/25.4) > .01 {
		t.Fatal("default Carta dimensions are incorrect")
	}
	for _, v := range []map[string]any{{"paperSize617": "legal"}, {"showQR617": "false"}} {
		if err := musicPaperSettings617(v, map[string]any{}); err == nil {
			t.Fatal("invalid settings accepted", v)
		}
	}
}

func Test617TemplatePreviewDoesNotIssueOrSave(t *testing.T) {
	finalSample := musicPreviewCertificate617(map[string]any{"certificateType": "final"})
	eventSample := musicPreviewCertificate617(map[string]any{"certificateType": "event"})
	if !strings.Contains(strings.Join(certificateDetails610(finalSample), " "), "100%") || !strings.Contains(strings.Join(certificateDetails610(eventSample), " "), "Asistencia confirmada") {
		t.Fatal("preview sample does not match certificate kind")
	}
	s, db := fixture550(t)
	musicEnsure610(db)
	st, _ := s.loadLegacy()
	st.State.Database = db
	_ = s.saveLegacy(st)
	st, _ = s.loadLegacy()
	before, _ := json.Marshal(st.State)
	preview := func(actor, personal string, generation int64, value map[string]any) *httptest.ResponseRecorder {
		body, _ := json.Marshal(musicRequest610{ActorID: actor, Action: "template.preview", Value: value, RestoreGeneration: generation})
		r := httptest.NewRequest("POST", "/api/music/v610", bytes.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+actor)
		if personal != "" {
			r.Header.Set("X-SGL-Personal", personal)
			r = r.WithContext(context.WithValue(r.Context(), personalContext555{}, true))
		}
		w := httptest.NewRecorder()
		s.music610(w, r)
		return w
	}
	generation := legacyCanonicalRestoreGeneration(st.State)
	w := preview("admin", "", generation, certificateTemplate617())
	if w.Code != 200 || w.Header().Get("Content-Type") != "application/pdf" || !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) || !bytes.Contains(w.Body.Bytes(), []byte("935.43 612.28")) {
		t.Fatal("preview failed", w.Code, w.Body.String()[:min(w.Body.Len(), 300)])
	}
	after, _ := s.loadLegacy()
	actual, _ := json.Marshal(after.State)
	if !bytes.Equal(before, actual) {
		t.Fatal("preview changed templates, issued certificates or progress")
	}
	for _, actor := range []string{"p", "administrativo"} {
		if response := preview(actor, "", generation, certificateTemplate617()); response.Code != 403 {
			t.Fatal("unauthorized preview", actor, response.Code)
		}
	}
	personalCode := preview("admin", "1", generation, certificateTemplate617()).Code
	staleCode := preview("admin", "", generation+1, certificateTemplate617()).Code
	if personalCode != 403 || staleCode != 409 {
		t.Fatal("personal or stale preview allowed", personalCode, staleCode)
	}
	bad := certificateTemplate617()
	bad["paperSize617"] = "random"
	if preview("admin", "", generation, bad).Code != 400 {
		t.Fatal("bad paper accepted in preview")
	}
	code, result := call610(t, s, "admin", "", "template.save", certificateTemplate617())
	if code != 200 {
		t.Fatal(code, result)
	}
	var id string
	for key, raw := range mapField(result, "templates") {
		template := legacyMap(raw)
		if template["name"] == "Plantilla de prueba" {
			id = key
			if template["paperSize617"] != "oficio" || template["showQR617"] != false {
				t.Fatal("paper or QR setting not persisted")
			}
		}
	}
	if id == "" {
		t.Fatal("saved template missing")
	}
	r := httptest.NewRequest("GET", "/api/music/v610?actorId=admin&resource=template-preview&id="+id, nil)
	r.Header.Set("Authorization", "Bearer admin")
	w = httptest.NewRecorder()
	s.music610(w, r)
	if w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte("935.43 612.28")) {
		t.Fatal("saved template preview failed", w.Code)
	}
}

func Test617FontCatalogUsesBundledAndInstalledFiles(t *testing.T) {
	fonts := musicScanFonts617(nil, true)
	families := map[string]bool{}
	for _, f := range fonts {
		families[f.Family] = true
		raw, err := musicFontFiles610.ReadFile(f.path)
		if _, parseErr := parseMusicSFNT613(raw); err != nil || parseErr != nil || f.Source != "bundled" {
			t.Fatal("bundled font not printable", f.Family, err, parseErr)
		}
	}
	for _, family := range []string{"DejaVu Sans", "DejaVu Serif", "DejaVu Sans Mono"} {
		if !families[family] {
			t.Fatal("missing bundled family", family)
		}
	}
	dir := t.TempDir()
	raw := fontData613(t)
	_ = os.WriteFile(filepath.Join(dir, "real.ttf"), raw, 0600)
	_ = os.WriteFile(filepath.Join(dir, "duplicate.ttf"), raw, 0600)
	_ = os.WriteFile(filepath.Join(dir, "invalid.ttf"), []byte("invalid font"), 0600)
	fonts = musicScanFonts617([]string{dir}, false)
	if len(fonts) != 1 || fonts[0].Source != "server" {
		t.Fatal("installed font detection or deduplication failed", fonts)
	}
	restricted := append([]byte{}, raw...)
	for at := 12; at < 12+int(binary.BigEndian.Uint16(restricted[4:]))*16; at += 16 {
		if string(restricted[at:at+4]) == "OS/2" {
			offset := int(binary.BigEndian.Uint32(restricted[at+8:]))
			binary.BigEndian.PutUint16(restricted[offset+8:], 2)
		}
	}
	if family, _, _ := musicFontName617(restricted); family != "" {
		t.Fatal("restricted font listed for embedding")
	}
	s, _ := fixture550(t)
	for _, actor := range []string{"admin", "p"} {
		r := httptest.NewRequest("GET", "/api/music/v610?actorId="+actor+"&resource=font-catalog", nil)
		r.Header.Set("Authorization", "Bearer "+actor)
		w := httptest.NewRecorder()
		s.music610(w, r)
		if actor == "p" {
			if w.Code != 403 {
				t.Fatal("aspirant accessed font catalog")
			}
			continue
		}
		if w.Code != 200 || strings.Contains(w.Body.String(), `"path"`) || !strings.Contains(w.Body.String(), "DejaVu Sans") {
			t.Fatal("catalog missing or exposes paths", w.Code)
		}
	}
	for _, source := range s.musicFonts617()[:1] {
		r := httptest.NewRequest("GET", fmt.Sprintf("/api/music/v610?actorId=admin&resource=font-source&id=%s", source.ID), nil)
		r.Header.Set("Authorization", "Bearer admin")
		w := httptest.NewRecorder()
		s.music610(w, r)
		if w.Code != 200 || !strings.HasPrefix(w.Header().Get("Content-Type"), "font/") {
			t.Fatal("bundled font cannot load", w.Code)
		}
	}
}
