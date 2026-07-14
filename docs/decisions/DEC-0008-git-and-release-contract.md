# DEC-0008 — Git and release contract

**Status:** Accepted  
**Owner:** dragondad22  
**Date:** 2026-07-11  
**Source decision:** D8

## Context

All changes need protected traceability, but applications, packages, infrastructure,
data, documentation, and monorepos do not share one valid version or release mechanism.

## Decision

Use protected default branches and the default flow Ready issue → issue branch → scoped
PR → required gates → squash merge. Use Conventional Commit-style surviving titles and
one source of change records. Select version, publication, deployment, signing, and merge
queue adapters from project context and policy.

## Consequences

Milestones represent releases only. Emergency work uses a governed break-glass path.
Release operations are transactional, evidence-backed, audience-aware, and tied to
immutable source/artifact identity without forcing invented SemVer on every repository.
Draft PR status communicates that more implementation, verification, or internal review
is expected; a completed, verified, internally reviewed PR is moved to ready for review.
[DEC-0020](DEC-0020-distinct-pull-request-review.md) defines the required distinct review
pass and its separation from automated checks, self-review, outcome approval, and stronger
qualified assurance.

## Source

[Discovery decision D8](../discovery/CODEX_STARTER_KIT_REVIEW.md#d8) and its universal Git
and release contracts.
