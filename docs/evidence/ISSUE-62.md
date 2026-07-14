# Issue #62 — Distinct pull-request review decision record

**Date:** 2026-07-14
**Change owner:** dragondad22
**Issue:** [#62](https://github.com/dragondad22/codex-starter-kit/issues/62)
**Parent:** [#4](https://github.com/dragondad22/codex-starter-kit/issues/4)

## Approval and promotion

The product owner approved the universal review answer on 2026-07-14 and authorized its
durable promotion through issue #62. DEC-0020 is the authoritative result. It requires a
distinct capable review pass for every PR while preserving stronger human independence,
multiple-reviewer, CODEOWNER, and qualification requirements from effective policy.

The decision distinguishes implementation, implementer self-review, automated checks,
change review, requested-outcome authority, and qualified assurance. A separate capable
AI context can satisfy the universal review pass, but it does not silently satisfy a
policy requirement for a human, organizational separation, or domain qualification.

## Changed records

- Added DEC-0020 and routed it through the decision and documentation indexes.
- Linked the new review contract from DEC-0008, DEC-0017, and DEC-0019.
- Added `Distinct review pass` to the canonical glossary.
- Reconciled the owner persona's multiple-role context, outcome authority, and
  anti-assumptions without treating the owner as a code or assurance specialist.
- Added the review pass to the professional-baseline lifecycle and PRD implementation
  decisions.
- Assigned the universal review control to `github-delivery` while leaving stronger
  independence and qualifications to applicable assurance packs.

## Verification evidence

The issue branch runs:

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
git diff --check
```

The required `go test ./...` result is supplied by CI because Go is unavailable in the
local environment. The completing PR retains exact check and review evidence.

## Downstream reconciliation

Issue #64 is directly blocked by #62 and #63. Closing #62 removes only this blocker; #64
remains `Blocked` until #63 also closes. Its Work Manager prototype must model distinct
review results separately from checks, outcome approval, and stronger qualified assurance,
including source identity and invalidation when affected inputs change.
