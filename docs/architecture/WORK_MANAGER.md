# Work Manager — Managed-Task Lifecycle Contract

**Status:** Implemented credential-free lifecycle and deterministic GitHub transport

**Issues:** [#71](https://github.com/dragondad22/codex-starter-kit/issues/71), [#46](https://github.com/dragondad22/codex-starter-kit/issues/46), [#74](https://github.com/dragondad22/codex-starter-kit/issues/74)

**GitHub adapter:** [#72 contract](GITHUB_ADAPTER.md); live sandbox deferred to #73/#76

## Public lifecycle seam

The engine exposes `InspectManagedTask`, `PlanManagedTask`, `ApplyManagedTask`,
`VerifyManagedTask`, and `ManagedTaskStatus`. `ManageTask` composes the same operations
into one request and returns the complete journey. The CLI exposes that composite route as:

```text
starter-kit manage-task --input <managed-task.json>
```

The input is one strict JSON document with `request`, `capability`, and `observation`.
Unknown fields and trailing JSON are rejected. `request.intent` contains no credential;
it records only the expected mode and actor. An external composite request may also carry
its non-secret bounded execution mandate outside desired intent. The in-memory adapter performs no network,
credential lookup, installation, or external effect.

The outer lifecycle evidence remains schema version 1 with additive fields. Desired intent
schema v1 remains readable as historical one-task evidence. Desired intent schema v2 is
the governed executable-work route:

- `WorkDesiredIntent` binds one stable managed ID to the governed source revision,
  operating-profile revision, input digests, expected actor, immutable target IDs,
  relationships, lifecycle values, direct and parent-derived Phase, promotion route, review requirements, and
  desired completion state.
- `WorkCapability` reports online/fresh state, actor/mode, immutable target ownership,
  API version, exact permissions, rate budgets, limitations, evidence mode,
  configuration revision, observation time, and expiry.
- `WorkObservation` contains normalized immutable IDs and semantic task state. Raw HTTP,
  GraphQL, tokens, and transport requests are not retained.
- `WorkPlan` is content-addressed and binds the inspection, desired source, operating
  profile, observation, configuration, actor expectation, target IDs, expiry, impact,
  recovery, and ordered semantic effects.
- `WorkEffectReceipt` preserves each applied or non-pass effect with plan, operation,
  actor, authority, source, observation, target, attempt, retry, recovery, mandate, and time facts.

Capability, observation, plan, and effect-result schemas validate required identities,
timestamps, permissions, revisions, outcomes, attempts, and bounded retry ranges. All
adapter data is untrusted; secret-shaped normalized data is rejected before state, and
secret-shaped adapter diagnostics are redacted before receipts.

Go type names are implementation labels, not compatibility authority. Compatibility is
the emitted JSON plus black-box behavior through the CLI and engine seam.

## Governed executable-work qualification

Schema-v2 intent carries a canonical `ExecutableIssueContract`, exact governed-source
bindings, and any question/research subtype contract. The visible issue uses one versioned
schema marker and exact sections for summary, current context, governing references,
scope/exclusions, acceptance, verification, task-specific authority, and the two Ready
assertions. Governing-reference IDs correspond exactly to safe repository paths and
SHA-256 digests. The adapter parses the visible body; hidden metadata retains only stable
machine projection facts and cannot substitute for missing human sections.

Inspection produces a content-addressed `ManagedWorkQualification` bound to the issue,
source, operating profile, observation, configuration, and immutable target. It is
provenance, not effect authority. One deterministic pre-work disposition is retained:
`fresh`, `mechanical-drift-repaired`, `contained-context-refreshed`,
`needs-refinement`, `already-delivered`, or `blocked`. Planning proceeds only from
`fresh`; the two past-tense repair dispositions are reported only after apply and verify
receipts prove convergence. Source, acceptance, authority, risk, title/type, or
human-owned context changes yield `needs-refinement`. A current-context refresh is
available only when the task contract names the exact stale context-fragment digest as
refreshable and all other semantic sections remain unchanged.
When facts conflict, deterministic precedence is `already-delivered`, then `blocked`, then
`needs-refinement`: exact current delivery prevents duplicate work, a live block prevents
execution, and refinement governs the remaining semantic conflicts.

The check runs at selection/start and after material change. No timestamp participates in
qualification, so age alone cannot make Ready work stale. A matching old qualification is
not a permanent certificate: changed observation or source identity invalidates its plan.
No GitHub comment or standalone pass artifact is required for unchanged fresh work.

Related delivery is not a caller-supplied boolean. The governed GitHub observation reads
bounded issue-timeline cross-references, re-reads same-repository pull requests, and
accepts `already-delivered` only for a merged PR containing one exact, versioned delivery
claim bound to managed ID, source revision, executable-contract digest, and implemented
repository paths and digests. The PR must target the current default branch, its merge
commit must remain reachable from one immutable default-branch head, and every claimed
file digest must still match at that head. The claimed path set must exactly equal the
PR's bounded changed-file manifest; deletion-bearing or empty PRs cannot prove complete
delivery. Historical claims for other governed revisions are ignored. Open, reverted,
removed, different-revision, or otherwise partial claimed delivery returns to refinement.
Issue #75 emits the same delivery claim during branch/PR delivery.

Question and research subtype fields round-trip through the same visible issue contract.
Question relationship and answer authority are explicit. Research objective, intended
use, scope/exclusions, provenance, effort, authority, stopping, output, freshness, and
review are required. Closing promotion must match the governed destination, bind its
exact repository digest, and contain the reciprocal managed issue identity. Question
completion also posts and re-observes one canonical issue comment linking that promoted
record; the promoted record carries a structured, collision-safe backlink to the exact
issue. A no-promotion exception requires a visible closing resolution in the executable
contract rather than a machine-only boolean.

## Policy and adapter ownership

Work Manager derives the effective task before planning. It promotes a blocked task to
Readiness `ready` only when every natively observed blocker is closed, never changes Status
as a side effect of that promotion, reports Phase from the parent when direct Phase is
absent, maps a closed task to Status `done`, and preserves parent, blocker, promotion, and
distinct-review facts. The adapter observes and attempts semantic effects; it cannot
select policy, credentials, broader authority, or a passing result.
Closed questions require either a durable promotion route or an explicit no-promotion
resolution; closed research requires a durable promoted output. Implementation work
requires a named distinct-context review role.

Phase projection uses the immutable `Phase` field and `Phase 0`–`Phase 8` option IDs.
Roadmap features may carry a direct assignment. Ordinary children retain a blank Phase
field while `DerivedFacts` reports the adapter-observed native parent's immutable Phase
option with source `parent`; caller text alone cannot establish that fact. This avoids
copying context onto every child. A non-feature direct assignment is cross-cutting
and requires a durable reason. A child cannot duplicate its parent's value directly.
Unsupported values, missing/stale identities, or an unjustified direct assignment stop
before durable state. Reconciliation sets a justified direct option and clears a copied
option from ordinary child work.

Horizon uses the independent `Now`, `Next`, and `Later` catalog. A Ready feature requires
one direct Horizon when that capability is configured. Ordinary children keep their own
field blank and derive parent context from the native parent observation. Work Manager
sets or clears Horizon by immutable ID without changing Status or Readiness. Derived facts
report both Horizon and Phase capability as `configured` or `not-configured`; saved views
remain optional human-owned presentation, and Milestone remains a separate finite-release
dimension.

For the selected task, the immutable plan derives and applies a bounded reconciliation
slice containing the selected item, its one parent, and its direct dependents. A closed
selected item becomes Status `done`; any started or completed child keeps an incomplete
parent `in-progress`; and all closed children close the parent as `done` only when the
input explicitly confirms that the parent completion contract is satisfied. An
all-children-closed parent without that confirmation is rejected instead of being left
open as an unexplained placeholder.

Each direct dependent supplies governed identity and an explicit Ready-eligibility fact;
the adapter supplies its complete native blocker slice. The final closed blocker promotes
an eligible dependent from `blocked` to `ready`
without selecting Status `next`; any open blocker retains `blocked`. Related corrections
are ordered parent-first and then by dependent managed ID. Their plans and receipts retain
the exact target, operations, semantic before/after lifecycle values, source, observation,
actor, authority, attempt, and result. Completed related effects survive interruption,
and the next plan contains only residual drift.

The adapter refreshes the selected issue's native parent, the parent's complete bounded
sub-issue slice, the selected issue's blockers and direct dependents, and every dependent's
complete blocker slice. A native issue closure cannot be reversed by stale intent, while
governed intent may still request closure; native relationship facts remain authoritative.
The intent retains only governed identities, parent-completion satisfaction, Ready
eligibility, and desired lifecycle policy. Observed parent Status and closure form the
baseline before native child progress derives `backlog`, `in-progress`, or `done`; caller
copies of those lifecycle facts are not trusted. Missing endpoints, stable identities,
Project items, lifecycle options, expected relationships, or exact selected-child membership
produce a non-pass instead of falling back to issue prose or caller-supplied state.
#74 composes this reconciliation path with authoritative issue bodies, subtype completion,
Horizon, Phase, and broader executable-work governance.

Creating a missing task and reconciling its Project/relationship state are separate
effects. A completed create receipt therefore survives a denied or interrupted Project
effect, and the next observation/plan contains only the remaining semantic difference.
An ambiguous create may be re-observed through its stable non-secret marker; discovery
prevents a duplicate create but does not falsely report unobserved Project state complete.
If lookup remains inconclusive, disposition stays `ambiguous` and planning is blocked.

## Persistence, replay, and recovery

The credential-free request, normalized inspection, active plan, receipts, verification,
retry schedule, and disposition are atomically replaced at
`.starter-kit/work-manager/state.json`. The document has its own SHA-256 state digest and
fails closed if altered. It stores neither credentials nor raw transport requests.
Apply first acquires the repository lifecycle lease, so create/verify/work operations and
concurrent work applies cannot race effects or overwrite receipts.

Every prospective external effect, including a schema-v1 compatibility request, requires
`ApplyManagedTaskWithMandate` and a content-addressed DEC-0022
`WorkExecutionMandate`. It binds owner/approval, immutable target, actor and
credential mode, the selected operation and root managed item, exact permissions,
operating profile, input and governed-source digests, the full governance digest including
refreshable-context authority, source revisions, desired-resource digests, managed IDs,
semantic effect/operation classes,
data/cost/destructive ceilings, cumulative effect count, expiry, retention, and recovery
owner. Mandate usage is retained independently of the singleton active operation so
switching work and returning cannot reset its ceiling. The integrity-protected mandate
ledger is also stored outside the replaceable active-work directory, so removing that
directory cannot recreate first-use authority. Each usage slot is durably reserved before
the adapter call, so a crash or lost response cannot restore spent authority. Memory
effects and effect-free plans need
no external mandate. Historical schema-v1 evidence remains readable but cannot authorize
a new unmandated effect.
Once the Work Manager evidence directory exists, a missing, corrupt, or integrity-invalid
state file fails closed; only a genuinely uninitialized directory may create an empty
mandate-usage ledger.

An unchanged refreshed observation produces an effect-free no-change plan while retaining
prior receipts. Apply re-observes capability and task state immediately before effects.
A changed source, operating profile, observation, actor, permission, target,
configuration, field/option identity, or expiry makes the plan `stale` and requires a new
inspection and plan.

Explicit non-pass dispositions include `queued-offline`, `handshake-required`, `denied`,
`ambiguous`, `retry-pending`, `retry-exhausted`, `stale`, and verification `fail`.
Rate-limited receipts retain bounded attempt, retry, and reset times. Reconnect is not
authority: a fresh matching handshake is required. Exhausted retry remains blocked until
the recorded reset passes.

## Evidence boundary and limitations

The in-memory adapter remains the credential-free contract double. The production
`githubadapter` implements the same interface with native REST/GraphQL, an injected
ephemeral credential provider, immutable target IDs, bounded pagination, explicit
transport outcomes, and simulated/live receipt separation. See the
[GitHub adapter contract](GITHUB_ADAPTER.md).

Neither in-memory nor deterministic HTTP-fixture evidence is live GitHub evidence. #73
owns separately approved sandbox provisioning. #15 owns the deterministic reconciliation
backstop; #46 owns governed Phase field, option, and assignment configuration while its
repository-specific receipt also observes an optional human-owned saved view;
#74 consumes those results while owning full intake, subtype, Horizon, Phase, and Project
governance; #75 owns branch/PR/review/gate delivery; and #76 owns aggregate live
qualification, including the live item/parent/dependent reconciliation receipt.

The current route manages one selected task plus its bounded parent/direct-dependent
reconciliation slice, discovers native relationships read-only, and reconciles an existing
Phase assignment by immutable option identity. It does not create credentials, provision
repositories or Projects, configure rules or workflows, publish a release, or claim
private/paid/GHES support. Native Linux, macOS, and Windows support is claimed only after
the exact completing revision passes the repository matrix.

Project-level Phase configuration uses the engine's existing content-addressed external-
resource lifecycle rather than the one-task intent. Its v1 implementation names remain
`Sandbox*` for compatibility, but #46 supplies a distinct operational target, user-token
authority profile, manifest, mandate, and state repository. The plan covers exactly one
Phase catalog, feature #1–#9 assignments, and the repository's observed optional `Phases`
view; it cannot inherit #73's sandbox authority or widen routine Work Manager effects.
Work Manager does not prescribe saved-view presence or layout.
