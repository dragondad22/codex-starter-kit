# Issue #63 — GitHub authentication and transport research record

**Date:** 2026-07-14
**Change owner:** dragondad22
**Issue:** [#63](https://github.com/dragondad22/codex-starter-kit/issues/63)
**Parent:** [#4](https://github.com/dragondad22/codex-starter-kit/issues/4)

## Scope and authority

The owner authorized one bounded research session comparing GitHub App installation,
user-token/GitHub CLI, and Actions `GITHUB_TOKEN` identities using current official GitHub
documentation and read-only probes of this repository and Project #8. The session did not
register or install an App, create or rotate a credential, use a paid service, broaden
permissions, mutate repository/Project content, or qualify a production adapter.

The durable research record is decision input, not product or architecture authority. Its
recommendation requires later promotion. Exact permissions, plan-dependent behavior, and
negative paths require the live sandbox qualification planned after the Work Manager
prototype.

## Changed records

- Added the authentication and transport evaluation with objective, stopping conditions,
  method, official provenance, local observation boundary, conflicting user-owned Project
  evidence, supported-mode recommendation, recovery contract, limitations, invalidation
  conditions, and downstream obligations.
- Added the issue-backed GitHub executable-work decision map and linked its resolved
  research ticket to issue #63.
- Enabled the accepted `type:question` and `type:research` labels required by DEC-0013.
- Created #62, #63, and #64 as native children of #4 and added them to Project #8.
- Initially set #62 and #63 to `In progress` / `Ready`; set #64 to `Backlog` / `Blocked`
  with native dependencies on #62 and #63. PR #66 subsequently closed #62 as `Done`, so
  #63 is now #64's remaining active blocker.
- Routed the research, planning map, and this evidence record from the documentation index.

## Research outcome

The recommendation assigns GitHub App installation to preferred unattended and CI
reconciliation, user tokens to interactive/personal-account/recovery work, and Actions
`GITHUB_TOKEN` to repository-local CI. The production adapter uses native Go HTTP with
version-pinned REST plus GraphQL where Projects or relationship coverage requires it;
GitHub CLI remains optional setup and diagnostics.

GitHub currently prevents one least-authority identity from covering every 1.0 audience.
Fine-grained PATs and documented user-owned Project mutations do not cover the personal
Project route, while a classic PAT has broader repository authority. Full App support
therefore begins with organization-owned Projects after qualification; the personal route
must surface and explicitly accept user-token breadth. GitHub Enterprise Server,
App-based user-owned Project mutation, private-plan features, and webhook durability
remain limitations until separately qualified.

## Read-only evidence snapshot

The selected account-specific credential resolved through the API as `dragondad22`. REST
reported administrator access to the public `dragondad22/codex-starter-kit` repository.
GraphQL read user-owned Project #8, its 15 fields, and a 5,000-point user budget; REST read
#4's native sub-issues and #15's empty blocked-by set. No mutation was used during the
research session.

A separate local diagnosis reproduced the open GitHub CLI multi-account keychain defect:
the default config displayed `dragondad22` while an implicit API call used
`chris-zoolytix`. All research and GitHub publication operations therefore selected the
account-specific keyring entry explicitly and verified `/user` before effects.

## Verification evidence

Before publication, the available local checks reported:

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
git diff --check
```

The Python suite passed 33 tests and documentation validation passed. `go test ./...` was
not run because Go is unavailable in this local environment; CI must provide the required
Go 1.26.5 result. Final `git diff --check` and documentation verification are rerun on the
issue branch after this evidence record is added.

## Downstream work

Issue #62 promoted the independently reviewed PR-assurance answer through DEC-0020 and is
closed as `Done`. Issue #64 consumes that answer and this research to prototype the
lifecycle-engine-facing Work Manager boundary. It remains blocked until #63 closes. The
later sandbox ticket must
qualify the exact identity/permission matrix, lost-response and rate-limit recovery,
Project-field migration, revocation, plan availability, and negative paths before the
final #4 decomposition can claim support.
