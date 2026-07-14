# Issue #51 — Installable plugin status tracer record

**Date:** 2026-07-14

**Change owner:** dragondad22

**Issue:** [#51](https://github.com/dragondad22/codex-starter-kit/issues/51)

## Delivered slice

- Added a valid skills-only plugin manifest and repository development marketplace.
- Added the single progressively disclosed `$starter-kit-status` skill, its fail-closed
  compatibility/presentation contract, and positive/negative scenario fixtures.
- Added the standalone engine's non-mutating `capabilities` metadata command. It reports
  build, protocol, operation, and schema facts while retaining `unverified` provenance.
- Added a deterministic evaluation oracle for development/CI only. It is not a runtime
  dependency and does not replace the engine or the skill.
- Added installation, authority, compatibility, fallback, surface, offline, and current
  packaging limitations to the product and architecture documentation.

The plugin uses no app, connector, MCP server, hook, scheduled task, browser, telemetry,
or shell wrapper. Capability detection and plugin installation have no repository input or
repository mutation. Status is invoked only after compatibility, external provenance, and
read authority are established; every engine result list and non-pass lifecycle is
preserved.

## Scenario coverage

The checked-in evaluation set covers explicit and implicit routing, unrelated negative
routing, managed and unmanaged repositories, `managed_degraded` non-pass output, malformed
output, missing engine, incompatible engine, unverified engine, and administrative
unavailability. The policy oracle checks exact protocol/operation/schema facts, provenance,
authority, status envelope shape, capability mode, invocation boundary, and preservation
of engine diagnostics.

Live model routing is not represented as deterministic native support by these fixtures.
Cross-model/client context-budget and handoff qualification remains issue #54.

## Local verification

The supported plugin and skill validators passed against the final local file tree:

```text
python3 /home/chris/.codex/skills/.system/skill-creator/scripts/quick_validate.py plugins/codex-starter-kit/skills/status
python3 /home/chris/.codex/skills/.system/plugin-creator/scripts/validate_plugin.py plugins/codex-starter-kit
```

An isolated temporary `CODEX_HOME` then successfully added the repository marketplace,
listed `codex-starter-kit@codex-starter-kit-development` as available, installed version
`0.1.0`, and listed it as installed and enabled. The temporary profile emitted only the
expected warning that PATH helper aliases are refused under `/tmp`; the skills-only plugin
requires no helper alias. The user's real Codex profile and plugin configuration were not
changed.

The local Python suite passed 24 tests; documentation validation, plugin validation, skill
validation, and `git diff --check` also passed. The required local Go command remains
unavailable because this host has no Go executable; that unavailable local capability was
not converted to a pass.

Draft PR #56 run `29296901686` tested source commit
`58104bb99b373a0b9c8729857f386d409d48a852` using pinned Python 3.12 and Go 1.26.5.
The Linux, macOS, and Windows foundation jobs all passed the Python suite, documentation
validation, `go test ./...`, and native evidence capture. Phase 1 native semantic
equivalence and the aggregate required-check gate also passed. The final-head checks
triggered by this evidence update supersede that initial run when determining merge
readiness.

## Limitations and downstream work

No verified packaged engine is published, so the checked-in development engine remains
unverified and the installed plugin truthfully selects `degraded-guidance` rather than
executing status. Packaging/signing/retained artifact qualification remains required for a
supported executable. Guided create and verify remain #52 and #53. Client/model/native
qualification and handoffs remain #54; IDE distribution remains `needs-review` because
official documentation conflicts. Public submission, publisher identity, legal/support
materials, and publication remain Phase 6.
