# GitHub live contract-test matrix

**Status:** Approved organization-only 1.0 qualification topology; provisioning owned by #73
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

The product owner narrowed the required Phase 3 / 1.0 topology on 2026-07-16. Required
qualification uses one dedicated GitHub Free organization; the personal user-owned
Project/classic-token route is not a 1.0 support claim and must remain `not-configured`
or `unsupported` until separately approved and qualified. This supersedes the broader
research proposal below wherever it described P as required.

The approved minimum topology is:

| Target | Purpose | Base disposition |
|---|---|---|
| **O — organization route**: dedicated public repository plus organization-owned Project; App installed only on that repository | Preferred unattended `app-installation` identity; organization Project permission; repository Issues, PR observation, rules, and cleanup | Required live target before App support is promoted |
| **P — personal route** | Deferred; no personal repository, user-owned Project, or classic PAT is provisioned | `not-configured` for 1.0; requires a separate future approval |
| **A — Actions route**: one minimal workflow in O | Repository-local `GITHUB_TOKEN`, least job permissions, denied Project and cross-repository behavior, workflow-trigger semantics | Required live target for CI authority claims |
| **F — deterministic fault harness**: in-memory adapter plus controllable HTTP double | Rate exhaustion, partial GraphQL data, lost response, offline/reconnect, webhook signature/GUID replay, time, and retry scheduling | Required repeatable evidence; never relabeled as live GitHub evidence |

The base matrix uses public repositories and standard GitHub-hosted runners. Current
GitHub documentation says standard Actions usage is free for public repositories and
repository rulesets are available for public repositories on GitHub Free. Private,
internal, paid-plan, organization-wide ruleset, larger-runner, hosted webhook, and GitHub
Enterprise Server claims remain `not-configured` or `unsupported` until a separately
approved target is qualified.

## Evidence vocabulary

Evidence mode and result are independent. A live request may fail, a deterministic case
may need review, and an optional live target may be not configured. Every receipt records
at least one required evidence mode and exactly one result; blank evidence is never a
pass.

| Evidence mode | Meaning |
|---|---|
| `memory` | Deterministic fake/fault evidence proves product logic, not GitHub behavior |
| `live` | A retained sandbox receipt exercises the exact identity, owner kind, visibility, plan, permission, endpoint, and API version |
| `read-only` | A current non-mutating probe informs the plan but cannot qualify a mutation |

| Result | Meaning |
|---|---|
| `pass` | Required observations and postconditions were proved by the required mode |
| `fail` | An exercised required behavior or postcondition did not hold |
| `not-applicable` | A recorded rule and target facts prove the case irrelevant |
| `not-configured` | The required target, identity, service, fixture, or feature was not provisioned |
| `unsupported` | An authoritative capability check proves the declared platform route is unavailable |
| `needs-review` | Evidence is conflicting, stale, partial, ambiguous, or requires qualified disposition |
| `accepted-exception` | A named authority accepts a result that remains `fail`, `not-configured`, or `needs-review`; the underlying result and expiry remain visible |

All rows below currently have result `not-configured`: this issue designed the matrix but
did not provision or execute it. The read-only snapshot is research input, not a
qualification result. Release qualification for a supported claim requires `pass` for
every required `memory` and `live` mode assigned to it. A documentation link, read-only
probe, accepted exception, or successful broader token does not substitute for a required
minimum-permission or one-less-permission pass.

## Identity, permission, and endpoint manifest

The base manifest is fixed rather than inferred at runtime. Each case mints or selects the
narrow identity named below. Seeder and optional webhook authority are outside Work
Manager authority and require their own approved operation. An endpoint absent from the
manifest must not be called; adding one invalidates the manifest and its least-authority
evidence.

