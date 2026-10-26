package main

import "testing"

func bible524DB(status string, readings map[string]any) map[string]any {
	mirror := map[string]any{}
	for id, value := range readings {
		mirror[id] = value
	}
	journal := map[string]any{}
	for id, value := range readings {
		journal[id] = value
	}
	return map[string]any{
		"users": []any{map[string]any{"id": "asp1", "role": "aspirant", "active": true, "accountState": "active"}},
		"progress": map[string]any{"asp1": map[string]any{
			"status":   map[string]any{"II-13": status},
			"readings": map[string]any{"II-13": journal},
			"submissions": map[string]any{"II-13": map[string]any{
				"bibleReadings": mirror,
				"updatedAt":     "2026-09-18T13:00:00Z",
			}},
		}},
		"programProgress": map[string]any{}, "bibleYear": map[string]any{}, "configs": map[string]any{},
		"tasks": []any{}, "materials": []any{}, "indicationRules": []any{}, "classes": []any{}, "audit": []any{},
	}
}

func bible524Source(deleteID string) map[string]any {
	return map[string]any{"progress": map[string]any{"asp1": map[string]any{
		"updatedAt": "2026-09-18T14:00:00Z",
		"submissions": map[string]any{"II-13": map[string]any{
			"bibleReadingDeletedIds": []any{deleteID},
			"updatedAt":              "2026-09-18T14:00:00Z",
		}},
	}}}
}

func TestSGL524BibleDeleteRemovesOnlyRequestedChapterAndKeepsDevelopment(t *testing.T) {
	readings := map[string]any{
		"II-13-L001": map[string]any{"completed": true, "passage": "Exo. 1"},
		"II-13-L002": map[string]any{"completed": true, "passage": "Exo. 2"},
	}
	merged := bible524DB("in_progress", readings)
	proposal := &legacyProposal{ActorID: "asp1", ActorRole: "aspirant", OwnerID: "asp1", RequirementID: "II-13", ProgramID: "class_grouped", OperationType: "bible_reading_delete"}
	if !mergeBibleReadingDelete(merged, bible524Source("II-13-L001"), proposal, "aspirant", map[string]bool{"asp1": true}, "2026-09-18T14:00:00Z") {
		t.Fatal("expected bible reading deletion to be applied")
	}
	progress := legacyMap(legacyMap(merged["progress"])["asp1"])
	submission := legacyMap(legacyMap(progress["submissions"])["II-13"])
	serverReadings := legacyMap(submission["bibleReadings"])
	journal := legacyMap(legacyMap(progress["readings"])["II-13"])
	if _, ok := serverReadings["II-13-L001"]; ok {
		t.Fatal("deleted reading remained in canonical submission")
	}
	if _, ok := journal["II-13-L001"]; ok {
		t.Fatal("deleted reading remained in canonical journal")
	}
	if _, ok := serverReadings["II-13-L002"]; !ok {
		t.Fatal("unrelated reading was removed from canonical submission")
	}
	if _, ok := journal["II-13-L002"]; !ok {
		t.Fatal("unrelated reading was removed from canonical journal")
	}
	if got := stringValue(legacyMap(progress["status"])["II-13"]); got != "in_progress" {
		t.Fatalf("status=%q, want in_progress", got)
	}
	if _, ok := submission["bibleReadingDeletedIds"]; ok {
		t.Fatal("transient deletion marker must not be persisted")
	}
}

func TestSGL524BibleDeleteLastChapterReturnsRequirementToPending(t *testing.T) {
	merged := bible524DB("in_progress", map[string]any{
		"II-13-L001": map[string]any{"completed": true, "passage": "Exo. 1"},
	})
	proposal := &legacyProposal{ActorID: "asp1", ActorRole: "aspirant", OwnerID: "asp1", RequirementID: "II-13", ProgramID: "class_grouped", OperationType: "bible_reading_delete"}
	if !mergeBibleReadingDelete(merged, bible524Source("II-13-L001"), proposal, "aspirant", map[string]bool{"asp1": true}, "2026-09-18T14:00:00Z") {
		t.Fatal("expected last reading deletion to be applied")
	}
	progress := legacyMap(legacyMap(merged["progress"])["asp1"])
	if got := stringValue(legacyMap(progress["status"])["II-13"]); got != "pending" {
		t.Fatalf("status=%q, want pending after deleting final chapter", got)
	}
}

func TestSGL524BibleDeleteCannotChangeApprovedRequirement(t *testing.T) {
	merged := bible524DB("approved", map[string]any{
		"II-13-L001": map[string]any{"completed": true, "passage": "Exo. 1"},
	})
	proposal := &legacyProposal{ActorID: "asp1", ActorRole: "aspirant", OwnerID: "asp1", RequirementID: "II-13", ProgramID: "class_grouped", OperationType: "bible_reading_delete"}
	if mergeBibleReadingDelete(merged, bible524Source("II-13-L001"), proposal, "aspirant", map[string]bool{"asp1": true}, "2026-09-18T14:00:00Z") {
		t.Fatal("approved requirement must reject Bible reading deletion")
	}
}
