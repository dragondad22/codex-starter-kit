# DEC-0003 — Three-layer distribution

**Status:** Accepted  
**Owner:** dragondad22  
**Date:** 2026-07-11  
**Source decision:** D3

## Context

Codex plugins provide the intended guided experience, but client versions, workspace
policy, offline operation, CI, and deterministic verification cannot depend on an
interactive plugin alone.

## Decision

Ship a Codex plugin as the preferred experience adapter, a standalone lifecycle engine as
the deterministic authority, and a managed-repository contract as the durable record.
All three exist from the first usable release. MCP is optional, not foundational.

## Consequences

Plugin updates and repository upgrades are separate. CI and developers call the same
engine interface. Repositories remain explainable and verifiable when the plugin is
missing, degraded, administratively disabled, or offline.

## Source

[Discovery decision D3](../discovery/CODEX_STARTER_KIT_REVIEW.md#d3) and its Codex
capability spike.
