package main

import (
	"reflect"
	"testing"
)

func Test597NewRolesAreMusicOnly(t *testing.T) {
	s, h := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	for _, tc := range []struct{ id, role, canonical string }{{"secretaria", "administrativo_sgm", "secretaria"}, {"tm", "lider_musica", "tutor"}} {
		u := legacyUser(db, tc.id)
		u["role"] = tc.role
		delete(u, "tutorProgramIds")
		delete(u, "managementProgramIds")
		if legacyRole(u) != tc.canonical || !managementProgram555(db, u, musicProgram553) {
			t.Fatal("missing music permission", tc.role)
		}
		for _, p := range []string{"class_grouped", "guide_major", "guide_master", "guide_master_advanced"} {
			if managementProgram555(db, u, p) {
				t.Fatal("cross-module permission", tc.role, p)
			}
		}
	}
	s.saveLegacy(st)
	for _, id := range []string{"secretaria", "tm"} {
		code, _ := musicLogin553(t, h, id, true)
		if code != 200 {
			t.Fatal("new role cannot login", id, code)
		}
	}
}
func Test597CreateRequiresTutorAndPreservesCreator(t *testing.T) {
	s, _ := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	actor := legacyUser(db, "administrativo")
	v := map[string]any{"id": "new597", "fullName": "Persona nueva", "email": "new597@example.test", "username": "new597", "region": "1", "password": tutorPassword556()}
	if _, reason := musicNewAspirant560(db, actor, v); reason == "" {
		t.Fatal("missing tutor accepted")
	}
	v["tutorId"] = "other-tm"
	u, reason := musicNewAspirant560(db, actor, v)
	if reason != "" || u["createdBy"] != "administrativo" || !legacyTutorAssigned(u, "other-tm", musicProgram553) {
		t.Fatal("cross-region tutor rejected", reason)
	}
	v["tutorId"] = "departamental"
	if _, reason = musicNewAspirant560(db, actor, v); reason == "" {
		t.Fatal("ordinary administrative assigned restricted authority")
	}
}
func Test597TransferNotifiesBothTutorsAndPreservesOtherModule(t *testing.T) {
	s, h := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	owner := legacyUser(db, "e")
	ensureMapField(owner, "tutorByProgram")["guide_major"] = "tg"
	s.saveLegacy(st)
	_, login := musicLogin553(t, h, "administrativo", true)
	code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", stringValue(login["sessionToken"]), map[string]any{"actorId": "administrativo", "ownerId": "e", "action": "assign_tutor", "restoreGeneration": 1, "expectedUpdatedAt": owner["updatedAt"], "value": map[string]any{"tutorId": "other-tm", "region": "2"}})
	if code != 200 {
		t.Fatal("transfer", code)
	}
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	owner = legacyUser(db, "e")
	if !legacyTutorAssigned(owner, "tg", "guide_major") || !legacyTutorAssigned(owner, "other-tm", musicProgram553) {
		t.Fatal("assignment isolation")
	}
	seen := map[string]bool{}
	for _, n := range st.Notifications {
		if n.Type == "tutor_changed" {
			seen[n.UserID] = true
		}
	}
	for _, id := range []string{"administrativo", "other-tm", "e"} {
		if !seen[id] {
			t.Fatal("notification missing", id)
		}
	}
	if !musicScope550(db, legacyUser(db, "administrativo"), owner) {
		t.Fatal("creator lost access")
	}
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", stringValue(login["sessionToken"]), map[string]any{"actorId": "administrativo", "ownerId": "d", "action": "assign_tutor", "restoreGeneration": 1, "expectedUpdatedAt": "", "value": map[string]any{"tutorId": "tm"}})
	if code != 403 {
		t.Fatal("out of scope transfer", code)
	}
}
func Test597SelfProfileEmailAndGeneration(t *testing.T) {
	s, h := excelFixture560(t)
	_, login := musicLogin553(t, h, "a", true)
	token := stringValue(login["sessionToken"])
	req := map[string]any{"actorId": "a", "ownerId": "a", "action": "profile", "restoreGeneration": 1, "expectedUpdatedAt": "", "value": map[string]any{"fullName": "Nuevo nombre", "email": "nuevo597@example.test", "region": "12"}}
	code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", token, req)
	if code != 400 {
		t.Fatal("email requires confirmation", code)
	}
	legacyMap(req["value"])["emailChangeConfirmed"] = true
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", token, req)
	if code != 200 {
		t.Fatal("own profile", code)
	}
	st, _ := s.loadLegacy()
	u := legacyUser(st.State.Database, "a")
	if u["email"] != "nuevo597@example.test" || u["region"] == "12" {
		t.Fatal("profile isolation")
	}
	req["restoreGeneration"] = 2
	req["expectedUpdatedAt"] = u["updatedAt"]
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", token, req)
	if code != 409 {
		t.Fatal("generation barrier", code)
	}
}
func Test597EmptyRemovalAndNamedAccountProtection(t *testing.T) {
	s, h := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	db["users"] = append(legacyUsers(db), map[string]any{"id": "ghost597", "active": true})
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if stringValue(m["id"]) == musicProgram553 {
			m["assignedUserIds"] = append(arrayField(m, "assignedUserIds"), "ghost597")
		}
	}
	before := cloneAny(legacyProgressForRequirement(db, "a", "guide_major", false))
	s.saveLegacy(st)
	_, login := musicLogin553(t, h, "admin", true)
	token := stringValue(login["sessionToken"])
	for _, tc := range []struct {
		id   string
		code int
	}{{"a", 409}, {"ghost597", 200}} {
		code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", token, map[string]any{"actorId": "admin", "ownerId": tc.id, "action": "remove_empty", "restoreGeneration": 1, "expectedUpdatedAt": ""})
		if code != tc.code {
			t.Fatal(tc.id, code)
		}
	}
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	if legacyUser(db, "ghost597") != nil || len(arrayField(db, "musicEmptyArchive597")) != 1 || !reflect.DeepEqual(before, legacyProgressForRequirement(db, "a", "guide_major", false)) {
		t.Fatal("unsafe empty cleanup")
	}
}
func Test597AssignTaskScopeAndDuplicate(t *testing.T) {
	_, h := excelFixture560(t)
	_, login := musicLogin553(t, h, "administrativo", true)
	req := map[string]any{"actorId": "administrativo", "ownerId": "e", "action": "assign_task", "restoreGeneration": 1, "expectedUpdatedAt": "", "value": map[string]any{"id": "task597", "requirementId": "LM-NUT-1", "plannedDate": "2026-10-12", "dueDate": "2026-10-13", "instructions": "Lectura personal"}}
	code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", stringValue(login["sessionToken"]), req)
	if code != 200 {
		t.Fatal("task assignment", code)
	}
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", stringValue(login["sessionToken"]), req)
	if code != 409 {
		t.Fatal("duplicate or stale task accepted", code)
	}
	req["ownerId"] = "d"
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", stringValue(login["sessionToken"]), req)
	if code != 403 {
		t.Fatal("foreign task", code)
	}
}
