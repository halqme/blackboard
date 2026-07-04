package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/halqme/blackboard/internal/store"
)

func TestCmdInitWithAgentFilesCreatesWorkflowDocsAndInstallsSkills(t *testing.T) {
	t.Setenv("BLACKBOARD_HOME", t.TempDir())
	wd := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	defer os.Chdir(prev)

	if err := cmdInit([]string{"--with-agent-files"}); err != nil {
		t.Fatalf("cmdInit() error = %v", err)
	}

	assertFileContains(t, filepath.Join(wd, "AGENTS.md"), "## Blackboard")
	assertFileContains(t, filepath.Join(wd, ".agents", "skills", "blackboard", "SKILL.md"), "Mandatory skill for repositories adopting blackboard workflow.")
}

func TestCmdInitWithoutAgentFilesSkipsWorkflowDocsAndSkillInstall(t *testing.T) {
	t.Setenv("BLACKBOARD_HOME", t.TempDir())
	wd := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	defer os.Chdir(prev)

	if err := cmdInit([]string{"--no-agent-files"}); err != nil {
		t.Fatalf("cmdInit() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(wd, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("AGENTS.md exists, want not exist; err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(wd, ".agents")); !os.IsNotExist(err) {
		t.Fatalf(".agents exists, want not exist; err=%v", err)
	}
}

func TestCmdInitInteractiveYesCreatesWorkflowDocsAndInstallsSkills(t *testing.T) {
	t.Setenv("BLACKBOARD_HOME", t.TempDir())
	wd := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	defer os.Chdir(prev)

	restoreIO := setCLIIO(t, strings.NewReader("y\n"), &bytes.Buffer{})
	defer restoreIO()

	if err := cmdInit(nil); err != nil {
		t.Fatalf("cmdInit() error = %v", err)
	}

	assertFileContains(t, filepath.Join(wd, "AGENTS.md"), "## Blackboard")
	assertFileContains(t, filepath.Join(wd, ".agents", "skills", "blackboard", "SKILL.md"), "Mandatory skill for repositories adopting blackboard workflow.")
}

func TestCmdGuidePrintsBootstrapGuidance(t *testing.T) {
	out := &bytes.Buffer{}
	restoreIO := setCLIIO(t, strings.NewReader(""), out)
	defer restoreIO()

	if err := cmdGuide(nil); err != nil {
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

func TestRunUnknownCommandSuggestsNextAction(t *testing.T) {
	err := Run([]string{"--version"})
	exitErr := assertExitError(t, err, 2)
	if !strings.Contains(exitErr.Message, "unknown command: --version") {
		t.Fatalf("error message = %q, want unknown command", exitErr.Message)
	}
	if !strings.Contains(exitErr.Message, "Try 'bb help'") {
		t.Fatalf("error message = %q, want help suggestion", exitErr.Message)
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
}

func TestIntermediateCommandsAreHelpOnly(t *testing.T) {
	for _, command := range rootCommands() {
		assertIntermediateCommandsHaveNoRun(t, []string{command.Name}, command)
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

func TestCmdTaskCreatesTaskAndPausesPreviousActiveTask(t *testing.T) {
	s := newTestStore(t)
	st := store.State{
		ProjectID: "proj1",
		Revision:  1,
		Tasks: []store.Task{{
			ID:        "task-old",
			Title:     "old task",
			Stage:     "intake",
			Status:    "active",
			CreatedAt: "2026-07-04T00:00:00Z",
			UpdatedAt: "2026-07-04T00:00:00Z",
		}},
	}
	if err := s.Save(st); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := cmdTaskNew([]string{"new", "task"}, s, st); err != nil {
		t.Fatalf("cmdTaskNew() error = %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Revision != 2 {
		t.Fatalf("revision = %d, want 2", got.Revision)
	}
	if len(got.Tasks) != 2 {
		t.Fatalf("tasks len = %d, want 2", len(got.Tasks))
	}
	if got.Tasks[0].Status != "paused" {
		t.Fatalf("old task status = %q, want paused", got.Tasks[0].Status)
	}
	if got.Tasks[1].Title != "new task" || got.Tasks[1].Status != "active" {
		t.Fatalf("new task = %+v, want active new task", got.Tasks[1])
	}
}

func TestCmdSubmitStoresArtifactAndAdvancesTaskStage(t *testing.T) {
	s := newTestStore(t)
	blobPath := filepath.Join(t.TempDir(), "proposal.txt")
	if err := os.WriteFile(blobPath, []byte("proposal body"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	st := store.State{
		ProjectID: "proj1",
		Revision:  1,
		Tasks: []store.Task{{
			ID:        "task-1",
			Title:     "task",
			Stage:     "intake",
			Status:    "active",
			CreatedAt: "2026-07-04T00:00:00Z",
			UpdatedAt: "2026-07-04T00:00:00Z",
		}},
	}
	if err := s.Save(st); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := cmdSubmit([]string{"proposal", "--file", blobPath, "--based-on", "v1"}, s, st); err != nil {
		t.Fatalf("cmdSubmit() error = %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Revision != 2 {
		t.Fatalf("revision = %d, want 2", got.Revision)
	}
	if got.Tasks[0].Stage != "proposal" {
		t.Fatalf("task stage = %q, want proposal", got.Tasks[0].Stage)
	}
	if len(got.Artifacts) != 1 {
		t.Fatalf("artifacts len = %d, want 1", len(got.Artifacts))
	}
	if got.Artifacts[0].Status != "active" || got.Artifacts[0].Kind != "proposal" {
		t.Fatalf("artifact = %+v, want active proposal", got.Artifacts[0])
	}
}

func TestCmdApproveMarksArtifactApproved(t *testing.T) {
	s := newTestStore(t)
	st := store.State{
		ProjectID: "proj1",
		Revision:  1,
		Tasks: []store.Task{{
			ID:        "task-1",
			Title:     "task",
			Stage:     "proposal",
			Status:    "active",
			CreatedAt: "2026-07-04T00:00:00Z",
			UpdatedAt: "2026-07-04T00:00:00Z",
		}},
		Artifacts: []store.Artifact{{
			ID:              "art-1",
			TaskID:          "task-1",
			Stage:           "proposal",
			Kind:            "proposal",
			Version:         1,
			Status:          "active",
			BlobHash:        "abc123",
			BasedOnRevision: 1,
			CreatedAt:       "2026-07-04T00:00:00Z",
		}},
	}
	if err := s.Save(st); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := cmdApprove([]string{"art-1", "--based-on", "v1"}, s, st); err != nil {
		t.Fatalf("cmdApprove() error = %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Revision != 2 {
		t.Fatalf("revision = %d, want 2", got.Revision)
	}
	if got.Artifacts[0].Status != "approved" {
		t.Fatalf("artifact status = %q, want approved", got.Artifacts[0].Status)
	}
}

func TestCmdArchiveMarksTaskArchived(t *testing.T) {
	s := newTestStore(t)
	st := store.State{
		ProjectID: "proj1",
		Revision:  1,
		Tasks: []store.Task{{
			ID:        "task-1",
			Title:     "task",
			Stage:     "proposal",
			Status:    "active",
			CreatedAt: "2026-07-04T00:00:00Z",
			UpdatedAt: "2026-07-04T00:00:00Z",
		}},
	}
	if err := s.Save(st); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := cmdArchive([]string{"--based-on", "v1"}, s, st); err != nil {
		t.Fatalf("cmdArchive() error = %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Revision != 2 {
		t.Fatalf("revision = %d, want 2", got.Revision)
	}
	if got.Tasks[0].Stage != "archived" || got.Tasks[0].Status != "archived" {
		t.Fatalf("task = %+v, want archived", got.Tasks[0])
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
	prevIn := cliIn
	prevOut := cliOut
	cliIn = in
	cliOut = out
	return func() {
		cliIn = prevIn
		cliOut = prevOut
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
