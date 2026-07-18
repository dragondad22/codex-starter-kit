# DEC-0019 — Operating-profile contract

**Status:** Accepted
**Owner:** dragondad22
**Date:** 2026-07-14
**Source decision:** Issue #47

## Context

DEC-0017 establishes one professional engineering baseline while allowing ceremony to
vary. The product still needs a compact configuration that lets a person delegate routine
delivery or collaborate closely, add assurance at the scope where it applies, and choose
how much evidence appears in the normal view. Combining those concerns into named quality
tiers would make low interaction look like permission for weaker work and make profile
changes silently alter historical claims.

## Decision

An **operating profile** is a versioned configuration of three independent axes:

1. **Engagement mode** is `delegated` or `collaborative`. `delegated` is the default and
   permits execution within established authority and approved defaults. `collaborative`
   adds checkpoints at planning and material decisions or effects. Neither changes the
   applicable controls or pass criteria.
2. **Assurance additions** are versioned requirements attached to repository, work-item,
   or release policy. The default declares no discretionary addition, but DEC-0017 and
   every automatically applicable or governing policy still apply. Additions compose;
   a narrower scope may strengthen a broader requirement but cannot remove it. A conflict
   that cannot be safely composed becomes `needs-review`.
3. **Evidence presentation** is `concise` or `expanded`; `concise` is the default. Both
   preserve the evidence required by effective policy and link every non-pass,
   limitation, accepted risk, and unresolved result. Retention is determined separately
   by effective policy and handling rules, not weakened by a presentation preference.

The default profile is therefore `delegated` engagement, no discretionary assurance
addition beyond effective policy, and a `concise` evidence view.

Delegated execution must interrupt before continuing when it encounters an unresolved
product, architecture, policy, regulatory, or risk decision; new, ambiguous, or broader
authority; destructive, externally visible, installation, or network effects outside an
approved scoped execution mandate; installation or credential broadening;
sensitive-data uncertainty or an unassured route; failed, missing, stale, or conflicting
required evidence; an unsafe/non-recoverable condition; or a material change to the
approved outcome, plan, assumptions, cost, or compatibility. Conversation and prior
approval do not imply authority for a different effect. DEC-0022 governs mandate
containment: a refreshed plan digest is not a different effect when its semantic delta
remains wholly inside the approved mandate.

Every delivery produces a quality receipt. Its concise view leads with the requested and
delivered outcome, applied profile and policy identities, checks and aggregate result,
evidence routes, effects and approvals, limitations, and unresolved work. The expanded
view exposes the complete inspectable package without becoming a second source of truth.

Profile selection and change are explicit, attributable, and prospective. A work-item
engagement or evidence-view override applies only to that item. Repository, work-item,
and release assurance additions remain attached to their governing scopes. Changing a
profile invalidates affected active plans and derived views for re-evaluation; it never
rewrites prior receipts, evidence, decisions, exceptions, or claims.

## Consequences

The product can offer a near-one-shot default without a lower-quality path. Implementers
must model the three axes separately, retain their source and scope, and compute the
effective profile alongside effective policy. Plugin prompts may reduce interaction but
must surface mandatory interrupts and mandate-containment failures. The engine must treat a
changed effective profile as a stale-plan input. Evidence views may vary in depth while
their underlying results and retention obligations remain identical.

Repository, work-item, and release policy implementations need additive composition and
explicit conflict handling. Quality-receipt implementations need stable routes from the
concise view to machine evidence and human-owned authority records. Historical profile
identity remains part of evidence provenance. DEC-0022 also clarifies that collaborative
engagement changes visibility and decision cadence rather than requiring confirmation of
every in-mandate effect.

[DEC-0020](DEC-0020-distinct-pull-request-review.md) requires a distinct PR-review pass in
every profile. Assurance additions may strengthen its reviewer independence, number, or
qualifications but cannot remove or weaken the universal pass.

## Source

Approved by the product owner through [issue #47](https://github.com/dragondad22/codex-starter-kit/issues/47).
DEC-0017 remains the governing universal quality boundary.
