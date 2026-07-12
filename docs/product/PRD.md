# Codex Starter Kit — Product Requirements

**Status:** Draft derived from approved discovery decisions  
**Source:** `docs/discovery/CODEX_STARTER_KIT_REVIEW.md`  
**Audience:** Product owner, implementers, reviewers, and stakeholders

## Problem Statement

Building responsibly with an AI requires a developer to coordinate product discovery,
architecture, implementation, tests, security, compliance, documentation, GitHub work,
release communication, and long-term maintenance. AI agents can take unsafe shortcuts,
omit negative paths, expose credentials, lose prior decisions, or report success without
evidence. Standards copied into repositories drift, while cloud-only standards fail
offline and cannot reproduce historical conformance.

The developer should not need to remember every obligation or repeatedly explain the
project to each AI. They need a guided system that turns an idea or existing repository
into managed, executable work and can prove what was checked, what applies, what remains
unresolved, and which risks were explicitly accepted.

## Solution

Build a Codex-native development system with a simple guided workflow backed by a
deterministic lifecycle engine. It scaffolds new projects, retrofits existing projects,
compiles versioned policy from project facts, prepares executable GitHub issues, manages
predictable repository growth, verifies evidence, communicates changes to technical and
nontechnical audiences, releases through project-appropriate adapters, and safely
upgrades managed repositories.

The product ships as three layers from its first usable release:

1. A Codex plugin containing progressively disclosed workflow skills.
2. A standalone cross-platform lifecycle engine used by Codex, developers, and CI.
3. A durable managed-repository contract containing pinned state, human records,
   provenance, routing, and evidence.

## Product Principles

- Easy means complexity hidden behind a coherent interface, not reduced rigor.
- No evidence means no pass.
- Decisions made in conversation become authoritative only when durably approved.
- Policy applicability is derived from recorded facts and versioned rules.
- Content classification, handling authorization, and product assurance are separate
  facts; acknowledgment cannot substitute for any of them.
- Human-owned records and generated views never silently overwrite one another.
- Every executable issue can be implemented without the originating conversation.
- New structure appears when a real capability requires it and follows explicit rules.
- Online convenience never removes offline reproducibility.
- Upgrades and tool installation are valid, explainable options—not silent side effects.
- Risks are owned, time-bounded or periodically reviewed, and never disguised as passes.

## User Stories

