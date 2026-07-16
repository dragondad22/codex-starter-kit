# Issue #72 — production GitHub adapter evidence

**Date:** 2026-07-15

**Issue:** [#72](https://github.com/dragondad22/codex-starter-kit/issues/72)

**Parent:** [#4](https://github.com/dragondad22/codex-starter-kit/issues/4)

## Delivered outcome

- Added native Go `githubadapter` transport behind the issue #71 `WorkAdapter` seam.
- Added a credential-free allowlisted target manifest and injected ephemeral credential
  contract for App installation, user-token, and repository-local Actions modes.
- Added actor/repository/Project/API/permission/rate handshake evidence.
- Added bounded REST/GraphQL observation, stable-marker create recovery, issue and
  Project reconciliation, immutable lifecycle-field updates, and verified replay.
- Extended capability and receipt schemas with identity, owner, API, rate, limitation,
  disposition, and evidence-mode facts while retaining issue #71 compatibility.

## Deterministic behavior evidence

Integration tests use native HTTP test servers and the public lifecycle-engine seam. They
cover full personal user-token and organization-App journeys; no-change replay; immutable
repository, Project, field, option, item, and issue IDs; REST and GraphQL pagination;
one-less permission; App installation/account binding; App/user-owner incompatibility;
Actions Project limitation; expiry/reconnect; marker ambiguity; lost-create-response
recovery without duplication; hidden 404 denial; validation failure; GraphQL partial
data; bounded rate evidence; simulated receipt mode; and absence of the ephemeral token
from durable Work Manager state.

The in-memory and GitHub adapters use the same `WorkAdapter` interface and Work Manager
policy, plan, receipt, verification, and state implementation. Adapter results do not
choose Readiness, Status, Phase, relationships, completion, review policy, or a fallback
credential.

## Security and recovery evidence

The adapter pins REST version `2026-03-10`, uses native Go HTTP without shell strings,
limits response decoding, bounds pagination, prevents cross-host REST pagination,
serializes Project mutations, and redacts transport detail at the receipt seam. Tokens
are omitted from JSON and never persisted. Live mode rejects non-HTTPS or non-GitHub.com
endpoints and requires an approved-target assertion.

Create is not treated as generically idempotent. Every attempt first reads the stable
marker; exactly one match recovers, zero permits create, and multiple matches stop as
ambiguous. Partial issue/Project/field sequences return recoverable non-pass results and
are re-read before the next immutable plan.

## Verification

Focused development checks passed after the red/green slices:

```text
go test ./githubadapter ./engine
go test -race ./githubadapter ./engine
go vet ./...
```

The completing revision must additionally pass the repository Python suite,
documentation validation, every Go package, release-change validation, diff hygiene, and
the Linux/macOS/Windows GitHub Actions matrix. Exact final results and distinct review are
recorded before merge.

## Explicit limitations

- All issue #72 adapter receipts are `simulated`; live cases are `not-configured`.
- No App registration/installation, token creation/storage/broadening, repository,
  Project, rule, workflow, paid/private feature, or external resource was changed.
- The adapter manages the narrow issue #72 one-task route. Native hierarchy,
  dependencies, issue templates, Horizon/Phase, question/research promotion, PR delivery,
  webhooks, and aggregate qualification remain #74–#76.
- #73 owns provisioning, bootstrap, target leases, cleanup, and live authority.

## GitHub reconciliation

Issue #71 merged and is closed / Project `Done`. Issue #72 was promoted from `Blocked` to
`Ready`, selected, and moved to `In progress` on its issue-named branch. Parent #4 remains
`In progress`. Direct dependent #73 remains `Blocked` until the completing #72 PR merges.
