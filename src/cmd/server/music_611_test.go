package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func Test611PersonalReminderBelongsToAssignedReviewer(t *testing.T) {
	_, db := fixture550(t)
	musicEnsure610(db)
	owner := legacyUser(db, "p")
	p := reviewProgressContainer(db, "p", musicProgram553)
	ensureMapField(p, "status")["LM-PRE-1"] = "in_review"
	ensureMapField(p, "submissions")["LM-PRE-1"] = map[string]any{"submittedAt": time.Now().Add(-73 * time.Hour).UTC().Format(time.RFC3339Nano)}
	count := func(actor string) int {
		return int(numericValue(musicFollowupView610(db, legacyUser(db, actor))["pendingCount"]))
	}
	if count("tutor") != 1 || count("admin") != 0 || count("departamental") != 0 || count("p") != 0 {
		t.Fatal("a personal reminder leaked to unrelated accounts")
	}
	owner["tutorByProgram"] = map[string]any{musicProgram553: "admin"}
	if count("admin") != 1 || count("tutor") != 0 {
		t.Fatal("per-program reassignment did not move the reminder")
	}
	owner["tutorByProgram"] = map[string]any{musicProgram553: ""}
	if count("tutor") != 0 {
		t.Fatal("explicitly unassigned tutor inherited the legacy assignment")
	}
	owner["createdBy"] = "admin"
	if count("admin") != 1 {
		t.Fatal("responsible creator without a tutor has no reminder")
	}
	ensureMapField(p, "status")["LM-PRE-1"] = "approved"
	if count("admin") != 0 {
		t.Fatal("approved requirements still cause a reminder")
	}
}

func Test611RemindersOnlyNotifyAssignedTutor(t *testing.T) {
	s, db := fixture550(t)
	musicEnsure610(db)
	p := reviewProgressContainer(db, "p", musicProgram553)
	ensureMapField(p, "status")["LM-PRE-1"] = "in_review"
	ensureMapField(p, "submissions")["LM-PRE-1"] = map[string]any{"submittedAt": time.Now().Add(-73 * time.Hour).UTC().Format(time.RFC3339Nano)}
	st, _ := s.loadLegacy()
	st.Notifications = nil
	st.State.Database = db
	musicFollowupNotifications610(&st)
	if len(st.Notifications) != 1 || st.Notifications[0].UserID != "tutor" || st.Notifications[0].OwnerID != "p" {
		t.Fatalf("unexpected recipients: %#v", st.Notifications)
	}
}

func exportFixture611(t *testing.T) *Server {
	t.Helper()
	s, db := fixture550(t)
	musicEnsure610(db)
	ensureMapField(musicData550(db), "events")["ev"] = map[string]any{
		"id": "ev", "name": "Encuentro de Música", "date": "2026-10-20", "place": "Central",
		"participants":         map[string]any{"p": map[string]any{"attended": true, "enrolledAt": "2026-10-01"}},
		"externalParticipants": map[string]any{"ext": map[string]any{"fullName": "=HYPERLINK(\"example\")", "attended": false}},
	}
	st, _ := s.loadLegacy()
	st.State.Database = db
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	return s
}

func resource611(s *Server, actor, query string, personal bool) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", "/api/music/v610?actorId="+actor+"&"+query, nil)
	r.Header.Set("Authorization", "Bearer "+actor)
	if personal {
		r.Header.Set("X-SGL-Personal", "1")
	}
	w := httptest.NewRecorder()
	s.securityHeaders(http.HandlerFunc(s.music610)).ServeHTTP(w, r)
	return w
}

func Test611EventExportIncludesAttendanceAndExternalParticipants(t *testing.T) {
	s := exportFixture611(t)
	w := resource611(s, "regional", "resource=event-participants-csv&eventId=ev", false)
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	data := bytes.TrimPrefix(w.Body.Bytes(), []byte{0xef, 0xbb, 0xbf})
	rows, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil || len(rows) != 3 {
		t.Fatal(err, rows)
	}
	attendance := map[string]bool{}
	for _, row := range rows[1:] {
		attendance[row[8]] = true
		if row[4] == "Externo" && !strings.HasPrefix(row[3], "'=") {
			t.Fatal("unsafe spreadsheet formula", row)
		}
	}
	if !attendance["Sí"] || !attendance["No"] {
		t.Fatal("attendance missing", rows)
	}
	for _, actor := range []string{"p", "tutor"} {
		if w := resource611(s, actor, "resource=event-participants-csv&eventId=ev", false); w.Code != 403 {
			t.Fatal(actor, w.Code)
		}
	}
	if w := resource611(s, "admin", "resource=event-participants-pdf&eventId=ev", true); w.Code != 403 {
		t.Fatal("personal view exported the roster")
	}
	if w := resource611(s, "admin", "resource=event-participants-csv&eventId=missing", false); w.Code != 404 {
		t.Fatal("missing event accepted")
	}
	w = resource611(s, "regional", "resource=event-participants-pdf&eventId=ev", false)
	if w.Code != 200 || !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF-")) {
		t.Fatal(w.Code)
	}
}

func Test611EventPDFPaginationAndVoteOnCustomCertificate(t *testing.T) {
	rows := []musicParticipant611{}
	for i := 0; i < 87; i++ {
		rows = append(rows, musicParticipant611{Name: fmt.Sprintf("Participante %02d Pérez", i), Kind: "Registrado", Church: "Iglesia Central", Region: "1", Attended: i%2 == 0})
	}
	pdf := musicEventPDF611(map[string]any{"name": "Encuentro de Música", "date": "2026-10-20", "place": "Central"}, rows)
	if bytes.Count(pdf, []byte("/Type /Page ")) < 3 || !bytes.Contains(pdf, []byte("Participante 86")) {
		t.Fatal("participants were truncated")
	}
	c := map[string]any{"kind": "ministry", "fullName": "Coro Armonía", "title": "Ministerio de Música", "registration": map[string]any{"boardVote": "32-2026"}}
	field := map[string]any{"text": "{nombre}", "x": 10., "y": 40., "width": 80., "size": 24.}
	template := map[string]any{"fields": []any{field}}
	cert := musicCertificateBytes610(c, template)
	if !bytes.Contains(cert, []byte("Voto de Junta")) || !bytes.Contains(cert, []byte("32-2026")) {
		t.Fatal("custom template lost the vote")
	}
	field["text"] = "Voto de Junta: {voto_junta}"
	cert = musicCertificateBytes610(c, template)
	if bytes.Count(cert, []byte("32-2026")) != 1 {
		t.Fatal("vote duplicated or variable unresolved")
	}
	if out := os.Getenv("MUSIC_QA_PDF_DIR"); out != "" {
		if err := os.MkdirAll(out, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(out+"/Participantes.pdf", pdf, 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(out+"/Certificado_voto.pdf", cert, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func Test611ManagementHistoryUnavailableToApplicants(t *testing.T) {
	s := exportFixture611(t)
	if w := resource611(s, "p", "resource=archive&year=2026", false); w.Code != 403 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w := resource611(s, "admin", "resource=archive&year=2026", true); w.Code != 403 {
		t.Fatal(w.Code, w.Body.String())
	}
}
