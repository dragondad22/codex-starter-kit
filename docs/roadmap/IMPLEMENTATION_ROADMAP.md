# Codex Starter Kit — Implementation Roadmap

**Status:** Draft sequencing proposal  
**Roadmap authority:** Live features and ordering are maintained in the
[Codex Starter Kit GitHub Project](https://github.com/users/dragondad22/projects/8).
This document is sequencing/reference architecture, not a hand-maintained status tracker.

## Sequencing Principle

Build tracer-bullet vertical slices through the lifecycle engine seam. Each phase
must be usable and testable end to end; avoid building all policy, all templates, or all
adapters horizontally before proving the contract.

## Phase 0 — Repository and Governance Foundation

Outcome: this kit repository follows the rules it will ship.

- Initialize Git and GitHub repository.
- Create the required GitHub Project, Status/Horizon/Readiness fields, views, rulesets,
  labels, issue/PR templates, and reconciliation baseline.
- Add concise root `AGENTS.md` and document routing.
- Promote approved discovery decisions into stable decision records.
- Select license, ownership, security policy, contribution policy, and release approach.
- Establish native Windows/macOS/Linux CI and dependency/provenance controls.
- Convert this roadmap into executable GitHub epics/features/tasks.

## Phase 1 — Walking Skeleton: Create and Verify

Outcome: one engine command creates a minimal managed repository and verifies it on all
three operating systems.

- Define schemas for project facts, policy lock, layout, managed files, plans, results,
  routes, and evidence.
- Select engine language/package/signing approach through a documented tool evaluation.
- Implement `inspect`, `plan`, `apply`, `status`, and `verify` for local filesystem/Git.
- Implement `core-trust` seed controls: truthful results, secrets, artifact ownership,
  coverage, recovery, and breadcrumbs.
- Render minimal `AGENTS.md`, brief, decision index, and conformance summary.
- Seed and confirm the project persona registry; route generated issues, specifications,
  documentation, and summaries through stable persona IDs.
- Prove idempotence, plan preconditions, conflict safety, rollback, malicious paths, and
  cross-platform semantic equivalence.

## Phase 2 — Codex Plugin Vertical Slice

Outcome: a user installs the plugin and completes guided create using the same engine.

- Scaffold/validate plugin manifest and marketplace development flow.
- Implement focused create/status/verify skills with progressive disclosure.
- Bundle the baseline pack for first-run offline use.
- Add capability detection and full/degraded/verification-only/unsupported modes.
- Evaluate skill routing, approval handoffs, context budget, and truthful failure states.
- Publish install/update documentation without coupling repository migration to plugin
  update.

## Phase 3 — GitHub Executable Work System

Outcome: managed repositories have synchronized Issues/Project and one Ready issue can be
executed end to end.

- Implement GitHub adapter and in-memory contract double.
- Bootstrap/reconcile labels, fields, views, rulesets, auto-add, and close-to-Done flow.
- Render/validate two-layer issues and readiness refresh.
- Implement sparse `type:question` and bounded `type:research` forms, subtype readiness,
  authorization, completion/promotion checks, and reciprocal source links.
- Add durable research-record routing and validate provenance, uncertainty, limitations,
  freshness, depth or effort, and stopping conditions.
- Implement branch/PR/gate/squash completion memory.
- Implement Horizon roadmap intake/promotion and Project drift reconciliation.
- Test partial failure, rate limits, field-option migration snapshots, and offline queues.

## Phase 4 — Retrofit and Claude Migration

Outcome: existing repositories receive a read-only assessment and safe semantic migration.

- Implement project/tool/layout discovery with confidence and provenance.
- Map logical directory roles and produce conformance/layout plans.
- Classify artifacts and preserve user ownership/history.
- Import known Claude-kit versions and record semantic mappings.
- Create findings/issues for existing nonconformance without claiming prior compliance.
- Test messy, partial, symlinked, monorepo, documentation, infrastructure, and data
  fixtures.

## Phase 5 — Signed Policy Distribution and Upgrade

Outcome: team members and CI resolve identical policy online/offline and safely upgrade.

- Specify pack schema, signatures, publisher trust, registry, immutable cache, and lock.
- Implement dependency/applicability compilation and effective-policy index.
- Implement mirror/vendor/air-gap workflows.
- Implement semantic pack diff, migration plan, evidence invalidation, rollback, and
  retained history.
- Add organization and repository policy layering with D2 enforcement.
- Exercise revoked, incompatible, tampered, missing, offline, and historical-pack cases.

## Phase 6 — Release and Communication

Outcome: at least three different project outputs release through one contract.

- Implement change records and audience-specific generated communication.
- Implement finite release Milestones, aggregate release issues, S.M.A.R.T. scope and
  trigger validation, scope-change memory, and release-readiness views.
- Implement application/service, library/package, and documentation/infrastructure
  release adapters as representative variants.
- Produce GitHub Release/tag, version transaction, provenance, evidence, and Project/
  milestone synchronization.
- Add signing/SBOM/attestation adapter triggers.
- Test partial publication, rollback, rerun, and emergency release paths.

## Phase 7 — Verified Sensitive-Data and Regulatory Coverage

Outcome: selected sensitive-data routes and named regulatory packs meet a later approved
release definition with qualified review and representative evidence fixtures.

- Implement the detailed sensitive-data and AI/tool execution assurance boundary from
  issue #21, including verified routes and truthful unsupported behavior.
- Select initial jurisdictions/industries from intended real projects.
- Establish authoritative-source monitoring and expert-review workflow.
- Build reusable data/privacy/access/audit/evidence controls first.
- Publish signed regulatory packs with applicability, limitations, and review cadence.
- Test applicable, not-applicable, uncertain, exception, independent-approval, and
  retention scenarios.
- Run pilot retrofits and releases with real representative repositories before claiming
  support.

## Phase 8 — Scale and Fleet Operations

Outcome: teams can govern many repositories without losing local reproducibility.

- Organization policy packs and minimum-version rules.
- Fleet inventory, update visibility, drift/reconciliation scheduling, and audit export.
- Shared cache/mirror support and workspace plugin distribution.
- Performance/context budgets and large monorepo behavior.
- Disaster recovery, registry outage, key rotation/revocation, and long-term evidence
  retention exercises.

## Cross-Cutting Definition of Done

Every phase requires:

- executable GitHub issues with readiness validated;
- tests at the lifecycle engine seam and adapter contracts;
- native Linux/macOS/Windows CI evidence;
- threat and failure-mode updates;
- user/developer/stakeholder documentation impact;
- policy/control IDs and breadcrumb validation;
- upgrade/migration implications;
- human conformance summary and machine evidence manifest;
- no unresolved prohibited exception or false-pass state.

## Immediate Investigation Issues

These must be resolved before Phase 1 implementation choices:

1. Engine implementation/package/signing evaluation across native target platforms.
2. Schema and content-addressed state format evaluation.
3. GitHub authentication/app/CLI adapter authority and rate-limit strategy.
4. Plugin distribution/publication and minimum Codex capability matrix.
5. Policy signing, trusted publisher, registry, mirror, and revocation design.
6. Later sensitive-data route assurance and initial regulatory coverage plan.
