# Issue #68 — live GitHub contract-test matrix research

**Date:** 2026-07-15  
**Issue:** [#68](https://github.com/dragondad22/codex-starter-kit/issues/68)

## Scope and authority

The owner authorized one bounded research session to finish the GitHub executable-work
decision frontier. The work used current official GitHub documentation and read-only API
observations. It explicitly excluded repository/Project/App/credential/webhook creation,
permission changes, mutating endpoints, paid services, and adapter implementation.

The durable output is
[`GITHUB_LIVE_CONTRACT_TEST_MATRIX.md`](../research/GITHUB_LIVE_CONTRACT_TEST_MATRIX.md).
It allocates the already-approved authentication and Work Manager contracts between live,
in-memory, read-only, not-applicable, not-configured, unsupported, and needs-review
evidence. It does not authorize provisioning or silently establish a product decision.

## Result

The smallest honest qualification topology requires:

- an organization-owned public repository/Project for the preferred App installation
  route;
- a personal public repository/user-owned Project for the explicit classic user-token
  fallback gap;
- a repository-local Actions job for `GITHUB_TOKEN` authority and denial behavior; and
- deterministic fault injection for unsafe or non-repeatable conditions such as rate
  exhaustion, partial GraphQL data, lost responses, offline operation, and webhook replay.

The base route requires no paid GitHub feature and no live webhook receiver. Private,
paid-plan, larger-runner, hosted-webhook, GitHub Enterprise Server, and App-based
user-owned Project claims remain explicitly unqualified.

## Safety, cost, and provisioning disposition

The matrix requires synthetic public fixtures, immutable target allowlists, one mutating
lease, serialized effects, stable run/case markers, reverse-order cleanup, a 24-hour
fixture cleanup target, proposed 30-day raw CI evidence, durable redacted summaries, and
no secrets in plans or artifacts. It never intentionally exhausts GitHub rate limits.

Provisioning remains a separate human approval. The approval must identify sandbox
owners/admins and plans; App owner/managers, selected repository, permissions, secret
store, rotation/revocation, and incident owner; acceptance or rejection of the classic
personal token; optional webhook/private targets; retention/cleanup; spending guardrails;
and the fallback/support limitation if any route is declined.

## Read-only evidence

On 2026-07-15 the selected user-token API actor was `dragondad22`. The operational
repository was public with immutable ID `1297824030`; Project #8 was a public user-owned
Project with 15 fields; the repository exposed one active `Protect main` ruleset; REST
and GraphQL budgets had nonzero capacity; and GitHub reported REST `2026-03-10` and
`2022-11-28` as supported. These observations prove no mutation or least-permission
claim.

## Verification and limitations

The completing PR records the repository documentation tests, document validation, and
Go suite. Official sources were retrieved on 2026-07-15 and are linked in the research
record. Changes to GitHub API versions/schema, Projects token support, permissions,
Actions semantics, rulesets/plans, rates, webhooks, retention, pricing, or target facts
invalidate affected matrix claims.

The research resolves decision-map ticket #4. Ticket #5 must now determine whether the
result is sufficient to decompose feature #4 and must keep actual sandbox provisioning
as an explicit human-gated task.
