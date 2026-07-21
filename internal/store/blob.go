package store

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	BlobStatusMissing   = "missing"
	BlobStatusCorrupted = "corrupted"
)

type BlobValidationError struct {
	Hash       string
	ActualHash string
	Kind       string
}

func (e *BlobValidationError) Error() string {
	switch e.Kind {
	case BlobStatusMissing:
		return fmt.Sprintf("blob %s is missing", e.Hash)
	case BlobStatusCorrupted:
		return fmt.Sprintf("blob %s is corrupted: actual sha256 is %s", e.Hash, e.ActualHash)
	default:
		return fmt.Sprintf("blob %s failed validation", e.Hash)
	}
}

// PutVerifiedBlob stores a blob under its SHA-256 digest. The blob is written to
// a temporary file and linked into place only after all bytes are durable, so a
// reader never observes a partially written content-addressed blob.
func (s Store) PutVerifiedBlob(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	hash := hex.EncodeToString(h.Sum(nil))

	if _, err := s.ValidateBlob(hash); err == nil {
		return hash, nil
	} else if validationErr, ok := err.(*BlobValidationError); !ok || validationErr.Kind != BlobStatusMissing {
		return "", err
	}

	if _, err := f.Seek(0, 0); err != nil {
		return "", err
	}
	blobDir := filepath.Join(s.Root, "blobs")
	if err := os.MkdirAll(blobDir, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(blobDir, ".blob-*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := io.Copy(tmp, f); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	dst := filepath.Join(blobDir, hash)
	if err := os.Link(tmpName, dst); err != nil {
		if os.IsExist(err) {
			if _, validateErr := s.ValidateBlob(hash); validateErr != nil {
				return "", validateErr
			}
			return hash, nil
		}
		return "", err
	}
	return hash, nil
}

// ValidateBlob verifies that the blob exists and that its bytes match the
// SHA-256 digest used as its content-addressed name.
func (s Store) ValidateBlob(hash string) (string, error) {
	f, err := os.Open(filepath.Join(s.Root, "blobs", hash))
	if err != nil {
		if os.IsNotExist(err) {
			return "", &BlobValidationError{Hash: hash, Kind: BlobStatusMissing}
		}
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	actual := hex.EncodeToString(h.Sum(nil))
	if actual != hash {
		return actual, &BlobValidationError{Hash: hash, ActualHash: actual, Kind: BlobStatusCorrupted}
	}
	return actual, nil
}
