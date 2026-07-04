package cli

import (
	"fmt"
	"strings"
)

func cmdHelp(args []string) error {
	if len(args) == 0 {
		fmt.Fprint(cliOut, renderGeneralHelp())
		return nil
	}

	spec, path := resolveCommandPath(commandCatalog, args)
	if spec == nil {
		return usage(unknownCommandMessage(args[0]))
	}

	fmt.Fprint(cliOut, renderCommandHelp(*spec, path))
	return nil
}

func renderGeneralHelp() string {
	var b strings.Builder
	b.WriteString("bb - blackboard multi-agent workflow runtime\n\n")
	b.WriteString("Usage:\n")
	b.WriteString("  bb <command> [options]\n\n")
	b.WriteString("Commands:\n")
	for _, spec := range commandCatalog {
		b.WriteString(fmt.Sprintf("  %-30s %s\n", compactUsage(spec), spec.Summary))
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

func renderCommandHelp(spec commandSpec, path []string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("bb %s\n\n", strings.Join(path, " ")))
	b.WriteString("Usage:\n")
	for _, line := range spec.Usage {
		b.WriteString("  " + line + "\n")
	}
	b.WriteString("\n")
	b.WriteString(spec.Help + "\n")
	if len(spec.Subcommands) > 0 {
		b.WriteString("\nSubcommands:\n")
		for _, child := range spec.Subcommands {
			b.WriteString(fmt.Sprintf("  %-30s %s\n", child.Name, child.Summary))
		}
		b.WriteString(fmt.Sprintf("\nTry 'bb help %s <subcommand>' for details.\n", strings.Join(path, " ")))
	}
	return b.String()
}

func compactUsage(spec commandSpec) string {
	if len(spec.Usage) == 0 {
		return spec.Name
	}
	return strings.TrimPrefix(spec.Usage[0], "bb ")
}

func unknownCommandMessage(name string) string {
	return fmt.Sprintf("unknown command: %s\n\nNext actions:\n  - Try 'bb help' to list available commands.\n  - Try 'bb help <command>' for command-specific usage.", name)
}

func commandUsage(path []string, problem string) error {
	next := fmt.Sprintf("Try 'bb help %s' for command usage.", strings.Join(path, " "))
	return usage(problem + "\n\nNext action:\n  - " + next)
}

func resolveCommandPath(specs []commandSpec, args []string) (*commandSpec, []string) {
	var path []string
	current := specs
	var spec *commandSpec
	for _, arg := range args {
		next := findSubcommand(current, arg)
		if next == nil {
			break
		}
		spec = next
		path = append(path, next.Name)
		current = next.Subcommands
	}
	return spec, path
}

func findSubcommand(specs []commandSpec, name string) *commandSpec {
	for i := range specs {
		if specs[i].Name == name {
			return &specs[i]
		}
	}
	return nil
}
