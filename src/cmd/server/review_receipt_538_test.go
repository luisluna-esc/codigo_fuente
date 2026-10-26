package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func review538Fixture(t *testing.T) (*Server, string) {
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

	return s, token
}
func Test538ReceiptRequiresDurableCompleteFiles(t *testing.T) {
	for _, kind := range []string{"complete", "missing", "metadata", "truncated", "wrong_owner", "disk_failure"} {
		t.Run(kind, func(t *testing.T) {
			s, token := review538Fixture(t)
			file := map[string]any{"id": "p", "ownerId": "asp1", "requirementId": "R1", "programId": "class_grouped", "kind": "photo", "type": "image/png", "size": 3, "dataUrl": "data:image/png;base64,YWJj"}
			files := []any{file}
			switch kind {
			case "missing":
				files = nil
			case "metadata":
				delete(file, "dataUrl")
			case "truncated":
				file["size"] = 5
			case "wrong_owner":
				file["ownerId"] = "other"
			case "disk_failure":
				blocker := filepath.Join(t.TempDir(), "blocker")
				os.WriteFile(blocker, []byte("x"), 0600)
				s.path = filepath.Join(blocker, "data.json")
				s.legacyStorageMode = "json-fallback"
			}
			sub := map[string]any{"text": "Primera linea\nSegunda linea\nTercera linea", "attachmentIds": []any{"p"}, "reportPhotoBoxes": map[string]any{"p": map[string]any{"page": 2, "x": 45}}, "reportLayoutPages": []any{"first", "second"}}
			body, _ := json.Marshal(map[string]any{"actorId": "asp1", "ownerId": "asp1", "operationId": "receipt-test", "requirementId": "R1", "programId": "class_grouped", "ackOnly": true, "contentHash": "same", "submission": sub, "files": files})
			send := func() *httptest.ResponseRecorder {
				r := httptest.NewRequest(http.MethodPost, "/api/reviews/submit", bytes.NewReader(body))
				r.Header.Set("Authorization", "Bearer "+token)
				rr := httptest.NewRecorder()
				s.submitReview(rr, r)
				return rr
			}
			rr := send()
			st, _ := s.loadLegacy()
			if kind != "complete" {
				if rr.Code == 200 {
					t.Fatal("accepted incomplete delivery")
				}
				if legacyRequirementStatus(st.State.Database, "asp1", "class_grouped", "R1") != "in_progress" {
					t.Fatal("failed delivery changed cached status")
				}
				if len(st.Notifications) != 0 || len(st.AppliedOperations) != 0 {
					t.Fatal("failed delivery published receipt")
				}
				return
			}
			if rr.Code != 200 {
				t.Fatal(rr.Code, rr.Body.String())
			}
			var ack map[string]any
			json.Unmarshal(rr.Body.Bytes(), &ack)
			if ack["filesReceived"] != true {
				t.Fatal("missing receipt")
			}
			stored := legacyMap(mapField(legacyProgressForRequirement(st.State.Database, "asp1", "class_grouped", false), "submissions")["R1"])
			if stored["reportPhotoBoxes"] == nil || stored["reportLayoutPages"] == nil {
				t.Fatal("lost geometry")
			}
			count := len(st.Notifications)
			rr = send()
			json.Unmarshal(rr.Body.Bytes(), &ack)
			st, _ = s.loadLegacy()
			if rr.Code != 200 || ack["filesReceived"] != true || len(st.Notifications) != count {
				t.Fatal("replay must not duplicate notices")
			}
		})
	}
}
