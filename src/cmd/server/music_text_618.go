package main

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf8"
)

type musicGlyph618 struct {
	char  rune
	style int
	width float64
}

// Plain strings plus validated style runs keep clipboard content safe. PDF
// layout uses the same embedded font metrics as the ordinary text renderer.
func (p *musicPDF610) richText618(f, t map[string]any, vars map[string]string) string {
	styles := []map[string]any{}
	ends := []int{}
	text := ""
	for _, raw := range arrayField(f, "runs618") {
		r := cloneMap610(f)
		delete(r, "runs618")
		for k, v := range legacyMap(raw) {
			r[k] = v
		}
		styles = append(styles, r)
		text += stringValue(r["text"])
		ends = append(ends, len(text))
	}
	if len(styles) == 0 {
		return p.templateText613(f, musicReplace618(stringValue(f["text"]), vars), t)
	}
	fonts := make([]musicTextFont613, len(styles))
	for i, r := range styles {
		fonts[i] = p.textFont613(r, t)
	}
	glyphs := []musicGlyph618{}
	style := 0
	for at := 0; at < len(text); {
		for style+1 < len(ends) && at >= ends[style] {
			style++
		}
		part := ""
		length := 0
		if loc := musicVariables618.FindStringIndex(text[at:]); loc != nil && loc[0] == 0 {
			length = loc[1]
			part = musicReplace618(text[at:at+length], vars)
		} else {
			ch, n := utf8.DecodeRuneInString(text[at:])
			part = string(ch)
			length = n
		}
		for _, ch := range part {
			glyphs = append(glyphs, musicGlyph618{ch, style, measureText613(string(ch), numericValue(styles[style]["size"]), fonts[style].metrics)})
		}
		at += length
	}
	pw, ph := musicPaper617(t)
	width := numericValue(f["width"]) * pw / 100
	lines := [][]musicGlyph618{}
	line := []musicGlyph618{}
	word := []musicGlyph618{}
	length := 0.0
	flushWord := func() {
		if len(word) == 0 {
			return
		}
		w := 0.
		for _, g := range word {
			w += g.width
		}
		space := musicGlyph618{' ', word[0].style, measureText613(" ", numericValue(styles[word[0].style]["size"]), fonts[word[0].style].metrics)}
		if len(line) > 0 && length+space.width+w > width {
			lines = append(lines, line)
			line = nil
			length = 0
		}
		if len(line) > 0 {
			line = append(line, space)
			length += space.width
		}
		for _, g := range word {
			if len(line) > 0 && length+g.width > width {
				lines = append(lines, line)
				line = nil
				length = 0
			}
			line = append(line, g)
			length += g.width
		}
		word = nil
	}
	for _, g := range glyphs {
		if g.char == '\n' {
			flushWord()
			lines = append(lines, line)
			line = nil
			length = 0
		} else if unicode.IsSpace(g.char) {
			flushWord()
		} else {
			word = append(word, g)
		}
	}
	flushWord()
	lines = append(lines, line)
	if f["explicitLines619"] == true || f["fitLines620"] == true {
		lines = [][]musicGlyph618{{}}
		for _, g := range glyphs {
			if g.char == '\n' {
				lines = append(lines, []musicGlyph618{})
			} else {
				lines[len(lines)-1] = append(lines[len(lines)-1], g)
			}
		}
	}
	spacing := 1.25
	if f["lineSpacing618"] != nil {
		spacing = numericValue(f["lineSpacing618"])
	}
	top := numericValue(f["y"]) * ph / 100
	var out strings.Builder
	for _, line := range lines {
		length, maxSize := 0., numericValue(f["size"])
		for _, g := range line {
			length += g.width
			maxSize = math.Max(maxSize, numericValue(styles[g.style]["size"]))
		}
		scale := 1.0
		if (f["explicitLines619"] == true || f["fitLines620"] == true) && length > width && length > 0 {
			scale = width / length
			length = width
		}
		x := numericValue(f["x"]) * pw / 100
		if f["align"] == "center" {
			x += (width - length) / 2
		} else if f["align"] == "right" {
			x += width - length
		}
		for at := 0; at < len(line); {
			end := at + 1
			part := string(line[at].char)
			w := line[at].width
			for end < len(line) && line[end].style == line[at].style {
				part += string(line[end].char)
				w += line[end].width
				end++
			}
			r := cloneMap610(styles[line[at].style])
			r["size"] = numericValue(r["size"]) * scale
			w *= scale
			r["align"] = "left"
			r["x"] = x / pw * 100
			r["y"] = (top + maxSize - numericValue(r["size"])) / ph * 100
			r["width"] = (w + .02) / pw * 100
			out.WriteString(p.templateText613(r, part, t))
			x += w
			at = end
		}
		top += maxSize * spacing
	}
	return out.String()
}
