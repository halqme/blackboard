package context

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/halqme/blackboard/internal/config"
	"github.com/halqme/blackboard/internal/store"
)

func Render(root string, cfg config.Config, st store.State, current *store.Task) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Blackboard Context\n\n")
	fmt.Fprintf(&b, "Current revision: v%d\n", st.Revision)
	fmt.Fprintf(&b, "Project: %s (%s)\n\n", cfg.Project.Name, cfg.Project.ID)
	if current != nil {
		fmt.Fprintf(&b, "## Current Task\n%s: %s\nStage: %s\nStatus: %s\n\n", current.ID, current.Title, current.Stage, current.Status)
	}
	fmt.Fprintf(&b, "## Active Artifacts\n")
	found := false
	for _, a := range st.Artifacts {
		if current != nil && a.TaskID == current.ID && (a.Status == "active" || a.Status == "approved") {
			fmt.Fprintf(&b, "- %s %s.v%d [%s]\n", a.ID, a.Kind, a.Version, a.Status)
			found = true
		}
	}
	if !found {
		fmt.Fprintf(&b, "- none\n")
	}
	fmt.Fprintf(&b, "\n## Git Status\n")
	out, err := exec.Command("git", "-C", root, "status", "--short").CombinedOutput()
	if err != nil {
		fmt.Fprintf(&b, "git status unavailable: %v\n", err)
	} else if strings.TrimSpace(string(out)) == "" {
		fmt.Fprintf(&b, "clean\n")
	} else {
		b.Write(out)
	}
	return b.String()
}

func StageInstruction(stage string) string {
	switch stage {
	case "implementation":
		return "Stage: implementation\nGoal: implement the active proposal or approved decision. Submit with: bb submit implementation --file <path> --based-on <revision>.\n"
	case "review":
		return "Stage: review\nGoal: review correctness, scope, maintainability, regressions, and tests. Submit with: bb submit review --file <path> --based-on <revision>.\n"
	case "proposal":
		return "Stage: proposal\nGoal: propose a concrete implementation/design plan. Submit with: bb submit proposal --file <path> --based-on <revision>.\n"
	case "decision":
		return "Stage: decision\nGoal: choose a proposal and record rationale. Submit with: bb submit decision --file <path> --based-on <revision>.\n"
	default:
		return fmt.Sprintf("Stage: %s\nSubmit with: bb submit %s --file <path> --based-on <revision>.\n", stage, stage)
	}
}
