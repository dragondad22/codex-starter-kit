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

The glossary now keeps every allowed Readiness, Status, and Horizon value beneath its
field definition. Standalone `Ready`, `Now (Horizon)`, `Next (Horizon)`, and
`Later (Horizon)` entries were removed so readers do not have to search elsewhere to
understand a field's complete state model.

## Remaining issue scope

The lifecycle engine and production GitHub adapter now implement the repeatable
reconciliation backstop through the Work Manager seam. One immutable plan can correct the
selected closed/open item, its parent, and its direct dependents. The plan is bound to the
governed source, normalized observation, target/configuration IDs, actor, permission,
expiry, and exact semantic before/after states.

The policy maps closed items to Status `Done`, reopens items only to the explicitly
supplied lifecycle state, keeps a partially delivered parent `In progress`, and closes an
all-children-complete parent only after an explicit completion-contract result. A parent
with every child closed but no satisfied completion result is rejected. A fully specified
dependent becomes Ready only after its final blocker closes and remains Backlog unless it
was separately selected.

Related effects are independently receipted. A selected-item success followed by a denied
parent correction retains both results; refreshed inspection produces a plan containing
only the unconverged parent and dependent. Parent closure uses a state-only GitHub patch
and preserves human-owned title, body, and labels.

This is a partial #15 result. The current route receives bounded relationship facts from
its caller; it does not yet refresh native GitHub hierarchy and dependency observations.
#15 retains that work because #74 consumes the completed reconciliation contract. Native
relationships must be observed rather than inferred from issue prose.

## Verification

The first public-seam test was recorded RED before production types existed: compilation
failed because parent completion, direct dependent context, and related observations were
absent. It then passed after the engine planned, applied, and verified the selected,
parent, and dependent corrections.

Focused deterministic coverage now includes:

- closed-item Status repair, incomplete-parent progress, and final-blocker promotion;
- all-children-complete parent closure plus effect-free replay;
- rejection of unexplained open parents after every child closes;
- no dependent promotion while any blocker remains open;
- rejection of a direct dependent cycle before durable state;
- reopening to an explicit lifecycle state;
- partial related-effect denial, restart, residual-only plan, and convergence;
- bounded production-adapter observation of selected and related immutable identities;
- parent issue closure and Status correction without rewriting human content; and
- existing stale target/configuration, permission denial, partial response, rate,
  ambiguity, replay, and native portability cases.

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
go test ./...
go run ./cmd/starter-kit changes check --repository .
git diff --check
```

The exact completing revision must pass native Linux, macOS, and Windows CI and a distinct
Standards/Spec review before merge. The approved #73 sandbox supplies only built-in
close-to-Done evidence. #15's native relationship observation and `GH-WORK-08` multi-item
live result are `not-configured` pending renewed bounded authority; they remain #15
acceptance gates and must pass before this issue or its draft PR can complete.
