package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
)

const musicAPI553 = "/api/musica"
const musicProgram553 = "music_ministry"

type musicContext553 struct{}

func musicRequest553(r *http.Request) bool { return r.Context().Value(musicContext553{}) == true }

func musicAllowed553(db map[string]any, actor map[string]any) bool {
	if actor == nil || !legacyActive(actor) || actor["musicAccessRemoved"] == true {
		return false
	}
	role := legacyRole(actor)
	if musicPolicies560[role].Notices && managementProgram555(db, actor, musicProgram553) {
		return true
	}
	if musicAssigned550(db, stringValue(actor["id"])) {
		return true
	}
	if role == "aspirant" {
		return musicAssigned550(db, stringValue(actor["id"]))
	}
	if role == "tutor" {
		return containsString(tutorPrograms548(db, actor), musicProgram553)
	}
	if role == "regional" {
		for _, raw := range legacyUsers(db) {
			if musicScope550(db, actor, legacyMap(raw)) {
				return true
			}
		}
	}
	return false
}

func musicKey553(key string) bool {
	return strings.HasPrefix(key, musicProgram553+":") || strings.Contains(key, "LM-")
}
func musicItem553(raw any) bool {
	m := legacyMap(raw)
	return stringValue(m["programId"]) == musicProgram553 || (stringValue(m["programId"]) == "" && strings.HasPrefix(stringValue(m["requirementId"]), "LM-"))
}

