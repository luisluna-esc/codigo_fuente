package main

import (
	"testing"
	"time"
)

func Test58FollowupNeverAutomaticallyBlocks(t *testing.T) {
	db, u, now := accessDB571()
	rec := &accessRecord571{BlockedAt: now.Add(-24 * time.Hour).Format(time.RFC3339), Drained: true}
	status := accessStatus571(db, u, "conquis", rec, now)
	if status["blocked"] != false || status["canExtend"] != false || rec.BlockedAt != "" {
		t.Fatal(status, rec)
	}
	rows := status["followup"].([]any)
	if len(rows) != 1 || rows[0].(map[string]any)["stage"] != 2 {
		t.Fatal(rows)
	}
	u["createdAt"] = now.AddDate(0, 0, -8).Format(time.RFC3339)
	if followup572(db, u, "conquis", now)[0].(map[string]any)["stage"] != 1 {
		t.Fatal("missing early warning")
	}
	u["createdAt"] = now.AddDate(0, 0, -2).Format(time.RFC3339)
	if followup572(db, u, "conquis", now)[0].(map[string]any)["stage"] != 0 {
		t.Fatal("new account warning")
	}
	if len(followup572(db, u, "music", now)) != 0 {
		t.Fatal("music followup changed")
	}
}

func Test58NinetyPercentSuppressesWarningButNotManualBlock(t *testing.T) {
	db, u, now := accessDB571()
	p := reviewProgressContainer(db, "learner", "guide_major")
	ids := requirements571["guide_major"]
	for _, id := range ids[:(len(ids)*9+9)/10] {
		ensureMapField(p, "status")[id] = "in_review"
	}
	if followup572(db, u, "conquis", now)[0].(map[string]any)["stage"] != 0 {
		t.Fatal("90 percent still warned")
	}
	status := accessStatus571(db, u, "conquis", &accessRecord571{Manual: true}, now)
	if status["blocked"] != true || status["canExtend"] != true {
		t.Fatal(status)
	}
}

func Test58ManualExtensionPersistsFor72HoursAndCannotRepeat(t *testing.T) {
	s, _ := fixture550(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	session := s.remoteSessions[hashToken("p")]
	session.Role = legacyRole(legacyUser(db, "p"))
	s.remoteSessions[hashToken("p")] = session
	db["modules"] = append(arrayField(db, "modules"), map[string]any{"id": "guide_major", "active": true, "assignedUserIds": []any{"p"}})
	if err := s.saveLegacy(st); err != nil {
		t.Fatal(err)
	}
	if w := accessCall571(t, s, "admin", "conquis", "block", []string{"p"}); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	start := time.Now().UTC()
	if w := accessCall571(t, s, "p", "conquis", "extend", nil); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	data, err := s.loadAccess571()
	if err != nil {
		t.Fatal(err)
	}
	rec := data["conquis"]["p"]
	until := accessTime571(rec.ExtensionUntil)
	if until.Before(start.Add(72*time.Hour)) || until.After(time.Now().Add(72*time.Hour)) || !rec.ExtensionUsed {
		t.Fatal(rec)
	}
	if w := accessCall571(t, s, "p", "conquis", "extend", nil); w.Code != 403 {
		t.Fatal("repeated extension", w.Code)
	}
	user := legacyUser(db, "p")
	if status := accessStatus571(db, user, "conquis", rec, start); status["blocked"] != false {
		t.Fatal(status)
	}
	if status := accessStatus571(db, user, "conquis", rec, until.Add(time.Second)); status["blocked"] != true || status["canExtend"] != false {
		t.Fatal(status)
	}
	rec.Disabled = true
	if status := accessStatus571(db, user, "conquis", rec, start); status["blocked"] != true || status["canExtend"] != false {
		t.Fatal("disabled bypass", status)
	}
	if w := accessCall571(t, s, "admin", "conquis", "release", []string{"p"}); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	data, _ = s.loadAccess571()
	if data["conquis"]["p"].Manual || data["conquis"]["p"].ExtensionUntil != "" {
		t.Fatal("release did not persist")
	}
	if w := accessCall571(t, s, "p", "music", "extend", nil); w.Code != 403 {
		t.Fatal("music extension", w.Code)
	}
}
