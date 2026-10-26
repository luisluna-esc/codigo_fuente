package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
	"testing"
	"time"
)

func Test612PrintableJPEGKeepsOriginalBytesAndDimensions(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 1200, 900, 1800))
	for y := 1200; y < 1800; y++ {
		for x := 0; x < 900; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: uint8(x ^ y), A: 255})
		}
	}
	var source bytes.Buffer
	if err := jpeg.Encode(&source, img, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
	url := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(source.Bytes())
	p := newMusicPDF610()
	id := p.image(url)
	if id == 0 || id != p.image(url) {
		t.Fatal("image missing or duplicated")
	}
	p.page(842, 595, "q 842 0 0 595 0 0 cm /Artwork Do Q\n"+pdfText610(45, 540, 28, "Nombre del certificado", false), map[string]int{"Artwork": id})
	pdf := p.bytes()
	if !bytes.Contains(pdf, source.Bytes()) || !bytes.Contains(pdf, []byte("/DCTDecode")) || !bytes.Contains(pdf, []byte("/Width 900 /Height 600")) || !bytes.Contains(pdf, []byte("Nombre del certificado")) {
		t.Fatal("JPEG re-encoded, resized or text rasterized")
	}
	if len(pdf) > source.Len()+150000 {
		t.Fatalf("unexpected JPEG inflation: original %d, PDF %d", source.Len(), len(pdf))
	}
	if out := os.Getenv("MUSIC_QA_PDF_DIR"); out != "" {
		if err := os.WriteFile(out+"/Certificado_JPEG_original.pdf", pdf, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func Test612TransparentArtworkRemainsLosslessAtFullResolution(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 800, 500))
	img.Set(300, 200, color.NRGBA{R: 50, G: 100, B: 200, A: 130})
	var source bytes.Buffer
	if err := png.Encode(&source, img); err != nil {
		t.Fatal(err)
	}
	url := "data:image/png;base64," + base64.StdEncoding.EncodeToString(source.Bytes())
	p := newMusicPDF610()
	id := p.image(url)
	if id == 0 {
		t.Fatal("PNG artwork missing")
	}
	object := p.objects[id-1]
	if !bytes.Contains([]byte(object), []byte("/Width 800 /Height 500")) || !bytes.Contains([]byte(object), []byte("/SMask")) || bytes.Contains([]byte(object), []byte("/DCTDecode")) {
		t.Fatal("PNG dimensions or transparency lost")
	}
}

func Test612IncompleteImageDoesNotPassHeaderOnlyValidation(t *testing.T) {
	var source bytes.Buffer
	if err := jpeg.Encode(&source, image.NewGray(image.Rect(0, 0, 40, 30)), nil); err != nil {
		t.Fatal(err)
	}
	url := func(b []byte) string { return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(b) }
	if !validMusicPhoto592(url(source.Bytes())) {
		t.Fatal("valid JPEG rejected")
	}
	truncated := source.Bytes()[:source.Len()-12]
	if validMusicPhoto592(url(truncated)) || newMusicPDF610().jpegOriginal612(truncated) != 0 {
		t.Fatal("truncated JPEG accepted")
	}
	p := newMusicPDF610()
	if id := p.jpegOriginal612(source.Bytes()); id == 0 || !bytes.Contains([]byte(p.objects[id-1]), []byte("/DeviceGray")) {
		t.Fatal("grayscale JPEG color space lost")
	}
}

func Test612ReportParagraphsSurviveLegacyCollapsedPlainText(t *testing.T) {
	for _, tc := range []struct {
		text, rich string
		rows       int
	}{
		{"Primera.Segunda.Tercera.", "<p>Primera.</p><p>Segunda.</p><p>Tercera.</p>", 3},
		{"Primera.Segunda.Tercera.", "Primera.<br>Segunda.<br>Tercera.", 3},
		{"Primera.\nSegunda.\nTercera.", "<p>Primera.</p><p>Segunda.</p><p>Tercera.</p>", 3},
		{"Texto actual", "<p>Texto anterior.</p><p>No vigente.</p><p>No usar.</p>", 1},
		{"", "<p> </p><p><br></p><figure><img src='photo'></figure>", 0},
	} {
		sub := map[string]any{"text": tc.text, "richText": tc.rich}
		if rows := reviewNormalReportTextRows(sub); rows != tc.rows {
			t.Fatalf("rows=%d, expected=%d, text=%q", rows, tc.rows, tc.text)
		}
	}
}

func Test612ReminderPollingDoesNotTransferArtworkOrMutateVersion(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	ensureMapField(musicData550(db), "templates")["print"] = map[string]any{"dataUrl": strings.Repeat("A", 200000), "name": "Original para impresión"}
	p := reviewProgressContainer(db, "p", musicProgram553)
	ensureMapField(p, "status")["LM-PRE-1"] = "in_review"
	ensureMapField(p, "submissions")["LM-PRE-1"] = map[string]any{"submittedAt": time.Now().Add(-73 * time.Hour).UTC().Format(time.RFC3339Nano)}
	st, _ := s.loadLegacy()
	st.State.Database = db
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	before, _ := s.loadLegacy()
	w := resource611(s, "tutor", "resource=followup", false)
	if w.Code != 200 || w.Body.Len() > 10000 || bytes.Contains(w.Body.Bytes(), []byte("dataUrl")) || !bytes.Contains(w.Body.Bytes(), []byte(`"pendingCount":1`)) {
		t.Fatal("reminder includes unrelated data or has wrong recipient", w.Code, w.Body.Len())
	}
	after, _ := s.loadLegacy()
	if before.State.ServerVersion != after.State.ServerVersion {
		t.Fatal("read-only poll changed server version")
	}
	if w := resource611(s, "p", "resource=followup", false); !bytes.Contains(w.Body.Bytes(), []byte(`"pendingCount":0`)) {
		t.Fatal("applicant received another person's review reminder")
	}
	for _, raw := range arrayField(db, "modules") {
		module := legacyMap(raw)
		if module["id"] == musicProgram553 {
			module["assignedUserIds"] = append(arrayField(module, "assignedUserIds"), "admin")
		}
	}
	after.State.Database = db
	if err := s.saveLegacy(after); err != nil {
		t.Fatal(err)
	}
	w = resource611(s, "admin", "", true)
	var personal map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &personal); err != nil || w.Code != 200 || mapField(personal, "permissions")["ministries.edit"] != false {
		t.Fatal("personal view inherited management permissions", w.Code)
	}
}
