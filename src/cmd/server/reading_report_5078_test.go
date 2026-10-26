package main

import "testing"

func TestReadingReportMode5078(t *testing.T) {
	db := map[string]any{
		"reportRequirementModes": map[string]any{
			"guide_major:REQ-01":   map[string]any{"mode": "reading_report"},
			"guide_master:REQ-02":  "reading_report",
			"class_grouped:REQ-03": map[string]any{"mode": "reading_report"},
		},
	}
	if !reviewRequirementReadingReportMode(db, "guide_major", "REQ-01") {
		t.Fatal("Guia Mayor must accept explicitly configured reading report mode")
	}
	if !reviewRequirementReadingReportMode(db, "guide_master", "REQ-02") {
		t.Fatal("Guia Mayor Master must accept explicitly configured reading report mode")
	}
	if !reviewRequirementReadingReportMode(db, "class_grouped", "REQ-03") {
		t.Fatal("class_grouped must accept explicitly configured reading report mode in 5.0.99")
	}
	if reviewRequirementDocumentMode(db, "guide_major", "REQ-01") {
		t.Fatal("reading report and document modes must be mutually exclusive")
	}
}

func TestReadingReportValidation5078(t *testing.T) {
	submission := map[string]any{
		"submissionMode": "reading_report",
		"deliveryMode":   "reading_report",
		"attachmentIds":  []any{},
		"readingReport": map[string]any{
			"elements":        []any{"Uno", "Dos", "Tres", "", ""},
			"impactPhrase":    "Frase significativa",
			"impactPage":      "42",
			"personal":        "Aplicacion personal",
			"professional":    "Aplicacion profesional",
			"spiritual":       "Aplicacion espiritual",
			"finalReflection": "Reflexion final",
		},
	}
	if ok, message := reviewSubmissionIsReadingReportComplete(submission); !ok {
		t.Fatalf("complete reading report must pass: %s", message)
	}
	if ok, message := reviewSubmissionIsInternalReport(submission, nil); !ok {
		t.Fatalf("reading report must be recognized by the review pipeline: %s", message)
	}

	report := legacyMap(submission["readingReport"])
	report["elements"] = []any{"Uno", "Dos", "", "", ""}
	if ok, _ := reviewSubmissionIsReadingReportComplete(submission); ok {
		t.Fatal("fewer than three key elements must be rejected")
	}
	report["elements"] = []any{"Uno", "Dos", "Tres", "", ""}
	submission["attachmentIds"] = []any{"photo-1"}
	if ok, _ := reviewSubmissionIsReadingReportComplete(submission); ok {
		t.Fatal("reading reports must reject attachments")
	}
}

func TestSafeReadingReportSubmission5078(t *testing.T) {
	submission := map[string]any{
		"submissionMode":       "reading_report",
		"deliveryMode":         "reading_report",
		"readingReportVersion": 1,
		"readingReport": map[string]any{
			"elements":        []any{"A", "B", "C", "", ""},
			"impactPhrase":    "Frase",
			"impactPage":      "7",
			"personal":        "P",
			"professional":    "Pr",
			"spiritual":       "E",
			"finalReflection": "R",
		},
	}
	safe := safeReviewSubmission(submission, "2026-09-14T00:00:00Z")
	if !reviewSubmissionIsReadingReport(safe) {
		t.Fatal("safe review copy lost reading_report mode")
	}
	if legacyMap(safe["readingReport"]) == nil {
		t.Fatal("safe review copy lost reading report form data")
	}
	if int(numericValue(safe["readingReportVersion"])) != 1 {
		t.Fatal("safe review copy lost reading report version")
	}
}
