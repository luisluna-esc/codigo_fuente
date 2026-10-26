package main

import (
	"net/http"
	"net/url"
	"strings"
)

// Contact permission is independent of hierarchy and evaluated on every request.
func musicWhatsappEligible598(db, actor map[string]any) bool {
	if actor == nil || !legacyActive(actor) || actor["musicAccessRemoved"] == true || !managementProgram555(db, actor, musicProgram553) {
		return false
	}
	switch legacyRole(actor) {
	case "admin", "departamental", "secretaria", "coordinador_musica", "administrativo", "regional":
		return true
	}
	return false
}
func musicWhatsapp598(db, actor map[string]any) bool {
	return musicPermit602(db, actor, "whatsapp") && musicWhatsappEligible598(db, actor) && containsString(arrayField(mapField(musicData550(db), "interface598"), "whatsappUsers"), stringValue(actor["id"]))
}
func musicContactURL598(db, actor, owner map[string]any) (string, string) {
	if !musicWhatsapp598(db, actor) || owner == nil || !legacyActive(owner) || !(musicScope550(db, actor, owner) || musicSeeTutor560(db, actor, owner)) {
		return "", "No tienes permiso para contactar esta cuenta."
	}
	raw := strings.TrimSpace(stringValue(owner["phone"]))
	if !strings.HasPrefix(raw, "+") {
		return "", "Registra el teléfono con + y código de país antes de abrir WhatsApp."
	}
	phone := ""
	for _, c := range raw[1:] {
		if c >= '0' && c <= '9' {
			phone += string(c)
		} else if !strings.ContainsRune(" ()-", c) {
			return "", "El teléfono contiene caracteres inválidos."
		}
	}
	if len(phone) < 8 || len(phone) > 15 || phone[0] == '0' {
		return "", "Revisa el teléfono y su código de país."
	}
	message := "Hola, " + stringValue(owner["fullName"]) + ". Soy " + stringValue(actor["fullName"]) + ", del Sistema de Gestión de Música. Quisiera comunicarme contigo sobre tu formación."
	if custom := strings.TrimSpace(stringValue(mapField(musicData550(db), "permissions602")["whatsappMessage"])); custom != "" {
		message = strings.NewReplacer("{nombre}", stringValue(owner["fullName"]), "{responsable}", stringValue(actor["fullName"])).Replace(custom)
	}
	return "https://wa.me/" + phone + "?text=" + url.QueryEscape(message), ""
}
func (s *Server) musicContact598(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", 405)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, 500, "No se pudo consultar el contacto.", nil)
		return
	}
	db := legacyMap(st.State.Database)
	actor, ok := s.musicActor550(r, db, r.URL.Query().Get("actorId"))
	if !ok || personalRequest555(r) {
		legacyError(w, 403, "Sesión no autorizada.", nil)
		return
	}
	link, reason := musicContactURL598(db, actor, legacyUser(db, r.URL.Query().Get("ownerId")))
	if reason != "" {
		legacyError(w, 403, reason, nil)
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "url": link})
}
