# Issue #27 — Seed core-trust verification coverage

**Date:** 2026-07-12
**Issue:** [#27](https://github.com/dragondad22/codex-starter-kit/issues/27)
**Parent:** [#2](https://github.com/dragondad22/codex-starter-kit/issues/2)

## Delivered outcome

The lifecycle-engine and CLI prepare/execute `verify` transaction evaluates six stable seed controls for an
explicit repository scope and lifecycle gate. Results preserve explicit states and
aggregate to `pass` only when every result passes. Secret scanning remains
`not-configured` rather than becoming false green. Issue #29 supplies a bounded create-v1
recovery protocol, but the recovery control remains `needs-review` until #30 binds the
executing binary to retained, current native test provenance.

Verification writes content-digested, self-describing machine evidence under
`.starter-kit/evidence/`, regenerates the human conformance summary, updates its managed
digest, records the requesting actor and authority in a structured operation event,
validates the resulting contract, and rolls replacements back on ordinary
commit/postcondition failure. A deterministic clock adapter makes controlled results
reproducible. Diagnostics emitted from repository problems pass through secret-pattern
redaction before entering evidence.

Verification commits validate their own evidence and generated projection rather than
requiring the entire repository to pass. A degraded repository can therefore retain a
truthful failed result while remaining `managed_degraded`; failed execution attempts are
recorded in the Git-local attempt ledger when repository evidence cannot be committed.

## Acceptance coverage

| Issue #27 requirement | Evidence | Disposition |
|---|---|---|
| `verify` operates on #26-created repository through engine seam | `TestVerifyCreatedRepositoryEmitsTruthfulSeedResults`, `TestCLIVerifyCreatedRepository` | Covered |
| Truth, secrets, ownership, coverage, recovery, breadcrumb controls | Six stable `CORE-*` results | Covered; secrets are `not-configured`; recovery is `needs-review` pending build/native provenance |
| Stable identity and one explicit state with evidence/rationale | Result schema plus contract validation | Covered |
| Scope, source revision, engine/schema/policy versions, time, redacted diagnostics | `VerificationResult`, injected clock, redaction | Covered; policy explicitly `not-configured` |
| Reproducible human summary from authoritative result | Rendered `CONFORMANCE.md` and updated managed digest | Covered |
| Non-pass fixtures cannot become pass | `TestOverallStateFixturesNeverProducePass`, `TestAggregateNeverConvertsExplicitNonPassStateIntoPass` | Covered |
| Equivalent controlled runs produce equivalent semantics | `TestVerifyEquivalentControlledRepositoriesProducesEquivalentSemantics` | Covered |
| Coverage and downstream consumers explicit | This record and engine interface document | Covered |

## Control and downstream coverage

| Control | Phase 1 promise | Named test evidence | Deferred consumer/impact |
|---|---|---|---|
| `CORE-TRUTH-001` | No evidence means no pass; preserve explicit states | `TestVerifyCreatedRepositoryEmitsTruthfulSeedResults`, `TestOverallStateFixturesNeverProducePass` | All later policy/control packs consume explicit state semantics |
| `CORE-SECRETS-001` | Secrets and sensitive diagnostics are not exposed | `TestDiagnosticsAreRedactedBeforeEvidence` | #28 expands leakage attacks; later policy work selects a scanner |
| `CORE-OWNERSHIP-001` | Human, generated, and machine ownership remain distinct | `TestVerifyCreatedRepositoryEmitsTruthfulSeedResults` | #29 reconciliation and upgrades preserve ownership/history |
| `CORE-COVERAGE-001` | Claims disclose scope and unsupported coverage | `TestVerifyCreatedRepositoryEmitsTruthfulSeedResults` | Plugin, release, and assurance views consume disclosure |
| `CORE-RECOVERY-001` | Recovery cannot pass without current evidence bound to the executing build | `TestVerifyCreatedRepositoryEmitsTruthfulSeedResults`; Issue #29 recovery fixtures | #30 binds released build/native provenance; later adapters add external-effect recovery |
| `CORE-ROUTES-001` | Stable breadcrumb IDs resolve to governed artifacts | `TestRequiredBreadcrumbCannotPassWhenMissing` | Plugin and governed breadcrumb routing consume stable IDs |

## Verification commands

```text
go test ./...
go vet ./...
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
git diff --check
```

Final local results and native CI evidence are retained in the completing pull request.

## Limitations

- No approved secret scanner exists; secret coverage cannot pass.
- #28 owns complete malicious-path/input and secret-leakage coverage.
- #30 owns exact released platform support and cross-platform semantic closure.
- Create-v1 recovery does not claim crash-atomic multi-file mutation or recovery for future
  external adapters; it uses state-last commit, conservative replay, and compensation.
- Human attestation, risk acceptance authorization, qualified review, signed policy packs,
  release gates, and remote evidence stores remain later vertical slices.
