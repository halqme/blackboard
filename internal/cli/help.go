package cli

import (
	"fmt"
	"strings"

	"github.com/halqme/blackboard/internal/cli/commandkit"
)

func cmdHelp(args []string) error {
	if len(args) == 0 {
		fmt.Fprint(commandkit.Out, renderGeneralHelp())
		return nil
	}

	command, path := resolveCommandPath(rootCommands(), args)
	if command == nil {
		return commandkit.Usage(unknownCommandMessage(args[0]))
	}

	fmt.Fprint(commandkit.Out, renderCommandHelp(*command, path))
	return nil
}

func renderGeneralHelp() string {
	var b strings.Builder
	b.WriteString("bb - blackboard multi-agent workflow runtime\n\n")
	b.WriteString("Usage:\n")
	b.WriteString("  bb <command> [options]\n\n")
	b.WriteString("Commands:\n")
	for _, command := range rootCommands() {
		b.WriteString(fmt.Sprintf("  %-30s %s\n", compactUsage(command), command.Abstract))
	}
	b.WriteString("  help, --help, -h              Show usage information\n\n")
	b.WriteString("Stages:\n")
	b.WriteString("  intake context proposal critique decision\n")
	b.WriteString("  implementation review verification handoff archived\n\n")
	b.WriteString("Exit codes:\n")
	b.WriteString("    0  success\n")
	b.WriteString("    1  general error\n")
	b.WriteString("    2  invalid usage\n")
	b.WriteString("    3  config error\n")
	b.WriteString("    4  task not found\n")
	b.WriteString("    5  invalid state transition\n")
	b.WriteString("   11  artifact validation error\n")
	b.WriteString("   12  lock conflict (CAS failure)\n\n")
	b.WriteString("All writes require --based-on <revision> (Compare-and-Swap).\n")
	b.WriteString("For command details, run 'bb help <command>'.\n")
	b.WriteString("For agents and automation, prefer explicit init flags over interactive prompts.\n")
	return b.String()
}

func renderCommandHelp(command Command, path []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("bb %s\n\n", strings.Join(path, " ")))
	b.WriteString("Usage:\n")
	b.WriteString("  " + command.Usage + "\n\n")
	b.WriteString(command.Abstract + "\n")
	if len(command.Children) > 0 {
		b.WriteString("\nSubcommands:\n")
		for _, child := range command.Children {
			b.WriteString(fmt.Sprintf("  %-30s %s\n", child.Name, child.Abstract))
		}
		b.WriteString(fmt.Sprintf("\nTry 'bb help %s <subcommand>' for details.\n", strings.Join(path, " ")))
	}
	return b.String()
}

func compactUsage(command Command) string {
	return strings.TrimPrefix(command.Usage, "bb ")
}

func resolveCommandPath(commandsList []Command, args []string) (*Command, []string) {
	var path []string
	current := commandsList
	var command *Command
	for _, arg := range args {
		next := findCommand(current, arg)
		if next == nil {
			break
		}
		command = next
		path = append(path, next.Name)
		current = next.Children
	}
	return command, path
}

func findCommand(commands []Command, name string) *Command {
	for i := range commands {
		if commands[i].Name == name {
			return &commands[i]
		}
	}
	return nil
}

func unknownCommandMessage(name string) string {
	return fmt.Sprintf("unknown command: %s\n\nNext actions:\n  - Try 'bb help' to list available commands.\n  - Try 'bb help <command>' for command-specific usage.", name)
}

func commandUsage(path []string, problem string) error {
	next := fmt.Sprintf("Try 'bb help %s' for command usage.", strings.Join(path, " "))
	return commandkit.Usage(problem + "\n\nNext action:\n  - " + next)
}
