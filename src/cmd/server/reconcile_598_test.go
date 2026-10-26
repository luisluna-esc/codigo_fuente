package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func Test598ActualServerRecovery(t *testing.T) {
	path := os.Getenv("SGM598_SERVER")
	backup := os.Getenv("SGM598_BACKUP")
	if path == "" || backup == "" {
		t.Skip("private production fixture is opt-in and never packaged")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var original legacyStore
	if json.Unmarshal(b, &original) != nil {
		t.Fatal("server parse")
	}
	b, err = os.ReadFile(backup)
	if err != nil {
		t.Fatal(err)
	}
	var local map[string]any
	if json.Unmarshal(b, &local) != nil {
		t.Fatal("backup parse")
	}
	ops := arrayField(legacyMap(local["database"]), "syncQueue")
	m := manifest598()
	fresh := func() legacyStore { b, _ := json.Marshal(original); var v legacyStore; json.Unmarshal(b, &v); return v }
	next := fresh()
	before := legacyMap(cloneAny(next.State.Database))
	filesHash := hash598(next.State.Files)
	receipt, msg := reconcileState598(&next, m.ActorID, ops, "2026-10-12T21:30:00Z")
	if msg != "" {
		t.Fatal(msg)
	}
	db := legacyMap(next.State.Database)
	if next.State.ServerVersion != original.State.ServerVersion+1 || next.State.RestoreGeneration != 3 || next.State.RestoreID != original.State.RestoreID || hash598(next.State.Files) != filesHash {
		t.Fatal("state/file preservation")
	}
	if !musicAssigned550(db, franklin598) || !musicAssigned550(db, noemi598) || legacyUser(db, ghost598) != nil || len(legacyUsers(db)) != len(legacyUsers(before))-1 {
		t.Fatal("recovery did not take effect")
	}
	for _, raw := range legacyUsers(before) {
		u := legacyMap(raw)
		id := stringValue(u["id"])
		if id == ghost598 {
			continue
		}
		actual := legacyUser(db, id)
		old := legacyMap(cloneAny(u))
		if id == noemi598 {
			delete(old, "password")
			delete(old, "updatedAt")
			actual = legacyMap(cloneAny(actual))
			delete(actual, "password")
			delete(actual, "updatedAt")
		}
		if hash598(old) != hash598(actual) {
			t.Fatal("unrelated user field changed", id)
		}
	}
	for _, key := range []string{"progress", "bibleYear", "tasks", "materials", "music550", "clubs", "configs"} {
		if hash598(before[key]) != hash598(db[key]) {
			t.Fatal("unrelated database changed", key)
		}
	}
	for id, p := range legacyMap(before["programProgress"]) {
		if id != ghost598 && hash598(p) != hash598(legacyMap(db["programProgress"])[id]) {
			t.Fatal("progress changed", id)
		}
	}
	for _, raw := range arrayField(before, "modules") {
		mod := legacyMap(raw)
		if mod["id"] == musicProgram553 {
			continue
		}
		for _, v := range arrayField(db, "modules") {
			if legacyMap(v)["id"] == mod["id"] && hash598(v) != hash598(mod) {
				t.Fatal("other module changed")
			}
		}
	}
	if len(arrayField(receipt, "operationIds")) != 3 {
		t.Fatal("receipt")
	}
	nextVersion := next.State.ServerVersion
	legacyUser(db, noemi598)["password"] = map[string]any{"later": "new legitimate password"}
	snapshot := hash598(next)
	_, msg = reconcileState598(&next, m.ActorID, ops, "later")
	if msg != "" || next.State.ServerVersion != nextVersion || hash598(next) != snapshot {
		t.Fatal("retry overwrote newer data")
	}
	for _, kind := range []string{"generation", "password", "ghost", "operation"} {
		v := fresh()
		db := legacyMap(v.State.Database)
		o := ops
		switch kind {
		case "generation":
			v.State.RestoreGeneration = 4
		case "password":
			legacyUser(db, noemi598)["password"] = map[string]any{"new": "credential"}
		case "ghost":
			legacyUser(db, ghost598)["fullName"] = "Persona real"
		case "operation":
			o = append([]any{}, ops...)
			o = o[:2]
		}
		beforeHash := hash598(v)
		_, err := reconcileState598(&v, m.ActorID, o, "now")
		if err == "" || hash598(v) != beforeHash {
			t.Fatal("unsafe conflict", kind)
		}
	}
	// Exercise the authenticated HTTP route, including Music's route allowlist.
	srv, h := musicFixture553(t)
	if err := srv.saveLegacy(fresh()); err != nil {
		t.Fatal(err)
	}
	request := map[string]any{"actorId": m.ActorID, "operations": ops}
	code, _ := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/reconcile598", "fixture-client", request)
	if code == 200 {
		t.Fatal("non-admin token accepted")
	}
	code, body := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/reconcile598", "fixture-admin", request)
	if code != 200 || body["ok"] != true {
		t.Fatal("HTTP recovery failed", code, body["error"])
	}
	saved, _ := srv.loadLegacy()
	savedVersion := saved.State.ServerVersion
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/reconcile598", "fixture-admin", request)
	if code != 200 {
		t.Fatal("HTTP retry failed", code)
	}
	saved, _ = srv.loadLegacy()
	if saved.State.ServerVersion != savedVersion {
		t.Fatal("HTTP retry duplicated change")
	}
	t.Logf("PASS: 3 operations reconciled, %d files unchanged, generation 3 retained, other users and all existing progress preserved", len(next.State.Files))
	if out := os.Getenv("SGM598_RESULT"); out != "" {
		clean := fresh()
		_, err := reconcileState598(&clean, m.ActorID, ops, "2026-10-12T21:30:00Z")
		if err != "" {
			t.Fatal(err)
		}
		bytes, _ := json.Marshal(clean)
		if strings.Contains(out, "/private/") {
			if os.WriteFile(out, bytes, 0600) != nil {
				t.Fatal("write result")
			}
		}
	}
}
