package store

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestEnsureCreatesSQLiteStateDatabase(t *testing.T) {
	home := t.TempDir()
	t.Setenv("BLACKBOARD_HOME", home)

	s := New("proj1")
	if err := s.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(s.Root, "state.db")); err != nil {
		t.Fatalf("state.db not created: %v", err)
	}
}

func TestUpdateAndLoadRoundTripThroughSQLite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("BLACKBOARD_HOME", home)

	s := New("proj1")
	if err := s.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	updated, err := s.Update(1, func(st *State) error {
		st.Tasks = append(st.Tasks, Task{
			ID:        "task-1",
			Title:     "do thing",
			Stage:     "proposal",
			Status:    "active",
			CreatedAt: "2026-07-04T00:00:00Z",
			UpdatedAt: "2026-07-04T00:01:00Z",
		})
		st.Artifacts = append(st.Artifacts, Artifact{
			ID:              "art-1",
			TaskID:          "task-1",
			Stage:           "proposal",
			Kind:            "proposal",
			Version:         1,
			Status:          "active",
			BlobHash:        "abc123",
			BasedOnRevision: 1,
			CreatedAt:       "2026-07-04T00:02:00Z",
		})
		return nil
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got.ProjectID != updated.ProjectID || got.Revision != updated.Revision {
		t.Fatalf("Load() = %+v, want %+v", got, updated)
	}
	if len(got.Tasks) != 1 || got.Tasks[0] != updated.Tasks[0] {
		t.Fatalf("Load() tasks = %+v, want %+v", got.Tasks, updated.Tasks)
	}
	if len(got.Artifacts) != 1 || !reflect.DeepEqual(got.Artifacts[0], updated.Artifacts[0]) {
		t.Fatalf("Load() artifacts = %+v, want %+v", got.Artifacts, updated.Artifacts)
	}
}
