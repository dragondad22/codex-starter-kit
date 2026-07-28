# Assurance-Factory Seam Inventory

**Scope:** Design-first feature
[#83](https://github.com/dragondad22/codex-starter-kit/issues/83)

**Snapshot:** `origin/main` at `1b66fe58e06324fd4b1d09b60d8afbd53baf19bf`,
2026-07-27

**Status:** Planning inventory only, not a completed research record and not product,
architecture, policy, or execution authority

## Purpose

This inventory identifies what the repository already owns, what is still being
qualified, and which decisions are genuinely missing before an assurance-factory
capability can be decomposed into executable work. It prevents #83 from creating a
second lifecycle, evidence, issue, or authority model merely because a future
multi-actor workflow needs to consume those facts.

## Stable seams to consume

| Existing seam | Current owner and behavior | Constraint on #83 |
|---|---|---|
| Product and human authority | PRD, personas, accepted decisions, specifications, policy, risks, and approved issue contracts | Generated artifacts may derive and route authority; they may not silently replace it. |
| Lifecycle operations | The lifecycle engine owns `create`, `retrofit`, `inspect`, `plan`, `apply`, `verify`, `status`, and `upgrade` semantics | A factory cannot introduce a competing public lifecycle or let an actor bypass engine plans and evidence. |
| Professional quality | DEC-0017 supplies one applicability-aware baseline with explicit non-pass states | Autonomy, actor count, model choice, or low interaction cannot create a lower-quality passing path. |
| Operating profile | DEC-0019 separates engagement, additive assurance, and evidence presentation | Actor coordination may change cadence and inputs, not applicable controls or pass criteria. |
| Distinct evaluation | DEC-0020 requires a capable review outside the implementation context | Specialist evaluators must extend declared capability and separation; checks or role names do not prove either. |
| Effect authority | DEC-0022 binds external work, retries, recovery, and cleanup to a scoped execution mandate | Work packages, mission state, or dashboards cannot self-authorize an effect. |
| Governed work | Work Manager owns executable issue contracts, freshness, desired issue/Project state, native relationships, plans, effect receipts, replay, and bounded reconciliation | Later mission state must reference or extend these identities without duplicating readiness, issue, Project, or receipt authority. |
| Policy and controls | Policy Compiler and Control Evaluator own applicability and `pass`, `fail`, `not-applicable`, `not-configured`, `needs-review`, and accepted-exception semantics | Evaluator errors, missing capability, disagreement, and absent evidence cannot collapse into pass. |
| Evidence and projections | Evidence Store owns attributable evidence; DEC-0004 separates structured state, generated views, and human-owned records | Validation contracts and Mission Control must be reproducible projections with source identity, freshness, limitations, and conflict behavior. |
| Conversational capture | Issue #84 and the issue-tracker contract govern search, containment, promotion, and natural capture checkpoints | Candidate findings require a distinct disposition step before they become defects, decisions, or new Project items. |

## In-progress dependencies

- [#75](https://github.com/dragondad22/codex-starter-kit/issues/75) is
  `In progress / Ready` and is proving one Ready issue through branch, pull request,
  exact-head checks, distinct review, squash merge, completion memory, and Project
  reconciliation.
- [#76](https://github.com/dragondad22/codex-starter-kit/issues/76) is
  `Backlog / Blocked` and owns aggregate qualification of the GitHub executable-work
  contract.
- Validation-contract authority can be resolved before #76 because it strengthens the
  existing single-writer path without assuming orchestration. A new mission boundary
  cannot be approved until #76 establishes the completed delivery seam it would extend.

## Missing contracts

The repository has no approved answer for:

1. validation-contract authority, ownership, identity, lifecycle, invalidation, and
   evidence linkage;
2. assertion classes and the evidence methods capable of evaluating them;
3. actor roles, capability declarations, least-knowledge inputs, tool/data authority,
   isolation, results, and limitations;
4. candidate-finding reproduction, disposition, correction, promotion, disagreement,
   and stopping rules;
5. durable mission identity, bounded actor episodes, work-package handoff,
   orchestration, replay, and recovery relative to the lifecycle engine and Work
   Manager;
6. provider-neutral work-package and result semantics, capability parity, privacy,
   cost, compatibility, and fallback behavior; or
7. Mission Control projection freshness, truthful non-pass views, decision queues, and
   action authority.

No implementation module, prototype, provider adapter, evaluator kit, orchestration
runtime, or dashboard should be inferred from these gaps.

## Progress since snapshot

- #100 resolved item 1 and promoted the result as
  [DEC-0023](../decisions/DEC-0023-validation-manifest-authority-and-lifecycle.md).
- #101 completed bounded
  [assertion/evidence-method research](VALIDATION_ASSERTION_EVIDENCE_METHODS.md) for item
  2. Promotion and executable design remain downstream work.

## Refinement conclusion

The smallest useful frontier is validation-contract authority and lifecycle. It can
improve predetermined outcome traceability for the current single-writer delivery path
without approving multi-actor architecture. The completed research supplies a ten-class
mapping and multi-axis method model without approving a schema or architecture. Actor,
finding, mission, provider, and Mission Control contracts remain separate downstream
decisions because each has distinct authority, dependency, evidence, and handoff value.

The active sequence and GitHub work-item identities are preserved in the
[Assurance Factory Decision Map](../roadmap/ASSURANCE_FACTORY_DECISION_MAP.md). The live
Project and native issue relationships remain lifecycle and execution authority.
