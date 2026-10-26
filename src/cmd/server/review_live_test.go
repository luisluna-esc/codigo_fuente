package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
)

func TestReviewItemReturnsOnlyRequestedCanonicalDelivery(t *testing.T) {
	dir := t.TempDir()
	s := &Server{
		cfg:  Config{AdminToken: "admin-token", ClientToken: "client-token"},
		path: filepath.Join(dir, "data", "sgl-server-data.json"),
	}
	db := map[string]any{
		"settings": map[string]any{},
		"users": []any{
			map[string]any{"id": "admin1", "fullName": "Administrador", "role": "admin", "active": true, "accountState": "active"},
			map[string]any{"id": "t1", "fullName": "Tutor Uno", "role": "tutor", "active": true, "accountState": "active"},
			map[string]any{"id": "t2", "fullName": "Tutor Dos", "role": "tutor", "active": true, "accountState": "active"},
			map[string]any{"id": "asp1", "fullName": "Aspirante Uno", "role": "aspirant", "tutorId": "t1", "active": true, "accountState": "active"},
		},
		"progress": map[string]any{"asp1": map[string]any{
			"status": map[string]any{"R1": "in_review"},
			"submissions": map[string]any{"R1": map[string]any{
				"text": "Primera linea\nSegunda linea\nTercera linea", "attachmentIds": []any{"photo-r1"}, "updatedAt": "2026-09-17T22:00:00Z",
			}},
			"updatedAt": "2026-09-17T22:00:00Z",
		}},
		"programProgress": map[string]any{},
		"revision":        3.0,
	}
	files := []any{
		map[string]any{"id": "photo-r1", "ownerId": "asp1", "programId": "class_grouped", "requirementId": "R1", "kind": "photo", "type": "image/jpeg", "dataUrl": "data:image/jpeg;base64,YWJj"},
		map[string]any{"id": "private-other", "ownerId": "asp1", "programId": "class_grouped", "requirementId": "R2", "kind": "photo", "type": "image/jpeg", "dataUrl": "data:image/jpeg;base64,ZGVm"},
	}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 41, Database: db, Files: files}}); err != nil {
		t.Fatal(err)
	}

	call := func(actorID, token string) *httptest.ResponseRecorder {
		q := url.Values{"actorId": {actorID}, "ownerId": {"asp1"}, "programId": {"class_grouped"}, "requirementId": {"R1"}}
		req := httptest.NewRequest(http.MethodGet, "http://example/api/reviews/item?"+q.Encode(), nil)
		req.Header.Set("X-SGL-Token", token)
		rr := httptest.NewRecorder()
		s.legacyReviewItem(rr, req)
		return rr
	}

	for _, tc := range []struct {
		name, actor, token string
	}{
		{name: "admin", actor: "admin1", token: "admin-token"},
		{name: "assigned tutor", actor: "t1", token: "client-token"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := call(tc.actor, tc.token)
			if rr.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
			}
			var payload struct {
				OK            bool           `json:"ok"`
				Status        string         `json:"status"`
				Submission    map[string]any `json:"submission"`
				Files         []any          `json:"files"`
				ServerVersion int64          `json:"serverVersion"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if !payload.OK || payload.Status != "in_review" || payload.ServerVersion != 41 {
				t.Fatalf("payload=%#v", payload)
			}
			if stringValue(payload.Submission["text"]) == "" {
				t.Fatalf("submission missing: %#v", payload.Submission)
			}
			if len(payload.Files) != 1 || stringValue(legacyMap(payload.Files[0])["id"]) != "photo-r1" {
				t.Fatalf("files=%#v", payload.Files)
			}
		})
	}

	rr := call("t2", "client-token")
	if rr.Code != http.StatusForbidden {
		t.Fatalf("unassigned tutor status=%d body=%s", rr.Code, rr.Body.String())
	}
}