// Project a copy: the SGL canonical state and its other programs stay intact.
func musicProject553(source any, actorID, role string) map[string]any {
	db := legacyMap(cloneAny(source))
	if db == nil {
		return nil
	}
	actor := legacyUser(db, actorID)
	if role == "aspirant" && legacyRole(actor) != "aspirant" {
		actor = legacyMap(cloneAny(actor))
		actor["role"] = "aspirant"
	}
	users := []any{}
	owners := map[string]bool{}
	visible := map[string]bool{}
	learner := role == "aspirant" || role == "pastor_aspirant"
	directory := map[string]map[string]any{}
	for _, raw := range arrayField(db, "musicTutorDirectory560") {
		u := legacyMap(raw)
		directory[stringValue(u["id"])] = u
	}
	for _, raw := range legacyUsers(db) {
		u := legacyMap(raw)
		id := stringValue(u["id"])
		owner := musicAssigned550(db, id) && (actorID == "" || musicScope550(db, actor, u))
		keep := id == actorID || owner || musicSeeTutor560(db, actor, u) || directory[id] != nil
		if musicPolicies560[role].Profile && (actorID == "" || managementProgram555(db, actor, musicProgram553)) && musicAllowed553(db, u) {
			keep = true
		}
		if learner {
			keep = id == actorID || legacyTutorAssigned(actor, id, musicProgram553)
		}
		if !keep {
			continue
		}
		visible[id] = true
		if owner {
			owners[id] = true
		}
		if !learner && !owner && id != actorID && !musicManager550(role) {
			users = append(users, map[string]any{"id": id, "fullName": u["fullName"], "role": u["role"], "region": u["region"], "createdBy": u["createdBy"], "active": legacyActive(u), "accountState": "active", "musicIsTutor560": true})
			continue
		}
		if learner && id != actorID {
			users = append(users, map[string]any{"id": id, "fullName": u["fullName"], "role": u["role"], "active": legacyActive(u), "accountState": "active"})
			continue
		}
		for _, k := range []string{"tutorByProgram", "tutorPrograms", "programTutors"} {
			if m := legacyMap(u[k]); m != nil {
				out := map[string]any{}
				if v, ok := m[musicProgram553]; ok {
					out[musicProgram553] = v
				}
				u[k] = out
			}
		}
		for _, k := range []string{"tutorProgramIds", "programIds", "assignedProgramIds", "requirementViewPrograms", "regionalProgramIds", "managementProgramIds"} {
			if a, ok := u[k]; ok {
				list := []any{}
				if containsString(anySlice548(a), musicProgram553) {
					list = append(list, musicProgram553)
				}
				u[k] = list
			}
		}
		users = append(users, u)
	}
	if !learner {
		for id, u := range directory {
			if !visible[id] {
				users = append(users, u)
				visible[id] = true
			}
		}
	}
	delete(db, "musicEmptyArchive597")
	db["users"] = users
	db["clubs"] = []any{}
	db["classes"] = []any{}
	db["progress"] = map[string]any{}
	db["bibleYear"] = map[string]any{}
	db["errors"] = []any{}
	modules := []any{}
	for _, raw := range arrayField(db, "modules") {
		m := legacyMap(raw)
		if stringValue(m["id"]) != musicProgram553 {
			continue
		}
		ids := []any{}
		for _, id := range arrayField(m, "assignedUserIds") {
			if owners[stringValue(id)] {
				ids = append(ids, id)
			}
		}
		m["assignedUserIds"] = ids
		modules = append(modules, m)
	}
	db["modules"] = modules
	progress := map[string]any{}
	for id, raw := range legacyMap(db["programProgress"]) {
		if !owners[id] {
			continue
		}
		if p, ok := legacyMap(raw)[musicProgram553]; ok {
			progress[id] = map[string]any{musicProgram553: p}
		}
	}
	db["programProgress"] = progress
	for _, key := range []string{"configs", "reportRequirementModes"} {
		out := map[string]any{}
		for k, v := range legacyMap(db[key]) {
			if musicKey553(k) {
				out[k] = v
			}
		}
		db[key] = out
	}
	for _, key := range []string{"tasks", "materials", "indicationRules"} {
		out := []any{}
		for _, raw := range arrayField(db, key) {
			m := legacyMap(raw)
			if musicItem553(raw) && (key != "tasks" || visible[stringValue(m["assigneeId"])]) {
				out = append(out, raw)
			}
		}
		db[key] = out
	}
	if settings := legacyMap(db["settings"]); settings != nil {
		for _, k := range []string{"serverToken", "serverUrl", "nameAutocompleteHistory"} {
			delete(settings, k)
		}
		if m := legacyMap(settings["reportRequirementModes"]); m != nil {
			out := map[string]any{}
			for k, v := range m {
				if musicKey553(k) {
					out[k] = v
				}
			}
			settings["reportRequirementModes"] = out
		}
	}
	audit := []any{}
	for _, raw := range arrayField(db, "audit") {
		m := legacyMap(raw)
		if musicKey553(stringValue(m["entity"])) || stringValue(m["programId"]) == musicProgram553 {
			audit = append(audit, raw)
		}
	}
	db["audit"] = audit
	if actor != nil && db["music550"] != nil {
		db["music550"] = musicView550(db, actor)
	}
	return db
}

func musicFileIDs553(value any, ids map[string]bool) {
	switch v := value.(type) {
	case map[string]any:
		for k, x := range v {
			if removedMusicIntegrationKey(k) {
				continue
			}
			if (k == "fontAsset613" || k == "fileId610" || k == "fileId" || k == "photoFileId" || k == "primaryDocumentId" || k == "documentFileId" || k == "watermarkFileId") && stringValue(x) != "" {
				ids[stringValue(x)] = true
			}
			if k == "attachmentIds" || k == "fileIds" {
				for _, id := range anySlice548(x) {
					ids[stringValue(id)] = true
				}
			}
			musicFileIDs553(x, ids)
		}
	case []any:
		for _, x := range v {
			musicFileIDs553(x, ids)
		}
	}
}

