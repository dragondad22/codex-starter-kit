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

At the surface where a persona naturally makes a decision, expose the minimum relevant
outcome, scope, state, coverage, uncertainty, downstream impact, and breadcrumb needed to
interpret that surface correctly. Keep deeper detail in one authoritative record rather
than duplicating it. Derived operational views must reconcile when authoritative inputs
change; essential context must not depend on the human knowing which hidden document to
open.

## Consequences

A generated registry owns current paths and load-when metadata. Broken, duplicated,
orphaned, or recursively mandatory routes fail validation. Backlinks and impact lists are
generated; root routing has an explicit context budget.

## Source

[Discovery decision D11](../discovery/CODEX_STARTER_KIT_REVIEW.md#d11) and its breadcrumb
interface workshop. The point-of-use application was clarified through
[issue #23](https://github.com/dragondad22/codex-starter-kit/issues/23).
