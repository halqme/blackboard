package commands

import (
	"fmt"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func CmdArtifactList(args []string, s store.Store, st store.State) error {
	if commandkit.Has(args, "--json") {
		return commandkit.PrintJSON(st.Artifacts)
	}
	for _, a := range st.Artifacts {
		fmt.Fprintf(commandkit.Out, "%s %s %s\n", a.ID, a.Kind, a.Status)
	}
	return nil
}
