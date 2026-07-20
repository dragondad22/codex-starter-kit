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
view. No live Project effect was attempted by the correction itself. The deliberate
owner-approval seam was first exercised by the historical bounded non-pass below and then
completed by the exact-head zero-effect qualification after transient-read recovery was
implemented.

## Zero-effect completion attempt

Owner comment
[`5017894175`](https://github.com/dragondad22/codex-starter-kit/issues/46#issuecomment-5017894175)
was verified as an unchanged `OWNER` record authored by `dragondad22` at
`2026-07-20T00:17:09Z`. It approves candidate
`368f043accbc297fcbea1ddf41d564b60aff8eb0`, the pinned owner/repository/Project,
reconciler actor, recovery owner, classic OAuth scopes `gist`, `project`, `read:org`,
`repo`, and `workflow`, semantic `projects:write`, the exact 20 Phase resources, and a
24-hour validity interval. The execution package deliberately narrowed the owner's
20-effect ceiling to zero. This historical approval and result remain evidence for the
older candidate; they do not authorize or describe the later passing run.

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

## Transient read correction

The subsequent bounded diagnosis recorded in issue comments `5017922768` and
`5017941354` established that the red capability attempts received intermittent GitHub
ProjectsV2 REST `503` responses. The adapter had made one read and then reported a
semantic Project identity or field-inventory problem. Later independent `200` responses
did not invalidate that captured non-pass and were not treated as replay evidence.

The correction keeps the lifecycle and adapter interfaces small and moves the recovery
policy behind the native HTTP implementation. Idempotent REST `GET` reads now receive at
most three attempts for `502`, `503`, and `504`. A `429` is eligible only when its valid
`Retry-After` delay fits the two-second aggregate wait ceiling. Context cancellation and
deadlines interrupt waiting. REST POST and other effect requests, authentication,
permission, not-found, semantic identity mismatch, malformed successful responses, and
all other non-transient results remain single-attempt. Exhaustion is retained as a
distinct provider-transient capability or observation problem without provider bodies or
credentials.

Deterministic public-seam cases cover exact two-read `503` recovery for Project identity
and fields, bounded persistent-transient user and Project identity non-pass with no
effects, `502`/`504` recovery, bounded and cumulative `429` delay handling, canceled
waiting, mixed rate/transient exhaustion, strict `Retry-After` parsing, and the required
no-retry paths. This
does not rewrite the historical zero-effect non-pass above. The later exact-head run used
a separate owner record and content-addressed mandate.

## Exact-head zero-effect qualification

Owner comment
[`5028032912`](https://github.com/dragondad22/codex-starter-kit/issues/46#issuecomment-5028032912)
was verified as an unchanged `OWNER` record authored by `dragondad22` at
`2026-07-20T22:52:22Z`. It approves candidate
`bf9827c124a7d24ae14802c65600a4083d9cb14b`, the same immutable target, reconciler actor,
complete observed classic OAuth scope set, semantic `projects:write`, exact 20 Phase
resources, and a zero-effect ceiling through `2026-07-21T22:52:22Z`.

The renewed [zero-effect mandate](issue-46-zero-effect-mandate-bf9827c.json) is content addressed
as `sha256:01665223c0df125ff080a12bf65a410c161504aed11ea99dabefe2ceb98be9b3`.
At the pinned observation time `2026-07-20T22:53:31Z`, the lifecycle observed every
resource and produced immutable plan
`sha256:b004f976c073a3322b0dba357b7e5c886af7f3f3f97e2919a8ced3ca53964a65`
with zero effects. Effect-free apply returned `no_change`; verification returned `pass`
for `GITHUB-SANDBOX-001`; reinspection produced the identical plan; and replay apply
returned `no_change`. Project effect count is zero.

The independently reviewable, content-addressed
[passing receipt](issue-46-zero-effect-pass.json) retains the exact source, approval,
mandate, semantic capability and authority, all 20 per-resource postconditions, immutable
plan, zero-effect apply result, verification controls/evidence, and identical replay. Its
canonical content excluding `evidence_id` is
`sha256:97add3f7a473c26505496b511e9c5b0e96d9273398ae5983746010cdac36c1cc`.
No effect receipts exist because both apply passes contained zero effects. Raw command
output remains subject to the mandate's 30-day ceiling and is not committed; the durable
normalized receipt excludes credentials, provider response bodies, and runner-local paths.

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

The executable candidate passed its exact-head zero-effect lifecycle. The evidence-only
completion revision still requires native CI and distinct Standards/Spec review; any
subsequent executable change invalidates the live qualification and requires a new
exact-source mandate and run.
