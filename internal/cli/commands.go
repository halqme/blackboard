package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/halqme/blackboard/internal/config"
	bbctx "github.com/halqme/blackboard/internal/context"
	"github.com/halqme/blackboard/internal/fsm"
	"github.com/halqme/blackboard/internal/store"
)

type commandSpec struct {
	Name        string
	Summary     string
	Usage       []string
	Help        string
	Run         func(args []string) error
	Subcommands []commandSpec
}

var commandCatalog = []commandSpec{
	{
		Name:    "init",
		Summary: "Initialize a new blackboard project",
		Usage: []string{
			"bb init [--with-agent-files|--no-agent-files]",
		},
		Help: "Create blackboard.yaml and initialize the project store.",
		Run:  cmdInit,
	},
	{
		Name:    "guide",
		Summary: "Show blackboard rollout guidance for repos and agents",
		Usage: []string{
			"bb guide",
		},
		Help: "Explain how to roll blackboard into a repository and agent setup.",
		Run:  cmdGuide,
	},
	{
		Name:    "status",
		Summary: "Show current workflow state",
		Usage: []string{
			"bb status [--json]",
		},
		Help: "Show the current revision, active task, and stage.",
		Run: func(args []string) error {
			return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error {
				return cmdStatus(args, st)
			})
		},
	},
	{
		Name:    "task",
		Summary: "Task operations",
		Usage: []string{
			"bb task <subcommand>",
		},
		Help: "Operate on tasks.",
		Subcommands: []commandSpec{
			{
				Name:    "new",
				Summary: "Create a new task",
				Usage: []string{
					"bb task new <title>",
				},
				Help: "Create a new task and pause any currently active task.",
				Run: func(args []string) error {
					return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error {
						return cmdTaskNew(args, s, st)
					})
				},
			},
		},
	},
	{
		Name:    "next",
		Summary: "Show the next valid stage or action",
		Usage: []string{
			"bb next [--json]",
		},
		Help: "Show the current stage and the next valid stage transition.",
		Run: func(args []string) error {
			return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error {
				return cmdNext(args, st)
			})
		},
	},
	{
		Name:    "context",
		Summary: "Show working context (project, task, artifacts)",
		Usage: []string{
			"bb context [--json]",
		},
		Help: "Render the current project context for a human or agent.",
		Run: func(args []string) error {
			return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error {
				return cmdContext(args, root, cfg, st)
			})
		},
	},
	{
		Name:    "stage",
		Summary: "Show stage-specific instructions",
		Usage: []string{
			"bb stage <stage>",
		},
		Help: "Show instructions for one workflow stage.",
		Run:  cmdStage,
	},
	{
		Name:    "submit",
		Summary: "Submit an artifact (proposal, impl, etc.)",
		Usage: []string{
			"bb submit <kind> --file <path> --based-on <revision>",
		},
		Help: "Submit an artifact and optionally advance the task stage.",
		Run: func(args []string) error {
			return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error {
				return cmdSubmit(args, s, st)
			})
		},
	},
	{
		Name:    "artifact",
		Summary: "Artifact operations",
		Usage: []string{
			"bb artifact <subcommand>",
		},
		Help: "Inspect project artifacts.",
		Subcommands: []commandSpec{
			{
				Name:    "list",
				Summary: "List all artifacts",
				Usage: []string{
					"bb artifact list [--json]",
				},
				Help: "Inspect recorded artifacts for the current project.",
				Run: func(args []string) error {
					return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error {
						return cmdArtifactList(args, s, st)
					})
				},
			},
		},
	},
	{
		Name:    "approve",
		Summary: "Approve an artifact",
		Usage: []string{
			"bb approve <artifact-id> --based-on <revision>",
		},
		Help: "Mark an artifact as approved using CAS semantics.",
		Run: func(args []string) error {
			return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error {
				return cmdApprove(args, s, st)
			})
		},
	},
	{
		Name:    "archive",
		Summary: "Archive the current task",
		Usage: []string{
			"bb archive --based-on <revision>",
		},
		Help: "Archive the active task using CAS semantics.",
		Run: func(args []string) error {
			return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error {
				return cmdArchive(args, s, st)
			})
		},
	},
}

// ── init ────────────────────────────────────────────────────────────────

