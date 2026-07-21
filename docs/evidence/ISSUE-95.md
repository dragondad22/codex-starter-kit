# Issue #95 — singular actionable task evidence

**Date:** 2026-07-21

**Issue:** [#95](https://github.com/dragondad22/codex-starter-kit/issues/95)

## Delivered contract

- Added a pre-Ready task-fitness gate: one task is one singular, actionable,
  independently completable delivery with one coherent completion transition, review
  boundary, evidence boundary, and Project state.
- Added the operational decomposition test. Work becomes separate native child tasks when
  a section can be delivered, evidenced, and closed independently while leaving a useful,
  truthful outcome, unless transactional correctness, rollout, or recovery requires one
  transition.
- Distinguished delivery boundaries from implementation detail. Multiple files, modules,
  tests, commits, or coordinated atomic steps do not alone require decomposition.
- Required freshness and fitness checks before implementation planning and after material
  change. Structured plans coordinate a fit task; runtimes without that facility use an
  agent-neutral ordered checklist with exactly one active step.
- Made both plan forms subordinate to issue authority: neither can expand issue scope or
  authority.
- Preserved one writer per mutable boundary while allowing read-only or isolated
  independent review.

## Intake and template behavior

The task form now states the fitness test and asks for one coherent scope, ownership,
review, and evidence boundary. Initiative refinement requires a bounded decomposition
outline before executable children become Ready. A coherent bug remains an Intake defect
record and may be promoted in place; independently completable remediations become native
child tasks.

These additions are descriptions and Markdown guidance only. Existing issue-form field
IDs, headings, parser contracts, and lifecycle-engine schemas remain unchanged.

## Deterministic validation and limitation

Documentation validation now requires the canonical task-fitness, decomposition,
agent-neutral planning, and writer-boundary markers across the repository rules,
issue-tracker contract, lifecycle guide, and issue forms. Unit tests cover acceptance of
the complete contract and rejection when planning/decomposition guidance is absent.

The validator protects durable policy presence; it does not claim to decide semantic task
singularity. That judgment uses current acceptance, authority, dependencies, implemented
state, review boundaries, and evidence boundaries during refinement and freshness review.

## Verification state

The exact verification results and independent review state are recorded on the pull
request that completes #95. Required local gates are the Python test suite, documentation
validation, Go tests, structured change check, and whitespace validation.

## Reconciliation and handoff

Issue #95 is an independently completable correction related to #74, not a new native
blocker for #75. Completing it leaves parent #4 in progress and allows the selected
feature sequence to continue with #75 and then #76.
