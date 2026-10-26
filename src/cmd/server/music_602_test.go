package main

import (
	"testing"
)

func Test602PermissionPrecedenceProjectionAndRoleScope(t *testing.T) {
	s, _ := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	actor := legacyUser(db, "regional")
	cfg := map[string]any{"roles": map[string]any{"regional": map[string]any{"reports.approve": false}}, "users": map[string]any{"regional": map[string]any{"reports.approve": true}}}
	musicData550(db)["permissions602"] = cfg
	if !musicPermit602(db, actor, "reports.approve") {
		t.Fatal("user override")
	}
	delete(mapField(cfg, "users"), "regional")
	if musicPermit602(db, actor, "reports.approve") {
		t.Fatal("role denial")
	}
	out := musicView550(db, actor)
	if out["permissions602"] != nil || mapField(out, "effective602")["reports.approve"] != false {
		t.Fatal("projection")
	}
	db["music550"] = out
	if musicPermit602(db, actor, "reports.approve") {
		t.Fatal("double projection restored revoked capability")
	}
	if musicScope550(db, actor, legacyUser(db, "a")) {
		t.Fatal("permissions broadened owner scope")
	}
}
func Test602MentoringOnlyAssignedTutorAndIdempotent(t *testing.T) {
	s, _ := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	actor := legacyUser(db, "regional")
	owner := legacyUser(db, "b")
	v := map[string]any{"id": "message-1", "text": "Gracias por perseverar."}
	if err := musicMutation602(db, actor, owner, "mentor602", v, "now", false); err != "" {
		t.Fatal(err)
	}
	if err := musicMutation602(db, actor, owner, "mentor602", v, "now", false); err != "" {
		t.Fatal(err)
	}
	if len(arrayField(musicData550(db), "mentoring602")) != 1 {
		t.Fatal("duplicate")
	}
	ensureMapField(owner, "tutorByProgram")[musicProgram553] = "tm"
	if err := musicMutation602(db, actor, owner, "mentor602", v, "now", false); err == "" {
		t.Fatal("former tutor posted")
	}
	if len(arrayField(musicView550(db, actor), "mentoring602")) != 0 {
		t.Fatal("former tutor reads note")
	}
	if len(arrayField(musicView550(db, owner), "mentoring602")) != 1 {
		t.Fatal("learner missing note")
	}
}
func Test602AdminGuardValidationAuditAndDisable(t *testing.T) {
	s, _ := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	admin := legacyUser(db, "admin")
	reg := legacyUser(db, "regional")
	v := map[string]any{"roles": map[string]any{"regional": map[string]any{"people.create": false}}, "users": map[string]any{}}
	if musicMutation602(db, reg, nil, "permissions602", v, "now", false) == "" {
		t.Fatal("non admin wrote")
	}
	if musicMutation602(db, admin, nil, "permissions602", v, "now", true) == "" {
		t.Fatal("personal admin wrote")
	}
	if err := musicMutation602(db, admin, nil, "permissions602", v, "now", false); err != "" {
		t.Fatal(err)
	}
	if musicPermit602(db, reg, "people.create") || len(arrayField(musicData550(db), "permissionsAudit602")) != 1 {
		t.Fatal("missing effective permission or audit")
	}
	mapField(mapField(v, "roles"), "regional")["invented"] = true
	if musicMutation602(db, admin, nil, "permissions602", v, "now", false) == "" {
		t.Fatal("unknown accepted")
	}
}
func Test602HTTPRevocationAndDriveGate(t *testing.T) {
	s, h := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	musicData550(db)["permissions602"] = map[string]any{"users": map[string]any{"regional": map[string]any{"people.create": false}}}
	s.saveLegacy(st)
	_, login := musicLogin553(t, h, "regional", true)
	token := stringValue(login["sessionToken"])
	code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", token, map[string]any{"actorId": "regional", "action": "create_aspirant", "restoreGeneration": 1, "value": map[string]any{}})
	if code != 403 {
		t.Fatal("direct create bypass", code)
	}
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/reports/drive/attempt", token, map[string]any{"actorId": "regional", "programId": musicProgram553, "ownerId": "b"})
	if code != 403 {
		t.Fatal("Drive disabled bypass", code)
	}
}
func Test602AchievementsOnlyApprovedAndScoped(t *testing.T) {
	s, _ := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	actor := legacyUser(db, "regional")
	p := reviewProgressContainer(db, "b", musicProgram553)
	p["status"] = map[string]any{}
	ensureMapField(p, "status")[musicRequirements550[0]] = "approved"
	ensureMapField(p, "status")[musicRequirements550[1]] = "in_review"
	out := musicView550(db, actor)
	m := mapField(out, "milestones602")
	if len(m) != 0 {
		t.Fatal("invalid milestone projection", m)
	}
}
func Test602CustomWhatsappDoesNotGrantOptIn(t *testing.T) {
	s, _ := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	a := legacyUser(db, "regional")
	b := legacyUser(db, "b")
	b["phone"] = "+59171234567"
	musicData550(db)["permissions602"] = map[string]any{"whatsappMessage": "Hola {nombre}, soy {responsable}."}
	if u, _ := musicContactURL598(db, a, b); u != "" {
		t.Fatal("missing opt in")
	}
	ensureMapField(musicData550(db), "interface598")["whatsappUsers"] = []any{"regional"}
	u, err := musicContactURL598(db, a, b)
	if err == "" || u != "" {
		t.Fatal(u, err)
	}
}
func Test602HTTPMentoringDurableAndPermissionRevoked(t *testing.T) {
	s, h := excelFixture560(t)
	_, login := musicLogin553(t, h, "regional", true)
	token := stringValue(login["sessionToken"])
	code, out := musicCall553(t, h, "GET", musicAPI553+"/api/music/module?actorId=regional", token, nil)
	if code != 200 {
		t.Fatal(code)
	}
	body := map[string]any{"actorId": "regional", "ownerId": "b", "action": "mentor602", "baseUpdatedAt": mapField(out, "data")["updatedAt"], "restoreGeneration": 1, "value": map[string]any{"id": "http-note", "text": "Tu esfuerzo cuenta."}}
	code, out = musicCall553(t, h, "POST", musicAPI553+"/api/music/module", token, body)
	if code != 200 {
		t.Fatal(code, out)
	}
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	if len(arrayField(musicData550(db), "mentoring602")) != 1 {
		t.Fatal("not saved")
	}
	musicData550(db)["permissions602"] = map[string]any{"users": map[string]any{"regional": map[string]any{"mentor.write": false}}}
	s.saveLegacy(st)
	body["baseUpdatedAt"] = mapField(out, "data")["updatedAt"]
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/music/module", token, body)
	if code != 403 {
		t.Fatal("revoked write permitted", code)
	}
}
func Test602SyncCannotBypassReportDenial(t *testing.T) {
	s, _ := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	actor := legacyUser(db, "regional")
	musicData550(db)["permissions602"] = map[string]any{"users": map[string]any{"regional": map[string]any{"reports.approve": false}}}
	p := reviewProgressContainer(db, "b", musicProgram553)
	ensureMapField(p, "status")[musicRequirements550[0]] = "in_review"
	candidate := legacyMap(cloneAny(db))
	ensureMapField(reviewProgressContainer(candidate, "b", musicProgram553), "status")[musicRequirements550[0]] = "approved"
	if musicSyncGuard602(db, actor, candidate) == "" {
		t.Fatal("bulk sync bypass")
	}
}
