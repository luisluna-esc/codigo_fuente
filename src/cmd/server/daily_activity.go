package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type legacyDailyActivityItem struct {
	UserID      string   `json:"userId"`
	FullName    string   `json:"fullName"`
	Role        string   `json:"role"`
	Count       int      `json:"count"`
	Points      int      `json:"points"`
	Reports     int      `json:"reports"`
	Modules     int      `json:"modules"`
	Chats       int      `json:"chats"`
	Submissions int      `json:"submissions"`
	Approvals   int      `json:"approvals"`
	Reviews     int      `json:"reviews"`
	Actions     []string `json:"actions,omitempty"`
}

type legacyActivityScore struct {
	Points      int
	Reports     int
	Modules     int
	Chats       int
	Submissions int
	Approvals   int
	Reviews     int
}

type legacyTrackActivityRequest struct {
	UserID string `json:"userId"`
	Action string `json:"action"`
	Entity string `json:"entity"`
	Detail string `json:"detail"`
	Module string `json:"module"`
	At     string `json:"at"`
}

func legacyDailyRecognitionHour(db any) int {
	settings := mapField(legacyMap(db), "settings")
	keys := []string{"dailyRecognitionHour", "interactionRecognitionHour", "activityRecognitionHour"}
	for _, key := range keys {
		value := strings.TrimSpace(stringValue(settings[key]))
		if value == "" {
			continue
		}
		if dot := strings.Index(value, "."); dot >= 0 {
			value = value[:dot]
		}
		if n, err := strconv.Atoi(value); err == nil && n >= 0 && n <= 23 {
			return n
		}
	}
	return 18
}

func legacyDailyActivityDate(now time.Time, hour int) string {
	local := now.Local()
	if local.Hour() < hour {
		local = local.AddDate(0, 0, -1)
	}
	return local.Format("2006-01-02")
}

func legacyNonAdminActiveUsers(db any) map[string]map[string]any {
	users := map[string]map[string]any{}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		id := strings.TrimSpace(stringValue(u["id"]))
		role := strings.TrimSpace(legacyRole(u))
		if id == "" || !legacyActive(u) || role == "admin" {
			continue
		}
		users[id] = u
	}
	return users
}

func legacyActivityScoreFor(action, entity, detail string) legacyActivityScore {
	text := strings.ToLower(strings.TrimSpace(strings.Join([]string{action, entity, detail}, " ")))
	if text == "" {
		return legacyActivityScore{}
	}
	ignored := []string{"inicio de sesi", "cierre de sesi", "login", "logout", "sincroniz", "heartbeat", "notific", "aviso", "sesion remota", "sesión remota", "reconocimiento diario programado"}
	for _, part := range ignored {
		if strings.Contains(text, part) {
			return legacyActivityScore{}
		}
	}

	score := legacyActivityScore{Points: 1}
	add := func(points int) {
		if points > score.Points {
			score.Points = points
		}
	}
	if strings.Contains(text, "chat") || strings.Contains(text, "mensaje") || strings.Contains(text, "conversaci") {
		score.Chats++
		add(2)
	}
	if strings.Contains(text, "naveg") || strings.Contains(text, "modulo") || strings.Contains(text, "módulo") || strings.Contains(text, "pantalla") || strings.Contains(text, "opcion") || strings.Contains(text, "opción") || strings.Contains(text, "abrir") || strings.Contains(text, "consult") {
		score.Modules++
		add(1)
	}
	if strings.Contains(text, "informe") || strings.Contains(text, "reporte") || strings.Contains(text, "pdf") || strings.Contains(text, "word") || strings.Contains(text, "drive") || strings.Contains(text, "evidencia") || strings.Contains(text, "archivo adjunt") {
		score.Reports++
		add(3)
	}
	if strings.Contains(text, "escrit") || strings.Contains(text, "redact") || strings.Contains(text, "guardar avance") || strings.Contains(text, "avance guardado") || strings.Contains(text, "informe guardado") || strings.Contains(text, "archivo adjunt") {
		score.Reports++
		add(5)
	}
	if strings.Contains(text, "requisito enviado") || strings.Contains(text, "enviado a revis") || strings.Contains(text, "enviar a revis") || strings.Contains(text, "enviar informe") || strings.Contains(text, "informe enviado") || strings.Contains(text, "envio") || strings.Contains(text, "envío") || strings.Contains(text, "presentado") {
		score.Submissions++
		score.Reports++
		add(6)
	}
	if strings.Contains(text, "aprob") {
		score.Approvals++
		score.Reviews++
		score.Reports++
		add(8)
	}
	if strings.Contains(text, "revis") || strings.Contains(text, "correcci") || strings.Contains(text, "correg") || strings.Contains(text, "rechaz") || strings.Contains(text, "devoluci") {
		score.Reviews++
		add(4)
	}
	if strings.Contains(text, "tarea") {
		add(3)
	}
	if strings.Contains(text, "asistencia") {
		add(3)
	}
	if strings.Contains(text, "material") {
		add(3)
	}
	if strings.Contains(text, "planific") {
		add(4)
	}
	if strings.Contains(text, "modificar") || strings.Contains(text, "actualiz") || strings.Contains(text, "editar") {
		add(3)
	}
	return score
}

