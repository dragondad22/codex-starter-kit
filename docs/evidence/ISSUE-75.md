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

The reviewed workflow candidates
[`issue-75-contract.yml`](issue-75-contract.yml) and
[`issue-75-contract-cleanup.yml`](issue-75-contract-cleanup.yml), plus the role-isolated
[`issue-75-sandbox-stage-plan.yml`](issue-75-sandbox-stage-plan.yml) and
[`issue-75-sandbox-stage-apply.yml`](issue-75-sandbox-stage-apply.yml), and credential-free
[`issue-75-delivery-input.yml`](issue-75-delivery-input.yml) and
[`issue-75-cleanup-plan.yml`](issue-75-cleanup-plan.yml), do not prove that
the journey ran. The main workflow requires exact request, bound active mandate, and 40-character
Starter Kit revision artifacts before it emits a credential-free envelope. Each dispatch
may then execute only one semantic transition; a changed observation requires another
dispatch through the same envelope gate. Evidence artifacts retain for 30 days and contain
no credential material.

Sandbox resources progress organically through separately planned/applied stages:
`issues-setup` emits immutable issue identities; the delivery-input workflow binds those
identities and the generator-derived final fixture workflow digest into one complete
governed request/mandate artifact; `issues-governed` consumes that exact artifact and
patches the three native fixture bodies with their managed markers, metadata, and
executable contracts; `project-setup` then consumes their node IDs to set exact Project
Status/Readiness; `relationships-setup` consumes the same issue handoff; and
`file-initial`/`file-stale` prepare the exact-head check fixture. Seeder stages use only
`contract-seeder`; Project/relationship stages use only `contract-reconciler`. Apply
regenerates and byte-compares the credential-free input, binds the downloaded plan to its
own active mandate, verifies convergence, and performs a second read-only plan to retain
the postcondition and issue identity handoff.

Cleanup is invoked explicitly after terminal replay or for recovery, never after an
ordinary one-transition dispatch. A credential-free builder combines the four exact
stage-planning artifacts into one content-addressed #75-native bundle containing exactly
four ordered apply inputs: `cleanup-delivery` (seeder), `cleanup-file` (seeder),
`cleanup-relationships` (reconciler), and `cleanup-issues` (seeder). The bundle binds one
episode, source, target, approval, and active stage mandates. Its delivery stage preserves
the exact reciprocal `Closes #<delivery>` marker; the other three use the exact episode
marker. The protected combined runner selects only the key required for each generic
sandbox apply, requires every stage to converge, and retains per-stage results plus one
cleanup receipt. This is a reviewed executable candidate, not evidence that cleanup or
any other live effect occurred.

Each later dispatch must supply the prior run and exact state artifact. The workflow
rejects symlinks and every payload path except the prior transition receipt and
`.starter-kit/delivery/state.json`, `.starter-kit/work-manager/state.json`, and
`.starter-kit/work-mandates.json`. It admits only the latest non-expired canonical state
artifact from this workflow, binds its manifest to the exact source, mandate, delivery
resource digest, and predecessor run, and rejects initial-state or older-artifact replay.
It restores only those state files with owner-only permissions and uploads the next
integrity-protected state with the transition receipt.
The exact request uses repository `.` so state remains at the sandbox
workspace root. Without prior state, only the initial `create-branch` transition is
admissible. Completing the journey therefore requires repeated dispatches plus external
check and distinct-review perturbations; one workflow run is one transition, not live
qualification.

One human authority action remains before the live workflow is configured: create and
protect a `contract-delivery` environment and populate exactly
`CSK_RECONCILER_APP_PRIVATE_KEY` and `CSK_SEEDER_APP_PRIVATE_KEY`. The current composed
runner exposes the union of those two keys within that job. Dispatch therefore requires an
explicit owner assertion that the environment exists and is approved. Reviewer and rules
secrets remain absent from that environment and retain their separate protected routes.
