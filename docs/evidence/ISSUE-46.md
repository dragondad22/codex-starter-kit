# Issue #46 — governed roadmap Phase evidence

**Date:** 2026-07-18
**Issue:** [#46](https://github.com/dragondad22/codex-starter-kit/issues/46)

## Delivered contract

Phase is a distinct Project planning field with exactly `Phase 0` through `Phase 8`.
Features receive direct assignments. Ordinary children keep their own Phase field blank
and expose parent-derived context through Work Manager. A non-feature direct assignment
is accepted only as explicitly reasoned cross-cutting work; duplicating the parent's value
is rejected.

Work Manager validates the finite vocabulary and complete immutable field/option catalog
before planning. Parent-derived context passes only when the adapter observes the selected
issue's native parent and that parent's immutable Phase option. The GitHub adapter also
rejects renamed, duplicate, wrong-type, missing, extra, or stale Phase catalog state, sets
a direct value by immutable option ID, clears copied child values, and re-reads the
postcondition. Horizon, Status, Readiness, Milestone, and Phase remain independent.

## Live Project result

The approved public sandbox organization Project #1 already contains one Phase field:

- field `PVTSSF_lADOEjyyNM4Bdm9FzhYHTZI`;
- option IDs `7fcb7c26`, `e6cbdc17`, `db48cb41`, `3a97d4af`, `e8eef021`,
  `358327da`, `e3063f78`, `3c19af01`, and `865934cf` for Phase 0–Phase 8.

The operational user Project #8 was inspected before mutation and contained no Phase
field or naming collision. Actor `dragondad22` (account `19365745`, node
`MDQ6VXNlcjE5MzY1NzQ1`) created field
`PVTSSF_lAHOASd_cc4BdI9qzhYRk9k` with these immutable options:

| Value | Option ID | Feature |
|---|---|---|
| Phase 0 | `221d176d` | #1 |
| Phase 1 | `f817c01d` | #2 |
| Phase 2 | `8188d955` | #3 |
| Phase 3 | `6b779f39` | #4 |
| Phase 4 | `a7bbab56` | #5 |
| Phase 5 | `2880879a` | #6 |
| Phase 6 | `d4e86930` | #7 |
| Phase 7 | `85d21677` | #8 |
| Phase 8 | `6d252c8e` | #9 |

A postcondition read confirmed features #1–#9 map respectively to Phase 0–Phase 8 and
that no other item has a direct Phase. The run made one field creation attempt followed by
nine item assignment attempts; all returned success and the read matched the intended
state.

The owner authorized proceeding through #46, but the CLI effects were not preceded by a
retained immutable effect-plan/review identity and no safe create replay was attempted.
The live application result is therefore `needs-review`, not a qualification pass, even
though its current postcondition is correct. The field and mappings are retained; recovery
must inspect them by immutable identity rather than rerun field creation.

The production adapter now validates the complete existing catalog before item effects,
but Work Manager does not yet own an idempotent field/option/view creation plan. Therefore
the requested repeatable configuration reconciliation is also incomplete; direct CLI
field creation is not the supported fallback.

## Explicit non-pass

The required saved `Phases` view is `not-configured`. GitHub's public CLI and GraphQL
schema expose no Project-view creation mutation, so the agent cannot create or verify the
human-owned view without changing authority or using an unsupported interface. The owner
must create one `Phases` view in Project #8, grouped or ordered by Phase and showing useful
Status, Readiness, and native sub-issue progress context. That action must not imply phase
or release completion.

A native GraphQL read found the existing `All Tasks`, `Roadmap`, `Current Work`, `Backlog`,
and `Needs Refinement` views and no `Phases` view. This distinguishes a confirmed missing
view from an unavailable inventory while retaining creation as a human action.

Issue #46 and its PR cannot pass until the view exists, its identity/layout is re-read,
and distinct Spec review accepts the complete result.

## Deterministic verification

Coverage includes direct immutable-option projection and replay, native-parent-bound
context without copied assignment, clearing a duplicated child value, justified
cross-cutting assignment, invalid Phase, orphan/duplicate parent assignment, incomplete
catalog, renamed option, duplicate/wrong-type field, stale option identity, and the
existing target/configuration/partial-effect failure cases.

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
go test ./...
go vet ./...
go run ./cmd/starter-kit changes check --repository .
git diff --check
```

The exact completing revision still requires native CI and distinct Standards/Spec review.
