# Issue #100 — Validation-manifest decision record

**Date:** 2026-07-27
**Change owner:** dragondad22
**Issue:** [#100](https://github.com/dragondad22/codex-starter-kit/issues/100)
**Parent:** [#83](https://github.com/dragondad22/codex-starter-kit/issues/83)

## Approval and promotion

The product owner approved the validation-manifest authority and lifecycle through a
one-question-at-a-time resolution of issue #100. DEC-0023 is the durable authority. It
defines a generated immutable checklist whose exact governed inputs and derived
assertion/evidence obligations determine its identity without creating requirements,
approving work, or authorizing effects.

The owner rejected a universal separate-approval rule. Human-owned standing policy or a
bounded per-work decision determines the approval checkpoint using applicable work,
actor, risk, timing, cost, environment, and effect facts. DEC-0022 remains the separate
effect-authority contract.

## Changed records

- Added DEC-0023 and linked it from the decision and documentation indexes.
- Added `Validation manifest` and `Manifest assessment` to the canonical glossary.
- Added the owner outcome and implementation boundaries to the PRD.
- Assigned source composition, evaluation, immutable storage, lifecycle assessment, and
  approval boundaries across the architecture without selecting a new runtime module.
- Resolved ticket #1 in the assurance-factory decision map while leaving evidence-method
  classification and executable decomposition with #101 and #108.
- Linked DEC-0019 and DEC-0022 to the approval-versus-authority distinction.

## Negative-path disposition

| Scenario | Required result |
|---|---|
| Authoritative sources conflict semantically | `needs-review`; correct the source or record a governed decision/exception, then regenerate |
| A bound source, profile, policy, schema, compiler, assertion, or obligation changes | prior manifest is `stale`; create a new immutable manifest |
| Manifest schema, digest, or provenance is invalid | reject input; do not assign a lifecycle disposition |
| Required evaluator or approval route is absent | `not-configured`; inspection may continue, evaluation/effects stop |
| Required capability is unavailable | `unsupported` |
| Evidence is absent or insufficient | assertion records its required non-pass; manifest never passes |
| Prior evidence looks similar | do not reuse without unchanged fingerprint, permitted method, freshness, scope validity, and attribution |
| Risk is accepted | retain the underlying result |
| A current manifest or approval exists but effect authority does not | stop; neither replaces the DEC-0022 execution mandate |

## Verification evidence

The documentation branch passed the complete documentation-change command set:

```text
python3 -m unittest discover -s tests -p "test_*.py" — 41 tests passed
python3 scripts/validate_docs.py — passed
go test ./... — passed
starter-kit changes validate --repository . — passed; 20 records valid
git diff --check — passed
```

The completing pull request and native CI retain exact source-revision, check, and review
evidence.

## Downstream reconciliation

Closing #100 removes its blocker from #101, #102, #104, and #107. Only #101 becomes fully
unblocked; it becomes `Ready / Backlog`, not `Next`, until explicitly selected. #102
remains blocked by #101, #104 retains #76, #102, and #103, #107 retains the remaining
decision-map blockers, and #108 remains blocked by #101. Parent #83 remains
`In progress / Needs refinement`.