func cmdInit(args []string) error {
	if _, err := os.Stat("blackboard.yaml"); err == nil {
		return configErr("blackboard.yaml already exists")
	}
	writeAgentFiles, err := shouldWriteAgentFiles(args)
	if err != nil {
		return err
	}
	wd, _ := os.Getwd()
	id := strings.ToLower(strings.ReplaceAll(filepath.Base(wd), " ", "-"))
	content := fmt.Sprintf(`version: 1
project:
  id: %s
  name: %s
context:
  files:
    - README.md
    - AGENTS.md
commands:
  test: ""
  lint: ""
  typecheck: ""
`, id, filepath.Base(wd))
	if err := os.WriteFile("blackboard.yaml", []byte(content), 0o644); err != nil {
		return err
	}
	if writeAgentFiles {
		if err := ensureAgentWorkflowFiles(); err != nil {
			return err
		}
		if err := installAgentSkills(); err != nil {
			return err
		}
	}
	cfg, _ := config.Load(".")
	s := store.New(cfg.Project.ID)
	if err := s.Ensure(); err != nil {
		return err
	}
	fmt.Fprintln(cliOut, "initialized blackboard project:", cfg.Project.ID)
	fmt.Fprintln(cliOut, "store:", s.Root)
	return nil
}

func cmdGuide(_ []string) error {
	fmt.Fprint(cliOut, `Blackboard rollout guide

1. Run 'bb init' inside the target repository.
2. Add or update repo instructions in AGENTS.md.
3. Install the blackboard skill into .agents when agent files are enabled.
4. Teach agents to use the repo workflow through those files, not through hidden prompts.
5. Use --with-agent-files for agents and automation to skip the interactive prompt.
6. Use 'bb help' as the source of truth for command shape.

Files involved:
- blackboard.yaml: repo configuration
- AGENTS.md: repo-local workflow rules
- .agents/: repo-local agent skill install
`)
	return nil
}

// ── status ──────────────────────────────────────────────────────────────

func cmdStatus(args []string, st store.State) error {
	if has(args, "--json") {
		return printJSON(st)
	}
	fmt.Fprintf(cliOut, "Current revision: v%d\n", st.Revision)
	if t := currentTask(st); t != nil {
		fmt.Fprintf(cliOut, "Current task: %s — %s\n", t.ID, t.Title)
		fmt.Fprintf(cliOut, "Stage: %s\n", t.Stage)
		fmt.Fprintf(cliOut, "Status: %s\n", t.Status)
	} else {
		fmt.Fprintln(cliOut, "Current task: none")
	}
	return nil
}

// ── task ────────────────────────────────────────────────────────────────

func cmdTaskNew(args []string, s store.Store, st store.State) error {
	if len(args) < 1 {
		return commandUsage([]string{"task", "new"}, "task title is required")
	}
	t := store.Task{ID: store.NewID("task"), Title: strings.Join(args, " "), Stage: "intake", Status: "active", CreatedAt: store.Now(), UpdatedAt: store.Now()}
	for i := range st.Tasks {
		if st.Tasks[i].Status == "active" {
			st.Tasks[i].Status = "paused"
			st.Tasks[i].UpdatedAt = store.Now()
		}
	}
	st.Tasks = append(st.Tasks, t)
	st.Revision++
	if err := s.Save(st); err != nil {
		return err
	}
	fmt.Fprintf(cliOut, "created task %s at revision v%d\n", t.ID, st.Revision)
	return nil
}

// ── next ────────────────────────────────────────────────────────────────

func cmdNext(args []string, st store.State) error {
	t := currentTask(st)
	if t == nil {
		return taskNotFound("no active task")
	}
	n := fsm.Next(t.Stage)
	if n == "" {
		n = "none"
	}
	if has(args, "--json") {
		return printJSON(map[string]any{"task_id": t.ID, "stage": t.Stage, "next": n, "revision": fmt.Sprintf("v%d", st.Revision)})
	}
	fmt.Fprintf(cliOut, "Current revision: v%d\n", st.Revision)
	fmt.Fprintf(cliOut, "Current stage: %s\n", t.Stage)
	fmt.Fprintf(cliOut, "Next: %s\n", n)
	return nil
}

// ── context ─────────────────────────────────────────────────────────────

func cmdContext(args []string, root string, cfg config.Config, st store.State) error {
	t := currentTask(st)
	if has(args, "--json") {
		return printJSON(map[string]any{"revision": fmt.Sprintf("v%d", st.Revision), "project": cfg.Project, "task": t, "artifacts": activeArtifacts(st, t)})
	}
	fmt.Fprint(cliOut, bbctx.Render(root, cfg, st, t))
	return nil
}

