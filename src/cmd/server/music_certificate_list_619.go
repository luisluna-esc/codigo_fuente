package main

// Do not send template artwork, fonts or document bytes when listing issued
// certificates. Keep scope checks identical for current and historical rows.
func musicCertificateList619(db, actor map[string]any) []any {
	d := musicData550(db)
	rows := []any{}
	seen := map[string]bool{}
	add := func(raw any) {
		c := legacyMap(raw)
		ownerID := stringValue(c["ownerId"])
		if ownerID != stringValue(actor["id"]) && !musicLeaders610(legacyRole(actor)) && !musicScope550(db, actor, legacyUser(db, ownerID)) && !(ownerID == "" && c["ministryId"] != nil) {
			return
		}
		identity := stringValue(c["token"])
		if identity == "" {
			identity = stringValue(c["id"])
		}
		if seen[identity] {
			return
		}
		seen[identity] = true
		copy := map[string]any{}
		for k, v := range c {
			if k != "templateSnapshot610" {
				copy[k] = v
			}
		}
		copy = cloneMap610(copy)
		copy["canDelete618"] = musicCanDeleteCertificate618(db, actor, c)
		rows = append(rows, copy)
	}
	for _, raw := range mapField(d, "certificates") {
		add(raw)
	}
	for _, raw := range arrayField(d, "certificateHistory610") {
		add(raw)
	}
	return rows
}
