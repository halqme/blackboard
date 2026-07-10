package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/halqme/blackboard/internal/store"
)

func TestValidateSubmitArgsParsesInputs(t *testing.T) {
	kind, file, err := validateSubmitArgs([]string{"proposal", "--file", "x.md", "--based-on", "v1"})
	if err != nil || kind != "proposal" || file != "x.md" {
		t.Fatalf("kind=%q file=%q err=%v", kind, file, err)
	}
}

func TestStoreArtifactAppendsArtifact(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "proposal.txt")
	if err := os.WriteFile(file, []byte("proposal body"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	st := store.State{Revision: 1, Tasks: []store.Task{{ID: "task-1", Stage: "intake", Status: "active"}}}
	art, err := storeArtifact(&st, "proposal", file)
	if err != nil {
		t.Fatalf("storeArtifact() error = %v", err)
	}
	if len(st.Artifacts) != 1 || art.Kind != "proposal" {
		t.Fatalf("state = %+v art = %+v", st, art)
	}
}

func TestAdvanceTaskStageMovesActiveTask(t *testing.T) {
	st := store.State{Tasks: []store.Task{{ID: "task-1", Stage: "intake", Status: "active"}}}
	advanceTaskStage(&st, "proposal")
	if st.Tasks[0].Stage != "proposal" {
		t.Fatalf("stage = %q, want proposal", st.Tasks[0].Stage)
	}
}

func TestCmdSubmitStoresArtifactAndAdvancesTaskStage(t *testing.T) {
	s := newTestStore(t)
	blobPath := filepath.Join(t.TempDir(), "proposal.txt")
	if err := os.WriteFile(blobPath, []byte("proposal body"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	st := seedTestStore(t, s, []store.Task{{ID: "task-1", Title: "task", Stage: "intake", Status: "active", CreatedAt: "2026-07-04T00:00:00Z", UpdatedAt: "2026-07-04T00:00:00Z"}}, nil)
	if err := CmdSubmit([]string{"proposal", "--file", blobPath, "--based-on", "v2"}, s, st); err != nil {
		t.Fatalf("CmdSubmit() error = %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Revision != 3 || got.Tasks[0].Stage != "proposal" || len(got.Artifacts) != 1 {
		t.Fatalf("state = %+v", got)
	}
}

func TestCmdApproveMarksArtifactApproved(t *testing.T) {
	s := newTestStore(t)
	st := seedTestStore(t, s,
		[]store.Task{{ID: "task-1", Title: "task", Stage: "proposal", Status: "active", CreatedAt: "2026-07-04T00:00:00Z", UpdatedAt: "2026-07-04T00:00:00Z"}},
		[]store.Artifact{{ID: "art-1", TaskID: "task-1", Stage: "proposal", Kind: "proposal", Version: 1, Status: "active", BlobHash: "abc123", BasedOnRevision: 1, CreatedAt: "2026-07-04T00:00:00Z"}},
	)
	if err := CmdApprove([]string{"art-1", "--based-on", "v2"}, s, st); err != nil {
		t.Fatalf("CmdApprove() error = %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Artifacts[0].Status != "approved" {
		t.Fatalf("artifact status = %q, want approved", got.Artifacts[0].Status)
	}
}

func TestCmdArchiveMarksTaskArchived(t *testing.T) {
	s := newTestStore(t)
	st := seedTestStore(t, s, []store.Task{{ID: "task-1", Title: "task", Stage: "proposal", Status: "active", CreatedAt: "2026-07-04T00:00:00Z", UpdatedAt: "2026-07-04T00:00:00Z"}}, nil)
	if err := CmdArchive([]string{"--based-on", "v2"}, s, st); err != nil {
		t.Fatalf("CmdArchive() error = %v", err)
	}
	got, err := s.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Tasks[0].Stage != "archived" || got.Tasks[0].Status != "archived" {
		t.Fatalf("task = %+v, want archived", got.Tasks[0])
	}
}
