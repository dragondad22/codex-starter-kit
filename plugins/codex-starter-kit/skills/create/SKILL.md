---
name: starter-kit-create
description: Guide approved creation of a Codex Starter Kit managed repository. Use when a user asks to initialize, create, bootstrap, or set up a new repository under Starter Kit management.
---

# Starter Kit Create

Guide managed-repository creation through the standalone lifecycle engine. Keep inspection,
planning, human review, effect approval, application, and result reporting separate.

Do not invent a brief, persona confirmation, special-data answer, notice acknowledgment,
engine/baseline provenance, plan approval, or repository authority.

## Routing boundary

Use this skill for an explicit `$starter-kit-create` request or a focused request to
initialize, bootstrap, or create a new Starter Kit managed repository. Do not route generic
requests to create files, events, issues, applications, or content. Do not use it to
retrofit, repair, reconcile, verify, or upgrade an existing repository.

## Capability boundary

1. Establish the requested repository as an absolute native path and the requested
   operation as `create`.
2. Resolve the engine without installing, replacing, or updating it. Invoke executable and
   argument vector equivalent to `[engine, "capabilities"]`, never a composed shell string.
3. Apply [references/create-contract.md](references/create-contract.md). Full create
   requires a verified compatible engine, the approved baseline identity/material,
   `inspect`, `plan`, `apply`, and `status`, authorized local process/filesystem access,
   and the ability to retain the exact plan safely. A filename, cache hit, self-reported
   identity, or plugin version is not proof.
   Validate the bundled
   [baseline manifest](../../baselines/professional-v1/baseline.json) and its content digest
   before using its offline professional-baseline projection. External evidence must still
   verify the containing plugin snapshot.
4. If a prerequisite is missing, incompatible, unverified, unavailable offline, or denied,
   do not inspect, plan, or apply. Report the truthful capability mode, exact failed fact,
   remediation, and reviewed direct-engine fallback. Do not download or enable anything.

## Human-owned inputs

Gather only these inputs before planning:

- the human-supplied project brief;
- explicit confirmation that this exact brief is approved;
- explicit confirmation of the seed project-owner persona; and
- the special-data-handling declaration `No`, `Yes`, or `Unsure`.

Never draft missing authority and then mark it approved. For `Yes` or `Unsure`, present the
concise notice in the contract and require explicit acknowledgment. Acknowledgment permits
only safe workflow continuation; it is not handling authorization, route assurance,
classification, legal review, product assurance, or permission to expose the content.
Do not ask the user to provide specially handled content.

The current create-v1 engine request does not persist the declaration. Identify it as a
session-scoped workflow fact and a coverage limitation; never claim it was written to the
managed repository.

## Inspect and plan

1. Invoke executable and argument vector equivalent to
   `[engine, "inspect", "--repository", absolute_repository]`. Present unmanaged/managed,
   Git, user-content, contract, problem, and precondition facts without mutation.
2. Stop on existing/user content or a managed contract unless the engine returns an exact
   supported no-change create path. Create never supplies retrofit or reconciliation
   authority.
3. Invoke executable and argument vector equivalent to
   `[engine, "plan", "--operation", "create", "--repository", absolute_repository,
   "--brief", approved_brief, "--approve-brief", "--confirm-owner-persona"]`.
4. Validate the complete plan envelope from the contract. Retain the exact JSON bytes in a
   private temporary file outside the target repository and retain `plan_id` separately.
   If the host cannot do that safely and exactly, stop before effects.
5. Show the plan ID and precondition, every proposed path with ownership/source/digest,
   whether it is `no_change`, the result-evidence destination, and these current seed
   limits: policy is `not_configured`, initial verification has not run, no control pass is
   implied, special-data declaration is not persisted, and multi-file mutation uses
   recovery/compensation rather than crash-atomic commit.

Do not hide planned file content from a user who asks to review it, but do not echo the
brief or other content unnecessarily. Never present plan generation as effect approval.

## Apply boundary

Ask the user to approve application of the exact retained plan ID after the review. A
general request to “set it up,” brief approval, persona confirmation, or notice
acknowledgment is not apply approval. If approval is absent, ambiguous, or declined, stop
with the plan retained and no repository effect.

After explicit approval, invoke executable and argument vector equivalent to
`[engine, "apply", "--plan", private_plan_path, "--plan-id", retained_plan_id]`.
Do not regenerate, edit, or substitute the plan. Do not retry a changed precondition with a
new plan without a new review and approval.

## Result

Validate the result or structured failure envelope and preserve all status, changed-file,
conflict, recovery, evidence, stage, cause, and recoverability facts. Keep `applied`,
`no_change`, `failed`, and reconciliation-required outcomes distinct.

- `applied` reports only the exact changed paths and evidence returned.
- `no_change` is an engine result for an already valid unchanged contract, not a guessed
  conversational success.
- stale preconditions and new/existing content stop; preserve content and prepare a new
  plan only after explicit review.
- interrupted setup may replay only the same immutable plan when the engine says so.
- rollback failure is non-recoverable and requires preserved evidence and human
  reconciliation; never report success.

Do not run verification automatically. Offer the focused verify workflow only when it is
implemented and separately requested.