func legacyCountableInteraction(action string) bool {
	return legacyActivityScoreFor(action, "", "").Points > 0
}

func legacyAuditDay(at string) string {
	at = strings.TrimSpace(at)
	if at == "" {
		return ""
	}
	if t, err := time.Parse(time.RFC3339Nano, at); err == nil {
		return t.Local().Format("2006-01-02")
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04:05", at, time.Local); err == nil {
		return t.Local().Format("2006-01-02")
	}
	if len(at) >= 10 {
		return at[:10]
	}
	return ""
}

// legacyAuditEvaluationDay asigna cada interacción al periodo que termina en la
// próxima hora diaria de reconocimiento. Ejemplo con corte a las 18:00:
// 17:30 del día 5 pertenece al periodo 5; 18:01 del día 5 ya pertenece al 6.
// Así, una vez emitida la felicitación del día, los puntos nuevos no modifican
// ese reconocimiento y quedan acumulados para el siguiente periodo.
func legacyAuditEvaluationDay(at string, hour int) string {
	at = strings.TrimSpace(at)
	if at == "" {
		return ""
	}
	var parsed time.Time
	var ok bool
	if t, err := time.Parse(time.RFC3339Nano, at); err == nil {
		parsed, ok = t.Local(), true
	} else if t, err := time.ParseInLocation("2006-01-02 15:04:05", at, time.Local); err == nil {
		parsed, ok = t.Local(), true
	}
	if !ok {
		return legacyAuditDay(at)
	}
	if hour < 0 || hour > 23 {
		hour = 18
	}
	if parsed.Hour() >= hour {
		parsed = parsed.AddDate(0, 0, 1)
	}
	return parsed.Format("2006-01-02")
}

