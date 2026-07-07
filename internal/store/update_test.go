package store

import (
	"errors"
	"testing"
)

func TestUpdateRejectsStaleRevision(t *testing.T) {
	home := t.TempDir()
	t.Setenv("BLACKBOARD_HOME", home)

	s := New("proj1")
	if err := s.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	first, err := s.Update(1, func(st *State) error {
		st.Tasks = append(st.Tasks, Task{
			ID:        "task-1",
			Title:     "first",
			Stage:     "intake",
			Status:    "active",
			CreatedAt: "2026-07-07T00:00:00Z",
			UpdatedAt: "2026-07-07T00:00:00Z",
		})
		return nil
	})
	if err != nil {
		t.Fatalf("first Update() error = %v", err)
	}
	if first.Revision != 2 {
		t.Fatalf("first revision = %d, want 2", first.Revision)
	}

	_, err = s.Update(1, func(st *State) error {
		st.Tasks[0].Title = "stale overwrite"
		return nil
	})
	if !errors.Is(err, ErrRevisionMismatch) {
		t.Fatalf("stale Update() error = %v, want ErrRevisionMismatch", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Revision != 2 {
		t.Fatalf("revision = %d, want 2", got.Revision)
	}
	if got.Tasks[0].Title != "first" {
		t.Fatalf("task title = %q, want first", got.Tasks[0].Title)
	}
}

func TestUpdateMutatesLatestPersistedState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("BLACKBOARD_HOME", home)

	s := New("proj1")
	if err := s.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	if _, err := s.Update(1, func(st *State) error {
		st.Tasks = append(st.Tasks, Task{
			ID:        "task-1",
			Title:     "persisted",
			Stage:     "intake",
			Status:    "active",
			CreatedAt: "2026-07-07T00:00:00Z",
			UpdatedAt: "2026-07-07T00:00:00Z",
		})
		return nil
	}); err != nil {
		t.Fatalf("first Update() error = %v", err)
	}

	updated, err := s.Update(2, func(st *State) error {
		if len(st.Tasks) != 1 || st.Tasks[0].Title != "persisted" {
			t.Fatalf("mutate received %+v, want latest persisted state", st.Tasks)
		}
		st.Tasks[0].Title = "updated"
		return nil
	})
	if err != nil {
		t.Fatalf("second Update() error = %v", err)
	}
	if updated.Revision != 3 {
		t.Fatalf("revision = %d, want 3", updated.Revision)
	}
	if updated.Tasks[0].Title != "updated" {
		t.Fatalf("task title = %q, want updated", updated.Tasks[0].Title)
	}
}

func TestUpdateRollsBackRevisionAndMutationOnError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("BLACKBOARD_HOME", home)

	s := New("proj1")
	if err := s.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	wantErr := errors.New("stop")
	_, err := s.Update(1, func(st *State) error {
		st.Tasks = append(st.Tasks, Task{ID: "task-1"})
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Update() error = %v, want %v", err, wantErr)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Revision != 1 {
		t.Fatalf("revision = %d, want 1", got.Revision)
	}
	if len(got.Tasks) != 0 {
		t.Fatalf("tasks = %+v, want none", got.Tasks)
	}
}
