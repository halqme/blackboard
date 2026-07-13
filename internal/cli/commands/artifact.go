package commands

import (
	"fmt"
	"strings"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func CmdArtifactList(args []string, s store.Store, st store.State) error {
	if commandkit.Has(args, "--json") {
		return commandkit.PrintJSON(st.Artifacts)
	}
	for _, a := range st.Artifacts {
		dependencies := "-"
		if len(a.DependsOn) > 0 {
			dependencies = strings.Join(a.DependsOn, ",")
		}
		fmt.Fprintf(commandkit.Out, "%s %s %s depends-on=%s\n", a.ID, a.Kind, a.Status, dependencies)
	}
	return nil
}
