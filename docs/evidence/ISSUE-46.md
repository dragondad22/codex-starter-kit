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

The owner authorized proceeding through #46, but the original CLI effects were not
preceded by a retained immutable effect-plan/review identity and no safe create replay was
attempted. The original live application result is therefore `needs-review`, not a
qualification pass, even though its current postcondition is correct. The field and
mappings are retained; recovery must inspect them by immutable identity rather than rerun
field creation.

The production adapter now validates the complete existing catalog—including the exact
option count, names, and immutable IDs—before item effects.
The engine's content-addressed external-resource lifecycle also plans the exact Phase
field, nine options, `Phases` view, and nine feature assignments under a separate
operational target and DEC-0022 mandate. It verifies the API user and classic `project`
scope while retaining and exactly binding every other observed classic OAuth scope. Apply
accepts only an independently retained mandate JSON artifact whose approver and recovery
owner match the pinned target owner, whose approval identity is a retained #46 owner record,
target, actors, timestamps, authority profile, resource digests, ceilings, retention, and
recovery owner validate through the lifecycle seam; caller flags cannot manufacture it.
The command checks the trusted execution time is within the retained approval interval
before constructing transport for either planning or apply.
Before observation, the adapter also resolves the configured user/Project-number REST
route and requires its Project node ID plus owner login, immutable ID, and kind to match
the GraphQL target and retained owner. A mixed REST/GraphQL target is therefore non-pass
before any effect.
The adapter re-observes view and item postconditions through GraphQL and rejects stale
identity or human-owned drift rather than duplicating it. Clean-create tests omit
provider-assigned field/option identities, retain GitHub's returned IDs, converge, and
separately prove that an already-pinned stale identity remains non-pass. The historical
`Sandbox*` type names remain v1 compatibility labels; the operational mandate does not
inherit sandbox authority.

## Saved view and current live state

The owner created and renamed the human-facing view before the supported automation route
was applied. A native GraphQL read now identifies `Phases` as
`PVTV_lAHOASd_cc4BdI9qzgLBdLU`, table view number 6. It is grouped and sorted ascending by
the immutable Phase field and displays Title, Status, Readiness, and native sub-issue
progress. This makes Phase membership and progress understandable without making Phase a
release or completion signal.

GitHub API version `2026-03-10` documents saved-view creation for user-owned Projects at
[`POST /users/{user_id}/projectsV2/{project_number}/views`](https://docs.github.com/en/rest/projects/views#create-a-view-for-a-user-owned-project).
That endpoint does not support
GitHub App user, App installation, or fine-grained personal access tokens. The adapter
therefore uses the classic user-token route only after native actor/scope verification.
Its request schema does not expose grouping or sorting. The adapter consequently returns
`not-configured` without creating a partial view when the required grouped/ascending-sorted
`Phases` view is absent; it can still verify and replay the existing matching human-created
view. No live Project effect was attempted for this correction. A fresh independently
retained mandate covering the exact final source, full observed scope set, and final resource
digests is required before effect-free live planning/apply/replay can be final evidence.
That deliberate owner-approval seam was exercised by the bounded attempt below; its
non-pass result remains the current completion boundary.

## Zero-effect completion attempt

Owner comment
[`5017894175`](https://github.com/dragondad22/codex-starter-kit/issues/46#issuecomment-5017894175)
was verified as an unchanged `OWNER` record authored by `dragondad22` at
`2026-07-20T00:17:09Z`. It approves candidate
`368f043accbc297fcbea1ddf41d564b60aff8eb0`, the pinned owner/repository/Project,
reconciler actor, recovery owner, classic OAuth scopes `gist`, `project`, `read:org`,
`repo`, and `workflow`, semantic `projects:write`, the exact 20 Phase resources, and a
24-hour validity interval. The execution package deliberately narrowed the owner's
20-effect ceiling to zero.

The retained [zero-effect mandate](issue-46-zero-effect-mandate.json) is content addressed
as `sha256:b09038fbd3ed1a2f0a38b92ff8f9c033b404c25b3a12090b27c96ad4639822c1`.
It binds all 20 resource digests, `max_effects: 0`, public Project metadata, zero-dollar
cost, no deletion or human-view overwrite, 30-day retention, and recovery owner
`dragondad22` from approval time through `2026-07-21T00:17:09Z`.

At `2026-07-20T00:22:35Z`, a fresh isolated evidence-state repository ran the reviewed
candidate through lifecycle planning. Capability inspection returned non-pass before an
immutable plan was emitted:

```text
reconciler: Project immutable identity or owner is unavailable or mismatched
sandbox capability is unavailable or stale
```

Execution stopped immediately. Apply, verification, and replay were not attempted, and
the Project effect count is zero. A subsequent read-only request to the exact configured
user/Project route returned the expected Project node, number, owner login, numeric owner
ID, and owner kind. That later match narrows the failure to an unavailable or inconsistent
capability observation rather than proving durable identity drift; it does not rewrite the
non-pass, authorize a retry, or qualify the live lifecycle. The redacted, content-addressed
[non-pass evidence](issue-46-zero-effect-plan-non-pass.json) omits credentials, provider
response bodies, and temporary paths. Its canonical content excluding the `evidence_id`
field is addressed as
`sha256:567464257526a4b61514ae53d4e40f6ebe15acf34dd76e77e4c47183e5846f58`.

## Deterministic verification

Coverage includes direct immutable-option projection and replay, native-parent-bound
context without copied assignment, clearing a duplicated child value, justified
cross-cutting assignment, invalid Phase, orphan/duplicate parent assignment, incomplete
catalog, renamed option, duplicate/wrong-type field, stale option identity, exact saved-
view observation and re-read, explicit non-pass without partial creation when required
grouping/sorting is not expressible, unavailable user view route, immutable Project-item
field update/replay, verified API user/classic scope, and the existing target,
configuration, and partial-effect failure cases. It also covers unexpected broader classic
OAuth scopes, extra Phase options, provider-ID adoption with stale-ID rejection, and both
configured Project owner routes for option reconciliation, including rejection when the
REST Project identity or owner differs from the GraphQL target. Project-item observation,
immutable-content lookup, and postcondition reads follow bounded GraphQL cursors; a later-page
assignment is found and exhaustion remains an explicit non-pass rather than evidence of
absence.

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
go test ./...
go vet ./...
go run ./cmd/starter-kit changes check --repository .
git diff --check
```

The exact completing revision still requires a fresh passing zero-effect lifecycle,
native CI, and distinct Standards/Spec review.
