package commands

import (
	"errors"
	"fmt"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

func updateState(s store.Store, expectedRevision int, mutate func(*store.State) error) (store.State, error) {
	st, err := s.Update(expectedRevision, mutate)
	if errors.Is(err, store.ErrRevisionMismatch) {
		return store.State{}, commandkit.LockConflict(fmt.Sprintf("revision mismatch: based on v%d", expectedRevision))
	}
	return st, err
}

func requiredBasedOnRevision(args []string) (int, error) {
	basedOn := commandkit.Value(args, "--based-on")
	if basedOn == "" {
		return 0, commandkit.Usage("--based-on is required")
	}
	revision, err := store.ParseRevision(basedOn)
	if err != nil || revision < 1 {
		return 0, commandkit.Usage("invalid --based-on revision: " + basedOn)
	}
	return revision, nil
}
