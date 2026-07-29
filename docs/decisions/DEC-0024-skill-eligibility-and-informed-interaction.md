# DEC-0024 — Skill eligibility and informed interaction

**Status:** Accepted
**Owner:** dragondad22
**Date:** 2026-07-29
**Source decision:** Issue #87

## Context

Codex can apply reusable workflows through standalone skills or skills distributed in a
plugin. That makes a skill an attractive home for any useful agent behavior, but a large
catalog creates routing ambiguity, consumes discovery context, burdens users with names
to remember, and duplicates rules already owned by product specifications, repository
instructions, policy, deterministic engine operations, or tool integrations.

The owner's experience with an intensive project-discovery workflow also exposed a
separate problem. Persistent questioning can uncover useful detail, but asking someone
to accept a recommendation before showing the relevant context, reasonable alternatives,
consequences, reversibility, uncertainty, and decision authority turns discovery into
uninformed approval. The desired behavior is deep enough to reach decision-ready
understanding while remaining natural, efficient, and useful even when no skill is
installed or activated.

DEC-0018 already makes focused `create`, `status`, and `verify` skills an experience
adapter over the deterministic lifecycle engine. The product needs a repeatable boundary
for future candidates without making those three names a catalog template.

## Decision

### Universal informed-interaction contract

Every Starter Kit interaction surface applies these guarantees whether or not a skill is
active:

1. Investigate authorized, discoverable context before asking the user to repeat it.
2. Ask only questions whose answers materially change the outcome, recommendation,
   authority, or safe next action. Prefer one focused question at a time when the answer
   will shape the next question; batch only independent, low-cost facts when doing so is
   clearer.
3. Continue discovery until the current decision is ready, not until a fixed questionnaire
   or artifact list is exhausted. Persistent questioning is a means, not the outcome.
4. Before requesting a material decision, present the known context, realistic options,
   material consequences, reversibility, important uncertainty, the recommended option
   and its basis, and who has authority to decide. If that basis is not yet sufficient,
   investigate or ask rather than recommend with false confidence.
5. Distinguish advice, conversational agreement, durable promotion, and effect authority.
   A skill or agent may recommend and prepare, but it may not infer missing human
   authority, silently write an authoritative record, mutate a tracker, or implement a
   consequential decision merely because the conversation appears settled.
6. Route only approved, durable conclusions to the appropriate decision, specification,
   issue, plan, persona, architecture, or evidence record. Do not generate personas,
   architecture decisions, or other artifacts merely because a template permits them.
7. Preserve explicit uncertainty, disagreement, missing capability, and non-pass states.
   Offer a bounded fallback or safe stop instead of manufacturing completeness.

Engagement mode may change cadence and evidence presentation under DEC-0019, but it does
not remove these guarantees. Delegated work may resolve routine choices within existing
authority without interruption; collaborative work may expose more intermediate
reasoning. Both must make material owner decisions informed.

### Skill eligibility contract v1

A **skill** is task-specific reusable guidance that helps an AI apply judgment or
orchestrate tools for one recognizable user goal. A proposed Starter Kit capability
becomes a skill only when evidence supports every condition below:

1. **Recognizable goal:** users naturally ask for one coherent outcome and can understand
   the skill's name without learning internal architecture.
2. **Activation boundary:** positive triggers, negative triggers, and overlap with nearby
   skills are clear enough to test explicit and implicit activation.
3. **Reusable procedural value:** repeated use benefits materially from focused
   instructions, references, examples, or orchestration beyond the base model and
   universal repository guidance.
4. **Appropriate execution surface:** the work requires adaptive model judgment or
   tool-sequence guidance. Deterministic lifecycle behavior remains in the engine;
   standing rules remain in specifications, policy, templates, or `AGENTS.md`; external
   capabilities remain in their adapters, MCP servers, connectors, hooks, or automations.
5. **Bounded contract:** inputs, outputs, possible effects, facts that must not be inferred,
   stop/ask/decline behavior, authority, data access, dependencies, portability,
   degraded behavior, and fallback are explicit.
6. **Qualifiable behavior:** representative and negative scenarios can test activation,
   non-activation, semantic outcome, user experience, dependency failure, unsupported
   hosts, authority containment, and prohibited effects.
7. **Net catalog value:** the recurring value exceeds the routing ambiguity, context,
   maintenance, versioning, ownership, and user-memory cost introduced by another skill.

Failing any condition routes the capability to its proper surface or retains it as
research. Related behavior may share governed references, but several skills must not copy
universal interaction or authority rules into divergent prompt contracts. A broad router
that merely chooses among unrelated jobs is not a substitute for recognizable goals. A
candidate is too broad when its natural triggers imply independent outcomes, authorities,
effects, or fallbacks. Split it only when each result is itself a recognizable user goal;
do not create several narrow names when one coherent goal and shared contract would make
users remember artificial internal stages. A semantic change to these seven conditions
requires an explicit DEC-0024 amendment or superseding decision.

