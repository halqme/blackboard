package commands

import (
	"fmt"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func validateDependencies(st store.State, artifactID string, dependsOn []string) error {
	seen := map[string]bool{}
	byID := map[string]store.Artifact{}
	for _, a := range st.Artifacts {
		byID[a.ID] = a
	}
	for _, id := range dependsOn {
		if id == artifactID {
			return commandkit.ArtifactValidation("artifact cannot depend on itself: " + id)
		}
		if seen[id] {
			return commandkit.ArtifactValidation("duplicate dependency: " + id)
		}
		seen[id] = true
		a, ok := byID[id]
		if !ok {
			return commandkit.ArtifactValidation("dependency not found: " + id)
		}
		if a.Status != "approved" {
			return commandkit.ArtifactValidation(fmt.Sprintf("dependency must be approved: %s is %s", id, a.Status))
		}
		if reaches(st, id, artifactID, map[string]bool{}) {
			return commandkit.ArtifactValidation("dependency cycle detected through: " + id)
		}
	}
	return nil
}

func reaches(st store.State, from, target string, seen map[string]bool) bool {
	if from == target {
		return true
	}
	if seen[from] {
		return false
	}
	seen[from] = true
	for _, a := range st.Artifacts {
		if a.ID != from {
			continue
		}
		for _, next := range a.DependsOn {
			if reaches(st, next, target, seen) {
				return true
			}
		}
	}
	return false
}

func supersedePreviousAndMarkStale(st *store.State, newArtifact store.Artifact) {
	var superseded []string
	for i := range st.Artifacts {
		a := &st.Artifacts[i]
		if a.ID == newArtifact.ID || a.TaskID != newArtifact.TaskID || a.Kind != newArtifact.Kind {
			continue
		}
		if a.Status == "active" || a.Status == "approved" {
			a.Status = "superseded"
			superseded = append(superseded, a.ID)
		}
	}
	markDependentsStale(st, superseded)
}

func markDependentsStale(st *store.State, roots []string) {
	queue := append([]string(nil), roots...)
	seen := map[string]bool{}
	for len(queue) > 0 {
		root := queue[0]
		queue = queue[1:]
		if seen[root] {
			continue
		}
		seen[root] = true
		for i := range st.Artifacts {
			a := &st.Artifacts[i]
			if a.Status == "superseded" || a.Status == "stale" {
				continue
			}
			for _, dependency := range a.DependsOn {
				if dependency == root {
					a.Status = "stale"
					queue = append(queue, a.ID)
					break
				}
			}
		}
	}
}