// ── stage ───────────────────────────────────────────────────────────────

func cmdStage(args []string) error {
	if len(args) < 1 {
		return commandUsage([]string{"stage"}, "usage: bb stage <stage>")
	}
	stage := args[0]
	if !fsm.IsStage(stage) {
		return commandUsage([]string{"stage"}, "unknown stage: "+stage)
	}
	fmt.Fprint(cliOut, bbctx.StageInstruction(stage))
	return nil
}

// ── submit ──────────────────────────────────────────────────────────────

func cmdSubmit(args []string, s store.Store, st store.State) error {
	if len(args) < 1 {
		return commandUsage([]string{"submit"}, "usage: bb submit <kind> --file <path> --based-on <revision>")
	}
	kind := args[0]
	file := value(args, "--file")
	based := value(args, "--based-on")
	if file == "" {
		return commandUsage([]string{"submit"}, "--file is required")
	}
	if based == "" {
		return commandUsage([]string{"submit"}, "--based-on is required")
	}
	baseRev, err := store.ParseRevision(based)
	if err != nil {
		return commandUsage([]string{"submit"}, "invalid --based-on revision")
	}
	if baseRev != st.Revision {
		return lockConflict(fmt.Sprintf("lock conflict: current revision is v%d, but artifact is based on v%d", st.Revision, baseRev))
	}
	t := currentTask(st)
	if t == nil {
		return taskNotFound("no active task")
	}
	if kind == "" {
		return artifactValidation("artifact kind is required")
	}
	hash, err := s.PutBlob(file)
	if err != nil {
		return err
	}
	version := nextArtifactVersion(st, t.ID, kind)
	for i := range st.Artifacts {
		if st.Artifacts[i].TaskID == t.ID && st.Artifacts[i].Kind == kind && st.Artifacts[i].Status == "active" {
			st.Artifacts[i].Status = "superseded"
		}
	}
	a := store.Artifact{ID: store.NewID("art"), TaskID: t.ID, Stage: kind, Kind: kind, Version: version, Status: "active", BlobHash: hash, BasedOnRevision: baseRev, CreatedAt: store.Now()}
	st.Artifacts = append(st.Artifacts, a)
	if fsm.IsStage(kind) && fsm.CanTransition(t.Stage, kind) {
		for i := range st.Tasks {
			if st.Tasks[i].ID == t.ID {
				st.Tasks[i].Stage = kind
				st.Tasks[i].UpdatedAt = store.Now()
			}
		}
	} else if fsm.IsStage(kind) && t.Stage == kind {
		// same-stage resubmission allowed
	} else if fsm.IsStage(kind) {
		return invalidTransition(fmt.Sprintf("invalid transition: %s -> %s", t.Stage, kind))
	}
	st.Revision++
	if err := s.Save(st); err != nil {
		return err
	}
	fmt.Fprintf(cliOut, "submitted %s.v%d as %s\n", kind, version, a.ID)
	fmt.Fprintf(cliOut, "Current revision: v%d\n", st.Revision)
	return nil
}

// ── artifact ────────────────────────────────────────────────────────────

func cmdArtifactList(args []string, s store.Store, st store.State) error {
	arts := st.Artifacts
	sort.Slice(arts, func(i, j int) bool { return arts[i].CreatedAt < arts[j].CreatedAt })
	if has(args, "--json") {
		return printJSON(arts)
	}
	if len(arts) == 0 {
		fmt.Fprintln(cliOut, "no artifacts")
		return nil
	}
	for _, a := range arts {
		fmt.Fprintf(cliOut, "%s  %s.v%d  task=%s  status=%s  based-on=v%d\n", a.ID, a.Kind, a.Version, a.TaskID, a.Status, a.BasedOnRevision)
	}
	return nil
}

// ── approve ─────────────────────────────────────────────────────────────

