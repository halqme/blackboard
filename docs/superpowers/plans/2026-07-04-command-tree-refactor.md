# Command Tree Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor the CLI command registry so each command is declared as a reusable `Command` value, parent commands register child command declarations, and help/dispatch are generated from the same recursive command tree.

**Architecture:** Keep the existing `Run(args []string)` public entry point and existing command handler functions. Replace the current giant `commandCatalog` literal with `rootCommands()` plus per-command declaration functions such as `taskCommand()` and `taskNewCommand()`. Intermediate commands with children only display help; leaf commands with `Run` execute.

**Tech Stack:** Go, existing `internal/cli` package, repo-local `just check` verification.

---

## File Structure

- Modify: `internal/cli/commands.go`
  - Owns `Command`, `rootCommands()`, and per-command declaration functions.
  - Keeps existing command handler implementations such as `cmdInit`, `cmdTaskNew`, and `cmdArtifactList`.
- Modify: `internal/cli/run.go`
  - Dispatches recursively through `rootCommands()`.
  - Treats intermediate commands as help nodes.
- Modify: `internal/cli/help.go`
  - Renders `bb help`, parent command help, and leaf command help from `rootCommands()`.
- Modify: `internal/cli/commands_test.go`
  - Adds declaration-shape tests and updates existing command coverage helpers to use `Command`.
- Run: `just check`
  - Final verification.

---

### Task 1: Lock In Command Declaration Shape

**Files:**
- Modify: `internal/cli/commands_test.go`

- [ ] **Step 1: Write the failing tests**

Add tests that describe the desired declaration API and recursive behavior. These tests should fail against the current `commandSpec` plus global `commandCatalog` shape.

```go
func TestRootCommandsAreDeclaredByCommandFunctions(t *testing.T) {
	commands := rootCommands()

	if findCommand(commands, "task") == nil {
		t.Fatalf("rootCommands() missing task command")
	}
	if findCommand(commands, "artifact") == nil {
		t.Fatalf("rootCommands() missing artifact command")
	}
}

func TestParentCommandsRegisterSubcommandDeclarations(t *testing.T) {
	task := taskCommand()
	child := findCommand(task.Children, "new")
	if child == nil {
		t.Fatalf("taskCommand() missing new subcommand")
	}
	if child.Run == nil {
		t.Fatalf("task new command missing Run")
	}
	if child.Usage != "bb task new <title>" {
		t.Fatalf("task new usage = %q, want %q", child.Usage, "bb task new <title>")
	}

	artifact := artifactCommand()
	if findCommand(artifact.Children, "list") == nil {
		t.Fatalf("artifactCommand() missing list subcommand")
	}
}

func TestIntermediateCommandsAreHelpOnly(t *testing.T) {
	for _, command := range rootCommands() {
		assertIntermediateCommandsHaveNoRun(t, []string{command.Name}, command)
	}
}
```

Add these helpers near the existing test helpers:

```go
func findCommand(commands []Command, name string) *Command {
	for i := range commands {
		if commands[i].Name == name {
			return &commands[i]
		}
	}
	return nil
}

func assertIntermediateCommandsHaveNoRun(t *testing.T, path []string, command Command) {
	t.Helper()
	if len(command.Children) > 0 && command.Run != nil {
		t.Fatalf("intermediate command %q must not have Run", strings.Join(path, " "))
	}
	for _, child := range command.Children {
		assertIntermediateCommandsHaveNoRun(t, append(path, child.Name), child)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
go test ./internal/cli -run 'TestRootCommandsAreDeclaredByCommandFunctions|TestParentCommandsRegisterSubcommandDeclarations|TestIntermediateCommandsAreHelpOnly'
```

Expected: build failure or test failure because `rootCommands`, `taskCommand`, `artifactCommand`, and `Command` do not exist yet.

- [ ] **Step 3: Commit is not required yet**

Do not commit after the failing test step. Keep this as the red phase for Task 2.

---

### Task 2: Introduce `Command` Declarations

**Files:**
- Modify: `internal/cli/commands.go`
- Modify: `internal/cli/commands_test.go`

- [ ] **Step 1: Replace `commandSpec` with `Command`**

In `internal/cli/commands.go`, replace:

```go
type commandSpec struct {
	Name        string
	Summary     string
	Usage       []string
	Help        string
	Run         func(args []string) error
	Subcommands []commandSpec
}
```

with:

```go
type Command struct {
	Name     string
	Usage    string
	Abstract string
	Run      func(args []string) error
	Children []Command
}
```

- [ ] **Step 2: Replace the giant catalog with declaration functions**

Remove the `var commandCatalog = []commandSpec{...}` literal and add:

