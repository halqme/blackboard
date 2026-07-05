package commands

import (
	"fmt"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/config"
	"github.com/halqme/blackboard/internal/store"
)

func CmdInit(args []string) error {
	writeAgentFiles, err := shouldWriteAgentFiles(args)
	if err != nil {
		return err
	}
	if err := CmdWriteConfig(nil); err != nil {
		return err
	}
	if writeAgentFiles {
		if err := CmdWriteAgentFiles(nil); err != nil {
			return err
		}
		if err := CmdInstallSkill(nil); err != nil {
			return err
		}
	}
	cfg, _ := config.Load(".")
	s := store.New(cfg.Project.ID)
	if err := s.Ensure(); err != nil {
		return err
	}
	fmt.Fprintln(commandkit.Out, "initialized blackboard project:", cfg.Project.ID)
	fmt.Fprintln(commandkit.Out, "store:", s.Root)
	return nil
}
