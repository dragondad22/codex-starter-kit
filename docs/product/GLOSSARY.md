# Codex Starter Kit — Glossary

**Status:** Initial canonical vocabulary  
**Authority:** Human-owned domain language

Use these terms in issues, specifications, policy, tests, interfaces, and documentation.
Do not introduce a synonym for a governed term without updating this reference.

## Terms

### Acceptance criteria

Observable conditions used to determine whether work or a specification has been
completed correctly. Good criteria describe externally verifiable outcomes and important
negative paths rather than only listing implementation steps.

### Accepted exception

A separately approved disposition allowing work to cross a specific gate despite an
underlying failed or incomplete control. It never changes that control to `pass`.

### Accepted residual risk

Risk inherent in a chosen architecture or operating model whose acceptance is reviewed
periodically rather than tied to a promised remediation date.

### Actionable work item

A work item whose objective, scope, authority, dependencies, evidence, and subtype-specific
completion contract are sufficient for an assigned human or AI to proceed without
inventing a new decision. It may be implementation, question, or research work.

### Adapter

A concrete implementation that connects an external dependency or product surface to a
module interface. Examples include GitHub, filesystem, registry, and Codex adapters.

### Breadcrumb

A stable, validated reference that lets a human or AI load additional governed material
only when relevant.

### Bug

A verified defect where observed behavior is incorrect, unsafe, inaccessible, or
misleading relative to an applicable requirement, specification, or reasonable supported
expectation.

### Canary repository

A changing real-world repository used to discover new product conditions and regressions.
It supplements repeatable qualification evidence but cannot be the sole basis for a
release pass. A confirmed supported-contract failure remains blocking or requires a
truthful approved limitation.

### Capability handshake

A non-mutating collection and evaluation of host, plugin, engine, repository, baseline,
native-environment, approval, and authority facts for one requested workflow. It retains
unknown and conflicting facts and selects a workflow capability mode; it does not install,
upgrade, grant authority, or establish conformance.

### Change record

A validated, human-owned structured description of one material change, including its
category, audiences or internal-only disposition, affected product components, breaking
state, and durable issue or pull-request references. Generated changelogs and release
summaries derive from it.

### Conformance

An evidence-backed result for an explicit scope, policy version, source revision, and
lifecycle gate. Avoid the unqualified term “compliant.”

### Content classification

A recorded statement about the confidentiality, privacy, contractual, regulatory, or
other handling needs of information. It does not by itself authorize an actor or tool to
handle the content and does not prove that a product route satisfies those needs.

### Control

A versioned requirement with applicability, evaluation, enforcement, exception, evidence,
invalidation, and routing rules.

### Conversational capture

The agent behavior of noticing when discussion produces durable untracked work, a
consequential question, risk, dependency, or decision; searching existing GitHub work;
and offering the owner the appropriate issue or authoritative-record action at a natural
checkpoint. It does not silently create work, approve decisions, or track ordinary
clarification.

### Corrective exception

A time-limited accepted exception for a condition expected to be remediated.

### Decision record

A durable human-owned record of an approved consequential choice, its context,
alternatives or tradeoffs, consequences, owner, and history. It explains what governs
later work and why; it is more authoritative than the discussion that produced it.

### Distinct review pass

An evidence-producing PR evaluation by a declared capable human or AI reviewer that did
not implement the change in the same working context. Automated checks and implementer
self-review support it but do not fulfill it; effective policy may require stronger human
independence or qualifications.

### Effective policy

The deterministic result of compiling universal, project-type, triggered, organization,
repository, and approved-risk policy layers for a project.

### Engagement mode

The configured degree and timing of human participation: `delegated` execution within
established authority and approved defaults, or `collaborative` checkpoints at planning
and material decisions or effects. It changes interaction, not the professional
engineering baseline or the truth of control results.

### Epic

A parent work item grouping multiple features or tasks toward a larger outcome. An epic
normally becomes complete through its sub-issues rather than acting as one large
implementation task.

### Evidence

Versioned, attributable information sufficient to support a control result. Logs or
claims are not automatically sufficient evidence.

### Executable issue

A Ready GitHub issue containing a human summary and complete implementation brief that an
authorized AI or developer can execute without new product or policy decisions.

### Execution mandate

A versioned human-owned authorization envelope naming immutable targets, exact credential
identities and permissions, exact governed resource specifications, permitted semantic
effects, data/cost/compatibility/destructive ceilings, expiry, and cleanup/recovery
ownership. Content-addressed plans may execute without another prompt only when the
lifecycle engine proves they are wholly contained by the mandate; a marker alone does not
establish ownership.

### Feature

A user- or stakeholder-visible capability represented on the Horizon roadmap. A feature
may begin as lightweight intake and later be refined or decomposed into executable work.

### Handling authorization

Permission for a named actor to expose specified content to a named tool, service, or
environment for a bounded purpose. It is separate from content classification and from
evidence that the route provides suitable handling guarantees.

### Horizon

Feature roadmap intent in the GitHub Project: `Now`, `Next`, `Later`, or blank. It is not
execution Status or release membership. Its values are:

