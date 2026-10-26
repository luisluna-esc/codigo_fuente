package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoteReviewSubmitAppliesDeltaAndNotifies(t *testing.T) {
	dir := t.TempDir()
	s := &Server{
		cfg:            Config{PublicWeb: true, ServerName: "Test"},
		path:           filepath.Join(dir, "data", "sgl-server-data.json"),
		remoteSessions: map[string]remoteSession{}, loginAttempts: map[string]loginAttempt{},
	}
	db := map[string]any{
		"settings": map[string]any{},
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Administrador", "username": "admin", "role": "admin", "active": true, "accountState": "active", "password": testPasswordRecord("Admin123!")},
			map[string]any{"id": "t1", "fullName": "Tutor Uno", "username": "tutor", "role": "tutor", "active": true, "accountState": "active", "password": testPasswordRecord("Tutor123!")},
			map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "username": "asp1", "role": "aspirant", "tutorId": "t1", "active": true, "accountState": "active", "password": testPasswordRecord("Asp12345!")},
		},
		"progress": map[string]any{"asp1": map[string]any{
			"status":      map[string]any{"R1": "in_progress"},
			"submissions": map[string]any{"R1": map[string]any{"text": "borrador", "attachmentIds": []any{}}},
		}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"tasks":     []any{map[string]any{"id": "task1", "assigneeId": "asp1", "requirementId": "R1", "programId": "class_grouped", "presentedAt": ""}},
		"materials": []any{}, "modules": []any{map[string]any{"id": "class_grouped", "active": true, "assignedUserIds": []any{"asp1", "asp2"}}}, "classes": []any{}, "audit": []any{}, "revision": 1.0,
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 5, Database: db, Files: []any{}}}); err != nil {
		t.Fatal(err)
	}

	loginBody, _ := json.Marshal(map[string]any{"identity": "asp1", "password": "Asp12345!", "deviceId": "phone-1"})
	loginReq := httptest.NewRequest(http.MethodPost, "http://example/api/remote/login", bytes.NewReader(loginBody))
	loginReq.RemoteAddr = "203.0.113.7:1234"
	loginRR := httptest.NewRecorder()
	s.remoteLogin(loginRR, loginReq)
	if loginRR.Code != 200 {
		t.Fatalf("login=%d %s", loginRR.Code, loginRR.Body.String())
	}
	var login map[string]any
	_ = json.Unmarshal(loginRR.Body.Bytes(), &login)
	token, _ := login["sessionToken"].(string)
	if token == "" {
		t.Fatal("missing remote token")
	}

	file := map[string]any{"id": "file1", "ownerId": "asp1", "programId": "class_grouped", "requirementId": "R1", "name": "evidencia.png", "kind": "photo", "type": "image/png", "size": 3.0, "dataUrl": "data:image/png;base64,YWJj"}
	normalReportText := "Primera linea del informe\nSegunda linea del informe\nTercera linea del informe"
	reviewBody, _ := json.Marshal(map[string]any{
		"installationId": "phone-1", "actorId": "asp1", "actorRole": "aspirant", "ownerId": "asp1",
		"operationId": "SGL-REV-ASP1-R1-TEST", "requirementId": "R1", "requirementTitle": "Requisito de prueba", "programId": "class_grouped", "contentHash": "abc",
		"submission": map[string]any{"text": normalReportText, "richText": "<p>Primera linea del informe</p><p>Segunda linea del informe</p><p>Tercera linea del informe</p>", "completionDate": "2026-08-31", "attachmentIds": []any{"file1"}, "updatedAt": "2026-08-31T21:30:00Z"},
		"files":      []any{file},
	})
	req := httptest.NewRequest(http.MethodPost, "http://example/api/reviews/submit", bytes.NewReader(reviewBody))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	s.submitReview(rr, req)
	if rr.Code != 200 {
		t.Fatalf("submit=%d %s", rr.Code, rr.Body.String())
	}

	st, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	p := legacyMap(legacyMap(st.State.Database)["progress"])
	asp := legacyMap(p["asp1"])
	if got := stringValue(legacyMap(asp["status"])["R1"]); got != "in_review" {
		t.Fatalf("status=%s", got)
	}
	sub := legacyMap(legacyMap(asp["submissions"])["R1"])
	if stringValue(sub["text"]) != normalReportText {
		t.Fatalf("submission not applied: %#v", sub)
	}
	if len(st.State.Files) != 1 {
		t.Fatalf("files=%d", len(st.State.Files))
	}
	seenAdmin, seenTutor := false, false
	for _, n := range st.Notifications {
		if n.SourceOperationID != "SGL-REV-ASP1-R1-TEST" {
			continue
		}
		if n.UserID == "admin1" {
			seenAdmin = true
		}
		if n.UserID == "t1" {
			seenTutor = true
		}
	}
	if !seenAdmin || !seenTutor {
		t.Fatalf("notifications admin=%v tutor=%v", seenAdmin, seenTutor)
	}

	// Idempotent retry must not duplicate state or notifications.
	req2 := httptest.NewRequest(http.MethodPost, "http://example/api/reviews/submit", bytes.NewReader(reviewBody))
	req2.Header.Set("Authorization", "Bearer "+token)
	rr2 := httptest.NewRecorder()
	s.submitReview(rr2, req2)
	if rr2.Code != 200 {
		t.Fatalf("retry=%d %s", rr2.Code, rr2.Body.String())
	}
	var payload map[string]any
	_ = json.Unmarshal(rr2.Body.Bytes(), &payload)
	if payload["deduplicated"] != true {
		t.Fatalf("retry was not deduplicated: %#v", payload)
	}

	// A second device or browser with a different operationId must not create
	// another review while the same requirement is already in_review.
	var otherPayload map[string]any
	if err := json.Unmarshal(reviewBody, &otherPayload); err != nil {
		t.Fatal(err)
	}
	otherPayload["operationId"] = "SGL-REV-ASP1-R1-TEST-OTHER"
	otherPayload["contentHash"] = "abc-other"
	otherPayload["submission"] = map[string]any{"text": "segundo envio", "richText": "<p>segundo envio</p>", "completionDate": "2026-09-01", "attachmentIds": []any{"file1"}, "updatedAt": "2026-09-01T10:00:00Z"}
	otherBody, _ := json.Marshal(otherPayload)
	req3 := httptest.NewRequest(http.MethodPost, "http://example/api/reviews/submit", bytes.NewReader(otherBody))
	req3.Header.Set("Authorization", "Bearer "+token)
	rr3 := httptest.NewRecorder()
	s.submitReview(rr3, req3)
	if rr3.Code != 200 {
		t.Fatalf("second device submit=%d %s", rr3.Code, rr3.Body.String())
	}
	var payload3 map[string]any
	_ = json.Unmarshal(rr3.Body.Bytes(), &payload3)
	if payload3["deduplicated"] != true || payload3["alreadyInReview"] != true || payload3["clearLocalOperation"] != true {
		t.Fatalf("second device was not blocked as in_review: %#v", payload3)
	}

	st2, err := s.loadLegacy()
	if err != nil {
		t.Fatal(err)
	}
	p2 := legacyMap(legacyMap(st2.State.Database)["progress"])
	asp2 := legacyMap(p2["asp1"])
	sub2 := legacyMap(legacyMap(asp2["submissions"])["R1"])
	if stringValue(sub2["text"]) != normalReportText {
		t.Fatalf("second device overwrote original submission: %#v", sub2)
	}
	for _, n := range st2.Notifications {
		if n.SourceOperationID == "SGL-REV-ASP1-R1-TEST-OTHER" {
			t.Fatalf("second device created duplicate notification: %#v", n)
		}
	}
}

