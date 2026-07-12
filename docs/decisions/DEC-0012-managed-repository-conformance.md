# DEC-0012 — Managed-repository conformance

**Status:** Accepted  
**Owner:** dragondad22  
**Date:** 2026-07-11  
**Source decision:** D12

## Context

A managed repository needs a precise product promise. Generic green badges conceal
coverage gaps, stale evidence, inapplicable controls, and accepted risks.

## Decision

Conformance is versioned evidence for an explicit scope and gate. Every control has a
stable identity and one explicit state: pass, fail, not applicable, not configured, needs
review, or accepted exception with the underlying result retained. Releases produce
human summaries and machine evidence manifests and disclose all coverage limits.

## Consequences

Pass always points to current evidence. Not applicable records facts and rules. Risks
record owner, rationale, scope, approval, expiry/review, compensating controls, and closure.
Retrofit cannot rewrite prior history; upgrades cannot silently weaken controls;
independent runs against the same versions must reproduce the result.

## Source

[Discovery decision D12](../discovery/CODEX_STARTER_KIT_REVIEW.md#d12) and its fourteen-part
managed-repository contract.
