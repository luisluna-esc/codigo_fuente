package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
)

func body595() map[string]any {
	return map[string]any{"actorId": "a", "ownerId": "a", "operationId": "audit-595", "programId": "guide_major", "requirementId": "GM-PRE-2", "contentHash": "v1", "ackOnly": true, "submission": map[string]any{"text": "Primera línea\nSegunda línea\nTercera línea", "updatedAt": "2026-10-11T11:00:00Z"}, "files": []any{}}
}
func fixture595(t *testing.T) (*Server, http.Handler, string) {
	s, h := musicFixture553(t)
	_, login := musicLogin553(t, h, "a", false)
	return s, h, stringValue(login["sessionToken"])
}
func Test595RejectUnassignedSubmission(t *testing.T) {
	s, h, token := fixture595(t)
	body := body595()
	body["programId"] = "guide_master"
	body["requirementId"] = "GMM-PRE-1"
	before, _ := s.loadLegacy()
	snap, _ := json.Marshal(before.State.Database)
	code, result := musicCall553(t, h, "POST", "/api/reviews/submit", token, body)
	if code != 403 {
		t.Fatalf("unassigned submission accepted: %d %v", code, result)
	}
	after, _ := s.loadLegacy()
	afterJSON, _ := json.Marshal(after.State.Database)
	if string(snap) != string(afterJSON) {
		t.Fatal("rejection mutated database")
	}
}
func Test595RejectOldGenerationReview(t *testing.T) {
	s, h, token := fixture595(t)
	st, _ := s.loadLegacy()
	st.State.RestoreGeneration = 2
	s.saveLegacy(st)
	body := body595()
	body["restoreGeneration"] = 1
	code, p := musicCall553(t, h, "POST", "/api/reviews/submit", token, body)
	if code != 409 || p["code"] != "RESTORE_REQUIRED" {
		t.Fatalf("old-generation review allowed: %d %v", code, p)
	}
}
func Test595OperationCollisionNotAcknowledged(t *testing.T) {
	_, h, token := fixture595(t)
	body := body595()
	code, p := musicCall553(t, h, "POST", "/api/reviews/submit", token, body)
	if code != 200 {
		t.Fatal(code, p)
	}
	body["requirementId"] = "GM-PRE-3"
	code, p = musicCall553(t, h, "POST", "/api/reviews/submit", token, body)
	if code != 409 || p["clearLocalOperation"] == true {
		t.Fatalf("unreceived report acknowledged: %d %v", code, p)
	}
}
func Test595RecoveryForStaffTraining(t *testing.T) {
	for _, source := range []bool{false, true} {
		t.Run(map[bool]string{false: "capsule", true: "backup"}[source], func(t *testing.T) {
			s, c := fixtureBackup554(t)
			st, _ := s.loadLegacy()
			legacyUser(st.State.Database, "a")["role"] = "administrativo"
			s.saveLegacy(st)
			if !source {
				c.SourceUser = nil
			}
			rr := recoveryCall548(s.adminRecovery548, "adm", recoveryRequest548{Capsule: c})
			if rr.Code != 200 {
				t.Fatalf("staff recovery refused: %d %s", rr.Code, rr.Body.String())
			}
		})
	}
}
func Test595StaleReviewerCannotApproveResubmission(t *testing.T) {
	s, h, token := fixture595(t)
	body := body595()
	code, p := musicCall553(t, h, "POST", "/api/reviews/submit", token, body)
	if code != 200 {
		t.Fatal(code, p)
	}
	_, login := musicLogin553(t, h, "admin", false)
	admin := stringValue(login["sessionToken"])
	st, _ := s.loadLegacy()
	sub := legacyMap(mapField(legacyProgressForRequirement(st.State.Database, "a", "guide_major", false), "submissions")["GM-PRE-2"])
	old := stringValue(sub["updatedAt"])
	sub["updatedAt"] = "2026-10-11T12:00:00Z"
	sub["text"] = "Versión nueva\nQue no vio\nEl revisor"
	s.saveLegacy(st)
	code, p = musicCall553(t, h, "POST", "/api/reviews/decision", admin, map[string]any{"actorId": "admin", "ownerId": "a", "programId": "guide_major", "requirementId": "GM-PRE-2", "decision": "approved", "expectedUpdatedAt": old})
	if code != 409 {
		t.Fatalf("stale review allowed: %d %v", code, p)
	}
}
func Test595AnonymousCannotWriteCriticalEndpoints(t *testing.T) {
	_, h := musicFixture553(t)
	for _, path := range []string{"/api/reviews/submit", "/api/reviews/decision", "/api/admin/recovery", "/api/admin/restore", "/api/progress/reset", "/api/admin/changes", "/api/backup"} {
		r := httptest.NewRequest("POST", path, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code >= 200 && w.Code < 300 {
			t.Fatal("anonymous allowed", path)
		}
	}
}

func Test595SubmissionReviewCorrectionResubmission(t *testing.T) {
	s, h, token := fixture595(t)
	body := body595()
	_, login := musicLogin553(t, h, "tg", false)
	tutor := stringValue(login["sessionToken"])
	post := func(b map[string]any) {
		t.Helper()
		code, p := musicCall553(t, h, "POST", "/api/reviews/submit", token, b)
		if code != 200 {
			t.Fatal(code, p)
		}
	}
	post(body)
	st, _ := s.loadLegacy()
	firstVersion := st.State.ServerVersion
	firstNotices := len(st.Notifications)
	post(body)
	st, _ = s.loadLegacy()
	if st.State.ServerVersion != firstVersion || len(st.Notifications) != firstNotices {
		t.Fatal("duplicate changed state or notices")
	}
	decision := map[string]any{"actorId": "tg", "ownerId": "a", "programId": "guide_major", "requirementId": "GM-PRE-2", "decision": "needs_changes", "feedback": "Amplía tu reflexión", "ackOnly": true}
	code, p := musicCall553(t, h, "POST", "/api/reviews/decision", tutor, decision)
	if code != 200 {
		t.Fatal(code, p)
	}
	body["operationId"] = "corrected-595"
	body["contentHash"] = "v2"
	body["submission"] = map[string]any{"text": "Reflexión corregida\nExplicación ampliada\nAplicación personal", "updatedAt": "2026-10-11T13:00:00Z"}
	post(body)
	st, _ = s.loadLegacy()
	sub := legacyMap(mapField(legacyProgressForRequirement(st.State.Database, "a", "guide_major", false), "submissions")["GM-PRE-2"])
	decision["decision"] = "approved"
	decision["feedback"] = ""
	decision["expectedUpdatedAt"] = sub["updatedAt"]
	decision["expectedSubmittedAt"] = sub["submittedAt"]
	code, p = musicCall553(t, h, "POST", "/api/reviews/decision", tutor, decision)
	if code != 200 {
		t.Fatal(code, p)
	}
	post(body)
	st, _ = s.loadLegacy()
	if legacyRequirementStatus(st.State.Database, "a", "guide_major", "GM-PRE-2") != "approved" {
		t.Fatal("retry overwrote approval")
	}
}

func Test595SubmissionCannotImpersonateOwner(t *testing.T) {
	_, h, token := fixture595(t)
	for _, field := range []string{"actorId", "ownerId"} {
		body := body595()
		body[field] = "b"
		code, _ := musicCall553(t, h, "POST", "/api/reviews/submit", token, body)
		if code != 403 {
			t.Fatal(field, code)
		}
	}
}

func Test595ReviewDiskFailurePreservesCanonicalState(t *testing.T) {
	s, h, token := fixture595(t)
	code, p := musicCall553(t, h, "POST", "/api/reviews/submit", token, body595())
	if code != 200 {
		t.Fatal(code, p)
	}
	_, login := musicLogin553(t, h, "tg", false)
	tutor := stringValue(login["sessionToken"])
	st, _ := s.loadLegacy()
	before, _ := json.Marshal(st)
	if e := os.MkdirAll(s.legacyPath()+".tmp", 0700); e != nil {
		t.Fatal(e)
	}
	code, _ = musicCall553(t, h, "POST", "/api/reviews/decision", tutor, map[string]any{"actorId": "tg", "ownerId": "a", "programId": "guide_major", "requirementId": "GM-PRE-2", "decision": "approved"})
	if code != 500 {
		t.Fatal("failure not exercised", code)
	}
	after, _ := s.loadLegacy()
	encoded, _ := json.Marshal(after)
	if string(before) != string(encoded) {
		t.Fatal("failed review changed cached state")
	}
}

func Test595LoadingPreservesHistoricalEvidenceAndApproval(t *testing.T) {
	sub := map[string]any{"primaryDocumentId": "old-pdf", "attachmentIds": []any{"old-pdf"}, "approval": map[string]any{"reviewerId": "t"}, "feedback": "Anterior"}
	file := map[string]any{"id": "old-pdf", "ownerId": "a", "requirementId": "R", "type": "application/pdf", "dataUrl": "data:application/pdf;base64,b2xk"}
	st := legacyStore{State: legacyState{Database: map[string]any{"progress": map[string]any{"a": map[string]any{"status": map[string]any{"R": "approved"}, "submissions": map[string]any{"R": sub}}}}, Files: []any{file}}}
	before, _ := json.Marshal(mapField(legacyMap(st.State.Database), "progress"))
	originalFiles, _ := json.Marshal(st.State.Files)
	normalizeLegacyStore(&st)
	after, _ := json.Marshal(mapField(legacyMap(st.State.Database), "progress"))
	afterFiles, _ := json.Marshal(st.State.Files)
	if string(before) != string(after) || string(originalFiles) != string(afterFiles) {
		t.Fatal("loading destroyed historical work")
	}
}

func Test595AutosaveCannotWriteUnassignedProgram(t *testing.T) {
	s, h, token := fixture595(t)
	st, _ := s.loadLegacy()
	source := legacyMap(cloneAny(st.State.Database))
	p := reviewProgressContainer(source, "a", "guide_master")
	ensureMapField(p, "status")["GMM-PRE-1"] = "in_progress"
	ensureMapField(p, "submissions")["GMM-PRE-1"] = map[string]any{"text": "No asignado"}
	code, _ := musicCall553(t, h, "POST", "/api/sync", token, map[string]any{"actorId": "a", "actorRole": "aspirant", "ownerId": "a", "programId": "guide_master", "requirementId": "GMM-PRE-1", "restoreGeneration": 1, "baseVersion": 7, "installationId": "scope595", "operationId": "scope595", "operationType": "save_progress", "database": source, "files": []any{}, "ackOnly": true})
	after, _ := s.loadLegacy()
	if legacyRequirementStatus(after.State.Database, "a", "guide_master", "GMM-PRE-1") == "in_progress" {
		t.Fatalf("unassigned autosave persisted: HTTP %d", code)
	}
}

func Test595Browser(t *testing.T) {
	if os.Getenv("SGL595_BROWSER") != "1" {
		t.Skip("browser opt-in")
	}
	s, h := setup555(t)
	st, _ := s.loadLegacy()
	p := reviewProgressContainer(legacyMap(st.State.Database), "a", musicProgram553)
	ensureMapField(p, "status")["LM-PRE-2"] = "in_review"
	ensureMapField(p, "submissions")["LM-PRE-2"] = map[string]any{"text": "Informe de prueba\nSegunda línea\nTercera línea", "updatedAt": "2026-10-11T11:00:00Z", "submittedAt": "2026-10-11T11:01:00Z"}
	s.saveLegacy(st)
	server := httptest.NewServer(h)
	defer server.Close()
	cmd := exec.Command("node", "../../../tests/browser-audit-595.cjs")
	cmd.Env = append(os.Environ(), "SGL_BROWSER_BASE="+server.URL)
	out, e := cmd.CombinedOutput()
	t.Log(string(out))
	if e != nil {
		t.Fatal(e)
	}
}

func Test595AspirantCannotSelfApproveViaSync(t *testing.T) {
	s, h, token := fixture595(t)
	st, _ := s.loadLegacy()
	source := legacyMap(cloneAny(st.State.Database))
	p := reviewProgressContainer(source, "a", "guide_major")
	ensureMapField(p, "status")["GM-PRE-2"] = "approved"
	ensureMapField(p, "submissions")["GM-PRE-2"] = map[string]any{"text": "Autorización inventada", "approval": map[string]any{"reviewerId": "admin"}}
	code, _ := musicCall553(t, h, "POST", "/api/sync", token, map[string]any{"actorId": "a", "actorRole": "aspirant", "ownerId": "a", "programId": "guide_major", "requirementId": "GM-PRE-2", "restoreGeneration": 1, "baseVersion": 7, "installationId": "self595", "operationId": "self595", "operationType": "save_progress", "database": source, "files": []any{}, "ackOnly": true})
	after, _ := s.loadLegacy()
	if legacyRequirementStatus(after.State.Database, "a", "guide_major", "GM-PRE-2") == "approved" {
		t.Fatalf("self approval persisted: HTTP %d", code)
	}
}

func Test595ConcurrentRetriesApplyExactlyOnce(t *testing.T) {
	s, h, token := fixture595(t)
	body, _ := json.Marshal(body595())
	start := make(chan struct{})
	results := make(chan int, 32)
	for i := 0; i < 32; i++ {
		go func() {
			<-start
			r := httptest.NewRequest("POST", "/api/reviews/submit", bytes.NewReader(body))
			r.Header.Set("Authorization", "Bearer "+token)
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			results <- w.Code
		}()
	}
	close(start)
	for i := 0; i < 32; i++ {
		if code := <-results; code != 200 {
			t.Fatalf("retry failed %d", code)
		}
	}
	st, _ := s.loadLegacy()
	count := 0
	for _, op := range st.AppliedOperations {
		if op.OperationID == "audit-595" {
			count++
		}
	}
	if count != 1 || st.State.ServerVersion != 8 {
		t.Fatalf("duplicate commit: count=%d version=%d", count, st.State.ServerVersion)
	}
}
