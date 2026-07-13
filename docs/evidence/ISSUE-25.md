# Issue #25 — Toolchain decision and verification record

**Date:** 2026-07-12
**Change owner:** dragondad22
**Issue:** [#25](https://github.com/dragondad22/codex-starter-kit/issues/25)

## Decision and authority

The owner approved the Go recommendation from the bounded lifecycle-engine toolchain
evaluation. DEC-0015 is the durable authority; the evaluation retains candidates,
criteria, primary sources, tradeoffs, downstream impact, exclusions, and invalidation
triggers.

The decision selects Go 1.26.5 for initial contributor and CI builds, a
standard-library-only Phase 1 implementation, native Linux/macOS/Windows evidence, a
language-neutral JSON/operation seam, SHA-256 manifests, GitHub artifact attestations,
offline verification material, and Rust as the preferred reimplementation fallback.

## Changed records

- Added DEC-0015 and D15 source history.
- Promoted the research record from recommendation to approved decision provenance.
- Updated architecture, PRD, personas, support boundaries, documentation routing, and the
  decision index.
- Extended mechanical decision-route validation and fixtures from D1–D14 to D1–D15.

## Verification evidence

The implementation branch ran:

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
git diff --check
```

Before and after review corrections, 19 tests passed, documentation validation passed,
and the diff check passed. CI and the completing pull request retain the final evidence
and supersede these local results if the source changes.

## Limitations and invalidation

No engine, CLI, installer, release binary, or signing pipeline is implemented by this
decision. Exact schemas, package layout, OS/runtime support, and release membership remain
separate work. DEC-0015 and the evaluation list the conditions that return the selection
to review.
