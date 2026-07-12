# DEC-0009 — Native platform support

**Status:** Accepted  
**Owner:** dragondad22  
**Date:** 2026-07-11  
**Source decision:** D9

## Context

Requiring a Unix compatibility layer on Windows or embedding OS-specific shell behavior
would contradict the cross-platform product and reproduce portability failures found in
the Claude kit.

## Decision

Support Linux, macOS, and Windows natively in the first release. WSL, Git Bash, and
containers are optional adapters. Universal formats and workflows are platform-neutral
and do not depend on Bash, PowerShell, GNU tools, or one filesystem's behavior.

## Consequences

The lifecycle engine owns native paths, atomic writes, locks, process execution,
permissions/capability detection, and line endings. A published CI-backed support matrix
defines exact OS, architecture, filesystem, and Codex-client compatibility.

## Source

[Discovery decision D9](../discovery/CODEX_STARTER_KIT_REVIEW.md#d9).
