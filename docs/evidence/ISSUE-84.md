# Issue #84 — Conversational capture evidence

**Date:** 2026-07-18
**Change owner:** dragondad22
**Issue:** [#84](https://github.com/dragondad22/codex-starter-kit/issues/84)
**Parent:** [#4](https://github.com/dragondad22/codex-starter-kit/issues/4)

## Outcome and authority

The product owner selected #84 for immediate implementation after confirming that the
equivalent Claude-kit behavior reduces forgotten work and inconsistent tracking. Existing
DEC-0005, DEC-0011, DEC-0013, and DEC-0022 semantics supply the authority boundary: GitHub
is operational memory, only consequential material crosses the conversational threshold,
standing workflow belongs in bootstrap and governed records, and approved decisions are
promoted beyond issue discussion.

The implemented contract requires an agent to search open and closed issues and offer a
capture action at a natural checkpoint when conversation reveals durable work, questions,
decisions, risks, or dependencies. It preserves ordinary clarification in conversation,
routes contained work to the active issue, blocks material implementation without
Readiness `Ready`, and does not authorize silent issue creation or decision approval.

## Delivered surfaces

- The repository `AGENTS.md` carries the concise point-of-use rule for active agents.
- The issue-tracker guide defines thresholds, checkpoints, routing, negative paths, active
  issue identity, and authoritative decision promotion.
- The PRD records the owner outcome and standing product behavior; the owner and AI actor
  records explain the attention and routing responsibilities.
- Lifecycle `create` emits the same concise rule in the generated managed-repository
  `AGENTS.md` and retains its generated ownership and content digest.
- Documentation validation protects the root contract, and lifecycle integration coverage
  exercises generated instructions through the public `create` interface.

## Verification evidence

The Python conversational capture validator followed a recorded RED/GREEN cycle: its
focused import/behavior test failed before the public validator existed, then passed after
the validator was implemented. Repository documentation validation also failed with five
missing governed markers before the root and issue-tracker rules were added.

The lifecycle `create` RED was reconstructed against the exact pre-change `main` source
`b6b3a42` in a temporary detached worktree. Only the public lifecycle assertion was added;
the production source remained unchanged. The focused command was:

```text
go test ./engine -run TestCreateRoutesConversationalCaptureThroughGeneratedAgentInstructionsRed -count=1
```

It failed because the generated `AGENTS.md` lacked all four asserted routes: existing
GitHub search, lifecycle-specific new work, authoritative decision promotion, and
Readiness `Ready`. The temporary assertion and worktree were removed after the result.
The committed focused test then passed against the implementation:

```text
go test ./engine -run TestCreateRoutesConversationalCaptureThroughGeneratedAgentInstructions -count=1
```

Required final commands are:

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
go test ./...
go run ./cmd/starter-kit changes check --repository .
git diff --check
```

The local environment resolves Python 3.12.3 and Go 1.26.5 for Linux amd64. Final results
are 35 passing Python unit tests, a passing documentation validator, a passing full Go
suite across every package, and a clean `git diff --check`. Product `changes check`
validated version `0.3.0`, 14 unreleased records, 13 external records, and one internal
record. Independent schema/digest verification also validated changelog digest
`sha256:2a4e5f03f255762b9978a4708e491a1df867e89a96f0f86356e12aebf31d900e`.

## Limitations and downstream work

This slice governs durable instructions and generated bootstrap content. It does not prove
that every model will recognize every conversational threshold, perform remote GitHub
orchestration, or automatically reconcile issues. Issue #74 remains responsible for the
broader executable-work intake/readiness engine route. Any future automated capture must
retain owner authority, duplicate handling, subtype semantics, and the non-flooding
boundary established here.
