# Issue #101 — Validation assertion/evidence research

**Date:** 2026-07-28

**Issue:** [#101](https://github.com/dragondad22/codex-starter-kit/issues/101)

**Parent:** [#83](https://github.com/dragondad22/codex-starter-kit/issues/83)

## Research outcome

The bounded
[research record](../research/VALIDATION_ASSERTION_EVIDENCE_METHODS.md) classifies ten
representative assertion classes and identifies a multi-axis method profile. Evidence
action, observation boundary, execution mode, capability/separation, and coverage remain
distinct so an automated black-box specialist evaluation can be represented truthfully
without collapsing unlike facts into one method label.

The result preserves DEC-0023: authoritative sources determine assertions, the immutable
manifest records acceptable evidence constraints, and downstream plans and receipts
record selected evaluator implementations, attempts, and results. The research does not
approve a schema, module, evaluator kit, provider, quality score, or product architecture.

## Coverage and negative paths

- Mapped functional acceptance; static structure; lifecycle and authority; negative paths
  and recovery; security/privacy; accessibility; persona/UX; architecture/maintainability;
  documentation/support; and process/provenance assertions.
- Preserved `fail`, `not-applicable`, `not-configured`, `needs-review`, `unsupported`,
  invalid-input, and accepted-exception semantics without manufacturing pass.
- Distinguished objective violations, supported findings, subjective recommendations,
  unresolved intent, and human-owned risk acceptance.
- Required exact source, scope, evaluator capability, separation, coverage, environment,
  freshness, evidence route, limitation, and reuse traceability.
- Identified the open schema boundary between `unsupported` capability/manifest state and
  the Control Evaluator's current assertion result vocabulary for #108 to preserve rather
  than guess.

## Source and authority

Repository sources were inspected at exact `origin/main` revision `98f7cfe`. External
primary sources were retrieved on 2026-07-28 and versioned in the research record. The
external sources clarify method semantics only; they do not become product authority.

Material conclusions remain research until an applicable owner-approved decision promotes
them. #108 may use this record to decide whether the validation-only path is ready for
executable decomposition, while #102 and #103 retain actor and finding-disposition
authority.

## Verification evidence

The initial documentation package passed:

```text
python3 -m unittest discover -s tests -p "test_*.py" — 41 tests passed
python3 scripts/validate_docs.py — passed
go test ./... — passed
starter-kit changes check --repository . — passed; generated changelog current
git diff --check — passed
```

The completing pull request retains the exact final source, distinct review, and native CI
evidence.

## Downstream reconciliation

Completing #101 removes the last blocker from #102 and #108, which become Readiness
`Ready` while remaining Status `Backlog` unless deliberately selected. #103 remains
blocked by #102. Parent #83 remains `In progress / Needs refinement` because later
decision-map questions and delivery remain unresolved.
