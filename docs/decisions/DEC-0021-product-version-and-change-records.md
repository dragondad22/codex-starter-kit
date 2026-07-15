# DEC-0021 — Product version and change records

**Status:** Accepted
**Owner:** dragondad22
**Date:** 2026-07-15
**Source decision:** Issue #78

## Context

The plugin already carried development versions while the standalone engine reported
build metadata and the repository had no consolidated changelog. Git history and issue
evidence preserved detailed traceability, but they could not deterministically generate
current user, operator, developer, security, and stakeholder release communication.
Minor and patch releases before `1.0.0` require one truthful product identity without
collapsing independently evolving compatibility contracts into that identity.

## Decision

Use one Codex Starter Kit product release version across same-release distribution
surfaces, beginning at `0.3.0` to align the existing plugin development distribution. The
root schema-v1 `product-version.json` record is authoritative. The standalone CLI and
engine capability handshake report it, and validation requires the plugin manifest and
both plugin capability contracts to match it.

Keep lifecycle protocol, JSON schema, policy-pack, professional-baseline, managed-state,
and other compatibility versions independent. They change only when their own contracts
change and never merely to follow a product release.

Use validated schema-v1 JSON change records as the single human-owned source of release
communication. Each material change records its category, audiences, components, durable
issue or pull-request references, breaking state, and either external audiences or an
explicit internal-only disposition. Generate `CHANGELOG.md`, audience-filtered summaries,
and prepared release manifests from those records; do not maintain duplicate summaries by
hand.

Release preparation and publication are distinct. Preparation requires an explicit
greater stable SemVer, date, and admission binding exact records to Milestone, aggregate
issue, and approver authority; archives exact records with digests; synchronizes declared
same-release versions, regenerates views, and records `prepared` plus `published: false`.
It uses a durable recovery journal and performs no Git or external effect. A later
approved adapter publishes only an exact merged candidate through a tag and GitHub
Release, records evidence, and reconciles the aggregate release issue, Milestone, and
Project.

## Consequences

Every material pull request must add a change record or an explicit internal-only record,
and CI can reject malformed records, version skew, or a stale changelog. A version bump or
prepared changelog is never proof of publication. Existing notable development work is
backfilled as Unreleased communication without claiming a historical public release.

The first implementation provides local record validation, rendering, stale-view checks,
and compensated preparation. Tagging, GitHub Release creation, artifact publication,
signing, candidate qualification, and approval remain downstream release-adapter work.

## Source

Approved through [issue #78](https://github.com/dragondad22/codex-starter-kit/issues/78).
