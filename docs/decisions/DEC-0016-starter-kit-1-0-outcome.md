# DEC-0016 — Codex Starter Kit 1.0.0 outcome

**Status:** Accepted
**Owner:** dragondad22
**Date:** 2026-07-13
**Source decision:** Issue #23

## Context

DEC-0014 defines how a finite release is represented and evaluated, but deliberately did
not decide what the Codex Starter Kit must deliver before claiming its initial stable
supported contract. Creating a `1.0.0` Milestone without an approved outcome, evidence
contract, authority model, and limitation boundary would make a planning container appear
to be a release commitment.

This decision configures the Codex Starter Kit repository's own `1.0.0`. It does not make
all of its evidence depth, approvers, pilots, or signing choices mandatory for every
repository managed by the product.

## Decision

### Outcome and trigger

`1.0.0` requires the complete roadmap outcomes from Phase 0 through Phase 6:

1. repository and governance foundation;
2. native lifecycle-engine create and verify walking skeleton;
3. Codex plugin vertical slice;
4. GitHub executable-work system;
5. existing-repository retrofit and Claude-kit semantic migration;
6. signed policy distribution and managed-repository upgrade; and
7. release and stakeholder communication through representative adapters.

Phase 7 comprehensive sensitive-data and regulatory assurance and Phase 8 fleet
operations are excluded. The first-release special-data declaration, concise notice,
universal trust controls, and truthful `needs-review` or `unsupported` behavior remain
required; deferring comprehensive assurance does not permit a false pass.

The trigger is purely outcome-bound, with no deadline. The engine evaluates the complete
outcome, explicit dispositions and owners, Milestone work, evidence, blockers, and exact
source/artifact set. It recommends Release Candidate eligibility or reports gaps at the
aggregate release issue. The product owner must manually promote the release; no trigger
changes state or publishes automatically. Minor and patch releases are also manually
initiated and pass their applicable aggregate gates.

### Stable compatibility boundary

The stable observable contract includes lifecycle operation names and documented JSON
semantics; supported CLI and exit meanings; managed-repository schemas, ownership,
provenance, routes, and evidence; policy-pack identity, lock, control IDs, and upgrade;
documented plugin workflows and engine compatibility; GitHub hierarchy, Project fields,
readiness, and release semantics; and a supported migration path for persisted 1.x state.

Additive 1.x changes are allowed. A breaking observable or persisted-contract change
requires a compatible migration or the next major version. Internal Go packages, types,
and layout; exact prose formatting not separately declared contractual; and unlisted
platform or client versions are not stable public interfaces.

### Qualification evidence

Release evidence combines:

- an automated native fixture matrix covering application, library, infrastructure,
  data, and documentation repositories;
- three human-operated pilots outside this repository: greenfield guided create,
  existing-repository retrofit/GitHub work, and Claude-kit semantic migration;
- collective coverage of three representative release-adapter families, direct
  engine/CI use, offline policy resolution, solo and small-team handoff, all three
  special-data declarations, upgrade/reconciliation, and every supported native OS; and
- current security, compatibility, documentation, aggregate, and rollback evidence.

Release-blocking qualification uses immutable snapshots pinned by source revision and
digest. Changing real-world canary repositories supplement those snapshots for discovery
but are never the sole pass basis. A confirmed canary failure within the supported
contract blocks release or requires an approved truthful limitation; useful cases are
promoted into immutable snapshots. The exact canary repositories require a later bounded
selection with license, provenance, privacy, maintenance, and representativeness review.

Each GitHub pilot creates a unique private repository from its snapshot under a dedicated
test owner or organization. Every run-created Project, Milestone, issue, branch, pull
request, ruleset, release, and related resource carries the run identity, source digest,
owner, and expiry. Durable evidence and a cleanup manifest are exported before cleanup.
Successful runs delete and verify all run-owned resources; failed or cancelled runs use a
declared diagnostic retention period and scheduled cleanup backstop. Pre-existing,
unowned, source, and snapshot resources are never mutated or deleted. Cleanup failure is
an explicit non-pass. Qualification must exercise setup, success cleanup,
cancellation/failure retention, scheduled cleanup, idempotent rerun, insufficient cleanup
authority, partial deletion, and durable evidence survival.

