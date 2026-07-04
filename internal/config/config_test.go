package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRootFindsNearestBlackboardYAML(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, "proj1", "Demo")
	sub := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	got, err := FindRoot(sub)
	if err != nil {
		t.Fatalf("FindRoot() error = %v", err)
	}
	if got != root {
		t.Fatalf("FindRoot() = %q, want %q", got, root)
	}
}

func TestLoadParsesProjectContextAndCommands(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, "proj1", "Demo")

	got, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Version != 1 || got.Project.ID != "proj1" || got.Project.Name != "Demo" {
		t.Fatalf("Load() config = %+v", got)
	}
	if len(got.Context.Files) != 2 || got.Context.Files[0] != "README.md" {
		t.Fatalf("Load() context = %+v", got.Context.Files)
	}
	if got.Commands["test"] != "go test ./..." {
		t.Fatalf("Load() commands = %+v", got.Commands)
	}
}

func TestLoadRejectsMissingProjectID(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "blackboard.yaml"), []byte("version: 1\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := Load(root); err == nil {
		t.Fatal("Load() error = nil, want missing project.id")
	}
}

func writeConfig(t *testing.T, root, id, name string) {
	t.Helper()
	content := "version: 1\nproject:\n  id: " + id + "\n  name: " + name + "\ncontext:\n  files:\n    - README.md\n    - AGENTS.md\ncommands:\n  test: \"go test ./...\"\n"
	if err := os.WriteFile(filepath.Join(root, "blackboard.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
