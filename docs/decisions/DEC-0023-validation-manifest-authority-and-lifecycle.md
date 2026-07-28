# DEC-0023 — Validation-manifest authority and lifecycle

**Status:** Accepted
**Owner:** dragondad22
**Date:** 2026-07-28
**Source decision:** Issue #100

## Context

The professional engineering baseline, effective policy, operating profile, Ready issue,
human-owned specifications and decisions, and scoped execution mandate already govern one
delivery. The product needs a deterministic way to assemble their validation obligations
without making generated output a second specification, policy, approval, or effect
authority. Later assurance-factory research also needs a stable assertion-to-evidence
handoff before it can classify evidence methods or propose an executable slice.

A human-edited validation checklist would drift from its sources. A per-run document whose
identity included actors, timestamps, attempts, or results would prevent deterministic
reconstruction. A universal rule requiring or waiving separate human approval would also
ignore legitimate differences in work, actor, timing, cost, environment, and risk.

## Decision

A **validation manifest** is a generated, read-only checklist for one governed work scope.
One Ready issue anchors the initial scope. The manifest binds the exact applicable
human-owned specifications, decisions, architecture contracts, personas, risks, effective
policy, operating profile, and scoped assurance requirements. These inputs compose
additively; none silently overrides another. A semantic conflict produces
`needs-review`. A qualified human corrects an authoritative source or records an
applicable governed decision or exception, then regenerates the manifest. The generated
manifest is never edited to resolve authority.

Source resolution keeps missing inputs explicit. A source required by the Ready scope
whose applicability or content cannot be resolved produces `needs-review`. An explicitly
required but unconfigured source class or route produces `not-configured`; an unavailable
source capability produces `unsupported`. Invalid source schema, digest, or provenance is
rejected input and receives no lifecycle disposition. No expected source is silently
omitted.

The manifest records what must be proved and the acceptable evidence constraints. Each
assertion contains a stable assertion ID, exact source references, a normalized claim,
applicability and lifecycle gate, acceptable evidence constraints, freshness and
qualification requirements, and the required non-pass result when evidence is absent or
insufficient. Test implementations, assigned actors, execution attempts, results,
retries, and findings remain downstream plan, receipt, evidence, or disposition facts.
Issue #101 will classify assertion and evidence methods without changing this boundary.

The manifest ID is the digest of its canonical contents, including governed work scope
and source revision; exact source IDs and digests; effective-policy and operating-profile
revisions; manifest schema and compiler versions; and derived assertion and
evidence-obligation definitions. Identical inputs produce the same ID. Run facts belong
to separate validation receipts. Retained canonical inputs, versions, manifest, and
receipts reconstruct what was required and what a run observed.

Any bound input change makes the whole prior manifest stale and produces a new immutable
manifest. The old manifest and its receipts remain historical evidence but cannot support
a current pass. A new manifest may reuse prior evidence only when the assertion
fingerprint is unchanged, its evidence method permits reuse, and the evidence remains
fresh and valid for the new scope. The new receipt attributes that reuse; similar prose
or a prior aggregate pass is insufficient.

The lifecycle engine computes a **manifest assessment** rather than mutating the manifest.
Its dispositions are `current`, `stale`, `needs-review`, `not-configured`, and
`unsupported`. Invalid schema, digest, or provenance is rejected input and receives no
lifecycle disposition. Assessment occurs when compiling or inspecting, when planning,
immediately before evaluation, during verification, and whenever status is requested.
Only `current` may proceed. If a bound input changes during or after evaluation, produced
results remain historical but cannot establish a current pass. The manifest itself never
passes. Assertions may be `not-applicable` only with rationale, and risk acceptance
preserves the underlying result.

A **validation approval rule** is the human-owned governed choice that determines whether
evaluation of a current manifest requires an additional approval. Standing policy may
evaluate work type, actor, risk, timing, cost, environment, and effect scope. A qualified
human may record a narrower per-work rule with scope, rationale, and expiry. With no
applicable rule, inspection may continue but evaluation or effects stop at
`not-configured`. The rule is reassessed before evaluation because actor, timing, or
estimated cost may change. It is not effect authority: a validation manifest or approval
never replaces the DEC-0022 execution mandate or other required authority.

## Consequences

The product gains a reproducible assertion checklist without moving human-owned intent
into generated state. Manifest compilation and assessment must be observable through the
lifecycle-engine seam, but this decision does not select a new module, schema, assertion
taxonomy, evaluator kit, orchestration runtime, or provider. Those choices remain with
issues #101 and #108 and their downstream decision sequence.

Plans and receipts must retain the exact manifest ID and assessment inputs. Quality
receipts may summarize manifest-backed results but remain generated evidence views.
Implementations must distinguish approval from authority, current assessment from
assertion result, malformed input from lifecycle non-pass, and historical evidence from
evidence valid for the current manifest.

Rejected alternatives are a human-edited manifest, a manifest that authorizes effects,
run metadata inside manifest identity, partial in-place manifest mutation, inferred
evidence reuse, and a universal separate-approval rule. Each would respectively create
drift, broaden generated authority, destroy deterministic identity, obscure invalidation,
overstate evidence, or remove the human's governed choice.

## Source

Approved by the product owner through
[issue #100](https://github.com/dragondad22/codex-starter-kit/issues/100) as the first
investigation in the
[assurance-factory decision map](../roadmap/ASSURANCE_FACTORY_DECISION_MAP.md).
