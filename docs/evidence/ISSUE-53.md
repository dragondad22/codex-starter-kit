# Issue #53 — Guided truthful verification record

**Date:** 2026-07-14

**Change owner:** dragondad22

**Issue:** [#53](https://github.com/dragondad22/codex-starter-kit/issues/53)

## Delivered slice

- Updated the skills-only plugin to version `0.3.0` and added only the focused
  `$starter-kit-verify` skill beside create/status.
- Added capability, explicit metadata, immutable plan, exact execution approval, local
  evidence-effect, result, redaction, fallback, and failure contracts.
- Preserved all six control states and accepted-exception underlying state; aggregate pass
  fails closed on any non-pass, absent required evidence, incomplete coverage, or evaluator
  failure.
- Added deterministic development/CI evaluation code and scenarios. It is neither runtime
  plugin code nor a control/conformance authority.
- Added install/update, authority, evidence, sensitive-data, compatibility, offline,
  unsupported-surface, direct-engine/CI fallback, and current packaging limitations.

Verification-only is not read-only: it permits only explicitly authorized bounded local
evidence regeneration while prohibiting create/apply/migration effects. The plugin never
opens/transmits referenced evidence automatically, installs evaluators, accepts risk, or
overrides engine results.

## Scenario coverage

The checked-in suite covers explicit and implicit routing, unrelated negative routing,
every `pass`, `fail`, `not-applicable`, `not-configured`, `needs-review`, and
`accepted-exception` state, underlying exception state, mixed aggregate priority, declined
execution approval, stale plan, malformed output, evaluator failure, redacted diagnostics,
and plugin-unavailable fallback. Results preserve mode, invocation boundary, controls,
coverage, diagnostics, and evidence/event paths.

Live model/context-budget/handoff qualification remains #54 rather than being represented
as a deterministic scenario pass.

## Verification and limitations

The local Python suite passed 27 tests. Documentation validation, all three skill
validators, plugin validation, and `git diff --check` passed. An isolated temporary
`CODEX_HOME` added the local marketplace, installed/listed plugin `0.3.0` as enabled, and
confirmed only cached `create`, `status`, and `verify` skills. Expected `/tmp` PATH-helper
warnings have no effect on this skills-only plugin. The user's real Codex profile and
repositories were unchanged.

Local Go remains unavailable; Go and native Linux/macOS/Windows evidence must pass in CI
before merge, and unavailable local Go is not a pass.

No verified packaged engine exists, so an ordinary development installation remains
`degraded-guidance` and performs no verification through the plugin. The conditional
qualified path and every explicit state are exercised by deterministic scenarios. Live
client/model qualification, packaged provenance, signed policy distribution, approved
secret scanning, public publication, and production assurance remain downstream work.
