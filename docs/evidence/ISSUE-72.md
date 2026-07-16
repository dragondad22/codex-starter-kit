# Issue #72 — production GitHub adapter evidence

**Date:** 2026-07-15

**Issue:** [#72](https://github.com/dragondad22/codex-starter-kit/issues/72)

**Parent:** [#4](https://github.com/dragondad22/codex-starter-kit/issues/4)

## Delivered outcome

- Added native Go `githubadapter` transport behind the issue #71 `WorkAdapter` seam.
- Added a credential-free allowlisted target manifest and injected ephemeral credential
  contract for App installation, user-token, and repository-local Actions modes.
- Added actor/repository/Project/API/granted-permission/configuration/rate handshake
  evidence before mutation.
- Added bounded REST/GraphQL observation, stable-marker create recovery, issue and
  Project reconciliation, immutable lifecycle-field updates, and verified replay.
- Extended capability and receipt schemas with identity, owner, API, rate, limitation,
  disposition, and evidence-mode facts while retaining issue #71 compatibility.

## Deterministic behavior evidence

Integration tests use native HTTP test servers and the public lifecycle-engine seam. They
cover full personal user-token and organization-App journeys; no-change replay; immutable
repository, Project, field, option, item, and issue IDs; REST and GraphQL pagination;
one-less permission; App-JWT observation of installation/account/slug and a mismatched
API installation negative; App/user-owner incompatibility;
Actions Project limitation; expiry/reconnect; marker ambiguity; lost-create-response
recovery without duplication; hidden 404 denial; validation failure; GraphQL partial
data; distinct authentication/authorization/not-found/validation results; durable
exponential rate attempts; human body/label preservation; partial recovery containing
only remaining semantic operations; mutation postcondition re-reads; simulated receipt
mode; and absence of the ephemeral credential from durable Work Manager state.

The in-memory and GitHub adapters use the same `WorkAdapter` interface and Work Manager
policy, plan, receipt, verification, and state implementation. Adapter results do not
choose Readiness, Status, Phase, relationships, completion, review policy, or a fallback
credential.

## Security and recovery evidence

The adapter pins REST version `2026-03-10`, uses native Go HTTP without shell strings,
limits response decoding, bounds pagination, prevents cross-host REST pagination,
serializes and paces live mutations, and redacts provider/transport detail at the receipt
seam. The user/installation token and handshake-only App JWT are omitted from JSON and
never persisted. Live mode rejects non-HTTPS or non-GitHub.com
endpoints and requires an approved-target assertion.

Create is not treated as generically idempotent. Every attempt first reads the stable
marker; exactly one match recovers, zero permits create, and multiple matches stop as
ambiguous. Partial issue/Project/field sequences return recoverable non-pass results and
are re-read before the next immutable plan.

## Verification

The exact reviewed candidate passed:

```text
python3 -m unittest discover -s tests -p "test_*.py"
  33 tests passed
python3 scripts/validate_docs.py
  Documentation validation passed.
go test -count=1 ./...
  all packages passed
go test -race -count=1 ./engine ./githubadapter
  both packages passed
go vet ./...
  passed
starter-kit changes validate --repository .
starter-kit changes check --repository .
  product 0.3.0; 12 Unreleased; 11 external; 1 internal
git diff --check
  passed
```

The completing PR must additionally pass the repository Linux/macOS/Windows GitHub
Actions matrix on this exact source before merge.

## Distinct review

Two independent reviewers compared `origin/main...6c978fb`. The standards review and the
issue #72 specification review both passed the exact candidate. Earlier review rounds
identified preservation of human-owned issue content, authoritative mutation re-reads,
redaction, rate evidence, no-op replay, distinct transport outcomes, durable exponential
attempts, partial recovery, API-observed App installation identity, explicit-state
preservation, and effect-authority rate provenance. Commits `8ca7ca2`, `29a51b4`, and
`6c978fb` resolved those findings, and the final reviewers reported no remaining blocker
or scope regression.

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
