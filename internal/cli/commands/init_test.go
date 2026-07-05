package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	if err := CmdInit([]string{"--with-agent-files"}); err != nil {
		t.Fatalf("CmdInit() error = %v", err)
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

	if err := CmdInit([]string{"--no-agent-files"}); err != nil {
		t.Fatalf("CmdInit() error = %v", err)
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

	restoreIO := setCommandIO(t, strings.NewReader("y\n"), &bytes.Buffer{})
	defer restoreIO()

	if err := CmdInit(nil); err != nil {
		t.Fatalf("CmdInit() error = %v", err)
	}

	assertFileContains(t, filepath.Join(wd, "AGENTS.md"), "## Blackboard")
	assertFileContains(t, filepath.Join(wd, ".agents", "skills", "blackboard", "SKILL.md"), "Mandatory skill for repositories adopting blackboard workflow.")
}
