package cli

import (
	"fmt"

	"github.com/halqme/blackboard/internal/cli/commandkit"
)

var Version = "dev"

func cmdVersion(_ []string) error {
	fmt.Fprintln(commandkit.Out, Version)
	return nil
}

func versionCommand() Command {
	return Command{
		Name:     "version",
		Usage:    "bb version",
		Abstract: "Show bb version",
		Run:      cmdVersion,
	}
}
