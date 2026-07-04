# Specification

> Blackboard システムの設計仕様。
> アーキテクチャ上の原則、状態管理モデル、CAS、終了コードなどを定義する。

## Core Architectural Principle

Keep the following boundaries strict.

### Git

Git stores adopted long-term project knowledge and source code.

Examples:

```txt
AGENTS.md
CLAUDE.md
README.md
docs/adr/
docs/specs/
source code
tests
```

Do not store temporary agent workflow state in Git.

Do not commit:

```txt
raw agent output
temporary critiques
handoff notes
run logs
unapproved decisions
internal agent state
temporary prompts
conversation transcripts
```

### Blackboard

`bb` stores authoritative workflow state.

Examples:

```txt
Task
Stage
Artifact
Decision
Approval
Agent Record
Worktree Mapping
Run
Project Revision
```

The blackboard is the source of truth for current workflow state.

### Historical Recall

Historical agent sessions are not authoritative blackboard state.

A historical recall provider such as `ctx` may be used to search previous sessions.

Historical recall is evidence, not truth.

Do not treat recalled conversation content as an approved design decision or active workflow artifact.

Relevant findings from historical recall must be summarized and promoted into a normal `bb` artifact before they become part of the active workflow.

---

## Non-Goals

Do not turn `bb` into:

* an agent runner
* a process supervisor
* a scheduler
* an autonomous coding agent
* a replacement for Git
* a chat history database
* a general-purpose agent memory system

Agent processes are created, executed, and terminated externally.

`bb` records their state and outputs only.

---

## Repository State Rules

Never create:

```txt
.blackboard/
```

inside the repository.

Blackboard state must live outside the Git repository.

Default location:

```txt
~/.local/share/blackboard/
```

When set, prefer:

```sh
BLACKBOARD_HOME
```

Project configuration lives in:

```txt
blackboard.yaml
```

`blackboard.yaml` is Git-managed long-term configuration.

It must not contain task-local state.

---

## Revision and CAS Rules

Every blackboard write must be based on an explicit revision or artifact.

For artifact submission:

```sh
bb submit <kind> --file <path> --based-on <revision-or-artifact-id>
```

The `--based-on` argument is mandatory.

Do not bypass it.

Example:

```sh
bb status
# Current revision: v12

bb submit proposal \
  --file proposal.md \
  --based-on v12
```

All writes use Compare-and-Swap semantics.

If the current revision has changed, the write must fail.

---

## Lock Conflict Handling

Exit Code:

```txt
12
```

means:

```txt
revision mismatch
CAS failure
lock conflict
```

When Exit Code 12 occurs:

1. Do not immediately retry.
2. Run `bb status` and `bb context`.
3. Compare the new state with the assumptions used to produce the failed artifact.
4. Determine what changed.
5. Reconsider the work.
6. Regenerate the artifact if necessary.
7. Submit again using the new revision.

Never mechanically resubmit stale output.

A CAS conflict means the assumptions may no longer be valid.

---

## Human Approval Handling

Exit Code:

```txt
9
```

means:

```txt
human approval required
```

Stop the workflow.

Do not continue automatically.

Do not attempt to infer approval.

---

## Stage Model

Standard stages are:

```txt
intake
context
proposal
critique
decision
implementation
review
verification
handoff
archived
```

The normal forward flow is:

```txt
intake
-> context
-> proposal
-> critique
-> decision
-> implementation
-> review
-> verification
-> handoff
-> archived
```

Some workflows may skip stages.

Examples:

```txt
intake
-> implementation
-> review
-> verification
-> archived
```

or:

```txt
intake
-> proposal
-> critique
-> decision
-> archived
```

Do not assume every task uses every stage.

Use:

```sh
bb next
```

to determine the valid next action.

---

## Allowed Backward Transitions

The workflow is a finite state machine, not a one-way pipeline.

Allowed backward transitions include:

```txt
review -> implementation
review -> proposal
review -> decision

verification -> implementation
verification -> review

critique -> proposal

decision -> proposal
decision -> critique
```

Agents may recommend a backward transition.

The transition itself must be managed through `bb`.

Do not silently continue using artifacts invalidated by a backward transition.

---

## Artifact Rules

Artifacts are immutable records.

Never overwrite an existing artifact.

Multiple versions of the same kind may exist.

Examples:

```txt
proposal.v1
proposal.v2
implementation.v1
implementation.v2
```

Artifact statuses include:

```txt
active
superseded
rejected
approved
stale
```

### active

The current usable version.

### superseded

Replaced by a newer version.

### rejected

Explicitly rejected.

### approved

Explicitly accepted for use.

### stale

Its assumptions or dependencies are no longer valid.

---

## Stale Artifact Rules

Never use stale artifacts as inputs for new work.

An artifact becomes stale when:

* one of its dependencies is superseded
* a backward transition invalidates its assumptions
* the workflow has moved to a new version of an earlier artifact

Stale artifacts may be inspected for historical understanding only.

They must not be treated as current requirements or approved decisions.

---

## Artifact Ownership

Agents must not directly modify artifacts created by other agents.

When revising another artifact:

1. Read the existing artifact.
2. Create a new artifact version.
3. Submit it through `bb`.
4. Allow `bb` to manage superseded and stale status.

Do not mutate history.

---

## Exit Codes

Treat exit codes as part of the public interface.

```txt
0   success
1   general error
2   invalid usage
3   config error
4   task not found
5   invalid state transition
6   missing required artifact
7   agent execution failed
8   verification failed
9   human approval required
10  git worktree error
11  artifact validation error
12  lock conflict
```

Do not collapse meaningful exit codes into a generic error.

Adapters and extensions must preserve their semantics.

---

## Implementation Boundaries

The core `bb` CLI owns:

```txt
configuration
project discovery
state persistence
FSM validation
revision handling
CAS
artifact lifecycle
stale propagation
decision state
approval state
worktree metadata
verification records
context rendering
```

Agent integrations own:

```txt
calling bb
rendering tool results
mapping exit codes to agent guidance
exposing commands or tools
```

Agent integrations must not:

```txt
write directly to the database
reimplement FSM validation
reimplement CAS
modify artifact status directly
perform silent retries after Exit Code 12
```

---

## Pi Extension Boundary

The Pi extension must remain thin.

It may expose tools such as:

```txt
blackboard_status
blackboard_next
blackboard_context
blackboard_stage
blackboard_submit
blackboard_artifact_list
blackboard_decide
blackboard_approve
blackboard_verify
blackboard_archive
```

The extension must invoke the `bb` CLI.

It must not become a second blackboard implementation.

---

## Development Priorities

Prefer:

```txt
correctness
clear state transitions
immutable history
explicit failure
deterministic behavior
small interfaces
structured output
testability
```

over:

```txt
clever abstractions
implicit state
automatic recovery
hidden retries
large framework layers
agent-specific coupling
```

The runtime should remain understandable from the CLI, SQLite state, and artifact history.

---

## Testing Requirements

At minimum, test:

```txt
successful CAS write
CAS conflict
missing --based-on
invalid state transition
archived terminal state
artifact version increment
superseded artifact handling
stale propagation
stale artifact rejection
approval flow
verification failure
project discovery
BLACKBOARD_HOME override
absence of repository-local .blackboard state
```

Any change to FSM, artifact lifecycle, or CAS semantics requires tests.

---

## Final Rule

The blackboard must remain the authoritative workflow runtime.

Agents may think independently.

Agents may disagree.

Agents may be replaced.

Conversation history may disappear.

The workflow must still remain recoverable from:

```txt
Git
bb state
artifacts
decisions
revision history
```

Design every feature with that constraint in mind.