### Scope and distribution

Skill location expresses audience, not authority:

- repository-scoped standalone skills are for demonstrated workflows specific to this
  repository or one of its areas;
- user-scoped skills are personal workflows useful across repositories;
- administrator and system skills are environment- or platform-owned;
- workspace availability and administration govern shared plugin access and policy; they
  do not create another local repository-skill scope or authorize local effects; and
- the Starter Kit plugin distributes qualified product workflows to supported plugin
  surfaces.

Availability, discovery, installation, or activation grants no filesystem, process,
network, connector, data-handling, decision, tracker, or effect authority. Each dependency
and effect retains its own contract. Repository or personal experimentation is not product
qualification, and moving a skill into the plugin requires an issue-specific Ready
delivery with current evidence.

### Qualification and portfolio disposition

Each product skill must retain:

- two or three representative user goals before broader expansion;
- explicit, implicit, and should-not-trigger prompts;
- incomplete-input, excessive-questioning, premature-recommendation, and artifact-sprawl
  scenarios where interaction is material;
- missing dependency, unsupported surface, degraded/fallback, and semantic-drift cases;
- attempted authority leakage, tracker mutation, promotion, data access, and
  implementation without authorization;
- equivalent required semantics across each claimed model and host, without requiring
  identical prose; and
- exact skill, plugin, host, dependency, fixture, and source identities in evidence.

The current portfolio is dispositioned as follows:

| Candidate | Disposition | Reason and route |
|---|---|---|
| `create`, `status`, `verify` | Retain as product skills | Each has a recognizable engine-backed lifecycle goal, bounded effects, tested activation, capability modes, and direct-engine fallback under DEC-0018. |
| General `think-it-through` | Personal skill | Useful across repositories, but its informed-decision lessons are universal product behavior rather than a Starter Kit product skill. |
| Issue #114 informed project/feature shaping | Eligible after #87; remains blocked pending task refinement | It may warrant one opt-in deep shaping skill because it owns a recognizable discovery outcome and durable artifact routing. It must demonstrate value beyond ordinary conversation and avoid a predetermined artifact factory. |
| General `review-this` | Personal skill | Broad second-opinion review remains useful personally; it does not itself satisfy the Starter Kit's distinct governed PR-review contract. |
| Issue #88 governed review | Eligible after #87 and #75; remains blocked | A product skill may guide the distinct review workflow once governed delivery supplies the exact source/evidence seam and task-specific qualification. |
| General `find-the-cause` | Personal skill | No demonstrated Starter Kit-specific procedural value currently exceeds the catalog cost. |
| General `make-the-change` | Ordinary governed delivery behavior | Ready issue execution belongs in repository instructions, issue contracts, the Work Manager, lifecycle engine, and adapters. A product skill would duplicate standing workflow and risk becoming an authority-shaped router. |
| Issue #89 handoff and resumption | Different product surface first | Durable execution state and resumption belong in Work Manager state/projections. A later skill is justified only if guided human routing adds demonstrated value after that deterministic seam exists. |
| Issue #90 selective skill management | Needs research | Current host scope, enable/disable, administration, and plugin behavior require bounded research. No manager or product skill is assumed. |

Approval of this decision does not make any candidate implementation Ready.

## Consequences

The product has a stable answer to both “why is this a skill?” and “why is this not another
surface?” without turning catalog symmetry into product architecture. Skill proposals
must carry evidence, a downstream owner, and a supported distribution route. Removing or
combining a skill remains valid when its value no longer exceeds its discovery and
maintenance burden.

Issue #114 can now be refined as the next informed-shaping question. Issue #88 retains
both #75 and this decision as predecessors. Issue #89 should first define the Work Manager
projection it depends on. Issue #90 remains bounded research. Parent #86 remains open while
those concrete children are incomplete.

Rejected alternatives are retaining only the existing three skills forever, turning every
useful behavior into a product skill, packaging all personal workflows, using one broad
manager/router skill, or relying on conversational taste without qualification. Those
options respectively freeze learning, create catalog sprawl, confuse personal and product
ownership, hide unrelated authority behind one entry point, or make promotion
irreproducible.

Return this decision to review when supported skill activation or distribution changes
materially, catalog scale invalidates discovery assumptions, cross-model qualification
cannot preserve required semantics, or a candidate exposes an authority or product-surface
conflict the eligibility test cannot resolve.

## Source

Approved by the product owner through
[issue #87](https://github.com/dragondad22/codex-starter-kit/issues/87). The bounded
[skill eligibility evaluation](../research/CODEX_SKILL_ELIGIBILITY_EVALUATION.md)
preserves current official capability evidence, alternatives, the complete comparison,
qualification plan, uncertainty, and downstream routing.