| Identity | Exact positive authority | Required one-less or denied case |
|---|---|---|
| O reconciler App installation | Selected O repository; repository Metadata read, Issues write, Pull requests read, Checks read, Actions read, Commit statuses read; organization Projects write | Omit Issues for issue mutation; Projects read for Project mutation; unselected repository; wrong installation; expired/revoked/suspended installation token |
| O rules inspector | Selected O repository; Metadata and Administration read | Administration omitted cannot inspect |
| O rules applier | One-operation selected-repository token; Metadata read and Administration write | Inspector token cannot mutate; token is destroyed after disable/delete cleanup |
| P reconciler classic user token | Not provisioned or required for 1.0 | `not-configured`; no classic-token fallback may be selected automatically |
| A issue job | `contents: read`, `issues: write`, every other job permission `none` | Project and other-repository access denied; a second job with `issues: read` cannot mutate |
| A observation job | `contents: read`, `pull-requests: read`, `checks: read`, `statuses: read`, `actions: read`, every other job permission `none` | No issue, Project, ruleset, review submission, merge, contents-write, or workflow-write authority |
| Fixture seeder App installation | Separate private App installed only on the O sandbox repository; Metadata read, Contents write, Pull requests write, Workflows write, and Issues write | Seeder installation credential is unavailable to Work Manager, Actions jobs, and reviewer; omitted Workflows permission cannot seed/change the fixture workflow |
| Distinct test reviewer | Dedicated user account, different from sandbox owner, seeder App managers, and implementation actor; repository role Write; selected-repository fine-grained token with Metadata/Contents read and Pull requests write | Pull requests read cannot submit a review; same-actor or stale-head review is observed but cannot satisfy distinct review |
| Optional webhook manager | Separate one-operation identity with Metadata read and Webhooks write | Omitted by default; built-in `GITHUB_TOKEN` cannot configure it |

| Endpoint family | Allowed operations |
|---|---|
| Handshake | REST actor, repository, installation/repositories, rate-limit, API-version metadata; GraphQL viewer/rateLimit and repository/Project immutable IDs |
| Issues | REST labels and issue create/read/update; native sub-issue and dependency create/read/delete endpoints |
| Projects | Versioned REST Project fields and items for O/P owner routes, plus views only when an explicit team-specific manifest requests them; GraphQL Project V2 item/field mutations only where the REST inventory lacks the required operation |
| Pull requests and gates | REST/GraphQL read of PR, reviews, review requests, checks, statuses, workflows/runs, branch and merge metadata |
| Rules | REST repository ruleset read/create/update/delete against fixture branches only |
| Webhooks | REST repository webhook create/read/delivery/redelivery/delete only when `GH-WEB-02` is separately approved |

The manifest forbids merge, branch protection bypass, repository administration other
than isolated ruleset operations, repository deletion, member administration, Secrets,
and Actions variable/secret access. Project views are optional, human-owned presentation;
the fixture owner may preprovision them for observation, but their presence or layout is
not a universal qualification requirement. Built-in auto-add/close workflows are
preprovisioned by the fixture owner because the documented public API surface does not
establish a complete workflow-configuration contract.

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

`Current result` is `not-configured` for every row because no qualification environment
was provisioned. The required-evidence column describes what a later run must produce;
it is not a claim that the behavior has already been observed.

### Identity, capability, and version

| ID | Claim and action | Required evidence | Current result | Negative or stopping condition |
|---|---|---|---|---|
| `GH-ID-01` | Handshake records host, REST/GraphQL versions, mode, immutable actor/install/account, repository/Project owner kind and IDs, granted versus required permissions, feature/plan limits, and budgets | O/P/A `live`; F `memory` schema validation | `not-configured` | A mismatch is `fail`, `unsupported` only after an authoritative capability check, or `needs-review` if ambiguous; never switch credentials automatically |
| `GH-ID-02` | App installation token expires/refreshes and remains selected-repository scoped | O `live` | `not-configured` | Expired, revoked, suspended, or uninstalled App retains desired intent and reports exact recovery; no retry loop |
| `GH-ID-03` | User token authenticates the expected API login and accesses only the accepted personal route | P `live` | `not-configured` | Wrong active account, expired/revoked token, fine-grained/App user-Project rejection |
| `GH-ID-04` | Actions token is unique to the job, repository-local, and least-permission | A `live` | `not-configured` | Project and cross-repository requests denied; token-created workflow effects follow documented trigger limitations |
| `GH-ID-05` | REST sends `X-GitHub-Api-Version: 2026-03-10`; adapter reports supported range | O/P/A `live`; supported-version inventory `read-only` | `not-configured` | Unsupported version returns explicit version failure; no unversioned fallback |

### Bootstrap, issue rendering, readiness, and lifecycle projection

