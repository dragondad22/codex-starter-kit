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

### Actionable work item

A work item whose objective, scope, authority, dependencies, evidence, and subtype-specific
completion contract are sufficient for an assigned human or AI to proceed without
inventing a new decision. It may be implementation, question, or research work.

### Accepted exception

A separately approved disposition allowing work to cross a specific gate despite an
underlying failed or incomplete control. It never changes that control to `pass`.

### Accepted residual risk

Risk inherent in a chosen architecture or operating model whose acceptance is reviewed
periodically rather than tied to a promised remediation date.

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

### Corrective exception

A time-limited accepted exception for a condition expected to be remediated.

### Decision record

A durable human-owned record of an approved consequential choice, its context,
alternatives or tradeoffs, consequences, owner, and history. It explains what governs
later work and why; it is more authoritative than the discussion that produced it.

### Effective policy

The deterministic result of compiling universal, project-type, triggered, organization,
repository, and approved-risk policy layers for a project.

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

### Feature

A user- or stakeholder-visible capability represented on the Horizon roadmap. A feature
may begin as lightweight intake and later be refined or decomposed into executable work.

### Promotion

The explicit transfer of a material issue result into the authoritative decision,
specification, policy, human-owned record, or structured state that governs its meaning.
The destination and issue retain reciprocal references.

### Question work item

A `type:question` issue for a consequential unresolved question whose answer must outlive
the current conversation, materially affects work, requires named authority or evidence,
or is likely to be referenced again. Its issue discussion is not authoritative.

### Handling authorization

Permission for a named actor to expose specified content to a named tool, service, or
environment for a bounded purpose. It is separate from content classification and from
evidence that the route provides suitable handling guarantees.

### Horizon

Feature roadmap intent in the GitHub Project: `Now`, `Next`, `Later`, or blank. It is not
execution Status or release membership.

### Issue

A GitHub work record describing an outcome, question, research effort, defect, or task.
An issue coordinates work and preserves history; only promoted decisions, specifications,
policy, human-owned records, or structured state become authoritative for their domains.

### Later (Horizon)

A plausible future feature without a release commitment. It remains visible for direction
and refinement but must not be presented as promised scope.

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

### Milestone

The single finite GitHub manifest for one named release. It identifies approved release
membership but does not by itself prove aggregate release readiness. Epics and Horizon
express different concepts and are not milestones.

### Next (Horizon)

An intentional feature candidate adjacent to or after current commitments. It is not yet
committed to a release. Do not confuse it with `Next` in execution Status, which means a
work item is queued to start.

### Now (Horizon)

A feature that is part of committed current product direction. It may span releases and
does not join a release until assigned to that release's approved Milestone.

### Persona

An evidence-backed human audience perspective with goals, motivations, constraints,
authority, risks, and communication needs. An AI actor is not a persona.

### Policy pack

An immutable signed/versioned bundle of focused standards, controls, templates, schemas,
routing metadata, and migrations.

### Product assurance

Evidence that an end-to-end product route—including the AI client, connected tools,
services, environment, authority, and data flow—provides required handling guarantees for
an explicit scope. User acknowledgment or tool availability is not product assurance.

### PRD (Product Requirements Document)

A human-owned product document explaining the problem, intended users and outcomes,
requirements, boundaries, success measures, and product-level testing decisions. It says
what the product must achieve, not every implementation detail.

### Pull request (PR)

A proposed set of repository changes submitted for checks, review, and merge. The PR links
the work item, governing records, changed artifacts, verification, evidence, deviations,
and follow-up work.

### Ready

The work-item readiness state indicating that objective, scope, decisions, references,
acceptance, evidence, dependencies, and authority satisfy the applicable subtype contract.

### Research record

A durable human-owned account of bounded research, including its questions or mapping
objective, intended use, method, sources, findings, conflicting evidence, uncertainty,
limitations, and freshness. It informs but does not silently establish a decision.

### Research work item

A `type:research` issue authorizing bounded evidence-producing work by a human or AI at a
declared depth or effort, with explicit stopping conditions, output, and review needs.

### Release

A governed publication, deployment, or delivery of an identified source and artifact set
to an audience or environment. Its mechanics and versioning depend on the selected
release adapter.

### Release candidate

A release scope and exact source/artifact set undergoing final aggregate verification.
New scope normally enters only to resolve blockers or required corrections.

### Release issue

The aggregate executable issue that owns one release's scope, exclusions, gates, evidence,
limitations, approvals, publication, rollback, communication, and completion.

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

### Roadmap

The ordered view of feature direction maintained through live Project issues and Horizon.
It is not a release manifest or a duplicate document listing current issue status.

### Semantic Versioning (SemVer)

The `MAJOR.MINOR.PATCH` version convention used when selected for a versioned contract.
Broadly, major versions communicate incompatible contract changes, minor versions add
backward-compatible capability, and patch versions make backward-compatible fixes.
Not every project output uses SemVer.

### S.M.A.R.T. Release

A release that is Scoped, Measurable, Approved, Releasable, and Triggered. Its trigger may
be outcome-, time-, event-, or hybrid-bound.

### Specification (spec)

A human-owned description of behavior or qualities to build and verify, tied to personas,
scenarios, requirements, constraints, acceptance criteria, and governing decisions. A PRD
describes product outcomes; specifications provide the more focused contract for delivery.

### Special-data-handling declaration

The v1 project-level answer to whether the project intentionally contains or processes
information requiring special confidentiality, privacy, contractual, or regulatory
handling: `No`, `Yes`, or `Unsure`. `Yes` and `Unsure` trigger a notice and truthful
coverage limits; no answer grants handling authority or establishes conformance.

### Status

Execution lifecycle on the GitHub Project: `Backlog`, `Next`, `In progress`, or `Done`.
Do not use Status to express roadmap intent or readiness.

### Task

One independently executable implementation or operational slice with explicit scope,
acceptance criteria, verification, dependencies, and completion evidence.
