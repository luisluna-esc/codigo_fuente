package main

import (
	"bytes"
	"compress/zlib"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed fonts610
var musicFontFiles610 embed.FS

type musicFont610 struct {
	Name, File             string
	Widths                 []int
	Ascent, Descent, Flags int
	BBox                   []int
	compressed             string
	length                 int
}

var musicFonts610 = func() []musicFont610 {
	b, _ := musicFontFiles610.ReadFile("fonts610/metrics.json")
	var fonts []musicFont610
	if err := json.Unmarshal(b, &fonts); err != nil {
		panic(err)
	}
	for i := range fonts {
		raw, e := musicFontFiles610.ReadFile("fonts610/" + fonts[i].File)
		if e != nil {
			panic(e)
		}
		var out bytes.Buffer
		w := zlib.NewWriter(&out)
		_, _ = w.Write(raw)
		_ = w.Close()
		fonts[i].compressed = out.String()
		fonts[i].length = len(raw)
	}
	return fonts
}()

func (p *musicPDF610) fonts610() {
	for i, font := range musicFonts610 {
		data := p.add(fmt.Sprintf("<< /Length %d /Length1 %d /Filter /FlateDecode >>\nstream\n%sendstream", len(font.compressed), font.length, font.compressed))
		descriptor := p.add(fmt.Sprintf("<< /Type /FontDescriptor /FontName /%s /Flags %d /FontBBox [%d %d %d %d] /ItalicAngle 0 /Ascent %d /Descent %d /CapHeight %d /StemV 80 /FontFile2 %d 0 R >>", font.Name, font.Flags, font.BBox[0], font.BBox[1], font.BBox[2], font.BBox[3], font.Ascent, font.Descent, font.Ascent, data))
		var widths strings.Builder
		for _, n := range font.Widths {
			fmt.Fprintf(&widths, "%d ", n)
		}
		p.objects[2+i] = fmt.Sprintf("<< /Type /Font /Subtype /TrueType /BaseFont /%s /Encoding /WinAnsiEncoding /FirstChar 32 /LastChar 255 /Widths [%s] /FontDescriptor %d 0 R >>", font.Name, widths.String(), descriptor)
	}
}
func pdfMeasure610(text string, size float64, font int) float64 {
	if font < 0 || font >= len(musicFonts610) {
		font = 0
	}
	width := 0
	for _, r := range text {
		code := int(r)
		if code >= 32 && code <= 255 {
			width += musicFonts610[font].Widths[code-32]
		} else if r == '–' || r == '—' {
			width += musicFonts610[font].Widths[int('-')-32]
		} else {
			width += musicFonts610[font].Widths[int('?')-32]
		}
	}
	return float64(width) * size / 1000
}
func pdfTextFont610(x, y, size float64, text string, font int) string {
	return fmt.Sprintf("BT /F%d %.2f Tf 1 0 0 1 %.2f %.2f Tm %s Tj ET\n", font+1, size, x, y, pdfLiteral610(text))
}
func pdfFieldFont610(f map[string]any) int {
	font := 0
	family := strings.ToLower(stringValue(f["font"]))
	if strings.Contains(family, "georgia") || strings.Contains(family, "times") || strings.Contains(family, "palatino") {
		font = 2
	}
	if f["bold"] == true {
		font++
	}
	return font
}
