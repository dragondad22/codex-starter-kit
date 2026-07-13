# Issue #23 — 1.0.0 outcome and professional-baseline evidence

**Date:** 2026-07-13
**Issue:** [#23](https://github.com/dragondad22/codex-starter-kit/issues/23)
**Release issue:** [#45](https://github.com/dragondad22/codex-starter-kit/issues/45)
**Milestone:** [1.0.0](https://github.com/dragondad22/codex-starter-kit/milestone/1)

## Delivered contract

- DEC-0016 defines this repository's stable `1.0.0` outcome through Phases 0–6,
  compatibility boundary, outcome trigger, qualification evidence, pilot isolation and
  cleanup, artifact trust, authority separation, limitations, membership, exact candidate,
  and immutable publication history.
- DEC-0017 establishes the universal professional engineering baseline: project size and
  delegated interaction never create a lower-quality passing mode. Engagement,
  project-specific assurance, and evidence presentation remain independently configurable.
- DEC-0011 now routes minimum governed context to the human's natural decision surface
  while preserving one deeper authoritative record.
- Product requirements, personas, lifecycle architecture, roadmap obligations, and
  canonical vocabulary use the same meanings.
- Documentation validation now requires every durable `DEC-NNNN` record—not only the
  original discovery decisions—to be indexed and structurally complete.

## Release and coverage state

The native `1.0.0` Milestone is the finite manifest. Aggregate issue #45 reports the
release as `Committed`, not Release Candidate-ready, and owns the outcome, initial
membership, exclusions, coverage gaps, gates, authority, trigger, and completion contract.

Initial membership includes Phase features #1–#7, their existing Phase 0–6 child work,
the open Project-reconciliation defect #15, initial-release coverage feature #31,
hierarchy correction #32, vocabulary validator #40, the aggregate release issue, and the
three follow-ups created from refinement. Phase 7–8 features #8–#9 and Later features #18,
#19, and #21 remain explicitly outside `1.0.0`.

Aggregate issue #45 records the product owner as the accountable membership owner and
preserves outcome-level rationale, acceptance/evidence expectations, and gap impact for
each admitted phase. Blank implementation assignees therefore do not mean ownerless scope.
The `Needs refinement` state on Phases 2–6 truthfully reports missing executable
decomposition and remains a Release Candidate blocker.

| Phase | Current coverage truth |
|---|---|
| 0 | Feature #1 and its four children are complete; final-candidate evidence must refresh. |
| 1 | Feature #2 and its six children are complete for the source-runtime contract; packaged release evidence remains later. |
| 2 | Feature #3 is required but still needs decomposition; operating-profile decision #47 is Ready for human work. |
| 3 | Feature #4 is partially delivered; #15 and #31 remain gaps and phase-visibility task #46 is Ready. |
| 4 | Feature #5 is required and undecomposed. |
| 5 | Feature #6 is required and undecomposed. |
| 6 | Feature #7 has release-governance decisions but still lacks executable adapter, signing, pilot, publication, rollback, and communication coverage; snapshot/canary selection #48 is Ready for human work. |

## GitHub reconciliation

| Item | Status | Readiness | Horizon / relationship |
|---|---|---|---|
| #3 | `Backlog` | `Needs refinement` | Horizon `Now`; #47 is a native child. |
| #4 | `In progress` | `Needs refinement` | Horizon `Now`; delivered #16 and open #15 plus #46 are native children. |
| #5–#6 | `Backlog` | `Needs refinement` | Horizon `Now`; both require decomposition. |
| #7 | `In progress` | `Needs refinement` | Horizon `Now`; #22, #23, #45, and #48 are native children. |
| #31 | `Backlog` | `Needs refinement` | Horizon `Now`; explicitly committed non-phased release coverage work. |
| #45 | `In progress` | `Blocked` | Native child of #7; blocked by incomplete Phase 2–6 outcomes and release gates. |
| #46 | `Backlog` | `Ready` | Native child of #4 and assigned to `1.0.0`. |
| #47 | `Backlog` | `Ready` | Human-owned question work; native child of #3 and assigned to `1.0.0`. |
| #48 | `Backlog` | `Ready` | Human-owned bounded research; native child of #7 and assigned to `1.0.0`. |

This reconciliation deliberately moves committed features #3–#7 from Horizon `Next` or
`Later` to `Now`. Milestone membership is still separate from Horizon and does not change
their incomplete Readiness state.

## Follow-up boundaries

- [#46](https://github.com/dragondad22/codex-starter-kit/issues/46) implements the Project
  `Phase` field and view without using Milestones or copied child facts.
- [#47](https://github.com/dragondad22/codex-starter-kit/issues/47) resolves the compact
  operating-profile choices above the non-configurable professional baseline.
- [#48](https://github.com/dragondad22/codex-starter-kit/issues/48) selects the real
  qualification snapshots and canaries under the approved research boundary.
- [#31](https://github.com/dragondad22/codex-starter-kit/issues/31) remains responsible for
  a derived initial-release coverage, gap, and downstream-impact view.

This change does not implement release automation, signing infrastructure, pilots, policy
distribution, the Codex plugin, retrofit, or the remaining Phase 2–6 capabilities. It does
not claim that the product is ready for Release Candidate status.

## Verification

The implementation branch uses:

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
go test ./...
git diff --check
```

Local results before internal review: 22 Python tests passed, documentation validation
passed, and the diff check passed. The pinned Go 1.26.5 toolchain is not installed in this
local environment, so `go test ./...` could not run locally. The completing pull request's
native CI must supply that result before the change is ready for review. Any source change
after these results requires affected verification to be rerun.
