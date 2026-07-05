package commands

import (
	"fmt"
	"os"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/fsm"
	"github.com/halqme/blackboard/internal/store"
)

func CmdSubmit(args []string, s store.Store, st store.State) error {
	if len(args) < 2 {
		return commandkit.Usage("kind and --file are required\n\nNext action:\n  - Try 'bb help submit' for command usage.")
	}
	kind := args[0]
	file := commandkit.Value(args, "--file")
	based := commandkit.Value(args, "--based-on")
	if file == "" || based == "" {
		return commandkit.Usage("--file and --based-on are required\n\nNext action:\n  - Try 'bb help submit' for command usage.")
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	id := store.NewID("art")
	taskID := currentTaskID(st)
	ver := nextArtifactVersion(st, kind, taskID)
	art := store.Artifact{ID: id, TaskID: taskID, Stage: currentStage(st), Kind: kind, Version: ver, Status: "active", BlobHash: fmt.Sprintf("%x", len(b)), BasedOnRevision: st.Revision, CreatedAt: store.Now()}
	st.Artifacts = append(st.Artifacts, art)
	if fsm.IsStage(kind) {
		for i := range st.Tasks {
			if st.Tasks[i].ID == taskID {
				st.Tasks[i].Stage = kind
				st.Tasks[i].UpdatedAt = store.Now()
			}
		}
	}
	st.Revision++
	if err := s.Save(st); err != nil {
		return err
	}
	fmt.Fprintf(commandkit.Out, "submitted %s.%d as %s\n", kind, ver, id)
	return nil
}
