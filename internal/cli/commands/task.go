package commands

import (
	"fmt"
	"strings"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func CmdTaskNew(args []string, s store.Store, st store.State) error {
	if len(args) < 1 {
		return commandkit.Usage("task title is required\n\nNext action:\n  - Try 'bb help task new' for command usage.")
	}

	updated, err := updateState(s, st.Revision, func(current *store.State) error {
		pauseActiveTask(current)
		createTask(current, strings.Join(args, " "))
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(commandkit.Out, "created task %s at revision v%d\n", updated.Tasks[len(updated.Tasks)-1].ID, updated.Revision)
	return nil
}

func pauseActiveTask(st *store.State) {
	for i := range st.Tasks {
		if st.Tasks[i].Status == "active" {
			st.Tasks[i].Status = "paused"
			st.Tasks[i].UpdatedAt = store.Now()
		}
	}
}

func createTask(st *store.State, title string) store.Task {
	t := store.Task{ID: store.NewID("task"), Title: title, Stage: "intake", Status: "active", CreatedAt: store.Now(), UpdatedAt: store.Now()}
	st.Tasks = append(st.Tasks, t)
	return t
}
