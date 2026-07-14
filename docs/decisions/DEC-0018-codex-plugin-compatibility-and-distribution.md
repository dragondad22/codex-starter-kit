# DEC-0018 — Codex plugin compatibility and distribution

**Status:** Accepted
**Owner:** dragondad22
**Date:** 2026-07-13
**Source decision:** Issue #50

## Context

DEC-0003 makes an installable Codex plugin the preferred experience adapter while keeping
the lifecycle engine and managed repository independently authoritative. Phase 2 cannot
implement that adapter truthfully without deciding its initial component shape,
development distribution, compatibility evidence, capability modes, offline boundary,
and publication handoff.

Current official documentation supports skills-only plugins and local, repository,
Git-backed, npm-backed, workspace, and public-directory distribution paths. It does not
document a version string that alone proves the required plugin capabilities, and current
official pages conflict about IDE-extension plugin availability. A client-version guess
or silent selection of every optional plugin capability would create false support and
unnecessary authority.

## Decision

Phase 2 uses a **skills-only Codex plugin** for focused `create`, `status`, and `verify`
guidance. It does not add an app/connector, MCP server, hook, browser extension, scheduled
task, or remote service. Skills invoke the standalone engine through its public
executable/JSON contract; they never become conformance authority.

Development uses a repository marketplace and local plugin source. Qualification adds a
Git-backed marketplace snapshot pinned to an immutable source identity. npm distribution
is deferred until a separately approved need justifies its package-manager, registry,
authentication, and resolution surface. Phase 2 prepares a publication-ready skills-only
package and review cases but does not submit or publish it publicly. Publisher identity,
legal/support materials, signed release artifacts, release approval, and universal
directory publication remain Phase 6 authorities.

Compatibility is a non-mutating capability handshake, not a guessed minimum Codex
version. It records the host surface; plugin identity, source, enabled state, and
capabilities; engine resolution, identity, provenance, protocol/schema range, and
operations; managed-repository pins; baseline identity and compatibility; relevant
filesystem, process, sandbox, approval, network, and offline facts; and the requested
operation's authority and content-handling boundary. A recorded version is evidence for
that run, never sufficient proof by itself.

The handshake selects one workflow-specific mode:

| Mode | Contract |
|---|---|
| `full` | A compatible installed/enabled plugin, verified compatible engine and baseline, required operation, permissions, and approvals may guide and execute the requested supported workflow through the engine. |
| `degraded-guidance` | A missing, disabled, incompatible, or unverified execution capability permits read-only guidance, exact remediation, and direct-engine fallback only. |
| `verification-only` | A verified compatible engine may provide status, verification, and evidence when mutation is unavailable or unauthorized; no create/apply or migration effect is permitted. |
| `unsupported` | Unknown or failed required guarantees, policy conflict, unavailable safe fallback, or an unimplemented operation stops with bounded diagnostics and remediation. |

These modes describe workflow capability, not repository conformance. `full` can
truthfully return non-pass controls; no narrower mode may simulate an engine result or
appear as full conformance.

Codex CLI is the required development surface once its marketplace, installation, skill,
and workflow behavior passes qualification. ChatGPT desktop Codex is a candidate supported
surface requiring manual evidence on every claimed native environment. IDE extension
plugin distribution remains `needs-review` because official documentation conflicts;
observed installed skills do not prove a general marketplace or administration contract.
ChatGPT Work web is unsupported for the local lifecycle path without a separately verified
host route, and Chat/mobile are unsupported for Phase 2 local lifecycle operation.

Plugin, engine, repository schema, baseline/policy packs, and templates keep independent
versions and identities. Plugin installation or update does not install or replace an
engine, change repository pins, resolve policy online, or migrate a repository. A
compatible direct engine remains the required non-plugin and CI fallback.

Supported offline use requires the plugin marketplace snapshot or cache, verified engine,
baseline, compatibility metadata, and trusted roots to be provisioned in advance. Offline
execution verifies those local inputs and never silently fetches or runs remote content.

## Consequences

Issues #51–#54 can implement one narrow vertical slice without introducing optional
integration authority. The status tracer owns the minimal plugin, repository marketplace,
handshake, and first engine call. Guided create requires `full`; guided verify may use
`full` or `verification-only`; qualification must exercise every mode and retain the exact
client/plugin/engine/native identities and documentation conflicts.

The plugin can be useful before public publication, but no development marketplace,
cache, source build, or observed client version becomes a stable release claim. Requiring
users to install Go, an unverified engine, or silent network resolution remains outside
the supported 1.0 outcome.

Trust, authority, data access, cost, compatibility, and fallback remain independently
visible. The Phase 2 plugin adds no connector data path or external-service authority;
future apps, MCP, hooks, analytics, or remote services require their own approved use case
and review.

Return this decision to review if official plugin packaging or distribution changes, a
required surface cannot run the skills-only workflow, safe capability probing cannot
distinguish the four modes, the engine handshake requires a breaking contract, offline
operation gains an undisclosed network/package dependency, public review requires a
materially different package, or native qualification reveals semantic drift.

## Source

Approved by the product owner through [issue #50](https://github.com/dragondad22/codex-starter-kit/issues/50).
The bounded [compatibility and distribution evaluation](../research/CODEX_PLUGIN_COMPATIBILITY_EVALUATION.md)
preserves sources, method, observations, conflict, limitations, alternatives, and
downstream impact.
