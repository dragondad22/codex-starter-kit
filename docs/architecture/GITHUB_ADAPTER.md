# GitHub Adapter — Identity and Transport Contract

**Status:** Implemented deterministic production transport; isolated sandbox baseline qualified

**Issue:** [#72](https://github.com/dragondad22/codex-starter-kit/issues/72)

**Sandbox and live authority:** [#73](https://github.com/dragondad22/codex-starter-kit/issues/73)

## Interface and ownership

The `githubadapter` Go module implements the lifecycle engine's existing `WorkAdapter`
interface. Work Manager continues to own desired lifecycle policy, immutable plans,
preconditions, receipts, verification, and recovery. The adapter owns only ephemeral
credential acquisition, a fixed target manifest, normalized GitHub observation, and
allowlisted REST/GraphQL effects.

Issue #73 adds a separate `SandboxAdapter` seam for baseline and fixture resources whose
authority differs from routine one-task reconciliation. `NewSandbox` accepts one immutable
organization repository/Project allowlist plus three role expectations and injected
credential providers. It aggregates reconciler, seeder, and rules installation authority
without making any one credential a fallback for another. See
[GitHub Sandbox Bootstrap](SANDBOX_BOOTSTRAP.md).

#46 reuses that content-addressed external-resource lifecycle for a separately authorized
operational Project configuration; it does not reuse sandbox authority. The retained v1
type names are wire-compatibility labels. For a user-owned Project, the adapter binds the
numeric owner and Project identities to one explicitly selected user-token route, verifies
the API actor and complete observed classic OAuth scope set (including `project`), and
resolves the owner/number REST route to the same immutable Project node used by GraphQL.
It inventories fields through REST and views/items through GraphQL, and permits only the
reviewed Project resources. GitHub App and
fine-grained-token routes are rejected for user-owned saved-view creation.

`githubadapter.New` accepts one credential-free configuration, an injected credential
provider, and a native Go HTTP client. The configuration names the host, pinned REST
version, expected mode and actor, repository and Project immutable IDs/owners, lifecycle
field/option IDs, permission manifest, pagination bound, and evidence mode. The provider
returns an ephemeral user or installation token plus expected actor, permission, expiry,
account, and installation facts. App mode also supplies an ephemeral App JWT solely for
the non-mutating installation-identity query.

Both secret values have `json:"-"`, are added only to request authorization headers, and never enter
desired intent, capability output, observations, plans, receipts, durable state, or
diagnostics. GitHub CLI, a shell, keychain discovery, and automatic account selection are
not runtime dependencies.

## Supported identity roles

| Mode | Implemented deterministic contract | Current live result |
|---|---|---|
| `app-installation` | Expected App slug and numeric installation/account are API-observed, then bound to the organization-owned Project, selected repository, mint-response permissions, and expiry before effects | #73 qualified the three named App roles against the approved organization sandbox; the Work Adapter's multi-item reconciliation route remains unqualified live |
| `user-token` | Expected API user, accepted owner route, selected repository/Project, permissions, expiry, and API actor are bound before effects | #46's separately approved external-resource lifecycle passed exact-head zero-effect observation, verification, and replay for the operational Phase catalog; the routine Work Adapter user-token route remains unqualified live |
| `actions-job` | Repository actor and target can be inspected | `unsupported` for the Project route; repository-local authority is never promoted to Project or cross-repository authority |

App installation mode rejects a user-owned Project rather than selecting a user token.
Live mode additionally requires an explicit approved-target assertion and the pinned
GitHub.com HTTPS API host. GHES and `app-user` remain unsupported.

## Handshake

`Capability` performs non-mutating native HTTP requests and returns schema-versioned,
credential-free facts:

1. expected mode and ephemeral credential identity, expiry, and permissions;
2. API actor kind and login, or App slug plus installation ID/account from the authenticated App installation response;
3. repository node ID and owner;
4. Project number and node ID plus owner login, immutable ID, and kind, proving that the
   REST owner route and GraphQL target identify the same Project;
5. pinned REST version `2026-03-10` and a successful GraphQL compatibility query;
6. required versus granted permissions, retaining and exactly binding every observed
   classic-user scope or the
   App-installation mint response bound to its permission revision;
7. current Project lifecycle field and option identities before any effect; and
8. REST and GraphQL limit, used, remaining, and reset budgets, limitations,
   configuration digest, evidence mode, and
   freshness.

Wrong actor, account, installation, owner, immutable ID, permission, API version, expired
credential, unsupported owner/mode combination, or unapproved live target stops before an
effect. Reconnect reacquires the explicitly selected mode and repeats the handshake; it
does not broaden authority or switch credentials.

## Observation and effects

Observation follows bounded REST `Link` pages and GraphQL Project-item cursors, matches
the exact non-secret `starter-kit-managed:<managed-id>` marker, and normalizes issue,
Project item, lifecycle option, native parent, parent Phase option, and managed metadata
identities. The capability handshake requires exactly one single-select `Phase` field and
the complete named Phase 0–8 option catalog whenever Phase is configured. Renamed,
duplicate, wrong-type, missing, extra, or stale catalog state is `needs-review`. For an existing selected
issue it reads the version-pinned native parent, sub-issue, `blocked_by`, and `blocking`
endpoints; reads each dependent's complete blocker slice; and resolves the parent,
siblings, and dependents through immutable issue and Project-item identities. The engine
compares that graph with governed intent before planning. Zero selected matches means the
task is absent. Multiple markers are `ambiguous`; unavailable relationship endpoints,
missing stable identities or Project items, incomplete parent membership, pagination
exhaustion, and GraphQL partial data are explicit non-pass results, never partial success
or a fallback to issue prose.

The adapter accepts only the two semantic effects produced by Work Manager:

- `create-task` re-reads the marker before POST. One existing match recovers a lost
  response; multiple matches remain ambiguous.
- `reconcile-task` carries an ordered list containing only the remaining semantic
  operations: issue metadata, issue closure/reopening, Project membership, Readiness, and
  Status, plus direct Phase where configured. A related parent closure patches only issue state and therefore does not rewrite
  human-owned title, body, or labels. The adapter skips already-converged operations and
  re-reads every mutation before reporting it applied. Phase is set by immutable option ID
  for directly assigned work and cleared from ordinary children that derive it from a parent.

Expired/invalid authentication, insufficient authorization, hidden-resource 404,
validation failure, offline transport, GraphQL partial errors, bounded pagination
exhaustion, and rate delay remain distinct outcomes. Rate receipts retain durable attempt,
maximum attempts, exponential retry time, and reset time without response bodies or
credentials. Mutation calls are adapter-serialized; live mode enforces at least one
second between them.

The external-resource adapter retries only idempotent REST `GET` reads. A read receives at
most three attempts: `502`, `503`, and `504` use short bounded backoff, while `429` is
eligible only with a valid `Retry-After` value that fits the two-second aggregate wait
budget. Cancellation or deadline expiry interrupts the wait and returns through the
lifecycle seam. REST effects and every non-`GET` request remain single-attempt; semantic
identity mismatch, authentication, permission, absence, other status responses, and
successful-but-malformed payloads are not retried. Exhausted eligible reads report the
provider as transiently unavailable rather than converting transport availability into
Project identity or inventory drift. Diagnostics retain neither response bodies nor
credentials.

## Evidence boundary

Deterministic tests use native `httptest` REST/GraphQL fixtures through the real adapter
and public lifecycle seam. They cover both supported owner routes, complete
create/project/update/verify/no-change replay, secret-free state, REST/GraphQL pagination,
one-less permission, identity/owner mismatch, expiry/reconnect, ambiguous markers,
lost-response recovery, hidden resources, validation, partial GraphQL data, rate
scheduling, native hierarchy/dependency observation, unavailable relationship endpoints,
Actions limitations, bounded transient user/Project identity and field-inventory
recovery, single-attempt semantic and effect failures, cancellation, and unsupported
combinations.

Those receipts are labeled `simulated`. They prove implementation semantics and native
HTTP portability, not a GitHub permission or service claim. No live target, token, App,
Project, issue, or paid feature was created or mutated for #72. #73 subsequently
qualified the isolated sandbox baseline. #46 later observed, verified, and replayed the
approved operational Phase catalog through a separately approved owner route with zero
effects. That result qualifies the exact external-resource lifecycle candidate, not the
routine Work Adapter Phase effect. #46 observed the matching human-created `Phases` view;
automated creation remains `not-configured` when the required grouping and sorting cannot
be expressed. #76 owns aggregate qualification and support claims.

The Project-configuration route plans the complete Phase field/option catalog, the saved
`Phases` view, and feature #1–#9 assignments as immutable resources. Existing name/type,
option, node, item, layout, visible-field, grouping, sorting, or assignment conflicts stop
instead of creating duplicates or overwriting human state. A missing view may be created
through the version-pinned REST route only when that route can express its complete desired
configuration, and must be re-observed through GraphQL. GitHub's current create-view request
accepts name, layout, filter, and visible fields but not grouping or sorting, so a missing
grouped/sorted `Phases` view is `not-configured` without creating a partial view; an existing
matching human-created view remains observable and replayable. `404`, denial, partial data,
and missing postconditions retain explicit non-pass results. Feature assignments update only
the approved item, field, and option IDs and are re-read before pass.

Clean creation may leave field, option, and view identities unbound in desired input. The
adapter retains GitHub's returned immutable IDs in receipts and normalized observation, and
subsequent effects in that lifecycle adopt the observed identities. A manifest that already
pins an identity still fails closed when the provider identity changes. Option inventory and
mutation use the configured user- or organization-owner route rather than assuming one owner
kind.
