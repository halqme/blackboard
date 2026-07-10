# blackboard / `bb`

`blackboard` is a small working prototype of the `bb` CLI: a shared workflow runtime for externally executed coding agents.

`bb` is not an autonomous agent or process manager. It stores authoritative workflow state so multiple agents can coordinate through tasks, stages, artifacts, approvals, and revisions without sharing conversation history.

Workflow metadata is stored in SQLite under `BLACKBOARD_HOME` (or the default blackboard data directory). Artifact contents are stored separately as content-addressed blobs.

## Build

```sh
just build
```

## Try it

```sh
export BLACKBOARD_HOME=$PWD/.bb-home
just build
./bb init
./bb guide
./bb task new "Add feature"
./bb status
./bb context
./bb stage proposal

echo '# Proposal' > proposal.md
./bb submit proposal --file proposal.md --based-on v2
./bb artifact list
```

## Implemented commands

- `bb init`
- `bb guide`
- `bb status [--json]`
- `bb task new <title>`
- `bb next [--json]`
- `bb context [--json]`
- `bb stage <stage>`
- `bb submit <kind> --file <path> --based-on <revision>`
- `bb artifact list [--json]`
- `bb approve <artifact-id> --based-on <revision>`
- `bb archive --based-on <revision>`

## Important behavior

- Writes use revision-checked Compare-and-Swap semantics.
- `--based-on` is mandatory for revision-sensitive public writes.
- Revision mismatch exits with code 12 and does not commit stale state.
- Revision checking, mutation, and revision increment are committed atomically in SQLite.
- Artifacts are stored as content-addressed blobs.
- Previous active artifacts of the same kind are marked `superseded`.
- Forward stage skips are allowed for simplified workflows.
- `archived` is terminal.

## Architecture

Keep these boundaries strict:

- Git stores adopted long-term project knowledge and source code.
- `bb` stores authoritative current workflow state.
- Historical recall providers such as `ctx` provide evidence, not authoritative state.

See `docs/specification.md` for the full architecture and state model, and `docs/operations.md` for workflow behavior.

## Not implemented yet

- artifact dependency graph and precise stale propagation
- `decide`, `verify`, `worktree`, and `agent` commands
- Pi Coding Agent extension wrapper
