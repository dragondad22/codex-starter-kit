# Issue #71 — deterministic managed-task lifecycle evidence

**Date:** 2026-07-15  
**Issue:** [#71](https://github.com/dragondad22/codex-starter-kit/issues/71)  
**Parent:** [#4](https://github.com/dragondad22/codex-starter-kit/issues/4)

## Delivered contract

- Added the production Work Manager contract behind lifecycle-engine inspect, plan,
  apply, verify, and status operations, plus one composite `ManageTask` request.
- Added strict language-neutral `starter-kit manage-task --input <json>` behavior.
- Added a credential-free in-memory adapter behind the same transport seam that #72 can
  implement for live GitHub without moving policy into the adapter.
- Added versioned desired intent, capability, normalized observation, immutable plan,
  per-effect receipt, verification, retry, and durable status values.
- Added self-digested atomic state at `.starter-kit/work-manager/state.json`; neither
  credentials nor raw transport requests are represented or persisted.
- Removed the issue #64 terminal prototype after its claimed scenarios were represented
  by production behavior and tests.

## Behavior and negative-path evidence

Engine and CLI integration tests cover the complete composite request; restart-safe
status; semantic no-change replay; governed-source, operating-profile, observation, and
configuration/option migration staleness; denied authority; ambiguous create recovery;
partial create/Project success and remaining-effect resume; offline/reconnect handshake;
expiry; bounded rate retry and reset; blocker-driven Readiness without Status selection;
parent Phase inheritance and incomplete-parent status; completion-to-Done and promotion
facts; separate review requirements; lifecycle-lease serialization; full capability
freshness at apply and verify; unresolved ambiguous-create blocking; invalid adapter-result
`needs-review`; strict schema and secret-shaped input/observation rejection before state;
adapter-detail redaction; and state-integrity failure after tamper.
Question/research completion validation keeps promotion outputs distinct from issue and
Project state, and mismatched observed managed IDs stop before planning.

The production policy owns desired lifecycle and relationship facts. The adapter owns
normalized observation and effect attempts only. Missing evidence and every failure remain
explicit; in-memory results are never relabeled as live GitHub evidence.

## Prototype absorption

The four issue #64 values are preserved as production values: credential-free desired
intent, capability plus observation, immutable plan, and per-effect receipt. Stable
markers, immutable IDs, semantic comparison, source/configuration/profile freshness,
partial-effect receipts, offline intent, bounded retry, and dependency-derived Readiness
all have automated production coverage. The prototype command, reducer, model, and notes
were deleted; the issue #64 evidence record remains the historical design source.

## Verification

The exact local branch used the cached Go 1.26.5 toolchain. Focused engine and CLI suites
passed during each RED/GREEN cycle. The pre-review candidate passed:

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
/home/chris/.cache/codex-starter-kit/toolchains/go1.26.5/bin/go test ./...
/home/chris/.cache/codex-starter-kit/toolchains/go1.26.5/bin/go vet ./...
/home/chris/.cache/codex-starter-kit/toolchains/go1.26.5/bin/go run ./cmd/starter-kit changes check --repository .
git diff --check
```

Results: 33 Python tests passed; documentation validation passed; every Go package and
`go vet` passed; release-change validation reported product `0.3.0`, eleven Unreleased
records, ten external records, and one internal-only record; and the diff check passed.
Distinct review and native CI results on the completing revision supersede this local
Linux evidence if the source changes.

## Distinct review

Two independent review axes examined the implementation diff from `origin/main` through
commit `9b52b0d`:

- Standards review: passed with no remaining violations of repository instructions,
  issue-tracker policy, or the Work Manager contract.
- Specification review: passed with no remaining mismatch against issue #71's acceptance
  criteria and the absorbed issue #64 scenarios.

Initial findings included untrusted adapter persistence, missing lifecycle serialization,
incomplete capability/configuration binding, insufficient verification capability,
unresolved ambiguous-create handling, incomplete schema validation, erased non-pass
outcomes, and incomplete question/promotion and managed-identity checks. Commits
`ebfdf2e` and `9b52b0d` resolved those findings, and both reviewers re-ran focused engine
and CLI tests before passing the exact implementation source.

## Limitations and downstream work

- No GitHub request, credential, repository, Project, paid feature, or external resource
  was created or changed by the implemented route.
- The in-memory adapter is deterministic development evidence, not live qualification.
- The current contract manages one task. Full governed intake/hierarchy/Horizon/Phase
  behavior remains #74 after #72 and the separately authorized #73 sandbox.
- GitHub authentication, REST/GraphQL transport, pagination, partial GraphQL errors,
  webhook hints, and live permission/rate behavior remain #72 and #76.
- Packaged binary provenance and exact native support require the completing CI matrix.

## GitHub reconciliation

Issue #71 was Project `Ready` and `Next` before implementation. Work began on the
issue-named branch and moved #71 to `In progress`; parent #4 remained `In progress`, and
direct dependent #72 remained `Backlog` / `Blocked`. Local implementation, verification,
distinct review, and commit evidence are complete. Issue #71 remains `In progress` until
the normal pull-request gates and squash merge complete; therefore #72 remains blocked.