func TestReviewNormalReportRequiresThreeTextRows(t *testing.T) {
	photo := map[string]any{"id": "photo1", "kind": "photo", "type": "image/png"}
	cases := []struct {
		name       string
		submission map[string]any
		files      []any
		wantOK     bool
	}{
		{name: "image only is rejected", submission: map[string]any{"attachmentIds": []any{"photo1"}}, files: []any{photo}, wantOK: false},
		{name: "one line is rejected", submission: map[string]any{"text": "Una sola linea", "attachmentIds": []any{}}, wantOK: false},
		{name: "two lines are rejected", submission: map[string]any{"text": "Primera linea\nSegunda linea", "attachmentIds": []any{}}, wantOK: false},
		{name: "three explicit lines are accepted", submission: map[string]any{"text": "Primera linea\nSegunda linea\nTercera linea", "attachmentIds": []any{}}, wantOK: true},
		{name: "three lines plus image are accepted", submission: map[string]any{"text": "Primera linea\nSegunda linea\nTercera linea", "attachmentIds": []any{"photo1"}}, files: []any{photo}, wantOK: true},
		{name: "rich text lines are accepted", submission: map[string]any{"richText": "<p>Primera linea</p><p>Segunda linea</p><p>Tercera linea</p>", "attachmentIds": []any{}}, wantOK: true},
		{name: "canonical wrapped rows are accepted", submission: map[string]any{"text": strings.Repeat("a", reviewNormalReportRowChars*2+1), "attachmentIds": []any{}}, wantOK: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, _ := reviewSubmissionIsInternalReport(tc.submission, tc.files)
			if ok != tc.wantOK {
				t.Fatalf("ok=%v want=%v rows=%d", ok, tc.wantOK, reviewNormalReportTextRows(tc.submission))
			}
		})
	}
}
