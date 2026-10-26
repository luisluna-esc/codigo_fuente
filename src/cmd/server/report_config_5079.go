package main

import (
	"net/http"
)

// reportConfig5079 returns only the report-delivery assignments needed by
// connected clients. It is intentionally lightweight so an Aspirant can refresh
// the Admin configuration before opening a requirement without downloading the
// full database snapshot.
func (s *Server) reportConfig5079(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := s.auth(r); !ok {
		legacyError(w, http.StatusUnauthorized, "Token o sesión no válidos.", nil)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	st, err := s.loadLegacy()
	if err != nil {
		legacyError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	modes := map[string]any{}
	if st.State.Database != nil {
		db := legacyMap(st.State.Database)
		if root := legacyMap(db["reportRequirementModes"]); len(root) > 0 {
			for key, value := range root {
				modes[key] = value
			}
		} else if settings := legacyMap(db["settings"]); len(settings) > 0 {
			for key, value := range legacyMap(settings["reportRequirementModes"]) {
				modes[key] = value
			}
		}
	}

	jsonOut(w, http.StatusOK, map[string]any{
		"ok":                     true,
		"compatVersion":          "5.6",
		"serverVersion":          st.State.ServerVersion,
		"reportRequirementModes": modes,
	})
}
