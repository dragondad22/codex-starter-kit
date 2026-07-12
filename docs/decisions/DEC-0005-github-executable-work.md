# DEC-0005 — GitHub executable work

**Status:** Accepted  
**Owner:** dragondad22  
**Date:** 2026-07-11  
**Source decision:** D5

## Context

Issues must serve humans and preserve enough implementation context for another AI or
developer to execute work without the originating conversation. Roadmap and execution
state drift when duplicated in documents.

## Decision

Every managed repository uses GitHub Issues and exactly one linked GitHub Project. An
executable issue contains a human summary and complete AI brief and must pass Readiness.
Status tracks execution; Horizon tracks feature intent; Readiness tracks executability.
The live Project is the roadmap authority and is reconciled automatically.

## Consequences

Implementation cannot begin from Intake or Needs refinement. Changed decisions, policy,
references, or facts invalidate readiness instead of authorizing invention during coding.
Completed issues retain PR, evidence, communication, and deviation memory.

## Source

[Discovery decision D5](../discovery/CODEX_STARTER_KIT_REVIEW.md#d5), derived from the
Claude kit's issue and roadmap standards.
