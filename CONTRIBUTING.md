# Contributing

Thanks for helping build the Codex Starter Kit. The project is currently establishing its
foundation; start by reading the [public roadmap](docs/roadmap/IMPLEMENTATION_ROADMAP.md)
and [architecture](docs/architecture/ARCHITECTURE.md).

## Before writing code

1. Search open and closed GitHub issues for related work.
2. Open an issue, or ask to take an existing one.
3. Wait until the issue's Project Readiness is `Ready` before substantial implementation.
   Executor-routing labels such as `ready-for-agent` do not override a `Blocked` gate.
4. Create a branch named `<type>/<issue-number>-<short-slug>`.

Ready issues include a short human summary and a complete execution brief. If the issue
requires a new decision, is stale, or lacks acceptance/evidence requirements, refine it
before implementation.

## Pull requests

- Keep one coherent issue as the primary scope.
- Link it with `Closes #N` when the PR completes the work.
- Explain deviations and follow-up work rather than rewriting history.
- Update applicable product, architecture, persona, policy, and public documentation.
- Add a validated change record under `changes/unreleased/`, using an explicit
  internal-only disposition when no external audience should see the change.
- Regenerate `CHANGELOG.md` and require `starter-kit changes check --repository .` to pass.
- Include verification evidence and coverage limits.
- Use a Conventional Commit-style PR title; squash merge is the default.
- Use draft status only while more implementation, verification, or internal review is
  expected. Mark a completed and internally reviewed PR ready for review.

External pull requests enter the same triage workflow as issues. Maintainer or
collaborator work already in progress is not treated as unsolicited intake.

## Standards

The approved constraints are summarized in [AGENTS.md](AGENTS.md). Until executable
policy packs exist, the discovery decisions and architecture documents are authoritative.

## Local validation

Run the Python foundation checks and Go engine suite before opening a pull request:

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
go test ./...
```

CI runs the equivalent commands with Python 3.12 and Go 1.26.5 on native Ubuntu, macOS,
and Windows runners. Do not add a universal shell dependency to make validation pass.

## Conduct and security

Be respectful, specific, and evidence-driven. Report vulnerabilities privately according
to [SECURITY.md](SECURITY.md), not through public issues.
