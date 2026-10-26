package main

import (
	"net/http"
	"testing"
)

func Test539ApprovedResetPermissions(t *testing.T) {
	for _, c := range []struct {
		actor, token string
		want         int
	}{{"asp1", "cli", 403}, {"t1", "cli", 200}, {"t2", "cli", 403}, {"admin1", "adm", 200}} {
		t.Run(c.actor, func(t *testing.T) {
			s := progressResetTestServer(t)
			rr := callProgressReset(t, s, c.token, map[string]any{"actorId": c.actor, "ownerId": "asp1", "programId": "class_grouped", "requirementId": "R2", "scope": "requirement", "operationId": "approved-reset"})
			if rr.Code != c.want {
				t.Fatal(rr.Code, rr.Body.String())
			}
			if c.want == http.StatusForbidden {
				st, _ := s.loadLegacy()
				if legacyRequirementStatus(st.State.Database, "asp1", "class_grouped", "R2") != "approved" || len(st.State.Files) != 2 {
					t.Fatal("approved evidence changed")
				}
			}
		})
	}
}
