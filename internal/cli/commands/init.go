package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/config"
	"github.com/halqme/blackboard/internal/store"
)

func CmdInit(args []string) error {
	if _, err := os.Stat("blackboard.yaml"); err == nil {
		return commandkit.ConfigErr("blackboard.yaml already exists")
	}
	writeAgentFiles, err := shouldWriteAgentFiles(args)
	if err != nil {
		return err
	}
	wd, _ := os.Getwd()
	id := strings.ToLower(strings.ReplaceAll(filepath.Base(wd), " ", "-"))
	content := fmt.Sprintf(`version: 1
project:
  id: %s
  name: %s
context:
  files:
    - README.md
    - AGENTS.md
commands:
  test: ""
  lint: ""
  typecheck: ""
`, id, filepath.Base(wd))
	if err := os.WriteFile("blackboard.yaml", []byte(content), 0o644); err != nil {
		return err
	}
	if writeAgentFiles {
		if err := ensureAgentWorkflowFiles(); err != nil {
			return err
		}
		if err := installAgentSkills(); err != nil {
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
