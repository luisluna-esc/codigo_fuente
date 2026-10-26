package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// This path imports only progress/files for one existing account. It never
// replaces the database, sessions, roles, credentials or restore generation.
func recoveryHasWork554(record map[string]any) bool {
	status := stringValue(record["status"])
	if status != "" && status != "pending" && status != "assigned" {
		return true
	}
	submission := legacyMap(record["submissions"])
	for _, key := range []string{"text", "richText", "completionDate"} {
		if strings.TrimSpace(stringValue(submission[key])) != "" {
			return true
		}
	}
	if len(arrayField(submission, "attachmentIds")) > 0 || len(mapField(submission, "readingReport")) > 0 || len(mapField(submission, "devotion531")) > 0 || record["readings"] != nil {
		return true
	}
	return false
}
func (s *Server) recoveryCatalog554() (map[string]map[string]bool, error) {
	root := s.cfg.WebRoot
	if root == "" {
		root = filepath.Join(filepath.Dir(filepath.Dir(s.path)), "app")
	}
	out := map[string]map[string]bool{}
	for _, name := range []string{"requirements_master_v2.json", "program_catalog.json"} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return nil, fmt.Errorf("No se pudo comprobar el catálogo de requisitos actual. Actualiza el paquete completo de SGL.")
		}
		var catalog map[string]any
		if json.Unmarshal(data, &catalog) != nil {
			return nil, fmt.Errorf("El catálogo actual no es válido.")
		}
		programs := mapField(catalog, "programs")
		if name == "requirements_master_v2.json" {
			programs = map[string]any{"class_grouped": catalog}
		}
		for pid, raw := range programs {
			ids := map[string]bool{}
			for _, area := range arrayField(legacyMap(raw), "areas") {
				for _, req := range arrayField(legacyMap(area), "requirements") {
					ids[stringValue(legacyMap(req)["id"])] = true
				}
			}
			out[pid] = ids
		}
	}
	return out, nil
}
func (s *Server) prepareBackupRecovery554(database any, c recoveryCapsule548) (recoveryCapsule548, map[string]any, []string, error) {
	db := legacyMap(database)
	source := c.SourceUser
	fail := func(message string) (recoveryCapsule548, map[string]any, []string, error) {
		return c, nil, nil, fmt.Errorf("%s", message)
	}
	if stringValue(source["id"]) != c.OwnerID {
		return fail("La identidad de la copia no coincide con su expediente.")
	}
	norm := func(v any) string { return strings.ToLower(strings.TrimSpace(stringValue(v))) }
	target := legacyUser(db, c.OwnerID)
	if target == nil {
		for _, raw := range legacyUsers(db) {
			u := legacyMap(raw)
			match := false
			for _, k := range []string{"email", "username"} {
				if norm(source[k]) != "" && norm(source[k]) == norm(u[k]) {
					match = true
				}
			}
			if match {
				if target != nil {
					return fail("El correo o usuario corresponde a más de una cuenta. Corrige las cuentas antes de recuperar.")
				}
				target = u
			}
		}
	}
	if !leader555(db, target) {
		return fail("No se encontró la misma cuenta activa del aspirante en este servidor.")
	}
	// Never use a matching name alone or silently accept a conflicting identity.
	for _, k := range []string{"email", "username"} {
		if norm(source[k]) != "" && norm(target[k]) != "" && norm(source[k]) != norm(target[k]) {
			return fail("El correo o usuario de la copia no coincide con la cuenta actual.")
		}
	}
	catalog, err := s.recoveryCatalog554()
	if err != nil {
		return c, nil, nil, err
	}
	oldOwner := c.OwnerID
	c.OwnerID = stringValue(target["id"])
	allowed := map[string]bool{}
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if m["active"] != false && containsString(arrayField(m, "assignedUserIds"), c.OwnerID) {
			allowed[stringValue(m["id"])] = true
		}
	}
	if len(arrayField(db, "modules")) == 0 {
		allowed["class_grouped"] = true
	}
	skipped := []string{}
	programs := map[string]any{"class_grouped": c.Progress}
	for pid, v := range c.Programs {
		programs[pid] = v
	}
	clean := map[string]any{}
	for pid, raw := range programs {
		p := legacyMap(raw)
		next := map[string]any{}
		for _, field := range []string{"submissions", "status", "readings"} {
			values := map[string]any{}
			for id, v := range mapField(p, field) {
				if !allowed[pid] || !catalog[pid][id] {
					if recoveryHasWork554(recoveryRecord548(p, id)) {
						key := pid + " · " + id
						if !containsString(stringsToAny554(skipped), key) {
							skipped = append(skipped, key)
						}
					}
					continue
				}
				values[id] = cloneAny(v)
			}
			next[field] = values
		}
		if allowed[pid] {
			clean[pid] = next
		}
	}
	c.Progress = legacyMap(clean["class_grouped"])
	c.Programs = map[string]any{}
	for pid, v := range clean {
		if pid != "class_grouped" {
			c.Programs[pid] = v
		}
	}
	files := []any{}
	for _, raw := range c.Files {
		f := legacyMap(raw)
		if stringValue(f["ownerId"]) != oldOwner {
			return fail("La copia contiene archivos de otro usuario.")
		}
		pid := stringValue(f["programId"])
		if pid == "" {
			pid = "class_grouped"
		}
		rid := stringValue(f["requirementId"])
		if !allowed[pid] || !catalog[pid][rid] {
			continue
		}
		f = legacyMap(cloneAny(f))
		f["ownerId"] = c.OwnerID
		files = append(files, f)
	}
	c.Files = files
	if !allowed["class_grouped"] && !allowed["guide_major"] && !allowed["guide_master"] && !allowed["guide_master_advanced"] {
		c.Bible = nil
	}
	return c, target, skipped, nil
}
func stringsToAny554(in []string) []any {
	out := []any{}
	for _, v := range in {
		out = append(out, v)
	}
	return out
}

func recoveryStateFingerprint554(state legacyState, owner string) string {
	db := legacyMap(state.Database)
	user := legacyUser(db, owner)
	identity := map[string]any{}
	for _, key := range []string{"id", "email", "username", "role", "active", "accountState"} {
		identity[key] = user[key]
	}
	modules := []string{}
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if m["active"] != false && containsString(arrayField(m, "assignedUserIds"), owner) {
			modules = append(modules, stringValue(m["id"]))
		}
	}
	sort.Strings(modules)
	files := []any{}
	for _, raw := range state.Files {
		if stringValue(legacyMap(raw)["ownerId"]) == owner {
			files = append(files, raw)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return stringValue(legacyMap(files[i])["id"]) < stringValue(legacyMap(files[j])["id"])
	})
	data, _ := json.Marshal(map[string]any{"identity": identity, "modules": modules, "programs": recoveryPrograms548(db, owner), "bible": mapField(db, "bibleYear")[owner], "files": files, "generation": state.RestoreGeneration, "restoreId": state.RestoreID})
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