func legacyDailyActivityItems(db any, day string) []legacyDailyActivityItem {
	users := legacyNonAdminActiveUsers(db)
	hour := legacyDailyRecognitionHour(db)
	items := map[string]*legacyDailyActivityItem{}
	actions := map[string]map[string]bool{}
	for _, raw := range arrayField(legacyMap(db), "audit") {
		entry := legacyMap(raw)
		actorID := strings.TrimSpace(stringValue(entry["actorId"]))
		if actorID == "" {
			actorID = strings.TrimSpace(stringValue(entry["userId"]))
		}
		if actorID == "" || users[actorID] == nil {
			continue
		}
		action := strings.TrimSpace(stringValue(entry["action"]))
		entity := strings.TrimSpace(stringValue(entry["entity"]))
		detail := strings.TrimSpace(stringValue(entry["detail"]))
		module := stringValue(entry["module"])
		if module == "" {
			module = stringValue(entry["programId"])
		}
		if module == musicProgram553 || strings.Contains(entity+" "+detail, "LM-") || module == "" && stringValue(users[actorID]["primaryProgramId"]) == musicProgram553 {
			continue
		}
		score := legacyActivityScoreFor(action, entity, detail)
		if score.Points <= 0 || legacyAuditEvaluationDay(stringValue(entry["at"]), hour) != day {
			continue
		}
		item := items[actorID]
		if item == nil {
			u := users[actorID]
			item = &legacyDailyActivityItem{UserID: actorID, FullName: legacyName(u), Role: legacyRole(u)}
			items[actorID] = item
		}
		item.Count++
		item.Points += score.Points
		item.Reports += score.Reports
		item.Modules += score.Modules
		item.Chats += score.Chats
		item.Submissions += score.Submissions
		item.Approvals += score.Approvals
		item.Reviews += score.Reviews
		if actions[actorID] == nil {
			actions[actorID] = map[string]bool{}
		}
		if action != "" && len(actions[actorID]) < 8 {
			actions[actorID][action] = true
		}
	}
	out := make([]legacyDailyActivityItem, 0, len(items))
	for id, item := range items {
		sample := make([]string, 0, len(actions[id]))
		for action := range actions[id] {
			sample = append(sample, action)
		}
		sort.Strings(sample)
		item.Actions = sample
		out = append(out, *item)
	}
	return out
}

func legacyTopDailyInteractions(db any, day string, limit int) []legacyDailyActivityItem {
	if limit <= 0 {
		limit = 5
	}
	items := legacyDailyActivityItems(db, day)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Points != items[j].Points {
			return items[i].Points > items[j].Points
		}
		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count
		}
		return strings.ToLower(items[i].FullName) < strings.ToLower(items[j].FullName)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

func legacyTopDailyReports(db any, day string, limit int) []legacyDailyActivityItem {
	if limit <= 0 {
		limit = 5
	}
	items := legacyDailyActivityItems(db, day)
	filtered := make([]legacyDailyActivityItem, 0, len(items))
	for _, item := range items {
		if item.Reports > 0 || item.Submissions > 0 || item.Approvals > 0 || item.Reviews > 0 {
			filtered = append(filtered, item)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		left := filtered[i].Reports*10 + filtered[i].Submissions*12 + filtered[i].Approvals*15 + filtered[i].Reviews*4
		right := filtered[j].Reports*10 + filtered[j].Submissions*12 + filtered[j].Approvals*15 + filtered[j].Reviews*4
		if left != right {
			return left > right
		}
		if filtered[i].Points != filtered[j].Points {
			return filtered[i].Points > filtered[j].Points
		}
		return strings.ToLower(filtered[i].FullName) < strings.ToLower(filtered[j].FullName)
	})
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered
}

func legacyDailyActivityMessage(day string, top []legacyDailyActivityItem) string {
	parts := make([]string, 0, len(top))
	for i, item := range top {
		name := strings.TrimSpace(item.FullName)
		if name == "" || name == "<nil>" {
			name = item.UserID
		}
		parts = append(parts, fmt.Sprintf("%d. %s: %d pts", i+1, name, item.Points))
	}
	if len(parts) == 0 {
		return fmt.Sprintf("Resumen del %s: todavía no hay actividad suficiente.", day)
	}
	return fmt.Sprintf("Resumen interno del %s: %s.", day, strings.Join(parts, " | "))
}

func legacyDailyWinnerMessage(day string, winner legacyDailyActivityItem) string {
	return fmt.Sprintf("🏆 Hoy ganaste %d puntos. ¡Excelente avance, sigue adelante!", winner.Points)
}

func legacyDailyPersonalMessage(item legacyDailyActivityItem) string {
	return fmt.Sprintf("Hoy sumaste %d puntos. Cada avance cuenta. Sigue adelante.", item.Points)
}

