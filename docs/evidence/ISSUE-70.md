# Issue #70 — Phase 3 executable-work decomposition

**Date:** 2026-07-15  
**Issue:** [#70](https://github.com/dragondad22/codex-starter-kit/issues/70)

## Decision and authority

After review-assurance issue #62, authentication/transport research #63, Work Manager
prototype #64, and live qualification research #68 completed, the product owner approved
the proposed six-slice decomposition. No unresolved contract branch required another
question or research ticket. The accepted answer is promoted into decision-map ticket
#5; the live Project and native issues remain execution authority.

## Published decomposition

Issues #71–#76 are two-layer tracer-bullet children of feature #4. Each has an end-to-end
human outcome, production brief, acceptance and negative paths, verification,
documentation/evidence impact, milestone, executor routing, and explicit blockers.
Each body also records current context; stable governing decisions, personas, policy, and
specifications; scope and exclusions; dependencies/human actions; and both checked
readiness assertions from the executable-work template. Existing #15 and #46 were
retrofitted to the same current contract before their future dependency promotion.

The graph is:

```text
#71 → #72 → #73
#73 → #15 → #74
#73 → #46 → #74
#73 → #74
#72 → #75; #74 → #75
#73 → #76; #74 → #76; #75 → #76
```

#75 also records #72 directly because the delivery slice consumes the GitHub adapter as
well as governed-work behavior. #74 consumes existing #15 and #46 instead of duplicating
their closed-state reconciliation and Phase work. Closed #16 remains the governing input
for sparse question/research forms, readiness, promotion, and completion.
Within that boundary, #73 owns built-in close-to-Done bootstrap/configuration and fixture
proof; #15 distinctly owns the production reconciliation/audit backstop for automation
that is absent, delayed, missed, denied, or partial.

## Project reconciliation

- #71 is `Ready` and deliberately selected as Status `Next`.
- #72–#76 are Status `Backlog` / Readiness `Blocked` by their native dependencies or the
  explicit #73 provisioning authority gate.
- #15 and #46 are Status `Backlog` / Readiness `Blocked` by #73 before their live
  integration into #74.
- #4 is Status `In progress` / Readiness `Ready` as the decomposed delivery container;
  child Readiness still controls which work may execute.
- #70 is Status `In progress` while this documentation/evidence promotion is delivered;
  its completing PR closes it and moves it to `Done`.

Removing `needs-triage` from #4 records that no parent-level refinement remains. Child
Readiness still controls execution independently; parent Ready never unblocks a child.

## Verification and limitations

The publishing audit verifies issue bodies, native parent relationships, blocker
relationships, labels, Milestone membership, Project Status/Readiness, and preservation
of #15, #16, and #46. Repository documentation tests, validation, and native Go tests run
on the completing source revision.

This decomposition does not provision GitHub sandboxes, register Apps, create or broaden
credentials, enable paid services, implement production behavior, or turn the unexecuted
#68 qualification rows into passes. Those effects remain governed by their child issues.
