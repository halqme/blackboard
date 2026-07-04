package context

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/halqme/blackboard/internal/config"
	"github.com/halqme/blackboard/internal/store"
)

func TestRenderIncludesCurrentTaskAndActiveArtifacts(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "blackboard.yaml"), []byte("version: 1\nproject:\n  id: proj1\n  name: Demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("readme"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := execGit(root, "init"); err != nil {
		t.Fatalf("git init error = %v", err)
	}

	cfg := config.Config{}
	cfg.Project.ID = "proj1"
	cfg.Project.Name = "Demo"
	st := store.State{
		Revision:  2,
		Tasks:     []store.Task{{ID: "task-1", Title: "task", Stage: "proposal", Status: "active"}},
		Artifacts: []store.Artifact{{ID: "art-1", TaskID: "task-1", Kind: "proposal", Version: 1, Status: "active"}},
	}
	out := Render(root, cfg, st, &st.Tasks[0])
	if !strings.Contains(out, "Current revision: v2") || !strings.Contains(out, "art-1 proposal.v1 [active]") {
		t.Fatalf("Render() output missing expected content:\n%s", out)
	}
}

func TestStageInstruction(t *testing.T) {
	out := StageInstruction("proposal")
	if !strings.Contains(out, "Stage: proposal") || !strings.Contains(out, "bb submit proposal") {
		t.Fatalf("StageInstruction() = %q", out)
	}
}

func execGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}