- **Now:** The feature is part of committed current product direction. It may span
  releases and does not join a release until assigned to that release's approved
  Milestone.
- **Next:** The feature is an intentional candidate adjacent to or after current
  commitments. It is not yet committed to a release. Do not confuse Horizon `Next` with
  Status `Next`, which means a work item is selected as the immediate execution queue.
- **Later:** The feature is plausible future direction without a release commitment. It
  remains visible for direction and refinement but must not be presented as promised
  scope.
- **Blank:** The item is not independently placed on the feature roadmap. Tasks normally
  inherit roadmap context from their parent and leave Horizon blank.

### Issue

A GitHub work record describing an outcome, question, research effort, defect, or task.
An issue coordinates work and preserves history; only promoted decisions, specifications,
policy, human-owned records, or structured state become authoritative for their domains.

### Lifecycle engine

The deterministic authority that implements `create`, `retrofit`, `inspect`, `plan`,
`apply`, `verify`, `status`, and `upgrade`.

### Logical directory role

A stable semantic home such as source, integration tests, decisions, or evidence that a
project maps to an appropriate physical path.

### Managed repository

A repository with a valid local contract for pinned state, policy, provenance, routing,
human records, evidence, and GitHub synchronization. It does not imply every control
passes.

### Managed-work qualification

A content-addressed pre-work provenance record binding an executable issue contract,
governed sources, operating profile, current observation, Project configuration, and
immutable target. It does not authorize external effects; a DEC-0022 execution mandate is
separate.

### Milestone

The single finite GitHub manifest for one named release. It identifies approved release
membership but does not by itself prove aggregate release readiness. Epics and Horizon
express different concepts and are not milestones.

### Operating profile

A versioned governed configuration of engagement mode, scoped assurance additions, and
evidence presentation. The default is `delegated` engagement, no discretionary assurance
addition beyond effective policy, and a `concise` evidence view. It may add controls,
approvals, and detail above the universal professional engineering baseline; it is not a
code-quality tier. Retention remains an effective-policy and handling requirement rather
than a presentation preference. Profile changes are prospective and never rewrite prior
evidence or claims.

### Persona

An evidence-backed human audience perspective with goals, motivations, constraints,
authority, risks, and communication needs. An AI actor is not a persona.

### Phase

An ordered roadmap outcome used to sequence product capability and its evidence. Phase
membership must be explicit, either directly on cross-cutting work or derived from a
native parent assigned to the phase. The governed values are `Phase 0` through `Phase 8`.
Roadmap features carry the direct Project assignment; ordinary children leave the field
blank and expose parent-derived context. A cross-cutting direct assignment requires a
durable reason. A phase is not a sprint, Horizon value, execution state, or release
Milestone; completing one does not publish a release.

### Policy pack

An immutable signed/versioned bundle of focused standards, controls, templates, schemas,
routing metadata, and migrations.

### PRD (Product Requirements Document)

A human-owned product document explaining the problem, intended users and outcomes,
requirements, boundaries, success measures, and product-level testing decisions. It says
what the product must achieve, not every implementation detail.

### Pre-work freshness disposition

The single result of checking a Ready item against current authoritative facts at
selection/start or after material change: `fresh`, `mechanical-drift-repaired`,
`contained-context-refreshed`, `needs-refinement`, `already-delivered`, or `blocked`.
Age alone is not freshness.

### Product assurance

Evidence that an end-to-end product route—including the AI client, connected tools,
services, environment, authority, and data flow—provides required handling guarantees for
an explicit scope. User acknowledgment or tool availability is not product assurance.

### Product version

The SemVer identity shared by same-release Codex Starter Kit distribution surfaces. It is
distinct from lifecycle protocol, JSON schema, policy-pack, baseline, managed-state, and
other compatibility versions.

### Professional engineering baseline

The universal applicability-aware quality bar for every supported deliverable, regardless
of project size, audience, or engagement mode. It includes relevant coding and external
standards, security, complete user experience, testing, documentation, maintainability,
and acceptance verification. Applicability may differ, but a lower-quality passing mode
does not exist.

### Promotion

The explicit transfer of a material issue result into the authoritative decision,
specification, policy, human-owned record, or structured state that governs its meaning.
The destination and issue retain reciprocal references.

### Pull request (PR)

A proposed set of repository changes submitted for checks, review, and merge. The PR links
the work item, governing records, changed artifacts, verification, evidence, deviations,
and follow-up work.

### Qualification snapshot

An immutable representative repository state pinned by source revision and digest and
used to produce repeatable release-blocking evidence. Unlike a canary repository, it does
not change during qualification.

### Quality receipt

A concise human view of what was requested and delivered, which professional-baseline
checks applied, what evidence supports the result, and every limitation or non-pass. It
routes to deeper evidence and does not replace it.

### Question work item

A `type:question` issue for a consequential unresolved question whose answer must outlive
the current conversation, materially affects work, requires named authority or evidence,
or is likely to be referenced again. Its issue discussion is not authoritative.

### Readiness

Executability on the GitHub Project: `Intake`, `Needs refinement`, `Ready`, or `Blocked`.
Readiness answers whether authorized work can start now. Its states are:

