# Guided verification contract

This contract applies DEC-0018 and the lifecycle engine's verify-v1 interface. Workflow
mode, repository lifecycle, verification aggregate, control state, and risk acceptance are
independent facts.

## Compatibility and authority

Verification requires capability schema `1`, engine `starter-kit`, protocol
`starter-kit.lifecycle` version `1`, operations `verify-plan`, `verify`, `inspect`, and
`status`, and verification/result schema version `1`. Retained external evidence must
verify the resolved engine and applicable baseline/policy compatibility facts.

The repository must be safely inspectable and the current user/workspace/admin/sandbox
policy must authorize process execution, repository reads, lifecycle locking, evidence and
event writes, conformance projection replacement, managed-digest update, and ordinary
rollback. Network access and repository content transmission are not required or implied.

`verification-only` allows those bounded local verification effects while prohibiting
create/apply/migration effects. Unknown or denied evidence-write authority is
`unsupported`, not read-only verification.

## Immutable plan

A verify plan is valid only when it is a JSON object with:

- `schema_version: 1`;
- a SHA-256 `plan_id` retained separately;
- a non-empty canonical `repository` and SHA-256 `repository_digest`; and
- the exact non-empty `scope`, `gate`, `actor`, and `authority` supplied by the user.

The plan contains no execution result. Retain exact bytes privately outside the repository.
Approval binds this exact plan and its precondition. A changed repository, altered metadata,
or regenerated plan requires new review and approval.

## Verification result

Accept only a JSON object containing:

- schema version `1`, SHA-256 verification/evidence identities, ownership
  `machine-evidence`, and source `engine:verify:v1`;
- exact scope/gate/actor/authority;
- non-empty source revision/snapshot, engine/repository/policy versions, and verification
  time;
- `overall_state` in the six allowed states;
- a non-empty `controls` array;
- string-array `coverage_limitations`; and
- non-empty local `evidence_path` and `event_path`.

Each control requires a stable ID; exactly one allowed state; summary; optional rationale;
structured evidence references; and redacted string diagnostics. Allowed states are:

- `pass`;
- `fail`;
- `not-applicable`;
- `not-configured`;
- `needs-review`; and
- `accepted-exception`.

Only `accepted-exception` carries `underlying_state`, which must remain an explicit non-pass
state. Preserve it and its risk evidence. Never relabel an exception as pass.

Cross-check fail closed: aggregate `pass` is possible only when every evaluated control
passes, required current evidence exists, coverage is complete for the stated scope/gate,
and no evaluator failed. For non-pass aggregates, preserve the engine's priority/meaning;
do not conversationally recompute a greener state. Empty controls or missing evidence
cannot pass. Malformed, partial, contradictory, unredacted-sensitive, or unknown output is
`unsupported` and no conformance result may be claimed.

## Effects and failure

Execution may create `.starter-kit/evidence/verify-*.json` and
`.starter-kit/events/verify-*.json`, regenerate `docs/evidence/CONFORMANCE.md`, and update
the generated file's managed digest. It may use Git-local `starter-kit-attempts` evidence
when repository evidence cannot safely commit. These are local mutation effects requiring
the exact-plan approval.

A stale precondition, lock failure, evaluator failure, persistence failure, rollback
failure, or malformed output stays explicit. Do not repair, retry with new inputs, accept
risk, install a tool, or suppress a diagnostic without separate authority. Diagnostics
remain engine-redacted; never ask for the concealed content.

## Direct and CI fallback

After prerequisites and authority are established, the reviewed direct sequence is:

```text
starter-kit capabilities
starter-kit verify-plan --repository <absolute-path> --scope <scope> --gate <gate> --actor <actor> --authority <authority>
starter-kit verify --plan <exact-private-plan.json> --plan-id <retained-plan-id>
```

CI may call the same engine seam directly and must not depend on plugin routing or
conversation. These commands are explanatory fallback, not permission to run while a
prerequisite or authority is missing/unverified/denied. Plugin update never regenerates
evidence, changes a repository, updates policy, or accepts risk.
