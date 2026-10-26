package main

import (
	"fmt"
	"strconv"
	"strings"
)

type musicTextFont613 struct {
	name         string
	metrics      musicFont610
	bold, italic bool
}

func (p *musicPDF610) textFont613(f, t map[string]any) musicTextFont613 {
	index := pdfFieldFont610(f)
	result := musicTextFont613{fmt.Sprintf("F%d", index+1), musicFonts610[index], index%2 == 1, false}
	id := stringValue(f["fontAsset613"])
	if id == "" {
		return result
	}
	font, ok := p.fontMetrics613[id]
	if !ok {
		raw := musicFontRaw613(map[string]any{"dataUrl": mapField(t, "fontPrograms613")[id]})
		var err error
		font, err = musicCachedFont613(raw)
		if err != nil {
			return result
		}
		p.fontMetrics613[id] = font
		programType, fontType, fontKey := "", "TrueType", "FontFile2"
		if font.openType {
			programType = " /Subtype /OpenType"
			fontType = "Type1"
			fontKey = "FontFile3"
			p.openType613 = true
		}
		data := p.add(fmt.Sprintf("<< /Length %d /Length1 %d /Filter /FlateDecode%s >>\nstream\n%sendstream", len(font.compressed), font.length, programType, font.compressed))
		angle := 0
		if font.italic {
			angle = -12
		}
		descriptor := p.add(fmt.Sprintf("<< /Type /FontDescriptor /FontName /%s /Flags %d /FontBBox [%d %d %d %d] /ItalicAngle %d /Ascent %d /Descent %d /CapHeight %d /StemV 80 /%s %d 0 R >>", font.Name, font.Flags, font.BBox[0], font.BBox[1], font.BBox[2], font.BBox[3], angle, font.Ascent, font.Descent, font.Ascent, fontKey, data))
		var widths strings.Builder
		for _, n := range font.Widths {
			fmt.Fprintf(&widths, "%d ", n)
		}
		resource := "L" + strconv.Itoa(len(p.fontMetrics613))
		p.fonts613[resource] = p.add(fmt.Sprintf("<< /Type /Font /Subtype /%s /BaseFont /%s /Encoding /WinAnsiEncoding /FirstChar 32 /LastChar 255 /Widths [%s] /FontDescriptor %d 0 R >>", fontType, font.Name, widths.String(), descriptor))
		// Resource IDs are private to a PDF; keep their name alongside the metric.
		font.File = resource
		p.fontMetrics613[id] = font
	}
	return musicTextFont613{font.File, font.musicFont610, font.bold, font.italic}
}

func measureText613(text string, size float64, font musicFont610) float64 {
	width := 0
	for _, r := range text {
		code := int(r)
		if r == '–' || r == '—' {
			code = int('-')
		}
		if code < 32 || code > 255 {
			code = int('?')
		}
		width += font.Widths[code-32]
	}
	return float64(width) * size / 1000
}

func wrapText613(text string, width, size float64, font musicFont610) []string {
	measure := func(value string) float64 { return measureText613(value, size, font) }
	if width <= 0 {
		width = 1
	}
	result := []string{}
	for _, paragraph := range strings.Split(text, "\n") {
		line := ""
		for _, word := range strings.Fields(paragraph) {
			candidate := word
			if line != "" {
				candidate = line + " " + word
			}
			if line != "" && measure(candidate) > width {
				result = append(result, line)
				line = ""
			}
			for measure(word) > width {
				chars := []rune(word)
				at := 1
				for at < len(chars) && measure(string(chars[:at+1])) <= width {
					at++
				}
				result = append(result, string(chars[:at]))
				word = string(chars[at:])
			}
			if line != "" {
				line += " "
			}
			line += word
		}
		result = append(result, line)
	}
	return result
}

func (p *musicPDF610) templateText613(f map[string]any, text string, t map[string]any) string {
	var b strings.Builder
	size := numericValue(f["size"])
	if size <= 0 {
		size = 12
	}
	pw, ph := musicPaper617(t)
	width := numericValue(f["width"]) * pw / 100
	x, y := numericValue(f["x"])*pw/100, ph-numericValue(f["y"])*ph/100-size
	font := p.textFont613(f, t)
	color := "0.07 0.18 0.29"
	if value, err := strconv.ParseUint(strings.TrimPrefix(stringValue(f["color"]), "#"), 16, 32); err == nil {
		color = fmt.Sprintf("%.3f %.3f %.3f", float64(value>>16)/255, float64((value>>8)&255)/255, float64(value&255)/255)
	}
	fmt.Fprintf(&b, "q %s%s rg %s RG\n", p.opacity618(f["opacity618"]), color, color)
	shear := 0.
	if f["italic"] == true && !font.italic {
		shear = .22
	}
	mode := 0
	if f["bold"] == true && !font.bold {
		mode = 2
	}
	lines := wrapText613(text, width, size, font.metrics)
	if f["explicitLines619"] == true || f["fitLines620"] == true {
		lines = musicExplicitLines619(text)
	}
	for _, line := range lines {
		xx := x
		lineSize := size
		length := measureText613(line, lineSize, font.metrics)
		if (f["explicitLines619"] == true || f["fitLines620"] == true) && length > width && length > 0 {
			lineSize *= width / length
			length = width
		}
		if f["align"] == "center" {
			xx += (width - length) / 2
		} else if f["align"] == "right" {
			xx += width - length
		}
		fmt.Fprintf(&b, "BT /%s %.2f Tf %d Tr %.3f w 1 0 %.3f 1 %.2f %.2f Tm %s Tj ET\n", font.name, lineSize, mode, lineSize/40, shear, xx, y, pdfLiteral610(line))
		if f["underline"] == true && line != "" {
			fmt.Fprintf(&b, "%.3f w %.2f %.2f m %.2f %.2f l S\n", lineSize/20, xx, y-lineSize*.13, xx+length, y-lineSize*.13)
		}
		spacing := 1.25
		if f["lineSpacing618"] != nil {
			spacing = numericValue(f["lineSpacing618"])
		}
		y -= size * spacing
	}
	b.WriteString("Q\n")
	return b.String()
}

func musicSignatures613(t map[string]any) map[string]map[string]any {
	result := map[string]map[string]any{}
	if t["textOnly620"] == true {
		return result
	}
	for key, x := range map[string]float64{"department": 16.4, "coordinator": 52.3} {
		f := legacyMap(mapField(t, "signatures613")[key])
		if f == nil {
			f = map[string]any{"x": x, "y": 82.2, "width": 23.8, "size": 10., "font": "Arial", "align": "left", "color": "#1f2933"}
		}
		result[key] = f
	}
	return result
}
