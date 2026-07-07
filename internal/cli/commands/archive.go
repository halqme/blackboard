package commands

import (
	"fmt"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func CmdArchive(args []string, s store.Store, _ store.State) error {
	expectedRevision, err := requiredBasedOnRevision(args)
	if err != nil {
		return err
	}

	var archivedTaskID string
	_, err = updateState(s, expectedRevision, func(st *store.State) error {
		for i := range st.Tasks {
			if st.Tasks[i].Status == "active" {
				st.Tasks[i].Status = "archived"
				st.Tasks[i].Stage = "archived"
				archivedTaskID = st.Tasks[i].ID
				return nil
			}
		}
		return commandkit.TaskNotFound("no active task")
	})
	if err != nil {
		return err
	}

	fmt.Fprintln(commandkit.Out, "archived", archivedTaskID)
	return nil
}
