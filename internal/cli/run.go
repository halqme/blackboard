package cli

import (
	"github.com/halqme/blackboard/internal/config"
	"github.com/halqme/blackboard/internal/store"
)

func Run(args []string) error {
	if len(args) == 0 {
		return cmdHelp(nil)
	}
	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		return cmdHelp(args[1:])
	}

	spec := findSubcommand(commandCatalog, args[0])
	if spec == nil {
		return usage(unknownCommandMessage(args[0]))
	}
	return runCommand(spec, []string{spec.Name}, args[1:])
}

func runCommand(spec *commandSpec, path, args []string) error {
	if len(spec.Subcommands) > 0 {
		if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
			return cmdHelp(path)
		}
		child := findSubcommand(spec.Subcommands, args[0])
		if child == nil {
			return commandUsage(path, "unknown subcommand: "+args[0])
		}
		return runCommand(child, append(path, child.Name), args[1:])
	}
	if has(args, "--help") || has(args, "-h") {
		return cmdHelp(path)
	}
	if spec.Run == nil {
		return commandUsage(path, "usage: "+spec.Usage[0])
	}
	return spec.Run(args)
}

func withProject(fn func(root string, cfg config.Config, s store.Store, st store.State) error) error {
	root, err := config.FindRoot(".")
	if err != nil {
		return configErr(err.Error())
	}
	cfg, err := config.Load(root)
	if err != nil {
		return configErr(err.Error())
	}
	s := store.New(cfg.Project.ID)
	if err := s.Ensure(); err != nil {
		return err
	}
	st, err := s.Load()
	if err != nil {
		return err
	}
	return fn(root, cfg, s, st)
}
