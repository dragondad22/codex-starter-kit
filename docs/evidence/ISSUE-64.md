# Issue #64 — deterministic Work Manager prototype

**Date:** 2026-07-14  
**Issue:** [#64](https://github.com/dragondad22/codex-starter-kit/issues/64)

## Outcome and scope

The throwaway Go prototype answers the lifecycle-engine-facing interface question without
performing GitHub, credential, filesystem, or other external effects. A pure reducer in
`internal/prototype/workmanager` owns all state transitions. The terminal-only command in
`cmd/work-manager-prototype` renders the complete in-memory state after each action and is
run with:

```text
go run ./cmd/work-manager-prototype
```

The prototype deliberately narrows its claim to contract learning. It is not the
production Work Manager, adapter, persistence format, compatibility schema, or live
GitHub qualification.

## Interface conclusion

The production boundary should preserve four independently versioned values:

| Value | Owner and purpose |
|---|---|
| `DesiredIntent` | Work Manager policy and desired work state, including stable managed IDs, relationships, lifecycle fields, review roles, promotion/completion evidence routes, governed source identity, and immutable target IDs; contains no credential material |
| `Capability` and `Observation` | Adapter-reported identity, authority, budgets, limitations, immutable GitHub IDs, configuration revision, and normalized current state; partial or stale observations are not success |
| `Plan` | Lifecycle-engine immutable semantic delta with source, observation, repository, Project, field, option, authority, impact, approval, and recovery preconditions |
| `EffectReceipt` | Per-effect applied, ambiguous, denied, rate-limited, or recovery result with plan, operation, actor, authority, source, observation, and target provenance retained across partial failure without claiming distributed atomicity |

The Work Manager derives readiness, parent/child and blocker transitions, Horizon/Phase
projection, review requirements, research/question promotion, and completion. The adapter
does not choose those policies, switch credentials, broaden authority, or turn missing
evidence into a pass. Stable non-secret markers reconcile ambiguous creates. Immutable
IDs, configuration revisions, and semantic comparison reconcile updates.

## Scenario conclusions

The seeded and documented manual paths cover:

- #15-style closed-issue drift: one effect repairs Project Status, and later plans contain
  only the remaining delta;
- #16-style question completion: the promoted durable decision route remains distinct
  from issue and Project state;
- #46 Phase context: the feature carries direct Phase membership while native children
  inherit it instead of receiving copied field values;
- repeated desired state: refreshed observation plus the same desired source converges
  without repeating completed effects;
- lost create response: the result stays `ambiguous` until stable-marker lookup discovers
  exactly one issue, preventing blind duplicate creation while leaving unobserved Project
  and relationship effects in the next plan;
- partial failure and rate limiting: completed observations/receipts survive while only
  credential-free desired intent is queued;
- Project option migration: changed IDs make a plan stale, and only a new governed Work
  Manager source revision accepts the replacement identities;
- desired-source drift: apply rechecks the supplied plan ID, source, observation, target,
  configuration, actor, permission, and rate preconditions before representing an effect;
- offline/reconnect: reconnect is not authority; a fresh actor/capability/precondition
  handshake precedes re-planning; and
- dependency completion: closing #64 promotes its fully unblocked dependent to Readiness
  `Ready` without silently selecting Status `Next`, while incomplete parent #4 remains
  `In progress`.

Review requirements are represented separately from implementation, checks, outcome
approval, assurance approval, and completion, preserving DEC-0020.

## Verification and evidence state

The local repository-scoped Go 1.26.5 toolchain formatted the prototype, compiled and ran
the terminal command, and passed `go test ./...`. The required Python unit suite and
documentation validator also passed. In keeping with issue #64's explicit throwaway
prototype scope, no production-style test suite was added. The terminal harness was
driven through the documented key sequences and the final dispositions were recorded:

| Path | Final disposition |
|---|---|
| complete replay (`p a p a p a p`) | `converged` |
| lost create, marker lookup, re-plan (`p a p l u p`) | `planned` remaining Project/relationship delta |
| two bounded rate-limit attempts (`p a p r h p r h p`) | `rate-limited` |
| observed rate reset and fresh handshake (`... t h p`) | `planned` |
| option migration rejected at apply, then accepted/re-planned (`p m a g h p`) | `planned` |
| offline, reconnect, handshake, plan (`o p n p h p`) | `planned` |
| blocker completion and derived reconciliation (`c p`) | `planned` |
| desired-source drift rejected at apply (`p s a`) | `stale` |

These are attributable manual prototype observations, not production or live-GitHub
qualification. The GitHub Actions matrix remains the native Linux, macOS, and Windows
compilation evidence route for the PR.

No production GitHub adapter or live sandbox claim is covered by this prototype.

## Production absorption and deletion

Production work may lift the four-boundary shape and scenario requirements. It must not
lift the prototype's synthetic node IDs, terse messages, revision suffixes, unvalidated
JSON, in-memory-only state, or deliberate lack of tests. Production still needs schemas,
durable state/evidence, expiry, pagination, partial GraphQL error handling, permission
manifests, bounded retry scheduling, and sandbox-qualified transport behavior.

Delete `cmd/work-manager-prototype` and `internal/prototype/workmanager` after the
production Work Manager and in-memory adapter absorb the prototype-supported boundary. The decision
map, this evidence record, and the later production contract remain the durable result.

## Downstream handoff

The authentication and Work Manager questions are resolved. The next decision-map ticket
can define the smallest live GitHub contract-test matrix, report ownership, visibility,
permission, plan, cost, retention, cleanup, and fallback implications, and request
separate approval before provisioning any sandbox resources.
