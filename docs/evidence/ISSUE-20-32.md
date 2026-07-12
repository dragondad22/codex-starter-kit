# Issues #20 and #32 — Change and verification record

**Date:** 2026-07-12  
**Change owner:** dragondad22  
**Issues:** [#20](https://github.com/dragondad22/codex-starter-kit/issues/20),
[#32](https://github.com/dragondad22/codex-starter-kit/issues/32)

## Changed contract

- DEC-0001 retains its original regulated-project statement as superseded history and
  governs v1 through the amended sensitive-data assurance boundary.
- V1 distinguishes content classification, handling authorization, and product assurance;
  records the `No`/`Yes`/`Unsure` declaration; presents a notice for `Yes`/`Unsure`; and
  uses truthful `needs-review` or `unsupported` outcomes when assurance is absent.
- Detailed sensitive-data execution and verified regulatory coverage remain Later work
  in issue #21.
- Issue decomposition uses GitHub's native parent/sub-issue relationship and reconciles
  parent and child Project state independently.

## Verification evidence

The implementation branch ran:

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
git diff --check
```

Results before review corrections: 17 tests passed, documentation validation passed, and
the diff check passed. After review corrections, 19 tests passed and both documentation
validation and the diff check passed again. CI and the completing pull request retain the
final evidence and supersede these local results if the source changes.

The GitHub API also reported issues #25–#30 as native sub-issues of #2 and issue #32 as a
native sub-issue of #1. Project Status and Readiness were reconciled separately.

## Coverage and limitations

This change defines and validates the documentation contract; the lifecycle engine and
plugin do not yet implement it. The claim scan rejects known unqualified first-release
assurance phrases but cannot replace human review of novel wording. Formal route
assurance, DLP/egress enforcement, provider certification, and qualified regulatory pack
coverage are not configured and must not be reported as passing.
