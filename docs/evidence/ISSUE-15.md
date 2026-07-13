# Issue #15 — Project-state reconciliation rule

**Date:** 2026-07-13  
**Issue:** [#15](https://github.com/dragondad22/codex-starter-kit/issues/15)

## Confirmed drift

After issues #25, #26, and #27 completed, feature #2 still reported Status `Backlog`.
Issue #28 remained `Blocked` although its only blocker, #26, was complete. Issue #29 also
remained `Blocked` after both #26 and #27 completed. The Project therefore contradicted
its issue dependency and native hierarchy facts.

The live Project was reconciled on 2026-07-13:

| Item | Status | Readiness | Rationale |
|---|---|---|---|
| #2 | `In progress` | `Ready` | Three of six child slices are complete and the feature remains open. |
| #28 | `Next` | `Ready` | #26 is complete and this slice was selected as the immediate next item. |
| #29 | `Backlog` | `Ready` | #26 and #27 are complete, but this slice has not been selected as next. |
| #30 | `Backlog` | `Blocked` | #28 and #29 remain unresolved blockers. |

Issue #30 retains the `ready-for-agent` routing label because its brief is complete and
its intended executor is known. The label contract was clarified in both repository
guidance and the live GitHub label description: it cannot override Readiness `Blocked`.

Feature #1 also remained open and `In progress` although all four native children (#11,
#12, #20, and #32) were closed and `Done`. Its completion condition named delivery
through those children and no outstanding task was recorded, so #1 was closed and moved
to `Done` on 2026-07-13.

## Durable rule

Agent and issue-tracker guidance now treats starting, completion, reopening, dependency
changes, and child transitions as reconciliation checkpoints. The touched item, its
parent, and its direct dependents must be audited. Completing the final blocker promotes
a dependent to Readiness `Ready`; it moves to Status `Next` only through deliberate work
selection. A partially delivered open parent is Status `In progress`, not `Backlog`.
When every child is complete, the parent closes as `Done`; genuinely remaining acceptance
work must be represented by a concrete attached child before the parent stays open.

The field vocabulary also states explicitly that Status `Backlog` is not Horizon `Later`
and is not a synonym for a Scrum product backlog.

## Remaining issue scope

This record establishes the standing behavior and repairs the observed slice. Issue #15
remains open for a supported automation or repeatable reconciliation backstop that proves
closed items and dependency changes cannot silently leave the Project stale.

## Verification

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
go test ./...
```
