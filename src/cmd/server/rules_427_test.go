package main

import "testing"

func TestSGL545AllowsAllLeadershipPrograms(t *testing.T) {
	db := map[string]any{
		"modules": []any{
			map[string]any{"id": "class_grouped", "assignedUserIds": []any{"leader1", "masterOnly"}},
			map[string]any{"id": "guide_major", "assignedUserIds": []any{"leader1"}},
			map[string]any{"id": "guide_master", "assignedUserIds": []any{"leader1", "masterOnly"}},
		},
		"classes": []any{
			map[string]any{"id": "class1", "assignedUserIds": []any{"leader1", "masterOnly", "normal"}},
		},
	}

	legacyEnforceSGL427Rules(db)

	mods := arrayField(db, "modules")
	var master map[string]any
	for _, raw := range mods {
		m := legacyMap(raw)
		if stringValue(m["id"]) == "guide_master" {
			master = m
		}
	}
	if master == nil {
		t.Fatal("guide_master module missing")
	}
	masterIDs := arrayField(master, "assignedUserIds")
	if !containsString(masterIDs, "leader1") {
		t.Fatal("all leadership program assignments must be preserved")
	}
	if !containsString(masterIDs, "masterOnly") {
		t.Fatal("valid guide_master assignment was removed")
	}

	classes := arrayField(db, "classes")
	if len(classes) != 1 {
		t.Fatalf("classes len=%d", len(classes))
	}
	classIDs := arrayField(legacyMap(classes[0]), "assignedUserIds")
	if !containsString(classIDs, "masterOnly") {
		t.Fatal("guide_master class instructor assignment must be preserved")
	}
	if !containsString(classIDs, "normal") {
		t.Fatal("normal class instructor was removed")
	}
}
