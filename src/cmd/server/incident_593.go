package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const incident593 = `{"removeClassIds":["usr_10dbea8e-4f34-43b7-829c-99ac532af145","usr_13fb2c69-11ed-4eb5-8729-30adf2469d32","usr_287e39b0-80d6-4cca-abba-9ee0c280a486","usr_2a24dd34-baf5-4629-931a-ae8649e0a15c","usr_313ebd46-d1cf-4a81-b09f-49576534d32f","usr_39f03584-e49f-4146-8608-17f710e5c112","usr_50401bc6-8a3a-42de-b2f0-cc85890564df","usr_5101836c-b84a-4856-8625-d635fbee42cd","usr_5881f4bd-85dd-4e53-bde9-5ffc3aecda18","usr_5b9fd5df-8d85-40ab-9bd3-0f2af37d19d1","usr_729b3bad-4397-4929-9476-3e8316a92d6f","usr_75e08376-9216-4aeb-825b-8581a9d161cd","usr_824cda26-5284-411f-9b70-8398b11bbdaa","usr_84c1f9d3-c888-4954-bb51-45baafce3350","usr_85094f66-e557-4328-a099-ea41b70d4359","usr_87d655d7-6b38-45f8-a61e-703ed183dcde","usr_88b0ed5f-caa5-4652-89d5-a96fc0564751","usr_8b7ae3e9-017a-4d80-8dd6-c2c381c040da","usr_8ec4c1a9-d91f-40df-a83b-db015c77a88f","usr_93354e5e-e307-471c-ac37-6d0389998029","usr_94385fbd-f6ce-4da1-af7d-a469f52bf8f8","usr_967cc2e1-b000-4f2a-a435-3dbbe628a16f","usr_9d5db3c5-4d1f-475b-a885-812f11d5ac19","usr_9d6adfc0-4524-4e13-b8b2-26eb0b3ca897","usr_a0c212e7-16d0-4092-968c-04c648b4dba0","usr_acb2891c-9744-4540-8cbb-d3c2dc27f5e5","usr_accc10cd-09c7-4714-93dc-103e0398c3c9","usr_b0857dbd-f464-46d3-8a79-a3e0c88f3a3d","usr_b1ec0a3b-d230-493e-bbe5-4a2ac154ebe4","usr_b1fe4b09-f814-4a14-9675-3455dbd1a8bb","usr_bbfe2461-3533-4502-a0c1-e9dcfca7d899","usr_bcb11ecb-f4ae-45fe-9588-882795be3fda","usr_bfe8882f-37b1-4def-b711-16eed36102db","usr_c3fa9d1c-a83b-4e46-a4bc-274aba2f3b66","usr_cbfc7091-b5fb-4818-8961-fe3c4433b84c","usr_cdbf99e8-53b8-431b-8bc4-fa9db7c94efd","usr_d25b7b93-cdd2-46fe-8589-26d8a232e66b","usr_e86fad35-804b-497a-8670-720f33ff870d","usr_f7bd383f-7e42-47b3-9647-57debdcc2e5d","usr_f80e1a58-e1c7-43f8-b192-a574e3c74a67"],"restoreMusicIds":["usr_10dbea8e-4f34-43b7-829c-99ac532af145","usr_13fb2c69-11ed-4eb5-8729-30adf2469d32","usr_2867d68f-5efa-4f4a-b851-3e2a021db15d","usr_287e39b0-80d6-4cca-abba-9ee0c280a486","usr_2a24dd34-baf5-4629-931a-ae8649e0a15c","usr_313ebd46-d1cf-4a81-b09f-49576534d32f","usr_39f03584-e49f-4146-8608-17f710e5c112","usr_4299c4d9-fc2e-40e5-92ec-e82e61722971","usr_4fcef000-bd25-46b9-9677-78c5f435943d","usr_50401bc6-8a3a-42de-b2f0-cc85890564df","usr_5101836c-b84a-4856-8625-d635fbee42cd","usr_538fc4d9-2791-4d5d-9189-e90bd3b8cbce","usr_5881f4bd-85dd-4e53-bde9-5ffc3aecda18","usr_5b9fd5df-8d85-40ab-9bd3-0f2af37d19d1","usr_729b3bad-4397-4929-9476-3e8316a92d6f","usr_75e08376-9216-4aeb-825b-8581a9d161cd","usr_824cda26-5284-411f-9b70-8398b11bbdaa","usr_84c1f9d3-c888-4954-bb51-45baafce3350","usr_85094f66-e557-4328-a099-ea41b70d4359","usr_87d655d7-6b38-45f8-a61e-703ed183dcde","usr_88b0ed5f-caa5-4652-89d5-a96fc0564751","usr_8b7ae3e9-017a-4d80-8dd6-c2c381c040da","usr_8ec4c1a9-d91f-40df-a83b-db015c77a88f","usr_9218c4d5-bf29-4c38-b5ee-3e4ac1f0abc0","usr_93354e5e-e307-471c-ac37-6d0389998029","usr_94385fbd-f6ce-4da1-af7d-a469f52bf8f8","usr_967cc2e1-b000-4f2a-a435-3dbbe628a16f","usr_9a3e7a8e-eb5d-4133-acbe-18f3aa98297f","usr_9d5db3c5-4d1f-475b-a885-812f11d5ac19","usr_9d6adfc0-4524-4e13-b8b2-26eb0b3ca897","usr_a0c212e7-16d0-4092-968c-04c648b4dba0","usr_acb2891c-9744-4540-8cbb-d3c2dc27f5e5","usr_accc10cd-09c7-4714-93dc-103e0398c3c9","usr_b0857dbd-f464-46d3-8a79-a3e0c88f3a3d","usr_b1ec0a3b-d230-493e-bbe5-4a2ac154ebe4","usr_b1fe4b09-f814-4a14-9675-3455dbd1a8bb","usr_bbfe2461-3533-4502-a0c1-e9dcfca7d899","usr_bcb11ecb-f4ae-45fe-9588-882795be3fda","usr_bf273961-4fce-4b74-b5b4-a29381235ce4","usr_bfe8882f-37b1-4def-b711-16eed36102db","usr_c3fa9d1c-a83b-4e46-a4bc-274aba2f3b66","usr_cbfc7091-b5fb-4818-8961-fe3c4433b84c","usr_cdbf99e8-53b8-431b-8bc4-fa9db7c94efd","usr_d25b7b93-cdd2-46fe-8589-26d8a232e66b","usr_e86fad35-804b-497a-8670-720f33ff870d","usr_eb925923-add7-434e-bd3b-f0260d729c89","usr_f5a95bbf-f73b-4f07-b645-0c1a445fdcc5","usr_f710ed59-70f5-453f-ab1c-d01463b8b7be","usr_f7bd383f-7e42-47b3-9647-57debdcc2e5d","usr_f80e1a58-e1c7-43f8-b192-a574e3c74a67"]}`

