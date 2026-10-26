package main

import (
	"fmt"
	"time"
)

// Follow-up is informational; it never changes account access.
func followup572(db map[string]any, u map[string]any, scope string, now time.Time) []any {
	rows := []any{}
	if scope != "conquis" || legacyRole(u) == "admin" {
		return rows
	}
	names := map[string]string{"class_grouped": "Agrupada", "guide_major": "Guía Mayor", "guide_master": "Máster", "guide_master_advanced": "Máster Avanzado"}
	for _, pid := range []string{"class_grouped", "guide_major", "guide_master", "guide_master_advanced"} {
		ids := requirements571[pid]
		if !assignedProgram555(db, stringValue(u["id"]), pid) || len(ids) == 0 {
			continue
		}
		p := legacyProgressForRequirement(db, stringValue(u["id"]), pid, false)
		statuses := mapField(p, "status")
		subs := mapField(p, "submissions")
		presented := 0
		deliveries := map[string]time.Time{}
		var latest time.Time
		add := func(id string, at time.Time) {
			if at.After(deliveries[id]) && !at.After(now) {
				deliveries[id] = at
			}
		}
		for _, id := range ids {
			sub := legacyMap(subs[id])
			at := accessTime571(sub["updatedAt"])
			if at.After(latest) && !at.After(now) {
				latest = at
			}
			status := stringValue(statuses[id])
			if status == "in_review" || status == "approved" || status == "needs_changes" {
				presented++
				sent := accessTime571(sub["submittedAt"])
				if sent.IsZero() {
					sent = at
				}
				add(id, sent)
			}
		}
		total, done := 0, 0
		for _, raw := range arrayField(db, "tasks") {
			t := legacyMap(raw)
			program := stringValue(t["programId"])
			if program == "" {
				program = "class_grouped"
			}
			if program != pid || stringValue(t["assigneeId"]) != stringValue(u["id"]) || t["active"] == false || stringValue(t["deletedAt"]) != "" {
				continue
			}
			id := stringValue(t["requirementId"])
			at := accessTime571(t["presentedAt"])
			add(id, at)
			due := accessTime571(t["dueDate"])
			if !due.IsZero() && !due.Before(now.AddDate(0, 0, -14)) && !due.After(now) {
				total++
				if !at.IsZero() || statuses[id] == "in_review" || statuses[id] == "approved" {
					done++
				}
			}
		}
		week, previous := 0, 0
		var lastDelivery time.Time
		for _, at := range deliveries {
			if at.After(lastDelivery) {
				lastDelivery = at
			}
			if !at.Before(now.AddDate(0, 0, -7)) {
				week++
			} else if !at.Before(now.AddDate(0, 0, -14)) {
				previous++
			}
		}
		percent := float64(presented) * 100 / float64(len(ids))
		reasons := []string{}
		stage := 0
		age := accessTime571(u["createdAt"])
		oldEnough := age.IsZero() || !age.After(now.AddDate(0, 0, -7))
		twoWeeks := age.IsZero() || !age.After(now.AddDate(0, 0, -14))
		if percent < 90 && oldEnough {
			if week < 2 {
				stage = 1
				reasons = append(reasons, "Menos de dos entregas en los últimos siete días.")
			}
			if twoWeeks && previous < 2 {
				reasons = append(reasons, "Menos de dos entregas en la semana anterior.")
			}
			if twoWeeks && (week < 2 || previous < 2 || total > 0 && done*100 < total*75) {
				stage = 2
			}
			if twoWeeks && (latest.IsZero() || latest.Before(now.AddDate(0, 0, -14))) {
				reasons = append(reasons, "Sin avances guardados en las últimas dos semanas.")
			}
			if total > 0 && done*100 < total*75 {
				reasons = append(reasons, fmt.Sprintf("Tareas cumplidas: %d de %d con vencimiento reciente.", done, total))
			}
		}
		last, delivery := "", ""
		if !latest.IsZero() {
			last = latest.Format(time.RFC3339)
		}
		if !lastDelivery.IsZero() {
			delivery = lastDelivery.Format(time.RFC3339)
		}
		tutorName := stringValue(legacyUser(db, legacyTutorForProgram(u, pid))["fullName"])
		rows = append(rows, map[string]any{"tutorName": tutorName, "programId": pid, "name": names[pid], "percent": percent, "lastProgressAt": last, "lastDeliveryAt": delivery, "deliveriesWeek": week, "deliveriesPreviousWeek": previous, "tasksDone": done, "tasksTotal": total, "stage": stage, "reasons": reasons})
	}
	return rows
}
