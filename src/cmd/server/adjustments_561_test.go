package main

import (
	"reflect"
	"testing"
)

func Test561DeletionMatrix(t *testing.T) {
	s, _ := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	ids := []string{"admin", "departamental", "secretaria", "coordinador", "administrativo", "regional", "tm", "a"}
	expected := map[string]map[string]bool{
		"admin":         {"departamental": true, "secretaria": true, "coordinador": true, "administrativo": true, "regional": true, "tm": true, "a": true},
		"departamental": {"secretaria": true, "coordinador": true, "administrativo": true, "regional": true, "tm": true, "a": true},
		"secretaria":    {"coordinador": true, "administrativo": true, "regional": true, "tm": true, "a": true},
		"coordinador":   {"administrativo": true, "regional": true, "tm": true, "a": true},
		"regional":      {},
	}
	for _, a := range ids {
		for _, o := range ids {
			if got := musicDelete560(db, legacyUser(db, a), legacyUser(db, o)); got != expected[a][o] {
				t.Errorf("%s removes %s: %v", a, o, got)
			}
		}
	}
	if musicDelete560(db, legacyUser(db, "regional"), legacyUser(db, "d")) {
		t.Fatal("foreign region")
	}
}
func Test561RemovePreservesCardAndUnassigns(t *testing.T) {
	s, h := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	before := cloneAny(legacyProgressForRequirement(db, "departamental", musicProgram553, false))
	legacyUser(db, "a")["tutorByProgram"] = map[string]any{musicProgram553: "departamental", "guide_major": "tg"}
	s.saveLegacy(st)
	_, login := musicLogin553(t, h, "admin", true)
	token := stringValue(login["sessionToken"])
	req := map[string]any{"actorId": "admin", "ownerId": "departamental", "action": "remove", "restoreGeneration": 1, "expectedUpdatedAt": ""}
	code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", token, req)
	if code != 400 {
		t.Fatal("region required", code)
	}
	req["value"] = map[string]any{"region": "MBOS"}
	code, out := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", token, req)
	if code != 200 {
		t.Fatal(code, out)
	}
	st, _ = s.loadLegacy()
	db = legacyMap(st.State.Database)
	u := legacyUser(db, "departamental")
	if legacyRole(u) != "departamental" || u["musicManagementRemoved597"] != true || u["region"] != "MBOS" || !musicAssigned550(db, "departamental") || !reflect.DeepEqual(before, legacyProgressForRequirement(db, "departamental", musicProgram553, false)) {
		t.Fatal("card lost", u)
	}
	if !musicUnassignedMember561(db, legacyUser(db, "a")) || legacyTutorAssigned(legacyUser(db, "a"), "departamental", musicProgram553) || !legacyTutorAssigned(legacyUser(db, "a"), "tg", "guide_major") {
		t.Fatal("dependent not safely unassigned")
	}
	_, login = musicLogin553(t, h, "regional", true)
	token = stringValue(login["sessionToken"])
	code, q := musicCall553(t, h, "GET", musicAPI553+"/api/accounts/unassigned?actorId=regional", token, nil)
	if code != 200 || len(arrayField(q, "rows")) != 0 {
		t.Fatal(code, q)
	}
	code, q = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/unassigned", token, map[string]any{"actorId": "regional", "ownerId": "departamental", "restoreGeneration": 1})
	if code != 403 {
		t.Fatal(code, q)
	}
	st, _ = s.loadLegacy()
	if legacyUser(st.State.Database, "departamental")["region"] != "MBOS" {
		t.Fatal("request changed region")
	}
}
func Test561AssignmentPermissions(t *testing.T) {
	for _, id := range []string{"admin", "departamental", "secretaria", "coordinador", "administrativo", "regional", "tm", "a"} {
		t.Run(id, func(t *testing.T) {
			s, h := excelFixture560(t)
			_, login := musicLogin553(t, h, id, true)
			st, _ := s.loadLegacy()
			u := legacyUser(st.State.Database, "a")
			code, out := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", stringValue(login["sessionToken"]), map[string]any{"actorId": id, "ownerId": "a", "action": "assign_tutor", "restoreGeneration": 1, "expectedUpdatedAt": u["updatedAt"], "value": map[string]any{"tutorId": "regional", "region": "MBOS"}})
			allowed := id == "admin" || id == "departamental" || id == "secretaria" || id == "coordinador"
			if (code == 200) != allowed {
				t.Fatal(code, out)
			}
		})
	}
}
func Test561ImportAtomicPreviewAndRetry(t *testing.T) {
	s, h := excelFixture560(t)
	_, login := musicLogin553(t, h, "admin", true)
	token := stringValue(login["sessionToken"])
	row := map[string]any{"id": "import-new", "tutorId": "tm", "fullName": "María López", "email": "maria@example.test", "region": "MBOS", "role": "secretaria", "programId": musicProgram553, "training": true, "birthDate": "1990-01-02", "password": tutorPassword556()}
	req := map[string]any{"actorId": "admin", "action": "preview", "batchId": "batch561", "restoreGeneration": 1, "rows": []any{row}}
	code, out := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/import", token, req)
	if code != 200 || out["ok"] != true {
		t.Fatal(code, out)
	}
	st, _ := s.loadLegacy()
	if legacyUser(st.State.Database, "import-new") != nil {
		t.Fatal("preview wrote user")
	}
	req["action"] = "commit"
	code, out = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/import", token, req)
	if code != 200 || out["ok"] != true {
		t.Fatal(code, out)
	}
	code, out = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/import", token, req)
	if code != 200 || out["alreadyImported"] != true {
		t.Fatal("retry", code, out)
	}
	st, _ = s.loadLegacy()
	u := legacyUser(st.State.Database, "import-new")
	if u["mission"] != "Misión Boliviana Occidental Sur" || u["role"] != "secretaria" || !musicAssigned550(legacyMap(st.State.Database), "import-new") {
		t.Fatal(u)
	}
	req["action"] = "preview"
	req["batchId"] = "other"
	code, out = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/import", token, req)
	if out["ok"] != false {
		t.Fatal("duplicate accepted", code, out)
	}
	_, login = musicLogin553(t, h, "secretaria", true)
	req["actorId"] = "secretaria"
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/import", stringValue(login["sessionToken"]), req)
	if code != 403 {
		t.Fatal("nonadmin import", code)
	}
}
