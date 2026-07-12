# DEC-0004 — State and document authority

**Status:** Accepted  
**Owner:** dragondad22  
**Date:** 2026-07-11  
**Source decision:** D4

## Context

Policy computation needs validated structured state, while humans need durable prose and
editable records. Bidirectional free-form synchronization creates competing truths and
silent overwrites.

## Decision

Schema-versioned structured state is authoritative for lifecycle facts, policy inputs,
versions, and provenance. Generated views are reproducible projections with source
digests. Approved briefs, personas, decisions, risks, specifications, and stakeholder
documents are human-owned records.

## Consequences

Human proposals pass validation before updating machine state. Conflicts stop for a
reviewable reconciliation plan. Generated files never silently become authoritative
input, and human-owned prose is never silently regenerated.

## Source

[Discovery decision D4](../discovery/CODEX_STARTER_KIT_REVIEW.md#d4) and its three ownership
classes.
