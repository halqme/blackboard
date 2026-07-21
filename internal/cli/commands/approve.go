package commands

import (
	"fmt"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func CmdApprove(args []string, s store.Store, _ store.State) error {
	if len(args) == 0 {
		return commandkit.Usage("artifact id required\n\nNext action:\n  - Try 'bb help approve' for command usage.")
	}
	expectedRevision, err := requiredBasedOnRevision(args)
	if err != nil {
		return err
	}
	id := args[0]

	_, err = updateState(s, expectedRevision, func(st *store.State) error {
		for i := range st.Artifacts {
			if st.Artifacts[i].ID != id {
				continue
			}
			if st.Artifacts[i].Status == "stale" {
				return commandkit.ArtifactValidation("stale artifact cannot be approved: " + id)
			}
			if _, err := s.ValidateBlob(st.Artifacts[i].BlobHash); err != nil {
				return commandkit.ArtifactValidation("artifact blob is invalid: " + id + ": " + err.Error())
			}
			st.Artifacts[i].Status = "approved"
			return nil
		}
		return commandkit.ArtifactValidation("artifact not found: " + id)
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(commandkit.Out, "approved %s\n", id)
	return nil
}
