# Issue #78 — Product version and release-change evidence

**Date:** 2026-07-15
**Issue:** [#78](https://github.com/dragondad22/codex-starter-kit/issues/78)
**Parent:** [#7](https://github.com/dragondad22/codex-starter-kit/issues/7)
**Aggregate release issue:** [#45](https://github.com/dragondad22/codex-starter-kit/issues/45)

## Delivered contract

- Established root schema-v1 `product-version.json` `0.3.0` as the product release
  identity and synchronized the plugin manifest and both plugin capability contracts.
- Added `starter-kit version`; the engine capability handshake and newly generated
  verification evidence report the same product version while preserving separate
  protocol, schema, policy, baseline, and managed-state compatibility identities.
- Promoted the approved design into DEC-0021 and published maintainer/user guidance.
- Added human-owned schema-v1 JSON change records with deterministic categories,
  audiences, components, issue/PR references, breaking state, and explicit internal-only
  disposition.
- Added all-audience, audience-filtered, and release-bounded Markdown rendering with a
  deterministic source digest, explicit validation, and a stale-generated-changelog
  check. Backfilled concise notable development capabilities as Unreleased without
  claiming historical publication.
- Added local release preparation that requires a greater stable SemVer, explicit date,
  and approved admission binding exact records to a Milestone and aggregate issue;
  archives exact record and admission bytes with SHA-256 digests, synchronizes
  same-release version surfaces, regenerates the changelog, and records `prepared` plus
  `published: false`.
- Added the release-change check to all three native CI foundation jobs and the pull
  request completion contract.

## Behavior and negative-path evidence

Public CLI integration tests exercise canonical product identity, the capability report,
verification evidence identity, deterministic Unreleased rendering, audience filtering,
CI validation counts, version skew, unsafe and duplicate IDs, trailing JSON, unknown
audiences, missing internal-only disposition, stale generated views, successful prepared
archive creation, and refusal of a non-incrementing version without mutation.

PR CI additionally exposed native checkout line-ending variance: Windows checked out the
generated changelog with CRLF while the deterministic renderer emitted LF. The freshness
check now normalizes CRLF only at the comparison boundary, retains strict content
comparison, and has regression coverage alongside the stale-view negative path.

Release preparation preflights all inputs, admits only named records, uses same-directory
temporary files and backups for native replacement, and records originals in a durable
recovery journal before mutation. The recovery command restores interrupted work. It
never performs Git, GitHub, network, approval, signing, artifact, or deployment effects.
Archived records and admissions are re-read through authority, identity, and digest checks
before rendering; record IDs cannot be reused across release history.

## Verification

The final pre-review local tree used the repository's cached, exact Go 1.26.5 toolchain:

```text
python3 -m unittest discover -s tests -p "test_*.py"
python3 scripts/validate_docs.py
/home/chris/.cache/codex-starter-kit/toolchains/go1.26.5/bin/go test ./...
/home/chris/.cache/codex-starter-kit/toolchains/go1.26.5/bin/go run ./cmd/starter-kit changes check --repository .
git diff --check
```

Results: 33 Python tests passed; documentation validation passed; every Go package passed;
the release-change check reported product `0.3.0`, ten Unreleased records, nine external
records, and one internal-only record; and the diff check passed. Final CI and review
results supersede this pre-review snapshot if the source changes.

## Limitations and downstream work

- No Git tag, GitHub Release, package, executable, signature, SBOM, attestation,
  deployment, or public release was created. Product version `0.3.0` remains a development
  identity, not a supported-publication claim.
- Multi-file local preparation is recoverable rather than crash-atomic. After hard
  termination, a retained transaction journal blocks another preparation until explicit
  `release recover` compensation succeeds.
- Publication requires a dependent bounded GitHub release adapter, an exact merged
  candidate, its own finite Milestone and aggregate release issue, applicable gates and
  evidence, explicit human approval, remote tag/Release verification, and truthful partial
  failure reconciliation.
- The `1.0.0` aggregate issue remains incomplete. This foundation does not make it Release
  Candidate-ready or satisfy signing, packaging, pilots, rollback, or publication gates.

## GitHub reconciliation

Issue #78 was Ready before implementation, is a native child of #7, belongs to the `1.0.0`
Milestone, and moved to Project Status `In progress` when work began. Parent #7 remains `In
progress`; aggregate issue #45 remains open. Completion reconciliation occurs only after
the implementation is committed and its required review state is known.
