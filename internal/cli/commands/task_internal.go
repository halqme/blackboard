package commands

import (
	"fmt"
	"strings"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func CmdPauseActiveTask(_ []string, s store.Store, st store.State) error {
	_, err := updateState(s, st.Revision, func(current *store.State) error {
		pauseActiveTask(current)
		return nil
	})
	return err
}

func CmdCreateTask(args []string, s store.Store, st store.State) error {
	if len(args) < 1 {
		return commandkit.Usage("task title is required")
	}

	var task store.Task
	updated, err := updateState(s, st.Revision, func(current *store.State) error {
		task = createTask(current, strings.Join(args, " "))
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(commandkit.Out, "created task %s at revision v%d\n", task.ID, updated.Revision)
	return nil
}
