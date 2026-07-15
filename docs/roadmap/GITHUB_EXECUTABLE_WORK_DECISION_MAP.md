# GitHub Executable-Work Decision Map

**Scope:** GitHub feature [#4](https://github.com/dragondad22/codex-starter-kit/issues/4)
**Status:** Resolved planning record; not product, architecture, policy, or Project authority
**Finish line:** Enough decisions and evidence to decompose #4 into sequenced Ready
sub-issues without implementation-time invention

## #1: What review assurance must executable work require?

Blocked by: None
Type: Grilling
GitHub work item: [#62](https://github.com/dragondad22/codex-starter-kit/issues/62)

### Question

How must PR review remain distinct from implementation and automated checks while still
supporting solo owners and risk-dependent separation of duties?

### Answer

Agreed on 2026-07-14: every PR requires a distinct review pass by a capable reviewer who
did not perform the implementation in the same working context. The reviewer may be a
human or a separate AI context. Automated checks and implementer self-review are
supporting evidence, not the review. Effective policy may require stronger human
independence or qualifications; missing required expertise blocks the affected gate.
The product owner can review outcome alignment without being treated as the code-language
expert. The answer was promoted through `type:question` issue #62 as DEC-0020 and
reconciled with DEC-0008, DEC-0017, DEC-0019, the persona registry, lifecycle, glossary,
and policy ownership.

On 2026-07-14 the accepted `type:question` label was enabled and #62 was created as a
native child of #4. PR #66 promoted the answer, and #62 is now closed and `Done`. This
removes one dependency from #64, which remains `Backlog` / `Blocked` by #63.

## #2: Which GitHub authentication and transport contract should the kit support?

Blocked by: None
Type: Research
GitHub work item: [#63](https://github.com/dragondad22/codex-starter-kit/issues/63)

### Question

Using official GitHub sources and read-only probes, compare GitHub App installation
identity, user-token/`gh` identity, Actions `GITHUB_TOKEN`, API surfaces, permissions,
rate limits, retries, idempotency, offline queues, revocation, cost, native-platform
compatibility, and fallbacks. Which identities and transports belong in the 1.0 supported
contract, and which remain explicit limitations?

### Answer

Resolved on 2026-07-14 by
[the authentication and transport evaluation](../research/GITHUB_AUTHENTICATION_AND_TRANSPORT_EVALUATION.md).
Use three explicit 1.0 modes: GitHub App installation is the preferred unattended and CI
reconciler; a user token is the interactive, personal-account, and recovery route; Actions
`GITHUB_TOKEN` is repository-local CI authority only. Defer GitHub App user tokens. The
production adapter uses native Go HTTP with version-pinned REST plus GraphQL where Project
or relationship coverage requires it; GitHub CLI is optional setup/diagnostics, never a
runtime or identity dependency. Credentials are injected ephemerally and validated against
the API actor, immutable repository/Project owners, and required permission set before
effects. Retries are bounded and rate-aware; ambiguous creates reconcile through stable
managed markers; webhooks are signed, deduplicated hints; offline queues contain
credential-free desired state and always revalidate on reconnect.

Current GitHub behavior prevents a single least-authority identity from covering every
1.0 audience: fine-grained PATs and current user-owned Project REST mutations do not cover
the personal Project route, while classic PAT scope is broader. Full App support therefore
begins with organization-owned Projects after sandbox qualification; the personal-account
route uses an explicitly accepted user token and surfaces its breadth. GitHub Enterprise
Server, App-based user-owned Project mutation, private-repository plan features, and
webhook durability remain explicit limitations until separately qualified. Ticket #3 must
consume this identity, capability, rate, desired-state, and recovery contract; ticket #4
must exercise the exact permission matrix and negative paths.

## #3: What Work Manager boundary makes GitHub reconciliation deterministic?

Blocked by: #1, #2
Type: Prototype
GitHub work item: [#64](https://github.com/dragondad22/codex-starter-kit/issues/64)

### Question

What lifecycle-engine-facing interface and versioned data contract let the Work Manager
own desired issue, hierarchy, readiness, Project, review, and completion state while an
in-memory double and production GitHub adapter own transport? Exercise stale inputs,
partial failure, retries, field-option migration, offline queuing, and recovery without
embedding policy in the adapter.

### Answer

Resolved on 2026-07-14 by the throwaway logic prototype and durable notes delivered
through issue #64. Use four versioned boundaries behind the lifecycle-engine `plan` and
`apply` seam:

1. a credential-free `DesiredIntent` owned by the Work Manager, bound to governed source
   identity and containing stable managed IDs, desired lifecycle/relationship/review
   state, and immutable GitHub target identities;
2. adapter-reported `Capability` and `Observation` values containing the explicit actor,
   mode, minimum permissions, budgets, limitations, immutable GitHub identities,
   configuration revision, and normalized current state;
3. an immutable `Plan` bound to desired-source, observation, repository, Project, field,
   and option identities and containing only the current semantic delta; and
4. per-effect receipts that preserve explicit applied, ambiguous, denied, rate-limited,
   and recovery outcomes across partial failure.

The Work Manager derives readiness, hierarchy, dependency, phase, review, question or
research promotion, and completion policy. The adapter observes and executes effects but
does not choose policy, credentials, or broader authority. Stable non-secret markers
reconcile ambiguous creates; immutable IDs and semantic comparison reconcile updates.
Offline and rate-limited queues retain desired intent, source hashes, and expiry rather
than credentials or raw transport requests. Reconnect always repeats identity,
capability, and precondition checks.

The prototype seeds #15's closed-item Project drift, #16's promoted question result,
#46's inherited Phase context, and #64's distinct review requirement. Its conclusion and
deletion/absorption boundary are recorded in
[`NOTES.md`](../../internal/prototype/workmanager/NOTES.md) and
[`ISSUE-64.md`](../evidence/ISSUE-64.md). Production schema, persistence, adapter, and
sandbox qualification remain downstream work; the prototype JSON and synthetic IDs are
not compatibility promises.

## #4: How will live GitHub behavior be qualified safely?

Blocked by: None
Type: Research
GitHub work item: [#68](https://github.com/dragondad22/codex-starter-kit/issues/68)

### Question

Define the smallest live contract-test matrix that can prove authentication, permissions,
Issues/Project behavior, rulesets, PR lifecycle, rate limits, reconciliation, and recovery
without contaminating an operational repository or overstating unsupported behavior.

### Answer

Resolved on 2026-07-15 by
[`GITHUB_LIVE_CONTRACT_TEST_MATRIX.md`](../research/GITHUB_LIVE_CONTRACT_TEST_MATRIX.md).
The minimum honest topology has two public/free live routes—an organization repository
and organization-owned Project for App-installation support, plus a personal repository
and user-owned Project for the explicit classic user-token fallback—one repository-local
Actions route, and a deterministic fault harness. Live GitHub evidence covers exact
identities, minimum and one-less permissions, semantic Issues/Project/PR/rules effects,
API headers, recovery lookup, and cleanup. The fault harness covers conditions that are
unsafe or non-repeatable to provoke live, including rate exhaustion, partial GraphQL
data, offline operation, and webhook replay; those results never masquerade as live
evidence.

The base matrix uses synthetic public fixtures, standard runners, a selected-repository
private test App, no paid feature, and no live webhook receiver. It specifies immutable
target allowlists, stable run markers, serialized mutation, reverse cleanup, a 24-hour
fixture-cleanup target, proposed 30-day raw evidence, durable redacted summaries, explicit
limitations, and invalidation triggers. Private/paid targets, hosted webhooks, GitHub
Enterprise Server, and App-based user-owned Project mutation remain unqualified.

Provisioning is still a separate external action. Approval must name both sandbox owners
and plans; App owner/managers, permissions, selected repository, secret store,
rotation/revocation and incident owner; acceptance or rejection of the broad classic
personal token; optional targets; retention/cleanup; budget guardrails; and fallback
claims.

## #5: Is the path clear enough to decompose and execute feature #4?

Blocked by: None
Type: Grilling
GitHub work item: [#70](https://github.com/dragondad22/codex-starter-kit/issues/70)

### Question

Do the resolved contracts cover every Phase 3 outcome and negative path well enough to
turn #4 into native, tracer-bullet sub-issues with explicit dependencies, acceptance,
evidence, documentation impact, and readiness—while preserving #15, #16, and #46?

### Answer

Resolved yes on 2026-07-15. Tickets #1–#4 and their promoted records cover every Phase 3
outcome and negative path without requiring implementation-time product, architecture,
policy, regulatory, or risk invention. The product owner approved six tracer-bullet
slices, published as native children of #4:

1. [#71](https://github.com/dragondad22/codex-starter-kit/issues/71) — manage one task
   deterministically through the lifecycle engine;
2. [#72](https://github.com/dragondad22/codex-starter-kit/issues/72) — reconcile one
   managed task through GitHub;
3. [#73](https://github.com/dragondad22/codex-starter-kit/issues/73) — bootstrap an
   isolated GitHub executable-work sandbox after explicit provisioning approval;
4. [#74](https://github.com/dragondad22/codex-starter-kit/issues/74) — govern executable
   work from intake through readiness;
5. [#75](https://github.com/dragondad22/codex-starter-kit/issues/75) — deliver one Ready
   issue through squash completion; and
6. [#76](https://github.com/dragondad22/codex-starter-kit/issues/76) — qualify the GitHub
   executable-work contract and publish truthful evidence.

The dependency order is #71 → #72 → #73; existing #15 and #46 then join #73 as blockers
of #74; #72 and #74 block #75; and #73–#75 block #76. Provisioning within #73 remains a
named human-owned external authority gate. Closed #16 governs question/research semantics
rather than becoming duplicate work. #71 is the single selected Ready item; other
incomplete slices remain Backlog/Blocked until reconciliation promotes them.

This resolves the design frontier. The operational Project and native issue hierarchy
now govern delivery; this document preserves only the reasoning and reciprocal routes.
