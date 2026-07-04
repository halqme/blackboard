# blackboard / bb demo

This is a small working prototype of the `bb` CLI.

It intentionally uses a JSON metadata store for the demo so the command behavior is easy to inspect. The package layout keeps the store behind `internal/store`, so replacing it with SQLite is straightforward.

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

- `--based-on` is mandatory for writes.
- Revision mismatch exits with code 12.
- Artifacts are stored as content-addressed blobs.
- Previous active artifacts of the same kind are marked `superseded`.
- Forward stage skips are allowed for simplified workflows.
- `archived` is terminal.

## Not implemented yet

- SQLite backend
- artifact dependency graph and precise stale propagation
- `decide`, `verify`, `worktree`, and `agent` commands
- Pi Coding Agent extension wrapper
