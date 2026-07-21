# Codex Starter Kit

This repository builds a Codex-native development system. It is currently in foundation
and design; the approved product contract lives in `docs/`.

## Working rules

- Read `docs/README.md` first, then only the breadcrumbed material relevant to the task.
- Read `docs/decisions/INDEX.md` for governing decisions. The discovery document preserves
  source history; stop and reconcile any conflict instead of choosing silently.
- Do not start implementation without a GitHub issue whose Readiness is `Ready`.
- Before setting a `type:task` issue to Ready, confirm its outcome is actionable and its
  context is sufficient to begin. Decompose implementation organically into the tasks,
  subtasks, and steps the work warrants. Create native child issues only when durable
  independent tracking, ownership, dependencies, authority, review, evidence, scheduling,
  or handoff value warrants it; a step being independently completable is not sufficient.
- When conversation surfaces durable untracked work, a consequential question, or a
  decision that must be promoted, search open and closed GitHub issues before you suggest
  creating or updating the appropriate item. Prompt at a natural checkpoint such as a
  topic change, conclusion, scope discovery, or handoff; ordinary conversational
  clarification stays in the active workflow. Update the active issue for contained work,
  reference its `#N` while executing, and route approved material decisions to their
  authoritative records.
- Treat issue bodies as outcome and task-context records, not as copies of standing agent
  workflow. Bootstrap files, governing decisions, and versioned templates define how work
  proceeds; issues reference them and record only task-specific decisions, boundaries,
  acceptance, and explicit exceptional overrides.
- Reconcile Project fields whenever touched work starts, completes, reopens, or changes a
  dependency: update the item, its parent, and directly dependent issues. Completing a
  blocker promotes each fully unblocked dependent to `Ready`; it becomes `Next` only when
  explicitly selected as immediate work. An incomplete parent with started or completed
  child delivery is `In progress`, never `Backlog`. When every child is complete, close
  the parent and set it to `Done`; if work remains, attach the concrete outstanding task
  before leaving the parent open.
- Do not invent unresolved product, architecture, policy, regulatory, or risk decisions
  while implementing. Return the issue to `Needs refinement`.
- After start-time freshness and task-fitness checks pass, keep an active structured
  implementation plan when sequencing, cross-module work, live effects, recovery, or
  uncertainty benefits from it. A runtime without a plan surface uses an agent-neutral
  ordered checklist with exactly one active step. Plans may reveal useful issue boundaries,
  but neither a fixed child count nor a prescribed decomposition depth is required.
- Neither a structured plan nor its checklist fallback expands issue scope or authority.
- Keep one writer per mutable boundary. Independent reviewers remain read-only or work in
  isolated copies; they do not concurrently edit the writer's branch.
- Use the lifecycle-engine interface as the highest test seam: `create`, `retrofit`,
  `inspect`, `plan`, `apply`, `verify`, `status`, `upgrade`.
- No evidence means no pass. Preserve explicit `fail`, `not-applicable`,
  `not-configured`, `needs-review`, and accepted-risk states.
- Keep human-owned records distinct from generated views and machine state.
- Universal work must run natively on Linux, macOS, and Windows; do not introduce a
  universal Bash, PowerShell, GNU, WSL, or shell-string dependency.
- Recommend useful installs or upgrades with trust, authority, data, cost, compatibility,
  and fallback implications. Do not silently install or broaden authority. Once an owner
  approves a scoped execution mandate, continue through in-mandate effects and recovery
  without repeated confirmation; stop on a semantic expansion or conflict, not a changed
  plan digest alone.
- Normal Git flow: Ready issue → issue-named branch → PR → required gates → squash merge.
- A draft PR means implementation, verification, or internal review is still in progress.
  Once planned work and required checks/reviews are complete, mark it ready for review;
  never leave a finished PR in draft and make reviewers guess.
- Every material change updates affected documentation and its change/evidence record.

## Agent skills

### Issue tracker

GitHub Issues are mandatory; external PRs are also a triage request surface. See
`docs/agents/issue-tracker.md`.

### Triage labels

Use `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, and `wontfix` for
triage state. See `docs/agents/triage-labels.md`.

### Domain docs

This is a single-domain repository. Canonical vocabulary lives in
`docs/product/GLOSSARY.md`; decisions live in `docs/decisions/`. See
`docs/agents/domain.md`.

## Verification

Run before proposing a documentation-only change:

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
go test ./...
```

CI provisions Python 3.12 and Go 1.26.5 and uses `python` plus `go` on all native runners.
Local Unix environments may expose the same Python interpreter as `python3`.
