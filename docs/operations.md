# Operations

> エージェントが blackboard を操作する際の手順と規律。
> コンテキストの取得、アーティファクトの提出、レビュー、エージェント登録、ワークツリーなどの実務的な使い方を定義する。

## Context Rules

Use:

```sh
bb context
```

for authoritative working context.

The rendered context may include:

```txt
Project context
Current task
Current stage
Relevant active artifacts
Approved artifacts
Git status
Git diff
Verification results
Current revision
```

Do not rely on hidden context from another agent.

Do not assume another agent's conversation is visible.

---

## Scope Discipline

Agents must remain within their assigned task and stage.

Do not:

* make unrelated refactors
* modify unrelated files
* expand project scope without explicit justification
* rewrite another agent's work outside the current task
* bypass an approved decision without explaining why

When a task reveals unrelated problems, record them separately instead of silently fixing them.

---

## Decision Discipline

A decision artifact is authoritative only when valid and active or approved.

When a workflow includes a decision stage:

* implementation must follow the valid decision
* implementation must not silently contradict the decision
* a conflicting implementation requires a backward transition
* ignoring an approved decision requires an explicit reason

Use:

```sh
bb decide
```

for creating or recording a decision.

Use:

```sh
bb approve
```

for approving an artifact or decision.

Do not confuse decision creation with approval.

---

## Review Discipline

Review must evaluate the implementation against:

* the current task
* active proposal artifacts
* approved decision artifacts
* project constraints
* repository state
* tests and verification commands

Review should examine:

```txt
correctness
scope control
regression risk
maintainability
security-sensitive behavior
test coverage
```

A review must not approve known serious defects.

When serious problems are found, recommend the appropriate backward transition.

Possible outcomes include:

```txt
approve
return to implementation
return to proposal
return to decision
require human approval
```

---

## Verification Discipline

Verification must use the project commands defined in `blackboard.yaml` or the `.just` file.

Typical commands may include:

```txt
just test
just lint
just check
```

Do not invent verification results.

Do not claim a command passed unless it was actually executed successfully.

If verification cannot be performed, state exactly why.

Verification failure must not be silently ignored.

---

## Historical Recall

Historical recall is optional.

When configured, use:

```sh
bb recall "<query>"
```

for questions such as:

```txt
Why was this design rejected previously?
Did we already try this implementation?
Which previous session modified this file?
What caused the earlier CAS failure?
```

Historical recall internally uses `ctx` to retrieve session history.

Do not treat recall results as authoritative.

The correct flow is:

```txt
historical session
-> recall
-> analysis
-> new bb artifact
-> authoritative workflow input
```

Do not copy raw session history directly into the current workflow unless necessary.

Prefer concise, evidence-based summaries.

---

## Agent State

Agents are externally executed.

`bb` may record:

```txt
agent identity
task
stage
status
parent agent
generated artifacts
```

Useful commands include:

```sh
bb agent list
bb agent show <agent-id>
bb agent tree
bb agent sync
```

Agents must not inspect another agent's private conversation state.

Agents may inspect only shared artifacts and state exposed through `bb`.

---

## Agent Registration

When agent tracking is active, register the agent:

```sh
bb agent register <agent-name> --task <task-id>
```

Update stage or status when appropriate:

```sh
bb agent update <agent-id> --stage review
bb agent update <agent-id> --status running
```

On success:

```sh
bb agent complete <agent-id>
```

On failure:

```sh
bb agent fail <agent-id> --reason "<reason>"
```

Agent registration is metadata only.

It must not imply that `bb` controls the agent process.

---

## Worktrees

Worktrees are for source changes only.

They are not the source of truth for workflow state.

Do not store blackboard metadata in a worktree.

Useful commands may include:

```sh
bb worktree create
bb worktree adopt <agent>
```

Keep source changes isolated when multiple agents work concurrently.

Do not assume filesystem state alone describes the current workflow.