| ID | Claim and action | Required evidence | Current result | Negative or stopping condition |
|---|---|---|---|---|
| `GH-BOOT-01` | Reconcile the exact managed label vocabulary, create a missing managed label, correct managed color/description drift, and preserve an unrecognized human label | O/P `live`; F `memory` | `not-configured` | No deletion or rewrite of unrecognized labels; collision or permission denial produces no partial bootstrap pass |
| `GH-BOOT-02` | Inspect exact Project fields/options and optional human-owned views, recreate one isolated field/option, and refresh immutable configuration IDs | O/P `live`; F `memory` | `not-configured` | Missing or duplicate governed field, wrong type, or stale ID invalidates the configuration and any old plan before effects; view variation is informational unless an explicit manifest governs that view |
| `GH-BOOT-03` | Prove the preprovisioned auto-add filter and close-to-Done workflow with marked and unmarked issues | O/P `live` | `not-configured` | Existing matching items are not assumed to backfill; an unmarked issue must not be added; workflow configuration remains human-owned |
| `GH-ISSUE-01` | Render and parse the two-layer issue contract: concise human summary plus validated machine details and governed source references | F `memory`; O/P `live` round trip | `not-configured` | Missing summary, invalid/missing details, unknown schema, or unresolved source reference is `fail` and blocks mutation/readiness |
| `GH-READY-01` | Refresh readiness from current dependencies, governed source, policy facts, and Project state before plan and apply | F `memory`; O/P `live` observations | `not-configured` | Stale/missing source, facts, dependency, Project configuration, or accepted plan returns `Needs refinement`/new plan rather than guessing |
| `GH-QRES-01` | Complete question and research fixtures only after their required answer/output is promoted to its governed record and linked from the issue | F `memory`; O/P `live` issue/Project projection | `not-configured` | Issue prose is not authority; missing promotion, validation, source link, or qualified acceptance prevents completion |
| `GH-HORIZON-01` | Intake a feature, assign Now/Next/Later Horizon, inherit Phase to children, and promote Horizon without conflating Status or Readiness | F `memory`; O/P `live` Project projection | `not-configured` | Child Phase conflict, invalid promotion, or Status/Readiness used as Horizon is `fail` |

### Issues, relationships, Project, and replay

| ID | Claim and action | Required evidence | Current result | Negative or stopping condition |
|---|---|---|---|---|
| `GH-WORK-01` | Create one marked issue, re-observe immutable ID, replay unchanged desired state | O/P `live`; F `memory` | `not-configured` | Exactly one issue exists; marker collision or multiple matches is `needs-review` |
| `GH-WORK-02` | Update title/body/labels, close/reopen, and re-run semantic comparison | O/P `live`; F `memory` | `not-configured` | Formatting/order normalization does not cause loops; omitted permission changes nothing |
| `GH-WORK-03` | Create parent/child and blocker/dependent relationships, then remove/reapply | O/P `live`; F `memory` lifecycle derivation | `not-configured` | Native relationships match stable IDs; missing blocker prevents Ready; completing final blocker promotes Ready but not Status Next |
| `GH-WORK-04` | Add issue/PR to Project, update Status/Readiness and one ordinary field by immutable IDs | O/P `live` with distinct identities | `not-configured` | Name-only lookup is rejected; read-only permission and wrong owner kind are denied |
| `GH-WORK-05` | Attempt an accepted plan after isolated Project field/option recreation | O/P `live`; F `memory` | `not-configured` | Old field/option/configuration identity invalidates before effects and produces a new plan |
| `GH-WORK-06` | Inject response loss after a successful live issue create; reconcile by stable marker | O/P `live` through faulting client shim; F `memory` | `not-configured` | No blind create retry; zero or multiple lookup matches remain ambiguous |
| `GH-WORK-07` | Apply a multi-effect plan with one accepted effect followed by injected failure | F `memory`; O/P `live` postcondition reads | `not-configured` | Completed receipt/observation remains; next plan contains only remaining semantic delta |
| `GH-WORK-08` | Close a linked work item and reconcile item, parent, and direct dependents | O/P `live`; F `memory` | `not-configured` | Reproduces the #15 class and proves closed-to-Done plus parent/dependency rules |

### PR, checks, review, and rules

