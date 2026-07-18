# DEC-0022 — Scoped execution mandates and issue responsibility

**Status:** Accepted
**Owner:** dragondad22
**Date:** 2026-07-17
**Source decision:** Issue #73

## Context

The product promises delegated delivery without weaker engineering, but DEC-0019's broad
interrupt list and issue #73's exact-plan approvals turned ordinary external delivery and
recovery into repeated permission prompts. Content-addressed plans are valuable integrity
evidence; a new digest caused by refreshed observations or partial completion is not by
itself new authority. Issue bodies also began repeating standing workflow, causing a task
brief to silently redefine how an agent collaborated with its owner.

Issues have different valid maturity states. An Intake idea or design-first initiative
can preserve demand, context, breadcrumbs, known constraints, open decisions, and a
promotion gate without pretending to be executable. Only Ready implementation work needs
a decision-complete execution brief.

## Decision

Human approval attaches to a versioned **execution mandate**. A mandate identifies the
owner and approval record; immutable targets and credential identities; exact permission
and compatibility profiles; allowed semantic effects and exact resource-spec digests;
data, cost, and destructive ceilings; bounded-run marker context; expiry; and cleanup,
retention, and recovery ownership. A marker is descriptive evidence, not sufficient proof
of ownership. The lifecycle engine still emits an exact
content-addressed plan, but may apply, retry, reconcile, and clean up without another
checkpoint when it proves the plan is wholly contained by the active mandate.

Execution stops before effects when the mandate is missing, expired, stale, conflicting,
or would be exceeded by a changed target, actor, permission, effect class, data route,
cost, compatibility assumption, destructive scope, or unrecognized human-owned state.
A digest change caused only by current observation, completed effects, bounded retry, or
recovery is not a semantic expansion. Exact-plan approval remains a supported stronger
interaction preference and a historical evidence format, not the delegated default.

Collaborative engagement provides visibility and checkpoints at genuine decisions. It
does not mean confirmation for every in-mandate effect. Delegated and collaborative modes
share the same controls and stop conditions; they differ in communication cadence. This
narrows DEC-0019's blanket external-effect interrupt rule: externality is a fact recorded
in the mandate, while new or broader authority remains a mandatory interrupt.

Issue templates are lifecycle-specific:

- an Intake idea preserves opportunity and provenance;
- a design-first initiative preserves enough context to lead research, decisions, and
  decomposition while remaining non-executable;
- a Ready task supplies a verifiable execution brief; and
- question and research items preserve their subtype-specific resolution contracts.

Issues state outcomes, task-specific context, governing references, genuine open
decisions, boundaries, acceptance, and evidence. Bootstrap files, governing decisions,
and versioned templates define standing execution behavior. An issue may override that
behavior only explicitly, exceptionally, and with rationale.

## Consequences

Plans and receipts retain exact identities, but routine recovery no longer turns integrity
hashes into permission prompts. The engine records both mandate and plan identity so an
auditor can prove what authority covered each effect. A containment failure produces a
semantic explanation rather than asking an owner to approve an opaque replacement hash.

Issue validation becomes state-aware: completeness for Intake or Needs refinement is not
misrepresented as Ready, and missing implementation detail is not invented prematurely.
Native Readiness remains the execution gate.

Existing exact-plan receipts and approvals remain immutable v1 evidence. New mandate
authorization is versioned and prospective.

## Source

Approved by the product owner in the active Codex collaboration and recorded in
[#73](https://github.com/dragondad22/codex-starter-kit/issues/73#issuecomment-5009113729).
