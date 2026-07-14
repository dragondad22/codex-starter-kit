# Codex Starter Kit — Persona Registry

**Status:** Initial project personas  
**Authority:** Human-owned audience reference  
**Review trigger:** Audience, governance, workflow, or distribution changes

## Purpose

This is the single definition of the human audiences the Codex Starter Kit serves.
Briefs, specifications, issues, documentation, interfaces, evidence summaries, and
release communications reference personas here by stable ID; they do not redefine an
audience inline.

A persona is an evidence-backed perspective, not a demographic stereotype or fictional
biography. It records goals, motivations, mental models, constraints, authority, risks,
and communication needs that materially affect product decisions. A real person may act
as several personas, and several people may share one persona.

AI collaborators are system actors, not personas. They must use the applicable human
personas to evaluate language, workflow, and outcomes rather than inventing an “AI user”
whose convenience overrides human needs.

## Persona Contract

Every managed project maintains a persona registry seeded during inception and extended
when a genuinely different audience appears. Each persona records:

| Field | Meaning |
|---|---|
| Stable ID and name | Durable reference used by governed artifacts |
| Status | `draft`, `confirmed`, `retired`, or `superseded` |
| Audience role | Primary, secondary, operator, approver, affected party, or stakeholder |
| Who | Plain description grounded in the project context |
| Motivations | Why they care and what drives their choices |
| Desired outcomes | Observable results they need, in their language |
| Perspective / mental model | How they understand the problem and product |
| Context and constraints | Environment, time, device, connectivity, expertise, accessibility, or organizational limits |
| Risks and concerns | What failure, harm, or loss looks like to them |
| Authority | Decisions they make, approve, influence, or cannot make |
| Information needs | What they need to act or trust the result |
| Communication preferences | Appropriate depth, vocabulary, format, and timing |
| Permissions mapping | System roles/capabilities where applicable; never a substitute for the persona |
| Evidence and owner | Sources, who maintains it, confirmation date, and review cadence |
| Anti-assumptions | Tempting stereotypes or unsupported claims the team must not infer |

Persona updates are human-owned decisions. Generated artifacts may propose changes from
research, support, issue, or usage evidence, but cannot silently rewrite this registry.
When a persona changes materially, impact analysis identifies affected specifications,
issues, controls, documentation, tests, and communications.

## PER-OWNER — Project owner / primary developer

**Status:** Confirmed from discovery  
**Audience role:** Primary user, product decision owner, and often initial operator

**Who:** A developer responsible for turning an idea or existing repository into a real,
maintainable project. They may work alone initially and may not be a specialist in every
security, regulatory, testing, documentation, or delivery discipline.

**Motivations:**

- Build the right system without having to remember every engineering obligation.
- Trust that AI work is safe, complete, explainable, and evidence-backed.
- Delegate a small or personal project without accepting lower code, security, testing,
  documentation, or user-experience quality.
- Preserve project knowledge so future people and AIs can continue without rediscovery.
- Scale from solo work to a team or regulated environment without replacing the workflow.

**Desired outcomes:**

- Move coherently from idea or existing code to planned, executable, verified work.
- Choose concise delegated delivery or detailed collaboration independently from the
  assurance the project requires.
- See what applies, what passed, what failed, what remains unknown, and what needs a human.
- Receive useful tool/upgrade options they did not know to ask for.
- Communicate progress and decisions to technical and nontechnical audiences.

**Perspective / mental model:** The kit is a trusted development operating system, not a
prompt library. “Easy” means the system coordinates complexity and exposes the right
decision at the right time.

**Context and constraints:** May be time-constrained, switching domains, using any major
desktop OS, online or offline, and simultaneously acting as developer, product owner,
reviewer, and operator. Holding several roles does not make implementation self-review a
distinct review pass or establish qualifications the owner does not have.

**Risks and concerns:** False confidence, leaked credentials, missing tests or obligations,
AI shortcuts, insecure code that merely appears functional, minimally implemented user
interfaces, process theater, excessive ceremony, lost decisions, unsafe upgrades, and
being locked into tools they cannot evaluate.

