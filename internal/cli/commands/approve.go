package commands

import (
	"fmt"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func CmdApprove(args []string, s store.Store, st store.State) error {
	if len(args) == 0 {
		return commandkit.Usage("artifact id required\n\nNext action:\n  - Try 'bb help approve' for command usage.")
	}
	id := args[0]
	for i := range st.Artifacts {
		if st.Artifacts[i].ID == id {
			st.Artifacts[i].Status = "approved"
			st.Revision++
			if err := s.Save(st); err != nil {
				return err
			}
			fmt.Fprintf(commandkit.Out, "approved %s\n", id)
			return nil
		}
	}
	return commandkit.ArtifactValidation("artifact not found: " + id)
}
