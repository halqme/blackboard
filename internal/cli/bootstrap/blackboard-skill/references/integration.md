# Blackboard Integration Reference

Use this file when the task is to roll blackboard out to other repositories or to improve bootstrap reliability.

## Discovery model

There are three layers, and each has a different job:

1. Global skill
   Teaches the agent how to bring blackboard into a repository.
2. Agent-side hook
   Nudges the agent to check whether bootstrap is needed.
3. Repo-local `AGENTS.md`
   Carry the durable workflow rules for that repository.

If you skip repo-local files, the setup gets brittle fast.

## Recommended hook message

Keep the hook short. Example intent:

```txt
This repository may need blackboard bootstrap or blackboard workflow review.
Use `$blackboard`, then check `bb guide` and `bb help`.
```

The hook should remind, not automate workflow writes.

## Repo-local bootstrap

Prefer letting `bb init --with-agent-files` write the first draft.

Then review and adjust it for the repository.

## Repo-local AGENTS.md snippet

Use a short snippet rather than a giant policy dump:

```md
## Blackboard

This repository uses blackboard workflow.

At task start:

1. Run `bb status`
2. Run `bb context`
3. Run `bb next`

If you work within a stage, run `bb stage <stage>`.
Use `bb help` when command usage is unclear.
```

## Failure handling

- `blackboard.yaml not found`
  Means the repo is not initialized for blackboard, or you are in the wrong directory. Run `bb guide` before guessing.
- usage error or missing command shape
  Run `bb help`.

## Important distinction

The skill is for rollout. The repo-local files and `bb` output are for operation.
