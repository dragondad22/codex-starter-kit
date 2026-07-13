# Lifecycle Engine — Phase 1 Create Interface

**Status:** Implemented development slice
**Decision:** [DEC-0015](../decisions/DEC-0015-lifecycle-engine-toolchain.md)
**Issues:** [#26](https://github.com/dragondad22/codex-starter-kit/issues/26)–[#28](https://github.com/dragondad22/codex-starter-kit/issues/28)

## Interface

The `starter-kit` CLI and Go callers cross the same lifecycle-engine seam. The CLI emits
JSON on standard output and diagnostics on standard error:

```text
starter-kit inspect --repository <path>
starter-kit create --repository <path> --brief <text> --approve-brief --confirm-owner-persona
starter-kit plan --operation create --repository <path> --brief <text> --approve-brief --confirm-owner-persona
starter-kit apply --plan <plan.json> --plan-id <sha256:...>
starter-kit status --repository <path>
starter-kit verify-plan --repository <path> --scope <scope> --gate <gate> --actor <actor> --authority <authority>
starter-kit verify --plan <verify-plan.json> --plan-id <sha256:...>
```

`create` is the focused convenience operation for a create plan. The caller supplies the
brief and separately confirms brief approval and the seed owner persona; omission stops
planning rather than inventing human authority. `plan --operation create` produces the
same immutable result for unchanged inputs. The caller reviews and stores that JSON plan,
then supplies both the plan document and its separately retained identifier to `apply`.
Apply re-hashes the plan, re-inspects content/Git preconditions, constrains every path to
the repository, stages and verifies content, acquires the lifecycle lock, refuses existing
targets, commits state last, validates the complete managed contract, rolls back a failed
commit where possible, and returns a structured result. Repeating create produces
`no_change` only when the manifest, state, and every managed artifact remain valid.

The current seam implements `create`, `inspect`, `plan`, `apply`, `status`, and seed
`verify`. `retrofit` and `upgrade` remain later phases. A missing operation must not be
represented as available.

## Seed verification

`verify-plan` captures an immutable, reviewable repository precondition plus the explicit
scope, lifecycle gate, requesting actor, and authority. `verify` consumes that plan and
its separately retained identifier, rechecks the repository precondition, and records
evidence regeneration in a content-addressed operation event. Each result is exactly
one of `pass`, `fail`, `not-applicable`, `not-configured`,
`needs-review`, or `accepted-exception`; an accepted exception retains its underlying
state. Aggregate `pass` is possible only when every evaluated control passes.

| Control | Seed behavior |
|---|---|
| `CORE-TRUTH-001` | Passes when the result model is explicit and pass states cite current evidence |
| `CORE-SECRETS-001` | `not-configured` until an approved scanner provides defensible coverage |
| `CORE-OWNERSHIP-001` | Passes only for a complete valid managed-file ownership/provenance contract |
| `CORE-COVERAGE-001` | Passes when evaluated controls and coverage limits are disclosed |
| `CORE-RECOVERY-001` | `not-configured` until #29 supplies interruption and stale-lock evidence |
| `CORE-ROUTES-001` | Passes only when stable seed routes parse and resolve |

The engine injects a clock so controlled runs can reproduce timestamps and semantics.
Machine evidence records scope, gate, source revision/snapshot, engine and repository
schema versions, policy state, controls, evidence references, limitations, and redacted
diagnostics under `.starter-kit/evidence/`. Each evidence document carries a digest over
its content with the digest field blank. The human `CONFORMANCE.md` projection and its
managed-file digest are replaced with rollback data under the lifecycle lock. Dynamic
verification evidence is schema/provenance/digest validated as part of the managed
contract.

## Versioned JSON contracts

Every document/result includes `schema_version: 1`. Plan identity is the SHA-256 digest
of its canonical Go-encoded JSON with an empty `plan_id`; file digests are SHA-256. Plans
contain content-based repository and Git precondition digests, proposed paths, ownership,
provenance source, content, content digest, and the approved-input digests/confirmations.
Plans also declare the reserved `.starter-kit/events/` machine-evidence path with
ownership and source. A successful mutation stages its plan ID, operation, status,
repository digest, and changed paths with the other files and commits that event before
authoritative state. Failed accepted applies record the same structured failure evidence;
no-change applies record their evaluation event. This is a state-last recoverable baseline,
not a claim of crash-safe atomicity; #29 owns interruption recovery. Go types are not
durable authority: compatibility is defined by observable JSON fields and black-box
behavior through the engine seam.

Failure-only event directories do not by themselves assert that a managed contract is
present, so correcting the precondition and replanning remains possible. A lock rejection
cannot safely mutate the lock-protected repository surface; it is recorded instead in the
Git-local `starter-kit-attempts` ledger and returned as structured failure JSON.

The operation-acceptance boundary follows validation of the request or immutable plan and
the authorized repository root. Malformed create input, a plan with invalid identity,
schema, approvals, or digest fields, and a repository path whose authority cannot be
established are rejected inputs rather than accepted operations. They return redacted
caller diagnostics and deliberately cause no repository or Git effects: rejected input
cannot supply the authority or evidence destination used to record its own rejection.
After this boundary, validation failures are operation results and emit structured,
redacted evidence. #29 owns the complete failure and recovery transition matrix.

Machine authority is stored under `.starter-kit/`. Human-owned records are seeded under
`docs/` and are never silently replaced. Generated views identify their role through the
managed-file manifest.

| Artifact | Ownership | Purpose |
|---|---|---|
| `.starter-kit/project.json` | managed | Approved/detected seed project facts and lifecycle |
| `.starter-kit/policy-lock.json` | managed | Truthful `not_configured` seed policy state until #27 |
| `.starter-kit/layout.json` | managed | Logical role-to-path mapping |
| `.starter-kit/managed-files.json` | managed | Ownership, provenance digest, and path manifest |
| `.starter-kit/state.json` | managed | Lifecycle, schema, and engine state; written last |
| `.starter-kit/routes.json` | generated | Stable artifact-ID resolution |
| `.starter-kit/events/*.json` | machine-evidence | Self-describing operation results with plan, source, ownership, status, and diagnostics |
| Git-local `starter-kit-attempts/*.json` | machine-evidence | Lock-rejected attempts recorded outside the unavailable repository lock |
| `AGENTS.md` | generated | Concise repository orientation and routes |
| `docs/product/BRIEF.md` | human-owned | Approved seed project brief |
| `docs/product/PERSONAS.md` | human-owned | Confirmed seed persona registry |
| `docs/decisions/INDEX.md` | human-owned | Durable decision index |
| `docs/evidence/CONFORMANCE.md` | generated | Truthful initial not-yet-verified summary |

## Hostile-input safety

Create and apply use one portable path policy before effects. Planned paths must be clean
printable-ASCII relative paths with forward slashes; empty/relative segments, Windows
absolute forms, reserved characters and device names, trailing-dot/space aliases, and
case-fold collisions are rejected on every host. The conservative ASCII boundary avoids
claiming Unicode normalization equivalence before a versioned normalization policy exists.

Create treats existing files and directories as user-owned repository content and refuses
to infer reconciliation authority. Apply accepts only the exact seed create artifact set,
ownership classes, and provenance sources, refuses existing targets, and rejects symlinked
managed artifacts even when linked content has the expected digest. Repository-root and
reserved-directory symlinks or junctions are rejected rather than silently followed;
repository authorization also rejects a symlinked ancestor rather than accepting only the
final path component.

All Git effects use a structured executable plus argument vector. The engine supplies an
allowlisted native process environment, removes inherited Git override variables, disables
interactive prompts and system/global configuration, disables repository-local filesystem
monitor execution and hook discovery, and prevents optional read-command locks. Repository
content is never interpolated into a shell command.

Token-, credential-, and private-key-shaped content is rejected before entering repository
paths, create or verification plans, apply staging, verification metadata, or evidence.
Diagnostics from untrusted plans and managed state are redacted; invalid paths are not
echoed during pre-transaction validation. The seed contract validates layout roles, routes,
lifecycle/engine state, project approvals, policy state, ownership, provenance, and content
digests. Self-consistent but semantically malicious state therefore reports
`managed_degraded` rather than managed.

## Current limits

- Create accepts only an empty Git working tree apart from `.git`; retrofit is deferred.
- Phase 1 uses the Go standard library and the structured `git`
  executable/argument/environment seam.
- Staging, exclusive locking, state-last commit, postcondition validation, and best-effort
  rollback are present. #29 owns interruption fixtures, stale-lock recovery, durable
  recovery evidence, stronger atomicity, and deeper conflict/reconciliation behavior.
- Portable path, ownership, secret, malformed-state, malicious-plan, and structured Git
  defenses are implemented. Symlink fixtures execute where native creation is available;
  Windows CI adds a native directory-junction fixture, and the case-collision fixture records
  the runner's filesystem behavior. #30 owns broader reparse-point capability evidence,
  exact filesystem support claims, and released native support closure.
- Seed verification is implemented, but secrets and recovery remain `not-configured`.
  #28–#30 own hostile-input, interruption/recovery, and native support closure; no current
  aggregate verification result is expected to pass.
- Runtime support is not published until #30 proves native semantic equivalence and exact
  OS/architecture/filesystem assumptions.
