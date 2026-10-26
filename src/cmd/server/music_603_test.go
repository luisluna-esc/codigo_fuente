package main

import (
	"testing"
	"time"
)

func Test603RemovedFeaturesCannotBeReenabled(t *testing.T) {
	db := map[string]any{"music550": map[string]any{"permissions602": map[string]any{"users": map[string]any{"admin": map[string]any{"whatsapp": true, "achievements": true}}}}}
	actor := map[string]any{"id": "admin", "role": "admin"}
	for _, key := range []string{"whatsapp", "achievements"} {
		if musicPermit602(db, actor, key) {
			t.Fatal("removed feature enabled", key)
		}
	}
	if musicPermit602(db, actor, "birthdays.reminders") || musicPermit602(db, actor, "ui.help") {
		t.Fatal("optional defaults must be off")
	}
	if !musicPermit602(db, actor, "birthdays.all") || !musicPermit602(db, actor, "birthdays.calendar") {
		t.Fatal("calendar defaults missing")
	}
}
func Test603BirthdayAllMusicAndScopedOverride(t *testing.T) {
	learner := map[string]any{"id": "a", "role": "aspirant", "birthDate": "1980-02-04", "fullName": "A"}
	other := map[string]any{"id": "b", "role": "aspirant", "birthDate": "1990-05-06", "fullName": "B"}
	db := map[string]any{"users": []any{learner, other}, "modules": []any{map[string]any{"id": musicProgram553, "assignedUserIds": []any{"a", "b"}}}}
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	if got := birthdayDirectory562(db, learner, "music", now); len(got) != 2 {
		t.Fatal(got)
	}
	musicData550(db)["permissions602"] = map[string]any{"users": map[string]any{"a": map[string]any{"birthdays.all": false}}}
	if got := birthdayDirectory562(db, learner, "music", now); len(got) != 1 {
		t.Fatal(got)
	}
}
func Test603SelfPlanningDeniedWithoutChangingExistingTasks(t *testing.T) {
	actor := map[string]any{"id": "a", "role": "aspirant"}
	old := map[string]any{"id": "t", "programId": musicProgram553, "createdBy": "a", "assigneeId": "a", "selfPlanned": true, "dueDate": "2026-10-20"}
	db := map[string]any{"tasks": []any{old}, "music550": map[string]any{"permissions602": map[string]any{"users": map[string]any{"a": map[string]any{"tasks.selfPlan": false}}}}}
	candidate := legacyMap(cloneAny(db))
	if reason := musicSyncGuard602(db, actor, candidate); reason != "" {
		t.Fatal("unchanged task blocked", reason)
	}
	legacyMap(arrayField(candidate, "tasks")[0])["dueDate"] = "2026-10-21"
	if musicSyncGuard602(db, actor, candidate) == "" {
		t.Fatal("self-plan edit bypassed denial")
	}
}
