# Work Manager — Managed-Task Lifecycle Contract

**Status:** Implemented credential-free lifecycle and deterministic GitHub transport

**Issue:** [#71](https://github.com/dragondad22/codex-starter-kit/issues/71)

**GitHub adapter:** [#72 contract](GITHUB_ADAPTER.md); live sandbox deferred to #73/#76

## Public lifecycle seam

The engine exposes `InspectManagedTask`, `PlanManagedTask`, `ApplyManagedTask`,
`VerifyManagedTask`, and `ManagedTaskStatus`. `ManageTask` composes the same operations
into one request and returns the complete journey. The CLI exposes that composite route as:

```text
starter-kit manage-task --input <managed-task-v1.json>
```

The input is one strict JSON document with `request`, `capability`, and `observation`.
Unknown fields and trailing JSON are rejected. `request.intent` contains no credential;
it records only the expected mode and actor. The in-memory adapter performs no network,
credential lookup, installation, or external effect.

Every boundary is schema version 1:

- `WorkDesiredIntent` binds one stable managed ID to the governed source revision,
  operating-profile revision, input digests, expected actor, immutable target IDs,
  relationships, lifecycle values, Phase, promotion route, review requirements, and
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
  actor, authority, source, observation, target, attempt, retry, recovery, and time facts.

Capability, observation, plan, and effect-result schemas validate required identities,
timestamps, permissions, revisions, outcomes, attempts, and bounded retry ranges. All
adapter data is untrusted; secret-shaped normalized data is rejected before state, and
secret-shaped adapter diagnostics are redacted before receipts.

Go type names are implementation labels, not compatibility authority. Compatibility is
the emitted JSON plus black-box behavior through the CLI and engine seam.

## Policy and adapter ownership

Work Manager derives the effective task before planning. It promotes a blocked task to
Readiness `ready` only when every natively observed blocker is closed, never changes Status
as a side effect of that promotion, inherits Phase from the parent when direct Phase is
absent, maps a closed task to Status `done`, and preserves parent, blocker, promotion, and
distinct-review facts. The adapter observes and attempts semantic effects; it cannot
select policy, credentials, broader authority, or a passing result.
Closed questions require either a durable promotion route or an explicit no-promotion
resolution; closed research requires a durable promoted output. Implementation work
requires a named distinct-context review role.

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

An unchanged refreshed observation produces an effect-free no-change plan while retaining
prior receipts. Apply re-observes capability and task state immediately before effects.
A changed source, operating profile, observation, actor, permission, target,
configuration, field/option identity, or the plan's own expiry makes the plan `stale` and
requires a new inspection and plan.

Capability preconditions bind the semantic authority contract rather than one token
snapshot. A freshly minted credential may have a later observation/expiry timestamp and
different REST or GraphQL counters without invalidating an otherwise unexpired plan.
Actor, mode, account/installation, exact permissions, limitations, evidence mode, target,
API version, and configuration revision remain digest-bound; changes to any of those facts
still stop apply as stale.

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
backstop; #74 consumes that result while owning full intake, subtype, Horizon, Phase, and
Project governance; #75 owns branch/PR/review/gate delivery; and #76 owns aggregate live
qualification, including the live item/parent/dependent reconciliation receipt.

The current draft route manages one selected task plus its bounded parent/direct-dependent
reconciliation slice and discovers native relationships read-only. It does not create
credentials, provision repositories or Projects, configure rules or workflows, publish a
release, or claim private/paid/GHES support. Native Linux, macOS, and Windows support is
claimed only after the exact completing revision passes the repository matrix.
