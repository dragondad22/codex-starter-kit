# Issue #75 — governed squash-delivery evidence

**Date:** 2026-07-21

**Issue:** [#75](https://github.com/dragondad22/codex-starter-kit/issues/75)

**Parent:** [#4](https://github.com/dragondad22/codex-starter-kit/issues/4)

**State:** Development candidate; seven pre-effect, one provider-effect-attempt, and two
post-effect live qualification failures reproduced; current-source verification state is
recorded below; live qualification pending

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
- Required check identity now includes the ruleset's GitHub App integration ID; an
  unbound same-name status is distinct evidence and cannot supersede the required App
  check. The fixture rule binds `contract-delivery` to GitHub Actions integration `15368`.
- Universal review is a required exact-source declaration naming actor, role, capability,
  distinct implementation/review contexts, approval and findings routes, limitations,
  and any stronger-policy requirement. GitHub approval is eligible only when its exact
  head and normalized evidence match that declaration.

## Policy boundary

Only current exact-head checks and reviews are eligible. Required checks may be pending,
passed, or failed; review may be unrequested, pending, approved, or changes-requested.
A capable review outside the implementation context is distinct from checks, effective
rules, merger authority, and optional product outcome approval. Stronger qualified
independence composes only when governed intent requires it. Review requests are routing,
not approval evidence, and one review evidence identity cannot satisfy two required roles.

The engine admits only an exact issue-named branch, same-repository issue-linked PR with
reciprocal `Closes #N`,
expected base/head, canonical #74 delivery claim, exact required-check catalog, supported
squash method, current capability, and matching DEC-0022 mandate. Wrong, partial, stale,
ambiguous, unsupported, or closed-unmerged observations remain waiting or non-pass.
Rules grant no bypass inference. Operational baseline-rules mutation remains outside
scope; one marker-owned sandbox fixture ruleset is separately planned, applied, verified,
and cleaned through the existing rules App.

GitHub does not reliably report a retrospective merge method. A qualifying merged
observation therefore combines current default-branch reachability with a retained
successful squash-effect receipt for the exact issue/source/PR/head/merge tuple. Only that
qualifying merge may invoke Work Manager completion reconciliation with the same mandate.

## Deterministic coverage in the development candidate

Engine and native HTTP fixture tests exercise branch and PR absence, claimed draft PR,
draft-to-ready transition, check pending/pass/fail, review request/pending/approval,
changes requested, optional separate product approval, merge-ready exact head,
closed-unmerged state, stale head and capability rejection, effective-rule mismatch,
wrong or changed apply-time actor/permission/expiry, missing mandate, cumulative mandate
use, same-name wrong-integration check evidence, missing reciprocal closure, corrupt state,
single-attempt ambiguous-effect recovery and unresolved non-pass, squash observation,
restart, completion reconciliation, and no-change replay.

The completion path composes the existing Work Manager parent/direct-dependent behavior
rather than introducing another Project mutation implementation. Adapter fixtures use
native Go HTTP requests and credential-free normalized evidence; they do not establish a
live GitHub service or permission claim.

## Current verification state

This record intentionally does not claim live qualification or completion. The first
independent Standards and Spec reviews found eight blockers, and subsequent audit found
credential-binding, exact-ruleset, and bypass-identity gaps. The branch now contains
regression-covered remediation. Complete local Python/Go/documentation gates, Go vet and
race tests, embedded workflow syntax validation, and final independent Standards and Spec
reviews passed at source `21f6d653de102528f8491cdab5af5cb774c62424`; its native
GitHub Actions checks also passed.

The four reviewed workflows were then installed byte-for-byte in sandbox commit
`854c2a95bbc045396ab99cd85d20aab676abb94e`. Approved planning run
`30159572938` failed before any external effect because the strict `sandbox-live-plan`
input interface rejected the generator-emitted `stage_contract` field. A public-CLI
regression reproduces that exact schema mismatch, and the live-plan input now accepts the
strictly typed stage contract without weakening unknown-field rejection. Refreshed gates,
Go vet and race tests, and independent Standards and Spec repair reviews pass. Native CI,
the separate completing-product-PR review, and the live sandbox journey remain pending at
the resulting source revision.

Replacement planning run `30160621400` then completed credentialed inspection and emitted
the expected immutable three-effect plan, but the workflow rejected the envelope because
an omitted empty `inspection.problems` slice serializes as `null` rather than literal
`[]`. No apply workflow was dispatched. The plan and apply postcondition assertions now
accept only an omitted, null, or empty problems field while rejecting wrong types and any
retained problem. Refreshed local Python, Go, documentation, vet, race, YAML, and assertion
checks and independent Standards and Spec reviews pass. Native CI and live retry remain
pending at the resulting source revision.

Planning retry `30160878551` passed with the strict assertions. Apply run `30160902136`
regenerated and bound that exact plan, then returned `non_pass` before any provider effect:
zero receipts were emitted and read-only inspection confirmed that no marker-owned issue
was created. The generator had authorized the App slug as the mandate actor while the
role-scoped adapter correctly reports the logical actor `seeder`; exact App identity
already remains separately bound through the authority credential identity. Generated
mandates now authorize the logical stage role and retain the exact App installation,
account, permissions, and compatibility in authority. The stage generator test now plans
and applies every emitted role-scoped mandate through the public sandbox lifecycle seam.
Refreshed local Python, Go, documentation, vet, and race gates and independent Standards
and Spec reviews pass. Native CI and live retry remain pending at the resulting source
revision.

Role-bound apply run `30161155746` then created the three approved marker-owned issues
exactly once as parent `#26`, delivery `#27`, and dependent `#28`, with three successful
receipts. Immediate verification did not observe the new list entries, and a later
read-only planning run `30161226802` observed all three exact native identities but still
planned updates because the engine compared the desired attribute map to the richer
identity handoff map exactly. No apply retry occurred. Sandbox matching now permits only
the adapter-curated `number`, database `id`, and `node_id` additions for initial fixture
issue identity handoff; any other unexpected observed attribute remains drift. A public
lifecycle regression covers accepted initial native identities, rejected unrelated
attributes, and drifted identities after those values become managed. Refreshed local
Python, Go, documentation, vet, and race gates and independent Standards and Spec reviews
pass. Native CI also passes at source
`31fd061522388eeb6a6ef7b39035ca7af6036114`. Read-only planning run `30161491466`
then observed all three issues as converged with zero effects.

Credential-free delivery-input run `30161558471` bound those exact issue identities and
the generator-derived final workflow digest to the governed request and mandate. Read-only
`issues-governed` planning run `30161587926` emitted exactly three issue reconciliation
effects with no inspection problems. Apply run `30161619161` stopped before regenerating
the stage input or invoking any provider effect because the downloaded planning artifact
retained the delivery input at `planning/delivery-input/issue-75-delivery-input.json`
while apply compared a nonexistent flattened path. The apply workflow now compares the
actual retained artifact path. Refreshed gates, native CI, independent review, workflow
reinstallation, and a newly source-bound live plan remain pending.

The repaired, source-bound `issues-governed` apply run `30161897182` then updated all
three issue contracts and converged with three receipts. Project setup apply run
`30161966733` reconciled five required Status/Readiness values and also converged.
Relationship planning runs `30161994402` and bounded read-only retry `30162029965` both
stopped before effects because the reconciler adapter queried Project views and workflows
even though the stage requested only issue relationships. The relationship token
intentionally has only `issues:write` and `metadata:read`; undeclared Project read access
is not added. Sandbox capability now queries Project identity only when Project resources
are requested or the bound credential explicitly declares Project authority; observation
queries Project inventory only for Project resources. The native relationship regression
uses the production-equivalent narrow permission set and fails on any Project request,
while a table regression requires identity and inventory reads for each of the six
supported Project resource kinds. Refreshed local Python, Go, documentation, vet, and
race gates pass. Final independent review, native CI, and a newly source-bound live plan
remain pending.

Source-bound relationship plan `30162446037` then emitted exactly the two approved native
relationship effects without Project access, and apply run `30162478736` converged with
both relationships present. File plan `30162511940` emitted the one expected
marker-owned workflow effect. Apply run `30162538860` reached the provider but retained an
`error` receipt, no file, and a failed missing-resource verification; no effect retry
occurred. The generated file-stage token requested only `contents:write` and
`metadata:read`, while a Contents API mutation under `.github/workflows/` also requires
`workflows:write`. Installation `147094309` currently exposes that owner-added permission,
so file create, update, and cleanup stages now request and bind the exact
`workflows:write` permission rather than broadening to another credential. Refreshed local
Python, Go, documentation, vet, and race gates pass. Final Standards review, native CI,
and a newly source-bound live plan remain pending.

The permission-corrected file apply `30162829001` then created and verified the exact
workflow and converged. Ruleset apply `30162880623` created marker-owned ruleset
`19734893` exactly once, but verification reported drift: GitHub returns
`do_not_enforce_on_create:false` in the required-check parameters when the request omits
that default. No ruleset apply retry occurred. The fixture now declares the canonical
false value explicitly, preserving active enforcement, strict current-head checks,
GitHub Actions integration `15368`, and an empty bypass list. A regression requires the
field to be present and false, and a native public-lifecycle regression requires the
canonical observed definition to plan no change with zero effects. Refreshed local Python,
Go, documentation, vet, and race gates pass. Final independent review, native CI, and a
newly source-bound read-only plan remain pending; that plan must recognize ruleset
`19734893` without an update effect.

The first governed-delivery transition run `30163177714` passed its credential-free
envelope and explicit combined-authority gates, then stopped before its planned
`create-branch` effect with `delivery capability changed before apply`. The App provider
had minted a second repository- and permission-scoped installation token between
capability planning and apply; GitHub assigned the second token a different expiry, so
the engine correctly treated it as a changed authority. No branch was created and the
retained transition artifact was empty. The provider now reuses one validated,
in-memory-only installation credential until its exact expiry while signing a fresh
short-lived App JWT for later identity queries. It drops the expired cached secret before
revalidating App and installation identity and minting a replacement. A concurrent
regression proves one mint, isolated returned permission data, App-JWT renewal before the
installation token expires, and installation-token replacement at expiry; the adapter's
independent changed-credential rejection remains unchanged. Refreshed gates, independent
review, native CI, and a newly source-bound delivery input remain pending before another
first transition.

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
[`issue-75-sandbox-stage-plan.yml`](issue-75-sandbox-stage-plan.yml) and
[`issue-75-sandbox-stage-apply.yml`](issue-75-sandbox-stage-apply.yml), and credential-free
[`issue-75-delivery-input.yml`](issue-75-delivery-input.yml), do not prove that
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
Status/Readiness; `relationships-setup` consumes the same issue handoff; `file-initial`
installs the check workflow on unprotected `main`; `rules-setup` then installs one
marker-owned active ruleset requiring the `contract-delivery` context from GitHub Actions
App integration `15368`; and `file-stale` updates the delivery head to exercise
invalidation. Seeder stages use only
`contract-seeder`; Project/relationship stages use only `contract-reconciler`; ruleset
stages use only `contract-rules`. Apply
regenerates and byte-compares the credential-free input, binds the downloaded plan to its
own active mandate, verifies convergence, and performs a second read-only plan to retain
the postcondition and issue identity handoff.

Cleanup is invoked explicitly after terminal replay or for recovery, never after an
ordinary one-transition dispatch. It uses the same independently planned/applied stage
contract in dependency order: `cleanup-delivery` (seeder), `cleanup-rules` (rules),
`cleanup-file` (seeder), `cleanup-relationships` (reconciler), and `cleanup-issues`
(seeder). The operator does not dispatch the next stage unless the current stage's apply,
verification, and effect-free postcondition converge. This keeps every private key in its
existing protected role environment, stops dependent cleanup after an exact-identity or
drift refusal, and retains each stage's active mandate and receipts. This is a reviewed
executable candidate, not evidence that cleanup or any other live effect occurred.

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
