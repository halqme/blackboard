package commands

import (
	"testing"

	"github.com/halqme/blackboard/internal/store"
)

func TestValidateDependencies(t *testing.T) {
	st := store.State{Artifacts: []store.Artifact{
		{ID: "approved", Status: "approved"},
		{ID: "active", Status: "active"},
	}}

	if err := validateDependencies(st, []string{"approved"}); err != nil {
		t.Fatalf("approved dependency rejected: %v", err)
	}
	if err := validateDependencies(st, []string{"missing"}); err == nil {
		t.Fatal("missing dependency was accepted")
	}
	if err := validateDependencies(st, []string{"active"}); err == nil {
		t.Fatal("unapproved dependency was accepted")
	}
	if err := validateDependencies(st, []string{"approved", "approved"}); err == nil {
		t.Fatal("duplicate dependency was accepted")
	}
}

func TestSupersedePreviousAndMarkStaleTransitively(t *testing.T) {
	st := store.State{Artifacts: []store.Artifact{
		{ID: "proposal-v1", TaskID: "task-1", Kind: "proposal", Status: "approved"},
		{ID: "implementation", TaskID: "task-1", Kind: "implementation", Status: "approved", DependsOn: []string{"proposal-v1"}},
		{ID: "verification", TaskID: "task-1", Kind: "verification", Status: "approved", DependsOn: []string{"implementation"}},
		{ID: "unrelated", TaskID: "task-1", Kind: "notes", Status: "approved"},
		{ID: "proposal-v2", TaskID: "task-1", Kind: "proposal", Status: "active"},
	}}

	supersedePreviousAndMarkStale(&st, st.Artifacts[4])

	want := map[string]string{
		"proposal-v1":    "superseded",
		"implementation": "stale",
		"verification":   "stale",
		"unrelated":      "approved",
		"proposal-v2":    "active",
	}
	for _, a := range st.Artifacts {
		if a.Status != want[a.ID] {
			t.Fatalf("%s status = %s, want %s", a.ID, a.Status, want[a.ID])
		}
	}
}
