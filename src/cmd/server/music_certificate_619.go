package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// A reset keeps design assets, never issued PDFs, people or activity history.
func musicResetDesign619(db map[string]any, files []any) (map[string]any, []any) {
	if db == nil {
		db = map[string]any{}
	}
	d := musicData550(db)
	templates := cloneMap610(mapField(d, "templates"))
	for _, raw := range templates {
		delete(legacyMap(raw), "updatedBy")
	}
	next := map[string]any{"users": []any{}, "modules": []any{}, "settings": map[string]any{}, "music550": map[string]any{"registrationSequence622": d["registrationSequence622"], "modalities622": cloneAny(d["modalities622"]), "schema610": true, "templates": templates, "publicQRURL613": d["publicQRURL613"]}}
	fonts := []any{}
	for _, raw := range files {
		f := legacyMap(raw)
		if f["kind"] != "music_font613" {
			continue
		}
		copy := cloneMap610(f)
		delete(copy, "ownerId")
		delete(copy, "createdAt")
		fonts = append(fonts, copy)
	}
	return next, fonts
}

func musicDate619(value string) string {
	if len(value) < 10 {
		return value
	}
	date, err := time.Parse("2006-01-02", value[:10])
	if err != nil {
		return value
	}
	months := []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}
	return fmt.Sprintf("%d de %s de %d", date.Day(), months[int(date.Month())-1], date.Year())
}

// Current copies follow the configured server; the unique token never changes.
func musicSyncQR619(db, c map[string]any) bool {
	value := ""
	if base := musicQRBase613(db); base != "" {
		value = base + "/api/music/certificate?token=" + stringValue(c["token"])
	}
	if stringValue(c["verificationURL610"]) == value {
		return false
	}
	c["verificationURL610"] = value
	return true
}

// CSS and PDF rotate clockwise about the same image center. The rectangle's
// complete rotated extent must remain on the sheet, including when resizing.
func musicImageBounds619(im map[string]any, t map[string]any) bool {
	w, h := musicPaper617(t)
	x, y := numericValue(im["x"])*w/100, numericValue(im["y"])*h/100
	iw, ih := numericValue(im["width"])*w/100, numericValue(im["height"])*h/100
	a := numericValue(im["rotation619"]) * math.Pi / 180
	bw, bh := math.Abs(iw*math.Cos(a))+math.Abs(ih*math.Sin(a)), math.Abs(iw*math.Sin(a))+math.Abs(ih*math.Cos(a))
	cx, cy := x+iw/2, y+ih/2
	return cx-bw/2 >= -.1 && cy-bh/2 >= -.1 && cx+bw/2 <= w+.1 && cy+bh/2 <= h+.1
}

func musicImageCommands619(im, t map[string]any, name, opacity string) string {
	w, h := musicPaper617(t)
	iw, ih := numericValue(im["width"])*w/100, numericValue(im["height"])*h/100
	cx := numericValue(im["x"])*w/100 + iw/2
	cy := h - numericValue(im["y"])*h/100 - ih/2
	a := -numericValue(im["rotation619"]) * math.Pi / 180
	co, si := math.Cos(a), math.Sin(a)
	return fmt.Sprintf("q %s%.6f %.6f %.6f %.6f %.3f %.3f cm %.3f 0 0 %.3f %.3f %.3f cm /%s Do Q\n", opacity, co, si, -si, co, cx, cy, iw, ih, -iw/2, -ih/2, name)
}

func musicSignature619(f map[string]any) map[string]any {
	copy := cloneMap610(f)
	copy["explicitLines619"] = true
	return copy
}

// Plain signature lines fit horizontally without adding unwanted line breaks.
func musicExplicitLines619(text string) []string {
	return strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
}
