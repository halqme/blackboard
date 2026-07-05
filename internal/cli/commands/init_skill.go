package commands

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/halqme/blackboard/internal/cli/commandkit"
)

func CmdWriteConfig(_ []string) error {
	if _, err := os.Stat("blackboard.yaml"); err == nil {
		return commandkit.ConfigErr("blackboard.yaml already exists")
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
	return os.WriteFile("blackboard.yaml", []byte(content), 0o644)
}

func CmdWriteAgentFiles(_ []string) error {
	return ensureAgentWorkflowFile("AGENTS.md", agentsWorkflowSnippet)
}

func CmdInstallSkill(_ []string) error {
	return writeInstalledSkill(filepath.Join(".agents", "skills", "blackboard"))
}

func shouldWriteAgentFiles(args []string) (bool, error) {
	if commandkit.Has(args, "--with-agent-files") {
		return true, nil
	}
	if commandkit.Has(args, "--no-agent-files") {
		return false, nil
	}
	return commandkit.PromptYesNo("Add blackboard workflow guidance to AGENTS.md?")
}

func ensureAgentWorkflowFile(path, snippet string) error {
	if b, err := os.ReadFile(path); err == nil {
		if strings.Contains(string(b), "## Blackboard") {
			return nil
		}
		content := strings.TrimRight(string(b), "\n") + "\n\n" + snippet + "\n"
		return os.WriteFile(path, []byte(content), 0o644)
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(snippet+"\n"), 0o644)
}

func writeInstalledSkill(root string) error {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return os.ErrNotExist
	}
	sourceRoot := filepath.Join(filepath.Dir(file), "..", "bootstrap", "blackboard-skill")
	return filepath.WalkDir(sourceRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		target := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}

const agentsWorkflowSnippet = `## Blackboard

This repository uses blackboard workflow.

At task start:

1. Run ` + "`bb status`" + `
2. Run ` + "`bb context`" + `
3. Run ` + "`bb next`" + `

If you work within a stage, run ` + "`bb stage <stage>`" + `.
Use ` + "`bb help`" + ` when command usage is unclear.
`