// Repair only the exact incident documented by the supplied backups.
// Never overwrite progress, people, tutors, files or other module enrollments.
func repairDatabase593(db map[string]any) bool {
	if db == nil || db["incidentMusic593"] == true {
		return false
	}
	var plan struct {
		RemoveClassIds  []string
		RestoreMusicIds []string
	}
	if json.Unmarshal([]byte(incident593), &plan) != nil {
		return false
	}
	for _, id := range plan.RemoveClassIds {
		u := legacyUser(db, id)
		if u == nil || stringValue(legacyMap(u["tutorByProgram"])[musicProgram553]) == "" {
			return false
		}
	}
	var class, music map[string]any
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if m["id"] == "class_grouped" {
			class = m
		}
		if m["id"] == musicProgram553 {
			music = m
		}
	}
	if class == nil || music == nil {
		return false
	}
	remove := map[string]bool{}
	for _, id := range plan.RemoveClassIds {
		remove[id] = true
	}
	kept := []any{}
	for _, id := range arrayField(class, "assignedUserIds") {
		if !remove[stringValue(id)] {
			kept = append(kept, id)
		}
	}
	class["assignedUserIds"] = kept
	ids := append([]any{}, arrayField(music, "assignedUserIds")...)
	for _, id := range plan.RestoreMusicIds {
		u := legacyUser(db, id)
		if u != nil && legacyActive(u) && u["musicAccessRemoved"] != true && !containsString(ids, id) {
			ids = append(ids, id)
		}
	}
	music["assignedUserIds"] = ids
	db["incidentMusic593"] = true
	return true
}

func (s *Server) repairIncident593() error {
	st, err := s.loadLegacy()
	if err != nil {
		return err
	}
	candidate := legacyMap(cloneAny(st.State.Database))
	if !repairDatabase593(candidate) {
		return nil
	}
	// A durable full backup is required before persisting any repair.
	raw, err := json.Marshal(st)
	if err != nil {
		return err
	}
	dir := filepath.Join(filepath.Dir(s.path), "Backups")
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	name := filepath.Join(dir, "antes-reparacion-5.9.3-"+time.Now().UTC().Format("20060102T150405.000000000")+".json")
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = f.Write(raw)
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	st.State.Database = candidate
	st.State.ServerVersion++
	st.State.RestoreGeneration++
	st.State.RestoreID = "repair593_" + token(16)
	st.State.RestoredAt = time.Now().UTC().Format(time.RFC3339)
	st.State.UpdatedAt = st.State.RestoredAt
	if err = s.saveLegacy(st); err != nil {
		return fmt.Errorf("guardar recuperación (respaldo %s): %w", name, err)
	}
	return nil
}
