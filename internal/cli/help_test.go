package cli

import (
	"bytes"
	"strings"
	"testing"

	cmds "github.com/halqme/blackboard/internal/cli/commands"
)

func TestCmdGuidePrintsBootstrapGuidance(t *testing.T) {
	out := &bytes.Buffer{}
	restoreIO := setCLIIO(t, strings.NewReader(""), out)
	defer restoreIO()

	if err := cmds.CmdGuide(nil); err != nil {
		t.Fatalf("cmdGuide() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "blackboard.yaml") {
		t.Fatalf("guide output missing blackboard.yaml: %q", got)
	}
	if !strings.Contains(got, ".agents") {
		t.Fatalf("guide output missing .agents: %q", got)
	}
	if !strings.Contains(got, "AGENTS.md") || !strings.Contains(got, ".agents") {
		t.Fatalf("guide output missing repo instructions: %q", got)
	}
	if !strings.Contains(got, "bb task new <title>") {
		t.Fatalf("guide output missing task creation hint: %q", got)
	}
}

func TestRunGuidePrintsBootstrapGuidance(t *testing.T) {
	out := &bytes.Buffer{}
	restoreIO := setCLIIO(t, strings.NewReader(""), out)
	defer restoreIO()

	if err := Run([]string{"guide"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !strings.Contains(out.String(), "Blackboard rollout guide") {
		t.Fatalf("Run(guide) output = %q, want rollout guide", out.String())
	}
}

func TestRunVersionFlagPrintsVersion(t *testing.T) {
	out := &bytes.Buffer{}
	restoreIO := setCLIIO(t, strings.NewReader(""), out)
	defer restoreIO()

	prev := Version
	Version = "test-version"
	defer func() { Version = prev }()

	if err := Run([]string{"--version"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(out.String(), "test-version") {
		t.Fatalf("output = %q, want version", out.String())
	}
}

func TestRunUnknownCommandSuggestsNextAction(t *testing.T) {
	err := Run([]string{"--bogus"})
	exitErr := assertExitError(t, err, 2)
	if !strings.Contains(exitErr.Message, "unknown command: --bogus") {
		t.Fatalf("error message = %q, want unknown command", exitErr.Message)
	}
	if !strings.Contains(exitErr.Message, "Try 'bb help'") {
		t.Fatalf("error message = %q, want help suggestion", exitErr.Message)
	}
}

func TestRunHelpDoesNotShowInternalCommandsByDefault(t *testing.T) {
	out := &bytes.Buffer{}
	restoreIO := setCLIIO(t, strings.NewReader(""), out)
	defer restoreIO()

	if err := Run([]string{"help"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	got := out.String()
	if strings.Contains(got, "write-config") {
		t.Fatalf("help output = %q, should hide internal commands", got)
	}
}

func TestRunHelpDeepShowsInternalCommands(t *testing.T) {
	out := &bytes.Buffer{}
	restoreIO := setCLIIO(t, strings.NewReader(""), out)
	defer restoreIO()

	if err := Run([]string{"help", "--deep"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "write-config") || !strings.Contains(got, "install-skill") {
		t.Fatalf("help output = %q, want init internals", got)
	}
	if !strings.Contains(got, "pause-active-task") || !strings.Contains(got, "advance-task-stage") {
		t.Fatalf("help output = %q, want workflow internals", got)
	}
}

func TestRunHelpForCommandPrintsCommandSpecificUsage(t *testing.T) {
	out := &bytes.Buffer{}
	restoreIO := setCLIIO(t, strings.NewReader(""), out)
	defer restoreIO()

	if err := Run([]string{"help", "task"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Usage:\n  bb task <subcommand>") {
		t.Fatalf("help output = %q, want parent task usage", got)
	}
	if !strings.Contains(got, "Subcommands:") || !strings.Contains(got, "new") {
		t.Fatalf("help output = %q, want task subcommands", got)
	}
}

func TestRunCommandHelpFlagPrintsCommandSpecificUsage(t *testing.T) {
	out := &bytes.Buffer{}
	restoreIO := setCLIIO(t, strings.NewReader(""), out)
	defer restoreIO()

	if err := Run([]string{"task", "--help"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Usage:\n  bb task <subcommand>") {
		t.Fatalf("help output = %q, want parent task usage", got)
	}
}

func TestRunNestedCommandHelpFlagPrintsLeafUsage(t *testing.T) {
	out := &bytes.Buffer{}
	restoreIO := setCLIIO(t, strings.NewReader(""), out)
	defer restoreIO()

	if err := Run([]string{"task", "new", "--help"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Usage:\n  bb task new <title>") {
		t.Fatalf("help output = %q, want leaf task usage", got)
	}
}

func TestRunParentHelpListsSubcommands(t *testing.T) {
	out := &bytes.Buffer{}
	restoreIO := setCLIIO(t, strings.NewReader(""), out)
	defer restoreIO()

	if err := Run([]string{"help", "artifact"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Subcommands:") {
		t.Fatalf("help output = %q, want subcommand section", got)
	}
	if !strings.Contains(got, "list") {
		t.Fatalf("help output = %q, want artifact list subcommand", got)
	}
}

func TestRunNestedHelpPrintsSubcommandUsage(t *testing.T) {
	out := &bytes.Buffer{}
	restoreIO := setCLIIO(t, strings.NewReader(""), out)
	defer restoreIO()

	if err := Run([]string{"help", "artifact", "list"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Usage:\n  bb artifact list [--json]") {
		t.Fatalf("help output = %q, want nested subcommand usage", got)
	}
}