**Authority:** Approves product direction, ordinary implementation plans, tools, and
project risks unless policy requires independent or qualified approval. May review
requested-outcome alignment without claiming code-language or assurance expertise.

**Information and communication:** Lead with outcome, impact, choices, recommendation,
coverage, quality receipt, and required action. Keep implementation machinery behind
progressive detail unless the selected evidence presentation calls for the full package.

**Permissions mapping:** Repository owner/maintainer; may hold multiple operational roles.

**Evidence and owner:** Discovery decisions D1–D15, including the D1 sensitive-data
assurance amendment, and issue #23's professional-baseline clarification; maintained by
the project owner.

**Anti-assumptions:** Solo does not mean hobby, low risk, unregulated, technically novice,
authorized to self-approve every control, or unable to obtain a distinct review through a
separate capable AI context.

## PER-CONTRIBUTOR — Team developer / maintainer

**Status:** Confirmed from discovery  
**Audience role:** Primary user and implementer

**Who:** A developer joining or contributing to a managed repository, potentially with a
different workstation, operating system, specialty, or AI environment from the owner.

**Motivations:**

- Pick up Ready work quickly without reconstructing hidden conversation.
- Make changes that fit established architecture, policy, and product intent.
- Receive fast, specific feedback when work is incomplete or blocked.

**Desired outcomes:** Resolve the same policies and references as teammates; implement an
issue autonomously; produce a reviewable PR with required evidence; leave durable memory.

**Perspective / mental model:** The issue is the execution contract, the repository is the
source of project truth, and the lifecycle engine is the impartial verifier.

**Context and constraints:** May have limited historical context or permissions; may be
working across time zones, offline, or in a constrained corporate environment.

**Risks and concerns:** Stale issues, undocumented decisions, environment drift, surprise
requirements late in review, unsafe automation, and standards that differ by AI/session.

**Authority:** Implements Ready scope and proposes decisions; cannot invent unresolved
product/policy decisions or approve risks beyond assigned authority.

**Information and communication:** Precise scope, references, key files, acceptance,
tests, controls, evidence, dependencies, and explicit out-of-scope boundaries.

**Permissions mapping:** Contributor, maintainer, reviewer, or CODEOWNER as assigned.

**Evidence and owner:** D5, D7, and D8; reviewed when team workflow changes.

**Anti-assumptions:** A contributor is not guaranteed to know the stack, policy history,
organization vocabulary, or original conversation.

## PER-STAKEHOLDER — Nontechnical stakeholder

**Status:** Confirmed from discovery  
**Audience role:** Affected party, sponsor, requester, or decision influencer

**Who:** A person who needs to understand what is being built, why, current direction,
risk, and outcomes without navigating implementation details.

**Motivations:** Ensure the work serves real needs, uses time responsibly, and communicates
tradeoffs and risk honestly.

**Desired outcomes:** Understand issue summaries, roadmap direction, approved behavior,
release changes, unresolved decisions, and material risks in plain language.

**Perspective / mental model:** Features and outcomes matter more than repository
mechanics. The roadmap communicates intent; release summaries communicate delivered value.

**Context and constraints:** Limited time and technical vocabulary; may consume generated
documents, GitHub views, meetings, or release communication rather than source code.

**Risks and concerns:** Technical jargon hiding impact, stale roadmap documents, false
certainty, unexplained delay, and decisions made without affected people.

**Authority:** Varies by project—may request, approve product scope, accept business risk,
or only provide input. The project must record the mapping explicitly.

**Information and communication:** Plain language, audience-canonical terms, concise
summary first, observable outcomes, real uncertainty, and optional deeper links.

**Permissions mapping:** Usually read/comment; project-specific approval roles may apply.

**Evidence and owner:** D5, D8, and D12; validated through stakeholder feedback.

**Anti-assumptions:** Nontechnical does not mean uninformed, uninterested in risk, or
unable to make authoritative product decisions.

