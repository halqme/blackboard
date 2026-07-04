---
name: blackboard
description: Mandatory skill for repositories adopting blackboard workflow.
---

# Blackboard Skill

Use this skill when the task is to bring blackboard into a repository, not when the task is merely to operate an already-managed workflow.

This skill should stay thin. Do not turn it into a second CLI manual.

## What this skill does

1. Tell the agent to use `bb guide` first for rollout instructions.
2. Tell the agent to use `bb help` for current command shape.
3. Explain the split between:
   - `bb` as the source of command truth
   - repo-local `AGENTS.md` as durable workflow instructions
   - agent-side hooks as optional nudges

If `bb` and the skill disagree, `bb` wins. The skill is not authoritative.

## Rollout workflow

When introducing blackboard into another repository:

1. Run `bb guide`.
2. Run `bb init`.
3. Prefer `bb init --with-agent-files` for agents and automation.
4. Review the generated or updated `AGENTS.md`. 
5. Keep repo-specific rules in those files, not in this skill.

## Boundaries

Do not hardcode detailed command syntax here beyond the minimum bootstrap path.

Do not copy full help text into this skill.

Do not rely on hidden prompts as the only place that explains blackboard. Repo-local files should carry the operational rules.

## Hooks

If an agent platform supports startup hooks, use them only to remind the agent that the repository may need blackboard bootstrap or blackboard workflow review.

Hooks should not mutate workflow state automatically.

## References

- For integration guidance and hook wording, read [references/integration.md](references/integration.md).
