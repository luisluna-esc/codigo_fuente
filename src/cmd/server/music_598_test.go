package main

import (
	"testing"
)

func Test598RegionalScopeAndRevokedTransfer(t *testing.T) {
	s, h := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	actor := legacyUser(db, "regional")
	owner := legacyUser(db, "a")
	owner["createdBy"] = "regional"
	ensureMapField(owner, "tutorByProgram")[musicProgram553] = ""
	if !musicScope550(db, actor, owner) {
		t.Fatal("creator must see unassigned")
	}
	ensureMapField(owner, "tutorByProgram")[musicProgram553] = "other-tm"
	if musicScope550(db, actor, owner) || musicAssignOwner597(db, actor, owner) {
		t.Fatal("creator retained transferred access")
	}
	if musicScope550(db, actor, legacyUser(db, "e")) {
		t.Fatal("region leaked")
	}
	if !musicScope550(db, actor, legacyUser(db, "b")) {
		t.Fatal("assigned foreign region missing")
	}
	legacyUser(db, "tm")["createdBy"] = "regional"
	if !musicSeeTutor560(db, actor, legacyUser(db, "tm")) || musicSeeTutor560(db, actor, legacyUser(db, "other-tm")) {
		t.Fatal("tutor creator scope")
	}
	s.saveLegacy(st)
	_, login := musicLogin553(t, h, "regional", true)
	token := stringValue(login["sessionToken"])
	code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", token, map[string]any{"actorId": "regional", "ownerId": "a", "action": "assign_tutor", "restoreGeneration": 1, "expectedUpdatedAt": owner["updatedAt"], "value": map[string]any{"tutorId": "regional"}})
	if code != 403 {
		t.Fatal("reclaimed transferred account", code)
	}
	code, snap := musicCall553(t, h, "GET", musicAPI553+"/api/accounts/view?actorId=regional", token, nil)
	if code != 200 {
		t.Fatal(code)
	}
	view := legacyMap(snap["database"])
	if legacyUser(view, "a") != nil || legacyUser(view, "e") != nil || legacyUser(view, "other-tm") != nil {
		t.Fatal("out-of-scope snapshot")
	}
}
func Test598WhatsappOptInRevocationAndProjection(t *testing.T) {
	s, h := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	actor := legacyUser(db, "regional")
	owner := legacyUser(db, "b")
	owner["phone"] = "+591 7123 4567"
	if link, _ := musicContactURL598(db, actor, owner); link != "" {
		t.Fatal("default enabled")
	}
	musicData550(db)["interface598"] = map[string]any{"driveVisible": true, "whatsappUsers": []any{"regional"}}
	if link, err := musicContactURL598(db, actor, owner); err == "" || link != "" {
		t.Fatal(link, err)
	}
	if link, _ := musicContactURL598(db, actor, legacyUser(db, "a")); link != "" {
		t.Fatal("contact scope leak")
	}
	s.saveLegacy(st)
	_, login := musicLogin553(t, h, "regional", true)
	token := stringValue(login["sessionToken"])
	code, snap := musicCall553(t, h, "GET", musicAPI553+"/api/accounts/view?actorId=regional", token, nil)
	if code != 200 || mapField(legacyMap(snap["database"]), "music550")["canWhatsapp598"] != false {
		t.Fatal("permission missing in snapshot", code)
	}
	code, _ = musicCall553(t, h, "GET", musicAPI553+"/api/music/contact?actorId=regional&ownerId=b", token, nil)
	if code != 403 {
		t.Fatal("removed contact allowed", code)
	}
	st, _ = s.loadLegacy()
	ensureMapField(musicData550(legacyMap(st.State.Database)), "interface598")["whatsappUsers"] = []any{}
	s.saveLegacy(st)
	code, _ = musicCall553(t, h, "GET", musicAPI553+"/api/music/contact?actorId=regional&ownerId=b", token, nil)
	if code != 403 {
		t.Fatal("revocation ignored", code)
	}
}
func Test598OnlyAdminCanEnableOptions(t *testing.T) {
	_, h := excelFixture560(t)
	_, login := musicLogin553(t, h, "regional", true)
	code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/music/module", stringValue(login["sessionToken"]), map[string]any{"actorId": "regional", "action": "interface_settings", "restoreGeneration": 1, "value": map[string]any{"driveVisible": true, "whatsappUsers": []any{"regional"}}})
	if code != 403 {
		t.Fatal("regional configured permissions", code)
	}
}

func Test598AdminOptionsPersistAndRegionalUnassignedCreation(t *testing.T) {
	s, h := excelFixture560(t)
	_, login := musicLogin553(t, h, "admin", true)
	code, out := musicCall553(t, h, "POST", musicAPI553+"/api/music/module", stringValue(login["sessionToken"]), map[string]any{"actorId": "admin", "action": "interface_settings", "restoreGeneration": 1, "value": map[string]any{"driveVisible": true, "whatsappUsers": []any{"regional", "a", "regional"}}})
	if code != 200 {
		t.Fatal(code, out)
	}
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	prefs := mapField(musicData550(db), "interface598")
	if prefs["driveVisible"] != nil || len(arrayField(prefs, "whatsappUsers")) != 0 {
		t.Fatal("invalid or duplicate recipient permitted", prefs)
	}
	v := map[string]any{"id": "new598", "fullName": "Nueva persona", "email": "new598@example.test", "username": "new598", "region": "1", "password": tutorPassword556()}
	actor := legacyUser(db, "regional")
	u, reason := musicNewAspirant560(db, actor, v)
	if reason != "" || u["createdBy"] != "regional" || legacyTutorForProgram(u, musicProgram553) != "" {
		t.Fatal(reason, u)
	}
	db["users"] = append(legacyUsers(db), u)
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if stringValue(m["id"]) == musicProgram553 {
			m["assignedUserIds"] = append(arrayField(m, "assignedUserIds"), "new598")
		}
	}
	if !musicUnassignedMember561(db, u) || !musicScope550(db, actor, u) {
		t.Fatal("new unassigned not visible")
	}
	delete(u, "createdBy")
	if musicScope550(db, actor, u) {
		t.Fatal("legacy creator inferred from region")
	}
}