| ID | Claim and action | Required evidence | Current result | Negative or stopping condition |
|---|---|---|---|---|
| `GH-BRANCH-01` | Accept only an issue-named fixture branch whose PR carries the stable issue link and expected head identity | F `memory`; O/P `live` | `not-configured` | Unlinked, ambiguously linked, wrong-issue, wrong-base, or stale-head branch cannot satisfy implementation |
| `GH-PR-01` | Observe seeded PRs through draft, ready, checks pending/pass/fail, changes requested, and merged/closed states | O/P `live`; branch content is seeder authority | `not-configured` | Missing/stale check or review evidence blocks the gate; head SHA change invalidates review evidence |
| `GH-PR-02` | Record implementation, checks, distinct change review, product outcome approval, and stronger assurance as separate roles/results | F `memory`; O/P `live` observations | `not-configured` | A ruleset, check, implementer self-review, or broad approval is not relabeled as capable/distinct/qualified review |
| `GH-PR-03` | On a qualifying squash merge, record issue, source, PR, head, merge commit, checks/reviews, and completion receipt in durable memory | F `memory`; O/P `live` | `not-configured` | Draft, closed-unmerged, merge/rebase method, changed head, or absent governed source cannot complete work |
| `GH-RULE-01` | Read a disabled fixture ruleset, separately approve activation, observe enforcement on a fixture branch, then disable/delete | O/P public/free `live` | `not-configured` | Read permission cannot mutate; no operational branch target; bypass is not tested |
| `GH-RULE-02` | Compare effective layered rule observation with desired repository rule | O/P `live` | `not-configured` | Unavailable plan/visibility feature yields `unsupported` only after capability proof, otherwise `not-configured`; never assume a paid upgrade |

### Transport, rate, webhook, offline, and recovery

| ID | Claim and action | Required evidence | Current result | Negative or stopping condition |
|---|---|---|---|---|
| `GH-TRANS-01` | Follow REST `Link` and GraphQL cursors with `per_page`/`first` set to 1 across three marked objects; preserve immutable IDs | O/P `live`; F `memory` boundary cases | `not-configured` | Missing page or duplicate cursor fails; partial GraphQL `data` plus required-field `errors` is not success |
| `GH-TRANS-02` | Preserve 401, 403, 404-for-hidden-resource, validation, and GraphQL error distinctions | O/P/A one-less/revoked `live`; F `memory` other responses | `not-configured` | No credential broadening or guessing that a 404 proves absence |
| `GH-TRANS-03` | Record primary/GraphQL budgets and reset headers from ordinary requests | O/P/A `live` | `not-configured` | Do not exhaust live limits; a point-in-time budget is not a rate-recovery pass |
| `GH-TRANS-04` | Inject primary exhaustion, secondary limit with/without `retry-after`, reset, exponential delay, and maximum attempts | F `memory` | `not-configured` | Serialize mutations, normally pause at least one second, stop at bound; never intentionally provoke live limits or bans |
| `GH-WEB-01` | Validate official HMAC-SHA-256 vectors, invalid/missing signature, GUID dedupe, and replay hint | F `memory` | `not-configured` | Payload is a hint and always triggers authoritative read; no receiver is required |
| `GH-WEB-02` | After separate receiver approval, observe signed delivery, duplicate redelivery, failure, and recent-delivery lookup | O `live` | `not-configured` | GitHub does not auto-redeliver failures and exposes recent deliveries for only three days; polling/full reconciliation remains required |
| `GH-OFF-01` | Queue credential-free desired intent while offline, expire/stale it, reconnect, reacquire credential, repeat handshake/read/plan | F `memory`; O/P reconnect `live` | `not-configured` | No raw HTTP, token, inferred current readiness, or automatic identity switch in the queue |

## Phase 3 coverage ledger

| Phase 3 outcome | Contract cases |
|---|---|
| Label vocabulary bootstrap | `GH-BOOT-01` |
| Project fields and options; optional presentation-view inventory | `GH-BOOT-02` |
| Auto-add and closed-to-Done automation | `GH-BOOT-03`, `GH-WORK-08` |
| Two-layer issue rendering and schema validation | `GH-ISSUE-01` |
| Hierarchy, dependencies, and stale-reference readiness refresh | `GH-WORK-03`, `GH-READY-01` |
| Question/research completion and promotion | `GH-QRES-01` |
| Horizon intake/promotion and inherited Phase context | `GH-HORIZON-01` |
| Issue-linked branch and PR lifecycle | `GH-BRANCH-01`, `GH-PR-01` |
| Distinct review roles and stale-head invalidation | `GH-PR-02` |
| Squash completion memory | `GH-PR-03` |
| Ruleset inspection and separately approved application | `GH-RULE-01`, `GH-RULE-02` |
| Authentication, pagination, rate, replay, offline, and recovery | `GH-ID-01`–`GH-ID-05`, `GH-WORK-01`–`GH-WORK-07`, `GH-TRANS-01`–`GH-OFF-01` |

