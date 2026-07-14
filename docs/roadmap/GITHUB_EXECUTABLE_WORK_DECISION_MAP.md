# GitHub Executable-Work Decision Map

**Scope:** GitHub feature [#4](https://github.com/dragondad22/codex-starter-kit/issues/4)
**Status:** Active planning aid; not product, architecture, policy, or Project authority
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

Unresolved. Issue #64 is a native child of #4 and is `Backlog` / `Blocked`. Its #62
dependency is complete; #63 remains its active native blocker. The prototype must use the
lifecycle-engine seam and incorporate existing children #15, #16, and #46 rather than
duplicating or bypassing them.

## #4: How will live GitHub behavior be qualified safely?

Blocked by: #2, #3
Type: Research

### Question

Define the smallest live contract-test matrix that can prove authentication, permissions,
Issues/Project behavior, rulesets, PR lifecycle, rate limits, reconciliation, and recovery
without contaminating an operational repository or overstating unsupported behavior.

### Answer

A dedicated Starter Kit sandbox repository and Project is the agreed proposed boundary;
most behavior remains covered by the in-memory adapter. Exact ownership, visibility,
plan-dependent features, identity, permissions, fixture retention, cost, cleanup, and
fallback remain unresolved until #2 and #3 complete. Provisioning is a separate external
action requiring explicit approval after those implications are reported.

## #5: Is the path clear enough to decompose and execute feature #4?

Blocked by: #3, #4
Type: Grilling

### Question

Do the resolved contracts cover every Phase 3 outcome and negative path well enough to
turn #4 into native, tracer-bullet sub-issues with explicit dependencies, acceptance,
evidence, documentation impact, and readiness—while preserving #15, #16, and #46?

### Answer

Unresolved. If yes, publish the decomposition to GitHub, reconcile every child and the
parent Project fields, remove `needs-triage`, and drive one Ready issue at a time. If no,
add only the newly discovered frontier tickets to this map.
