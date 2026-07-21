# Issue #75 — governed squash-delivery evidence

**Date:** 2026-07-21

**Issue:** [#75](https://github.com/dragondad22/codex-starter-kit/issues/75)

**Parent:** [#4](https://github.com/dragondad22/codex-starter-kit/issues/4)

**State:** Development candidate; final gates and live qualification pending

## Implemented deterministic contract

- Added a dedicated delivery lifecycle seam: `InspectDelivery`, `PlanDelivery`,
  `ApplyDelivery`, `VerifyDelivery`, and `DeliveryStatus`. It plans one semantic
  transition at a time rather than hiding branch, PR, review, merge, and reconciliation
  in one opaque operation.
- Added orthogonal observations and dispositions for absent/present branch, absent/draft/
  ready/closed/merged PR, exact-head checks, review routing and results, optional product
  approval, effective rules, merge reachability, and durable completion.
- Added mandate-contained effects for issue-named branch creation, claimed draft-PR
  creation, readying, reviewer routing, exact-head squash merge, and terminal Work Manager
  reconciliation. A narrowed effect credential may be separate from the read credential,
  but both remain bound to one immutable repository.
- Added exact postcondition recovery for ambiguous effect responses without retrying the
  external mutation. The shared integrity-protected mandate ledger reserves cumulative
  external-effect use before each call.
- Added integrity-protected delivery state and completion memory under
  `.starter-kit/delivery/state.json`. Completion binds managed issue/source, PR/head/merge,
  checks, reviews, approvals, rules, mandate, and nested Work Manager receipts. Exact replay
  is no-change and does not rewrite historical receipts.
- Added native GitHub observation for issue marker/linkage, branch/ref, linked PR,
  requested reviewers, check runs and commit statuses, review evidence/trust, effective
  branch rules, squash capability, and current default-branch delivery reachability.

## Policy boundary

Only current exact-head checks and reviews are eligible. Required checks may be pending,
passed, or failed; review may be unrequested, pending, approved, or changes-requested.
A capable review outside the implementation context is distinct from checks, effective
rules, merger authority, and optional product outcome approval. Stronger qualified
independence composes only when governed intent requires it. Review requests are routing,
not approval evidence, and one review evidence identity cannot satisfy two required roles.

The engine admits only an exact issue-named branch, same-repository issue-linked PR,
expected base/head, canonical #74 delivery claim, exact required-check catalog, supported
squash method, current capability, and matching DEC-0022 mandate. Wrong, partial, stale,
ambiguous, unsupported, or closed-unmerged observations remain waiting or non-pass.
Rules grant no bypass inference and operational rules mutation remains outside scope.

GitHub does not reliably report a retrospective merge method. A qualifying merged
observation therefore combines current default-branch reachability with a retained
successful squash-effect receipt for the exact issue/source/PR/head/merge tuple. Only that
qualifying merge may invoke Work Manager completion reconciliation with the same mandate.

## Deterministic coverage in the development candidate

Engine and native HTTP fixture tests exercise branch and PR absence, claimed draft PR,
draft-to-ready transition, check pending/pass/fail, review request/pending/approval,
changes requested, optional separate product approval, merge-ready exact head,
closed-unmerged state, stale head and capability rejection, effective-rule mismatch,
wrong actor or missing mandate, cumulative mandate use, single-attempt ambiguous-effect
recovery and unresolved non-pass, squash observation, restart, completion reconciliation,
and no-change replay.

The completion path composes the existing Work Manager parent/direct-dependent behavior
rather than introducing another Project mutation implementation. Adapter fixtures use
native Go HTTP requests and credential-free normalized evidence; they do not establish a
live GitHub service or permission claim.

## Current verification state

This record intentionally does not claim a final pass. Local implementation and focused
tests are still being completed on the issue branch. The exact completing revision,
complete repository gates, race/vet checks, independent Standards and Spec reviews,
GitHub Actions Linux/macOS/Windows matrix, and live sandbox journey have not yet been
recorded here.

## Pending live qualification and completion

The live journey requires one current content-addressed DEC-0022 mandate for its exact
source, issue, sandbox repository, actors, permissions, effects, limits, expiry, and
recovery. The existing fixture-seeder GitHub App installation may act as the logical
merger through a short-lived repository-narrowed token with `contents:write`,
`pull-requests:write`, and GitHub's mandatory `metadata:read` permission. The distinct reviewer and rules
identities remain separate; no new human account, bypass, baseline-rules mutation, or
generic credential fallback is part of the candidate.

Qualification must exercise a fresh marker-scoped issue/branch/PR through draft, ready,
checks, distinct review, stale-head invalidation, qualifying squash merge, selected-item/
parent/dependent reconciliation, cleanup, and replay. The completing product PR requires
its own distinct review; sandbox fixture review is not evidence for the product change.
Issue #76 owns aggregate live qualification and final support claims.
