# Issue #40 — Alphabetical glossary ordering

**Date:** 2026-07-13  
**Issue:** [#40](https://github.com/dragondad22/codex-starter-kit/issues/40)

## Delivered outcome

All 47 canonical `###` terms in the product glossary are in case-insensitive alphabetical
order. The term set and definition bodies are unchanged; only complete term blocks moved.

Domain-documentation guidance now states the standing order rule. The documentation
validator extracts canonical term headings and fails at the first mismatch with the
actual and expected terms, making future ordering drift visible in local and native CI.

## Test coverage

- An ordered mixed-case fixture passes.
- An out-of-order `Readiness`, `Question work item`, `Status` fixture fails with an
  actionable expected-term diagnostic.
- The repository's complete glossary passes the same validator used by CI.

## Verification

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
go test ./...
git diff --check
```

Local results: 21 Python tests passed; documentation validation passed; all Go packages
passed.
