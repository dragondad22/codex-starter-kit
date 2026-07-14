# DEC-0020 — Distinct pull-request review

**Status:** Accepted
**Owner:** dragondad22
**Date:** 2026-07-14
**Source decision:** Issue #62

## Context

Automated checks can prove specific properties, and an implementer can find defects by
reviewing their own work, but neither supplies a second evaluative perspective. Requiring
an unrelated human reviewer for every personal repository would make the solo-first
contract impractical; treating implementer self-review or a test run as approval would
instead create false assurance. Effective policy may also require independence or
qualifications that a general code reviewer, project owner, or AI cannot supply.

## Decision

Every pull request requires a **distinct review pass** before it is complete and ready for
external review or merge. The reviewer is a declared capable human or AI actor that did
not implement the change in the same working context. A separate AI context qualifies for
the universal minimum when it begins from durable issue, governing records, diff, and
evidence rather than the implementation session. Implementer self-review and automated
checks remain required or useful supporting evidence but are not the distinct review
pass.

The universal minimum is not automatically independent or qualified assurance. Effective
policy may require a human, CODEOWNER, organizational separation, multiple reviewers, or
a reviewer with specific security, legal, accessibility, safety, regulatory, or other
qualifications. Missing required capability or separation blocks the affected gate and
remains explicit; a weaker review is not relabeled as sufficient.

Review roles compose instead of collapsing into one approval. A product owner may review
requested-outcome alignment without claiming code-language or assurance expertise. A
change reviewer evaluates the change within declared capability. A qualified assurance
reviewer evaluates the controls within their remit. One person may hold several roles
where policy permits, but one approval cannot silently stand for a role or capability it
did not exercise.

The review result names the reviewed source identity, reviewer actor and context,
declared capability, applicable role, findings or approval, evidence routes, limitations,
and any stronger effective-policy requirement. A material source or governing-input
change invalidates affected review evidence and requires re-evaluation.

## Consequences

Solo owners can use a separate capable AI context for the universal code-change review
while retaining product-outcome authority. They cannot count the implementation session's
self-review or passing automation as that review. Team, regulated, and risk-sensitive
repositories can add stronger human independence and qualifications without changing the
universal baseline or inventing a quality tier.

The Work Manager must represent review requirements and results separately from checks,
implementation, outcome approval, and assurance approval. GitHub rulesets and branch
protection are enforcement adapters; their configuration does not itself prove that the
required reviewer was capable, distinct, independent, or qualified.

## Source

Approved by the product owner through
[#62](https://github.com/dragondad22/codex-starter-kit/issues/62). DEC-0008 supplies the
protected PR flow, DEC-0017 supplies the professional baseline, and DEC-0019 supplies
additive scoped assurance requirements.
