package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type musicParticipant611 struct {
	Name, Kind, Church, Region, EnrolledAt string
	Attended                               bool
}

func musicParticipants611(db, event map[string]any) []musicParticipant611 {
	rows := []musicParticipant611{}
	for id, raw := range mapField(event, "participants") {
		p, u := legacyMap(raw), legacyUser(db, id)
		name := stringValue(u["fullName"])
		if name == "" {
			name = stringValue(p["fullName"])
		}
		if name == "" {
			name = "Participante " + id
		}
		rows = append(rows, musicParticipant611{name, "Registrado", stringValue(u["church"]), stringValue(u["region"]), stringValue(p["enrolledAt"]), p["attended"] == true})
	}
	for _, raw := range mapField(event, "externalParticipants") {
		p := legacyMap(raw)
		rows = append(rows, musicParticipant611{stringValue(p["fullName"]), "Externo", stringValue(p["church"]), stringValue(p["region"]), stringValue(p["enrolledAt"]), p["attended"] == true})
	}
	sort.SliceStable(rows, func(i, j int) bool { return foldMusic610(rows[i].Name) < foldMusic610(rows[j].Name) })
	return rows
}

func musicAttendance611(attended bool) string {
	if attended {
		return "Sí"
	}
	return "No"
}

func musicCSVCell611(value string) string {
	value = strings.ReplaceAll(value, "\x00", "")
	trimmed := strings.TrimLeft(value, " \t\r\n")
	if strings.ContainsAny(value, "\t\r\n") || trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}

func (s *Server) musicEventExport611(w http.ResponseWriter, r *http.Request, db, actor map[string]any) {
	if personalRequest555(r) || !musicRoster610(legacyRole(actor)) {
		legacyError(w, 403, "La lista de participantes está disponible desde Regional en la vista de gestión.", nil)
		return
	}
	ev := legacyMap(mapField(musicData550(db), "events")[r.URL.Query().Get("eventId")])
	if ev == nil {
		legacyError(w, 404, "Evento no encontrado.", nil)
		return
	}
	rows := musicParticipants611(db, ev)
	w.Header().Set("Cache-Control", "no-store")
	if r.URL.Query().Get("resource") == "event-participants-csv" {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="Participantes_evento.csv"`)
		_, _ = w.Write([]byte{0xef, 0xbb, 0xbf})
		out := csv.NewWriter(w)
		_ = out.Write([]string{"N.º", "Evento", "Fecha del evento", "Nombre", "Tipo", "Iglesia", "Región", "Inscripción", "Asistió"})
		for i, row := range rows {
			cells := []string{fmt.Sprint(i + 1), stringValue(ev["name"]), stringValue(ev["date"]), row.Name, row.Kind, row.Church, row.Region, row.EnrolledAt, musicAttendance611(row.Attended)}
			for j, cell := range cells {
				cells[j] = musicCSVCell611(cell)
			}
			_ = out.Write(cells)
		}
		out.Flush()
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="Participantes_evento.pdf"`)
	_, _ = w.Write(musicEventPDF611(ev, rows))
}

func musicEventPDF611(event map[string]any, rows []musicParticipant611) []byte {
	p := newMusicPDF610()
	page, y := 0, 0.
	var content strings.Builder
	positions := []float64{40, 77, 338, 414, 665, 738}
	widths := []float64{30, 251, 66, 241, 63, 64}
	finish := func() {
		content.WriteString(pdfText610(40, 24, 9, fmt.Sprintf("Total: %d participantes · Página %d", len(rows), page), false))
		p.page(842, 595, content.String(), nil)
	}
	start := func() {
		page++
		content.Reset()
		content.WriteString("0.07 0.18 0.29 rg\n")
		content.WriteString(pdfText610(40, 554, 17, "Lista de inscritos y asistencia", true))
		y = 528
		for _, line := range pdfWrap610(stringValue(event["name"]), 758, 12) {
			content.WriteString(pdfText610(40, y, 12, line, true))
			y -= 16
		}
		content.WriteString(pdfText610(40, y, 10, stringValue(event["date"])+" · "+stringValue(event["place"]), false))
		y -= 29
		for i, label := range []string{"N.º", "Nombre", "Tipo", "Iglesia", "Región", "Asistió"} {
			content.WriteString(pdfText610(positions[i], y, 10, label, true))
		}
		y -= 23
	}
	start()
	for i, row := range rows {
		cells := []string{fmt.Sprint(i + 1), row.Name, row.Kind, row.Church, row.Region, musicAttendance611(row.Attended)}
		lines, count := make([][]string, len(cells)), 1
		for j, cell := range cells {
			lines[j] = pdfWrap610(cell, widths[j], 10)
			if len(lines[j]) > count {
				count = len(lines[j])
			}
		}
		height := float64(count)*13 + 12
		if y-height < 45 {
			finish()
			start()
		}
		for col, entries := range lines {
			for line, entry := range entries {
				content.WriteString(pdfText610(positions[col], y-float64(line)*13, 10, entry, false))
			}
		}
		y -= height
		fmt.Fprintf(&content, "0.85 0.88 0.90 RG 0.4 w 40 %.2f m 802 %.2f l S\n", y+12, y+12)
	}
	if len(rows) == 0 {
		content.WriteString(pdfText610(40, y, 11, "Este evento todavía no tiene participantes inscritos.", false))
	}
	finish()
	return p.bytes()
}
