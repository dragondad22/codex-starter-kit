# DEC-0011 — Governed breadcrumbs

**Status:** Accepted  
**Owner:** dragondad22  
**Date:** 2026-07-11  
**Source decision:** D11

## Context

Loading every standard wastes context and hides relevant rules, while informal links
become stale and change agent behavior without detection.

## Decision

Require stable IDs for governed policy, controls, decisions, and specifications. Use a
mechanically validated one-way context graph from orientation to workflow to policy to
evidence. Ordinary explanatory pages may use normal relative links unless another
artifact depends on them authoritatively.

## Consequences

A generated registry owns current paths and load-when metadata. Broken, duplicated,
orphaned, or recursively mandatory routes fail validation. Backlinks and impact lists are
generated; root routing has an explicit context budget.

## Source

[Discovery decision D11](../discovery/CODEX_STARTER_KIT_REVIEW.md#d11) and its breadcrumb
interface workshop.