// Merge the Music view into central data; absence in this view is never a
// deletion of an SGL account or another program.
func musicMerge553(central, incoming any) any {
	out := legacyMap(cloneAny(central))
	view := legacyMap(incoming)
	if out == nil || view == nil {
		return central
	}
	for _, key := range []string{"modules", "tasks", "materials", "indicationRules"} {
		list := []any{}
		for _, raw := range arrayField(out, key) {
			match := musicItem553(raw)
			if key == "modules" {
				match = stringValue(legacyMap(raw)["id"]) == musicProgram553
			}
			if !match {
				list = append(list, raw)
			}
		}
		for _, raw := range arrayField(view, key) {
			match := musicItem553(raw)
			if key == "modules" {
				match = stringValue(legacyMap(raw)["id"]) == musicProgram553
			}
			if match {
				if key == "modules" {
					// A projected snapshot is not evidence that omitted people
					// were removed. Explicit membership deltas handle removals.
					item := legacyMap(cloneAny(raw))
					ids := append([]any{}, arrayField(item, "assignedUserIds")...)
					for _, oldRaw := range arrayField(out, "modules") {
						old := legacyMap(oldRaw)
						if old["id"] == musicProgram553 {
							for _, id := range arrayField(old, "assignedUserIds") {
								if !containsString(ids, stringValue(id)) {
									ids = append(ids, id)
								}
							}
						}
					}
					item["assignedUserIds"] = ids
					raw = item
				}
				list = append(list, raw)
			}
		}
		if key == "modules" {
			found := false
			for _, raw := range list {
				if legacyMap(raw)["id"] == musicProgram553 {
					found = true
				}
			}
			if !found {
				for _, raw := range arrayField(out, key) {
					if legacyMap(raw)["id"] == musicProgram553 {
						list = append(list, raw)
					}
				}
			}
		}
		if _, ok := view[key]; ok {
			out[key] = list
		}
	}
	for _, key := range []string{"configs", "reportRequirementModes"} {
		if v, ok := view[key]; ok {
			m := ensureMapField(out, key)
			for k := range m {
				if musicKey553(k) {
					delete(m, k)
				}
			}
			for k, x := range legacyMap(v) {
				if musicKey553(k) {
					m[k] = x
				}
			}
		}
	}
	if v, ok := view["programProgress"]; ok {
		m := ensureMapField(out, "programProgress")
		for id, raw := range legacyMap(v) {
			target := ensureMapField(m, id)
			if p, ok := legacyMap(raw)[musicProgram553]; ok {
				target[musicProgram553] = p
			}
		}
	}
	users := legacyUsers(out)
	for _, raw := range arrayField(view, "users") {
		u := legacyMap(raw)
		id := stringValue(u["id"])
		old := legacyUser(out, id)
		if old == nil {
			if legacyRole(u) != "aspirant" && legacyRole(u) != "tutor" && !musicEventManager551(legacyRole(u)) {
				continue
			}
			users = append(users, cloneAny(u))
			continue
		}
		if !musicAllowed553(legacyMap(central), old) {
			continue
		}
		for k, v := range u {
			if k == "clubId" || k == "classTeaching" {
				continue
			}
			if k == "tutorByProgram" || k == "tutorPrograms" || k == "programTutors" {
				m := ensureMapField(old, k)
				delete(m, musicProgram553)
				if x, ok := legacyMap(v)[musicProgram553]; ok {
					m[musicProgram553] = x
				}
				continue
			}
			if k == "tutorProgramIds" || k == "programIds" || k == "assignedProgramIds" || k == "requirementViewPrograms" || k == "regionalProgramIds" || k == "managementProgramIds" {
				a := []any{}
				prior := anySlice548(old[k])
				if old[k] == nil {
					prior = implicitProgramList559(legacyMap(central), old, k)
				}
				for _, x := range prior {
					if stringValue(x) != musicProgram553 {
						a = append(a, x)
					}
				}
				if containsString(anySlice548(v), musicProgram553) {
					a = append(a, musicProgram553)
				}
				old[k] = a
				continue
			}
			old[k] = v
		}
	}
	out["users"] = users
	if settings := legacyMap(view["settings"]); settings != nil {
		target := ensureMapField(out, "settings")
		for k, v := range settings {
			if k != "serverUrl" && k != "serverToken" && k != "remoteAccessUrl" && k != "remoteAccessActive" && k != "appLogo" && k != "adminLogo" && k != "nameAutocompleteHistory" {
				target[k] = v
			}
		}
	}
	return out
}

