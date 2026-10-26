package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func Test547CatalogGrantsAreAdminOwned(t *testing.T) {
	db := fixture545()
	patch := map[string]any{"users": []any{map[string]any{"id": "music", "requirementViewPrograms": []any{"guide_major", "music_ministry"}}}}
	direct, owners := splitAdminPatchByOwnership(db, patch)
	if len(owners) != 0 || len(legacyUser(direct, "music")["requirementViewPrograms"].([]any)) != 2 {
		t.Fatalf("permission was routed to owner approval: %#v %#v", direct, owners)
	}
	s := &Server{cfg: Config{AdminToken: "adm", ClientToken: "cli"}, path: filepath.Join(t.TempDir(), "data.json")}
	if err := s.saveLegacy(legacyStore{State: legacyState{ServerVersion: 1, Database: db}}); err != nil {
		t.Fatal(err)
	}
	for _, programs := range [][]any{{"guide_major", "music_ministry"}, {}} {
		patch["users"] = []any{map[string]any{"id": "music", "requirementViewPrograms": programs}}
		raw, _ := json.Marshal(map[string]any{"actorId": "admin", "operationId": "catalog" + string(rune('0'+len(programs))), "patch": patch})
		req := httptest.NewRequest("POST", "/api/admin/changes", bytes.NewReader(raw))
		req.Header.Set("Authorization", "Bearer adm")
		rr := httptest.NewRecorder()
		s.legacyAdminChangeSubmit(rr, req)
		if rr.Code != 200 {
			t.Fatalf("%d %s", rr.Code, rr.Body.String())
		}
		st, _ := s.loadLegacy()
		got, _ := legacyUser(st.State.Database, "music")["requirementViewPrograms"].([]any)
		if len(got) != len(programs) {
			t.Fatalf("grant/revoke mismatch: %#v", got)
		}
	}
}
func Test547TutorCannotGrantCatalogToSelfOrNewUser(t *testing.T) {
	db := fixture545()
	legacyUser(db, "music")["requirementViewPrograms"] = []any{"music_ministry"}
	incoming := legacyMap(cloneAny(db))
	legacyUser(incoming, "music")["requirementViewPrograms"] = []any{"class_grouped"}
	incoming["users"] = append(legacyUsers(incoming), map[string]any{"id": "new", "role": "aspirant", "tutorId": "music", "active": true, "requirementViewPrograms": []any{"class_grouped"}})
	merged, _, err := mergeLegacyProposalOnServer(db, &legacyProposal{ActorID: "music", ActorRole: "tutor", Database: incoming})
	if err != nil {
		t.Fatal(err)
	}
	got := legacyUser(merged, "music")["requirementViewPrograms"].([]any)
	if len(got) != 1 || got[0] != "music_ministry" {
		t.Fatalf("self grant accepted: %#v", got)
	}
	if legacyUser(merged, "new")["requirementViewPrograms"] != nil {
		t.Fatal("new account gained unauthorized catalog permissions")
	}
	// A catalog grant must not give permission to read or review another class's reports.
	s := &Server{cfg: Config{ClientToken: "cli"}, path: filepath.Join(t.TempDir(), "data.json")}
	legacyUser(db, "music")["requirementViewPrograms"] = []any{"class_grouped"}
	if err := s.saveLegacy(legacyStore{State: legacyState{Database: db}}); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/api/reviews/item?actorId=music&ownerId=a&programId=class_grouped&requirementId=R", nil)
	req.Header.Set("X-SGL-Token", "cli")
	rr := httptest.NewRecorder()
	s.legacyReviewItem(rr, req)
	if rr.Code != 403 {
		t.Fatalf("catalog grant expanded review scope: %d %s", rr.Code, rr.Body.String())
	}
}