- **Intake:** The item has been captured but has not passed refinement. It is not
  executable.
- **Needs refinement:** The item has been promoted for clarification, but its objective,
  scope, authority, decisions, references, acceptance, evidence, or dependencies are not
  yet sufficient for execution.
- **Ready:** The item satisfies its subtype-specific execution contract, has the required
  authority, and has no unresolved blocker. Ready work may remain Status `Backlog` until
  deliberately selected.
- **Blocked:** A known unresolved dependency, control, or required human action prevents
  otherwise planned work from starting or continuing. When the final blocker resolves,
  re-evaluate the item and move it to Ready if no other readiness gap remains.

### Release

A governed publication, deployment, or delivery of an identified source and artifact set
to an audience or environment. Its mechanics and versioning depend on the selected
release adapter.

### Release candidate

A release scope and one exact source, dependency, policy, configuration, and artifact set
undergoing final aggregate verification. Any change invalidates that candidate and
requires a newly identified candidate with refreshed affected evidence and approvals.

### Release issue

The aggregate executable issue that owns one release's scope, exclusions, gates, evidence,
limitations, approvals, publication, rollback, communication, and completion.

### Release membership disposition

The explicit treatment of discovered work for a named release: admitted, present but
internal or non-user-visible, isolated or reverted, excluded with truthful disclosure, or
blocking. Source presence and Milestone membership remain distinct facts.

### Release preparation

The local transaction that archives admitted change records, synchronizes product version
surfaces, and generates dated communication while explicitly remaining unpublished. It
does not create or approve a tag, GitHub Release, artifact, deployment, or release claim.

### Release readiness

The evidence-backed determination that committed release scope is completed or validly
dispositioned and all applicable aggregate gates and approvals permit release. Milestone
percentage alone is not release readiness.

### Release target

A named finite release to which work has been explicitly committed through its GitHub
Milestone. It is separate from Horizon.

### Release trigger

The declared outcome, time, event, or hybrid condition that initiates final release
evaluation. Satisfying the trigger never overrides a failed or prohibited release gate.

### Research record

A durable human-owned account of bounded research, including its questions or mapping
objective, intended use, method, sources, findings, conflicting evidence, uncertainty,
limitations, and freshness. It informs but does not silently establish a decision.

### Research work item

A `type:research` issue authorizing bounded evidence-producing work by a human or AI at a
declared depth or effort, with explicit stopping conditions, output, and review needs.

### Roadmap

The ordered view of feature direction maintained through live Project issues and Horizon.
It is not a release manifest or a duplicate document listing current issue status.

### S.M.A.R.T. Release

A release that is Scoped, Measurable, Approved, Releasable, and Triggered. Its trigger may
be outcome-, time-, event-, or hybrid-bound.

### Semantic Versioning (SemVer)

The `MAJOR.MINOR.PATCH` version convention used when selected for a versioned contract.
Broadly, major versions communicate incompatible contract changes, minor versions add
backward-compatible capability, and patch versions make backward-compatible fixes.
Not every project output uses SemVer.

### Special-data-handling declaration

The v1 project-level answer to whether the project intentionally contains or processes
information requiring special confidentiality, privacy, contractual, or regulatory
handling: `No`, `Yes`, or `Unsure`. `Yes` and `Unsure` trigger a notice and truthful
coverage limits; no answer grants handling authority or establishes conformance.

### Specification (spec)

A human-owned description of behavior or qualities to build and verify, tied to personas,
scenarios, requirements, constraints, acceptance criteria, and governing decisions. A PRD
describes product outcomes; specifications provide the more focused contract for delivery.

### Status

Execution lifecycle on the GitHub Project: `Backlog`, `Next`, `In progress`, or `Done`.
Do not use Status to express roadmap intent or readiness. Its states are:

- **Backlog:** The item is tracked but not selected for immediate execution. Backlog does
  not mean Horizon `Later` and may contain Ready work.
- **Next:** The item is deliberately selected as the immediate execution queue but has not
  started. Starting it still requires Readiness `Ready`.
- **In progress:** Execution has started. An incomplete parent also remains In progress
  once any child delivery starts or completes.
- **Done:** The completion contract is satisfied and the item is closed. A parent becomes
  Done when every child is complete unless genuinely outstanding acceptance work is
  represented by a concrete attached child.

### Task

An actionable implementation or operational outcome with sufficient context to begin.
Its implementation may be decomposed organically into tasks, subtasks, and steps. A
separate native issue is used when durable tracking adds value, not merely because a step
could be completed independently.

### Workflow capability mode

The evidence-backed boundary of what the Codex plugin may do for one requested workflow.
Its values are `full`, `degraded-guidance`, `verification-only`, and `unsupported`.
`full` permits the supported engine-backed workflow when every required capability and
authority is established; `degraded-guidance` permits only guidance and fallback when
known facts establish that narrower safe boundary; `verification-only` permits supported
read-only status and verification but no mutation; and `unsupported` stops when required
guarantees are unknown, conflicting, failed, or unavailable. A workflow capability mode
is not a conformance result.
