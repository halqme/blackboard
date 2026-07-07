package store

import (
	"os"
	"path/filepath"
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

func TestSaveAndLoadRoundTripThroughSQLite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("BLACKBOARD_HOME", home)

	s := New("proj1")
	if err := s.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	want := State{
		ProjectID: "proj1",
		Revision:  3,
		Tasks: []Task{{
			ID:        "task-1",
			Title:     "do thing",
			Stage:     "proposal",
			Status:    "active",
			CreatedAt: "2026-07-04T00:00:00Z",
			UpdatedAt: "2026-07-04T00:01:00Z",
		}},
		Artifacts: []Artifact{{
			ID:              "art-1",
			TaskID:          "task-1",
			Stage:           "proposal",
			Kind:            "proposal",
			Version:         1,
			Status:          "active",
			BlobHash:        "abc123",
			BasedOnRevision: 2,
			CreatedAt:       "2026-07-04T00:02:00Z",
		}},
	}

	if err := s.Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got.ProjectID != want.ProjectID || got.Revision != want.Revision {
		t.Fatalf("Load() = %+v, want %+v", got, want)
	}
	if len(got.Tasks) != 1 || got.Tasks[0] != want.Tasks[0] {
		t.Fatalf("Load() tasks = %+v, want %+v", got.Tasks, want.Tasks)
	}
	if len(got.Artifacts) != 1 || got.Artifacts[0] != want.Artifacts[0] {
		t.Fatalf("Load() artifacts = %+v, want %+v", got.Artifacts, want.Artifacts)
	}
}

func TestSaveUpdatesRowsWithoutRecreatingThem(t *testing.T) {
	home := t.TempDir()
	t.Setenv("BLACKBOARD_HOME", home)

	s := New("proj1")
	if err := s.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	st := State{
		ProjectID: "proj1",
		Revision:  2,
		Tasks: []Task{{
			ID:        "task-1",
			Title:     "before",
			Stage:     "proposal",
			Status:    "active",
			CreatedAt: "2026-07-04T00:00:00Z",
			UpdatedAt: "2026-07-04T00:01:00Z",
		}},
		Artifacts: []Artifact{{
			ID:              "art-1",
			TaskID:          "task-1",
			Stage:           "proposal",
			Kind:            "proposal",
			Version:         1,
			Status:          "active",
			BlobHash:        "abc123",
			BasedOnRevision: 1,
			CreatedAt:       "2026-07-04T00:02:00Z",
		}},
	}
	if err := s.Save(st); err != nil {
		t.Fatalf("first Save() error = %v", err)
	}

	db, err := s.openDB()
	if err != nil {
		t.Fatalf("openDB() error = %v", err)
	}
	defer db.Close()

	var taskRowIDBefore, artifactRowIDBefore int64
	if err := db.QueryRow(`SELECT rowid FROM tasks WHERE id = ?`, "task-1").Scan(&taskRowIDBefore); err != nil {
		t.Fatalf("query task rowid before update: %v", err)
	}
	if err := db.QueryRow(`SELECT rowid FROM artifacts WHERE id = ?`, "art-1").Scan(&artifactRowIDBefore); err != nil {
		t.Fatalf("query artifact rowid before update: %v", err)
	}

	st.Revision = 3
	st.Tasks[0].Title = "after"
	st.Tasks[0].UpdatedAt = "2026-07-04T00:03:00Z"
	st.Artifacts[0].Status = "approved"
	if err := s.Save(st); err != nil {
		t.Fatalf("second Save() error = %v", err)
	}

	var taskRowIDAfter, artifactRowIDAfter int64
	if err := db.QueryRow(`SELECT rowid FROM tasks WHERE id = ?`, "task-1").Scan(&taskRowIDAfter); err != nil {
		t.Fatalf("query task rowid after update: %v", err)
	}
	if err := db.QueryRow(`SELECT rowid FROM artifacts WHERE id = ?`, "art-1").Scan(&artifactRowIDAfter); err != nil {
		t.Fatalf("query artifact rowid after update: %v", err)
	}

	if taskRowIDAfter != taskRowIDBefore {
		t.Fatalf("task row recreated: rowid before=%d after=%d", taskRowIDBefore, taskRowIDAfter)
	}
	if artifactRowIDAfter != artifactRowIDBefore {
		t.Fatalf("artifact row recreated: rowid before=%d after=%d", artifactRowIDBefore, artifactRowIDAfter)
	}
}

func TestSaveDeletesRowsRemovedFromState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("BLACKBOARD_HOME", home)

	s := New("proj1")
	if err := s.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	st := State{
		ProjectID: "proj1",
		Revision:  2,
		Tasks: []Task{{
			ID:        "task-1",
			Title:     "do thing",
			Stage:     "proposal",
			Status:    "active",
			CreatedAt: "2026-07-04T00:00:00Z",
			UpdatedAt: "2026-07-04T00:01:00Z",
		}},
		Artifacts: []Artifact{{
			ID:              "art-1",
			TaskID:          "task-1",
			Stage:           "proposal",
			Kind:            "proposal",
			Version:         1,
			Status:          "active",
			BlobHash:        "abc123",
			BasedOnRevision: 1,
			CreatedAt:       "2026-07-04T00:02:00Z",
		}},
	}
	if err := s.Save(st); err != nil {
		t.Fatalf("first Save() error = %v", err)
	}

	st.Revision = 3
	st.Tasks = nil
	st.Artifacts = nil
	if err := s.Save(st); err != nil {
		t.Fatalf("second Save() error = %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(got.Tasks) != 0 {
		t.Fatalf("Load() tasks = %+v, want none", got.Tasks)
	}
	if len(got.Artifacts) != 0 {
		t.Fatalf("Load() artifacts = %+v, want none", got.Artifacts)
	}
}