1. As a developer with an idea, I want a guided brief and inception flow so that the system understands the product before narrowing its design.
2. As a developer with an empty repository, I want the system to create a managed project so that I start with applicable controls and documentation.
3. As a developer with an existing repository, I want a read-only assessment and retrofit plan so that adopting the system does not destroy my work.
4. As a former Claude-kit user, I want my decisions and GitHub history migrated semantically so that I do not lose project memory.
5. As a developer, I want project facts recorded explicitly so that policy applicability is explainable.
6. As an owner of a project that may require special data handling, I want a short declaration and truthful capability limits so that ordinary work stays concise and absent assurance never looks like conformance.
7. As a developer, I want universal security and engineering controls enabled by default so that project size does not become an excuse for unsafe work.
8. As a policy owner, I want context-triggered controls so that irrelevant obligations do not obscure applicable ones.
9. As a reviewer, I want every control to have an explicit state and evidence so that green status is trustworthy.
10. As a risk owner, I want corrective exceptions and residual risks recorded so that unavoidable risk is visible and governed.
11. As an approver, I want prohibited exceptions enforced so that neither AI nor project owners can waive law or fabricate conformance.
12. As a solo developer, I want the workflow optimized for one person while showing missing separation of duties so that it can scale honestly.
13. As a team member, I want my AI to resolve the same policy versions and templates as everyone else so that behavior does not depend on a workstation.
14. As an offline developer, I want verified cached policy packs so that work remains reproducible without network access.
15. As an air-gapped team, I want mirrored or vendored signed packs so that restricted environments remain supportable.
16. As a developer, I want concise `AGENTS.md` routing and focused breadcrumbs so that Codex loads only relevant context.
17. As a maintainer, I want stable artifact IDs and generated routing links so that moving files does not create reference drift.
18. As a developer, I want predictable directory-role rules so that new project capabilities land in sensible locations.
19. As a maintainer of an unconventional repository, I want logical roles mapped to existing paths so that the kit respects valid conventions.
20. As a project owner, I want every work item in GitHub Issues and one synchronized Project so that the operational state is visible.
21. As a nontechnical stakeholder, I want a short issue summary so that I can understand the work without implementation details.
22. As an AI implementer, I want a complete execution brief so that I can implement a Ready issue without further decisions.
23. As a reviewer, I want stale issue references to invalidate readiness so that implementation does not follow obsolete decisions.
24. As a product owner, I want a Horizon roadmap over live feature issues so that ideas and direction do not drift into a separate document.
25. As a developer, I want GitHub Project state reconciled automatically so that closed work does not remain visibly in progress.
26. As a developer, I want protected issue-linked pull requests so that every material change has intent, scope, review, and evidence.
27. As a repository owner, I want release behavior selected for my output type so that infrastructure and documentation do not pretend to be packages.
28. As a user, I want release communication generated for my audience so that technical and nontechnical readers receive relevant information.
29. As a security reviewer, I want missing tooling to block at the earliest relevant risk boundary so that planning can continue without unsafe publication.
30. As a developer, I want recommended upgrades and tools explained so that I can benefit from capabilities I did not know existed.
31. As an administrator, I want tool recommendations to disclose permissions, data access, trust, cost, and fallback so that installation is informed.
32. As a CI operator, I want deterministic verification without an interactive AI so that conformance is reproducible.
33. As a release approver, I want a human conformance summary and machine evidence manifest so that release decisions are auditable.
34. As a project owner, I want plugin updates separated from repository upgrades so that an install cannot silently migrate my project.
35. As a maintainer, I want semantic policy and repository upgrade plans so that strengthened, weakened, or invalidated controls are visible.
36. As a developer, I want conflicts to stop for reconciliation so that generated state never silently defeats human edits.
37. As a Windows user, I want native support without mandatory WSL so that the universal workflow works on my platform.
38. As a macOS or Linux user, I want the same policy and evidence semantics so that OS differences do not alter conformance.
39. As an incident responder, I want a governed emergency path so that urgent work remains traceable and creates follow-up obligations.
40. As a future maintainer, I want historical packs, decisions, issues, and evidence linked so that I can reconstruct why the project changed.
41. As a project owner, I want consequential unresolved questions tracked without routine clarification flooding the board so that dependencies and answer authority stay visible.
42. As a decision maker, I want bounded research authorized by objective, depth, cost, and stopping conditions so that expensive evidence gathering is deliberate and reusable.
43. As a future maintainer, I want resolved questions and authoritative records to reference each other so that I can trace a conclusion without treating issue discussion as authority.
44. As a product owner, I want rolling Horizon intent separated from finite release membership so that current direction can evolve without making a named release endless.
45. As a release approver, I want one aggregate release issue with measurable scope, gates, evidence, trigger, and approval so that milestone completion cannot falsely imply readiness.

## Implementation Decisions

- The lifecycle engine is the primary external seam. Its interface is `create`,
  `retrofit`, `inspect`, `plan`, `apply`, `verify`, `status`, and `upgrade`.
- The Codex plugin is a guided adapter; it is not the sole enforcement authority.
- CI and direct developer use call the same engine interface.
- Structured state is authoritative for lifecycle facts and policy computation.
- V1 records whether special data handling is intentional as `No`, `Yes`, or `Unsure`;
  `Yes` and `Unsure` trigger a concise notice, acknowledgment, and explicit coverage
  limitations.
- Acknowledgment never authorizes tool access or transmission and never establishes
  sensitive-data or regulatory conformance.
- Human briefs, decisions, risks, specifications, and communications remain human-owned.
- Every managed project maintains one human-owned persona registry with stable IDs;
  governed artifacts reference personas rather than redefining audiences inline.
