# Guided create contract

This contract implements the create boundary from DEC-0018 and the lifecycle engine's
create-v1 interface. Workflow capability, repository lifecycle, and conformance are
separate facts.

## Compatibility facts

Full create requires all of the following:

- installed/enabled plugin and focused create skill are permitted by the host/workspace;
- capability schema `1`, engine name `starter-kit`, protocol
  `starter-kit.lifecycle` version `1`;
- `inspect`, `plan`, `apply`, and `status` operations are present;
- the plan/apply/result schema version is `1`;
- retained external evidence verifies the resolved engine artifact;
- the approved professional baseline identity, digest, compatibility metadata, and
  required local material are available and verified, including offline when offline use
  is claimed;
- read and mutation access to the target are authorized, and the current sandbox/approval
  policy permits each requested process/filesystem effect; and
- exact private plan retention and structured executable/argument invocation are
  available natively.

The plugin bundles `baselines/professional-v1/baseline.json` and its digest-bound
`professional-engineering.md` projection with identity
`baseline:professional-engineering:v1`, version `1.0.0`, and source decision `DEC-0017`.
The projection is locally available offline but is not a signed policy pack or conformance
evidence. Full use requires the manifest digest to match and external qualification to
verify the containing plugin snapshot.

The current repository publishes no verified packaged engine or signed baseline policy
pack. Ordinary source checkout/PATH use therefore remains `degraded-guidance`. The plugin
may explain and provide the direct commands for later review, but it must not execute
create or claim supported offline first-run operation.

## Special-data notice

For `Yes` or `Unsure`, say concisely:

> Your declaration triggered this notice. The current Codex, tool, service, and
> environment route is not assumed verified for specially handled content. Do not supply
> or transmit that content until handling authorization and route assurance are
> established. We can continue with metadata-only planning and remediation.
> Acknowledgment records only that you received this notice; it grants no handling
> authorization or product assurance.

The `No` path does not trigger a detailed privacy interview. Any contradiction discovered
later invalidates affected applicability/evidence. The create-v1 engine does not accept or
persist this declaration; disclose that coverage limit.

## Inspection and plan envelopes

Inspection must be a JSON object with schema version `1`, a non-empty canonical repository,
booleans for Git/managed/contract presence, string arrays for problems, non-negative user
file/directory counts, and non-empty snapshot/precondition digests.

A create plan must be a JSON object with:

- `schema_version: 1`, `operation: "create"`, canonical `repository`, and a SHA-256
  `repository_digest`;
- `plan_id` equal to the engine-supplied SHA-256 identity retained separately;
- `files` as structured path/ownership/source/digest/content objects;
- `no_change` as a boolean;
- `approval.brief_digest`, `approval.brief_approved: true`, and
  `approval.owner_persona_confirmed: true`; and
- a structured machine-evidence result path/ownership/source.

Reject malformed, partial, unsupported, or conflicting output. Never salvage an apply
authority from visible file content or an apparent plan ID.

## Effect review

Before apply, show at least:

- repository and immutable plan/precondition identities;
- all paths, ownership classes, provenance sources, and content digests;
- whether each path is human-owned, generated, managed, or machine evidence;
- `no_change` versus proposed writes;
- conflicts and user-owned content preservation;
- policy/control/baseline and special-data coverage limitations;
- recovery/replay behavior and the absence of crash-atomic/external-effect claims; and
- exact local effect authority requested.

Apply approval must bind the exact retained plan. No conversation acknowledgment can
authorize a changed or regenerated plan.

## Result envelopes

Successful apply output has schema version `1`, the matching plan ID, status `applied` or
`no_change`, and string arrays for `changed_files`, `recovery`, and `evidence`.

Apply failure output has schema version `1`, a structured result, `error`, `recoverable`,
and, when available, a structured failure with `stage`, `recoverable`, `changed_files`,
`conflicts`, `recovery`, `evidence`, and `cause`. Reconciliation during planning is a
schema-versioned object with repository, conflicts, problems, and recovery arrays.

Preserve the exact engine meaning. Empty evidence is not a pass; `recoverable: true` is
not permission for destructive repair; `recoverable: false` is never silently retried.

## Direct fallback

After prerequisites are independently satisfied and the user chooses direct operation,
the reviewed sequence is:

```text
starter-kit capabilities
starter-kit inspect --repository <absolute-path>
starter-kit plan --operation create --repository <absolute-path> --brief <approved-text> --approve-brief --confirm-owner-persona
starter-kit apply --plan <exact-private-plan.json> --plan-id <retained-plan-id>
starter-kit status --repository <absolute-path>
```

These are explanatory commands, not permission to execute them while an engine/baseline is
missing, incompatible, unverified, denied, or unavailable offline. Plugin update never
updates the engine, baseline, repository, or plan.
