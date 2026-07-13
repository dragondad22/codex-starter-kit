# Lifecycle Engine — Phase 1 Create Interface

**Status:** Implemented development slice
**Decision:** [DEC-0015](../decisions/DEC-0015-lifecycle-engine-toolchain.md)
**Issue:** [#26](https://github.com/dragondad22/codex-starter-kit/issues/26)

## Interface

The `starter-kit` CLI and Go callers cross the same lifecycle-engine seam. The CLI emits
JSON on standard output and diagnostics on standard error:

```text
starter-kit inspect --repository <path>
starter-kit create --repository <path>
starter-kit plan --operation create --repository <path>
starter-kit apply --plan <plan.json> --plan-id <sha256:...>
starter-kit status --repository <path>
```

`create` is the focused convenience operation for a create plan. `plan --operation
create` produces the same immutable result for unchanged inputs. The caller reviews and
stores that JSON plan, then supplies both the plan document and its separately retained
identifier to `apply`. Apply re-hashes the plan, re-inspects repository preconditions,
refuses existing targets, writes state last, verifies every content digest, and returns a
structured result. Repeating create on the unchanged managed repository returns an
explicit `no_change` plan and result.

The current seam implements `create`, `inspect`, `plan`, `apply`, and `status`. `verify`
is #27; `retrofit` and `upgrade` remain later phases. A missing operation must not be
represented as available.

## Versioned JSON contracts

Every document/result includes `schema_version: 1`. Plan identity is the SHA-256 digest
of its canonical Go-encoded JSON with an empty `plan_id`; file digests are SHA-256. Plans
contain the repository precondition digest, proposed paths, ownership, content, and
content digest. Go types are not durable authority: compatibility is defined by observable
JSON fields and black-box behavior through the engine seam.

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
| `AGENTS.md` | generated | Concise repository orientation and routes |
| `docs/product/BRIEF.md` | human-owned | Approved seed project brief |
| `docs/product/PERSONAS.md` | human-owned | Confirmed seed persona registry |
| `docs/decisions/INDEX.md` | human-owned | Durable decision index |
| `docs/evidence/CONFORMANCE.md` | generated | Truthful initial not-yet-verified summary |

## Current limits

- Create accepts only an empty Git working tree apart from `.git`; retrofit is deferred.
- Phase 1 uses the Go standard library and the structured `git` executable/argument seam.
- Transaction staging, locks, rollback, recovery evidence, and deeper stale-plan conflict
  handling are #29.
- Hostile path, symlink/junction, reserved-name, case-collision, and malicious-plan
  hardening are #28.
- Seed control evaluation and conformance evidence are #27; the initial summary explicitly
  reports that verification has not run and claims no pass.
- Runtime support is not published until #30 proves native semantic equivalence and exact
  OS/architecture/filesystem assumptions.