```go
func rootCommands() []Command {
	return []Command{
		initCommand(),
		guideCommand(),
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
```

Then add one declaration function per top-level command. Use the existing command handler functions and `withProject` wrappers directly:

```go
func initCommand() Command {
	return Command{
		Name:     "init",
		Usage:    "bb init [--with-agent-files|--no-agent-files]",
		Abstract: "Initialize a new blackboard project",
		Run:      cmdInit,
	}
}

func guideCommand() Command {
	return Command{
		Name:     "guide",
		Usage:    "bb guide",
		Abstract: "Show blackboard rollout guidance for repos and agents",
		Run:      cmdGuide,
	}
}

func statusCommand() Command {
	return Command{
		Name:     "status",
		Usage:    "bb status [--json]",
		Abstract: "Show current workflow state",
		Run: func(args []string) error {
			return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error {
				return cmdStatus(args, st)
			})
		},
	}
}

func taskCommand() Command {
	return Command{
		Name:     "task",
		Usage:    "bb task <subcommand>",
		Abstract: "Task operations",
		Children: []Command{
			taskNewCommand(),
		},
	}
}

func taskNewCommand() Command {
	return Command{
		Name:     "new",
		Usage:    "bb task new <title>",
		Abstract: "Create a new task",
		Run: func(args []string) error {
			return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error {
				return cmdTaskNew(args, s, st)
			})
		},
	}
}

func artifactCommand() Command {
	return Command{
		Name:     "artifact",
		Usage:    "bb artifact <subcommand>",
		Abstract: "Artifact operations",
		Children: []Command{
			artifactListCommand(),
		},
	}
}

func artifactListCommand() Command {
	return Command{
		Name:     "list",
		Usage:    "bb artifact list [--json]",
		Abstract: "List all artifacts",
		Run: func(args []string) error {
			return withProject(func(root string, cfg config.Config, s store.Store, st store.State) error {
				return cmdArtifactList(args, s, st)
			})
		},
	}
}
```

Implement `nextCommand`, `contextCommand`, `stageCommand`, `submitCommand`, `approveCommand`, and `archiveCommand` in the same style, preserving their current usage and handler behavior.

- [ ] **Step 3: Update test helper type names**

In `internal/cli/commands_test.go`, rename `commandSpec` references in helpers to `Command`, and rename fields:

```go
func assertCommandHelpCoverage(t *testing.T, path []string, command Command) {
	t.Helper()
	if strings.TrimSpace(command.Abstract) == "" {
		t.Fatalf("command %q missing abstract", strings.Join(path, " "))
	}
	if strings.TrimSpace(command.Usage) == "" {
		t.Fatalf("command %q missing usage", strings.Join(path, " "))
	}
	if len(command.Children) == 0 {
		if command.Run == nil {
			t.Fatalf("leaf command %q missing Run", strings.Join(path, " "))
		}
		return
	}
	for _, child := range command.Children {
		assertCommandHelpCoverage(t, append(path, child.Name), child)
	}
}
```

- [ ] **Step 4: Run focused tests**

Run:

```bash
go test ./internal/cli -run 'TestRootCommandsAreDeclaredByCommandFunctions|TestParentCommandsRegisterSubcommandDeclarations|TestIntermediateCommandsAreHelpOnly|TestCommandCatalogLeafCommandsHaveHelp'
```

Expected: tests may still fail because `run.go` and `help.go` still refer to old names. That is acceptable until Task 3.

---

### Task 3: Update Dispatch To Use `rootCommands()`

**Files:**
- Modify: `internal/cli/run.go`
- Modify: `internal/cli/help.go`

- [ ] **Step 1: Update dispatch helper names and fields**

In `internal/cli/run.go`, update the lookup to use `rootCommands()` and `Command.Children`:

```go
func Run(args []string) error {
	if len(args) == 0 {
		return cmdHelp(nil)
	}
	if args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		return cmdHelp(args[1:])
	}

	command := findCommand(rootCommands(), args[0])
	if command == nil {
		return usage(unknownCommandMessage(args[0]))
	}
	return runCommand(command, []string{command.Name}, args[1:])
}

func runCommand(command *Command, path, args []string) error {
	if len(command.Children) > 0 {
		if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
			return cmdHelp(path)
		}
		child := findCommand(command.Children, args[0])
		if child == nil {
			return commandUsage(path, "unknown subcommand: "+args[0])
		}
		return runCommand(child, append(path, child.Name), args[1:])
	}
	if has(args, "--help") || has(args, "-h") {
		return cmdHelp(path)
	}
	if command.Run == nil {
		return commandUsage(path, "usage: "+command.Usage)
	}
	return command.Run(args)
}
```

