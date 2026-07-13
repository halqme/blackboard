package commands

import (
	"fmt"
	"os"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/fsm"
	"github.com/halqme/blackboard/internal/store"
)

func CmdSubmit(args []string, s store.Store, _ store.State) error {
	kind, file, err := validateSubmitArgs(args)
	if err != nil {
		return err
	}
	expectedRevision, err := requiredBasedOnRevision(args)
	if err != nil {
		return err
	}
	dependsOn := values(args, "--depends-on")
	var art store.Artifact
	_, err = updateState(s, expectedRevision, func(st *store.State) error {
		if err := validateDependencies(*st, dependsOn); err != nil {
			return err
		}
		stored, err := storeArtifact(st, kind, file, dependsOn)
		if err != nil {
			return err
		}
		art = stored
		supersedePreviousAndMarkStale(st, art)
		advanceTaskStage(st, kind)
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(commandkit.Out, "submitted %s.%d as %s\n", art.Kind, art.Version, art.ID)
	return nil
}

func validateSubmitArgs(args []string) (kind string, file string, err error) {
	if len(args) < 2 {
		return "", "", commandkit.Usage("kind and --file are required\n\nNext action:\n  - Try 'bb help submit' for command usage.")
	}
	kind = args[0]
	file = commandkit.Value(args, "--file")
	if file == "" || commandkit.Value(args, "--based-on") == "" {
		return "", "", commandkit.Usage("--file and --based-on are required\n\nNext action:\n  - Try 'bb help submit' for command usage.")
	}
	if _, err := requiredBasedOnRevision(args); err != nil {
		return "", "", err
	}
	return kind, file, nil
}

func values(args []string, flag string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		if args[i] == flag && i+1 < len(args) {
			out = append(out, args[i+1])
			i++
		}
	}
	return out
}

func validateDependencies(st store.State, ids []string) error {
	seen := map[string]bool{}
	byID := map[string]store.Artifact{}
	for _, a := range st.Artifacts {
		byID[a.ID] = a
	}
	for _, id := range ids {
		if seen[id] {
			return commandkit.ArtifactValidation("duplicate dependency: " + id)
		}
		seen[id] = true
		a, ok := byID[id]
		if !ok {
			return commandkit.ArtifactValidation("dependency not found: " + id)
		}
		if a.Status != "approved" {
			return commandkit.ArtifactValidation("dependency is not approved: " + id)
		}
	}
	return nil
}

func storeArtifact(st *store.State, kind, file string, dependsOn []string) (store.Artifact, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return store.Artifact{}, err
	}
	id := store.NewID("art")
	taskID := currentTaskID(*st)
	ver := nextArtifactVersion(*st, kind, taskID)
	art := store.Artifact{ID: id, TaskID: taskID, Stage: currentStage(*st), Kind: kind, Version: ver, Status: "active", BlobHash: fmt.Sprintf("%x", len(b)), BasedOnRevision: st.Revision, CreatedAt: store.Now(), DependsOn: append([]string(nil), dependsOn...)}
	st.Artifacts = append(st.Artifacts, art)
	return art, nil
}

func supersedePreviousAndMarkStale(st *store.State, newest store.Artifact) {
	var superseded []string
	for i := range st.Artifacts {
		a := &st.Artifacts[i]
		if a.ID != newest.ID && a.TaskID == newest.TaskID && a.Kind == newest.Kind && (a.Status == "active" || a.Status == "approved") {
			a.Status = "superseded"
			superseded = append(superseded, a.ID)
		}
	}
	stale := map[string]bool{}
	for _, id := range superseded {
		stale[id] = true
	}
	changed := true
	for changed {
		changed = false
		for i := range st.Artifacts {
			a := &st.Artifacts[i]
			if a.ID == newest.ID || a.Status == "superseded" || a.Status == "stale" {
				continue
			}
			for _, dep := range a.DependsOn {
				if stale[dep] {
					a.Status = "stale"
					stale[a.ID] = true
					changed = true
					break
				}
			}
		}
	}
}

func advanceTaskStage(st *store.State, kind string) {
	if !fsm.IsStage(kind) {
		return
	}
	taskID := currentTaskID(*st)
	for i := range st.Tasks {
		if st.Tasks[i].ID == taskID {
			st.Tasks[i].Stage = kind
			st.Tasks[i].UpdatedAt = store.Now()
		}
	}
}
