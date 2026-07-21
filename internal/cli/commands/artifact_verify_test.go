package commands

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func TestCmdArtifactVerifyReportsValidBlob(t *testing.T) {
	s := newTestStore(t)
	hash := putTestBlob(t, s, "artifact body")
	st := seedTestStore(t, s, nil, []store.Artifact{{
		ID: "art-1", TaskID: "task-1", Stage: "proposal", Kind: "proposal",
		Version: 1, Status: "active", BlobHash: hash, BasedOnRevision: 1,
		CreatedAt: "2026-07-04T00:00:00Z",
	}})

	out := &bytes.Buffer{}
	restore := setCommandIO(t, strings.NewReader(""), out)
	defer restore()
	if err := CmdArtifactVerify(nil, s, st); err != nil {
		t.Fatalf("CmdArtifactVerify() error = %v", err)
	}
	if !strings.Contains(out.String(), "art-1 "+hash+" ok") {
		t.Fatalf("output = %q, want valid blob", out.String())
	}
}

func TestCmdArtifactVerifyReturnsExitCode11ForMissingBlob(t *testing.T) {
	s := newTestStore(t)
	hash := strings.Repeat("a", 64)
	st := seedTestStore(t, s, nil, []store.Artifact{{
		ID: "art-1", TaskID: "task-1", Stage: "proposal", Kind: "proposal",
		Version: 1, Status: "active", BlobHash: hash, BasedOnRevision: 1,
		CreatedAt: "2026-07-04T00:00:00Z",
	}})

	out := &bytes.Buffer{}
	restore := setCommandIO(t, strings.NewReader(""), out)
	defer restore()
	err := CmdArtifactVerify(nil, s, st)
	assertArtifactValidationExit(t, err)
	if !strings.Contains(out.String(), "missing") {
		t.Fatalf("output = %q, want missing status", out.String())
	}
}

func TestCmdArtifactVerifyReportsCorruptedBlobAsJSON(t *testing.T) {
	s := newTestStore(t)
	hash := putTestBlob(t, s, "artifact body")
	if err := os.WriteFile(filepath.Join(s.Root, "blobs", hash), []byte("corrupted"), 0o644); err != nil {
		t.Fatalf("corrupt blob: %v", err)
	}
	st := seedTestStore(t, s, nil, []store.Artifact{{
		ID: "art-1", TaskID: "task-1", Stage: "proposal", Kind: "proposal",
		Version: 1, Status: "active", BlobHash: hash, BasedOnRevision: 1,
		CreatedAt: "2026-07-04T00:00:00Z",
	}})

	out := &bytes.Buffer{}
	restore := setCommandIO(t, strings.NewReader(""), out)
	defer restore()
	err := CmdArtifactVerify([]string{"art-1", "--json"}, s, st)
	assertArtifactValidationExit(t, err)
	if !strings.Contains(out.String(), `"status": "corrupted"`) || !strings.Contains(out.String(), `"actual_hash"`) {
		t.Fatalf("output = %q, want corrupted JSON result", out.String())
	}
}

func putTestBlob(t *testing.T, s store.Store, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "artifact.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	hash, err := s.PutVerifiedBlob(path)
	if err != nil {
		t.Fatalf("PutVerifiedBlob() error = %v", err)
	}
	return hash
}

func assertArtifactValidationExit(t *testing.T, err error) {
	t.Helper()
	var exitErr commandkit.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error = %T %v, want commandkit.ExitError", err, err)
	}
	if exitErr.Code != 11 {
		t.Fatalf("exit code = %d, want 11", exitErr.Code)
	}
}