func (s *Server) ensureLegacyDailyActivityNotifications(st *legacyStore, now time.Time) bool {
	if st == nil || st.State.Database == nil {
		return false
	}
	musicChanged := ensureMusicDaily592(st, now)
	hour := legacyDailyRecognitionHour(st.State.Database)
	local := now.Local()
	// La felicitación es una ventana horaria exacta: fuera de la hora configurada
	// no se crea ningún aviso atrasado ni adelantado.
	if local.Hour() != hour {
		return false
	}
	day := local.Format("2006-01-02")
	top := legacyTopDailyInteractions(st.State.Database, day, 5)
	if len(top) == 0 {
		return musicChanged
	}
	before := len(st.Notifications)
	nowText := now.UTC().Format(time.RFC3339Nano)
	source := "daily_activity_" + day
	winner := top[0]
	allItems := legacyDailyActivityItems(st.State.Database, day)
	activityByUser := make(map[string]legacyDailyActivityItem, len(allItems))
	for _, item := range allItems {
		if item.UserID != "" && item.Points > 0 {
			activityByUser[item.UserID] = item
		}
	}
	for id, u := range legacyNonAdminActiveUsers(st.State.Database) {
		primary := strings.TrimSpace(stringValue(u["primaryProgramId"]))
		if musicAssigned550(legacyMap(st.State.Database), id) {
			other := ""
			for _, raw := range arrayField(legacyMap(st.State.Database), "modules") {
				m := legacyMap(raw)
				pid := stringValue(m["id"])
				if pid != musicProgram553 && assignedProgram555(legacyMap(st.State.Database), id, pid) {
					other = pid
					break
				}
			}
			if other == "" {
				continue
			}
			if primary == "" || primary == musicProgram553 {
				primary = other
			}
		}
		item, ok := activityByUser[id]
		if !ok || item.Points <= 0 {
			continue
		}
		title := "Puntos de hoy"
		message := legacyDailyPersonalMessage(item)
		if winner.UserID == id {
			title = "¡Felicitaciones!"
			message = legacyDailyWinnerMessage(day, item)
		}
		addLegacyNotification(st, legacyNotification{
			UserID: id, Type: "daily_activity", Title: title,
			Message: message, ActorID: "system", ActorName: "SGL", SourceOperationID: source, CreatedAt: nowText,
			ProgramID: primary,
		})
	}
	return len(st.Notifications) > before || musicChanged
}

// expireLegacyDailyActivityNotifications evita que una felicitación diaria quede
// pendiente y vuelva a mostrarse fuera de la hora para la que fue programada.
// El aviso del día solo puede permanecer sin leer durante esa ventana horaria.
func expireLegacyDailyActivityNotifications(st *legacyStore, now time.Time) bool {
	if st == nil || st.State.Database == nil {
		return false
	}
	hour := legacyDailyRecognitionHour(st.State.Database)
	local := now.Local()
	today := local.Format("2006-01-02")
	activeWindow := local.Hour() == hour
	readAt := now.UTC().Format(time.RFC3339Nano)
	changed := false
	for i := range st.Notifications {
		n := &st.Notifications[i]
		if n == nil || n.DeletedAt != "" || n.ReadAt != "" || (n.Type != "daily_activity" && n.Type != "daily_activity_winner") {
			continue
		}
		sourceDay := ""
		if strings.HasPrefix(n.SourceOperationID, "music_daily_activity_") {
			sourceDay = strings.TrimPrefix(n.SourceOperationID, "music_daily_activity_")
		}
		if strings.HasPrefix(n.SourceOperationID, "daily_activity_") {
			sourceDay = strings.TrimPrefix(n.SourceOperationID, "daily_activity_")
		}
		if !activeWindow || sourceDay == "" || sourceDay != today {
			n.ReadAt = readAt
			changed = true
		}
	}
	return changed
}

