package commands

import "github.com/halqme/blackboard/internal/store"

func currentTask(st store.State) *store.Task {
	for i := range st.Tasks {
		if st.Tasks[i].Status == "active" {
			return &st.Tasks[i]
		}
	}
	return nil
}

func currentTaskID(st store.State) string {
	if t := currentTask(st); t != nil {
		return t.ID
	}
	return ""
}

func currentStage(st store.State) string {
	if t := currentTask(st); t != nil {
		return t.Stage
	}
	return ""
}

func nextArtifactVersion(st store.State, kind, taskID string) int {
	max := 0
	for _, a := range st.Artifacts {
		if a.TaskID == taskID && a.Kind == kind && a.Version > max {
			max = a.Version
		}
	}
	return max + 1
}

func activeArtifacts(st store.State, t *store.Task) []store.Artifact {
	if t == nil {
		return nil
	}
	var out []store.Artifact
	for _, a := range st.Artifacts {
		if a.TaskID == t.ID && (a.Status == "active" || a.Status == "approved") {
			out = append(out, a)
		}
	}
	return out
}
