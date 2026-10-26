package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

func assignmentFixture559(t *testing.T) (*Server, http.Handler) {
	s, h := musicFixture553(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		u["email"] = stringValue(u["id"]) + "@example.test"
	}
	for _, id := range []string{"class_grouped", "guide_master", "guide_master_advanced"} {
		db["modules"] = append(arrayField(db, "modules"), map[string]any{"id": id, "active": true, "assignedUserIds": []any{}})
	}
	legacyUser(db, "departamental")["managementProgramIds"] = []any{musicProgram553, "guide_major"}
	legacyUser(db, "departamental")["tutorByProgram"] = map[string]any{"guide_major": "tg"}
	s.saveLegacy(st)
	return s, h
}

func Test559RemoteMusicAdminAssignsExistingAuthorityPersonalCard(t *testing.T) {
	s, h := assignmentFixture559(t)
	_, login := musicLogin553(t, h, "admin", true)
	_, department := musicLogin553(t, h, "departamental", true)
	token := stringValue(login["sessionToken"])
	departmentToken := stringValue(department["sessionToken"])
	path := musicAPI553 + "/api/accounts/roster?actorId=departamental"
	_, before := musicCall553(t, h, "GET", path, departmentToken, nil)
	patch := map[string]any{
		"modules":         []any{map[string]any{"id": musicProgram553, "assignedUserIds": []any{"a", "b", "departamental"}}},
		"users":           []any{map[string]any{"id": "departamental", "managementProgramIds": []any{musicProgram553}, "tutorByProgram": map[string]any{musicProgram553: "secretaria"}}},
		"programProgress": map[string]any{"departamental": map[string]any{musicProgram553: map[string]any{"status": map[string]any{}, "submissions": map[string]any{}}}},
	}
	code, result := musicCall553(t, h, "POST", musicAPI553+"/api/admin/changes", token, map[string]any{"actorId": "admin", "operationId": "first-personal-music", "programId": musicProgram553, "entityId": "departamental", "patch": patch, "ackOnly": true})
	if code != http.StatusOK {
		t.Fatalf("remote admin assignment rejected: %d %v", code, result)
	}
	st, _ := s.loadLegacy()
	if !musicAssigned550(legacyMap(st.State.Database), "departamental") {
		t.Fatal("personal assignment not stored")
	}
	u := legacyUser(st.State.Database, "departamental")
	if !containsString(arrayField(u, "managementProgramIds"), "guide_major") || stringValue(legacyMap(u["tutorByProgram"])["guide_major"]) != "tg" {
		t.Fatal("Music assignment erased other module permissions or tutor")
	}
	code, after := musicCall553(t, h, "GET", path+"&knownHash="+stringValue(before["rosterHash"]), departmentToken, nil)
	if code != 200 || after["unchanged"] == true || !musicAssigned550(legacyMap(after["database"]), "departamental") {
		t.Fatal("existing session did not receive assignment", code, after)
	}
	code, boot := musicCall553(t, h, "GET", musicAPI553+"/api/remote/bootstrap", departmentToken, nil)
	if code != 200 || !musicAssigned550(legacyMap(boot["database"]), "departamental") {
		t.Fatal("bootstrap did not receive assignment", code)
	}
}

func Test559ConcurrentMembershipChangesPreserveUnrelatedAssignments(t *testing.T) {
	for _, program := range []string{"class_grouped", "guide_major", "guide_master", "guide_master_advanced", musicProgram553} {
		t.Run(program, func(t *testing.T) {
			s, h := assignmentFixture559(t)
			_, login := musicLogin553(t, h, "admin", false)
			token := stringValue(login["sessionToken"])
			for index, id := range []string{"departamental", "secretaria"} {
				patch := map[string]any{"modules": []any{map[string]any{"id": program, "assignedUserIds": map[string]any{"__sglSetMembers": map[string]any{"add": []any{id}, "remove": []any{}}}}}}
				code, out := musicCall553(t, h, "POST", "/api/admin/changes", token, map[string]any{"actorId": "admin", "operationId": fmt.Sprintf("assign-%s-%d", program, index), "programId": program, "patch": patch})
				if code != 200 {
					t.Fatal(code, out)
				}
			}
			st, _ := s.loadLegacy()
			db := legacyMap(st.State.Database)
			if !assignedProgram555(db, "departamental", program) || !assignedProgram555(db, "secretaria", program) {
				t.Fatal("concurrent assignment lost an unrelated membership")
			}
		})
	}
}

