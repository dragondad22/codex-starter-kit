# Work Manager — Managed-Task Lifecycle Contract

**Status:** Implemented credential-free production slice  
**Issue:** [#71](https://github.com/dragondad22/codex-starter-kit/issues/71)  
**Live adapter and sandbox qualification:** Deferred to #72–#76

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
- `WorkCapability` reports online/fresh state, actor, mode, exact permissions,
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
Readiness `ready` only when every supplied native blocker is closed, never changes Status
as a side effect of that promotion, inherits Phase from the parent when direct Phase is
absent, maps a closed task to Status `done`, and preserves parent, blocker, promotion, and
distinct-review facts. The adapter observes and attempts semantic effects; it cannot
select policy, credentials, broader authority, or a passing result.

For the selected task, the immutable plan also reports derived parent status/closure from
the supplied parent and sibling facts. This represents the #64 rule that a started child
keeps an incomplete parent `in-progress` and all closed children close it `done`; the
one-task route does not apply a second parent effect. Multi-item reconciliation remains
#74.

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
configuration, field/option identity, or expiry makes the plan `stale` and requires a new
inspection and plan.

Explicit non-pass dispositions include `queued-offline`, `handshake-required`, `denied`,
`ambiguous`, `retry-pending`, `retry-exhausted`, `stale`, and verification `fail`.
Rate-limited receipts retain bounded attempt, retry, and reset times. Reconnect is not
authority: a fresh matching handshake is required. Exhausted retry remains blocked until
the recorded reset passes.

## Evidence boundary and limitations

The in-memory adapter is a production contract double and deterministic development
route. It is not live GitHub evidence. #72 owns the native Go GitHub transport and
identity modes; #73 owns separately approved sandbox provisioning; #74 owns full intake,
hierarchy, subtype, Horizon, Phase, and Project governance; #75 owns branch/PR/review/gate
delivery; and #76 owns aggregate live qualification.

The current route manages one task at a time. It does not create credentials, call
GitHub, provision repositories or Projects, configure rules or workflows, publish a
release, or claim private/paid/GHES support. Native Linux, macOS, and Windows support is
claimed only after the exact completing revision passes the repository matrix.
