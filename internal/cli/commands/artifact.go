package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/store"
)

type artifactBlobCheck struct {
	ArtifactID string `json:"artifact_id"`
	BlobHash   string `json:"blob_hash"`
	Status     string `json:"status"`
	ActualHash string `json:"actual_hash,omitempty"`
}

func CmdArtifactList(args []string, _ store.Store, st store.State) error {
	if commandkit.Has(args, "--json") {
		return commandkit.PrintJSON(st.Artifacts)
	}
	for _, a := range st.Artifacts {
		dependencies := "-"
		if len(a.DependsOn) > 0 {
			dependencies = strings.Join(a.DependsOn, ",")
		}
		fmt.Fprintf(commandkit.Out, "%s %s %s depends-on=%s\n", a.ID, a.Kind, a.Status, dependencies)
	}
	return nil
}

func CmdArtifactVerify(args []string, s store.Store, st store.State) error {
	artifactID := ""
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if artifactID != "" {
			return commandkit.Usage("at most one artifact id may be specified")
		}
		artifactID = arg
	}

	checks := make([]artifactBlobCheck, 0, len(st.Artifacts))
	failures := 0
	for _, artifact := range st.Artifacts {
		if artifactID != "" && artifact.ID != artifactID {
			continue
		}
		check := artifactBlobCheck{ArtifactID: artifact.ID, BlobHash: artifact.BlobHash, Status: "ok"}
		actual, err := s.ValidateBlob(artifact.BlobHash)
		if err != nil {
			var validationErr *store.BlobValidationError
			if !errors.As(err, &validationErr) {
				return err
			}
			check.Status = validationErr.Kind
			check.ActualHash = validationErr.ActualHash
			failures++
		} else {
			check.ActualHash = actual
		}
		checks = append(checks, check)
	}

	if artifactID != "" && len(checks) == 0 {
		return commandkit.ArtifactValidation("artifact not found: " + artifactID)
	}

	if commandkit.Has(args, "--json") {
		if err := commandkit.PrintJSON(checks); err != nil {
			return err
		}
	} else {
		for _, check := range checks {
			if check.Status == "corrupted" {
				fmt.Fprintf(commandkit.Out, "%s %s %s actual=%s\n", check.ArtifactID, check.BlobHash, check.Status, check.ActualHash)
				continue
			}
			fmt.Fprintf(commandkit.Out, "%s %s %s\n", check.ArtifactID, check.BlobHash, check.Status)
		}
	}

	if failures > 0 {
		return commandkit.ArtifactValidation(fmt.Sprintf("%d artifact blob(s) failed validation", failures))
	}
	return nil
}
