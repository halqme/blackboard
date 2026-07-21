package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPutVerifiedBlobAndValidateBlob(t *testing.T) {
	s := New("proj1")
	t.Setenv("BLACKBOARD_HOME", t.TempDir())
	s = New("proj1")
	if err := s.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	source := filepath.Join(t.TempDir(), "artifact.txt")
	if err := os.WriteFile(source, []byte("artifact body"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	hash, err := s.PutVerifiedBlob(source)
	if err != nil {
		t.Fatalf("PutVerifiedBlob() error = %v", err)
	}
	if len(hash) != 64 {
		t.Fatalf("hash = %q, want SHA-256", hash)
	}
	actual, err := s.ValidateBlob(hash)
	if err != nil {
		t.Fatalf("ValidateBlob() error = %v", err)
	}
	if actual != hash {
		t.Fatalf("actual hash = %q, want %q", actual, hash)
	}
}

func TestValidateBlobDistinguishesMissingAndCorrupted(t *testing.T) {
	t.Setenv("BLACKBOARD_HOME", t.TempDir())
	s := New("proj1")
	if err := s.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	_, err := s.ValidateBlob("missing")
	assertBlobValidationKind(t, err, BlobStatusMissing)

	source := filepath.Join(t.TempDir(), "artifact.txt")
	if err := os.WriteFile(source, []byte("artifact body"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	hash, err := s.PutVerifiedBlob(source)
	if err != nil {
		t.Fatalf("PutVerifiedBlob() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(s.Root, "blobs", hash), []byte("corrupted"), 0o644); err != nil {
		t.Fatalf("corrupt blob: %v", err)
	}

	_, err = s.ValidateBlob(hash)
	assertBlobValidationKind(t, err, BlobStatusCorrupted)
	if _, err := s.PutVerifiedBlob(source); err == nil {
		t.Fatal("PutVerifiedBlob() accepted a corrupted existing blob")
	}
}

func assertBlobValidationKind(t *testing.T, err error, want string) {
	t.Helper()
	var validationErr *BlobValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %T %v, want BlobValidationError", err, err)
	}
	if validationErr.Kind != want {
		t.Fatalf("validation kind = %q, want %q", validationErr.Kind, want)
	}
}
