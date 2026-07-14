# DEC-0017 — Professional engineering baseline

**Status:** Accepted
**Owner:** dragondad22
**Date:** 2026-07-13
**Source decision:** Issue #23

## Context

AI can produce a superficially functional result quickly while omitting security,
maintainability, testing, documentation, recovery, accessibility, or product-quality
work. Requiring every user to supervise every decision would make the kit inaccessible;
reducing rigor for small, personal, or near-one-shot projects would defeat the reason the
kit exists.

Project consequences and governance needs vary, but “AI-generated,” “personal,” “small,”
or “one-shot” are not acceptable reasons for careless implementation.

## Decision

Adopt the canonical principle **quality is invariant; ceremony is variable**.

Every supported deliverable silently meets a professional engineering baseline. As
applicable to its real domain and claims, it uses maintainable code and defined standards;
applies relevant external guidance such as OWASP; continuously considers security,
privacy, secrets, dependencies, permissions, input boundaries, data handling, and abuse;
designs complete user, accessibility, failure, recovery, setup, update, and support
experiences; tests for gaps, bugs, regressions, and unmet acceptance; maintains appropriate
user, help, operator, and developer documentation; and verifies that the result matches
the request. Unsupported, uncertain, failed, and not-applicable conditions remain explicit.

Applicability varies; quality does not. A genuinely irrelevant control records
`not-applicable` with its facts and rule. Project policy may add stricter controls,
independent approval, evidence retention, or regulated review. It may not select a
knowingly careless or insecure mode and still produce a passing professional-quality
claim. Risk acceptance preserves the underlying result and cannot redefine failed work as
well implemented.

Keep three operating concerns independent:

- **Engagement mode** controls how often a human participates, from delegated execution
  using approved defaults to collaborative decision-by-decision work.
- Project-specific assurance adds standards, controls, reviewers, and gates above the
  universal baseline.
- **Evidence presentation and retention** controls whether the normal human view is a
  concise quality receipt or a complete inspectable package without hiding failures or
  reducing the evidence required by policy.

Near-one-shot operation reduces interaction, not verification. The result includes a
concise quality receipt showing what was requested and delivered, what standards and
checks applied, what evidence supports the result, and what remains limited or unresolved.
Collaborative and governed operation expose more decisions and evidence but do not begin
from a higher implementation-quality standard.

The system routes the minimum relevant outcome, scope, state, coverage, uncertainty, and
breadcrumb to the surface where the human naturally makes the decision. Deeper detail
remains in one authoritative record. Derived views reconcile when their inputs change;
the product does not require the human to know which hidden document contains essential
context or create parallel hand-maintained truth.

Exploratory or disposable prototypes may narrow their claims and lifespan, but must be
isolated, truthfully labeled, and prevented from silently becoming supported deliverables
until they meet the same baseline.

## Consequences

The kit needs trusted default standards and applicability rules capable of producing
professional results without a long interview. Operating profiles may change authority,
independence, retained evidence, and point-of-use detail; they cannot be quality tiers.

Verification must assess code, security, tests, documentation, and the full user
experience rather than only successful execution or a minimally functional interface.
Failures and missing evidence remain visible even when the human asks for a concise result.
[DEC-0020](DEC-0020-distinct-pull-request-review.md) adds the universal distinct PR-review
pass without treating stronger policy-required independence or qualifications as optional
ceremony.

## Source

Approved through [issue #23](https://github.com/dragondad22/codex-starter-kit/issues/23).
