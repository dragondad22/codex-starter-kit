# Issue #74 — governed executable-work evidence

**Date:** 2026-07-21

**Issue:** [#74](https://github.com/dragondad22/codex-starter-kit/issues/74)

**Parent:** [#4](https://github.com/dragondad22/codex-starter-kit/issues/4)

## Delivered deterministic contract

- Added schema-v2 executable issue parsing/rendering and exact governed-source bindings
  behind the existing Work Manager lifecycle seam. Schema-v1 evidence remains readable.
- Added one content-addressed managed-work qualification with six explicit freshness
  dispositions. It is provenance, not DEC-0022 effect authority.
- Added narrow, exact stale-context-digest delegated Current-context repair; other visible
  semantic or human-owned changes return to refinement without mutation.
- Added question/research issue forms, visible subtype round-trip, output digest binding,
  reciprocal promoted-record validation, and an exact re-observed question closing-comment
  backlink.
- Added feature Horizon assignment and parent-derived child context independently of
  Status, Readiness, Phase, saved views, and finite Milestone.
- Added exact related-PR delivery claims. Only a merged same-repository PR bound to managed
  ID, source revision, executable-contract digest, current default-branch reachability,
  exact PR changed-file manifest, and current implemented-file bytes can yield
  `already-delivered`; historical revisions are ignored.
- Added a content-addressed DEC-0022 Work Manager execution mandate for every prospective
  external effect, including schema-v1 compatibility requests. It binds exact input,
  governed-source, full governance/refresh authority, selected operation/root item,
  desired-resource, actor, permission, ceiling, and cumulative-effect limits. Memory
  effects and effect-free inspection need no external authority.

## Deterministic validation matrix

The engine and native HTTP adapter cases cover canonical contract and subtype round-trip,
missing schema, unexpected preamble, edited acceptance, unbound governing references,
stale source digest, human-owned context conflict, delegated context refresh, unchanged
fresh work, mechanical Project drift, open native blockers, exact already-delivered work,
partial delivery, plan/effect qualification binding, missing external mandate, contained
mandate receipts, native parent-derived Horizon, copied child Horizon clearing, Ready
feature Horizon requirements, optional capability reporting, ambiguous type labels, and
current default-branch delivery observation including wrong-base, unreachable, removed,
changed-content, omitted-file, empty-PR, and historical-revision negatives. Mandate cases
cover exact governance/refresh authority, desired resources, operation/root identity,
cumulative usage across work switches, and missing/corrupt state.

Relationship coverage composes the existing #15 parent/dependent reconciliation through
the same plan/apply/verify/status seam. Subtype completion composes DEC-0013/#16, and Phase
projection composes the completed #46 catalog without requiring a saved view.

## Current verification state

The working-tree development candidate passed locally:

```text
go test ./...
  all packages passed
python3 -m unittest discover -s tests -p "test_*.py"
  37 tests passed
python3 scripts/validate_docs.py
  Documentation validation passed
go test -race ./engine ./githubadapter ./cli
  all selected packages passed
go vet ./...
  passed
starter-kit changes check --repository .
  passed
git diff --check
  passed
```

This is development evidence, not completion evidence. Exact-head review and
Linux/macOS/Windows CI remain completion gates after the live receipt is recorded.

## Approved live sandbox observation

Read-only qualification run
[`29861694892`](https://github.com/codex-starter-kit-labs/codex-starter-kit-sandbox/actions/runs/29861694892)
passed against Starter Kit commit `8c6b9dc1842f3f226a5422a422fa8cd4204ff4e2`.
The protected `contract-reconciler` environment minted an installation token for
`codex-starter-kit-labs-reconciler` installation `147093185`; the token response and
installation metadata agreed on permission revision
`972f4aa57f95fbf11bec751cf0a827be14e9f859ee5163c426552d81d44dc0e1` and included
`contents:read` plus the previously qualified reconciler permissions.

The App token read `README.md` from the isolated public sandbox at immutable commit
`e73da2428eb61242b6c128b6d69b61c77d2e5fc5`. Its decoded bytes matched approved digest
`sha256:5e24803c1e7aa208dc29c7f23d9f4d4c5559b20d7b32ad0e168614dd17a385ef`.
The redacted schema-v1 receipt has `evidence_mode: live` and `outcome: pass`; the workflow
artifact is retained for 30 days. The temporary qualification branch was removed after
the run. The versioned driver rejects mutable revisions, unsafe paths, missing
`contents:read`, non-file responses, and digest mismatches, and never serializes the App
key or installation token.

## Explicit limitations and handoff

- Deterministic full-journey GitHub transport receipts remain simulated. The approved
  read-only #74 source observation is live; no live #74 mutation is claimed.
- General GitHub setup and team-specific Project/view choices remain future optional
  GitHub App work. Saved-view presence or layout is not a #74 conformance gate.
- Issue #75 owns branch, PR, checks, distinct review, and merge delivery. It emits the
  delivery claim #74 now observes and uses an existing App installation token as logical
  `merger`; no new human account is required.
- Issue #76 owns aggregate live qualification and final support claims.

## Reconciliation state

The App permission and approved live evidence boundary are complete. Issue #74 remains
`In progress` until exact-head review and native CI complete. Parent #4 remains
`In progress`; #75 and #76 remain blocked by their native dependencies. Issue #95 is the
selected next item after #74, before #75.
