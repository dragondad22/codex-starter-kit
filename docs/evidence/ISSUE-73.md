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

## Applied baseline

- Project Status, Readiness, Horizon, and Phase fields/options were created and re-read.
- Execution table and Horizon roadmap views were created and re-read.
- Reviewer machine user `american-dragon-designs` (account ID `305973890`) has repository
  Write, with no organization-owner or repository-admin authority.
- Reconciler, seeder, and rules Apps are installed only on the sandbox repository. Four
  Actions environments isolate reconciler, seeder, rules, and reviewer secrets.
- Managed labels `type:task`, `ready-for-agent`, and `contract-run` were created and
  re-read by immutable node ID.
- The built-in Item closed workflow is enabled. Auto-add remains a named human UI
  checkpoint until its exact repository and `is:issue label:contract-run` filter are
  configured and re-read.

Secret values were never read. Environment-secret metadata established that the three
App private keys and reviewer token exist. The current seeder installation inventory is
explicitly non-pass until GitHub reports Workflows write in addition to Contents, Issues,
and Pull requests write.

## Implemented behavior

The engine supports strict sandbox inspect, immutable plan, apply, verify, status, and
composed bootstrap operations. Tests cover converging and replaying a missing resource,
name collision, stale observation, partial application and restart recovery, exact marked
cleanup that preserves human resources, active-lease refusal, unsupported kinds, and
sensitive-looking manifest rejection. State is integrity protected and receipts contain
no credentials.

The production GitHub sandbox adapter validates three distinct expiring App installation
roles against actor, account, installation, permissions, and the immutable target. It
uses native versioned REST and GraphQL transport to observe managed labels and Project
fields/options/views/workflows. Human-owned workflow configuration reports
`needs-review` when absent rather than selecting a substitute mechanism.

## Open qualification gates

- Seeder installation must report `workflows: write` after the human accepts the updated
  App permission.
- The Project auto-add workflow must be configured in the UI and observed enabled.
- Marked fixture, temporary fixture-only ruleset, reviewer, cleanup, and revocation cases
  must execute with retained redacted receipts.
- The exact completing candidate must pass repository validation, native CI, and distinct
  review before this evidence can record a live pass.

Until those gates complete, live aggregate status is `not-configured`; deterministic
engine and HTTP-fixture results do not substitute for it.