func musicValidatePatch553(patch any, db map[string]any) bool {
	p := legacyMap(patch)
	if p == nil {
		return true
	}
	for k, raw := range p {
		switch k {
		case "classes", "clubs", "progress", "bibleYear":
			return false
		case "modules":
			values := anySlice548(raw)
			if m := legacyMap(raw); m != nil {
				values = anySlice548(m["__sglReplace"])
			}
			for _, v := range values {
				if stringValue(legacyMap(v)["id"]) != musicProgram553 {
					return false
				}
			}
		case "programProgress":
			for id, v := range legacyMap(raw) {
				if id == "__sglDeleteKeys" || id == "__sglReplace" {
					return false
				}
				if !musicAssigned550(db, id) && !musicAssigned550(legacyMap(expandLegacyPatch(db, patch)), id) {
					return false
				}
				for program := range legacyMap(v) {
					if program != musicProgram553 {
						return false
					}
				}
			}
		case "tasks", "materials", "indicationRules":
			values := anySlice548(raw)
			if m := legacyMap(raw); m != nil {
				values = anySlice548(m["__sglReplace"])
			}
			for _, v := range values {
				if !musicItem553(v) {
					return false
				}
			}
		case "configs", "reportRequirementModes":
			values := legacyMap(raw)
			if replacement := legacyMap(values["__sglReplace"]); replacement != nil {
				values = replacement
			}
			for key := range values {
				if key == "__sglDeleteKeys" {
					for _, id := range anySlice548(values[key]) {
						if !musicKey553(stringValue(id)) {
							return false
						}
					}
					continue
				}
				if !musicKey553(key) {
					return false
				}
			}
		case "users":
			values := anySlice548(raw)
			if m := legacyMap(raw); m != nil {
				values = anySlice548(m["__sglReplace"])
			}
			for _, v := range values {
				u := legacyMap(v)
				for _, name := range anySlice548(u["__sglDeleteKeys"]) {
					if strings.Contains(stringValue(name), "Program") || stringValue(name) == "clubId" || stringValue(name) == "classTeaching" {
						return false
					}
				}
				old := legacyUser(db, stringValue(u["id"]))
				if old == nil && (strings.TrimSpace(stringValue(u["fullName"])) == "" || stringValue(u["role"]) == "") {
					return false
				}
				if old != nil && !musicAllowed553(db, old) {
					return false
				}
				for _, key := range []string{"tutorByProgram", "tutorPrograms", "programTutors"} {
					values := legacyMap(u[key])
					if replacement := legacyMap(values["__sglReplace"]); replacement != nil {
						values = replacement
					}
					for program := range values {
						if program == "__sglDeleteKeys" {
							for _, id := range anySlice548(values[program]) {
								if stringValue(id) != musicProgram553 {
									return false
								}
							}
							continue
						}
						if program != musicProgram553 {
							return false
						}
					}
				}
			}
		case "__sglDeleteKeys":
			return false
		}
	}
	return true
}

