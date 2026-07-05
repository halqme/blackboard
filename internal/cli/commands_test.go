package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func TestRootCommandsAreDeclaredByCommandFunctions(t *testing.T) {
	commands := rootCommands()

	if findCommand(commands, "task") == nil {
		t.Fatalf("rootCommands() missing task command")
	}
	if findCommand(commands, "artifact") == nil {
		t.Fatalf("rootCommands() missing artifact command")
	}
}

func TestParentCommandsRegisterSubcommandDeclarations(t *testing.T) {
	task := taskCommand()
	child := findCommand(task.Children, "new")
	if child == nil {
		t.Fatalf("taskCommand() missing new subcommand")
	}
	if child.Run == nil {
		t.Fatalf("task new command missing Run")
	}
	if child.Usage != "bb task new <title>" {
		t.Fatalf("task new usage = %q, want %q", child.Usage, "bb task new <title>")
	}

	artifact := artifactCommand()
	if findCommand(artifact.Children, "list") == nil {
		t.Fatalf("artifactCommand() missing list subcommand")
	}
	if findCommand(rootCommands(), "write-config") == nil {
		t.Fatalf("rootCommands() missing write-config command")
	}
}

func TestIntermediateCommandsAreHelpOnly(t *testing.T) {
	for _, command := range rootCommands() {
		assertIntermediateCommandsHaveNoRun(t, []string{command.Name}, command)
	}
}

func TestCommandCatalogLeafCommandsHaveHelp(t *testing.T) {
	for _, command := range rootCommands() {
		assertCommandHelpCoverage(t, []string{command.Name}, command)
	}
}

func TestCommandCatalogLeafCommandsRespondToHelp(t *testing.T) {
	for _, command := range rootCommands() {
		assertLeafHelpRuns(t, []string{command.Name}, command)
	}
}

func TestRunTaskUsesProjectConfigAndStore(t *testing.T) {
	home := t.TempDir()
	t.Setenv("BLACKBOARD_HOME", home)
	wd := t.TempDir()
	writeTestProject(t, wd, "proj1")
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	defer os.Chdir(prev)
	if err := Run([]string{"task", "new", "via run"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	s := store.New("proj1")
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Revision != 2 || len(got.Tasks) != 1 || got.Tasks[0].Title != "via run" {
		t.Fatalf("state = %+v, want one task from Run", got)
	}
}

func assertIntermediateCommandsHaveNoRun(t *testing.T, path []string, command Command) {
	t.Helper()
	if len(command.Children) > 0 && command.Run != nil {
		t.Fatalf("intermediate command %q must not have Run", strings.Join(path, " "))
	}
	for _, child := range command.Children {
		assertIntermediateCommandsHaveNoRun(t, append(path, child.Name), child)
	}
}

func newTestStore(t *testing.T) store.Store {
	t.Helper()
	home := t.TempDir()
	t.Setenv("BLACKBOARD_HOME", home)
	s := store.New("proj1")
	if err := s.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}
	return s
}

func writeTestProject(t *testing.T, dir, projectID string) {
	t.Helper()
	content := "version: 1\nproject:\n  id: " + projectID + "\n  name: Test Project\ncontext:\n  files:\n    - README.md\ncommands:\n  test: \"\"\n"
	if err := os.WriteFile(filepath.Join(dir, "blackboard.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if !strings.Contains(string(b), want) {
		t.Fatalf("%s does not contain %q; got %q", path, want, string(b))
	}
}

func setCLIIO(t *testing.T, in *strings.Reader, out *bytes.Buffer) func() {
	t.Helper()
	prevIn := commandkit.In
	prevOut := commandkit.Out
	commandkit.In = in
	commandkit.Out = out
	return func() {
		commandkit.In = prevIn
		commandkit.Out = prevOut
	}
}

func assertExitError(t *testing.T, err error, wantCode int) ExitError {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	exitErr, ok := err.(ExitError)
	if !ok {
		t.Fatalf("error = %T, want ExitError", err)
	}
	if exitErr.Code != wantCode {
		t.Fatalf("exit code = %d, want %d", exitErr.Code, wantCode)
	}
	return exitErr
}

func assertCommandHelpCoverage(t *testing.T, path []string, command Command) {
	t.Helper()
	if strings.TrimSpace(command.Abstract) == "" {
		t.Fatalf("command %q missing abstract", strings.Join(path, " "))
	}
	if strings.TrimSpace(command.Usage) == "" {
		t.Fatalf("command %q missing usage", strings.Join(path, " "))
	}
	if len(command.Children) == 0 {
		if command.Run == nil {
			t.Fatalf("leaf command %q missing Run", strings.Join(path, " "))
		}
		return
	}
	for _, child := range command.Children {
		assertCommandHelpCoverage(t, append(path, child.Name), child)
	}
}

func assertLeafHelpRuns(t *testing.T, path []string, command Command) {
	t.Helper()
	if len(command.Children) == 0 {
		out := &bytes.Buffer{}
		restoreIO := setCLIIO(t, strings.NewReader(""), out)
		defer restoreIO()

		args := append(append([]string{}, path...), "--help")
		if err := Run(args); err != nil {
			t.Fatalf("Run(%q) error = %v", strings.Join(args, " "), err)
		}
		if !strings.Contains(out.String(), "Usage:") {
			t.Fatalf("Run(%q) output = %q, want usage", strings.Join(args, " "), out.String())
		}
		return
	}
	for _, child := range command.Children {
		assertLeafHelpRuns(t, append(path, child.Name), child)
	}
}
