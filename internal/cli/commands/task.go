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
	t := store.Task{ID: store.NewID("task"), Title: strings.Join(args, " "), Stage: "intake", Status: "active", CreatedAt: store.Now(), UpdatedAt: store.Now()}
	for i := range st.Tasks {
		if st.Tasks[i].Status == "active" {
			st.Tasks[i].Status = "paused"
			st.Tasks[i].UpdatedAt = store.Now()
		}
	}
	st.Tasks = append(st.Tasks, t)
	st.Revision++
	if err := s.Save(st); err != nil {
		return err
	}
	fmt.Fprintf(commandkit.Out, "created task %s at revision v%d\n", t.ID, st.Revision)
	return nil
}
