# Issue #87 — Skill eligibility and informed-interaction decision record

**Date:** 2026-07-29
**Change owner:** dragondad22
**Issue:** [#87](https://github.com/dragondad22/codex-starter-kit/issues/87)
**Parent:** [#86](https://github.com/dragondad22/codex-starter-kit/issues/86)

## Approval and promotion

The product owner approved the layered contract after reviewing its practical
consequences and candidate dispositions. DEC-0024 is the durable authority:

- informed decision support is universal product behavior rather than an optional skill;
- a skill must pass recognizable-goal, activation, reusable-value, surface-ownership,
  bounded-contract, qualification, and net-catalog-value tests;
- scope and distribution express audience but grant no authority; and
- product-skill promotion still requires task-specific Ready work and evidence.

The owner also approved retaining product `create`, `status`, and `verify`; keeping the
four broad workflow skills personal or routing their behavior to existing product
surfaces; treating #114 and #88 as later qualified candidates; preferring deterministic
Work Manager state before #89 guidance; and leaving #90 as research.

## Changed records

- Added the current official-source evaluation and complete surface/candidate comparison.
- Added DEC-0024 and linked it from the decision and documentation indexes.
- Amended DEC-0018's dated capability snapshot to distinguish standalone IDE skills from
  plugins, which current documentation does not support in the IDE.
- Added the skill and informed-interaction outcomes to the PRD, Guide behavior, Context
  Router/plugin architecture, operations guidance, and canonical glossary.
- Added the structured product change record and regenerated the changelog.

## Negative-path disposition

| Scenario | Required result |
|---|---|
| The user asks for a material decision without enough context | Investigate or ask; do not manufacture a confident recommendation |
| The agent has a recommendation but alternatives/consequences are hidden | Present the decision brief before requesting approval |
| Discovery repeats known facts or follows a fixed questionnaire | Use discoverable context and ask only outcome-changing questions |
| A template supports many artifacts | Create only records justified by approved durable results |
| A candidate has no clear trigger or overlaps broadly | Refine, require explicit invocation, combine, route elsewhere, or reject |
| The behavior is deterministic or standing policy | Put it in the engine, policy, template, hook, automation, or repository contract |
| A skill or plugin is installed/active | Grant no implied data, tool, decision, tracker, or effect authority |
| A dependency or host capability is missing or unsupported | Preserve the explicit limitation and bounded fallback; do not simulate a result |
| Conversation reaches apparent agreement | Do not promote, mutate trackers, or implement without the separately required authority |
| One happy-path scenario passes | Do not claim qualification without activation, negative-path, authority, dependency, and supported-surface evidence |

## Verification evidence

The final exact-head verification record is completed before promotion:

```text
python3 -m unittest discover -s tests -p "test_*.py" — 41 tests passed
python3 scripts/validate_docs.py — passed
go test ./... — passed
starter-kit changes check --repository . — passed; generated changelog current
git diff --check — passed
```

The pull request and native CI retain exact source-revision, check, and distinct-review
evidence.

## Downstream reconciliation

Issue #87 completes the common contract but does not make a new skill implementation
Ready. Parent #86 remains in progress. #88 retains #75 as a separate blocker and requires
fresh task fitness after both predecessors complete. #89 routes first to a deterministic
Work Manager projection. #90 remains bounded host/distribution research. #114 is the next
informed-shaping candidate to refine, with its exact artifacts and qualification still
unresolved.
