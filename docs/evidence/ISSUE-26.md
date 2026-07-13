# Issue #26 — Create walking-skeleton coverage and evidence

**Date:** 2026-07-12
**Issue:** [#26](https://github.com/dragondad22/codex-starter-kit/issues/26)
**Parent:** [#2](https://github.com/dragondad22/codex-starter-kit/issues/2)

## Delivered outcome

The Go lifecycle engine and `starter-kit` CLI expose the first complete empty-repository
path through `inspect`, `create`, `plan`, `apply`, and `status`. A caller reviews a
schema-versioned JSON plan, retains its SHA-256 plan identifier, and supplies both to
apply. Apply rechecks identity and repository preconditions, refuses existing targets,
writes authoritative state last, verifies content digests, and returns a structured
result. An unchanged managed repository produces explicit `no_change` semantics.

The created repository separates managed machine state, generated projections, and
human-owned records. It includes stable routes, ownership/provenance digests, a concise
orientation, approved seed brief, confirmed persona registry, decision index, and a
truthful conformance summary that claims no verified controls.

## Acceptance coverage

| Issue #26 requirement | Evidence | Disposition |
|---|---|---|
| Engine-seam `inspect`, `create`, `plan`, `apply`, `status` against temporary real Git | `engine/engine_integration_test.go`, `cli/cli_integration_test.go` | Covered |
| Apply consumes plan identifier and rechecks preconditions | Plan hash verification plus changed-repository rejection test | Covered for empty create; deeper stale/conflict semantics are #29 |
| Minimal authoritative state, provenance, routes, orientation, brief, decisions, personas, summary | Deterministic create plan and applied-content assertions | Covered |
| Distinct human, generated, and managed ownership without silent overwrite | Managed-file manifest, ownership assertions, existing-target refusal | Covered for initial create; reconciliation is #29 |
| Platform-neutral paths and structured commands | Native Go path APIs and `exec.CommandContext` Git arguments; native CI | Covered for happy path; adversarial platform cases are #28/#30 |
| Stable rerun or explicit no change | Create-after-apply integration test | Covered |
| Tests cross lifecycle-engine seam | Public `engine` methods and CLI JSON adapter tests | Covered |
| Coverage and downstream effects explicit | This record and `LIFECYCLE_ENGINE.md` | Covered |

## Phase 1 roadmap coverage

| Phase 1 roadmap obligation | Owner/status after #26 | Downstream impact |
|---|---|---|
| Schemas for facts, policy lock, layout, managed files, plans, results, routes, evidence | Seed v1 JSON documents/results implemented; policy is explicitly `not_configured`; semantic expansion remains issue-owned | #27 adds verification evidence; #29 adds recovery state; later phases version/migrate |
| Engine language/package/signing evaluation | DEC-0015 / #25 complete | All engine and distribution consumers |
| Local filesystem/Git `inspect`, `plan`, `apply`, `status`, `verify` | All except `verify` implemented | #27 owns `verify` |
| Seed `core-trust` controls | Not implemented | #27 owns truthful evaluation and evidence |
| Minimal rendered orientation, brief, decisions, conformance | Implemented | #27 regenerates conformance from verification state |
| Seed and confirm persona registry with stable IDs | Initial `PER-OWNER` confirmed and routed | Later inception/persona work expands evidence and audiences |
| Idempotence, preconditions, conflicts, rollback, malicious paths, native equivalence | Stable no-change and basic precondition checks only | #28 hostile inputs, #29 recovery/conflicts, #30 full native equivalence |

## Verification commands

```text
go test ./...
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
git diff --check
```

Final local results and native CI evidence are recorded in the completing pull request.

## Explicit exclusions and limitations

- No control is verified and no conformance pass is emitted; #27 owns that result model.
- Writes are not yet a fully staged/locked/rollback-capable transaction; #29 owns it.
- Hostile paths, symlinks/junctions, reserved names, case collisions, and malicious plan
  documents are not yet an approved security seam; #28 owns them.
- Exact released runtime support, installers, signing, and native semantic equivalence
  remain #30 and the release workflow.
- GitHub, policy registry, plugin, retrofit, upgrade, and release adapters remain later
  roadmap slices and consume the language-neutral JSON/operation seam.
