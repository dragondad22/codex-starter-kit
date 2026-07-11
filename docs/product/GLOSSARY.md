# Codex Starter Kit — Glossary

**Status:** Initial canonical vocabulary  
**Authority:** Human-owned domain language

Use these terms in issues, specifications, policy, tests, interfaces, and documentation.
Do not introduce a synonym for a governed term without updating this reference.

## Terms

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

### Conformance

An evidence-backed result for an explicit scope, policy version, source revision, and
lifecycle gate. Avoid the unqualified term “compliant.”

### Control

A versioned requirement with applicability, evaluation, enforcement, exception, evidence,
invalidation, and routing rules.

### Corrective exception

A time-limited accepted exception for a condition expected to be remediated.

### Effective policy

The deterministic result of compiling universal, project-type, triggered, organization,
repository, and approved-risk policy layers for a project.

### Evidence

Versioned, attributable information sufficient to support a control result. Logs or
claims are not automatically sufficient evidence.

### Executable issue

A Ready GitHub issue containing a human summary and complete implementation brief that an
authorized AI or developer can execute without new product or policy decisions.

### Horizon

Feature roadmap intent in the GitHub Project: `Now`, `Next`, `Later`, or blank. It is not
execution Status.

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

### Persona

An evidence-backed human audience perspective with goals, motivations, constraints,
authority, risks, and communication needs. An AI actor is not a persona.

### Policy pack

An immutable signed/versioned bundle of focused standards, controls, templates, schemas,
routing metadata, and migrations.

### Ready

The issue readiness state indicating that scope, decisions, references, acceptance,
tests, evidence, dependencies, and authority are sufficient for implementation.

### Status

Execution lifecycle on the GitHub Project: `Backlog`, `Next`, `In progress`, or `Done`.
Do not use Status to express roadmap intent or readiness.
