package main

import (
	"bytes"
	"encoding/base64"
	"strings"
)

// Only a received PDF belonging to this annual requirement can enter review.
func musicAnnualPDF558(sub map[string]any, files []any, owner string) bool {
	if ok, _ := reviewSubmissionIsRequirementDocument(sub, files); !ok {
		return false
	}
	primary := strings.TrimSpace(stringValue(sub["primaryDocumentId"]))
	for _, raw := range files {
		f := legacyMap(raw)
		if stringValue(f["id"]) != primary {
			continue
		}
		if stringValue(f["ownerId"]) != owner || stringValue(f["programId"]) != musicProgram553 || stringValue(f["requirementId"]) != "LM-NUT-1" {
			return false
		}
		dataURL := stringValue(f["dataUrl"])
		if !strings.HasPrefix(dataURL, "data:application/pdf;base64,") {
			return false
		}
		data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURL, "data:application/pdf;base64,"))
		return err == nil && bytes.HasPrefix(data, []byte("%PDF-")) && reviewFilesComplete(sub, files, owner)
	}
	return false
}
