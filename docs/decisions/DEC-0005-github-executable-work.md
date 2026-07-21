# DEC-0005 — GitHub executable work

**Status:** Accepted  
**Owner:** dragondad22  
**Date:** 2026-07-11  
**Amended:** 2026-07-21 by owner clarification in issue #46
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

Project fields and their semantics are governed facts. Saved Project views are optional,
human-owned presentation: individuals and teams may create, omit, and arrange them to fit
their work. A managed repository must preserve those choices rather than prescribe or
normalize a universal layout. An explicitly requested view may be observed or managed only
within its own declared contract and authority; its absence does not invalidate the
underlying Project fields or executable-work contract.

## Consequences

Implementation cannot begin from Intake or Needs refinement. Changed decisions, policy,
references, or facts invalidate readiness instead of authorizing invention during coding.
Completed issues retain PR, evidence, communication, and deviation memory.
Teams may use the four D5 view examples as useful defaults, but conformance and automation
cannot require them unless a later team-specific contract explicitly adopts them.

## Source

[Discovery decision D5](../discovery/CODEX_STARTER_KIT_REVIEW.md#d5), derived from the
Claude kit's issue and roadmap standards.
