# Issue #47 — Operating-profile decision record

**Date:** 2026-07-14
**Change owner:** dragondad22
**Issue:** [#47](https://github.com/dragondad22/codex-starter-kit/issues/47)

## Decision and authority

The product owner approved the compact operating-profile contract on 2026-07-14.
DEC-0019 is the durable authority. DEC-0017 remains the non-negotiable professional
engineering baseline and prevents any profile from becoming a lower-quality passing tier.

The approved default is delegated engagement, no discretionary assurance addition beyond
effective policy, and concise evidence presentation. Collaborative engagement is the
explicit alternative. Assurance additions attach additively at repository, work-item, or
release scope. Evidence presentation may expand or contract the normal view but may not
hide non-pass results or discard required evidence.

## Changed records

- Added DEC-0019 and linked it from the decision index.
- Sharpened `Engagement mode` and `Operating profile` in the canonical glossary.
- Added product stories and implementation decisions for defaults, scoped composition,
  mandatory interrupts, and prospective profile changes.
- Added effective-policy composition and context-routing responsibilities to the
  architecture.
- Added operating-profile defaults, interrupts, and transition behavior to the lifecycle
  contract.
- Routed the decision and this evidence record from the documentation index.

## Verification evidence

The documentation change ran the repository's complete documentation-change command set:

```text
python3 -m unittest discover -s tests -p "test_*.py" — 27 tests passed
python3 scripts/validate_docs.py — passed
go test ./... — unavailable because Go is not installed on the local workstation
git diff --check — passed
```

The completing pull request and native CI run retain exact results and source revision.
The unavailable local Go capability remains explicit rather than represented as pass.

## Downstream decomposition

The approved contract separates later implementation into four independently testable
responsibilities:

1. Plugin interaction and mandatory-interrupt behavior for delegated and collaborative
   engagement.
2. Engine effective-profile identity, stale-plan invalidation, and lifecycle transitions.
3. Additive repository/work-item/release assurance composition in effective policy.
4. Concise/expanded quality-receipt projection and policy-governed evidence retention.

Issue #54 consumes the plugin-facing default and receipt/interrupt contract for Phase 2
qualification. Later engine, policy-distribution, GitHub work-item, and release issues
must retain the same axes and history boundary rather than inventing named quality tiers.
