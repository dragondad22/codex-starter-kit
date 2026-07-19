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
that no other item has a direct Phase. Their immutable issue-content and Project-item IDs
are now retained by the reviewed #46 configuration manifest rather than rediscovered by
title or issue number.

The owner authorized proceeding through #46, but the CLI effects were not preceded by a
retained immutable effect-plan/review identity and no safe create replay was attempted.
The live application result is therefore `needs-review`, not a qualification pass, even
though its current postcondition is correct. The field and mappings are retained; recovery
must inspect them by immutable identity rather than rerun field creation.

The production adapter now validates the complete existing catalog before item effects.
The engine's content-addressed external-resource lifecycle also plans the exact Phase
field, nine options, `Phases` view, and nine feature assignments under a separate
operational target and DEC-0022 mandate. It verifies the API user and classic `project`
scope, creates a missing user-owned view only through the version-pinned REST route,
re-observes view and item postconditions through GraphQL, and rejects stale identity or
human-owned drift rather than duplicating it. The historical `Sandbox*` type names remain
v1 compatibility labels; the operational mandate does not inherit sandbox authority.

## Saved view and current live state

The owner created and renamed the human-facing view before the supported automation route
was applied. A native GraphQL read now identifies `Phases` as
`PVTV_lAHOASd_cc4BdI9qzgLBdLU`, table view number 6. It is grouped and sorted ascending by
the immutable Phase field and displays Title, Status, Readiness, and native sub-issue
progress. This makes Phase membership and progress understandable without making Phase a
release or completion signal.

GitHub API version `2026-03-10` documents saved-view creation for user-owned Projects at
`POST /users/{user_id}/projectsV2/{project_number}/views`. That endpoint does not support
GitHub App user, App installation, or fine-grained personal access tokens. The adapter
therefore uses the classic user-token route only after native actor/scope verification.
The exact-source live configuration plan, mandate-bound apply/no-change receipt, replay,
and final independent review remain required before this record claims pass.

## Deterministic verification

Coverage includes direct immutable-option projection and replay, native-parent-bound
context without copied assignment, clearing a duplicated child value, justified
cross-cutting assignment, invalid Phase, orphan/duplicate parent assignment, incomplete
catalog, renamed option, duplicate/wrong-type field, stale option identity, exact saved-
view observation/creation/re-read, unavailable user view route, immutable Project-item
field update/replay, verified API user/classic scope, and the existing target,
configuration, and partial-effect failure cases.

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
go test ./...
go vet ./...
go run ./cmd/starter-kit changes check --repository .
git diff --check
```

The exact completing revision still requires native CI and distinct Standards/Spec review.
