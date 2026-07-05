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
	pauseActiveTask(&st)
	createTask(&st, strings.Join(args, " "))
	st.Revision++
	if err := s.Save(st); err != nil {
		return err
	}
	fmt.Fprintf(commandkit.Out, "created task %s at revision v%d\n", st.Tasks[len(st.Tasks)-1].ID, st.Revision)
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
