package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test548TutorCreatesSelectedModuleAndRejectsForgedModule(t *testing.T) {
	db := map[string]any{"users": []any{map[string]any{"id": "t", "role": "tutor", "active": true, "tutorProgramIds": []any{"music_ministry"}}}, "modules": []any{map[string]any{"id": "class_grouped", "active": true, "assignedUserIds": []any{}}, map[string]any{"id": "music_ministry", "active": true, "assignedUserIds": []any{}}}, "progress": map[string]any{}, "programProgress": map[string]any{}}
	source := legacyMap(cloneAny(db))
	source["users"] = append(arrayField(source, "users"), map[string]any{"id": "new", "role": "aspirant", "tutorId": "t", "active": true, "tutorProgramIds": []any{"class_grouped"}})
	legacyMap(arrayField(source, "modules")[1])["assignedUserIds"] = []any{"new"}
	result, _, err := mergeLegacyProposalOnServer(db, &legacyProposal{ActorID: "t", Database: source})
	if err != nil {
		t.Fatal(err)
	}
	merged := legacyMap(result)
	person := legacyUser(merged, "new")
	if !legacyTutorAssigned(person, "t", "music_ministry") || legacyTutorAssigned(person, "t", "class_grouped") {
		t.Fatal("wrong tutor scope")
	}
	if person["tutorProgramIds"] != nil {
		t.Fatal("injected tutor field retained")
	}
	if !containsString(arrayField(legacyMap(arrayField(merged, "modules")[1]), "assignedUserIds"), "new") {
		t.Fatal("chosen module lost")
	}
	if len(arrayField(legacyMap(arrayField(merged, "modules")[0]), "assignedUserIds")) != 0 {
		t.Fatal("unrequested class assigned")
	}
	legacyMap(arrayField(source, "modules")[0])["assignedUserIds"] = []any{"new"}
	if _, _, err = mergeLegacyProposalOnServer(db, &legacyProposal{ActorID: "t", Database: source}); err == nil {
		t.Fatal("forged module accepted")
	}
	if len(arrayField(legacyMap(arrayField(db, "modules")[1]), "assignedUserIds")) != 0 {
		t.Fatal("original mutated")
	}
}
func Test548ExplicitEmptyTutorScope(t *testing.T) {
	db := map[string]any{}
	if len(tutorPrograms548(db, map[string]any{"tutorProgramIds": []any{}})) != 0 {
		t.Fatal("empty scope widened")
	}
}
func Test548BirthdaysAuthenticatedMinimalDirectory(t *testing.T) {
	s, _ := recoveryFixture548(t)
	st, _ := s.loadLegacy()
	st.State.Database = map[string]any{"users": []any{map[string]any{"id": "a", "role": "admin", "fullName": "Ana", "birthDate": "1990-09-30", "photo": "data:image/png;base64,AA==", "password": "secret", "email": "private"}, map[string]any{"id": "b", "role": "tutor", "fullName": "Tutor", "birthDate": "1980-10-02"}}}
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/api/birthdays?actorId=a", nil)
	rr := httptest.NewRecorder()
	s.birthdays548(rr, req)
	if rr.Code != 401 {
		t.Fatal(rr.Code)
	}
	req.Header.Set("X-SGL-Token", "adm")
	rr = httptest.NewRecorder()
	s.birthdays548(rr, req)
	if rr.Code != 200 {
		t.Fatal(rr.Code)
	}
	var payload struct {
		People []map[string]any `json:"people"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.People) != 2 {
		t.Fatal("directory incomplete")
	}
	body := rr.Body.String()
	if strings.Contains(body, "1990") || strings.Contains(body, "password") || strings.Contains(body, "private") {
		t.Fatal("unnecessary profile data leaked")
	}
}
