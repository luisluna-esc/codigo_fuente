package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test549MinistryRolePermissionsAndConflicts(t *testing.T) {
	for _, role := range []string{"admin", "departamental", "secretaria", "administrativo", "aspirant", "tutor"} {
		t.Run(role, func(t *testing.T) {
			s, db := fixture548R3(t)
			db["users"] = append(arrayField(db, "users"), map[string]any{"id": "actor", "role": role, "active": true})
			db["modules"] = []any{map[string]any{"id": "music_ministry", "active": true, "assignedUserIds": []any{"p", "actor"}}}
			testEnrollConquistadores593(db)
			st, _ := s.loadLegacy()
			st.State.Database = db
			if err := s.saveLegacy(st); err != nil {
				t.Fatal(err)
			}
			raw := "token"
			s.remoteSessions[hashToken(raw)] = remoteSession{UserID: "actor", Role: role, ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339)}
			value := map[string]any{"type": "Coro", "name": "Coro de prueba", "church": "Iglesia", "district": "Distrito", "region": "Región", "mission": "Misión", "union": "Unión", "createdDate": "2026-10-02", "boardVote": "12", "director": "Director", "directorPhone": "12345678", "role": "admin"}
			call := func(actor, owner, base string) int {
				body, _ := json.Marshal(map[string]any{"actorId": actor, "ownerId": owner, "baseUpdatedAt": base, "restoreGeneration": 1, "value": value})
				r := httptest.NewRequest("POST", "/api/music-registration", bytes.NewReader(body))
				r.Header.Set("Authorization", "Bearer "+raw)
				w := httptest.NewRecorder()
				s.securityHeaders(http.HandlerFunc(s.ministryResource548)).ServeHTTP(w, r)
				return w.Code
			}
			owner := "p"
			if role == "aspirant" {
				owner = "actor"
			}
			allowed := role == "admin" || role == "departamental" || role == "secretaria"
			code := call("actor", owner, "")
			if allowed && code != 200 {
				t.Fatalf("save returned %d", code)
			}
			if !allowed && code != 403 {
				t.Fatalf("unauthorized save returned %d", code)
			}
			if allowed {
				if call("actor", owner, "") != 409 {
					t.Fatal("stale overwrite accepted")
				}
				st, _ = s.loadLegacy()
				saved := legacyMap(legacyUser(legacyMap(st.State.Database), owner)["musicMinistryRegistration"])
				if saved["name"] != "Coro de prueba" || saved["role"] != nil {
					t.Fatal("unsafe or missing fields")
				}
				if call("d", owner, "") != 403 {
					t.Fatal("forged actor accepted")
				}
			}
			if role == "aspirant" && call("actor", "p", "") != 403 {
				t.Fatal("aspirant edited another owner")
			}
		})
	}
}
func Test549WatermarkAdminOnlyAndReadScope(t *testing.T) {
	s, db := fixture548R3(t)
	db["users"] = append(arrayField(db, "users"), map[string]any{"id": "admin", "role": "admin", "active": true}, map[string]any{"id": "sec", "role": "secretaria", "active": true})
	st, _ := s.loadLegacy()
	st.State.Database = db
	s.saveLegacy(st)
	for _, id := range []string{"admin", "sec", "p"} {
		role := legacyRole(legacyUser(db, id))
		s.remoteSessions[hashToken(id)] = remoteSession{UserID: id, Role: role, ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339)}
	}
	call := func(id string, value map[string]any) int {
		body, _ := json.Marshal(map[string]any{"actorId": id, "ownerId": "p", "restoreGeneration": 1, "value": value})
		r := httptest.NewRequest("POST", "/api/report-watermark", bytes.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+id)
		w := httptest.NewRecorder()
		s.securityHeaders(http.HandlerFunc(s.ministryResource548)).ServeHTTP(w, r)
		return w.Code
	}
	disabled := map[string]any{"enabled": false, "size": 60, "opacity": .15}
	for _, id := range []string{"sec", "p"} {
		if call(id, disabled) != 403 {
			t.Fatal("non-admin watermark update")
		}
	}
	if call("admin", map[string]any{"enabled": true, "dataUrl": "data:image/png;base64,ZmFrZQ==", "size": 60, "opacity": .15}) != 400 {
		t.Fatal("invalid PNG accepted")
	}
	if call("admin", disabled) != 200 {
		t.Fatal("admin cannot disable watermark")
	}
	snapshot := s.filteredLegacySnapshot(legacyState{Database: db}, "sec", "secretaria")
	out := legacyMap(snapshot["database"])
	if len(legacyUsers(out)) != 5 || legacyMap(out["progress"])["p"] == nil || legacyUser(out, "d")["password"] != nil {
		t.Fatal("secretary read scope or secret filtering")
	}
	r := httptest.NewRequest("POST", "/api/indications", bytes.NewBufferString("{}"))
	r.Header.Set("Authorization", "Bearer sec")
	w := httptest.NewRecorder()
	s.securityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })).ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("secretary may edit indications")
	}
}