func (s *Server) musicPortal553(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix := strings.HasPrefix(r.URL.Path, musicAPI553+"/")
		sess, hasSession := s.remoteSessionFromRequest(r)
		if !prefix && (!hasSession || sess.Product != "music") {
			if hasSession && strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/api/remote/logout" {
				s.mu.Lock()
				state, err := s.loadLegacyLoginState()
				s.mu.Unlock()
				if err != nil {
					legacyError(w, 503, "No se pudieron comprobar los permisos.", nil)
					return
				}
				if !portalAllowed593(legacyMap(state.Database), legacyUser(state.Database, sess.UserID), false) {
					legacyError(w, 403, "Tu cuenta no tiene acceso a Conquistadores. Entra por el enlace de Música.", nil)
					return
				}
			}
			next.ServeHTTP(w, r)
			return
		}
		if !prefix {
			legacyError(w, 403, "Esta sesión corresponde a Música. Entra por el enlace de Música.", nil)
			return
		}
		request := r.Clone(context.WithValue(r.Context(), musicContext553{}, true))
		request.URL.Path = strings.TrimPrefix(r.URL.Path, musicAPI553)
		request.Header = request.Header.Clone()
		request.Header.Del("Accept-Encoding")
		path := request.URL.Path
		if p := request.URL.Query().Get("programId"); p != "" && p != musicProgram553 {
			legacyError(w, 403, "El programa solicitado no pertenece a Música.", nil)
			return
		}
		if id := request.URL.Query().Get("requirementId"); id != "" && !strings.HasPrefix(id, "LM-") {
			legacyError(w, 403, "El requisito solicitado no pertenece a Música.", nil)
			return
		}
		public := path == "/api/remote/login" || path == "/api/remote/info" || path == "/api/health" || path == "/api/music/certificate"
		allowed := map[string]bool{"/api/music/v610": true, "/api/accounts/reconcile598": true, "/api/access571": true, "/api/accounts/directory": true, "/api/accounts/unassigned": true, "/api/accounts/import": true, "/api/import/xlsx": true, "/api/export/xlsx": true, "/api/accounts/view": true, "/api/accounts/music": true, "/api/accounts/roster": true, "/api/accounts/tutor": true, "/api/health": true, "/api/remote/info": true, "/api/remote/login": true, "/api/remote/session": true, "/api/remote/bootstrap": true, "/api/remote/logout": true, "/api/remote/heartbeat": true, "/api/remote/files": true, "/api/remote/presence": true, "/api/admin/control": true, "/api/bootstrap": true, "/api/sync/diagnostics": true, "/api/sync": true, "/api/admin/changes": true, "/api/report-config": true, "/api/indications": true, "/api/messages": true, "/api/messages/send": true, "/api/messages/read": true, "/api/notifications": true, "/api/notifications/read": true, "/api/notifications/delete": true, "/api/activity/daily": true, "/api/activity/track": true, "/api/reviews/submit": true, "/api/reviews/item": true, "/api/reviews/own-changes": true, "/api/reviews/decision": true, "/api/progress/reset": true, "/api/proposals": true, "/api/proposals/status": true, "/api/proposals/decision": true, "/api/proposals/cancel": true, "/api/user-changes": true, "/api/user-changes/decision": true, "/api/music/module": true, "/api/music/contact": true, "/api/music/certificate": true, "/api/music-registration": true, "/api/report-watermark": true, "/api/birthdays": true, "/api/music-data/import": true, "/api/music-data/export": true, "/api/music-data/photos": true}
		if !allowed[path] {
			legacyError(w, 403, "Esta operación no está disponible en Música.", nil)
			return
		}
		if hasSession && sess.Product != "music" {
			legacyError(w, 401, "Inicia sesión desde la entrada de Música.", nil)
			return
		}
		// Login and public metadata have no snapshot to project. Avoid cloning the
		// database (and loading evidence bytes) on the critical authentication path.
		if public && (path == "/api/health" || path == "/api/remote/info") {
			next.ServeHTTP(w, request)
			return
		}
		if path == "/api/remote/login" && request.Method == http.MethodPost {
			body, e := io.ReadAll(io.LimitReader(request.Body, (1<<20)+1))
			if e != nil || len(body) > 1<<20 {
				legacyError(w, 400, "Solicitud inválida.", nil)
				return
			}
			request.Body = io.NopCloser(bytes.NewReader(body))
			request.ContentLength = int64(len(body))
			var login map[string]any
			if json.Unmarshal(body, &login) == nil && login["authOnly"] == true {
				next.ServeHTTP(w, request)
				return
			}
		}
		if path == "/api/accounts/view" && hasSession {
			request.Header.Set("Accept-Encoding", r.Header.Get("Accept-Encoding"))
			next.ServeHTTP(w, request)
			return
		}
		actorID, role := sess.UserID, sess.Role
		s.mu.Lock()
		state, err := s.loadLegacy()
		state.State.Database = cloneAny(state.State.Database)
		s.mu.Unlock()
		db := legacyMap(state.State.Database)
		if err != nil {
			legacyError(w, 500, "No se pudieron consultar los datos del servidor.", nil)
			return
		}
		if !public {
			if path == "/api/proposals" && request.URL.Query().Get("id") != "" {
				_, proposal := findCanonicalProposal(&state, request.URL.Query().Get("id"))
				if proposal != nil && proposal.ProgramID != musicProgram553 && !strings.HasPrefix(proposal.RequirementID, "LM-") {
					legacyError(w, 403, "Esta revisión heredada se administra desde SGL.", nil)
					return
				}
			}
			if !hasSession {
				scope, ok := s.auth(request)
				if !ok || scope != "admin" {
					legacyError(w, 401, "Inicia sesión en Música.", nil)
					return
				}
				role = "admin"
			}
			if hasSession && !musicAllowed553(db, legacyUser(db, actorID)) {
				legacyError(w, 403, "Tu cuenta no tiene acceso activo a Música.", nil)
				return
			}
			for _, key := range []string{"ownerId", "userId", "assigneeId", "peerId", "toUserId"} {
				id := request.URL.Query().Get(key)
				if id != "" && !musicAllowed553(db, legacyUser(db, id)) {
					legacyError(w, 403, "La cuenta solicitada no pertenece a Música.", nil)
					return
				}
			}
			if path == "/api/remote/files" {
				if id := request.URL.Query().Get("fileId"); id != "" {
					ids := map[string]bool{}
					musicFileIDs553(musicProject553(db, actorID, role), ids)
					if !ids[id] {
						legacyError(w, 403, "El archivo no está asignado a Música en tu alcance.", nil)
						return
					}
				}
			}
		}
		if request.Method == http.MethodPost && !public && path != "/api/remote/logout" && path != "/api/remote/heartbeat" {
			body, readErr := io.ReadAll(io.LimitReader(request.Body, (256<<20)+1))
			if readErr != nil || len(body) > 256<<20 {
				legacyError(w, 400, "Solicitud inválida.", nil)
				return
			}
			var payload map[string]any
			if len(body) > 0 && json.Unmarshal(body, &payload) == nil {
				if id := stringValue(payload["proposalId"]); id != "" {
					_, proposal := findCanonicalProposal(&state, id)
					if proposal != nil && proposal.ProgramID != musicProgram553 && !strings.HasPrefix(proposal.RequirementID, "LM-") {
						legacyError(w, 403, "Esta revisión heredada se administra desde SGL.", nil)
						return
					}
				}
				for _, id := range anySlice548(payload["recipientIds"]) {
					if !musicAllowed553(db, legacyUser(db, stringValue(id))) {
						legacyError(w, 403, "El destinatario no pertenece a Música.", nil)
						return
					}
				}
				for _, id := range anySlice548(payload["requestIds"]) {
					for _, change := range state.UserChanges {
						if change.ID == stringValue(id) && (!musicAllowed553(db, legacyUser(db, change.OwnerID)) || !musicValidatePatch553(change.Patch, db)) {
							legacyError(w, 403, "El cambio solicitado se administra desde SGL.", nil)
							return
						}
					}
				}
				for _, key := range []string{"ownerId", "userId", "assigneeId", "peerId", "toUserId", "recipientId"} {
					id := stringValue(payload[key])
					if id != "" && !musicAllowed553(db, legacyUser(db, id)) {
						legacyError(w, 403, "La cuenta solicitada no pertenece a Música.", nil)
						return
					}
				}
				if p := stringValue(payload["programId"]); p != "" && p != musicProgram553 {
					legacyError(w, 403, "Solo se permiten cambios del programa de Música.", nil)
					return
				}
				if id := stringValue(payload["requirementId"]); id != "" && !strings.HasPrefix(id, "LM-") {
					legacyError(w, 403, "El requisito no pertenece a Música.", nil)
					return
				}
				if !musicValidatePatch553(payload["patch"], db) {
					legacyError(w, 403, "El cambio incluye datos ajenos a Música.", nil)
					return
				}
				if patch := legacyMap(payload["patch"]); patch != nil {
					// Collection replacement must retain central rows excluded from the Music view.
					for _, key := range []string{"users", "modules", "tasks", "materials", "indicationRules"} {
						m := legacyMap(patch[key])
						if m == nil || m["__sglReplace"] == nil {
							continue
						}
						merged := legacyMap(musicMerge553(db, map[string]any{key: m["__sglReplace"]}))
						m["__sglReplace"] = merged[key]
					}
					if users := anySlice548(patch["users"]); users != nil {
						for _, raw := range users {
							u := legacyMap(raw)
							for _, key := range []string{"tutorByProgram", "tutorPrograms", "programTutors"} {
								m := legacyMap(u[key])
								if replacement := legacyMap(m["__sglReplace"]); replacement != nil {
									// Apply only the Music key under the canonical write lock.
									// A projected replacement must not freeze other programs.
									delta := map[string]any{"__sglDeleteKeys": []any{musicProgram553}}
									if v, ok := replacement[musicProgram553]; ok {
										delta = map[string]any{musicProgram553: v}
									}
									u[key] = delta
								}
							}
							for _, key := range []string{"tutorProgramIds", "programIds", "assignedProgramIds", "requirementViewPrograms", "regionalProgramIds", "managementProgramIds"} {
								if value, ok := u[key]; ok {
									add, remove := []any{}, []any{}
									if marker := legacyMap(legacyMap(value)["__sglSetMembers"]); marker != nil {
										if containsString(anySlice548(marker["add"]), musicProgram553) {
											add = append(add, musicProgram553)
										}
										if containsString(anySlice548(marker["remove"]), musicProgram553) {
											remove = append(remove, musicProgram553)
										}
									} else if containsString(anySlice548(value), musicProgram553) {
										add = append(add, musicProgram553)
									} else {
										remove = append(remove, musicProgram553)
									}
									u[key] = map[string]any{"__sglSetMembers": map[string]any{"add": add, "remove": remove}}
								}
							}
						}
					}
				}
				uploadedFiles := anySlice548(payload["files"])
				if incoming := legacyMap(payload["database"]); incoming != nil {
					payload["database"] = musicMerge553(db, incoming)
					files := []any{}
					byID := map[string]bool{}
					for _, raw := range state.State.Files {
						files = append(files, raw)
						byID[stringValue(legacyMap(raw)["id"])] = true
					}
					for _, raw := range anySlice548(payload["files"]) {
						id := stringValue(legacyMap(raw)["id"])
						if !byID[id] {
							files = append(files, raw)
						} else {
							for i, old := range files {
								if stringValue(legacyMap(old)["id"]) == id {
									files[i] = raw
									break
								}
							}
						}
					}
					payload["files"] = files
				}
				if uploaded := uploadedFiles; len(uploaded) > 0 {
					currentIDs, candidateIDs := map[string]bool{}, map[string]bool{}
					musicFileIDs553(musicProject553(db, actorID, role), currentIDs)
					candidate := payload["database"]
					if candidate == nil {
						candidate = expandLegacyPatch(db, payload["patch"])
					}
					musicFileIDs553(musicProject553(candidate, actorID, role), candidateIDs)
					centralIDs := map[string]bool{}
					for _, raw := range state.State.Files {
						centralIDs[stringValue(legacyMap(raw)["id"])] = true
					}
					for _, raw := range uploaded {
						id := stringValue(legacyMap(raw)["id"])
						if id == "" || centralIDs[id] && !currentIDs[id] || !centralIDs[id] && !candidateIDs[id] {
							legacyError(w, 403, "El archivo enviado no pertenece a Música en tu alcance.", nil)
							return
						}
					}
				}
				body, _ = json.Marshal(payload)
			}
			request.Body = io.NopCloser(bytes.NewReader(body))
			request.ContentLength = int64(len(body))
		}
		rr := httptest.NewRecorder()
		next.ServeHTTP(rr, request)
		for k, values := range rr.Header() {
			w.Header()[k] = values
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Del("Content-Length")
		var payload map[string]any
		if strings.Contains(rr.Header().Get("Content-Type"), "application/json") && json.Unmarshal(rr.Body.Bytes(), &payload) == nil {
			ids := map[string]bool{}
			projected := musicProject553(db, actorID, role)
			musicFileIDs553(projected, ids)
			var filter func(map[string]any)
			filter = func(value map[string]any) {
				for _, key := range []string{"proposals", "changes", "messages"} {
					if a, ok := value[key]; ok {
						rows := []any{}
						for _, raw := range anySlice548(a) {
							item := legacyMap(raw)
							keep := true
							if key == "proposals" {
								keep = stringValue(item["programId"]) == musicProgram553 || strings.HasPrefix(stringValue(item["requirementId"]), "LM-")
							}
							if key == "changes" {
								keep = musicAllowed553(db, legacyUser(db, stringValue(item["ownerId"]))) && musicValidatePatch553(item["patch"], db)
							}
							if key == "messages" {
								keep = musicAllowed553(db, legacyUser(db, stringValue(item["senderId"]))) && musicAllowed553(db, legacyUser(db, stringValue(item["recipientId"])))
							}
							if keep {
								filter(item)
								rows = append(rows, item)
							}
						}
						value[key] = rows
					}
				}
				if v := value["database"]; v != nil {
					value["database"] = musicProject553(v, actorID, role)
					musicFileIDs553(value["database"], ids)
				}
				if v := legacyMap(value["snapshot"]); v != nil {
					filter(v)
				}
				if a, ok := value["files"]; ok {
					files := []any{}
					for _, raw := range anySlice548(a) {
						if ids[stringValue(legacyMap(raw)["id"])] {
							files = append(files, raw)
						}
					}
					value["files"] = files
				}
				if a, ok := value["users"]; ok {
					users := []any{}
					for _, raw := range anySlice548(a) {
						m := legacyMap(raw)
						id := stringValue(m["userId"])
						if id == "" {
							id = stringValue(m["id"])
						}
						if musicAllowed553(db, legacyUser(db, id)) {
							users = append(users, raw)
						}
					}
					value["users"] = users
					if _, ok := value["count"]; ok {
						value["count"] = len(users)
					}
				}
			}
			filter(payload)
			stripRemoteCredentialFields508(payload)
			jsonOutRemote(w, r, rr.Code, payload)
			return
		}
		w.WriteHeader(rr.Code)
		_, _ = w.Write(rr.Body.Bytes())
	})
}

func (s *Server) musicWebRoutes553(m *http.ServeMux) {
	if !s.cfg.PublicWeb || s.cfg.WebRoot == "" {
		return
	}
	m.HandleFunc("/musica", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/musica/", http.StatusTemporaryRedirect)
	})
	root := filepath.Join(s.cfg.WebRoot, "musica")
	files := http.StripPrefix("/musica/", compressedStatic(root, http.FileServer(http.Dir(root))))
	m.HandleFunc("/musica/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.URL.Path == "/musica/sw.js" {
			w.Header().Set("Service-Worker-Allowed", "/musica/")
		}
		files.ServeHTTP(w, r)
	})
}
