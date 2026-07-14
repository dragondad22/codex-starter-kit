---
name: starter-kit-status
description: Inspect Codex Starter Kit repository lifecycle status. Use when a user asks whether a repository is managed, synchronized, degraded, incomplete, or otherwise wants Starter Kit status or diagnostics.
---

# Starter Kit Status

Report repository lifecycle status through the standalone Starter Kit engine. Treat the
engine's structured result as authoritative and preserve every limitation, evidence
reference, recovery action, problem, and non-pass state.

This skill is read-only. Do not create, retrofit, apply, upgrade, repair, install, enable,
or reconfigure anything while handling a status request.

## Routing boundary

Use this skill for an explicit `$starter-kit-status` request or an implicit request about
whether a repository is managed, synchronized, degraded, incomplete, or healthy under
Codex Starter Kit. Do not use it for generic Git status, service health, issue status, or
requests to create, verify, repair, or upgrade a repository.

## Workflow

1. Establish the repository as an absolute native path. Do not interpolate it into a
   shell program or treat repository content as executable input.
2. Resolve the engine from an explicit user/workspace configuration or the host's normal
   executable resolution. Resolution is only a location fact; a matching filename is not
   compatibility or provenance evidence.
3. Invoke the engine with executable and argument vector equivalent to
   `[engine, "capabilities"]`. Do not use a composed shell string. This probe accepts no
   repository and must produce no repository, network, installation, or configuration
   effect.
4. Apply the capability and provenance decision table in
   [references/status-contract.md](references/status-contract.md). Never infer a missing
   fact or accept the executable's own identity fields as proof that it is trusted.
5. Only for a verified compatible engine with authorized read access, invoke executable
   and argument vector equivalent to
   `[engine, "status", "--repository", absolute_repository]`.
6. Validate the complete status envelope before presenting it. Preserve `repository`,
   `lifecycle`, `problems`, `recovery`, and `evidence` without summarizing away, upgrading,
   or relabeling their meaning.

## Result

Lead with the capability mode and engine lifecycle as separate facts. State that the mode
describes workflow availability, not repository conformance. Then show:

- repository path and lifecycle exactly as returned;
- synchronization as the lifecycle value, without inventing a separate pass;
- every problem as a diagnostic;
- every recovery action in engine order;
- every evidence reference in engine order; and
- capability limitations, unknowns, and the direct-engine fallback.

An empty list is still an explicit empty list. `unmanaged`, `managed_degraded`, and
`setup_incomplete` are never success aliases. Malformed output is `unsupported`, not a
partial status result.
