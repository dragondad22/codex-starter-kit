# Issue #29 — Safe create replay, interruption, and reconciliation

**Date:** 2026-07-13
**Issue:** [#29](https://github.com/dragondad22/codex-starter-kit/issues/29)
**Parent:** [#2](https://github.com/dragondad22/codex-starter-kit/issues/2)

## Delivered outcome

The create-v1 lifecycle seam now treats an immutable plan as a durable recovery handle.
Exact applied and no-change replays are stable and non-mutating. An interrupted same-plan
apply may resume only when every committed artifact matches the plan and authoritative
state is still absent. Differing approved input, new human content, or an unknown staging
tree produces structured reconciliation without replacement.

Apply serializes mutation with a token-owned JSON lease. It never steals an active, recent,
or malformed lease. A dead stale lease is archived before recovery, and an abandoned stage
is removed only when its directory and transaction marker are bound to that archived lease
and the same plan. Stage content is digested into evidence before removal. Status reports
partial setup as `setup_incomplete` with safe recovery instructions and available evidence.

The protocol stages content, commits authoritative state last, verifies the complete
managed contract, and compensates ordinary commit/postcondition failures by removing files
created by the current attempt. This bounds truthful local recovery; it does not claim one
atomic transaction across multiple filesystem paths or future external effects.

## Failure-mode and transition matrix

| Transition or failure | Detection and stop condition | Result and evidence | Recovery behavior | Excluded external/downstream impact |
|---|---|---|---|---|
| Exact replay after `applied` | Managed contract and prior event match the immutable plan | Stable `applied` semantics plus a distinct content-addressed replay event | Return prior outcome without repository mutation | #30 proves native equivalence of this observable behavior |
| Exact replay after `no_change` | Managed contract and no-change event match the plan | Stable `no_change` semantics plus a distinct content-addressed replay event | Return prior outcome without repository mutation | Same #30 runtime-publication dependency |
| Repository or Git changed after planning | Recomputed content/Git precondition differs before mutation | Failed precondition event with redacted cause | Preserve content; inspect, resolve, and create a new reviewed plan | Remote state and future adapters are not modified |
| Changed approved create input | Existing managed contract does not match newly approved inputs | `ReconciliationRequired` with conflicts, problems, and actions | Preserve repository; explicitly review a later reconciliation/upgrade operation | Upgrade authorization is outside create-v1 |
| Existing user or new human content | Existing or newly introduced paths are not authorized by the plan | Structured reconciliation; conflicting paths are redacted when secret-shaped | Preserve every conflict; obtain explicit reconciliation authority | Retrofit remains a later lifecycle operation |
| Active or recent lifecycle lease | Lease parses and is recent or native process remains alive | Recoverable lock failure; Git-local attempt evidence where writable | Wait and replay the same immutable plan | No distributed or remote lock is claimed |
| Malformed lifecycle lease | Lease cannot be safely authenticated | Recoverable lock failure; lock remains untouched | Preserve lease for human inspection and authorized repair | No inference of ownership from malformed state |
| Dead stale lifecycle lease | Lease age exceeds the bound and native liveness says owner is gone | Lease archived under Git-local attempt evidence | Acquire a new token-owned lease and continue same plan | #30 validates native liveness semantics on supported systems |
| Abandoned matching stage | Stage name and marker match archived stale token and current plan | Full tree digest and recovery record written before removal | Remove only authenticated abandoned stage, then rebuild staging | External temporary stores are outside this local protocol |
| Unknown or mismatched stage lookalike | Reserved-prefix tree lacks matching stale lease/marker/plan | Failed `recover-stage`; artifact preserved | Explicit reconciliation; never delete based on prefix alone | User content cannot be classified as engine-owned implicitly |
| Partial matching committed prefix | State absent; each existing planned artifact exactly matches same plan | Recovery actions/evidence returned with final successful result | Preserve matching files and create only missing artifacts | File metadata atomicity beyond content is not claimed |
| Partial prefix has differing/link/unplanned content | Any existing artifact fails exact plan or portable-path semantics | Failed `reconcile` with structured conflicts | Preserve content and require review | No silent overwrite or semantic merge |
| Operation event exists before authoritative state | Event path appears in an otherwise partial prefix | Failed `reconcile`; event is preserved as a conflict | Review the event and repository evidence explicitly; do not trust self-asserted success | A self-consistent event is not proof that postconditions completed |
| Commit write/rename failure | Native filesystem operation returns an error | Failed event names stage, changed files, recovery, and evidence | Remove files created by current attempt where possible | Multi-path filesystem commit is compensating, not atomic |
| Postcondition validation failure | Complete managed contract fails after commit | Failure cannot become successful state | Roll back current-attempt files where possible; retry only after diagnosis | No evidence means no pass |
| Rollback failure | Compensation cannot restore the pre-attempt local surface | Explicit non-recoverable failure with retained changed-file list | Preserve evidence and require human reconciliation | Automated destructive repair is not authorized |
| Hard termination between file renames | Later status finds managed markers without authoritative state, or a lease/stage remains | `setup_incomplete`, recovery actions, and discoverable evidence refs | Replay the exact plan; matching prefix resumes, conflicts stop | Crash-atomic multi-file mutation is explicitly excluded |
| Future network, service, package, or plugin effect | No such effect exists in create-v1 | Not covered by `create-recovery:v1` capability | Each future adapter must define its own idempotency, evidence, and compensation | Downstream adapter work cannot inherit a false recovery claim |

## Named verification evidence

- `TestIdenticalAppliedCreatePlanReturnsStableIdempotentResult`
- `TestCreateAfterApplyReturnsExplicitNoChange`
- `TestRecoveryBearingCreateAndNoChangePlansRemainIdempotent`
- `TestCreateReturnsReviewableReconciliationForExistingUserContent`
- `TestApplyPersistsReviewableReconciliationWithoutReplacingNewHumanWork`
- `TestApplyRecoversDeadStaleLifecycleLeaseWithEvidence`
- `TestApplyDoesNotStealLiveLifecycleLease`
- `TestApplyResumesInterruptedMatchingCreateWithoutReplacingCommittedPrefix`
- `TestInterruptedResumeRejectsSelfConsistentForgedOperationEvent`
- `TestApplyPreservesUnrecognizedStagingLookalikeForReconciliation`
- `TestStatusExplainsIncompleteCreateAndSafeRecovery`
- `TestApplyRollsBackWhenCommitFails`
- `TestApplyRollsBackWhenPostconditionValidationFails`
- `TestCreateCommandEmitsStructuredReconciliation`
- `TestApplyCommandPreservesStructuredConflictAndRecoveryDetails`

## Verification commands

```text
go test ./...
go vet ./...
GOOS=windows GOARCH=amd64 go test -c ./engine
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
git diff --check
```

Final native Linux, macOS, and Windows results are retained in the completing pull request.

## Remaining limits and downstream impact

- #30 owns native semantic-equivalence closure, released OS/architecture/filesystem claims,
  and reparse-point capability evidence.
- Secret scanning remains `not-configured`. Recovery remains `needs-review` until #30 binds
  current native evidence to the executing build, so aggregate seed verification cannot pass.
- Retrofit, upgrade, plugin, package, remote-service, and release operations must define
  their own reconciliation and external-effect compensation before claiming recovery.
- Create-v1 recovery preserves content and reports safe actions; it does not infer human
  authority to merge or delete ambiguous material.