func Test559MusicPermissionEditsPreserveImplicitAndExplicitOtherPrograms(t *testing.T) {
	for _, implicit := range []bool{true, false} {
		t.Run(fmt.Sprint(implicit), func(t *testing.T) {
			s, h := assignmentFixture559(t)
			st, _ := s.loadLegacy()
			db := legacyMap(st.State.Database)
			u := legacyUser(db, "departamental")
			if implicit {
				delete(u, "managementProgramIds")
			} else {
				u["managementProgramIds"] = []any{musicProgram553, "guide_major", "guide_master"}
			}
			s.saveLegacy(st)
			_, login := musicLogin553(t, h, "admin", true)
			patch := map[string]any{"users": []any{map[string]any{"id": "departamental", "managementProgramIds": map[string]any{"__sglSetMembers": map[string]any{"remove": []any{musicProgram553}, "add": []any{}}}, "tutorByProgram": map[string]any{"__sglReplace": map[string]any{musicProgram553: "secretaria"}}}}}
			code, out := musicCall553(t, h, "POST", musicAPI553+"/api/admin/changes", stringValue(login["sessionToken"]), map[string]any{"actorId": "admin", "operationId": "remove-music-management", "programId": musicProgram553, "patch": patch})
			if code != 200 {
				t.Fatal(code, out)
			}
			st, _ = s.loadLegacy()
			db = legacyMap(st.State.Database)
			u = legacyUser(db, "departamental")
			if managementProgram555(db, u, musicProgram553) || managementProgram555(db, u, "guide_major") || managementProgram555(db, u, "guide_master") {
				t.Fatal("Music permission edit changed the other module permissions", u)
			}
			if implicit && managementProgram555(db, u, "guide_master_advanced") {
				t.Fatal("implicit legacy scope was erased")
			}
			if !implicit && managementProgram555(db, u, "guide_master_advanced") {
				t.Fatal("explicit legacy scope was widened")
			}
			if stringValue(legacyMap(u["tutorByProgram"])["guide_major"]) != "tg" {
				t.Fatal("projected tutor replacement erased another program")
			}
		})
	}
}

func Test559LostAdminAcknowledgementIsIdempotent(t *testing.T) {
	s, h := assignmentFixture559(t)
	_, login := musicLogin553(t, h, "admin", true)
	body := map[string]any{"actorId": "admin", "operationId": "original-frozen-559", "programId": musicProgram553, "restoreGeneration": 1, "patch": map[string]any{"modules": []any{map[string]any{"id": musicProgram553, "assignedUserIds": map[string]any{"__sglSetMembers": map[string]any{"add": []any{"secretaria"}, "remove": []any{}}}}}}}
	code, out := musicCall553(t, h, "POST", musicAPI553+"/api/admin/changes", stringValue(login["sessionToken"]), body)
	if code != 200 {
		t.Fatal(code, out)
	}
	version := out["serverVersion"]
	code, out = musicCall553(t, h, "POST", musicAPI553+"/api/admin/changes", stringValue(login["sessionToken"]), body)
	if code != 200 || out["deduplicated"] != true || out["operationId"] != "original-frozen-559" || out["serverVersion"] != version {
		t.Fatal("retry duplicated or lost the original operation", code, out)
	}
	st, _ := s.loadLegacy()
	count := 0
	for _, m := range arrayField(legacyMap(st.State.Database), "modules") {
		if stringValue(legacyMap(m)["id"]) == musicProgram553 {
			for _, id := range arrayField(legacyMap(m), "assignedUserIds") {
				if id == "secretaria" {
					count++
				}
			}
		}
	}
	if count != 1 {
		t.Fatal("retried membership appeared more than once")
	}
}

func Test559InvalidScopeAndStaleRestoreCannotApplyAssignments(t *testing.T) {
	s, h := assignmentFixture559(t)
	_, admin := musicLogin553(t, h, "admin", true)
	_, tutor := musicLogin553(t, h, "tm", true)
	body := map[string]any{"actorId": "admin", "operationId": "scope-559", "programId": musicProgram553, "restoreGeneration": 1, "patch": map[string]any{"modules": []any{map[string]any{"id": "guide_master", "assignedUserIds": []any{"secretaria"}}}}}
	if code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/admin/changes", stringValue(admin["sessionToken"]), body); code != 403 {
		t.Fatal("Music changed a foreign module", code)
	}
	body["patch"] = map[string]any{"modules": []any{map[string]any{"id": musicProgram553, "assignedUserIds": []any{"secretaria"}}}}
	body["actorId"] = "tm"
	if code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/admin/changes", stringValue(tutor["sessionToken"]), body); code != 403 {
		t.Fatal("tutor gained admin assignment rights", code)
	}
	body["actorId"] = "admin"
	st, _ := s.loadLegacy()
	st.State.RestoreGeneration = 2
	s.saveLegacy(st)
	code, out := musicCall553(t, h, "POST", musicAPI553+"/api/admin/changes", stringValue(admin["sessionToken"]), body)
	if code != 409 || out["code"] != "RESTORE_REQUIRED" {
		t.Fatal("stale operation crossed a restore", code, out)
	}
	st, _ = s.loadLegacy()
	if musicAssigned550(legacyMap(st.State.Database), "secretaria") {
		t.Fatal("stale assignment was applied")
	}
}

