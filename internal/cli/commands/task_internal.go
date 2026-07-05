package commands

import (
	"fmt"
	"strings"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func CmdPauseActiveTask(_ []string, s store.Store, st store.State) error {
	pauseActiveTask(&st)
	st.Revision++
	if err := s.Save(st); err != nil {
		return err
	}
	return nil
}

func CmdCreateTask(args []string, s store.Store, st store.State) error {
	if len(args) < 1 {
		return commandkit.Usage("task title is required")
	}
	task := createTask(&st, strings.Join(args, " "))
	st.Revision++
	if err := s.Save(st); err != nil {
		return err
	}
	fmt.Fprintf(commandkit.Out, "created task %s at revision v%d\n", task.ID, st.Revision)
	return nil
}
