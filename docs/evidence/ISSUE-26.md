# Issue #26 — Create walking-skeleton coverage and evidence

**Date:** 2026-07-12
**Issue:** [#26](https://github.com/dragondad22/codex-starter-kit/issues/26)
**Parent:** [#2](https://github.com/dragondad22/codex-starter-kit/issues/2)

## Delivered outcome

The Go lifecycle engine and `starter-kit` CLI expose the first complete empty-repository
path through `inspect`, `create`, `plan`, `apply`, and `status`. A caller reviews a
schema-versioned JSON plan after explicitly approving the supplied brief and confirming
the owner persona, retains its SHA-256 plan identifier, and supplies both to apply. Apply
rechecks identity and content/Git preconditions, constrains paths, stages and verifies
content, locks the lifecycle, refuses existing targets, writes authoritative state last,
validates the complete contract, rolls back failed commits where possible, and returns a
structured result. Accepted successful, failed, and no-change apply results are recorded
as self-describing machine evidence under the plan-declared `.starter-kit/events/` path;
successful event evidence is staged and committed before authoritative state. Only a valid
unchanged managed repository with the same approved inputs produces `no_change`.

The created repository separates managed machine state, generated projections, and
human-owned records. It includes stable routes, ownership/provenance digests, a concise
orientation, approved seed brief, confirmed persona registry, decision index, and a
truthful conformance summary that claims no verified controls.

## Acceptance coverage

| Issue #26 requirement | Evidence | Disposition |
|---|---|---|
| Engine-seam `inspect`, `create`, `plan`, `apply`, `status` against temporary real Git | `engine/engine_integration_test.go`, `cli/cli_integration_test.go` | Covered |
| Apply consumes plan identifier and rechecks preconditions | Plan hash verification, content/Git snapshot, same-count replacement test, changed-repository rejection | Covered for empty create; deeper conflict/reconciliation semantics are #29 |
| Minimal authoritative state, provenance, routes, orientation, brief, decisions, personas, summary | Deterministic create plan and applied-content assertions | Covered |
| Distinct human, generated, and managed ownership without silent overwrite | Explicit create approvals; self-classifying managed-file manifest; ownership/provenance assertions; existing-target refusal | Covered for initial create; reconciliation is #29 |
| Platform-neutral paths and structured commands | Native Go path APIs, root-constrained clean paths, symlink-parent rejection, and `exec.CommandContext` Git arguments; native CI | Covered baseline; full adversarial platform cases are #28/#30 |
| Stable rerun or explicit no change | Create-after-apply plus managed-drift/degraded-state tests | Covered |
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
| Idempotence, preconditions, conflicts, rollback, malicious paths, native equivalence | Valid-contract no-change, content/Git preconditions, root constraint, staging/lock/state-last commit, postcondition validation, and best-effort rollback baseline | #28 completes hostile inputs, #29 interruption/recovery/conflicts, #30 full native equivalence |

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
- #29 still owns interruption fixtures, stale-lock recovery, durable recovery evidence,
  stronger atomicity, and complete conflict/reconciliation semantics.
- #28 still owns junctions, reserved names, case collisions, malformed state, secret
  leakage, command injection, and the complete malicious-plan matrix.
- Exact released runtime support, installers, signing, and native semantic equivalence
  remain #30 and the release workflow.
- GitHub, policy registry, plugin, retrofit, upgrade, and release adapters remain later
  roadmap slices and consume the language-neutral JSON/operation seam.
