# DEC-0010 — Claude-kit semantic migration

**Status:** Accepted  
**Owner:** dragondad22  
**Date:** 2026-07-11  
**Source decision:** D10

## Context

Existing Claude-kit projects contain valuable decisions, interviews, standards, modules,
and GitHub history. File-for-file compatibility would import Claude-specific runtime
assumptions and create two competing instruction systems.

## Decision

Recognize known Claude-kit versions and produce a semantic migration plan. Preserve valid
human records and GitHub history; classify artifacts as adopt, transform,
retain-as-history, supersede, conflict, or unsupported. Do not promise runtime-file or
command compatibility.

## Consequences

Ambiguous or lossy mappings require human resolution. Temporary Claude/Codex coexistence
must declare authority and synchronization. Obsolete Claude runtime files are removed
only through a planned migration with a durable mapping record.

## Source

[Discovery decision D10](../discovery/CODEX_STARTER_KIT_REVIEW.md#d10).
