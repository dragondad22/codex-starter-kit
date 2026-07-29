# Codex skill eligibility and informed-interaction evaluation

**Status:** Approved and promoted to DEC-0024
**Issue:** [#87](https://github.com/dragondad22/codex-starter-kit/issues/87)
**Freshness:** 2026-07-29
**Decision:** [DEC-0024](../decisions/DEC-0024-skill-eligibility-and-informed-interaction.md)

## Objective and stopping conditions

Define an evidence-backed contract for when a Starter Kit capability should be a skill,
where that skill should live, and which interaction, authority, and qualification
guarantees it must preserve. Apply the contract to every current candidate so #86 and its
children can proceed without treating every useful behavior as a skill.

The evaluation stops when current official Codex skill/plugin capabilities are recorded;
all relevant product surfaces are distinguished; an eligibility and interaction contract
can disposition the entire candidate inventory; reasonable alternatives and negative
paths are explicit; and downstream work has a concrete owner and dependency. It does not
implement, install, package, enable, disable, publish, or qualify a new skill or plugin.

## Method and provenance

The evaluation combined:

1. current authoritative Starter Kit decisions, product, persona, architecture, plugin,
   and issue records at `origin/main` revision
   `df50af38641979456c2bfe8bf410109b1d4e78f3`;
2. the exact Ready issue #87 and its parent/candidate issue relationships;
3. the fresh Codex manual retrieved through the official OpenAI documentation helper on
   2026-07-29; and
4. observed local personal skill availability in the active Codex environment, used only
   to inventory candidates rather than establish a portable support claim.

Primary official sources:

- [Build skills](https://developers.openai.com/plugins/build/skills)
- [Build plugins](https://developers.openai.com/plugins/build/plugins)
- [Use plugins](https://learn.chatgpt.com/docs/plugins)

Official documentation establishes host capabilities, not Starter Kit product authority.
The active environment loads personal skills from `/home/chris/.codex/skills`, while the
current public documentation names `$HOME/.agents/skills`; that observation is
environment-specific and is not used as a supported location or portability claim.

## Current official capability snapshot

| Capability | Current documented fact | Starter Kit implication |
|---|---|---|
| Skill purpose | Task-specific reusable workflows made from `SKILL.md`, resources, and optional scripts | A skill is a focused method, not a general bucket for all agent behavior |
| Context loading | Name and description load first; full instructions load after selection | Names, descriptions, catalog size, and trigger overlap have real routing/context cost |
| Activation | Explicit mention or implicit description matching | Product qualification needs positive, negative, and overlap prompts |
| Invocation policy | Optional metadata can disable implicit invocation | High-risk or easily confused workflows may require explicit activation |
| Dependencies | Optional metadata may declare MCP tool dependencies | Declaration improves discovery but does not grant access, credentials, or effect authority |
| Standalone surfaces | ChatGPT desktop app, Codex CLI, and IDE extension | A repo/user skill can be available independently of plugin support |
| Local scopes | repository, user, administrator, and system locations | Scope expresses intended audience and ownership; it does not prove qualification |
| Repository discovery | `.agents/skills` is scanned from the working directory toward the repository root | Area-specific and root-shared workflows can be placed narrowly |
| User discovery | `$HOME/.agents/skills` | General cross-repository workflows can remain personal |
| Workspace boundary | Workspace controls can govern shared plugin availability and administration | Workspace access is distinct from local skill scope, tool access, and effect authority |
| Plugin packaging | One or more skills, optionally with connectors/MCP configuration and presentation assets | Plugin distribution is justified by a reusable product portfolio or integration boundary |
| Plugin surfaces | ChatGPT Work web, ChatGPT desktop Work/Codex, and Codex CLI; not Chat, IDE extension, or mobile | Standalone-skill availability and plugin availability must not be conflated |
| Enable/disable | Local skills can be disabled through Codex configuration; plugin controls are surface/admin dependent | #90 needs bounded research before proposing one management experience |

This snapshot changes DEC-0018's July 13 statement that official pages conflicted about
IDE plugin availability. Current documentation consistently distinguishes standalone
skills, which are available in the IDE, from plugins, which are not. It does not change
the Phase 2 choice to distribute `create`, `status`, and `verify` as a skills-only plugin
or the need to qualify each claimed supported surface.

## Product-surface comparison

| Surface | Best fit | Not its job | Starter Kit example |
|---|---|---|---|
| Product interaction contract | Universal conversational and decision quality across adapters | Optional activation or task-specific procedural detail | Investigate first; make material decisions informed |
| `AGENTS.md` / repository instructions | Standing repository workflow, constraints, commands, and routing | A reusable cross-project product workflow | Ready-before-work, one writer, required gates |
| Human-owned product/architecture docs | Durable outcomes, semantics, boundaries, and ownership | Step-by-step invocation behavior | Authority and evidence rules |
| Templates/assets | Repeatable artifact shape or starting content | Adaptive judgment and live orchestration | Issue forms, policy templates |
| Skill | Focused reusable model guidance for one recognizable goal | Deterministic invariants, broad policy, or unrelated routing | Guided `verify` |
| Lifecycle engine | Deterministic state, plans, effects, verification, and evidence | Open-ended discovery or conversational explanation | `create`, `status`, `verify` operations |
| Policy/control | Applicability, prohibition, required evidence, and non-pass semantics | Friendly task routing or tool orchestration | Earliest relevant risk gate |
| Hook/automation | Stable event- or schedule-driven action | A workflow that still needs material steering | Later deterministic maintenance trigger |
| MCP/connector/adapter | Authenticated external capability and transport | The human-facing method for deciding when/how to use it | GitHub API access |
| Plugin | Installable distribution of one or more qualified skills and optional integrations | Proof that every bundled capability is safe, authorized, or supported everywhere | Starter Kit skills-only adapter |

The important distinction is ownership. A skill may explain, sequence, and call another
surface, but it cannot take over that surface's authority.

## Eligibility contract v1 and quality matrix

A candidate must pass every eligibility row. A failed row is a routing result, not a
lower-quality skill. The seven-row contract is version 1 of the DEC-0024 eligibility test;
a semantic change requires an explicit decision amendment so later qualifications remain
reconstructable.

| Test | Evidence expected | Failure route |
|---|---|---|
| Recognizable user goal | Natural prompts and an instinctive outcome-oriented name | Ordinary behavior, split candidate, or reject |
| Clear activation boundary | Explicit/implicit triggers, negative triggers, overlap cases | Refine/split, explicit-only policy, or reject |
| Repeated procedural value | Two or three demonstrated use cases or repeated correction pattern | Keep conversational/personal or document normally |
| Adaptive workflow | Meaningful judgment or orchestration that cannot be a deterministic invariant | Engine, policy, template, hook, or automation |
| Bounded contract | Inputs, outputs, effects, authority, data, dependencies, stops, fallback | Needs refinement |
| Qualifiable semantics | Scenario suite with expected behavior and non-pass outcomes | Research or reject |
| Net catalog value | Benefit exceeds routing, context, maintenance, ownership, and memory cost | Combine, keep personal, or reject |

Passing eligibility is necessary but not sufficient for product distribution. A product
skill also needs a Ready implementation issue, owner, versioned package, supported-surface
claim, current dependency evidence, documentation, change/evidence record, and distinct
review.

## Informed discovery analysis

### Preserved value

The intensive “grilling” pattern contributes three useful properties:

- it does not stop at the first plausible answer;
- later questions adapt to earlier answers; and
- it can route approved results into briefs, personas, decisions, architecture, or other
  durable artifacts.

Those properties should be preserved for project and feature shaping. They do not require
the user to remember a large artifact taxonomy or decide every detail before receiving
context.

### Corrected failure mode

The harmful pattern is “recommend, then ask for agreement” when the user cannot yet see
what the choice changes or which alternatives are reasonable. The corrected sequence is:

1. inspect known authorized context;
2. identify the current decision and why it matters;
3. ask the smallest question that materially reduces uncertainty;
4. repeat only while the answer changes the path;
5. present options, consequences, reversibility, uncertainty, authority, and a reasoned
   recommendation; and
6. ask for the material decision, then route only approved durable results.

This remains compatible with one-question-at-a-time discovery. It also supports delegated
work: the agent resolves routine choices inside existing authority and interrupts only
for a material decision or expansion.

### Artifact routing

Discovery does not imply artifact production. A result is written only when it crosses
the applicable durable threshold:

| Result | Destination |
|---|---|
| Ordinary clarification | Current conversation |
| Stable product outcome or requirement | PRD or focused specification |
| Audience-specific need | Existing persona record, or a new persona only when genuinely distinct |
| Material product/architecture choice | Decision record |
| Executable work | Existing or new lifecycle-appropriate GitHub issue |
| Deterministic lifecycle fact | Structured managed-repository state through the engine |
| Research finding | Bounded research record; not authority until promoted |

Conversation, approval to promote, and authorization to implement remain separate.

## Candidate inventory and disposition

| Candidate | Eligibility result | Distribution/result |
|---|---|---|
| Plugin `create` | Pass | Retain product plugin skill; deterministic effects remain in engine |
| Plugin `status` | Pass | Retain product plugin skill; read-only state remains engine-owned |
| Plugin `verify` | Pass | Retain product plugin skill; evidence and non-pass semantics remain engine-owned |
| Personal `think-it-through` | Pass as general personal workflow; not Starter Kit-specific | Keep personal; promote informed-decision guarantees universally |
| #114 informed shaping | Plausible but unqualified | Refine one opt-in deep project/feature shaping skill after #87 |
| Personal `review-this` | Pass as broad personal workflow; not the governed review product contract | Keep personal |
| #88 governed review | Plausible after deterministic delivery seam | Blocked by #75 and #87; refine and qualify separately |
| Personal `find-the-cause` | No current product-specific evidence | Keep personal; do not add to product catalog |
| Personal `make-the-change` | Fails surface-ownership test for Starter Kit | Route to `AGENTS.md`, Ready issues, Work Manager, engine, and adapters |
| #89 handoff/resume | Deterministic state is prerequisite | Specify Work Manager projection first; reconsider guided skill only if proven useful |
| #90 selective management | Host/distribution question unresolved | Keep as bounded research; do not assume a manager skill |

## Alternatives and consequences

| Alternative | Benefit | Material cost | Disposition |
|---|---|---|---|
| Retain only `create`/`status`/`verify` permanently | Small stable catalog | Prevents evidence-backed new user goals | Reject as permanent policy; retain them now |
| Add repository-scoped development skills | Immediate team reuse without product distribution | Can mix internal workflow with customer contract | Allow only for demonstrated repository-specific work |
| Package all personal workflows | One installable portfolio | Confuses personal preference with product value and increases trigger overlap | Reject |
| Use one broad manager/router skill | One name to remember | Hides unrelated authority, becomes vague, and duplicates host routing | Reject |
| Keep general workflows personal | Fast iteration and cross-repository usefulness | No product support claim | Approve for current generic skills |
| Defer candidates until deterministic seams exist | Preserves authority and testability | Delays guided experience | Approve for #88/#89 where dependencies are real |
| Rely on model judgment without a contract | No artifact maintenance | Inconsistent promotion and repeated debate | Reject |

## Qualification plan

Every later product-skill delivery must define expected results for:

1. direct explicit invocation;
2. natural implicit activation;
3. prompts that must not activate it;
4. overlap with every adjacent skill;
5. incomplete or contradictory inputs;
6. discoverable facts already present, proving the skill does not re-ask needlessly;
7. adaptive one-question discovery and a bounded stopping condition;
8. a material decision where alternatives and consequences are not yet known;
9. premature recommendation and excessive-questioning prevention;
10. artifact requests that are not justified by a durable threshold;
11. missing, disabled, malformed, incompatible, or unauthorized dependencies;
12. unsupported host and explicitly narrower fallback;
13. denied or absent decision, data, tracker, network, installation, or effect authority;
14. attempted promotion, tracker mutation, or implementation from conversational agreement;
15. interruption, cancellation, replay, and stale context where effects are possible; and
16. semantic outcome across each claimed model/surface, retaining different prose as valid.

Qualification evidence names exact source revision, skill and plugin identity, host,
model/capabilities, dependency identities, scenario fixture version, observed result,
limitations, and reviewer. A green happy-path transcript alone is insufficient.

## Downstream dependency map

| Issue | Result of #87 |
|---|---|
| #86 | Remains `In progress / Needs refinement` while concrete children remain |
| #88 | Retains #75 and #87 as blockers until both are complete; then requires fresh refinement |
| #89 | #87 resolves eligibility, but Work Manager state/projection is the preferred first surface |
| #90 | Remains research; current scope/admin/distribution behavior must be evaluated before design |
| #114 | Becomes the next candidate for refinement after #87; no implementation is pre-approved |

## Uncertainty and limitations

- Official product documentation can change; capability claims are fresh only to
  2026-07-29 and trigger review under DEC-0018/DEC-0024 when they drift.
- This evaluation did not install, move, enable, disable, package, or execute a new skill.
- Personal skill behavior in the active environment is design input, not qualification
  evidence for the Starter Kit.
- Implicit routing and prose can vary by model. Qualification requires semantic boundaries,
  not exact wording.
- The evaluation does not decide #90's management experience, #89's state schema, #88's
  review implementation, or #114's exact shaping artifacts.
