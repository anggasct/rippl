# AGENTS.md

> Copy to the **code repository root**. Points implementers at project documentation — do not duplicate delivery process here.

## Agent quick start (one capability)

1. Read [project-docs/vibecoding-playbook.md](project-docs/vibecoding-playbook.md) — read order, guardrails, stop conditions, handoff format.
2. Open [project-docs/development-plan.md](project-docs/development-plan.md#delivery-queue) — pick **Priority 1** row that is unblocked, unclaimed, and has a ready spec.
3. Read the linked **CAP** and **spec** — plus `Repo Context To Inspect` and shipped refs in `api/`, `flows/`, `app-screen-mapping` for touched areas. If the spec has Implementation order, read the active step only.
4. **Claim** the queue row before editing code: set `Owner` = `Cursor`, `State` = `active`, `Mode` = `direct`, `Active task` = branch or session note.
5. Implement the **claimed capability** per spec acceptance criteria — scoped diff only; stop and ask if requirements are unclear (see playbook stop conditions).
6. Run spec verification commands; report using the [handoff format](project-docs/vibecoding-playbook.md#handoff-format).
7. After operator accept: update reference docs if contracts changed; release claim; update queue status in `development-plan.md` only.

## Project documentation

All delivery process, specs, and status live in `project-docs/` (symlink to Obsidian vault).

**Start here:**

0. [project-docs/bootstrap.md](project-docs/bootstrap.md) — first-time setup (skip once bootstrapped)
1. [project-docs/README.md](project-docs/README.md) — delivery process and CAP domain blocks
2. [project-docs/vibecoding-playbook.md](project-docs/vibecoding-playbook.md) — **read before implementing**
3. [project-docs/development-plan.md](project-docs/development-plan.md) — **canonical** status and delivery queue
4. [project-docs/development-flow-walkthrough.md](project-docs/development-flow-walkthrough.md) — extract → build loop

## Rules

- Implement from **spec** only — not from [project-docs/prd.md](project-docs/prd.md) directly.
- **Claim before code** — set `Owner`, `State` = `active`, `Mode`, and `Active task` on the queue row.
- If the spec has **Implementation order**, implement only the current step unless the spec allows batching.
- Status (✅ / 🔶 / ⬜) and build order live **only** in `development-plan.md`.
- Rippl is **local-only** — do not add network calls, telemetry, or cloud dependencies in MVP without ADR.
- If graph semantics, cache format, exit codes, or JSON schema is unclear: **stop** and ask — do not invent behavior.
- After shipping: update `project-docs/api/`, `project-docs/flows/`, and `project-docs/app-screen-mapping.md` when CLI contracts change.

## Conventions

- [project-docs/code-conventions.md](project-docs/code-conventions.md)
- [project-docs/go-conventions.md](project-docs/go-conventions.md)
- [project-docs/architecture.md](project-docs/architecture.md)

## Git

- **Branch:** `{type}/{short-slug}` — [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) types; no CAP IDs in branch names (default: `feat/cli-config`).
- **Commits to `main`:** use user-facing scopes in the subject for a readable public changelog — e.g. `feat(analyze): add JSON export`, `fix(graph): filter packages`. Put the CAP ID in the **PR body** or commit body, not in the scope (`cap-xxx` scopes appear in [CHANGELOG.md](CHANGELOG.md)).
- **Agent handoff commits** may still reference CAP in the body; avoid `feat(cap-N):` on merge commits that feed release-please.
- Details: [project-docs/vibecoding-playbook.md](project-docs/vibecoding-playbook.md#git-conventions), [CONTRIBUTING.md](CONTRIBUTING.md), [project-docs/execution-context.md](project-docs/execution-context.md#branch-and-worktree-policy).

## Verification

```bash
go test ./...
golangci-lint run ./...
```

Run commands listed in the active spec before handoff.

## Cursor integration

- **This file (`AGENTS.md`)** — repo-wide entry point for agents.
- **`.cursor/rules/`** — optional; must not contradict `project-docs/` or duplicate status.
