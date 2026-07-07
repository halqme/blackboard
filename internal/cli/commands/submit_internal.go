package commands

import (
	"fmt"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func CmdValidateSubmit(args []string) error {
	_, _, err := validateSubmitArgs(args)
	return err
}

func CmdStoreArtifact(args []string, s store.Store, st store.State) error {
	if len(args) < 2 {
		return commandkit.Usage("kind and --file are required")
	}

	var art store.Artifact
	_, err := updateState(s, st.Revision, func(current *store.State) error {
		stored, err := storeArtifact(current, args[0], commandkit.Value(args, "--file"))
		if err != nil {
			return err
		}
		art = stored
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(commandkit.Out, "stored %s.%d as %s\n", art.Kind, art.Version, art.ID)
	return nil
}

func CmdAdvanceTaskStage(args []string, s store.Store, st store.State) error {
	if len(args) < 1 {
		return commandkit.Usage("stage is required")
	}
	_, err := updateState(s, st.Revision, func(current *store.State) error {
		advanceTaskStage(current, args[0])
		return nil
	})
	return err
}
