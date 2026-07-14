---
name: starter-kit-verify
description: Guide truthful Codex Starter Kit repository verification. Use when a user asks to verify controls, conformance, evidence, coverage, or a lifecycle gate for a managed Starter Kit repository.
---

# Starter Kit Verify

Guide immutable verification planning and evidence regeneration through the standalone
Starter Kit engine. Preserve every explicit control state, aggregate state, limitation,
diagnostic, evidence reference, actor, and authority fact.

Do not invent scope, gate, actor, authority, plan approval, evaluator results, risk
acceptance, or evidence.

## Routing boundary

Use this skill for explicit `$starter-kit-verify` requests and focused requests to verify
Starter Kit controls, conformance, coverage, evidence, or a lifecycle gate. Do not route
generic identity/email/data verification, test-only requests, or questions that ask only
for repository lifecycle status. Do not use conversational judgment as a control evaluator.

## Capability boundary

1. Establish an absolute native repository path and the requested operation `verify`.
2. Resolve the engine without installing/updating it. Invoke executable and argument vector
   equivalent to `[engine, "capabilities"]`, not a composed shell string.
3. Apply [references/verify-contract.md](references/verify-contract.md). Verification
   requires a verified compatible engine with `verify-plan`, `verify`, `inspect`, and
   `status`, supported schema/protocol, a managed-repository contract, authorized local
   reads and evidence writes, exact private plan retention, and applicable baseline/policy
   compatibility facts.
4. `verification-only` permits this workflow when create/apply mutation is unavailable but
   verification evidence regeneration is explicitly authorized. It is not permission to
   create, repair, migrate, or upgrade the repository.
5. On missing, incompatible, disabled, unverified, denied, offline-unavailable, or
   conflicting prerequisites, stop with the truthful mode, failed facts, remediation, and
   direct engine/CI fallback. Do not install, enable, transmit, or broaden authority.

## Human-owned verification request

Gather the exact non-empty values required by the engine:

- verification `scope`;
- lifecycle `gate`;
- requesting `actor`; and
- `authority` for regenerating repository evidence.

Do not infer actor or authority from the user's account, role, original request, repository
ownership, risk record, or prior approval. Do not put secrets, specially handled content,
credentials, or private evidence in these metadata fields.

## Plan and review

1. Invoke executable and argument vector equivalent to
   `[engine, "verify-plan", "--repository", absolute_repository, "--scope", scope,
   "--gate", gate, "--actor", actor, "--authority", authority]`.
2. Validate the complete immutable plan envelope. Retain the exact JSON bytes in a private
   temporary file outside the repository and retain `plan_id` separately. Stop if native
   exact/private retention is unavailable.
3. Present repository and precondition digests, plan ID, scope, gate, actor, and authority.
   Explain that verify can write machine evidence and an operation event, regenerate
   `docs/evidence/CONFORMANCE.md`, update its managed digest, and roll back ordinary commit
   failures. It does not create a pass, accept risk, repair controls, install a scanner, or
   prove packaged-engine provenance.
4. Ask for explicit approval to execute this exact plan and regenerate evidence. Preparing
   the plan or providing metadata is not execution approval.

If the user declines or approval is ambiguous, stop before `verify` with the plan retained
and no verification effect.

## Execute

After exact-plan approval, invoke executable and argument vector equivalent to
`[engine, "verify", "--plan", private_plan_path, "--plan-id", retained_plan_id]`.
Do not edit/regenerate/substitute the plan or retry a stale plan. Repository changes after
planning require a new plan, review, and approval.

## Present truthful evidence

Validate the complete result using the contract, then preserve:

- verification/evidence identities, ownership, and source;
- scope, gate, actor, and authority;
- source revision/snapshot and engine/repository/policy versions;
- verification time and exact aggregate state;
- every control ID, state, underlying state, summary, rationale, evidence reference, and
  redacted diagnostic;
- every coverage limitation; and
- machine evidence and event paths.

Keep `pass`, `fail`, `not-applicable`, `not-configured`, `needs-review`, and
`accepted-exception` distinct. An accepted exception must retain its underlying non-pass
state and risk evidence; it is not pass. Never display aggregate pass when any control is
non-pass, required evidence is absent, coverage is incomplete, or an evaluator failed.

Do not hide, rewrite, or embellish diagnostics. Do not open or transmit referenced machine
evidence unless separately requested, authorized, and safe. Malformed/conflicting output
is `unsupported` with lifecycle/conformance unknown, never a partial pass.

On execution failure, preserve the engine's redacted diagnostic and any returned or
discoverable authorized Git-local attempt evidence. Stale plans need a new plan; evaluator
failure remains fail/non-pass; plugin unavailability uses direct engine/CI fallback.
