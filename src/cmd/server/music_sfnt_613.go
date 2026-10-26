package main

// Read the metrics required by the certificate writer without loading a native
// font engine. Programs are embedded once per PDF; templates retain only IDs.
import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"regexp"
	"unicode/utf16"
)

type musicSFNT613 struct {
	musicFont610
	openType, bold, italic bool
}

func parseMusicSFNT613(raw []byte) (font musicSFNT613, err error) {
	fail := func() (musicSFNT613, error) {
		return font, fmt.Errorf("Fuente TTF/OTF dañada o no compatible. Usa una fuente estática de hasta 4 MB.")
	}
	if len(raw) < 12 || len(raw) > 4<<20 {
		return fail()
	}
	tag := string(raw[:4])
	font.openType = tag == "OTTO"
	if !font.openType && tag != "\x00\x01\x00\x00" && tag != "true" {
		return fail()
	}
	u16 := func(b []byte, at int) int {
		if at < 0 || at+2 > len(b) {
			return -1
		}
		return int(binary.BigEndian.Uint16(b[at:]))
	}
	i16 := func(b []byte, at int) int {
		n := u16(b, at)
		if n > 32767 {
			n -= 65536
		}
		return n
	}
	u32 := func(b []byte, at int) int {
		if at < 0 || at+4 > len(b) {
			return -1
		}
		return int(binary.BigEndian.Uint32(b[at:]))
	}
	n := u16(raw, 4)
	if n < 1 || n > 128 || 12+n*16 > len(raw) {
		return fail()
	}
	tables := map[string][]byte{}
	for j := 0; j < n; j++ {
		at := 12 + j*16
		start, size := u32(raw, at+8), u32(raw, at+12)
		if start < 0 || size < 0 || start > len(raw) || size > len(raw)-start {
			return fail()
		}
		tables[string(raw[at:at+4])] = raw[start : start+size]
	}
	if tables["fvar"] != nil || (font.openType && tables["CFF "] == nil) || (!font.openType && tables["glyf"] == nil) {
		return fail()
	}
	head, hhea, maxp, hmtx := tables["head"], tables["hhea"], tables["maxp"], tables["hmtx"]
	if len(head) < 54 || len(hhea) < 36 || len(maxp) < 6 {
		return fail()
	}
	units, metrics, glyphs := u16(head, 18), u16(hhea, 34), u16(maxp, 4)
	if units < 16 || units > 16384 || metrics < 1 || metrics > glyphs || len(hmtx) < metrics*4+(glyphs-metrics)*2 {
		return fail()
	}
	if os2 := tables["OS/2"]; len(os2) >= 10 {
		license := u16(os2, 8)
		if license&0x202 != 0 {
			return font, fmt.Errorf("La licencia de esta fuente no permite incorporarla al PDF. Elige otra fuente.")
		}
		if len(os2) >= 64 {
			style := u16(os2, 62)
			font.bold, font.italic = style&32 != 0, style&1 != 0
		}
	}
	style := u16(head, 44)
	font.bold = font.bold || style&1 != 0
	font.italic = font.italic || style&2 != 0
	scale := func(v int) int { return int(float64(v) * 1000 / float64(units)) }
	font.Ascent, font.Descent, font.Flags = scale(i16(hhea, 4)), scale(i16(hhea, 6)), 32
	if font.italic {
		font.Flags |= 64
	}
	font.BBox = []int{scale(i16(head, 36)), scale(i16(head, 38)), scale(i16(head, 40)), scale(i16(head, 42))}
	font.Name = "MusicLocalFont"
	name := tables["name"]
	if len(name) >= 6 {
		count, start := u16(name, 2), u16(name, 4)
		if count > 1024 || 6+count*12 > len(name) {
			return fail()
		}
		for j := 0; j < count; j++ {
			at := 6 + j*12
			length, offset := u16(name, at+8), start+u16(name, at+10)
			if u16(name, at+6) != 6 || offset < 0 || offset+length > len(name) {
				continue
			}
			value := name[offset : offset+length]
			if u16(name, at) == 0 || u16(name, at) == 3 {
				codes := []uint16{}
				for k := 0; k+1 < len(value); k += 2 {
					codes = append(codes, binary.BigEndian.Uint16(value[k:]))
				}
				font.Name = string(utf16.Decode(codes))
			} else {
				font.Name = string(value)
			}
			break
		}
	}
	font.Name = regexp.MustCompile(`[^A-Za-z0-9_-]`).ReplaceAllString(font.Name, "")
	if font.Name == "" {
		font.Name = "MusicLocalFont"
	}
	if len(font.Name) > 100 {
		font.Name = font.Name[:100]
	}
	cmap := tables["cmap"]
	if len(cmap) < 4 || u16(cmap, 2) > 256 || 4+u16(cmap, 2)*8 > len(cmap) {
		return fail()
	}
	var mapping []byte
	format := 0
	for j := 0; j < u16(cmap, 2); j++ {
		at := 4 + j*8
		platform, encoding, offset := u16(cmap, at), u16(cmap, at+2), u32(cmap, at+4)
		if !(platform == 0 || platform == 3 && (encoding == 1 || encoding == 10)) || offset < 0 || offset+2 > len(cmap) {
			continue
		}
		f := u16(cmap, offset)
		if f != 4 && f != 12 || f < format {
			continue
		}
		length := u16(cmap, offset+2)
		if f == 12 {
			length = u32(cmap, offset+4)
		}
		if length < 16 || offset+length > len(cmap) {
			continue
		}
		mapping = cmap[offset : offset+length]
		format = f
	}
	if mapping == nil {
		return fail()
	}
	if format == 4 {
		segments := u16(mapping, 6) / 2
		if segments < 1 || 16+segments*8 > len(mapping) {
			return fail()
		}
	}
	if format == 12 {
		groups := u32(mapping, 12)
		if groups < 1 || groups > 65536 || 16+groups*12 > len(mapping) {
			return fail()
		}
	}
	glyph := func(code int) int {
		if format == 12 {
			groups := u32(mapping, 12)
			for j := 0; j < groups; j++ {
				at := 16 + j*12
				if code >= u32(mapping, at) && code <= u32(mapping, at+4) {
					return u32(mapping, at+8) + code - u32(mapping, at)
				}
			}
		} else {
			n := u16(mapping, 6) / 2
			for j := 0; j < n; j++ {
				end, start := u16(mapping, 14+j*2), u16(mapping, 16+n*2+j*2)
				if code < start || code > end {
					continue
				}
				delta, pos := i16(mapping, 16+n*4+j*2), 16+n*6+j*2
				offset := u16(mapping, pos)
				if offset == 0 {
					return (code + delta) & 65535
				}
				g := u16(mapping, pos+offset+(code-start)*2)
				if g > 0 {
					return (g + delta) & 65535
				}
				return 0
			}
		}
		return 0
	}
	font.Widths = make([]int, 224)
	for code := 32; code <= 255; code++ {
		g := glyph(code)
		if g < 0 || g >= glyphs {
			g = 0
		}
		metric := g
		if metric >= metrics {
			metric = metrics - 1
		}
		font.Widths[code-32] = scale(u16(hmtx, metric*4))
	}
	var compressed bytes.Buffer
	z := zlib.NewWriter(&compressed)
	_, _ = z.Write(raw)
	_ = z.Close()
	font.compressed, font.length = compressed.String(), len(raw)
	return font, nil
}