func (s *Server) legacyDailyActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	userID, _, ok := s.notificationActor(r, r.URL.Query().Get("userId"), &st)
	if !ok || userID == "" {
		legacyError(w, 403, "No autorizado para consultar interacción diaria.", nil)
		return
	}
	now := time.Now()
	if s.ensureLegacyDailyActivityNotifications(&st, now) {
		if err := s.saveLegacy(st); err != nil {
			legacyError(w, 500, err.Error(), nil)
			return
		}
	}
	day := strings.TrimSpace(r.URL.Query().Get("date"))
	if day == "" {
		day = legacyDailyActivityDate(now, legacyDailyRecognitionHour(st.State.Database))
	}
	limit := 5
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 20 {
			limit = n
		}
	}
	top := legacyTopDailyInteractions(st.State.Database, day, limit)
	topReports := legacyTopDailyReports(st.State.Database, day, limit)
	if legacyRole(legacyUser(st.State.Database, userID)) == "aspirant" {
		top = []legacyDailyActivityItem{}
		topReports = []legacyDailyActivityItem{}
		for _, item := range legacyDailyActivityItems(st.State.Database, day) {
			if item.UserID == userID {
				top = append(top, item)
				if item.Reports > 0 || item.Submissions > 0 || item.Approvals > 0 || item.Reviews > 0 {
					topReports = append(topReports, item)
				}
			}
		}
	}
	if musicRequest553(r) {
		top = []legacyDailyActivityItem{}
		for _, item := range musicDailyItems592(st.State.Database, day) {
			actor := legacyUser(st.State.Database, userID)
			owner := legacyUser(st.State.Database, item.UserID)
			if !canReadOwner555(legacyMap(st.State.Database), actor, owner, musicProgram553) {
				continue
			}
			top = append(top, item)
		}
		if len(top) > limit {
			top = top[:limit]
		}
		topReports = top
	}
	jsonOut(w, 200, map[string]any{
		"ok": true, "date": day, "recognitionHour": legacyDailyRecognitionHour(st.State.Database),
		"top": top, "topReports": topReports, "winner": func() any {
			if len(top) == 0 {
				return nil
			}
			return top[0]
		}(),
		"parameters": []string{"escribir informes", "guardar informes", "enviar a revisión", "aprobar", "solicitar corrección", "navegar por módulos", "chat", "tareas", "asistencia", "materiales"},
		"excludes":   []string{"admin", "login", "logout", "sincronización automática", "heartbeat", "avisos pasivos"},
		"serverTime": now.UTC().Format(time.RFC3339Nano),
	})
}

func (s *Server) legacyActivityTrack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	var req legacyTrackActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		legacyError(w, 400, "Solicitud de actividad inválida.", nil)
		return
	}
	req.UserID = strings.TrimSpace(req.UserID)
	req.Action = strings.TrimSpace(req.Action)
	if req.Action == "" {
		jsonOut(w, 200, map[string]any{"ok": true, "tracked": false})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	userID, role, ok := s.notificationActor(r, req.UserID, &st)
	if !ok || userID == "" {
		legacyError(w, 403, "No autorizado para registrar actividad.", nil)
		return
	}
	if role == "admin" || legacyActivityScoreFor(req.Action, req.Entity, req.Detail).Points <= 0 {
		jsonOut(w, 200, map[string]any{"ok": true, "tracked": false})
		return
	}
	db := legacyMap(st.State.Database)
	if db == nil {
		legacyError(w, 500, "Base de datos inválida.", nil)
		return
	}
	at := strings.TrimSpace(req.At)
	if at == "" {
		at = time.Now().UTC().Format(time.RFC3339Nano)
	}
	user := legacyUser(st.State.Database, userID)
	entry := map[string]any{
		"id":        "audit_" + token(10),
		"at":        at,
		"actorId":   userID,
		"actorName": legacyName(user),
		"action":    req.Action,
		"entity":    firstNonEmpty(req.Entity, req.Module, "activity"),
		"detail":    req.Detail,
	}
	audit := arrayField(db, "audit")
	audit = append([]any{entry}, audit...)
	if len(audit) > 2500 {
		audit = audit[:2500]
	}
	db["audit"] = audit
	db["updatedAt"] = time.Now().UTC().Format(time.RFC3339Nano)
	st.State.Database = db
	if err := s.saveLegacy(st); err != nil {
		legacyError(w, 500, err.Error(), nil)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "tracked": true})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
