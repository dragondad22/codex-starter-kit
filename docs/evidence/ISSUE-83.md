# Issue #83 — Assurance-factory refinement

**Date:** 2026-07-27

**Issue:** [#83](https://github.com/dragondad22/codex-starter-kit/issues/83)

## Owner selection and boundary

The product owner deliberately selected #83 for refinement and approved the validation
contract as the first investigation frontier. Evaluator contracts, finding disposition,
mission orchestration, provider interoperability, and Mission Control remain downstream
investigation branches, not approved requirements or architecture.

This refinement does not authorize implementation, a prototype, a plugin, a new lifecycle
module, an evaluator kit, a provider adapter, a dashboard, parallel mutation, credentials,
installation, paid services, sensitive-data transmission, or other external product
effects.

## Published investigation decomposition

The bounded
[seam inventory](../research/ASSURANCE_FACTORY_SEAM_INVENTORY.md) identifies current
authority and implementation ownership. The
[decision map](../roadmap/ASSURANCE_FACTORY_DECISION_MAP.md) sequences nine native
children:

| Issue | Outcome | Native blockers at publication |
|---|---|---|
| [#100](https://github.com/dragondad22/codex-starter-kit/issues/100) | Define validation-contract authority and lifecycle | None |
| [#101](https://github.com/dragondad22/codex-starter-kit/issues/101) | Map assertion classes to evidence methods | #100 |
| [#108](https://github.com/dragondad22/codex-starter-kit/issues/108) | Decide the validation-contract executable slice | #100, #101 |
| [#102](https://github.com/dragondad22/codex-starter-kit/issues/102) | Define actor capability and least-knowledge contracts | #100, #101 |
| [#103](https://github.com/dragondad22/codex-starter-kit/issues/103) | Define candidate-finding disposition and arbitration | #101, #102 |
| [#104](https://github.com/dragondad22/codex-starter-kit/issues/104) | Define mission orchestration and recovery boundaries | #76, #100, #102, #103 |
| [#105](https://github.com/dragondad22/codex-starter-kit/issues/105) | Research provider-neutral work-package interoperability | #102, #104 |
| [#106](https://github.com/dragondad22/codex-starter-kit/issues/106) | Define Mission Control projection and action authority | #93, #103, #104, #105 |
| [#107](https://github.com/dragondad22/codex-starter-kit/issues/107) | Decide the broader executable vertical-slice decomposition | #100–#106, #108 |

#93 remains the authority route for outcome-based efficiency measurement and #94 remains
the onboarding feature. #105 and #106 cross-link those issues without duplicating their
scope.

## Project reconciliation

- #83 is `In progress / Needs refinement / Horizon Next`.
- #100 is `Status Next / Readiness Ready` and is the selected investigation.
- #101–#108 are `Status Backlog / Readiness Blocked`.
- Every new issue is a native child of #83, is enrolled in Project #8, and has explicit
  native dependency relationships matching the table.
- Children carry no direct Horizon or Phase assignment.

The parent remains `Needs refinement` because its product and architecture questions are
not resolved. A Ready question authorizes only its bounded resolution and promotion; it
does not authorize product implementation.

## Verification and limitations

The refinement package passed:

- 41 Python documentation tests;
- `python3 scripts/validate_docs.py`;
- `go test ./...`;
- structured change-record validation; and
- `git diff --check origin/main...HEAD`.

The live GitHub audit verified issue bodies, native hierarchy, dependencies, labels, and
Project fields. Reciprocal scope links were added from #93 to #106 and from #94 to #105.

These records are a dated refinement snapshot. Current GitHub state remains operational
authority, and every ticket repeats freshness review when selected.

## Validation assertion research update

Issue #101 was explicitly selected on 2026-07-28 after #100 promoted DEC-0023. Its
[bounded research record](../research/VALIDATION_ASSERTION_EVIDENCE_METHODS.md) maps ten
representative assertion classes to a multi-axis evidence-method profile while preserving
authority, capability, coverage, non-pass, and reuse constraints.

The result does not establish product or architecture authority. Once #101 completes,
#102 and #108 become `Status Backlog / Readiness Ready`; #103 remains blocked by #102;
and #83 remains `In progress / Needs refinement`.
