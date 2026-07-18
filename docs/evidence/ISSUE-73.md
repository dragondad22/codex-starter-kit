# Issue #73 — isolated GitHub sandbox bootstrap evidence

**Date:** 2026-07-16

**Issue:** [#73](https://github.com/dragondad22/codex-starter-kit/issues/73)

**Parent:** [#4](https://github.com/dragondad22/codex-starter-kit/issues/4)

## Approved authority and target

The product owner approved external-effect plan `issue-73-bootstrap-v1` for GitHub Free
organization `codex-starter-kit-labs` (account ID `305967668`) only. The public sandbox
repository is `codex-starter-kit-labs/codex-starter-kit-sandbox` (REST ID `1303189066`,
node `R_kgDOTa0WSg`) and the public organization Project is #1 (node
`PVT_kwDOEjyyNM4Bdm9F`). The personal Project/classic-PAT route was explicitly removed
from required 1.0 qualification.

Public synthetic data, standard hosted runners, zero-dollar operation, one mutating
lease, fixture cleanup within 24 hours, 30-day raw-evidence retention, retained baseline
resources, and separately approved repository deletion are fixed guardrails.

On 2026-07-17 the product owner approved DEC-0022 and execution mandate
`issue-comment-5009113729`. The mandate replaces repeated exact-plan approval for this
bounded target. It authorizes the named reconciler, seeder, rules, and reviewer actors;
governed baseline resources; marker-scoped fixtures and rules; proof transitions;
credential-revocation evidence; recovery replanning; and cleanup. It expires after the
bounded #73 run and does not authorize operational targets, broader credentials,
private/paid data, webhooks, bypass, repository deletion, or unrecognized-resource edits.
Every effect still records its exact plan and mandate identities.

## Applied baseline

- Project Status, Readiness, Horizon, and Phase fields/options were created and re-read.
- Execution table and Horizon roadmap views were created and re-read.
- Reviewer machine user `american-dragon-designs` (account ID `305973890`) has repository
  Write, with no organization-owner or repository-admin authority.
- Reconciler, seeder, and rules Apps are installed only on the sandbox repository. Four
  Actions environments isolate reconciler, seeder, rules, and reviewer secrets.
- Managed labels `type:task`, `ready-for-agent`, and `contract-run` were created and
  re-read by immutable node ID.
- The built-in Item closed workflow and the exact repository-scoped
  `is:issue label:contract-run` auto-add workflow are enabled and were proved with
  positive, negative, and close-to-Done cases.

Secret values were never read. Environment-secret metadata established that the three
App private keys and reviewer token exist. The seeder installation reported Workflows,
Contents, Issues, and Pull requests write before fixture qualification began.

## Implemented behavior

The engine supports strict sandbox inspect, immutable plan, apply, verify, status, and
composed bootstrap operations. Tests cover converging and replaying a missing resource,
name collision, stale observation, partial application and restart recovery, exact marked
cleanup that preserves human resources, active-lease refusal, unsupported kinds, and
sensitive-looking manifest rejection. State is integrity protected and receipts contain
no credentials.

Schema-v2 mandate containment accepts regenerated recovery plans only when target, actor,
resource/effect kind, marker or baseline key, effect count, recovery owner, and expiry
remain bounded. Out-of-mandate plans stop before effects. Historical schema-v1 exact-plan
approval remains readable and executable for replay without rewriting prior evidence.

The production GitHub sandbox adapter validates three distinct expiring App installation
roles against actor, account, installation, permissions, and the immutable target. It
uses native versioned REST and GraphQL transport to observe managed labels and Project
fields/options/views/workflows. Human-owned workflow configuration reports
`needs-review` when absent rather than selecting a substitute mechanism.

## Live qualification result

The isolated bootstrap boundary is `pass`:

- Qualification apply run
  [`29610278071`](https://github.com/codex-starter-kit-labs/codex-starter-kit-sandbox/actions/runs/29610278071)
  converged the seeder, distinct reviewer, and rules roles. Its redacted receipts retain
  fixture PR/review identities, active ruleset identity `19106535`, and `http-401`
  revocation proofs for the seeder and rules App credentials.
- Proof planning run
  [`29625594695`](https://github.com/codex-starter-kit-labs/codex-starter-kit-sandbox/actions/runs/29625594695)
  emitted immutable plans from candidate `4163e8ad03d69253f064e964d422ea69bedf35b2`.
  Proof apply run
  [`29625632625`](https://github.com/codex-starter-kit-labs/codex-starter-kit-sandbox/actions/runs/29625632625)
  retained the active-rules `http-422` denial and post-revocation `http-401` evidence.
  Both receipts bind plan `sha256:8f35e8235e674c4e55cb8886a212ecd842740237c55660097895da2949e129fa`
  to mandate `sha256:e1e9a1dfd8905786bca27e91db6be3480cb7f5f7fad5f42f85f510bb9288170b`.
- Cleanup apply run
  [`29625729444`](https://github.com/codex-starter-kit-labs/codex-starter-kit-sandbox/actions/runs/29625729444)
  removed the marked workflow, ruleset, and fixture branches and closed retained fixture
  records. It preserved a non-pass receipt when GitHub auto-closed a PR after its branch
  was deleted, even though verification found the target state converged. That result
  drove the dependency-ordering correction.
- Recovery planning run
  [`29625829477`](https://github.com/codex-starter-kit-labs/codex-starter-kit-sandbox/actions/runs/29625829477)
  produced zero-effect plans for both cleanup roles from candidate
  `eadd27373b924738da4151c9ccdf5198a38fe80d`. Recovery apply run
  [`29625918740`](https://github.com/codex-starter-kit-labs/codex-starter-kit-sandbox/actions/runs/29625918740)
  accepted the engine's explicit `no_change` result and re-verified both roles as
  converged.

Earlier non-pass proof runs are retained rather than rewritten: planning run
`29625238825` rejected an unsupported mandate input before effects, apply run
`29625355700` rejected an actor outside the normalized role set before effects, and apply
run `29625450408` retained the provider-denial mismatch without claiming a proof pass.
The subsequent contained replans and fixes demonstrate that recovery did not require a
new approval or broaden the approved target.

Repository validation, native CI, and distinct review remain PR completion gates; they
do not change the recorded live sandbox qualification result.
