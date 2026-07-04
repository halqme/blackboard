
# AGENTS.md

## Project Overview

This repository contains `blackboard`, a multi-agent development workflow runtime exposed through the `bb` CLI.

`bb` is not an autonomous agent and is not an agent process manager. Its role is to maintain shared, durable workflow state for externally executed coding agents.

Agents do not share conversation history directly. They coordinate only through repository state, Git, and bb primitives (tasks, stages, artifacts, decisions, approvals, revisions, and optional historical recall providers).

The CLI executable name is `bb`. The formal project name is `blackboard`.

## Prerequisites

- **`just`** — command runner. See `just --list` for available commands.
- **`bb`** — built with `just build`.

## Development Commands

| Command       | Description                     |
|---------------|---------------------------------|
| `just build`  | Build the `bb` CLI              |
| `just test`   | Run all tests                   |
| `just lint`   | Run `go vet`                    |
| `just fmt`    | Format Go source code            |
| `just check`  | Run lint + test (verification)   |

## Required Agent Workflow

At the beginning of any task:

1. **Check current state**: `bb status`
2. **Read working context**: `bb context`
3. **Determine valid next action**: `bb next`

When working within a specific stage:

4. **Inspect stage state**: `bb stage <stage>`

Do not infer workflow state from previous conversation context. Always use the current blackboard state.

### Submitting work

```sh
bb submit <kind> --file <path> --based-on <revision-or-artifact-id>
```

The `--based-on` argument is mandatory. All writes use Compare-and-Swap semantics.

### Approval

```sh
bb approve <artifact-id> --based-on <revision>
```

### Archive

```sh
bb archive --based-on <revision>
```

## Verification

Before claiming work is complete, run:

```sh
just check
```

Do not assert success without evidence (test output, lint results, etc.). Investigate failures systematically — no blind retries or guesses.

## Git

Commit per logical unit of work. Write clear, descriptive commit messages. Do not commit with failing tests. Prefer small, focused commits over large monolithic ones.

## Detail References

- **システム仕様**: `docs/specification.md` — アーキテクチャ原則、ステージモデル、CAS、終了コードなど
- **操作手順**: `docs/operations.md` — コンテキスト取得、レビュー、エージェント登録、ワークツリーなど。実際の操作感の定義
- **Blackboard CLI help**: `bb help`

## Blackboard

This repository uses blackboard workflow.

At task start:

1. Run `bb status`
2. Run `bb context`
3. Run `bb next`

If you work within a stage, run `bb stage <stage>`.
Use `bb help` when command usage is unclear.