### Artifact trust and authority

Every supported artifact has a SHA-256 manifest, retained provenance, GitHub artifact
attestation, and an offline-verifiable bundle with the required trusted roots. Public
Windows executables are Authenticode-signed. Public macOS executables are code-signed and
notarized. Linux artifacts carry the universal evidence plus signing required by their
approved package or release adapter. Unsigned or unnotarized artifacts may be used only
for clearly labeled, isolated pre-release qualification and cannot satisfy the public
stable-release gate.

The product owner approves scope, membership and dispositions, permitted limitations and
risk, Release Candidate promotion, and publication. A different human independently
reviews aggregate evidence, compatibility, limitations, unresolved risks, and exact
source/artifact identity before stable publication. Qualified assurance review is added
only where a claim or control requires it; nobody self-approves evidence where independent
review is required. Signing authority is a separate bounded operational capability and
grants no scope, risk, or publication authority.

### Membership and candidate identity

The operating rule is **automatically discover; explicitly admit**. Work merged since the
previous release enters proposed scope automatically. It must then be admitted, classified
as present but internal/non-user-visible, isolated or reverted, excluded with truthful
disclosure, or treated as a blocker. Source presence is a fact independent of Milestone
membership.

Before Release Candidate status, committed work needs an owner, rationale, acceptance and
evidence expectations, downstream-impact context, and product-owner approval. Removal
needs a recorded reason, impact review, and truthful deferred, excluded, isolated, or
reverted disposition; it cannot preserve an outcome claim that is no longer satisfied.

The accountable release-membership owner is distinct from the implementation assignee,
which may remain unset until executable work is selected. An approved feature outcome may
be committed while its Project Readiness is `Needs refinement` only when the aggregate
release issue records that owner, rationale, outcome-level acceptance/evidence, and impact;
the missing decomposition remains an explicit Release Candidate blocker. Without those
facts the work remains proposed rather than committed.

A Release Candidate identifies one exact source revision, dependency and policy
resolution, configuration, and artifact set. Any change invalidates that candidate and
requires a newly identified candidate, refreshed affected evidence, reconciliation, and
renewed approvals. A published manifest and approval record are immutable. Corrections
use a subsequent release or an explicit withdrawal or revocation record.

### Limitations

The Phase 7 and 8 exclusions, GitHub-only collaboration, semantic-only Claude-kit
migration, exact published support matrix, approved adapter families, third-party/offline
availability limits, and the need for qualified professional judgment are acceptable when
communicated at the relevant installation, support, verification, and release surfaces.

Requiring users to install Go; an unavailable or unverified plugin; unconfigured universal
secret protection; unversioned recovery provenance; missing signed/offline policy
verification; no compatible repository upgrade; a false pass; missing native evidence;
unbounded release or pilot effects; or a breaking 1.x change without migration are release
blockers. A limitation may narrow a claim only through explicit approved disposition and
downstream-impact communication. It cannot silently remove a required outcome.

## Consequences

The `1.0.0` Milestone may now be created and populated with approved Phase 0–6 work. Its
aggregate release issue, not Milestone percentage, reports readiness. Current completion
of Phases 0–1 is partial product progress, not Release Candidate eligibility; missing
Phase 2–6 capabilities and qualification evidence remain explicit gaps.

The release engine must distinguish general truthful lifecycle invariants from repository
configuration. Other managed repositories may choose different triggers, approvers,
evidence depth, signing adapters, and permitted limitations without weakening universal
no-false-pass, exact-candidate, explicit-disposition, and immutable-history behavior.

## Source

Approved through [issue #23](https://github.com/dragondad22/codex-starter-kit/issues/23).
