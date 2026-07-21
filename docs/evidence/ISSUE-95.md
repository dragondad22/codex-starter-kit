# Issue #95 — organic task decomposition evidence

**Date:** 2026-07-21

**Issue:** [#95](https://github.com/dragondad22/codex-starter-kit/issues/95)

## Delivered contract and owner correction

The first implementation incorrectly interpreted planning decomposition as mandatory
native-issue decomposition for every independently completable unit. The owner corrected
that interpretation after PR #97, and #95 was reopened.

- A Ready task has an actionable outcome and sufficient context to begin; it need not be
  reduced to one independently completable implementation unit.
- Implementation decomposes organically into the tasks, subtasks, and steps warranted by
  the actual work. There is no prescribed child count or decomposition depth.
- Native child issues are created when durable independent tracking adds value, including
  distinct ownership, scheduling, dependencies, authority, review/evidence boundaries,
  release value, or handoff needs. Independent completable-ness alone is insufficient.
- Required freshness and fitness checks before implementation planning and after material
  change. Structured plans coordinate work; runtimes without that facility use an
  agent-neutral ordered checklist with exactly one active step.
- Made both plan forms subordinate to issue authority: neither can expand issue scope or
  authority.
- Preserved one writer per mutable boundary while allowing read-only or isolated
  independent review.

## Intake and template behavior

The task form asks for an actionable outcome and contextual decomposition. Initiative and
bug forms use the same durable-tracking criteria without requiring child issues merely
because remediation steps could be completed independently.

These additions are descriptions and Markdown guidance only. Existing issue-form field
IDs, headings, parser contracts, and lifecycle-engine schemas remain unchanged.

## Deterministic validation and limitation

Documentation validation now requires the canonical task-fitness, decomposition,
agent-neutral planning, and writer-boundary markers across the repository rules,
issue-tracker contract, lifecycle guide, and issue forms. Unit tests cover acceptance of
the complete contract and rejection when planning/decomposition guidance is absent.

The validator protects durable policy presence; it does not prescribe task granularity or
decide which implementation units deserve native issues. That contextual judgment remains
part of planning and tracker stewardship.

## Verification state

The exact verification results and independent review state are recorded on the pull
request that completes #95. Required local gates are the Python test suite, documentation
validation, Go tests, structured change check, and whitespace validation.

## Reconciliation and handoff

The rejected rigid rule is not a reason to split #75. After this correction, #75 returns
to its prior planning flow and parent #4 remains in progress.
