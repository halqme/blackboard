package commands

import (
	"fmt"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func CmdArchive(args []string, s store.Store, st store.State) error {
	for i := range st.Tasks {
		if st.Tasks[i].Status == "active" {
			st.Tasks[i].Status = "archived"
			st.Tasks[i].Stage = "archived"
			st.Revision++
			if err := s.Save(st); err != nil {
				return err
			}
			fmt.Fprintln(commandkit.Out, "archived", st.Tasks[i].ID)
			return nil
		}
	}
	return commandkit.TaskNotFound("no active task")
}