func Test559LiveReceptionIncludesProgressInEveryModule(t *testing.T) {
	for _, pair := range [][2]string{{"class_grouped", "R1"}, {"guide_major", "GM-NUT-1"}, {"guide_master", "GMM-PRE-1"}, {"guide_master_advanced", "GMMA-PRE-1"}, {musicProgram553, "LM-PRE-1"}} {
		t.Run(pair[0], func(t *testing.T) {
			s, h := assignmentFixture559(t)
			st, _ := s.loadLegacy()
			db := legacyMap(st.State.Database)
			for _, raw := range arrayField(db, "modules") {
				m := legacyMap(raw)
				if stringValue(m["id"]) == pair[0] && !containsString(arrayField(m, "assignedUserIds"), "a") {
					m["assignedUserIds"] = append(arrayField(m, "assignedUserIds"), "a")
				}
			}
			legacyUser(db, "tm")["tutorProgramIds"] = []any{pair[0], "class_grouped"}
			ensureMapField(legacyUser(db, "a"), "tutorByProgram")[pair[0]] = "tm"
			s.saveLegacy(st)
			_, login := musicLogin553(t, h, "a", false)
			_, tutor := musicLogin553(t, h, "tm", false)
			path := "/api/accounts/roster?actorId=tm"
			_, before := musicCall553(t, h, "GET", path, stringValue(tutor["sessionToken"]), nil)
			p := map[string]any{"status": map[string]any{pair[1]: "in_progress"}, "submissions": map[string]any{pair[1]: map[string]any{"text": "Informe guardado desde otro equipo.\nUna segunda línea de contenido.\nUna tercera línea conservada.", "updatedAt": "2026-10-04T18:00:00Z"}}}
			patch := map[string]any{"programProgress": map[string]any{"a": map[string]any{pair[0]: p}}}
			if pair[0] == "class_grouped" {
				patch = map[string]any{"progress": map[string]any{"a": p}}
			}
			code, out := musicCall553(t, h, "POST", "/api/sync", stringValue(login["sessionToken"]), map[string]any{"actorId": "a", "actorRole": "aspirant", "installationId": "second-device", "operationId": "save-" + pair[0], "operationType": "save_progress", "programId": pair[0], "ownerId": "a", "requirementId": pair[1], "baseVersion": st.State.ServerVersion, "restoreGeneration": 1, "patch": patch, "files": []any{}, "ackOnly": true})
			if code != 200 || out["operationId"] != "save-"+pair[0] {
				t.Fatal("save rejected", code, out)
			}
			code, after := musicCall553(t, h, "GET", path+"&knownHash="+stringValue(before["rosterHash"]), stringValue(tutor["sessionToken"]), nil)
			if code != 200 || after["unchanged"] == true {
				t.Fatal("progress was not signaled to the open session", code, after)
			}
			progress := legacyProgressForRequirement(after["database"], "a", pair[0], false)
			if stringValue(mapField(progress, "status")[pair[1]]) != "in_progress" {
				t.Fatal("new report missing from live snapshot")
			}
			_, same := musicCall553(t, h, "GET", path+"&knownHash="+stringValue(after["rosterHash"]), stringValue(tutor["sessionToken"]), nil)
			if same["unchanged"] != true {
				t.Fatal("unchanged state was downloaded again")
			}
		})
	}
}

func TestBrowserFixture559(t *testing.T) {
	if os.Getenv("SGL559_BROWSER_FIXTURE") != "1" {
		t.Skip("interactive fixture")
	}
	_, h := assignmentFixture559(t)
	var previous atomic.Bool
	oldFiles := http.FileServer(http.Dir(os.Getenv("SGL559_PREVIOUS_APP")))
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/__test__/assets" {
			previous.Store(r.URL.Query().Get("previous") == "1")
			w.WriteHeader(204)
			return
		}
		if previous.Load() && !strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, musicAPI553+"/") {
			oldFiles.ServeHTTP(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
	ln, err := net.Listen("tcp", "127.0.0.1:8900")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	tlsServer := httptest.NewUnstartedServer(wrapped)
	tlsListener, err := net.Listen("tcp", "127.0.0.1:8901")
	if err != nil {
		t.Fatal(err)
	}
	tlsServer.Listener = tlsListener
	tlsServer.StartTLS()
	defer tlsServer.Close()
	fmt.Println("SGL559_BROWSER_READY")
	http.Serve(ln, wrapped)
}
