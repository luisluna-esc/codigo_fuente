package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
)

func validMusicPhoto592(source string) bool {
	if !(strings.HasPrefix(source, "data:image/png;base64,") || strings.HasPrefix(source, "data:image/jpeg;base64,")) || len(source) > 2800000 {
		return false
	}
	raw, err := base64.StdEncoding.DecodeString(strings.SplitN(source, ",", 2)[1])
	if err != nil || len(raw) > 2*1024*1024 {
		return false
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || (format != "png" && format != "jpeg") || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 6000 || cfg.Height > 6000 {
		return false
	}
	if format == "png" {
		_, valid := musicImagePixels621(raw)
		return valid
	}
	_, _, err = image.Decode(bytes.NewReader(raw))
	return err == nil
}
func musicTemplateFields592(raw any) ([]any, bool) {
	rows, ok := raw.([]any)
	if !ok || len(rows) > 60 {
		return nil, false
	}
	out := []any{}
	color := regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	for _, r := range rows {
		f, ok := r.(map[string]any)
		if !ok {
			return nil, false
		}
		v := map[string]any{}
		for _, k := range []string{"x", "y", "width", "size"} {
			n, ok := f[k].(float64)
			if !ok || math.IsNaN(n) || math.IsInf(n, 0) {
				return nil, false
			}
			v[k] = n
		}
		x, y, w, size := numericValue(v["x"]), numericValue(v["y"]), numericValue(v["width"]), numericValue(v["size"])
		if x < 0 || x > 95 || y < 0 || y > 95 || w < 5 || x+w > 100.01 || size < 6 || size > 100 {
			return nil, false
		}
		text, font, align, col := stringValue(f["text"]), stringValue(f["font"]), stringValue(f["align"]), stringValue(f["color"])
		if len(text) > 4000 || len(font) < 1 || len(font) > 100 || strings.ContainsAny(font, "\"\\<>;\n\r") || !color.MatchString(col) || (align != "left" && align != "center" && align != "right") {
			return nil, false
		}
		if binding := stringValue(f["automatic622"]); binding != "" {
			if !musicAutomatic622(binding) || text != "{"+binding+"}" {
				return nil, false
			}
			v["automatic622"] = binding
		}
		v["text"] = text
		v["font"] = font
		v["align"] = align
		v["color"] = col
		v["bold"], _ = f["bold"].(bool)
		v["italic"], _ = f["italic"].(bool)
		v["underline"], _ = f["underline"].(bool)
		if id := stringValue(f["fontAsset613"]); id != "" {
			if !regexp.MustCompile(`^font613_[a-f0-9]{64}$`).MatchString(id) {
				return nil, false
			}
			v["fontAsset613"] = id
		}
		if !musicTextSettings618(f, v) {
			return nil, false
		}
		out = append(out, v)
	}
	return out, true
}
func musicHasProgress592(sub map[string]any) bool {
	if strings.TrimSpace(regexp.MustCompile(`<[^>]*>`).ReplaceAllString(stringValue(sub["text"])+stringValue(sub["richText"]), "")) != "" || stringValue(sub["primaryDocumentId"]) != "" || stringValue(sub["completionDate"]) != "" || len(arrayField(sub, "attachmentIds")) > 0 || len(arrayField(sub, "selections")) > 0 {
		return true
	}
	if len(mapField(mapField(sub, "bibleReadings"), "entries")) > 0 {
		return true
	}
	var anyContent func(any) bool
	anyContent = func(v any) bool {
		switch x := v.(type) {
		case string:
			return strings.TrimSpace(x) != ""
		case bool:
			return x
		case map[string]any:
			for _, child := range x {
				if anyContent(child) {
					return true
				}
			}
		}
		return false
	}
	return anyContent(sub["readingReport"])
}

// Each requirement contributes its current stage once in the evaluation period.
// Points belong to its leader, including when a tutor performs the review.
func musicDailyItems592(db any, day string) []legacyDailyActivityItem {
	out := []legacyDailyActivityItem{}
	hour := legacyDailyRecognitionHour(db)
	for id, u := range legacyNonAdminActiveUsers(db) {
		if !musicAssigned550(legacyMap(db), id) {
			continue
		}
		p := legacyProgressForRequirement(db, id, musicProgram553, false)
		statuses := mapField(p, "status")
		subs := mapField(p, "submissions")
		item := legacyDailyActivityItem{UserID: id, FullName: legacyName(u), Role: legacyRole(u)}
		for _, rid := range musicRequirements550 {
			status := stringValue(statuses[rid])
			sub := legacyMap(subs[rid])
			points := 0
			at := stringValue(sub["updatedAt"])
			switch status {
			case "approved":
				points = 3
				at = approvalAt555(sub).Format(time.RFC3339Nano)
			case "in_review":
				points = 2
				if s := stringValue(sub["submittedAt"]); s != "" {
					at = s
				}
			case "in_progress", "needs_changes", "assigned":
				if musicHasProgress592(sub) {
					points = 1
				}
			}
			if points == 0 || legacyAuditEvaluationDay(at, hour) != day {
				continue
			}
			item.Count++
			item.Points += points
			item.Reports++
			if points == 2 {
				item.Submissions++
			}
			if points == 3 {
				item.Approvals++
				item.Reviews++
			}
		}
		if item.Points > 0 {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Points == out[j].Points {
			return out[i].UserID < out[j].UserID
		}
		return out[i].Points > out[j].Points
	})
	return out
}
func ensureMusicDaily592(st *legacyStore, now time.Time) bool {
	if st == nil || st.State.Database == nil || now.Local().Hour() != legacyDailyRecognitionHour(st.State.Database) {
		return false
	}
	day := now.Local().Format("2006-01-02")
	items := musicDailyItems592(st.State.Database, day)
	before := len(st.Notifications)
	for i, item := range items {
		title, message := "Puntos de Música", legacyDailyPersonalMessage(item)
		if i == 0 {
			title = "¡Felicitaciones! · Música"
			message = legacyDailyWinnerMessage(day, item)
		}
		addLegacyNotification(st, legacyNotification{UserID: item.UserID, Type: "daily_activity", Title: title, Message: message, ActorID: "system", ActorName: "Música", SourceOperationID: "music_daily_activity_" + day, CreatedAt: now.UTC().Format(time.RFC3339Nano), ProgramID: musicProgram553})
	}
	return len(st.Notifications) > before
}