## Fixture and isolation contract

- Provision exactly one public O repository with one organization-owned Project and one
  public P repository with one user-owned Project. Names start with
  `codex-starter-kit-sandbox`; every run has an immutable run ID and every created object
  has a stable non-secret `starter-kit-contract:<run-id>:<case-id>` marker.
- Each route begins with these exact managed fixtures: labels `type:task`,
  `ready-for-agent`, and `contract-run`; single-select Project fields Status, Readiness,
  Horizon, and Phase; table view `Execution` and roadmap view `Horizon`; one auto-add
  workflow matching the run marker; and one closed-item workflow targeting Status Done.
  The run records immutable field, option, view, and workflow/configuration identities.
- The seeder creates exactly nine marked issues per route: parent, child, blocker,
  dependent, question, research, pagination-a, pagination-b, and pagination-c. It creates
  three fixture branches and two PRs: one draft/success-path PR and one failing/stale-head
  PR. One workflow supplies deterministic pending/pass/fail checks. One disabled ruleset
  targets only `contract/<run-id>/**`; its temporary activation and deletion are isolated
  `GH-RULE-01` effects. Create/replay and lost-response cases may create one additional
  issue each and must close them during case cleanup.
- The distinct test reviewer submits exactly one approval on the success-path PR and one
  changes-requested review on the failing PR. The seeder and Work Manager never submit a
  qualifying review. The run records reviewer account ID, repository role, credential
  permission facts, reviewed head SHA, and review ID without retaining the token.
- Project workflow configuration is asserted, not silently created by the adapter. If the
  named views or workflows cannot be provisioned exactly, their cases remain
  `not-configured`; another mechanism is not substituted.
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
  cleanup result; required evidence mode; result and any accepted-exception authority,
  reason, and expiry; limitations; and linked raw-evidence digest.

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
| Ownership | O requires an organization owner/App manager and Project administrator. P is outside the 1.0 claim. If required O roles are unavailable, affected live evidence is `not-configured`, not delegated to a broader token. |
| App trust | Register one private test App, selected-repository installation, no public listing, minimum permissions, and webhooks disabled unless GH-WEB-02 is separately selected. Record owner, managers, key rotation/revocation, and incident contact before use. |
| User token breadth | No classic PAT is approved or required. The reviewer uses a selected-repository fine-grained PAT with Contents read and Pull requests write only; it is not a Project or reconciliation fallback. |
| Data | Public synthetic fixture content only. No operational issue text, source, credentials, private metadata, or sensitive evidence enters the sandbox. |
| Cost | Public repositories with standard hosted runners have no Actions-minute charge under current docs. No paid upgrade, private target, larger runner, hosted receiver, or secret service is required for the base matrix. Usage and pricing remain invalidation triggers. |
| Compatibility | Base target is GitHub.com REST `2026-03-10` plus current GraphQL. GitHub Enterprise Server, private/paid behavior, App-based user-owned Project mutation, and webhook durability remain unsupported/not-configured until exact qualification. |
| Fallback | In-memory evidence remains available where live GitHub is unavailable. Missing live evidence narrows the support claim; it does not justify a personal route, broader token, or webhook authority. |

## Provisioning gate

This research does not authorize provisioning. Before any live setup, the owner must
approve the exact run manifest and confirm:

1. organization and personal sandbox owners, administrators, public visibility, and
   GitHub plan;
2. reconciler and seeder App owners/managers, selected repositories, exact permission
   manifests, key/secret stores, expiry, rotation, suspension/revocation, and incident
   responsibility;
3. whether the classic personal token fallback is accepted and, if so, its exact scopes,
   expiry, storage, and revocation path;
4. the distinct reviewer account owner, repository role, exact fine-grained permission,
   token lifecycle, independence from seeding/implementation, and availability;
5. whether optional live webhooks or any private/paid target are included, with hosting,
   data, retention, and cost implications;
6. fixture prefix, lease, 24-hour cleanup, 30-day raw evidence retention, and residual
   resource owner; and
7. spending/budget guardrails and the fallback/support claim if any target is declined.

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
- [Project view REST endpoints](https://docs.github.com/en/rest/projects/views)
- [Using the API to manage Projects](https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/using-the-api-to-manage-projects)
- [Automating Projects using Actions](https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/automating-projects-using-actions)
- [Adding items automatically](https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/adding-items-automatically)
- [Using built-in Project automations](https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/using-the-built-in-automations)
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
