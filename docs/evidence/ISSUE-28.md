# Issue #28 — Hostile path and input threat coverage

**Date:** 2026-07-13  
**Issue:** [#28](https://github.com/dragondad22/codex-starter-kit/issues/28)  
**Parent:** [#2](https://github.com/dragondad22/codex-starter-kit/issues/2)

## Delivered outcome

The Phase 1 create tracer bullet rejects hostile filesystem namespaces, forged create
contracts, secret-bearing plan material, unsafe Git environment overrides, and
self-consistent malicious managed state through the lifecycle-engine interface. Rejected
pre-transaction plans produce no managed-repository effects and emit redacted structured
failure evidence in the Git-local attempt ledger; malformed managed repositories remain
visible as `managed_degraded` with redacted diagnostics.

Invalid identity, schema, approval, digest, and repository-authority input is rejected
before operation acceptance and cannot select its own evidence destination. Tests require
that boundary to remain free of repository/Git effects and secret-bearing diagnostics.

The path and process implementations remain deep internal modules. Plugin, CLI, CI, and
tests continue to use the same lifecycle operations rather than learning filesystem or
process-defense details.

## Threat coverage

| Attack class | Operations | Expected result | Named test evidence | Exclusions | Residual risk | Downstream consumers |
|---|---|---|---|---|---|---|
| Traversal, Unix/Windows absolute injection, empty/relative segments | `apply` | Reject before managed effects; record redacted failure | `TestApplyRejectsSelfConsistentPlanPathOutsideRepository`; `TestApplyRejectsUnsafeCrossPlatformPathNamespaceBeforeManagedEffects` | No retrofit/upgrade paths exist in this slice. | New operation types could bypass the create-v1 contract if they do not reuse the policy. | Retrofit, upgrade, plugin adapters. |
| Reserved device names, trailing-dot aliases, ambiguous normalization | `apply` | Reject universally before managed effects | `TestApplyRejectsUnsafeCrossPlatformPathNamespaceBeforeManagedEffects` | Unicode is excluded from the v1 portable namespace. | A future Unicode policy requires its own versioned decision and fixtures. | Every later path-producing operation. |
| Case-fold collisions and existing user directory collision | `create`, `apply` | Reject; preserve user-owned namespace on the runner's detected case mode | `TestCreatePreservesExistingDirectoryThatCaseCollidesWithManagedPath`; `TestApplyRejectsUnsafeCrossPlatformPathNamespaceBeforeManagedEffects` | Retrofit reconciliation is not implemented. | #30 must publish exact filesystem case-mode support claims. | Retrofit and native support publication. |
| Repository ancestor aliases/root links, reserved parents, or managed artifact symlink/junction escape | `create`, `status`, `apply` | Canonicalize ancestor aliases into the plan; reject final-root or managed-path links; report drift as `managed_degraded` | `TestCreateCanonicalizesRepositoryRootBelowSymlinkedAncestor`; `TestCreateRejectsSymlinkRepositoryRoot`; `TestCreateRejectsReservedDirectorySymlinkEscapeDuringPlanning`; `TestCreateRejectsReservedDirectoryJunctionEscapeDuringPlanning`; `TestStatusRejectsManagedArtifactSymlinkEvenWhenContentDigestMatches` | Native symlink fixtures skip only where link creation is unavailable; Windows junction coverage is Windows-only. | #30 must publish broader reparse-point and capability evidence before runtime support. | Native support, retrofit, status, verification. |
| Added paths, forged result paths, ownership, or provenance | `apply`, `status` | Reject exact-contract mismatch before transaction or degrade persisted state | `TestApplyRejectsCreatePlanThatExpandsOrReclassifiesManagedWrites`; `TestApplyRejectsSecretBearingForgedResultPathBeforeManagedEffects`; `TestStatusFailsClosedForSelfConsistentAdversarialProvenance` | Only create-v1 artifacts are authorized. | Retrofit and upgrade need separately approved operation-specific contracts. | Retrofit, upgrade, policy and plugin writers. |
| Shell metacharacters, inherited Git overrides, or repository-local executable Git config | `inspect`, `create`, `apply` | Treat repository as one argument; ignore hostile environment and executable config | `TestLifecycleGitExecutionTreatsRepositoryMetacharactersAsOneArgument`; `TestLifecycleGitExecutionIgnoresHostileGitEnvironmentOverrides`; `TestInspectDoesNotExecuteRepositoryLocalFilesystemMonitor` | The Unix fsmonitor execution fixture is Unix-only; Windows still runs the same production override. | Future Git commands must assess additional command-specific executable configuration. | Every Git-backed engine operation and adapter. |
| Fixture secrets in briefs, repository paths, verification metadata, plan content/paths/digests, or state diagnostics | `create`, `apply`, `verify`, `status` | Reject before plan/staging; emit only redacted diagnostics/evidence | `TestCreateRejectsFixtureSecretWithoutEchoingItIntoDiagnostics`; `TestCreateRejectsSecretBearingRepositoryPathWithoutEchoingIt`; `TestPrepareVerifyRejectsSecretBearingMetadataBeforePlanOrEvidence`; `TestVerifyRejectsSecretBearingSelfConsistentMetadataBeforeEvidence`; `TestApplyRejectsFixtureSecretInSelfConsistentPlanBeforeStaging`; `TestApplyDoesNotEchoFixtureSecretFromHostilePlanPath`; `TestApplyRejectsSecretBearingForgedResultPathBeforeManagedEffects`; `TestApplyRejectsSecretBearingRepositoryDigestBeforeGeneratingEvidence`; `TestStatusRedactsFixtureSecretFromAdversarialOwnershipData` | Pattern rejection is not a comprehensive scanner. | `CORE-SECRETS-001` remains `not-configured`. | Plans, events, issue evidence, plugins, release gates. |
| Self-consistent malicious layout, routes, lifecycle state, ownership, or provenance | `inspect`, `status`, `create`, `verify` | Fail closed as `managed_degraded`; no false managed/no-change result | `TestStatusFailsClosedForSelfConsistentAdversarialManagedState`; `TestStatusFailsClosedForSelfConsistentAdversarialProvenance` | Only schema v1 is accepted. | Later schema versions need explicit migration validation. | Status, verification, upgrade and policy routing. |

## Native semantics and exclusions

Portable restrictions are enforced identically on Linux, macOS, and Windows rather than
depending on the current host filesystem. This intentionally rejects names that might be
legal on one host but unsafe or ambiguous on another. Native CI runs the same engine seam
suite on all three hosts.

Symlink fixtures use the native Go filesystem operation and skip when the runner cannot
create links. That skip is capability evidence, not a pass for link behavior. Windows CI
also creates a directory junction with the native command processor and verifies that the
reserved directory cannot escape. The case-collision fixture records whether `DOCS` and
`docs` resolve to the same native object and requires preservation in either mode. Issue
#30 must publish exact runner/filesystem assumptions, broader reparse-point capability,
and any unsupported state before runtime support is published.

Issue #29 owns interruption, stale locks, reconciliation results, and stronger recovery.
Later retrofit, upgrade, plugin, policy, and release work consumes the portable path,
operation-specific write contract, structured process, redaction, and degraded-state
semantics established here.

## Verification

```text
go test ./...
go vet ./...
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
git diff --check
```