- [ ] **Step 2: Move `findCommand` into production code**

In `internal/cli/help.go`, replace `findSubcommand` with:

```go
func findCommand(commands []Command, name string) *Command {
	for i := range commands {
		if commands[i].Name == name {
			return &commands[i]
		}
	}
	return nil
}
```

Remove the test-only `findCommand` helper from `commands_test.go` after this production function exists.

- [ ] **Step 3: Run dispatch-focused tests**

Run:

```bash
go test ./internal/cli -run 'TestRunTaskUsesProjectConfigAndStore|TestRunUnknownCommandSuggestsNextAction|TestRunCommandHelpFlagPrintsCommandSpecificUsage|TestRunNestedCommandHelpFlagPrintsLeafUsage'
```

Expected: pass after `run.go` and `findCommand` are updated.

---

### Task 4: Update Help Rendering To Use `Command`

**Files:**
- Modify: `internal/cli/help.go`
- Modify: `internal/cli/commands_test.go`

- [ ] **Step 1: Update general help**

In `internal/cli/help.go`, update `renderGeneralHelp`:

```go
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
```

- [ ] **Step 2: Update command help**

Replace `renderCommandHelp` with:

```go
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
```

- [ ] **Step 3: Update path resolution**

Replace `resolveCommandPath` with:

```go
func resolveCommandPath(commands []Command, args []string) (*Command, []string) {
	var path []string
	current := commands
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
```

Update `cmdHelp` to call `resolveCommandPath(rootCommands(), args)`.

- [ ] **Step 4: Update compact usage**

Replace `compactUsage` with:

```go
func compactUsage(command Command) string {
	return strings.TrimPrefix(command.Usage, "bb ")
}
```

- [ ] **Step 5: Run help-focused tests**

Run:

```bash
go test ./internal/cli -run 'TestRunHelpForCommandPrintsCommandSpecificUsage|TestRunParentHelpListsSubcommands|TestRunNestedHelpPrintsSubcommandUsage|TestCommandCatalogLeafCommandsRespondToHelp'
```

Expected: pass.

---

### Task 5: Clean Up Naming And Full Verification

**Files:**
- Modify: `internal/cli/commands_test.go`
- Run: `internal/cli/*.go`

- [ ] **Step 1: Rename catalog tests**

Rename test functions so they describe the new shape:

```go
func TestCommandTreeNodesHaveHelpMetadata(t *testing.T) {
	for _, command := range rootCommands() {
		assertCommandHelpCoverage(t, []string{command.Name}, command)
	}
}

func TestCommandTreeLeafCommandsRespondToHelp(t *testing.T) {
	for _, command := range rootCommands() {
		assertLeafHelpRuns(t, []string{command.Name}, command)
	}
}
```

- [ ] **Step 2: Remove obsolete names**

Run:

```bash
rg -n 'commandSpec|commandCatalog|Subcommands|Summary|Help|findSubcommand' internal/cli
```

Expected: no matches except unrelated prose if any. If matches remain in Go code, rename them to `Command`, `rootCommands`, `Children`, `Abstract`, or `findCommand` as appropriate.

- [ ] **Step 3: Format**

Run:

```bash
gofmt -w internal/cli/*.go
```

Expected: no output.

- [ ] **Step 4: Run full verification**

Run:

```bash
just check
```

Expected:

```text
go vet ./...
go test ./...
```

All packages pass.

- [ ] **Step 5: Review diff**

Run:

```bash
git diff -- internal/cli/commands.go internal/cli/run.go internal/cli/help.go internal/cli/commands_test.go
```

Expected:
- `Command` declarations live in `commands.go`.
- Parent commands only list children.
- Intermediate commands have no `Run`.
- Help and dispatch both read from `rootCommands()`.
- Tests cover recursive help and declaration metadata.

- [ ] **Step 6: Commit**

Only commit after `just check` passes.

```bash
git add internal/cli/commands.go internal/cli/run.go internal/cli/help.go internal/cli/commands_test.go internal/cli/utils.go docs/superpowers/plans/2026-07-04-command-tree-refactor.md
git commit -m "refactor: declare cli commands as recursive tree"
```

---

## Self-Review

- Spec coverage: Covers the agreed scope: ArgumentParser-like command declarations, parent-child registration, recursive subcommands, generated parent/leaf help, and leaf-only execution.
- Placeholder scan: No TBD/TODO placeholders are present.
- Type consistency: The plan consistently uses `Command`, `Children`, `Abstract`, `Usage`, `Run`, `rootCommands`, and `findCommand`.
- Scope check: This intentionally does not add declarative option parsing. Flags and positional arguments remain handled by existing command handlers.
