package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

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

func seedTestStore(t *testing.T, s store.Store, tasks []store.Task, artifacts []store.Artifact) store.State {
	t.Helper()
	st, err := s.Update(1, func(st *store.State) error {
		st.Tasks = append(st.Tasks, tasks...)
		st.Artifacts = append(st.Artifacts, artifacts...)
		return nil
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	return st
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

func setCommandIO(t *testing.T, in *strings.Reader, out *bytes.Buffer) func() {
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