func cmdApprove(args []string, s store.Store, st store.State) error {
	if len(args) < 1 {
		return commandUsage([]string{"approve"}, "usage: bb approve <artifact-id> --based-on <revision>")
	}
	based := value(args, "--based-on")
	if based == "" {
		return commandUsage([]string{"approve"}, "--based-on is required")
	}
	baseRev, err := store.ParseRevision(based)
	if err != nil {
		return commandUsage([]string{"approve"}, "invalid --based-on revision")
	}
	if baseRev != st.Revision {
		return lockConflict(fmt.Sprintf("lock conflict: current revision is v%d, based on v%d", st.Revision, baseRev))
	}
	id := args[0]
	ok := false
	for i := range st.Artifacts {
		if st.Artifacts[i].ID == id {
			st.Artifacts[i].Status = "approved"
			ok = true
		}
	}
	if !ok {
		return artifactValidation("artifact not found: " + id)
	}
	st.Revision++
	if err := s.Save(st); err != nil {
		return err
	}
	fmt.Fprintf(cliOut, "approved %s\nCurrent revision: v%d\n", id, st.Revision)
	return nil
}

// ── archive ─────────────────────────────────────────────────────────────

func cmdArchive(args []string, s store.Store, st store.State) error {
	based := value(args, "--based-on")
	if based == "" {
		return commandUsage([]string{"archive"}, "--based-on is required")
	}
	baseRev, err := store.ParseRevision(based)
	if err != nil {
		return commandUsage([]string{"archive"}, "invalid --based-on revision")
	}
	if baseRev != st.Revision {
		return lockConflict(fmt.Sprintf("lock conflict: current revision is v%d, based on v%d", st.Revision, baseRev))
	}
	t := currentTask(st)
	if t == nil {
		return taskNotFound("no active task")
	}
	if t.Stage == "archived" {
		return invalidTransition("task is already archived")
	}
	for i := range st.Tasks {
		if st.Tasks[i].ID == t.ID {
			st.Tasks[i].Stage = "archived"
			st.Tasks[i].Status = "archived"
			st.Tasks[i].UpdatedAt = store.Now()
		}
	}
	st.Revision++
	if err := s.Save(st); err != nil {
		return err
	}
	fmt.Fprintf(cliOut, "archived %s\nCurrent revision: v%d\n", t.ID, st.Revision)
	return nil
}

// ── shared helpers ──────────────────────────────────────────────────────

func currentTask(st store.State) *store.Task {
	for i := range st.Tasks {
		if st.Tasks[i].Status == "active" {
			return &st.Tasks[i]
		}
	}
	return nil
}

func activeArtifacts(st store.State, t *store.Task) []store.Artifact {
	if t == nil {
		return nil
	}
	var out []store.Artifact
	for _, a := range st.Artifacts {
		if a.TaskID == t.ID && (a.Status == "active" || a.Status == "approved") {
			out = append(out, a)
		}
	}
	return out
}

func nextArtifactVersion(st store.State, taskID, kind string) int {
	max := 0
	for _, a := range st.Artifacts {
		if a.TaskID == taskID && a.Kind == kind && a.Version > max {
			max = a.Version
		}
	}
	return max + 1
}

func shouldWriteAgentFiles(args []string) (bool, error) {
	if has(args, "--with-agent-files") {
		return true, nil
	}
	if has(args, "--no-agent-files") {
		return false, nil
	}
	return promptYesNo("Add blackboard workflow guidance to AGENTS.md?")
}

func ensureAgentWorkflowFiles() error {
	return ensureAgentWorkflowFile("AGENTS.md", agentsWorkflowSnippet)
}

func installAgentSkills() error {
	return writeInstalledSkill(filepath.Join(".agents", "skills", "blackboard"));
}

func ensureAgentWorkflowFile(path, snippet string) error {
	if b, err := os.ReadFile(path); err == nil {
		if strings.Contains(string(b), "## Blackboard") {
			return nil
		}
		content := strings.TrimRight(string(b), "\n") + "\n\n" + snippet + "\n"
		return os.WriteFile(path, []byte(content), 0o644)
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(snippet+"\n"), 0o644)
}

func writeInstalledSkill(root string) error {
	return fs.WalkDir(bootstrapFS, "bootstrap/blackboard-skill", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel("bootstrap/blackboard-skill", path)
		if err != nil {
			return err
		}
		target := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		b, err := bootstrapFS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
}

const agentsWorkflowSnippet = `## Blackboard

This repository uses blackboard workflow.

At task start:

1. Run ` + "`bb status`" + `
2. Run ` + "`bb context`" + `
3. Run ` + "`bb next`" + `

If you work within a stage, run ` + "`bb stage <stage>`" + `.
Use ` + "`bb help`" + ` when command usage is unclear.
`
