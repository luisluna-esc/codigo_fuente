package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func annualDocument558(choice string) (map[string]any, map[string]any) {
	data := []byte("%PDF-1.4\n1 0 obj << /Type /Catalog >> endobj\n%%EOF")
	sub := map[string]any{"submissionMode": "document", "deliveryMode": "document", "documentMode": true, "documentPageCount": 1, "primaryDocumentId": "annual-pdf", "attachmentIds": []any{"annual-pdf"}, "bibleReadings": map[string]any{"kind": "music_devotion", "choice": choice, "entries": map[string]any{"2026-01-01": map[string]any{"completed": true}}}}
	file := map[string]any{"id": "annual-pdf", "ownerId": "p", "programId": musicProgram553, "requirementId": "LM-NUT-1", "kind": "requirement_document", "type": "application/pdf", "name": "Certificado.pdf", "size": len(data), "dataUrl": "data:application/pdf;base64," + base64.StdEncoding.EncodeToString(data)}
	return sub, file
}
func submitPDF558(t *testing.T, s *Server, sub map[string]any, files []any) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(legacyReviewSubmitRequest{ActorID: "p", OwnerID: "p", ProgramID: musicProgram553, RequirementID: "LM-NUT-1", OperationID: "annual-pdf558", Submission: sub, Files: files})
	req := httptest.NewRequest(http.MethodPost, "/api/reviews/submit", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer p")
	w := httptest.NewRecorder()
	s.securityHeaders(http.HandlerFunc(s.submitReview)).ServeHTTP(w, req)
	return w
}
func Test558AnnualPDFEntersReviewWithoutConclusionAndKeepsReadings(t *testing.T) {
	for _, choice := range []string{"", "reavivados", "bible_plan"} {
		t.Run(choice, func(t *testing.T) {
			s := annualFixture551(t, choice)
			sub, file := annualDocument558(choice)
			prior, _ := s.loadLegacy()
			old := legacyMap(mapField(reviewProgressContainer(legacyMap(prior.State.Database), "p", musicProgram553), "submissions")["LM-NUT-1"])
			old["bibleReadings"] = cloneAny(sub["bibleReadings"])
			s.saveLegacy(prior)
			// An older certificate upload may omit the existing diary entirely.
			delete(sub, "bibleReadings")
			if w := submitPDF558(t, s, sub, []any{file}); w.Code != 200 {
				t.Fatal(w.Code, w.Body.String())
			}
			st, _ := s.loadLegacy()
			db := legacyMap(st.State.Database)
			p := reviewProgressContainer(db, "p", musicProgram553)
			if stringValue(mapField(p, "status")["LM-NUT-1"]) != "in_review" {
				t.Fatal("not in review")
			}
			saved := legacyMap(mapField(p, "submissions")["LM-NUT-1"])
			if stringValue(saved["primaryDocumentId"]) != "annual-pdf" || len(mapField(legacyMap(saved["bibleReadings"]), "entries")) != 1 {
				t.Fatal("certificate or reading lost")
			}
			if musicComplete550(db, "p") || stringValue(legacyMap(mapField(musicData550(db), "workflow")["p"])["state"]) == "approved" {
				t.Fatal("upload bypassed review/final approval")
			}
			before := len(st.Notifications)
			if w := submitPDF558(t, s, sub, []any{file}); w.Code != 200 {
				t.Fatal("retry", w.Code)
			}
			st, _ = s.loadLegacy()
			if len(st.Notifications) != before {
				t.Fatal("duplicate notifications")
			}
		})
	}
}
func Test558AnnualPDFRejectsMissingInvalidOrForeignFiles(t *testing.T) {
	mutations := map[string]func(map[string]any, map[string]any){
		"missing primary":     func(s, f map[string]any) { delete(s, "primaryDocumentId") },
		"missing link":        func(s, f map[string]any) { s["attachmentIds"] = []any{} },
		"empty bytes":         func(s, f map[string]any) { f["dataUrl"] = "data:application/pdf;base64," },
		"fake PDF":            func(s, f map[string]any) { f["dataUrl"] = "data:application/pdf;base64,YWJj"; f["size"] = 3 },
		"another owner":       func(s, f map[string]any) { f["ownerId"] = "other" },
		"another program":     func(s, f map[string]any) { f["programId"] = "guide_major" },
		"another requirement": func(s, f map[string]any) { f["requirementId"] = "LM-TEM-1" },
		"too many pages":      func(s, f map[string]any) { s["documentPageCount"] = 6 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			s := annualFixture551(t, "reavivados")
			sub, file := annualDocument558("reavivados")
			mutate(sub, file)
			if w := submitPDF558(t, s, sub, []any{file}); w.Code != 422 {
				t.Fatal(w.Code, w.Body.String())
			}
			st, _ := s.loadLegacy()
			if stringValue(mapField(reviewProgressContainer(legacyMap(st.State.Database), "p", musicProgram553), "status")["LM-NUT-1"]) != "in_progress" {
				t.Fatal("invalid PDF changed status")
			}
		})
	}
	t.Run("missing file", func(t *testing.T) {
		s := annualFixture551(t, "reavivados")
		sub, _ := annualDocument558("reavivados")
		if w := submitPDF558(t, s, sub, nil); w.Code != 422 {
			t.Fatal(w.Code)
		}
	})
}
func Test558AnnualGenericSyncCannotSubmitUnreceivedPDF(t *testing.T) {
	s := annualFixture551(t, "reavivados")
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	incoming := legacyMap(cloneAny(db))
	p := reviewProgressContainer(incoming, "p", musicProgram553)
	sub, _ := annualDocument558("reavivados")
	ensureMapField(p, "submissions")["LM-NUT-1"] = sub
	ensureMapField(p, "status")["LM-NUT-1"] = "in_review"
	merged, _, err := mergeLegacyProposalOnServer(db, &legacyProposal{ActorID: "p", ActorRole: "pastor_aspirant", Database: incoming, OperationType: "auto_sync"})
	if err != nil {
        if err != errLegacyProposalNotApplicable {t.Fatal(err)}
        return // No change applied: presentation must use the validated delivery channel.
    }
	if stringValue(mapField(reviewProgressContainer(legacyMap(merged), "p", musicProgram553), "status")["LM-NUT-1"]) == "in_review" {
		t.Fatal("generic sync bypassed PDF receipt validation")
	}
}
func TestBrowserFixture558(t *testing.T) {
	if os.Getenv("SGL558_BROWSER_FIXTURE") != "1" {
		t.Skip("interactive fixture")
	}
	s, h := setup555(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		u["email"] = stringValue(u["id"]) + "@example.test"
	}
	legacyUser(db, "admin")["recoveryKeys"] = []any{testPasswordRecord("FixtureRecovery1!"), testPasswordRecord("FixtureRecovery2!")}
	p := reviewProgressContainer(db, "a", musicProgram553)
	ensureMapField(p, "status")["LM-NUT-1"] = "in_progress"
	ensureMapField(p, "submissions")["LM-NUT-1"] = map[string]any{"updatedAt": "before", "bibleReadings": map[string]any{"kind": "music_devotion", "choice": "reavivados", "startDate": "2026-01-01", "entries": map[string]any{"2026-01-01": map[string]any{"completed": true, "date": "2026-01-01", "favoriteVerses": "Salmos 1:1", "impactStory": "Dios acompaña", "lifeLesson": "Confiar"}}}}
	p = reviewProgressContainer(db, "b", musicProgram553)
	ensureMapField(p, "status")["LM-NUT-1"] = "pending"
	delete(ensureMapField(p, "submissions"), "LM-NUT-1")
	s.saveLegacy(st)
	ln, err := net.Listen("tcp", "127.0.0.1:8899")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	fmt.Println("SGL558_BROWSER_READY")
	http.Serve(ln, h)
}