- Generated Markdown views carry source hashes and are reproducible projections.
- Policy is delivered through signed immutable packs pinned by ID, version, and digest.
- Effective policy layers universal, project-type, triggered, organization, repository,
  and approved-risk inputs without silent weakening.
- Logical directory roles map policy concepts to concrete paths.
- GitHub Issues and exactly one linked GitHub Project are mandatory.
- Ready issues use a human summary and a complete AI execution brief.
- Consequential questions and bounded research use subtype-specific Ready and completion
  contracts; ordinary clarification remains outside the Project.
- Material question answers are promoted with reciprocal issue/authority references, and
  research produces durable human-owned records without silently establishing decisions.
- Native GitHub Milestones are finite release manifests; Horizon remains rolling feature
  intent, and aggregate release issues own S.M.A.R.T. readiness and publication.
- The default delivery flow is Ready issue, issue branch, PR, gates, squash merge.
- Version and release adapters are selected from project outputs and policy.
- Linux, macOS, and Windows are native first-release targets.
- Claude-kit compatibility is semantic migration, not runtime-file compatibility.

## Testing Decisions

- Test observable behavior through the lifecycle engine interface.
- Use temporary real Git repositories and filesystem stand-ins at that seam.
- Use in-memory/fake adapters for GitHub, policy registry, signatures, clocks, process
  execution, and approval identity; contract-test production adapters separately.
- Golden tests cover generated human documents only where formatting is part of the
  interface; semantic assertions cover structured state.
- State-machine tests cover every allowed and rejected transition.
- Security tests cover path traversal, symlinks, malicious repositories, secret leakage,
  command injection, untrusted policy packs, signature failure, and authority escalation.
- Upgrade tests start at every supported prior schema/policy version and verify plans,
  conflicts, rollback, and evidence invalidation.
- Cross-platform CI runs native Windows, macOS, and Linux behavior.
- Routing tests prove relevant policy is discoverable without loading unrelated context.
- Audience tests verify issues, specs, documentation, interfaces, and release views serve
  their referenced personas without unsupported assumptions or internal-language leakage.
- Integration tests cover GitHub issue/project synchronization, idempotence, retries,
  partial failure, rate limits, and reconciliation.
- End-to-end tracer tests cover empty create, existing retrofit, Ready issue delivery,
  release evidence, offline verification, and managed upgrade.
- Sensitive-data boundary tests cover all three declaration values, concise ordinary
  flow, notice acknowledgment, absent-route `needs-review`/`unsupported` states, and no
  silent tool activation or transmission.

## Success Measures

- Zero false-pass states in conformance fixtures.
- A second AI can execute a Ready issue without additional product decisions.
- Create, retrofit, verify, and upgrade are idempotent or produce explicit conflicts.
- Every release result is reproducible from pinned engine/policy versions and source.
- Root routing stays within its declared context budget.
- Supported platform suites produce equivalent semantic results.
- GitHub Project drift is detected and reconciled without losing field assignments.
- Policy upgrades identify every changed obligation and invalidated evidence item.

## Out of Scope

- Supporting non-GitHub trackers in the first release.
- Preserving Claude-specific commands or permission files as runtime interfaces.
- Replacing legal, regulatory, security, accessibility, or domain experts where policy
  requires qualified human judgment.
- Detailed data taxonomies, DLP/egress enforcement, provider or environment certification,
  local/private AI runtimes, and comprehensive highly sensitive or regulated-content
  assurance in v1.
- Guaranteeing that all third-party tools or Codex clients are available offline.
- Silently installing tools, broadening permissions, or migrating repositories.
- A cloud-only conformance service as the sole authority.
- A hand-maintained roadmap document.

## Further Notes

This PRD intentionally specifies outcomes and interfaces before an implementation
language. Engine packaging, signing infrastructure, registry hosting, minimum OS/client
versions, verified sensitive-data routes, and regulatory packs require architecture
issues with evidence-backed tool selection and qualified review. The detailed
sensitive-data and AI/tool execution boundary is tracked as Later work in
[issue #21](https://github.com/dragondad22/codex-starter-kit/issues/21).
