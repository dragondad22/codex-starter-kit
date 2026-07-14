# Issue #52 — Guided managed-repository creation record

**Date:** 2026-07-14

**Change owner:** dragondad22

**Issue:** [#52](https://github.com/dragondad22/codex-starter-kit/issues/52)

## Delivered slice

- Updated the skills-only plugin to version `0.2.0` and added only the focused
  `$starter-kit-create` skill beside status.
- Added the create capability/input/notice/inspection/plan/review/apply/result contract
  using structured executable and argument-vector semantics.
- Bundled the approved DEC-0017 professional baseline as a versioned, digest-bound offline
  projection. It is explicitly not a signed policy pack or conformance evidence.
- Kept brief, persona, special-data notice, plan, and effect approvals separate. The skill
  discloses that create-v1 does not persist the special-data declaration.
- Added a deterministic development/CI evaluation oracle and checked-in routing, approval,
  declaration, no-change, conflict, failure, recovery, and prerequisite scenarios. The
  oracle is not a plugin runtime dependency or lifecycle authority.
- Added development install/update, authority, data, cost, compatibility, offline,
  recovery, unsupported-surface, and fallback documentation.

The plugin neither installs nor updates an engine/baseline/repository. It has no optional
integration, external-service, network, or shell-wrapper capability. Planning is read-only;
apply consumes the exact privately retained plan and separately retained ID only after
specific effect approval.

## Scenario coverage

The suite covers explicit and implicit create routing, unrelated negative routing, the
qualified applied path, exact no-change replay, declined apply approval, absent human
authority, each `No`/`Yes`/`Unsure` declaration, missing notice acknowledgment, malformed
plan, stale precondition, existing-content reconciliation before apply, interrupted setup,
non-recoverable rollback failure, missing engine, and missing baseline.

Results preserve outcome, invocation boundary, recoverability, conflicts, recovery, and
evidence. Live model/context-budget/handoff qualification remains #54 rather than being
represented as a deterministic scenario pass.

## Verification and limitations

The local Python suite passed 25 tests. Documentation validation, both skill validators,
plugin validation, and `git diff --check` passed. An isolated temporary `CODEX_HOME`
successfully added the repository marketplace, installed plugin version `0.2.0`, listed it
as installed/enabled, and exposed only the cached `create` and `status` skills. Its expected
warning refused PATH helper aliases under `/tmp`; this skills-only plugin requires none.
The user's real Codex profile, account configuration, and managed repositories were not
changed.

Local Go remains unavailable; required Go and native Linux/macOS/Windows evidence must pass
in CI before merge. That local unavailable capability is not a pass.

No verified packaged engine or signed baseline policy pack exists. The plugin does contain
the approved offline baseline projection, but ordinary local source does not externally
verify the containing snapshot. The implemented full path is therefore exercised against
qualified scenario facts, while an ordinary development installation truthfully remains
`degraded-guidance` and performs no create operation. The create-v1 engine also does not
persist the special-data declaration. Engine/policy packaging, declaration persistence,
live client/model qualification, guided verify, public publication, and production
assurance remain downstream work under their governing issues/phases.
