# Issue #75 — governed squash-delivery evidence

**Date:** 2026-07-21

**Issue:** [#75](https://github.com/dragondad22/codex-starter-kit/issues/75)

**Parent:** [#4](https://github.com/dragondad22/codex-starter-kit/issues/4)

**State:** Development candidate; eight pre-effect, two provider-effect-attempt, and five
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
restart, completion reconciliation, and no-change replay. Delivery observations
canonicalize check, review, approval, problem, and effective-rules evidence before
computing the optimistic-concurrency revision; reordered equivalent GitHub payloads retain
one revision while semantic gate changes still invalidate it.

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

Source `c5523d46f6c63e03d1693625ea5e9243fdf70252` then passed local gates, both
independent reviews, and native CI run `30163400672`. Delivery-input run `30163525467`
bound that exact source, and transition run `30163549727` created the exact issue-named
branch from sandbox `main` and converged at `pull-request-absent`. The next transition
run `30163603727` reached GitHub's create-PR endpoint but received 422; the retained
transition artifact is empty, no PR exists for the branch, and the branch remains at the
approved creation head. No POST retry occurred. The fixture had seeded the initial
workflow on `main` before branch creation, leaving no head-only commit from which GitHub
could create a PR.

The candidate now adds a predecessor-bound `file-candidate` stage after branch creation.
It consumes the integrity-checked successful create-branch state artifact and creates one
distinct intermediate workflow commit so draft-PR creation is meaningful. The existing
`file-stale` stage now binds the exact candidate head and remains later, after
candidate-head check and review evidence, to invalidate both against a new final head.
Repository-file apply rechecks the exact approved branch head immediately before its
Contents effect. A separate
`cleanup-orphan-branch` recovery stage does not weaken PR-bound `cleanup-delivery`: it
requires the integrity-checked successful create-branch state artifact, exact delivery
issue identity, and branch SHA; rechecks issue identity/marker, head, and all-state PR
absence immediately before deletion; refuses malformed or multi-page results and any PR
history; and uses only seeder `contents:write`, `metadata:read`, and
`pull-requests:read`.
These stage and authority changes advance the explicit sandbox configuration revision to
`issue-75-sandbox-config-v2`.

Source `466583879133ac2f54f0d0b8666b3d6f97c6f6d1` passed refreshed local gates, both
independent reviews, and native CI run `30164196138`. Sandbox PR `#29` passed the required
`contract-delivery` check and installed the three changed control workflows byte-for-byte
in commit `ef20cadd7b8cb85500131f055ee7601e8ad44499`. Orphan-cleanup plan
`30164330893` bound the exact successful branch-creation artifact, issue identity, branch
SHA, source, and mandate and emitted one delete effect with no problems. Apply run
`30164366716` issued that DELETE exactly once and retained an `applied` receipt, but its
immediate verification and separate postcondition read briefly observed the deleted ref
at the approved SHA. The exact ref subsequently returned `404`; read-only replay plan
`30164462373` retained no problems, `no_change:true`, zero effects, and an empty
observation.

The candidate now treats that result as GitHub read-after-delete propagation lag rather
than retrying the effect. Exact absent-branch observation polls only the idempotent ref GET
within a bounded exponential budget. It stops on `404`, surfaces a different SHA
immediately as drift, propagates cancellation, and retains the approved SHA as a residual
failure after budget exhaustion. Lifecycle regressions prove one DELETE followed by
bounded stale reads and convergence, effect-free replay, persistent-stale failure, changed
head handling, and cancellation. Refreshed gates, independent review, native CI, workflow
reinstallation, and a newly source-bound journey remain pending.

Source `72c0a16bcfc5ad8fbdb3430f8a2a9cebb1de8ddf` then passed refreshed gates,
independent review, and native CI run `30164650884`. The unchanged installed workflow
bytes still matched that source. Delivery-input run `30164767594` bound its exact request,
mandate, issue topology, and final workflow digest. Transition run `30164800746` passed
the credential-free envelope and authority gates but stopped before credentials or
effects because the predecessor gate treated the older successful state artifact
`30163549727` as a current episode even though the restore path would reject its different
source, mandate, and delivery-resource digest.

The predecessor gate now delegates continuity classification to the same exact reviewed
contract executable used for delivery. It strictly validates the latest unexpired
successful state artifact's allowlisted regular files, size ceiling, manifest schema and
run identity, required inventory, and every retained SHA-256 digest. An artifact matching
the current source, mandate, and delivery-resource digest still requires the exact
canonical predecessor; only a fully valid mismatched artifact is historical.

No workflow-level ref check authorizes the reset. When no predecessor was restored, the
runner requires the live inspection to produce exactly one `create-branch` effect before
Apply. A branch that exists before inspection yields a different plan and stops before
effects; a branch appearing after planning changes the engine's mandatory pre-apply
observation and also stops. Deterministic tests cover matching-state rejection, each
historical identity dimension, missing/malformed/extra/symlinked/hash-invalid artifacts,
unpaired mode flags, and zero/multiple/non-branch initial plans. This does not rewrite or
discard historical evidence and does not permit two heads for the same exact episode.

Source `a7961b59e314eb1d37d2ba33864da67dde1f847f` passed refreshed native CI run
`30165323279`. Sandbox PR `#30` passed `contract-delivery`, squash-merged the exact
contract workflow bytes as `ea58fa9b4f9a7e160ef82108f011762120bb3485`, and removed
its temporary branch. Delivery-input run `30165488626` then bound the new source,
topology, workflow digest, and active mandate. Transition run `30165513269` created only
the exact branch and retained its next state; separately planned/applied runs
`30165569364` and `30165599439` advanced that branch to candidate head
`f1ad09238895de4c483eb872fabdbd6430a7e108` with converged postconditions. Transition
run `30165630120` created draft sandbox PR `#31`; its exact-head `contract-delivery`
check passed.

The following transition `30165686763` stopped before effects because delivery
observation could not prove the governed squash method. The least-authority reconciler
App's repository response had not supplied positive repository-setting evidence, while
the installed effective ruleset declared only the required check. The candidate now
places `allowed_merge_methods: ["squash"]` in a marker-owned pull-request rule and derives
the effective restriction through the reconciler while the already-authorized merger App
supplies positive repository capability. Multiple effective rules are intersected with
the repository methods; omission, explicit disablement, conflicting restrictions,
last-push approval, and beta required-reviewer gates all fail closed. Refreshed gates,
independent review, ruleset reinstallation, and a newly source-bound journey remain
pending. Recovery plan `30165901548` bound exact PR `#31`, REST/GraphQL identities, and
head `f1ad09238895de4c483eb872fabdbd6430a7e108`; apply `30165989731` closed that
unmerged draft, deleted only its exact branch, verified convergence, and retained an
effect-free postcondition.

Ruleset plan `30166184938` then approved exactly one reconciliation to add the squash-only
pull-request rule. Apply run `30166211975` performed that effect, but immediate
verification and the second read-only plan refused convergence because GitHub added empty
`dismissal_restriction` and `required_reviewers` fields. The mutation was not retried.
The desired definition now includes those provider defaults explicitly so the installed
state and content-addressed plan can converge without semantic relaxation.

Because closed PR `#31` remains immutable timeline history for the first delivery branch,
the recovered next episode uses `contract/issue-75-20260721-02`. The delivery issue and
all other marker-owned fixture resources remain the same exact identities; the new branch
name prevents the historical closed PR from being mistaken for the current episode
without deleting or rewriting GitHub evidence.

Source `5f4b4788488a4eb3040a0d0c456b1c484358dde9` passed native CI run
`30166529683`. Delivery-input run `30167069149` bound that exact source and active
mandate. Transition `30167102252` created the rotated branch, and separately
planned/applied file runs `30167164594` and `30167192034` advanced it to candidate head
`6a9f54483f4cbd3330dbfdb20c47e65d3c58977c`. Transition `30167639726` created
draft PR `#32`; its exact candidate-head check passed. Transitions `30167693525` and
`30167747472` marked it ready and requested the configured reviewer. The distinct
`american-dragon-designs` review approved that exact candidate commit.

Stale-head plan `30174902363` then bound the approved candidate SHA and emitted one
repository-file effect. Apply run `30174931678` issued that Contents update exactly once
and retained its applied receipt. The immediate in-process verification briefly read the
prior candidate content and reported the file missing, while the following independent
postcondition already observed the exact final content with `no_change:true` and zero
effects. No mutation retry occurred. Repository-file observation now retries only the
idempotent Contents read within the existing bounded consistency budget, propagates
cancellation, and leaves persistent absence or drift non-pass. A regression reproduces a
stale first read followed by the exact new content. Cleanup regressions poll stale
approved content until `404`, retain persistent or unowned content as residual drift,
propagate cancellation, and prove one DELETE plus effect-free replay. Refreshed gates,
independent review, workflow reinstallation, and continuation from the already-created
exact final head remain pending.

The refreshed transition sequence then re-requested review for exact final head
`cbabcf3327451eaaa179dd6c46c4ae67e8cdd654` and squash-merged PR `#32` as
`9ac43a00f9807f223d4a2835dcb5b1dae327c1dd`. Post-merge observation initially remained
non-plannable because GitHub REST `2026-03-10` intentionally omits
`merge_commit_sha` from pull-request payloads. Live REST comparison and content reads
proved the merge and implemented bytes were current, while GraphQL resolved the immutable
PR node's `mergeCommit.oid` to the exact retained squash receipt. The adapter now uses
that GraphQL identity for both governed-work and delivery-lifecycle observation, retains
older-version REST conflict detection, and regression tests omit the retired REST field.
The full local Python, documentation, Go, vet, and race gates pass, and independent
Standards and Spec reviews report no findings. Completion reconciliation, cleanup,
replay, and refreshed native CI remain pending.

Native run `30175986968` passed Linux, macOS, Windows, equivalence, and aggregate
validation for source `077f6ee768d645987acc537badea9f10c1307d8c`. Because the
already-merged delivery state remains bound to its older source, recovery did not splice
new executable code into that episode. Exact plan/apply cleanup runs
`30176087976`/`30176116039`, `30176146257`/`30176168558`,
`30176186118`/`30176213806`, `30176238345`/`30176264183`, and
`30176291499`/`30176317018` removed the branch, ruleset, fixture workflow, native
relationships, and open fixture state in dependency order; each independent postcondition
was effect-free.

Fresh issues-setup plan `30176339070` and apply `30176364005` then reconciled and
independently verified the exact three fixture issues, but the post-effect handoff step
failed because its jq pipeline attempted to read `.plan` after changing the active input
to the issue-entry array. No create effect was retried. The always-retained postcondition
artifact recovers the exact source and issue identities. The handoff now retains the
postcondition root before deriving the issue map, and a workflow-contract regression
guards that binding. Refreshed gates, review, workflow installation, and continuation
from the recovered identities remain pending.

Source `38befd79e665fbd9b0e9757cefe093da96806026` passed native run
`30176448188`, and no-change issues runs `30176540032`/`30176566920` proved the
repaired handoff before the governed-input and setup sequence resumed. Initial delivery
run `30176875988` failed before an effect because historical merged PR `#32` still owns
the same `contract/issue-75-20260721-02` head identity and its older claim cannot be
treated as the new episode. The next fresh episode therefore rotates the exact delivery
identity to `contract/issue-75-20260721-03`; tests bind both delivery and sandbox staging
to that same branch. Refreshed gates, review, native CI, governed input, and continuation
remain pending.

Source `b00e19091a7c8271af1aba04ce3891d0091a1141` passed native run
`30177003777`. Delivery input `30177131395` and no-change governed-issue runs
`30177156382`/`30177182303` bound the fresh episode. Transition `30177210890`
created `contract/issue-75-20260721-03`; an incomplete read-only file-candidate identity
bundle failed in run `30177267076` before planning, and corrected plan/apply runs
`30177293334`/`30177314407` installed the exact candidate. Transitions
`30177336520`, `30177395371`, and `30177447403` created PR `#33`, marked it
ready, and requested its distinct reviewer. After exact candidate approval, stale-head
plan/apply `30225208689`/`30225234361` installed the final workflow digest, transition
`30225264268` re-requested review for the changed head, and the distinct reviewer approved
exact final head `4e46f10f2c9f381cc4c5422a35e39ff45d3a7098`. Transition
`30225415325` squash-merged PR `#33` as immutable GraphQL merge commit
`b36c3f022a934dd5e1ee676671370986d5292d94`.

Post-merge transition `30225474966` then rejected its plan before apply because GitHub
auto-closed delivery Issue `#27` one second after the merge, changing the native
precondition between reads. After that state settled, run `30225541230` again rejected
before apply because semantically equivalent delivery evidence was incorporated in
provider response order. Neither run emitted a completion receipt or mutated the parent,
dependent, or Project state. The adapter now canonicalizes evidence collections and
derives its rules revision from normalized effective semantics rather than raw rule-array
order. Regression coverage reverses equivalent rules and evidence while requiring one
observation revision, and every unrecognized or missing effective rule type becomes
sorted fail-closed problem evidence that invalidates the revision. Refreshed local Python,
documentation, Go, vet, and race gates pass, and independent Standards and Spec
re-reviews report no findings. Native CI, completion reconciliation, terminal replay, and
cleanup remain pending.

Source `d37a43839d8cb7dfc74994e9c6c8159483c616f9` passed native CI run
`30226344293`. A content-addressed mandate bound to that source then advanced the fresh
episode through branch creation, draft PR creation, ready state, review routing, and
stale-head invalidation. Sandbox PR `#34` now has exact final head
`70752fb69cc6465afdfd25fcfaec139c60519d71` and passed the required
`contract-delivery` check in run `30227492725`. Its approval remains bound to stale head
`08e5a85393710a5c78f82ac3815c3ecd430112cd`; transition run `30227502725`
requested a distinct review of the final head and stopped truthfully at `review-pending`.
No squash merge, completion reconciliation, terminal replay, or cleanup occurred.

The mandate for that episode expired at `2026-07-29T00:00:00Z` and cannot authorize a
later transition or a changed product source. On 2026-08-03 the product branch merged
current `main` at recovery checkpoint `c70a98a`; the conflict was limited to the generated
changelog and documentation index, and refreshed local Python, documentation, Go, vet,
race, release-record, and diff checks pass. The next live run requires a replacement
mandate bound to the recovered product candidate; prior receipts remain historical
evidence and must not be spliced into a newly authorized source.

The owner approved the bounded replacement mandate in issue comment `5168962792` for
source `c8dea833a06cff5dbffe81063ce22379a72034b1`. Exact cleanup plan/apply pairs
`30831793938`/`30831870213`, `30831995590`/`30832069885`,
`30832325384`/`30832380503`, and `30832469783`/`30832548910` then closed
unmerged PR `#34`, deleted its exact branch and marker-owned ruleset, removed both native
relationships, and closed fixture issues `#26`-`#28`; every successful stage retained
receipts and an effect-free converged postcondition. Cleanup-file plan/apply
`30832158398`/`30832206537` emitted no delete effect: the adapter found the exact
marker-owned initial workflow rather than the final-workflow digest required for terminal
cleanup and returned `needs-review`. The live file digest
`sha256:b60607dd2229857a46fe72499687b61d99e7f6f38d0cda448aea7b23b2fd4f85`
matches the source-generated `file-initial` resource exactly, so it remains a deliberate
no-change bootstrap input for the fresh episode and is still subject to terminal cleanup.

Closed PR `#34` permanently owns delivery branch `contract/issue-75-20260721-04` as
historical evidence. Reusing that identity would make the new episode ambiguous, so the
fresh delivery and sandbox generators advance together to
`contract/issue-75-20260721-05`. The resulting product source requires a newly bound
mandate before any fresh setup or delivery effect; the earlier approval does not authorize
a changed source revision.

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
App integration `15368` and allowing only squash merge; the delivery engine creates the
issue-named branch;
`file-candidate` binds the successful branch-creation state artifact and exact branch head
before creating the intermediate PR candidate; and, only after candidate-head check and
review evidence, `file-stale` binds that candidate head before updating the delivery head
to exercise invalidation. Seeder stages use only
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
If branch creation succeeds but no PR can be created, `cleanup-orphan-branch` is the
exclusive recovery route for that branch; it is not combined with `cleanup-delivery`.

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

The owner configured and protected the `contract-delivery` environment with exactly
`CSK_RECONCILER_APP_PRIVATE_KEY` and `CSK_SEEDER_APP_PRIVATE_KEY`. The composed runner
exposes the union of those two keys only within that job and requires an explicit owner
assertion on every dispatch. Reviewer and rules secrets remain absent from that
environment and retain their separate protected routes.
