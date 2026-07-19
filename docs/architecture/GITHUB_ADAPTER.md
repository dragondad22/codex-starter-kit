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
| `user-token` | Expected API user, accepted owner route, selected repository/Project, permissions, expiry, and API actor are bound before effects | #46's operational Phase catalog was applied through a separately approved owner CLI route; that result is `needs-review`, not a live Work Adapter user-token qualification |
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
4. Project node ID, owner login, and owner kind;
5. pinned REST version `2026-03-10` and a successful GraphQL compatibility query;
6. required versus granted permissions, using observed classic-user scopes or the
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

## Evidence boundary

Deterministic tests use native `httptest` REST/GraphQL fixtures through the real adapter
and public lifecycle seam. They cover both supported owner routes, complete
create/project/update/verify/no-change replay, secret-free state, REST/GraphQL pagination,
one-less permission, identity/owner mismatch, expiry/reconnect, ambiguous markers,
lost-response recovery, hidden resources, validation, partial GraphQL data, rate
scheduling, native hierarchy/dependency observation, unavailable relationship endpoints,
Actions limitations, and unsupported combinations.

Those receipts are labeled `simulated`. They prove implementation semantics and native
HTTP portability, not a GitHub permission or service claim. No live target, token, App,
Project, issue, or paid feature was created or mutated for #72. #73 subsequently
qualified the isolated sandbox baseline, and #46 applied and re-read the approved
operational Phase catalog through an owner CLI route. Neither result is a live pass for
the routine Work Adapter Phase effect. #46's saved `Phases` view remains `not-configured`;
#76 owns aggregate qualification and support claims.
