package commands

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func TestPublicWriteCommandsRejectStaleBasedOnRevision(t *testing.T) {
	t.Run("submit", func(t *testing.T) {
		s := newTestStore(t)
		st := store.State{
			ProjectID: "proj1",
			Revision:  1,
			Tasks: []store.Task{{
				ID: "task-1", Title: "task", Stage: "intake", Status: "active",
				CreatedAt: "2026-07-04T00:00:00Z", UpdatedAt: "2026-07-04T00:00:00Z",
			}},
		}
		if err := s.Save(st); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		advanceRevision(t, s, 1)

		file := filepath.Join(t.TempDir(), "proposal.md")
		if err := os.WriteFile(file, []byte("# Proposal"), 0o644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		err := CmdSubmit([]string{"proposal", "--file", file, "--based-on", "v1"}, s, st)
		assertLockConflict(t, err)

		got, err := s.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if got.Revision != 2 || len(got.Artifacts) != 0 || got.Tasks[0].Stage != "intake" {
			t.Fatalf("stale submit changed state: %+v", got)
		}
	})

	t.Run("approve", func(t *testing.T) {
		s := newTestStore(t)
		st := store.State{
			ProjectID: "proj1",
			Revision:  1,
			Tasks: []store.Task{{
				ID: "task-1", Title: "task", Stage: "proposal", Status: "active",
				CreatedAt: "2026-07-04T00:00:00Z", UpdatedAt: "2026-07-04T00:00:00Z",
			}},
			Artifacts: []store.Artifact{{
				ID: "art-1", TaskID: "task-1", Stage: "proposal", Kind: "proposal",
				Version: 1, Status: "active", BlobHash: "abc123", BasedOnRevision: 1,
				CreatedAt: "2026-07-04T00:00:00Z",
			}},
		}
		if err := s.Save(st); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		advanceRevision(t, s, 1)

		err := CmdApprove([]string{"art-1", "--based-on", "v1"}, s, st)
		assertLockConflict(t, err)

		got, err := s.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if got.Revision != 2 || got.Artifacts[0].Status != "active" {
			t.Fatalf("stale approve changed state: %+v", got)
		}
	})

	t.Run("archive", func(t *testing.T) {
		s := newTestStore(t)
		st := store.State{
			ProjectID: "proj1",
			Revision:  1,
			Tasks: []store.Task{{
				ID: "task-1", Title: "task", Stage: "proposal", Status: "active",
				CreatedAt: "2026-07-04T00:00:00Z", UpdatedAt: "2026-07-04T00:00:00Z",
			}},
		}
		if err := s.Save(st); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		advanceRevision(t, s, 1)

		err := CmdArchive([]string{"--based-on", "v1"}, s, st)
		assertLockConflict(t, err)

		got, err := s.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if got.Revision != 2 || got.Tasks[0].Status != "active" || got.Tasks[0].Stage != "proposal" {
			t.Fatalf("stale archive changed state: %+v", got)
		}
	})
}

func advanceRevision(t *testing.T, s store.Store, revision int) {
	t.Helper()
	if _, err := s.Update(revision, func(*store.State) error { return nil }); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
}

func assertLockConflict(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected lock conflict, got nil")
	}
	var exitErr commandkit.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error = %T %v, want commandkit.ExitError", err, err)
	}
	if exitErr.Code != 12 {
		t.Fatalf("exit code = %d, want 12", exitErr.Code)
	}
}
