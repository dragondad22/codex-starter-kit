# DEC-0006 — Stage-specific enforcement

**Status:** Accepted  
**Owner:** dragondad22  
**Date:** 2026-07-11  
**Source decision:** D6

## Context

Blocking all thought when a tool is absent prevents remediation, while allowing work to
proceed to publication creates the risk the missing control was intended to stop.

## Decision

Missing tools, integrations, checks, or evidence never pass. Each control declares the
earliest lifecycle gate it blocks. Safe read-only planning and constrained remediation may
continue, while commit, external mutation, merge, release, or conformance is blocked at
the relevant boundary.

## Consequences

The system first presents trusted upgrade, installation, configuration, manual-evidence,
and fallback paths. It never silently installs, broadens authority, transmits data, or
substitutes weaker evidence. D2 exception rules remain authoritative.

## Source

[Discovery decision D6](../discovery/CODEX_STARTER_KIT_REVIEW.md#d6) and its stage table.
