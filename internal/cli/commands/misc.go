package commands

import (
	"fmt"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	bbctx "github.com/halqme/blackboard/internal/context"
	"github.com/halqme/blackboard/internal/config"
	"github.com/halqme/blackboard/internal/fsm"
	"github.com/halqme/blackboard/internal/store"
)

func CmdGuide(_ []string) error {
	fmt.Fprint(commandkit.Out, `Blackboard rollout guide

1. Run 'bb init' inside the target repository.
2. If you are Claude and do not consult AGENTS.md or .agents/, check them first.
3. Add or update repo instructions in AGENTS.md.
4. Install the blackboard skill into .agents when agent files are enabled.
5. Add a task first with 'bb task new <title>' before continuing work.
6. Use 'bb help' as the source of truth for command shape.

Files involved:
- blackboard.yaml: repo configuration
- AGENTS.md: repo-local workflow rules
- .agents/: repo-local agent skill install
`)
	return nil
}

func CmdStatus(args []string, st store.State) error {
	t := currentTask(st)
	if commandkit.Has(args, "--json") {
		payload := map[string]any{
			"project_id":       st.ProjectID,
			"revision":         fmt.Sprintf("v%d", st.Revision),
			"task":             t,
			"has_active_task":  t != nil,
		}
		mergeNoTaskGuidance(payload, t == nil)
		return commandkit.PrintJSON(payload)
	}
	fmt.Fprintf(commandkit.Out, "Current revision: v%d\n", st.Revision)
	if t != nil {
		fmt.Fprintf(commandkit.Out, "Current task: %s — %s\n", t.ID, t.Title)
		fmt.Fprintf(commandkit.Out, "Stage: %s\n", t.Stage)
		fmt.Fprintf(commandkit.Out, "Status: %s\n", t.Status)
	} else {
		fmt.Fprintln(commandkit.Out, "Current task: none")
		printNoTaskGuidance()
	}
	return nil
}

func CmdNext(args []string, st store.State) error {
	t := currentTask(st)
	if t == nil {
		if commandkit.Has(args, "--json") {
			return commandkit.PrintJSON(map[string]any{
				"task_id":      nil,
				"stage":        nil,
				"next":         "create-task",
				"next_command": `bb task new "<title>"`,
				"revision":     fmt.Sprintf("v%d", st.Revision),
			})
		}
		fmt.Fprintln(commandkit.Out, "No active task.")
		fmt.Fprintln(commandkit.Out)
		fmt.Fprintln(commandkit.Out, "You should create a task before proceeding.")
		fmt.Fprintln(commandkit.Out, "Next command:")
		fmt.Fprintln(commandkit.Out, `  bb task new "<title>"`)
		return nil
	}
	n := fsm.Next(t.Stage)
	if n == "" {
		n = "none"
	}
	if commandkit.Has(args, "--json") {
		return commandkit.PrintJSON(map[string]any{"task_id": t.ID, "stage": t.Stage, "next": n, "revision": fmt.Sprintf("v%d", st.Revision)})
	}
	fmt.Fprintf(commandkit.Out, "Current stage: %s\nNext: %s\n", t.Stage, n)
	return nil
}

func CmdContext(args []string, root string, cfg config.Config, st store.State) error {
	current := currentTask(st)
	if commandkit.Has(args, "--json") {
		payload := map[string]any{"revision": fmt.Sprintf("v%d", st.Revision), "project": cfg.Project, "task": current, "artifacts": activeArtifacts(st, current), "has_active_task": current != nil}
		mergeNoTaskGuidance(payload, current == nil)
		return commandkit.PrintJSON(payload)
	}
	fmt.Fprint(commandkit.Out, bbctx.Render(root, cfg, st, current))
	if current == nil {
		fmt.Fprintln(commandkit.Out)
		printNoTaskGuidance()
	}
	return nil
}

func CmdStage(args []string) error {
	if len(args) == 0 {
		return commandkit.Usage("stage name required")
	}
	fmt.Fprint(commandkit.Out, bbctx.StageInstruction(args[0]))
	return nil
}

func printNoTaskGuidance() {
	fmt.Fprintln(commandkit.Out)
	fmt.Fprintln(commandkit.Out, "Next action:")
	fmt.Fprintln(commandkit.Out, "  No active task. Create one before making workflow changes.")
	fmt.Fprintln(commandkit.Out, "  Example:")
	fmt.Fprintln(commandkit.Out, `    bb task new "<title>"`)
}

func mergeNoTaskGuidance(payload map[string]any, noTask bool) {
	if !noTask {
		return
	}
	payload["next_action"] = "No active task. Create one before making workflow changes."
	payload["next_command"] = `bb task new "<title>"`
}
