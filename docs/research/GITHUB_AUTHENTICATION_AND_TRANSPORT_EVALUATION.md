# GitHub authentication and transport evaluation

**Status:** Completed research; recommendation awaits authoritative promotion
**Issue:** [#63](https://github.com/dragondad22/codex-starter-kit/issues/63)
**Planning source:**
[GitHub executable-work decision map](../roadmap/GITHUB_EXECUTABLE_WORK_DECISION_MAP.md#2-which-github-authentication-and-transport-contract-should-the-kit-support)
**Freshness:** 2026-07-14
**Promotion:** Requires an authoritative Phase 3 contract and executable decomposition;
this research record does not establish product or architecture authority

## Objective and stopping conditions

Select the narrowest supportable GitHub authentication and transport contract for the
1.0 Work Manager and production GitHub adapter. The result must compare GitHub App
installation identity, user-token and GitHub CLI identity, and Actions `GITHUB_TOKEN`;
cover permissions, attribution, expiry, revocation, rate limits, retries, idempotency,
offline behavior, cost, native compatibility, and fallback; and identify what needs live
qualification before a support claim.

The evaluation stops when official current GitHub documentation and read-only probes of
this repository are sufficient to assign every identity a supported role, recommend a
transport boundary, define safe failure/recovery behavior, and preserve material
uncertainty. It does not register or install a GitHub App, create or rotate credentials,
change repository or Project state, exercise a paid feature, or qualify a production
adapter.

## Method and provenance

Official GitHub documentation was retrieved on 2026-07-14 and is the primary source for
documented behavior. Read-only REST and GraphQL probes used an explicitly selected
`dragondad22` credential against `dragondad22/codex-starter-kit` and user-owned Project
#8. Local observations establish only this environment and cannot prove general support.

No probe printed a credential, created content, changed permissions, updated fields, or
invoked a mutating endpoint. The GitHub CLI is an observation and operator tool in this
research; it is not assumed to be the product transport.

## Governing constraints

- The lifecycle engine and Work Manager own desired state, policy, transitions, plans,
  evidence, and reconciliation. The GitHub adapter owns authentication and transport.
- External effects use explicit authority, precondition checks, idempotent desired-state
  reconciliation, and compensation; they do not claim distributed atomicity.
- Secrets never enter plans, issue bodies, logs, ordinary evidence, or offline queues.
- GitHub is required in 1.0, but network availability cannot become the sole authority for
  local status, policy, evidence, or reproducibility.
- Universal behavior runs natively on Linux, macOS, and Windows without requiring GitHub
  CLI, a shell string, WSL, or a platform-specific credential-store implementation.
- Missing identity, permission, freshness, plan capability, or verified postcondition is
  an explicit non-pass, never inferred success.

## Identity findings

### GitHub App installation identity

An installation access token represents the installed app, so activity is attributed to
the app rather than a human. It is the correct identity for unattended reconciliation
that does not need user intent on every request. The installation limits both repositories
and permissions, and token creation may narrow them further. Installation tokens expire
after one hour and can be minted again from the app identity; they should never be stored
as durable configuration.

The App registration and installation are separately privileged effects. An organization
owner may be required, and an administrator can restrict installation. The private key
and webhook secret introduce rotation, storage, and incident-response obligations. App
suspension, uninstallation, changed repository selection, or changed permissions must
invalidate affected capability rather than silently select a broader identity.

For GraphQL, a normal installation receives 5,000 points per hour and may scale with
repository and organization-user count to 12,500; Enterprise Cloud installations receive
10,000. REST has a separate budget, and both APIs also impose secondary limits.

**Material limitation:** GitHub documents installation tokens as valid for Projects API
automation generally, and demonstrates organization-owned Project automation with an
App. However, current REST endpoints for adding items to a **user-owned** Project
explicitly reject installation tokens, GitHub App user tokens, and fine-grained personal
tokens. GraphQL documentation does not resolve the exact user-owned mutation permission
matrix. Therefore 1.0 must not claim App-based user-owned Project reconciliation until a
live sandbox proves every required mutation. The preferred App contract is initially for
organization-owned Projects.

### User-token and GitHub CLI identity

A user token performs work with the user's authority and lifecycle. It is appropriate for
interactive setup, owner-approved local operations, personal-account compatibility, and a
recovery path when an App is unavailable. It is not the preferred unattended machine
identity: access changes when the person loses access, tokens expire or are revoked, and
automation is attributed to a human.

Fine-grained personal access tokens improve repository and permission scoping, but GitHub
currently lists user-owned Projects as an unsupported gap. A classic personal access token
with `project` and repository scope remains the documented user-owned Project route, but
classic `repo` authority spans all repositories the user can access rather than one
managed repository. That breadth must be disclosed and explicitly accepted; it cannot be
silently requested as the default.

GitHub CLI can securely broker a browser login and is useful for human setup and
diagnostics. It is not a stable engine dependency or an identity oracle. Environment
`GH_TOKEN` takes precedence over stored credentials, and GitHub CLI currently has an open
multi-account keychain defect in which the displayed account can differ from the token
actually retrieved. The adapter must validate the API actor and resource owner before any
effect rather than trusting a config label or active-account display.

GitHub App user access tokens are not required for the initial contract. They add an
OAuth/device flow, user-plus-app permission intersection, refresh-token handling, and
human attribution while retaining the user-owned Project uncertainty. They remain Later
unless a distinct user-delegation need cannot be met by the explicit user-token fallback.

### Actions `GITHUB_TOKEN`

At the beginning of each Actions job, GitHub creates a repository-scoped installation
token for that job. It expires when the job completes or at its effective maximum
lifetime. Workflow or job `permissions` should reduce it to the minimum required access.
An action can obtain the token from the `github.token` context even when it is not passed
explicitly, so third-party action review remains part of the trust boundary.

`GITHUB_TOKEN` is supported for repository-local CI verification and bounded event-driven
updates that its declared job permissions can perform. It is not the general Work Manager
identity: it is limited to the workflow repository, has a GraphQL allowance of 1,000
points per hour per repository, and cannot be assumed to own a cross-repository or
user-owned Project. When CI needs broader approved reconciliation, it should mint the
same scoped GitHub App installation token used by the production adapter rather than
store a broad user token.

## Recommended 1.0 identity contract

| Mode | 1.0 role | Support boundary |
|---|---|---|
| `app-installation` | Preferred unattended and CI reconciliation identity | Full support begins with organization-owned Projects after the exact permission and mutation matrix passes sandbox qualification |
| `user-token` | Interactive/local setup, personal-account route, and explicit recovery fallback | Fine-grained token where every capability supports it; classic PAT only for a documented gap such as user-owned Projects, with breadth, expiry, storage, and revocation implications surfaced |
| `actions-job` | Repository-local CI verification and bounded updates | Only the current workflow repository and explicitly declared job permissions; never presumed to be the global reconciler |
| `app-user` | Deferred | No distinct 1.0 need justifies its additional OAuth, refresh, attribution, and permission-intersection surface |
| unauthenticated | Read-only public discovery only | Never sufficient for readiness, private state, Project synchronization, or mutation |

Each adapter request receives an ephemeral credential from an injected credential
provider. The credential provider is outside the domain model; plans name the credential
mode and expected authority, never secret material or a secret-store path. At minimum the
non-mutating capability handshake records:

1. GitHub host and API family/version;
2. credential mode, actor kind, expected login or App identity, and installation/account;
3. repository immutable ID, owner, visibility, and observed role or App repository grant;
4. Project immutable ID, owner kind, and required field identities;
5. required versus observed permissions for the planned operations;
6. current REST and GraphQL budgets and reset times; and
7. every plan, account, feature, or endpoint limitation that narrows capability.

A user-mode handshake verifies the authenticated API login. An App-mode handshake
verifies App and installation identity, repository selection, and granted permissions.
Any mismatch is `unsupported` or `needs-review` according to the known facts; the adapter
does not switch accounts or credentials automatically.

## Permission segmentation

The adapter declares capability-specific permission sets rather than one permanent
all-powerful token:

| Capability | Expected minimum permission family | Notes |
|---|---|---|
| Inspect work | Metadata plus read Issues and Pull requests | Include only fields required for desired-state comparison |
| Reconcile issues, labels, hierarchy, and dependencies | Write Issues | REST now exposes sub-issue and dependency endpoints; sandbox must verify exact token support |
| Reconcile Project items and fields | Read/write Projects at the Project owner's scope | Organization Project App permission is distinct; user-owned Project token support is the documented gap |
| Inspect rules and repository settings | Read Administration | Treat unavailable plan features distinctly from missing permission |
| Apply repository rulesets/settings | Write Administration | High-authority, separately approved plan effect; never bundled into routine issue reconciliation |
| Observe PR gates/reviews | Read Pull requests, Checks, statuses, and Actions only as needed | Checks support differs by token type and requires endpoint qualification |
| Mutate PR metadata | Write Pull requests | Does not imply contents, workflow, merge, or bypass authority |
| Change repository files or workflows | No Work Manager permission by default | Normal issue branch and PR flow owns content; `Contents` or `Workflows` is a separate delivery effect |

The final permission manifest must be generated from the exact endpoint inventory and
verified in the sandbox. GraphQL permission errors do not reliably substitute for a
declared manifest; GitHub specifically recommends testing App GraphQL operations.

## Transport contract

### Native HTTP, not GitHub CLI

The production adapter should use Go's native HTTP/TLS and JSON support through the
lifecycle-engine seam. This keeps Linux, macOS, and Windows behavior under one testable
contract and avoids a runtime installation, shell, GitHub CLI config, credential-store,
or active-account dependency. GitHub CLI remains an optional setup, debugging, and
break-glass tool.

Use the REST API for repositories, issues, labels, milestones, comments, sub-issues,
dependencies, pull requests, rulesets, checks, and other resources with stable REST
coverage. Use GraphQL where Project v2 or a relationship is not equivalently supported by
REST. Stable node IDs connect the two surfaces; URLs are returned presentation data and
must not be parsed as identifiers.

The GitHub.com adapter pins the documented REST API version in
`X-GitHub-Api-Version`—currently `2026-03-10`—and declares its supported version range.
GitHub promises at least 24 months of support for the preceding REST version after a new
version is released, but additive schema changes and GraphQL changes still require
contract qualification.

GitHub Enterprise Server is an explicit 1.0 limitation unless separately qualified. Each
instance requires its own App registration and credentials, different base URLs, awareness
of delayed or absent API/webhook features, and administrator-configured rate limits. The
transport must not hard-code GitHub.com in domain data so a future adapter can supply
those capabilities honestly.

### Pagination, errors, and rate limits

- Follow REST `Link` headers and GraphQL cursors; every GraphQL connection uses `first` or
  `last` from 1–100.
- Preserve GraphQL `errors` even when partial `data` is present. Required partial data is
  not a successful observation.
- Serialize mutating requests and normally pause at least one second between them. Avoid
  concurrency that can trigger secondary limits.
- Honor `retry-after`; when the primary budget is exhausted, wait until
  `x-ratelimit-reset`; otherwise use bounded exponential backoff after at least one
  minute for repeated secondary-limit failures.
- Record rate-limit resource, limit, used/remaining, reset, attempt count, and terminal
  disposition without recording credentials or sensitive response content.
- Stop after a bounded retry count. Continuing while rate limited can cause an integration
  ban and is never a recovery strategy.

### Idempotency and reconciliation

GitHub does not document a generic idempotency-key guarantee for the required issue and
Project mutations. The adapter therefore cannot equate HTTP retry with safe replay.

Every effect uses a versioned desired object and stable local operation ID. Before a
mutation, the adapter reads current state and compares normalized semantic fields. After
the mutation, it re-reads the authoritative object and records the observed postcondition.
Updates and deletes are retried only when the operation is inherently idempotent or a
fresh read proves the intended state is still absent. Create operations use a stable,
non-secret managed marker and lookup so a lost response can be reconciled without creating
a duplicate. Ambiguous completion remains `needs-review` until observation resolves it.

Project field names and option names are presentation/configuration, not durable identity.
Plans bind Project, item, field, and option IDs plus the observed configuration version;
renamed, deleted, or recreated fields make the plan stale and trigger re-planning.

## Webhooks, offline work, and recovery

Webhooks are change hints, not authoritative state. Validate the SHA-256 signature,
deduplicate by delivery GUID, enqueue the affected resource, then reconcile from the API.
GitHub does not automatically retry failed webhook deliveries; recent deliveries can be
redelivered for only three days. A periodic full or bounded reconciliation scan is
therefore required even when webhooks are enabled.

Offline operation may inspect cached observations, prepare a proposed plan, and queue a
credential-free desired-state intent with its source hashes and expiry. It may not claim
current GitHub readiness, permission, rule, PR, or Project state. On reconnect, the engine
reacquires a credential, repeats the capability handshake and precondition reads, rejects
stale plans, then applies serially. The queue is never an append-only list of raw HTTP
requests because those requests may no longer be authorized, safe, or idempotent.

Revoked, expired, suspended, uninstalled, or narrowed credentials produce an explicit
authentication or authorization result. The engine retains safe local evidence and the
pending desired state, but does not loop, broaden scopes, switch to a human token, or ask
the AI to repair authority. Recovery reports the exact human or administrator action and
re-plans only after a fresh handshake succeeds.

## Cost, plan, and compatibility implications

- GitHub documents plan-dependent features and usage, not one universal feature set.
  Repository rulesets and protected branches are available for public repositories on
  GitHub Free, but private repositories require GitHub Pro, Team, or Enterprise Cloud;
  push rulesets have narrower paid-plan availability.
- GitHub-hosted Actions minutes and storage may incur cost for private repositories or
  usage beyond plan allowances. An App webhook receiver, secret manager, and monitoring
  service may also create non-GitHub operating cost.
- The 1.0 adapter must report `not-configured`, `unsupported`, or `needs-review` when a
  plan or account lacks a required feature. It must not recommend a paid upgrade without
  explaining cost, authority, compatibility, and fallback.
- A local, on-demand App adapter can avoid requiring an always-on hosted service in the
  first slice. Webhooks improve latency and polling cost but do not replace reconciliation.
- Native support belongs to the Go engine and HTTP contract. Availability of `gh`, a
  browser, a keychain backend, or a particular shell is not part of universal support.

## Local read-only observation snapshot

Observed on 2026-07-14:

| Fact | Observation | Claim boundary |
|---|---|---|
| Explicit API actor | `dragondad22` | Selected by account-specific keyring lookup and verified through the API; no default CLI account was trusted |
| Repository | Public `dragondad22/codex-starter-kit`; authenticated user reports admin/maintain/push/triage/pull | Current user observation only, not an App permission proof |
| Project | User-owned Project #8, public, open, 15 fields including Status, Horizon, Readiness, Parent issue, and Sub-issues progress | Read-only GraphQL access; no mutation support was tested |
| REST and GraphQL budgets | 5,000 each for the selected user; sampled GraphQL query cost 1 | One point in time; secondary limits and other identities differ |
| Native hierarchy | REST listed #4's three sub-issues; dependency read returned an empty set for #15 | Confirms current endpoint readability, not write authority or full dependency modeling |
| GitHub CLI | `2.45.0` on Linux x86-64 | Optional diagnostic tool only; older than current upstream and not a support baseline |
| Multi-account behavior | Default CLI account display and API identity previously diverged on this host | Matches open upstream cli/cli #12885; explicit token plus API identity avoided the defect |

## Recommendation and downstream obligations

Adopt the three supported modes in the identity table, with `app-installation` as the
preferred production identity, `user-token` as the explicit personal-account and recovery
route, and `actions-job` as a repository-local CI identity. Implement transport as native
REST plus GraphQL behind the GitHub adapter; do not make GitHub CLI a runtime dependency.

Ticket #3 must model credential mode, actor, host, immutable repository/Project IDs,
permission requirements, API/schema identities, desired-state markers, observations,
rate state, attempts, and explicit ambiguous/recovery results without putting a token in
the contract. Ticket #4 must qualify at least:

1. App installation issue/PR/rules behavior in an organization-owned sandbox Project;
2. user-token behavior for a personal-account repository and user-owned Project;
3. `GITHUB_TOKEN` repository-only behavior and denied cross-scope behavior;
4. every minimum permission set and one-less-permission negative case;
5. lost responses, duplicate delivery GUIDs, stale Project option IDs, partial GraphQL
   data, 401/403/404 distinctions, primary and secondary rate-limit recovery, and revoked
   or suspended identity; and
6. plan/feature differences for public Free versus any separately authorized private or
   paid qualification target.

Do not promote App-based user-owned Project support, GitHub Enterprise Server support,
private-repository ruleset support, webhook durability, or a minimum GitHub CLI version
without that exact live evidence.

## Primary sources

- [About authentication with a GitHub App](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/about-authentication-with-a-github-app)
- [Generating an installation access token](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-an-installation-access-token-for-a-github-app)
- [Choosing GitHub App permissions](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/choosing-permissions-for-a-github-app)
- [Managing personal access tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
- [Token expiration and revocation](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/token-expiration-and-revocation)
- [Actions `GITHUB_TOKEN`](https://docs.github.com/en/actions/concepts/security/github_token)
- [Using `GITHUB_TOKEN` in workflows](https://docs.github.com/en/actions/security-for-github-actions/security-guides/automatic-token-authentication)
- [Using the API to manage Projects](https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/using-the-api-to-manage-projects)
- [Automating Projects using Actions](https://docs.github.com/en/issues/planning-and-tracking-with-projects/automating-your-project/automating-projects-using-actions)
- [User-owned Project item REST endpoint](https://docs.github.com/en/rest/projects/items)
- [GraphQL rate and query limits](https://docs.github.com/en/graphql/overview/rate-limits-and-query-limits-for-the-graphql-api)
- [REST API best practices](https://docs.github.com/en/rest/using-the-rest-api/best-practices-for-using-the-rest-api)
- [REST API versions](https://docs.github.com/en/rest/about-the-rest-api/api-versions)
- [Handling failed webhook deliveries](https://docs.github.com/en/webhooks/using-webhooks/handling-failed-webhook-deliveries)
- [Validating webhook deliveries](https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries)
- [Repository ruleset availability](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets)
- [GitHub plans](https://docs.github.com/en/get-started/learning-about-github/githubs-plans)
- [GitHub Enterprise Server App differences](https://docs.github.com/en/enterprise-cloud@latest/apps/sharing-github-apps/making-your-github-app-available-for-github-enterprise-server)
- [GitHub CLI environment precedence](https://cli.github.com/manual/gh_help_environment)
- [GitHub CLI multi-account keychain defect](https://github.com/cli/cli/issues/12885)
