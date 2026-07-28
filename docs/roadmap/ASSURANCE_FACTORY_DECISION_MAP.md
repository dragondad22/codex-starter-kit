# Assurance Factory Decision Map

**Scope:** Design-first feature
[#83](https://github.com/dragondad22/codex-starter-kit/issues/83)

**Status:** Active planning record; not product, architecture, policy, or Project
authority

**Inventory:** [Current seams and missing contracts](../research/ASSURANCE_FACTORY_SEAM_INVENTORY.md)

**Finish line:** Enough promoted decisions and evidence to approve, narrow, defer, or
reject a finite graph of executable vertical slices without implementation-time invention

The live Project and native issue relationships govern lifecycle state. This map preserves
the compact reasoning sequence. Downstream answers may add, merge, or remove tickets as
the frontier advances.

## #1: What authority and lifecycle does a validation contract have?

Blocked by: None

Type: Grilling

GitHub work item:
[#100](https://github.com/dragondad22/codex-starter-kit/issues/100)

### Question

What is the validation contract, what authority does it have, and how does its derivation,
identity, invalidation, correction, and evidence linkage remain reproducible and
human-correctable without becoming a competing source of truth?

### Answer

Resolved on 2026-07-28 and promoted as
[DEC-0023](../decisions/DEC-0023-validation-manifest-authority-and-lifecycle.md). A
**validation manifest** is a generated immutable checklist for one Ready work scope. It
binds exact authoritative inputs and derived assertion/evidence obligations without
creating requirements, recording run results, approving work, or authorizing effects.

Its canonical contents determine its identity. Changed inputs produce a new manifest and
make the prior one stale; immutable historical receipts remain reconstructable, while
evidence reuse requires an unchanged assertion fingerprint plus method, freshness, and
scope validity. The lifecycle engine reassesses at compile or inspect, plan, immediately
before evaluation, verify, and status.

A validation approval rule is a governed human choice expressed through standing policy
or a bounded per-work decision. It may depend on work, actor, risk, timing, cost,
environment, and effect scope, but never replaces DEC-0022 effect authority. Issue #101
now owns evidence-method classification; issue #108 still owns executable decomposition.

## #2: Which assertions belong to which evidence methods?

Blocked by: #1

Type: Research

GitHub work item:
[#101](https://github.com/dragondad22/codex-starter-kit/issues/101)

### Question

Which classes of predetermined outcome assertion should be evaluated through automated
tests, static controls, black-box behavior, specialist review, human attestation, or
explicit non-pass states?

### Answer

Open. DEC-0023 now fixes the validation-manifest authority boundary. Research may begin
when #101 is explicitly selected; its answer must preserve DEC-0023's immutable manifest,
assertion handoff, explicit non-pass, and evidence-reuse rules.

## #3: Is the validation-only path clear enough for executable decomposition?

Blocked by: #1, #2

Type: Grilling

GitHub work item:
[#108](https://github.com/dragondad22/codex-starter-kit/issues/108)

### Question

Are tickets #1 and #2 sufficient to decompose a useful validation-manifest vertical slice
for the existing single-writer workflow, and what exact slice should be approved,
deferred, or rejected?

### Answer

Open. A yes answer may publish validation-manifest implementation children without
waiting for actor, finding, orchestration, provider, or Mission Control decisions.

## #4: What actor capability and least-knowledge contract is required?

Blocked by: #1, #2

Type: Grilling

GitHub work item:
[#102](https://github.com/dragondad22/codex-starter-kit/issues/102)

### Question

What provider-neutral actor contract supplies enough context and authority for capable
implementation or evaluation while enforcing one writer, least knowledge, evaluator
isolation, explicit limitations, and evidence-producing results?

### Answer

Open. It extends DEC-0020's distinct-review minimum without approving specialist kits or
claiming that a role label proves capability.

## #5: How are candidate findings disposed and disagreements stopped?

Blocked by: #2, #4

Type: Grilling

GitHub work item:
[#103](https://github.com/dragondad22/codex-starter-kit/issues/103)

### Question

What finding-disposition and arbitration contract preserves evidence and correction speed
while preventing silent loss, tracker spam, false authority, and unbounded disagreement?

### Answer

Open. It must distinguish candidate evidence, supported violations, duplicates,
in-mandate corrections, specialist referrals, durable issues, and human decisions.

## #6: Is a distinct mission-orchestration boundary needed?

Blocked by: #1, #4, #5, GitHub issue #76

Type: Grilling

GitHub work item:
[#104](https://github.com/dragondad22/codex-starter-kit/issues/104)

### Question

Does governed autonomous delivery require a new mission-orchestration module, and if so
what is its smallest non-overlapping identity, state, transition, recovery, and authority
contract above the lifecycle engine and Work Manager?

### Answer

Open. The no-new-module alternative must be evaluated. No prototype or runtime is
authorized by this ticket.

## #7: Can work packages remain semantically portable across providers?

Blocked by: #4, #6

Type: Research

GitHub work item:
[#105](https://github.com/dragondad22/codex-starter-kit/issues/105)

### Question

Which immutable work-package, capability, identity, privacy, cost, fallback, result, and
limitation semantics can be preserved across providers without requiring identical files,
commands, prompts, permission models, or runtime behavior?

### Answer

Open. The result will cross-link onboarding feature #94 without moving provider
interoperability into that feature.

## #8: What may Mission Control project or initiate?

Blocked by: #5, #6, #7, GitHub issue #93

Type: Grilling

GitHub work item:
[#106](https://github.com/dragondad22/codex-starter-kit/issues/106)

### Question

What may a future Mission Control surface truthfully display and initiate while remaining
a reproducible generated projection rather than a new authority?

### Answer

Open. Metric meaning remains owned by #93. No dashboard, service, database, or action
surface is authorized.

## #9: Is the broader factory path clear enough for executable decomposition?

Blocked by: #1, #2, #3, #4, #5, #6, #7, #8

Type: Grilling

GitHub work item:
[#107](https://github.com/dragondad22/codex-starter-kit/issues/107)

### Question

After the validation-only gate and downstream contracts resolve, what remaining
assurance-factory vertical-slice graph should be approved, deferred, narrowed, or
rejected?

### Answer

Open. Validation-only delivery may proceed through ticket #3. This later gate decides the
remaining broader factory graph; a no answer identifies the exact gap rather than
inventing implementation detail.
