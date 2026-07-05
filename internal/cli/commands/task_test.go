package commands

import (
	"testing"

	"github.com/halqme/blackboard/internal/store"
)

func TestPauseActiveTaskPausesCurrentTask(t *testing.T) {
	st := store.State{Tasks: []store.Task{{ID: "task-old", Status: "active", UpdatedAt: "2026-07-04T00:00:00Z"}}}
	pauseActiveTask(&st)
	if st.Tasks[0].Status != "paused" {
		t.Fatalf("status = %q, want paused", st.Tasks[0].Status)
	}
}

func TestCreateTaskAppendsActiveTask(t *testing.T) {
	st := store.State{}
	task := createTask(&st, "new task")
	if len(st.Tasks) != 1 || task.Title != "new task" || st.Tasks[0].Status != "active" {
		t.Fatalf("state = %+v task = %+v", st, task)
	}
}

func TestCmdTaskCreatesTaskAndPausesPreviousActiveTask(t *testing.T) {
	s := newTestStore(t)
	st := store.State{ProjectID: "proj1", Revision: 1, Tasks: []store.Task{{ID: "task-old", Title: "old task", Stage: "intake", Status: "active", CreatedAt: "2026-07-04T00:00:00Z", UpdatedAt: "2026-07-04T00:00:00Z"}}}
	if err := s.Save(st); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := CmdTaskNew([]string{"new", "task"}, s, st); err != nil {
		t.Fatalf("CmdTaskNew() error = %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Revision != 2 || len(got.Tasks) != 2 {
		t.Fatalf("state = %+v", got)
	}
	if got.Tasks[0].Status != "paused" {
		t.Fatalf("old task status = %q, want paused", got.Tasks[0].Status)
	}
	if got.Tasks[1].Title != "new task" || got.Tasks[1].Status != "active" {
		t.Fatalf("new task = %+v, want active new task", got.Tasks[1])
	}
}
