package main

import "testing"

func TestBibleVersion519IsSharedRemoteResource(t *testing.T) {
	db := map[string]any{
		"users": []any{
			map[string]any{"id": "admin1", "role": "admin"},
			map[string]any{"id": "asp1", "role": "aspirant"},
		},
		"bibleVersions": []any{
			map[string]any{"id": "bible1", "fileId": "bible-file", "name": "Version autorizada"},
		},
	}
	owners, shared := remoteFilePolicy(db, "asp1", "aspirant")
	if !owners["asp1"] {
		t.Fatal("aspirant owner scope is missing")
	}
	if !shared["bible-file"] {
		t.Fatal("administrator Bible version must be readable as a shared resource")
	}
	if !remoteFileAllowed(map[string]any{"id": "bible-file", "ownerId": "admin1"}, owners, shared) {
		t.Fatal("shared Bible file should be available to the authenticated aspirant")
	}
}

func TestSafeReviewSubmission519KeepsBibleJournal(t *testing.T) {
	input := map[string]any{
		"text":                "Genesis 1",
		"bibleReadingVersion": "5.19",
		"bibleVersionId":      "bible1",
		"bibleVersionName":    "Version autorizada",
		"bibleReadings": map[string]any{
			"L001": map[string]any{
				"completed":      true,
				"favoriteVerses": "Genesis 1:1",
				"impactStory":    "La creacion",
				"lifeLesson":     "Dios tiene proposito",
			},
		},
	}
	safe := safeReviewSubmission(input, "2026-09-18T00:00:00Z")
	if safe["bibleReadingVersion"] != "5.19" || safe["bibleVersionId"] != "bible1" || safe["bibleVersionName"] != "Version autorizada" {
		t.Fatalf("Bible journal metadata was not preserved: %#v", safe)
	}
	if legacyMap(safe["bibleReadings"]) == nil {
		t.Fatalf("structured Bible readings were not preserved: %#v", safe)
	}
}
