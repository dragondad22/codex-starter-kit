# Issue #50 — Codex plugin compatibility and distribution record

**Date:** 2026-07-13
**Change owner:** dragondad22
**Issue:** [#50](https://github.com/dragondad22/codex-starter-kit/issues/50)

## Decision and authority

The product owner reviewed the bounded compatibility and distribution evaluation and
approved its reasoning and recommendation. DEC-0018 is the durable authority. The
research record retains the objective, stopping conditions, official sources, local
observations, documentation conflict, uncertainty, limitations, trust/authority/data/cost
implications, invalidation triggers, and downstream implementation impact.

The decision selects a skills-only Phase 2 plugin, repository marketplace development,
immutable Git-backed qualification snapshots, capability evidence instead of a minimum
client-version guess, four truthful workflow modes, pre-provisioned verified offline
inputs, independent plugin/engine/repository/policy versions, and direct engine fallback.
Public submission and publication remain Phase 6 work.

## Changed records

- Added DEC-0018 and linked it from the decision index.
- Promoted the research record from recommendation to approved decision provenance.
- Added the capability handshake and surface boundary to the architecture.
- Added the capability handshake and workflow capability mode to the canonical glossary.
- Reconciled the decision-record format with the established approved issue-backed source
  pattern used by DEC-0016 onward.
- Routed the decision, research, and evidence records from the documentation index.
- Decomposed Epic #3 into native issues #50–#54 while retaining its existing operating-
  profile child #47 and independently reconciling Project readiness and dependencies.

## Verification evidence

Before and after internal review corrections, the implementation branch ran these
available local checks:

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
git diff --check
```

The Python suite passed 22 tests, documentation validation passed, and `git diff --check`
passed. After internal review corrections, the same applicable local checks again passed
with 22 Python tests.

The required `go test ./...` command could not run because the local IDE host did not
expose `go` after reload. That is an explicit missing local verification capability, not
a pass. Native CI and the completing pull request must run the Go suite before this work
completes.

## Limitations and downstream work

No plugin, manifest, marketplace entry, capability handshake, engine package, native
qualification, or public submission is implemented by this decision. Current official
documentation conflicts about IDE plugin availability, so IDE distribution remains
`needs-review`. The observed CLI and IDE versions describe one Linux environment only.

Issues #51–#54 implement and qualify the contract. Issue #47 must be reconciled before its
operating-profile choices affect guided create, verification presentation, or aggregate
qualification. Phase 6 retains public publisher identity, legal/support materials,
signing, attestation, final native artifacts, release approval, and publication.
