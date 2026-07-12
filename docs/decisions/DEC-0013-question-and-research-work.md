# DEC-0013 — Question and research work

**Status:** Accepted
**Owner:** dragondad22
**Date:** 2026-07-12
**Source decision:** D13

## Context

Consequential questions and evidence-gathering can block delivery as materially as
implementation work. Leaving them only in chat or a monolithic discovery document makes
ownership, authorization, cost, dependencies, and resolution difficult to track. Making
every clarification an issue would instead flood the Project, fragment coherent discovery
history, and mistake issue discussion for authoritative knowledge.

## Decision

Managed repositories support `type:question` for consequential unresolved questions and
`type:research` for bounded evidence-producing work. Either a human or AI may identify or
perform this work within declared authority. Both types have subtype-specific readiness
and completion contracts and use the normal Project lifecycle.

A question becomes a work item only when its answer must survive the current conversation,
blocks or materially changes planned work, requires named authority or evidence, or is
likely to be referenced again. Ordinary conversational clarification stays in the active
workflow. Closing a question does not make its issue or comments authoritative: material
answers are promoted into the applicable decision, specification, policy, human-owned
record, or structured state. That authoritative destination references the issue, and the
issue's closing comment references the promoted record.

Research is Ready only when it declares an objective or bounded exploratory mapping goal,
intended use, scope, source and provenance expectations, depth or effort budget, authority,
stopping conditions, output, and review needs. Its durable human-owned research record
preserves method, sources, findings, conflicting evidence, uncertainty, limitations, and
freshness. Research informs later decisions; it never silently establishes them.

The discovery document remains coherent source history. Question issues supplement it and
capture new durable uncertainties; they do not replace it wholesale.

## Consequences

Readiness must generalize from implementation alone to actionable work while preserving
subtype-specific validation. Labels, forms, lifecycle behavior, completion checks, and
GitHub reconciliation must support the new types without treating relationships as hard
dependencies unless progress is genuinely blocked. Research cost and external authority
remain visible authorization choices. Durable results remain usable outside GitHub and
retain reciprocal provenance.

## Source

[Discovery decision D13](../discovery/CODEX_STARTER_KIT_REVIEW.md#d13), approved through
[issue #16](https://github.com/dragondad22/codex-starter-kit/issues/16).