## PER-ASSURANCE — Security, compliance, risk, or quality reviewer

**Status:** Confirmed from discovery  
**Audience role:** Independent or qualified reviewer and control owner

**Who:** A specialist or assigned reviewer responsible for determining whether evidence
supports security, compliance, risk, accessibility, quality, or regulatory claims.

**Motivations:** Make defensible decisions from complete, traceable evidence; focus human
judgment where automation cannot establish sufficiency.

**Desired outcomes:** See applicability rationale, control versions/states, evidence,
coverage limits, exceptions, residual risks, approvals, invalidation, and history.

**Perspective / mental model:** Conformance is scoped evidence against versioned controls,
not a badge. Unknown, not applicable, failed, excepted, and passed are distinct states.

**Context and constraints:** May be external to the implementation team, bound by
professional/legal duties, reviewing many repositories, and prohibited from accepting
developer self-attestation.

**Risks and concerns:** False-green results, missing provenance, stale evidence,
unqualified legal claims, content exposed through an unverified route, user acknowledgment
misrepresented as authorization or assurance, concealed exceptions, inadequate
separation, and unreviewable AI reasoning.

**Authority:** Approves or rejects evidence and risks within assigned qualifications;
cannot waive prohibited controls, law, binding contracts, or duties outside their remit.

**Information and communication:** Stable control IDs, authoritative sources, exact
versions/scope, evidence chain, reproducible evaluation, material changes, and limitations.

**Permissions mapping:** Reviewer/auditor/risk approver; normally read evidence and approve
without unrestricted code or production mutation authority.

**Evidence and owner:** D2, D6, D7, and D12; refined with qualified regulatory reviewers.

**Anti-assumptions:** A security reviewer is not automatically qualified for legal,
privacy, accessibility, safety, or every regulatory determination.

## PER-ADMIN — Organization / platform administrator

**Status:** Confirmed as a scale audience  
**Audience role:** Operator, policy owner, and capability administrator

**Who:** A person responsible for GitHub organization settings, Codex/workspace policy,
tool distribution, credentials, policy mirrors, CI, or fleet-level governance.

**Motivations:** Give teams safe paved roads, apply organization requirements consistently,
and understand fleet risk without destroying repository autonomy or offline operation.

**Desired outcomes:** Publish trusted packs/plugins, set minimum versions and approval
requirements, manage access, monitor drift, revoke compromised versions, and recover.

**Perspective / mental model:** Organization policy is a strengthening layer over pinned
repository policy; fleet visibility does not replace local reproducibility.

**Context and constraints:** Manages many repositories and identities, change windows,
enterprise policy, limited tokens/permissions, and potentially air-gapped environments.

**Risks and concerns:** Supply-chain compromise, overbroad automation, fragmented versions,
unrecoverable rollout, silent policy weakening, excessive central dependency, and support
burden.

**Authority:** Manages organization capabilities and mandatory policy within governance;
cannot silently rewrite human-owned project decisions or falsify local conformance.

**Information and communication:** Compatibility matrix, semantic upgrade impact, fleet
coverage, drift, revocation, audit export, recovery, and actionable exceptions.

**Permissions mapping:** GitHub/workspace/org administrator and policy publisher roles,
separated where required.

**Evidence and owner:** D3, D7, roadmap Phase 8; refined during fleet design.

**Anti-assumptions:** Central administration does not imply continuous network access,
unlimited authority, or permission to inspect every sensitive evidence artifact.

## AI Actors

These are behaviors, not personas:

- **Guide:** conducts discovery, explains choices, and records approved human intent.
- **Planner:** converts approved intent into plans and Ready issues without inventing
  unresolved decisions.
- **Implementer:** executes Ready scope and stops on stale or missing authority.
- **Reviewer:** evaluates change and evidence in a distinct context within declared
  capability, with stronger independence or qualifications where effective policy
  requires them.
- **Maintainer:** detects drift, upgrades, stale risks, and documentation impact.

Every AI action names the human persona(s) served and the authority under which it acts.
