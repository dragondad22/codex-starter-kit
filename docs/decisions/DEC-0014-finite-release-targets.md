# DEC-0014 — Finite release targets

**Status:** Accepted
**Owner:** dragondad22
**Date:** 2026-07-12
**Source decision:** D14

## Context

Horizon communicates rolling product direction but does not define finite release
membership. Treating every `Now` feature as part of the current release would make the
release expand whenever direction changes, while milestone completion alone would not
prove that the assembled product passed release-wide gates.

## Decision

Keep Horizon and release targeting separate. `Now` means committed current product
direction, `Next` means an intentional candidate not yet committed to a release, `Later`
means plausible future direction without a release commitment, and blank means
unclassified or outside the feature roadmap.

Use exactly one native GitHub Milestone as the finite manifest for each named release.
Assign work only after its release membership is approved. Each release also has one
aggregate executable release issue that owns included scope and exclusions, required
outcomes, release-wide gates and evidence, known limitations, approvals, publication,
rollback, communication, and completion.

A release follows the S.M.A.R.T. Release contract: Scoped, Measurable, Approved,
Releasable, and Triggered. Its trigger may be outcome-, time-, event-, or hybrid-bound;
time never overrides a prohibited or unresolved gate. Release triggering is separate from
version selection, which remains contextual under DEC-0008. Where SemVer applies,
`1.0.0` establishes the initial stable supported contract.

Release scope progresses from proposed to committed to release candidate to released.
Additions, removals, and deferrals retain rationale and impact. Release-candidate additions
normally address blockers or required corrections. Deferral cannot hide a required failed
gate or remove an outcome while leaving the release claim unchanged.

## Consequences

The Project can show rolling direction and finite release membership without duplicating
either. A release is ready only when its committed scope is completed or validly
dispositioned and its aggregate issue has sufficient current evidence and approval;
milestone percentage and an empty `Now` column are never sufficient. The actual 1.0.0
outcome and membership require separate human approval before its Milestone is created.

## Source

[Discovery decision D14](../discovery/CODEX_STARTER_KIT_REVIEW.md#d14), approved through
[issue #22](https://github.com/dragondad22/codex-starter-kit/issues/22).
