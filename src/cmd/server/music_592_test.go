package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"testing"
	"time"
)

func photo592() string {
	var b bytes.Buffer
	png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 200, 100)))
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(b.Bytes())
}
func TestMusicPhoto592OwnershipAndVersion(t *testing.T) {
	s, h := excelFixture560(t)
	_, login := musicLogin553(t, h, "a", true)
	token := stringValue(login["sessionToken"])
	st, _ := s.loadLegacy()
	u := legacyUser(st.State.Database, "a")
	body := map[string]any{"actorId": "a", "ownerId": "a", "action": "self_photo", "restoreGeneration": 1, "expectedUpdatedAt": u["updatedAt"], "value": map[string]any{"photo": photo592()}}
	code, p := musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", token, body)
	if code != 200 {
		t.Fatal(code, p)
	}
	st, _ = s.loadLegacy()
	if stringValue(legacyUser(st.State.Database, "a")["photo"]) == "" {
		t.Fatal("photo not saved")
	}
	body["ownerId"] = "b"
	body["expectedUpdatedAt"] = legacyUser(st.State.Database, "b")["updatedAt"]
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/accounts/music", token, body)
	if code != 403 {
		t.Fatal("other photo allowed", code)
	}
}
func TestMusicPoints592(t *testing.T) {
	s, _ := excelFixture560(t)
	st, _ := s.loadLegacy()
	db := legacyMap(st.State.Database)
	p := reviewProgressContainer(db, "a", musicProgram553)
	now := time.Date(2026, 10, 9, 18, 15, 0, 0, time.Local)
	at := now.Add(-2 * time.Hour).Format(time.RFC3339Nano)
	for i, state := range []string{"in_progress", "in_review", "approved", "assigned", "in_progress"} {
		rid := musicRequirements550[i]
		ensureMapField(p, "status")[rid] = state
		sub := map[string]any{"updatedAt": at, "submittedAt": at}
		if i != 4 {
			sub["text"] = "Avance parcial"
		}
		if i == 2 {
			sub["approval"] = map[string]any{"at": at}
		}
		ensureMapField(p, "submissions")[rid] = sub
	}
	db["audit"] = []any{map[string]any{"actorId": "a", "module": musicProgram553, "action": "Navegar", "at": at}}
	items := musicDailyItems592(db, "2026-10-09")
	if len(items) != 1 || items[0].Points != 7 {
		t.Fatal(items)
	}
	if ensureMusicDaily592(&st, now.Add(-2*time.Hour)) {
		t.Fatal("early notification")
	}
	if !ensureMusicDaily592(&st, now) {
		t.Fatal("no notification")
	}
	n := len(st.Notifications)
	ensureMusicDaily592(&st, now)
	if len(st.Notifications) != n {
		t.Fatal("duplicate")
	}
	if len(musicDailyItems592(db, "2026-10-10")) != 0 {
		t.Fatal("stale points")
	}
	if !expireLegacyDailyActivityNotifications(&st, now.Add(2*time.Hour)) {
		t.Fatal("not expired")
	}
}
func TestMusicTemplate592Fields(t *testing.T) {
	s, h := excelFixture560(t)
	_, login := musicLogin553(t, h, "admin", true)
	token := stringValue(login["sessionToken"])
	field := map[string]any{"text": "{nombre}", "x": float64(10), "y": float64(40), "width": float64(80), "size": float64(30), "font": "Georgia", "align": "center", "color": "#193353"}
	v := map[string]any{"kind": "final", "dataUrl": photo592(), "fields": []any{field}, "qrX": 84, "qrY": 74}
	body := map[string]any{"actorId": "admin", "action": "template", "value": v, "restoreGeneration": 1}
	code, p := musicCall553(t, h, "POST", musicAPI553+"/api/music/module", token, body)
	if code != 200 {
		t.Fatal(code, p)
	}
	st, _ := s.loadLegacy()
	stored := legacyMap(mapField(musicData550(legacyMap(st.State.Database)), "templates")["final"])
	if len(arrayField(stored, "fields")) != 1 {
		t.Fatal(stored)
	}
	field["size"] = float64(-10)
	code, _ = musicCall553(t, h, "POST", musicAPI553+"/api/music/module", token, body)
	if code == 200 {
		t.Fatal("invalid size accepted")
	}
}
