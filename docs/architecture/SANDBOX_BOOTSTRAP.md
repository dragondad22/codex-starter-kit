# GitHub Sandbox Bootstrap

**Status:** Implemented engine contract; live qualification in progress

**Issue:** [#73](https://github.com/dragondad22/codex-starter-kit/issues/73)

## Boundary

The lifecycle engine exposes `InspectSandbox`, `PlanSandbox`, `ApplySandbox`,
`VerifySandbox`, `SandboxStatus`, and the composed `BootstrapSandbox` journey. A strict
versioned manifest binds the source revision, configuration revision, immutable GitHub
owner/repository/Project IDs, approved plan identity, marker prefix, and exact normalized
resources. Desired policy stays in the engine; adapters can observe and apply only the
semantic resource effects selected by the engine.

The approved live target is the public
`codex-starter-kit-labs/codex-starter-kit-sandbox` repository and organization Project
#1. Operational repositories and Projects, personal Projects, classic PATs, private or
paid targets, GHES, webhooks, and repository deletion are outside this contract.

## Safety and recovery

Inspection stops on stale/expired authority, configuration or immutable-target mismatch,
duplicate keys, unsupported kinds, sensitive-looking manifest material, or an
unrecognized resource colliding by kind and name. Apply requires the exact digest-bound
plan ID and recorded human approval, refreshes capability and observation before any
effect, and holds the repository lifecycle lease. It never steals an active lease.

Every attempted effect produces a credential-free receipt. Partial application remains
`non-pass`; a new inspection plans only the remaining semantic delta. Integrity-protected
state under `.starter-kit/sandbox/state.json` supports restart status and replay. Cleanup
accepts only an exact approved marker and removes only that managed key. Unrecognized
human-owned resources are preserved.

Project built-in workflow configuration remains human-owned. The adapter observes and
verifies its postcondition but returns `needs-review` instead of silently substituting an
API, Action, or broader credential when configuration is absent.

## Identity separation

The GitHub sandbox adapter aggregates three expiring selected-repository App installation
credentials: reconciler, seeder, and rules. Each role is checked against its expected
actor, account, numeric installation, permission manifest, and expiry. Tokens are injected
at request time, omitted from JSON, and never written to plans, receipts, or state. The
reviewer machine user and its fine-grained token are separate from all three App roles.

Deterministic tests use the same public lifecycle and adapter seams with an in-memory
adapter and native HTTP fixtures. Their evidence mode does not claim live GitHub behavior.
