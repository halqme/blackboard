package cli

import (
	cmds "github.com/halqme/blackboard/internal/cli/commands"
	"github.com/halqme/blackboard/internal/config"
	"github.com/halqme/blackboard/internal/store"
)

type Command struct {
	Name     string
	Usage    string
	Abstract string
	Run      func(args []string) error
	Children []Command
}

func rootCommands() []Command {
	return []Command{
		initCommand(),
		guideCommand(),
		versionCommand(),
		statusCommand(),
		taskCommand(),
		nextCommand(),
		contextCommand(),
		stageCommand(),
		submitCommand(),
		artifactCommand(),
		approveCommand(),
		archiveCommand(),
	}
}

func initCommand() Command {
	return Command{Name: "init", Usage: "bb init [--with-agent-files|--no-agent-files]", Abstract: "Initialize a new blackboard project", Run: cmds.CmdInit}
}

func guideCommand() Command {
	return Command{Name: "guide", Usage: "bb guide", Abstract: "Show blackboard rollout guidance for repos and agents", Run: cmds.CmdGuide}
}

func statusCommand() Command {
	return Command{Name: "status", Usage: "bb status [--json]", Abstract: "Show current workflow state", Run: func(args []string) error { return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error { return cmds.CmdStatus(args, st) }) }}
}

func taskCommand() Command {
	return Command{Name: "task", Usage: "bb task <subcommand>", Abstract: "Task operations; use 'bb task new' to add a task before other work", Children: []Command{taskNewCommand()}}
}

func taskNewCommand() Command {
	return Command{Name: "new", Usage: "bb task new <title>", Abstract: "Create a new task", Run: func(args []string) error { return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error { return cmds.CmdTaskNew(args, s, st) }) }}
}

func nextCommand() Command {
	return Command{Name: "next", Usage: "bb next [--json]", Abstract: "Show the next valid stage or action", Run: func(args []string) error { return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error { return cmds.CmdNext(args, st) }) }}
}

func contextCommand() Command {
	return Command{Name: "context", Usage: "bb context [--json]", Abstract: "Show working context (project, task, artifacts)", Run: func(args []string) error { return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error { return cmds.CmdContext(args, root, cfg, st) }) }}
}

func stageCommand() Command {
	return Command{Name: "stage", Usage: "bb stage <stage>", Abstract: "Show stage-specific instructions", Run: cmds.CmdStage}
}

func submitCommand() Command {
	return Command{Name: "submit", Usage: "bb submit <kind> --file <path> --based-on <revision>", Abstract: "Submit an artifact (proposal, impl, etc.)", Run: func(args []string) error { return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error { return cmds.CmdSubmit(args, s, st) }) }}
}

func artifactCommand() Command {
	return Command{Name: "artifact", Usage: "bb artifact <subcommand>", Abstract: "Artifact operations", Children: []Command{artifactListCommand()}}
}

func artifactListCommand() Command {
	return Command{Name: "list", Usage: "bb artifact list [--json]", Abstract: "List all artifacts", Run: func(args []string) error { return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error { return cmds.CmdArtifactList(args, s, st) }) }}
}

func approveCommand() Command {
	return Command{Name: "approve", Usage: "bb approve <artifact-id> --based-on <revision>", Abstract: "Approve an artifact", Run: func(args []string) error { return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error { return cmds.CmdApprove(args, s, st) }) }}
}

func archiveCommand() Command {
	return Command{Name: "archive", Usage: "bb archive --based-on <revision>", Abstract: "Archive the current task", Run: func(args []string) error { return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error { return cmds.CmdArchive(args, s, st) }) }}
}
