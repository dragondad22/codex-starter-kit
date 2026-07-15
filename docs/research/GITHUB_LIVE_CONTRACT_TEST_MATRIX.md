# GitHub live contract-test matrix

**Status:** Bounded research result; not provisioning approval or product authority  
**Issue:** [#68](https://github.com/dragondad22/codex-starter-kit/issues/68)  
**Retrieved:** 2026-07-15  
**Intended use:** Final decomposition of GitHub feature #4 and later adapter qualification

## Research boundary

This record defines the smallest evidence matrix that can qualify the GitHub adapter
contract without mutating the operational Starter Kit repository or claiming behavior
that was never exercised. It uses current official GitHub documentation and read-only
observations of the existing personal repository and Project. It did not create a
repository, Project, App, credential, webhook, paid feature, or mutating API request.

The authentication/transport contract is governed by
[`GITHUB_AUTHENTICATION_AND_TRANSPORT_EVALUATION.md`](GITHUB_AUTHENTICATION_AND_TRANSPORT_EVALUATION.md),
and the lifecycle-facing desired-state boundary is governed by the issue #64 conclusion
in the [decision map](../roadmap/GITHUB_EXECUTABLE_WORK_DECISION_MAP.md). This research
allocates evidence between those contracts; it does not replace them.

## Conclusion

One sandbox cannot prove every supported 1.0 identity honestly. The minimum topology is:

| Target | Purpose | Base disposition |
|---|---|---|
| **O — organization route**: dedicated public repository plus organization-owned Project; App installed only on that repository | Preferred unattended `app-installation` identity; organization Project permission; repository Issues, PR observation, rules, and cleanup | Required live target before App support is promoted |
| **P — personal route**: dedicated public personal repository plus user-owned Project | Interactive/recovery `user-token` route and the documented user-Project gap; use an explicitly accepted classic token only where the fine-grained/App routes are unsupported | Required live target before the personal route is promoted |
| **A — Actions route**: one minimal workflow in O or P | Repository-local `GITHUB_TOKEN`, least job permissions, denied Project and cross-repository behavior, workflow-trigger semantics | Required live target for CI authority claims |
| **F — deterministic fault harness**: in-memory adapter plus controllable HTTP double | Rate exhaustion, partial GraphQL data, lost response, offline/reconnect, webhook signature/GUID replay, time, and retry scheduling | Required repeatable evidence; never relabeled as live GitHub evidence |

The base matrix uses public repositories and standard GitHub-hosted runners. Current
GitHub documentation says standard Actions usage is free for public repositories and
repository rulesets are available for public repositories on GitHub Free. Private,
internal, paid-plan, organization-wide ruleset, larger-runner, hosted webhook, and GitHub
Enterprise Server claims remain `not-configured` or `unsupported` until a separately
approved target is qualified.

## Evidence vocabulary

Every case has exactly one disposition; blank evidence is not a pass.

| State | Meaning |
|---|---|
| `memory` | Deterministic fake/fault evidence proves product logic, not GitHub behavior |
| `live` | A retained sandbox receipt proves the exact identity, owner kind, visibility, plan, permission, endpoint, and API version |
| `read-only` | A current non-mutating probe informs the plan but cannot qualify a mutation |
| `not-applicable` | A recorded rule and facts prove the case irrelevant to the declared target |
| `not-configured` | The optional target, identity, service, or feature was not provisioned |
| `unsupported` | The platform or supported contract does not provide the route |
| `needs-review` | Evidence is conflicting, stale, partial, ambiguous, or requires human/qualified disposition |

The release qualification for a supported claim requires every `memory` and `live` row
assigned to that claim. A read-only probe, documentation link, or successful broader token
does not substitute for the live minimum-permission and one-less-permission cases.

## Identity and permission manifest

The production endpoint inventory generates the final manifest. The sandbox creates a
token for one case at a time; it does not maintain one permanently broad credential.

| Capability | App/user authority to qualify | Required negative case |
|---|---|---|
| Repository and actor handshake | Metadata read; explicit actor/App and installation; selected repository; immutable owner/repository ID | Wrong actor, wrong installation/account, unselected repository, expired/revoked/suspended token |
| Issue, label, hierarchy, and dependency reconciliation | Repository `Issues: write` | `Issues: read` or omitted; verify denial and no state change |
| Project read/reconciliation in O | Organization `Projects: read`/`write`; repository Issues/PR read only when the operation needs them | Repository Project permission alone, Projects read-only for mutation, stale Project/field/option IDs |
| Project reconciliation in P | Explicit user token with classic `project` scope; add repository scope only as required by target visibility/effects | Fine-grained PAT, App user token, and App installation token remain unsupported for user-owned Project item endpoints unless GitHub changes and the exact case passes |
| Ruleset inspection/application | Repository `Administration: read`/`write`; application is a separate high-authority approved effect | Administration read-only/omitted; disabled versus active rule; no bypass inference |
| PR/review/check observation | Pull requests, Checks, commit statuses, and Actions read only for the endpoint inventory | Missing one required read permission; ruleset/check success must not be treated as proof of reviewer capability or independence |
| PR metadata mutation, if included | `Pull requests: write` only | No Contents, Workflows, merge, or bypass authority inferred |
| Repository webhook configuration/redelivery, if enabled | `Webhooks: write` as a separately approved optional capability | Built-in `GITHUB_TOKEN` denial; missing/invalid signature; duplicate delivery GUID |
| Actions job | Explicit job `permissions`; normally `contents: read` plus the one repository-local write under test | Project access and another-repository access denied; no recursive-workflow assumption |

GitHub publishes a per-endpoint App permission table and returns
`X-Accepted-GitHub-Permissions` for REST troubleshooting. GraphQL permission behavior must
still be tested explicitly. A passing broad credential establishes no least-authority
claim.

### Source conflict and uncertainty

GitHub's general Projects API guide says its examples can use a classic personal token or
an App installation token. Its Actions guide is more specific: `GITHUB_TOKEN` cannot
access Projects, an App is recommended for organization Projects, and a personal token
is recommended for user Projects. The current Project-item REST reference is stricter
still: organization endpoints list App installation, App user, and fine-grained tokens,
while every user-owned item endpoint explicitly rejects those three token types. The
matrix therefore follows the endpoint-specific contract and requires separate O and P
evidence. It must be revisited if GitHub reconciles these surfaces or GraphQL behaves
differently.

The read-only user API response did not establish the current account plan. Public/free
eligibility is a documented base target, not an observation that this account or a future
organization is on a particular plan. Provisioning must record the actual plan before a
run. GraphQL has no REST-style date-version header, so schema/behavior freshness remains
an explicit qualification input even while REST is pinned.

## Contract matrix

### Identity, capability, and version

| ID | Claim and action | Target/evidence | Negative or stopping condition |
|---|---|---|---|
| `GH-ID-01` | Handshake records host, REST/GraphQL versions, mode, immutable actor/install/account, repository/Project owner kind and IDs, granted versus required permissions, feature/plan limits, and budgets | O/P/A `live`; F schema validation | Any mismatch is `unsupported` or `needs-review`; never switch credentials automatically |
| `GH-ID-02` | App installation token expires/refreshes and remains selected-repository scoped | O `live` | Expired, revoked, suspended, or uninstalled App retains desired intent and reports exact recovery; no retry loop |
| `GH-ID-03` | User token authenticates the expected API login and accesses only the accepted personal route | P `live` | Wrong active account, expired/revoked token, fine-grained/App user-Project rejection |
| `GH-ID-04` | Actions token is unique to the job, repository-local, and least-permission | A `live` | Project and cross-repository requests denied; token-created workflow effects follow documented trigger limitations |
| `GH-ID-05` | REST sends `X-GitHub-Api-Version: 2026-03-10`; adapter reports supported range | O/P/A `live`; version list `read-only` | Unsupported version returns explicit version failure; no unversioned fallback |

### Issues, relationships, Project, and replay

| ID | Claim and action | Target/evidence | Negative or stopping condition |
|---|---|---|---|
| `GH-WORK-01` | Create one marked issue, re-observe immutable ID, replay unchanged desired state | O and P `live`; F `memory` | Exactly one issue exists; marker collision or multiple matches is `needs-review` |
| `GH-WORK-02` | Update title/body/labels, close/reopen, and re-run semantic comparison | O and P `live`; F `memory` | Formatting/order normalization does not cause loops; omitted permission changes nothing |
| `GH-WORK-03` | Create parent/child and blocker/dependent relationships, then remove/reapply | O and P `live`; F derives lifecycle | Native relationships match stable IDs; missing blocker prevents Ready; completing final blocker promotes Ready but not Status Next |
| `GH-WORK-04` | Add issue/PR to Project, update Status/Readiness and one ordinary field by immutable IDs | O and P `live` with their distinct identities | Name-only lookup is rejected; read-only permission and wrong owner kind are denied |
| `GH-WORK-05` | Recreate one Project option/field in an isolated fixture and attempt an old plan | O and P `live`; F `memory` | Old field/option/configuration identity invalidates before effects and produces a new plan |
| `GH-WORK-06` | Inject response loss after a successful live issue create; reconcile by stable marker | O and P `live` through faulting client shim; F `memory` | No blind create retry; zero or multiple lookup matches remain ambiguous |
| `GH-WORK-07` | Apply a multi-effect plan with one accepted effect followed by injected failure | F `memory`, then O/P live postcondition reads | Completed receipt/observation remains; next plan contains only remaining semantic delta |
| `GH-WORK-08` | Close a linked work item and reconcile item, parent, and direct dependents | O and P `live`; F `memory` | Reproduces the #15 class and proves closed→Done plus parent/dependency rules |
| `GH-WORK-09` | Complete question and research fixtures | F `memory`; O/P `live` issue/Project projection | Cannot close without the required promoted-record/durable-output route; issue text never becomes authority |

### PR, checks, review, and rules

| ID | Claim and action | Target/evidence | Negative or stopping condition |
|---|---|---|---|
| `GH-PR-01` | Observe pre-created branch/PR through draft, ready, checks pending/pass/fail, changes requested, and merged/closed states | O and P `live`; branch content is fixture setup outside default Work Manager authority | Missing or stale check/review evidence blocks the modeled gate; head SHA change invalidates review evidence |
| `GH-PR-02` | Record implementation, checks, distinct change review, product outcome approval, and stronger assurance as separate roles/results | F `memory`; O/P live GitHub observations | A ruleset, check, implementer self-review, or broad approval is not relabeled as a capable/distinct/qualified review |
| `GH-RULE-01` | Read a disabled fixture ruleset, separately approve activation, observe enforcement on a fixture branch, then disable/delete | O `live`; P `live` public/free | Read permission cannot mutate; no operational branch target; bypass is tested only if separately approved |
| `GH-RULE-02` | Compare effective layered rule observation with desired repository rule | O/P `live` | Plan/visibility feature unavailable becomes `not-configured`/`unsupported`, not a paid-upgrade assumption |

### Transport, rate, webhook, offline, and recovery

| ID | Claim and action | Target/evidence | Negative or stopping condition |
|---|---|---|---|
| `GH-TRANS-01` | Follow REST `Link` and GraphQL cursors with `first`/`last` 1–100; preserve immutable IDs | O/P `live` with enough isolated fixtures; F boundary cases | Missing page or duplicate cursor fails; partial GraphQL `data` plus required-field `errors` is not success |
| `GH-TRANS-02` | Preserve 401, 403, 404-for-hidden-resource, validation, and GraphQL error distinctions | O/P/A one-less/revoked cases `live`; other responses F `memory` | No credential broadening or guessing that a 404 proves absence |
| `GH-TRANS-03` | Record primary/GraphQL budgets and reset headers from ordinary requests | O/P/A `live` | Do **not** exhaust live limits; a point-in-time budget is not a rate-recovery pass |
| `GH-TRANS-04` | Inject primary exhaustion, secondary limit with/without `retry-after`, reset, exponential delay, and maximum attempts | F `memory` | Serialize mutations, normally pause ≥1 second, stop at bound; never intentionally provoke GitHub limits or integration bans |
| `GH-WEB-01` | Validate official HMAC-SHA-256 vectors, invalid/missing signature, GUID dedupe, and replay hint | F `memory` required | Webhook remains optional `not-configured` live; payload is a hint and always triggers authoritative read |
| `GH-WEB-02` | If a receiver is separately approved, observe signed delivery, duplicate redelivery, failure, and recent-delivery lookup | O optional `live` | GitHub does not auto-redeliver failures and exposes recent deliveries for only three days; polling/full reconciliation remains required |
| `GH-OFF-01` | Queue credential-free desired intent while offline, expire/stale it, reconnect, reacquire credential, repeat handshake/read/plan | F `memory`; O/P reconnect handshake `live` | No raw HTTP, token, inferred current readiness, or automatic identity switch in the queue |

## Fixture and isolation contract

- Sandbox repository and Project names start with `codex-starter-kit-sandbox`; every run
  has an immutable run ID and every created object has a stable non-secret
  `starter-kit-contract:<run-id>:<case-id>` marker.
- No operational repository, Project, branch, ruleset, App, credential, webhook, or issue
  is a target. The adapter rejects target IDs not present in the signed/approved run
  manifest.
- Only one mutating run holds the sandbox lease. Mutations are serialized. Tests use
  synthetic public content and no production issue bodies, secrets, or user data.
- The run manifest records baseline IDs and expected cleanup. Cleanup deletes or closes
  marked fixtures in dependency-safe reverse order, re-reads the sandbox, and fails if
  any unrecognized or residual managed object remains.
- Keep the sandbox repositories/Projects for repeatability; clean per-run fixtures within
  24 hours. Repository deletion is a separately approved destructive action, not routine
  cleanup. GitHub notes that some deleted repositories are restorable for 90 days, so
  deletion is not an evidence-erasure guarantee.
- App keys and tokens are never fixtures. Inject them ephemerally from an approved secret
  provider, record only identity/permission/expiry facts, revoke test tokens when the
  run ends where supported, and never upload them in artifacts.

## Evidence and retention contract

Each case emits a machine receipt with: schema/run/case ID; source commit and adapter,
engine, desired-state, and API versions; timestamp; target host/owner kind/visibility/plan
and immutable IDs; credential mode and verified actor/installation (never the token);
required/observed permissions; endpoint operation and semantic request digest; redacted
status/error/rate headers; before/after IDs and normalized digests; effect attempts;
cleanup result; disposition; limitations; and linked raw-evidence digest.

Durable repository evidence keeps the matrix, concise run summary, limitations, and
content-addressed receipt digests. Proposed raw Actions artifacts/logs retain for 30 days,
matching this repository's existing native-evidence practice; per-run fixtures clean
within 24 hours. GitHub's default workflow artifact/log retention is 90 days and may be
configured from 1–90 days for public repositories, so the workflow must set the intended
value explicitly. Any longer retention or sensitive/private target requires effective
policy review rather than inheriting this proposal.

## Cost, authority, and compatibility implications

| Implication | Base recommendation and fallback |
|---|---|
| Ownership | O requires an organization owner/App manager and Project administrator; P requires the personal owner. If those roles are unavailable, affected live evidence is `not-configured`, not delegated to a broader token. |
| App trust | Register one private test App, selected-repository installation, no public listing, minimum permissions, and webhooks disabled unless GH-WEB-02 is separately selected. Record owner, managers, key rotation/revocation, and incident contact before use. |
| User token breadth | The user-owned Project REST route currently rejects fine-grained and App tokens. A classic token is an accepted broad fallback only after its scopes, expiry, storage, revocation, actor, and personal-account implications are approved. Otherwise P is unsupported. |
| Data | Public synthetic fixture content only. No operational issue text, source, credentials, private metadata, or sensitive evidence enters the sandbox. |
| Cost | Public repositories with standard hosted runners have no Actions-minute charge under current docs. No paid upgrade, private target, larger runner, hosted receiver, or secret service is required for the base matrix. Usage and pricing remain invalidation triggers. |
| Compatibility | Base target is GitHub.com REST `2026-03-10` plus current GraphQL. GitHub Enterprise Server, private/paid behavior, App-based user-owned Project mutation, and webhook durability remain unsupported/not-configured until exact qualification. |
| Fallback | Production can remain user-token interactive for the personal route and in-memory/polling-first where App/webhooks are unavailable. Missing live evidence narrows the support claim; it does not block deterministic local logic or justify broader authority. |

## Provisioning gate

This research does not authorize provisioning. Before any live setup, the owner must
approve the exact run manifest and confirm:

1. organization and personal sandbox owners, administrators, public visibility, and
   GitHub plan;
2. App owner/managers, selected repository, exact permission manifest, key/secret store,
   expiry, rotation, suspension/revocation, and incident responsibility;
3. whether the classic personal token fallback is accepted and, if so, its exact scopes,
   expiry, storage, and revocation path;
4. whether optional live webhooks or any private/paid target are included, with hosting,
   data, retention, and cost implications;
5. fixture prefix, lease, 24-hour cleanup, 30-day raw evidence retention, and residual
   resource owner; and
6. spending/budget guardrails and the fallback/support claim if any target is declined.

If approval is declined, decomposition can still implement the in-memory Work Manager and
adapter boundary, but App, personal-Project, or optional webhook claims remain explicitly
unqualified.

## Read-only observation snapshot

Observed on 2026-07-15 with no mutating endpoint:

- API actor `dragondad22`; public personal repository immutable ID `1297824030`; current
  actor reported admin/maintain/push/triage/pull permissions.
- User-owned public Project #8 immutable ID `PVT_kwHOASd_cc4BdI9q`, open, with 15 fields
  including immutable Status, Horizon, Readiness, parent, and sub-issue-progress IDs.
- One active repository ruleset named `Protect main`; this proves read access only.
- REST core and GraphQL budgets were 5,000-point resources with nonzero remaining values.
- GitHub reported REST versions `2026-03-10` and `2022-11-28` as supported.

This snapshot does not prove plan, App, one-less-permission, mutation, retry, cleanup, or
private-repository behavior.

## Freshness and invalidation

Re-run official-source review and affected cases when GitHub changes a REST version,
GraphQL schema, Projects endpoint/token support, App permission, Actions token semantics,
ruleset/plan availability, rate guidance, webhook delivery behavior, artifact retention,
pricing, or when the target owner/visibility/plan/App installation/permission manifest
changes. A source or target change invalidates only the affected evidence but prevents
the broader claim until it is refreshed.

## Primary official sources

- [Permissions required for GitHub Apps](https://docs.github.com/en/rest/authentication/permissions-required-for-github-apps)
- [Choosing permissions for a GitHub App](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/choosing-permissions-for-a-github-app)
- [Authenticating as a GitHub App installation](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation)
- [Project item REST endpoints](https://docs.github.com/en/rest/projects/items)
- [Using the API to manage Projects](https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/using-the-api-to-manage-projects)
- [Automating Projects using Actions](https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/automating-projects-using-actions)
- [`GITHUB_TOKEN`](https://docs.github.com/en/actions/concepts/security/github_token)
- [Using `GITHUB_TOKEN` in workflows](https://docs.github.com/en/actions/security-for-github-actions/security-guides/automatic-token-authentication)
- [REST API best practices](https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api)
- [GraphQL rate and query limits](https://docs.github.com/en/graphql/overview/rate-limits-and-query-limits-for-the-graphql-api)
- [REST API versions](https://docs.github.com/en/rest/about-the-rest-api/api-versions)
- [About repository rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets)
- [Validating webhook deliveries](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries)
- [Handling failed webhook deliveries](https://docs.github.com/en/webhooks/using-webhooks/handling-failed-webhook-deliveries)
- [Viewing webhook deliveries](https://docs.github.com/en/webhooks/testing-and-troubleshooting-webhooks/viewing-webhook-deliveries)
- [GitHub Actions billing](https://docs.github.com/en/billing/concepts/product-billing/github-actions)
- [Actions artifact and log retention](https://docs.github.com/en/organizations/managing-organization-settings/configuring-the-retention-period-for-github-actions-artifacts-and-logs-in-your-organization)
- [Deleting a repository](https://docs.github.com/en/repositories/creating-and-managing-repositories/deleting-a-repository)
- [Token expiration and revocation](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/token-expiration-and-revocation)
