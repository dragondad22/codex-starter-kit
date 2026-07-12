# DEC-0007 — Policy distribution and predictable layout

**Status:** Accepted  
**Owner:** dragondad22  
**Date:** 2026-07-11  
**Source decision:** D7

## Context

Copied standards drift and create noisy upgrades; cloud-only standards break offline and
historical proof. A rigid physical directory tree cannot fit every supported project type.

## Decision

Distribute signed immutable policy packs pinned by repository lockfile and digest, with a
verified local cache and optional mirror/vendoring. Map stable logical directory roles to
project paths and expand structure through inspected, planned layout rules.

## Consequences

Team members, AIs, and CI resolve identical policy online or offline. Project decisions
and evidence remain local. Policy and layout upgrades are semantic, explicit, and
transactional; a mutable latest version never defines conformance.

## Source

[Discovery decision D7](../discovery/CODEX_STARTER_KIT_REVIEW.md#d7) and the policy-pack and
logical-role workshops.
